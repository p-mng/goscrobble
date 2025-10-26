package main

import (
	"fmt"
	"maps"
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

func RunMainLoop(config Config) {
	log.Debug().Msg("started main loop")

	previouslyPlaying := map[string]PlaybackStatus{}
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

	sources := config.GetSources()
	sinks := config.GetSinks()
	log.Debug().Msg("set up sources and sinks")

	if len(sources) == 0 {
		log.Warn().Msg("no sources configured")
	}
	if len(sinks) == 0 {
		log.Warn().Msg("no sinks configured")
	}

	for {
		log.Debug().
			Dur("duration", time.Duration(config.PollRate)).
			Msg("waiting for next poll")
		time.Sleep(time.Second * time.Duration(config.PollRate))

		playbackStatus := make(map[string]PlaybackStatus)

		for _, source := range sources {
			status, err := source.GetInfo(playerBlacklist, parsedRegexes)
			if err != nil {
				log.Error().Err(err).Str("source", source.Name()).Msg("error getting current playback status")
			}
			maps.Copy(playbackStatus, status)
		}

		for player := range playbackStatus {
			if _, ok := previouslyPlaying[player]; !ok {
				log.Info().
					Str("player", player).
					Msg("new player found")
				previouslyPlaying[player] = PlaybackStatus{}
				scrobbledPrevious[player] = false
			}
		}

		for player := range previouslyPlaying {
			if _, ok := playbackStatus[player]; !ok {
				log.Info().
					Str("player", player).
					Msg("player disappeared")
				delete(previouslyPlaying, player)
				delete(scrobbledPrevious, player)
			}
		}

		for player, status := range playbackStatus {
			if !status.IsValid() {
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
				status.Position = time.Duration(0)
				status.Timestamp = time.Now()

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

				for _, sink := range sinks {
					sendNowPlaying(player, sink, status, config)
				}

				continue
			}

			status.Timestamp = previouslyPlaying[player].Timestamp

			if status.Position < minPlayTime || status.Status != PlaybackPlaying || scrobbledPrevious[player] {
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

			for _, sink := range sinks {
				sendScrobble(player, sink, status, config)
			}
		}
	}
}

func sendNowPlaying(player string,
	sink Sink,
	status PlaybackStatus,
	config Config,
) {
	log.Debug().
		Str("player", player).
		Str("sink", sink.Name()).
		Interface("status", status).
		Msg("updating now playing status")
	if err := sink.NowPlaying(status); err != nil {
		log.Error().
			Str("player", player).
			Str("sink", sink.Name()).
			Interface("status", status).
			Err(err).
			Msg("error updating now playing status")

		if config.NotifyOnError {
			sendNotification(
				fmt.Sprintf("%c error updating now playing status (%s)", RuneWarningSign, sink.Name()),
				fmt.Sprintf("error updating now playing status: %s", err.Error()),
				uint32(0),
			)
		}
	} else {
		log.Debug().
			Str("player", player).
			Str("sink", sink.Name()).
			Interface("status", status).
			Msg("updated now playing status")
	}
}

func sendScrobble(player string,
	sink Sink,
	status PlaybackStatus,
	config Config,
) {
	log.Debug().
		Str("player", player).
		Str("sink", sink.Name()).
		Interface("status", status).
		Msg("saving scrobble")
	if err := sink.Scrobble(status); err != nil {
		log.Error().
			Str("player", player).
			Str("sink", sink.Name()).
			Interface("status", status).
			Err(err).
			Msg("error saving scrobble")

		if config.NotifyOnError {
			sendNotification(
				fmt.Sprintf("%c error saving scrobble (%s)", RuneWarningSign, sink.Name()),
				fmt.Sprintf("error saving scrobble: %s", err.Error()),
				uint32(0),
			)
		}
	} else {
		log.Info().
			Str("player", player).
			Str("sink", sink.Name()).
			Interface("status", status).
			Msg("saved scrobble")
	}
}

func minPlayTime(status PlaybackStatus, config Config) (time.Duration, error) {
	if status.Duration < time.Duration(0) {
		return 0, fmt.Errorf("invalid track length: %d", status.Duration)
	}

	configDuration := time.Duration(config.MinPlaybackDuration * int64(time.Second))
	halfDuration := time.Duration(int64(status.Duration/100) * config.MinPlaybackPercent)

	return min(configDuration, halfDuration), nil
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
