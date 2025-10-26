package main

import (
	"fmt"
	"regexp"
	"time"

	"github.com/rs/zerolog/log"
)

var nowPlayingNotificationID uint32

const (
	RuneBeamedSixteenthNotes = '\u266C'
	RuneCheckMark            = '\u2713'
	RuneEmDash               = '\u2014'
	RuneWarningSign          = '\u26A0'
)

func RunMainLoop(config *Config) {
	log.Debug().Msg("started main loop")

	previouslyPlaying := map[string]Info{}
	scrobbledPrevious := map[string]bool{}

	var playerBlacklist []*regexp.Regexp

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
		playerBlacklist = append(playerBlacklist, compiled)
	}

	parsedRegexes := config.ParseRegexes()
	log.Debug().Msg("parsed match/replace expressions")

	for {
		log.Debug().
			Dur("duration", time.Duration(config.PollRate)).
			Msg("waiting for next poll")
		time.Sleep(time.Second * time.Duration(config.PollRate))

		playbackInfo, err := GetInfo(playerBlacklist, parsedRegexes)
		if err != nil {
			log.Error().Err(err).Msg("failed to get current playback status")
			continue
		}

		for player := range playbackInfo {
			if _, ok := previouslyPlaying[player]; !ok {
				log.Info().
					Str("player", player).
					Msg("new player found")
				previouslyPlaying[player] = Info{}
				scrobbledPrevious[player] = false
			}
		}

		for player := range previouslyPlaying {
			if _, ok := playbackInfo[player]; !ok {
				log.Info().
					Str("player", player).
					Msg("player disappeared")
				delete(previouslyPlaying, player)
				delete(scrobbledPrevious, player)
			}
		}

		for player, status := range playbackInfo {
			if !status.Valid() {
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

			if !status.Equals(previouslyPlaying[player]) {
				status.Position = 0
				status.Timestamp = time.Now().Unix()

				previouslyPlaying[player] = status
				scrobbledPrevious[player] = false

				log.Info().
					Str("player", player).
					Interface("status", status).
					Msg("started playback of new track")

				if config.NotifyOnScrobble {
					nowPlayingNotificationID = sendNotification(
						fmt.Sprintf("%c now playing: %s", RuneBeamedSixteenthNotes, status.Track),
						fmt.Sprintf("%s %c %s", status.JoinArtists(), RuneEmDash, status.Album),
						nowPlayingNotificationID,
					)
				}

				for _, provider := range config.Providers() {
					sendNowPlaying(player, provider, status, config)
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

			if config.NotifyOnScrobble {
				sendNotification(
					fmt.Sprintf("%c scrobbling: %s", RuneCheckMark, status.Track),
					fmt.Sprintf("%s %c %s", status.JoinArtists(), RuneEmDash, status.Album),
					uint32(0),
				)
			}

			scrobbledPrevious[player] = true

			for _, provider := range config.Providers() {
				sendScrobble(player, provider, status, config)
			}
		}
	}
}

func sendNowPlaying(player string,
	provider Provider,
	status Info,
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
			sendNotification(
				fmt.Sprintf("%c error updating now playing status (%s)", RuneWarningSign, provider.Name()),
				fmt.Sprintf("error updating now playing status: <b>%s</b>", err.Error()),
				uint32(0),
			)
		}
	} else {
		log.Debug().
			Str("player", player).
			Str("provider", provider.Name()).
			Interface("status", status).
			Msg("updated now playing status")
	}
}

func sendScrobble(player string,
	provider Provider,
	status Info,
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
			sendNotification(
				fmt.Sprintf("%c error saving scrobble (%s)", RuneWarningSign, provider.Name()),
				fmt.Sprintf("error saving scrobble: <b>%s</b>", err.Error()),
				uint32(0),
			)
		}
	} else {
		log.Info().
			Str("player", player).
			Str("provider", provider.Name()).
			Interface("status", status).
			Msg("saved scrobble")
	}
}

func minPlayTime(playbackInfo Info, config *Config) (int64, error) {
	if playbackInfo.Duration < 0 {
		return 0, fmt.Errorf("invalid track length: %d", playbackInfo.Duration)
	}

	half := int64((float64(playbackInfo.Duration) / 100) * float64(config.MinPlaybackPercent))
	return min(half, config.MinPlaybackDuration), nil
}

func sendNotification(summary, body string, replacesID uint32) uint32 {
	id, err := SendNotification(replacesID, summary, body)
	if err != nil {
		log.Error().
			Err(err).
			Interface("notification", map[string]string{
				"summary": summary,
				"text":    body},
			).
			Msg("error sending desktop notification via dbus")
		return replacesID
	}
	return id
}
