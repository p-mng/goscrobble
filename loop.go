package main

import (
	"fmt"
	"regexp"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog/log"
)

var nowPlayingNotificationID uint32

const (
	RuneBeamedSixteenthNotes = '\u266C'
	RuneCheckMark            = '\u2713'
)

func RunMainLoop(conn *dbus.Conn, config *Config) {
	log.Debug().Msg("started main loop")

	previouslyPlaying := map[string]NowPlaying{}
	scrobbledPrevious := map[string]bool{}

	var blacklist []*regexp.Regexp

	for _, expression := range config.Blacklist {
		compiled, err := regexp.Compile(expression)
		if err != nil {
			log.Warn().
				Str("expression", expression).
				Err(err).
				Msg("failed to compile regex blacklist entry")
			continue
		}

		log.Debug().
			Str("expression", expression).
			Msg("compiled regex blacklist entry")
		blacklist = append(blacklist, compiled)
	}

	parsedRegexes := config.ParseRegexes()
	log.Debug().Msg("parsed match/replace expressions")

	for {
		nowPlaying, err := GetNowPlaying(conn, blacklist, parsedRegexes)
		if err != nil {
			log.Error().Err(err).Msg("failed to get current playback status")
			continue
		}

		for player := range nowPlaying {
			if _, ok := previouslyPlaying[player]; !ok {
				log.Info().
					Str("player", player).
					Msg("new player found")
				previouslyPlaying[player] = NowPlaying{}
				scrobbledPrevious[player] = false
			}
		}

		for player := range previouslyPlaying {
			if _, ok := nowPlaying[player]; !ok {
				log.Info().
					Str("player", player).
					Msg("player disappeared")
				delete(previouslyPlaying, player)
				delete(scrobbledPrevious, player)
			}
		}

		for player, status := range nowPlaying {
			if !NowPlayingValid(status) {
				continue
			}

			minPlayTime, err := minPlayTime(status, config)
			if err != nil {
				log.Warn().
					Str("player", player).
					Interface("status", status).
					Err(err).
					Msg("cannot calculate minimum playback time")
				continue
			}

			if !NowPlayingEquals(status, previouslyPlaying[player]) {
				log.Info().
					Str("player", player).
					Interface("status", status).
					Msg("started playback of new track")

				status.Position = 0
				status.Timestamp = time.Now().Unix()

				previouslyPlaying[player] = status
				scrobbledPrevious[player] = false

				for _, provider := range config.Providers() {
					updateNowPlaying(player, provider, status, conn, config)
				}

				continue
			}

			status.Timestamp = previouslyPlaying[player].Timestamp

			if status.Position < minPlayTime || status.PlaybackStatus != PlaybackPlaying || scrobbledPrevious[player] {
				continue
			}

			log.Info().
				Str("player", player).
				Interface("status", status).
				Msg("scrobbling track")

			scrobbledPrevious[player] = true

			for _, provider := range config.Providers() {
				log.Debug().
					Str("player", player).
					Str("provider", provider.Name()).
					Interface("status", status).
					Msg("sending now playing and scrobble info")

				updateNowPlaying(player, provider, status, conn, config)
				sendScrobble(player, provider, status, conn, config)
			}
		}

		log.Debug().
			Dur("duration", time.Duration(config.PollRate)).
			Msg("waiting for next poll")
		time.Sleep(time.Second * time.Duration(config.PollRate))
	}
}

func updateNowPlaying(player string,
	provider Provider,
	status NowPlaying,
	conn *dbus.Conn,
	config *Config,
) {
	log.Debug().
		Str("player", player).
		Str("provider", provider.Name()).
		Interface("status", status).
		Msg("updating now playing status")
	if err := provider.NowPlaying(status); err != nil {
		log.Error().
			Str("player", player).
			Str("provider", provider.Name()).
			Interface("status", status).
			Err(err).
			Msg("error updating now playing status")

		if config.NotifyOnError {
			notify(
				conn,
				IconSyncError,
				"error updating now playing status",
				fmt.Sprintf("error updating now playing status: %s", err.Error()),
				false,
			)
		}
	} else {
		log.Debug().
			Str("player", player).
			Str("provider", provider.Name()).
			Interface("status", status).
			Msg("updated now playing status")

		if config.NotifyOnScrobble {
			notify(
				conn,
				IconSyncronizing,
				fmt.Sprintf("%c now playing (%s)", RuneBeamedSixteenthNotes, provider.Name()),
				fmt.Sprintf("now playing: %s - %s", status.JoinArtists(), status.Track),
				true,
			)
		}
	}
}

func sendScrobble(player string,
	provider Provider,
	status NowPlaying,
	conn *dbus.Conn,
	config *Config,
) {
	log.Debug().
		Str("player", player).
		Str("provider", provider.Name()).
		Interface("status", status).
		Msg("saving scrobble")
	if err := provider.Scrobble(status); err != nil {
		log.Error().
			Str("player", player).
			Str("provider", provider.Name()).
			Interface("status", status).
			Err(err).
			Msg("error saving scrobble")

		if config.NotifyOnError {
			notify(
				conn,
				IconSyncError,
				fmt.Sprintf("error saving scrobble (%s)", provider.Name()),
				fmt.Sprintf("error saving scrobble: %s", err.Error()),
				false,
			)
		}
	} else {
		log.Info().
			Str("player", player).
			Str("provider", provider.Name()).
			Interface("status", status).
			Msg("saved scrobble")

		if config.NotifyOnScrobble {
			notify(
				conn,
				IconSyncronizing,
				fmt.Sprintf("%c saved scrobble (%s)", RuneCheckMark, provider.Name()),
				fmt.Sprintf("saved scrobble: %s - %s", status.JoinArtists(), status.Track),
				false,
			)
		}
	}
}

func minPlayTime(nowPlaying NowPlaying, config *Config) (int64, error) {
	if nowPlaying.Duration < 0 {
		return 0, fmt.Errorf("invalid track length: %d", nowPlaying.Duration)
	}

	half := int64((float64(nowPlaying.Duration) / 100) * float64(config.MinPlaybackPercent))
	return min(half, config.MinPlaybackDuration), nil
}

func notify(conn *dbus.Conn, appIcon, summary, body string, isNowPlaying bool) {
	var replacesID uint32
	if isNowPlaying {
		replacesID = nowPlayingNotificationID
	}

	id, err := SendNotification(conn, replacesID, appIcon, summary, body)
	if err != nil {
		log.Error().
			Err(err).
			Interface("notification", map[string]string{
				"icon":    appIcon,
				"summary": summary,
				"text":    body},
			).
			Msg("error sending desktop notification via dbus")
	}

	if isNowPlaying {
		nowPlayingNotificationID = id
	}
}
