package main

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog/log"
)

// https://dbus.freedesktop.org/doc/dbus-specification.html
// https://specifications.freedesktop.org/mpris-spec/latest/
func GetInfo(
	playerBlacklist []*regexp.Regexp,
	regexes []ParsedRegexReplace,
) (map[string]PlaybackStatus, error) {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return nil, err
	}
	defer CloseDBus(conn)

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

	playerPlaybackStatus := map[string]PlaybackStatus{}

	for _, player := range playerNames {
		playerObj := conn.Object(player, "/org/mpris/MediaPlayer2")

		metadata, err1 := getProperty[map[string]dbus.Variant](playerObj, "org.mpris.MediaPlayer2.Player.Metadata")
		status, err2 := getProperty[string](playerObj, "org.mpris.MediaPlayer2.Player.PlaybackStatus")
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

		playbackStatus := PlaybackStatus{
			Artists:        *artists,
			Track:          *track,
			Album:          *album,
			Duration:       *duration / 1_000_000,
			Timestamp:      0,
			PlaybackStatus: *status,
			Position:       *position / 1_000_000,
		}

		playbackStatus.RegexReplace(regexes)

		playerPlaybackStatus[player] = playbackStatus
	}

	return playerPlaybackStatus, nil
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
