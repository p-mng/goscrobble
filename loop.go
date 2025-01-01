package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
)

const PollFrequency = 2000

func RunMainLoop(conn *dbus.Conn, config *Config) {
	previouslyPlaying := map[string]NowPlaying{}
	scrobbledPrevious := map[string]bool{}

	var blacklist []*regexp.Regexp

	for _, expression := range config.Blacklist {
		compiled, err := regexp.Compile(expression)
		if err != nil {
			log.Printf("error compiling regex blacklist entry: %v", err)
			continue
		}
		blacklist = append(blacklist, compiled)
	}

	if len(blacklist) > 0 {
		log.Printf("ignoring scrobbles from %d players", len(blacklist))
	}

	for {
		nowPlaying, err := GetNowPlaying(conn, blacklist)
		if err != nil {
			log.Printf("failed to get current playback status: %v", err)
			continue
		}

		for player := range nowPlaying {
			if _, ok := previouslyPlaying[player]; !ok {
				log.Printf("new player found: %s", player)
				previouslyPlaying[player] = NowPlaying{}
				scrobbledPrevious[player] = false
			}
		}

		for player := range previouslyPlaying {
			if _, ok := nowPlaying[player]; !ok {
				log.Printf("player disappeared: %s", player)
				delete(previouslyPlaying, player)
				delete(scrobbledPrevious, player)
			}
		}

		for player, status := range nowPlaying {
			minPlayTime, err := minPlayTime(status, config)
			if err != nil {
				log.Printf(
					"[%s] cannot calculate minimum playback time for %s by %s: %v",
					player,
					status.Track,
					strings.Join(status.Artists, ", "),
					err,
				)
				continue
			}

			if !NowPlayingEquals(status, previouslyPlaying[player]) {
				log.Printf("[%s] started playing %s by %s", player, status.Track, strings.Join(status.Artists, ", "))
				status.Position = 0
				previouslyPlaying[player] = status
				scrobbledPrevious[player] = false

				for _, provider := range config.Providers() {
					provider.NowPlaying(status)
				}

				continue
			}

			if status.Position < minPlayTime || status.PlaybackStatus != PlaybackPlaying || scrobbledPrevious[player] {
				continue
			}

			log.Printf(
				"[%s] scrobbling track %s by %s, played %s/%s",
				player,
				status.Track,
				strings.Join(status.Artists, ", "),
				formatDuration(minPlayTime),
				formatDuration(status.Duration),
			)
			scrobbledPrevious[player] = true

			for _, provider := range config.Providers() {
				provider.NowPlaying(status)
				provider.Scrobble(status)
			}
		}

		time.Sleep(time.Millisecond * PollFrequency)
	}
}

func minPlayTime(nowPlaying NowPlaying, config *Config) (int64, error) {
	if nowPlaying.Duration < 0 {
		return 0, fmt.Errorf("invalid track length: %d", nowPlaying.Duration)
	}

	half := int64((float64(nowPlaying.Duration) / 100) * float64(config.MinPlaybackPercent))
	return min(half, config.MinPlaybackDuration), nil
}

func formatDuration(duration int64) string {
	if duration < 0 {
		return "invalid duration"
	}

	minutes := duration / 60
	seconds := duration % 60

	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}
