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

type NowPlayingInfo struct {
	Artists   []string
	Track     string
	Album     string
	Duration  int64
	Timestamp int64

	PlaybackStatus string
	Position       int64
}

func (n NowPlayingInfo) JoinArtists() string {
	return strings.Join(n.Artists, ", ")
}

type ParsedRegexEntry struct {
	Match   *regexp.Regexp
	Replace string
	Artist  bool
	Track   bool
	Album   bool
}

func (n NowPlayingInfo) Equals(other NowPlayingInfo) bool {
	return reflect.DeepEqual(n.Artists, other.Artists) &&
		n.Track == other.Track &&
		n.Album == other.Album
}

func (n NowPlayingInfo) Valid() bool {
	switch {
	case n.Album == "":
		return false
	case n.Track == "":
		return false
	case n.JoinArtists() == "":
		return false
	case n.Duration == 0:
		return false
	default:
		return true
	}
}

// https://dbus.freedesktop.org/doc/dbus-specification.html
// https://specifications.freedesktop.org/mpris-spec/latest/
func GetNowPlaying(
	conn *dbus.Conn,
	playerBlacklist []*regexp.Regexp,
	regexes []ParsedRegexEntry,
) (map[string]NowPlayingInfo, error) {
	var dbusNames []string
	if err := conn.
		Object("org.freedesktop.DBus", "/org/freedesktop/DBus").
		Call("org.freedesktop.DBus.ListNames", 0).
		Store(&dbusNames); err != nil {
		return nil, err
	}

	var playerNames []string
	for _, name := range dbusNames {
		if strings.HasPrefix(name, "org.mpris.MediaPlayer2.") && !isBlacklisted(playerBlacklist, name) {
			playerNames = append(playerNames, name)
		}
	}

	info := map[string]NowPlayingInfo{}

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

		artists, err1 := getMapEntry[[]string](*metadata, "xesam:artist")
		track, err2 := getMapEntry[string](*metadata, "xesam:title")
		album, err3 := getMapEntry[string](*metadata, "xesam:album")
		duration, err4 := getMapEntry[int64](*metadata, "mpris:length")

		if err := errors.Join(err1, err2, err3, err4); err != nil {
			log.Warn().
				Str("player", player).
				Msg("error parsing metadata for player")
			continue
		}

		for _, r := range regexes {
			log.Debug().
				Str("expression", r.Match.String()).
				Str("replacement", r.Replace).
				Msg("running match/replace substitution")

			if r.Artist {
				var newArtists []string
				for _, artist := range *artists {
					newArtist := r.Match.ReplaceAllString(artist, r.Replace)
					newArtists = append(newArtists, newArtist)
				}
				*artists = newArtists
			}
			if r.Track {
				*track = r.Match.ReplaceAllString(*track, r.Replace)
			}
			if r.Album {
				*album = r.Match.ReplaceAllString(*album, r.Replace)
			}
		}

		info[player] = NowPlayingInfo{
			Artists:        *artists,
			Track:          *track,
			Album:          *album,
			Duration:       *duration / 1_000_000,
			Timestamp:      0,
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
