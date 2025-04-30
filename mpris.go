package main

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog/log"
)

// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Enum:Playback_Status
const (
	PlaybackPlaying = "Playing"
	PlaybackPaused  = "Paused"
	PlaybackStopped = "Stopped"
)

type NowPlaying struct {
	Artists   []string
	Track     string
	Album     string
	Duration  int64
	Timestamp int64

	PlaybackStatus string
	Position       int64
}

func NowPlayingEquals(left NowPlaying, right NowPlaying) bool {
	return reflect.DeepEqual(left.Artists, right.Artists) &&
		left.Track == right.Track &&
		left.Album == right.Album
}

func NowPlayingValid(n NowPlaying) bool {
	switch {
	case n.Album == "":
		return false
	case n.Track == "":
		return false
	case strings.Join(n.Artists, ", ") == "":
		return false
	case n.Duration == 0:
		return false
	default:
		return true
	}
}

func GetNowPlaying(conn *dbus.Conn, blacklist []*regexp.Regexp) (map[string]NowPlaying, error) {
	var dbusNames []string
	err := conn.Object("org.freedesktop.DBus", "/org/freedesktop/DBus").
		Call("org.freedesktop.DBus.ListNames", 0).Store(&dbusNames)
	if err != nil {
		return nil, err
	}

	var playerNames []string
	for _, name := range dbusNames {
		if strings.HasPrefix(name, "org.mpris.MediaPlayer2.") && !isBlacklisted(blacklist, name) {
			playerNames = append(playerNames, name)
		}
	}

	info := map[string]NowPlaying{}

	for _, player := range playerNames {
		playerObj := conn.Object(player, "/org/mpris/MediaPlayer2")

		metadata, err1 := getProperty[map[string]dbus.Variant](playerObj, "org.mpris.MediaPlayer2.Player.Metadata")
		playbackStatus, err2 := getProperty[string](playerObj, "org.mpris.MediaPlayer2.Player.PlaybackStatus")
		position, err3 := getProperty[int64](playerObj, "org.mpris.MediaPlayer2.Player.Position")

		if err := errors.Join(err1, err2, err3); err != nil {
			log.Error().
				Str("player", player).
				Err(err).
				Msg("error reading DBus properties for player")
			continue
		}

		track, err1 := getMapEntry[string](*metadata, "xesam:title")
		artists, err2 := getMapEntry[[]string](*metadata, "xesam:artist")
		album, err3 := getMapEntry[string](*metadata, "xesam:album")
		duration, err4 := getMapEntry[int64](*metadata, "mpris:length")

		if err := errors.Join(err1, err2, err3, err4); err != nil {
			log.Warn().
				Str("player", player).
				Msg("error parsing metadata for player")
			continue
		}

		info[player] = NowPlaying{
			Artists:        *artists,
			Track:          *track,
			Album:          *album,
			Duration:       *duration / 1_000_000,
			PlaybackStatus: *playbackStatus,
			Position:       *position / 1_000_000,
		}
	}

	return info, nil
}

func getProperty[E any](obj dbus.BusObject, property string) (*E, error) {
	value, err := obj.GetProperty(property)
	if err != nil {
		return nil, fmt.Errorf("failed to get property: %v", err)
	}

	if parsedValue, ok := value.Value().(E); ok {
		return &parsedValue, nil
	}
	return nil, errors.New("failed to read property from DBus object")
}

func getMapEntry[E any](metadata map[string]dbus.Variant, key string) (*E, error) {
	value, ok := metadata[key]
	if !ok {
		return nil, fmt.Errorf("map entry with key %s not found", key)
	}

	if parsedValue, ok := value.Value().(E); ok {
		return &parsedValue, nil
	}
	return nil, fmt.Errorf("invalid data type for map entry %s", key)
}

func isBlacklisted(blacklist []*regexp.Regexp, player string) bool {
	for _, re := range blacklist {
		if re.MatchString(player) {
			return true
		}
	}
	return false
}
