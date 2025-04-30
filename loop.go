package main

import (
	"fmt"
	"regexp"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog/log"
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

	for {
		nowPlaying, err := GetNowPlaying(conn, blacklist)
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
					log.Debug().
						Str("player", player).
						Str("provider", provider.Name()).
						Interface("status", status).
						Msg("sending now playing info")
					provider.NowPlaying(status)
				}

				continue
			}

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
				provider.NowPlaying(previouslyPlaying[player])
				provider.Scrobble(previouslyPlaying[player])
			}
		}

		log.Debug().
			Dur("duration", time.Duration(config.PollRate)).
			Msg("waiting for next poll")
		time.Sleep(time.Second * time.Duration(config.PollRate))
	}
}

func minPlayTime(nowPlaying NowPlaying, config *Config) (int64, error) {
	if nowPlaying.Duration < 0 {
		return 0, fmt.Errorf("invalid track length: %d", nowPlaying.Duration)
	}

	half := int64((float64(nowPlaying.Duration) / 100) * float64(config.MinPlaybackPercent))
	return min(half, config.MinPlaybackDuration), nil
}
