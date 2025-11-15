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
	log.Debug().Msg("starting main loop")

	previouslyPlaying := map[string]PlaybackStatus{}
	scrobbledPrevious := map[string]bool{}

	playerBlacklist := CompilePlayerBlacklist(config.Blacklist)
	parsedRegexes := config.ParseRegexes()

	sources := config.SetupSources()
	sinks := config.SetupSinks()

	ticker := time.NewTicker(time.Second * time.Duration(config.PollRate))

	for {
		RunMainLoopIteration(
			previouslyPlaying,
			scrobbledPrevious,
			playerBlacklist,
			parsedRegexes,
			sources,
			sinks,
			config.MinPlaybackDuration,
			config.MinPlaybackPercent,
			config.NotifyOnScrobble,
			config.NotifyOnError,
			SendNotification,
		)

		timestamp := <-ticker.C
		log.Debug().Time("timestamp", timestamp).Msg("finished main loop iteration")
	}
}

func RunMainLoopIteration(
	previouslyPlaying map[string]PlaybackStatus,
	scrobbledPrevious map[string]bool,
	playerBlacklist []*regexp.Regexp,
	parsedRegexes []ParsedRegexReplace,
	sources []Source,
	sinks []Sink,
	minPlaybackDuration int,
	minPlaybackPercent int,
	notifyOnScrobble bool,
	notifyOnError bool,
	notifier NotifierFunc,
) {
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

		minPlayTime, err := MinPlayTime(
			status.Duration,
			minPlaybackDuration,
			minPlaybackPercent,
		)
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

			if notifyOnScrobble {
				newID, err := notifier(
					nowPlayingNotificationID,
					fmt.Sprintf("%c now playing: %s", RuneBeamedSixteenthNotes, status.Track),
					fmt.Sprintf("%s %c %s", status.JoinArtists(), RuneEmDash, status.Album),
				)
				if err != nil {
					log.Error().Err(err).Msg("error sending notification")
				} else {
					nowPlayingNotificationID = newID
				}
			}

			for _, sink := range sinks {
				SendNowPlaying(player, sink, status, notifyOnError, notifier)
			}

			continue
		}

		status.Timestamp = previouslyPlaying[player].Timestamp

		if status.Position < minPlayTime || status.State != PlaybackPlaying || scrobbledPrevious[player] {
			continue
		}

		log.Info().
			Str("player", player).
			Interface("status", status).
			Msg("scrobbling track")

		if notifyOnScrobble {
			if _, err := notifier(
				uint32(0),
				fmt.Sprintf("%c scrobbling: %s", RuneCheckMark, status.Track),
				fmt.Sprintf("%s %c %s", status.JoinArtists(), RuneEmDash, status.Album),
			); err != nil {
				log.Error().Err(err).Msg("error sending notification")
			}
		}

		scrobbledPrevious[player] = true

		for _, sink := range sinks {
			SendScrobble(player, sink, status, notifyOnError, notifier)
		}
	}
}

func CompilePlayerBlacklist(blacklist []string) []*regexp.Regexp {
	var playerBlacklist []*regexp.Regexp

	for _, expression := range blacklist {
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

	return playerBlacklist
}

func SendNowPlaying(player string,
	sink Sink,
	status PlaybackStatus,
	notifyOnError bool,
	notifier NotifierFunc,
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

		if notifyOnError {
			if _, err := notifier(
				uint32(0),
				fmt.Sprintf("%c error updating now playing status (%s)", RuneWarningSign, sink.Name()),
				fmt.Sprintf("error updating now playing status: %s", err.Error()),
			); err != nil {
				log.Error().Err(err).Msg("error sending notification")
			}
		}
	} else {
		log.Debug().
			Str("player", player).
			Str("sink", sink.Name()).
			Interface("status", status).
			Msg("updated now playing status")
	}
}

func SendScrobble(player string,
	sink Sink,
	status PlaybackStatus,
	notifyOnError bool,
	notifier NotifierFunc,
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

		if notifyOnError {
			if _, err := notifier(
				uint32(0),
				fmt.Sprintf("%c error saving scrobble (%s)", RuneWarningSign, sink.Name()),
				fmt.Sprintf("error saving scrobble: %s", err.Error()),
			); err != nil {
				log.Error().Err(err).Msg("error sending notification")
			}
		}
	} else {
		log.Info().
			Str("player", player).
			Str("sink", sink.Name()).
			Interface("status", status).
			Msg("saved scrobble")
	}
}

func MinPlayTime(
	duration time.Duration,
	minPlaybackDuration int,
	minPlaybackPercent int,
) (time.Duration, error) {
	if duration < time.Duration(0) {
		return 0, fmt.Errorf("invalid track length: %d", duration)
	}

	configDuration := time.Duration(minPlaybackDuration * int(time.Second))
	halfDuration := time.Duration(minPlaybackPercent * int(duration/100))

	return min(configDuration, halfDuration), nil
}
