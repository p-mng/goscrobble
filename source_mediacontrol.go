package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"time"

	"github.com/rs/zerolog/log"
)

type MediaControlSource struct {
	Command   string
	Arguments []string
}

func (s MediaControlSource) Name() string {
	return "media-control"
}

func (s MediaControlSource) GetInfo(
	playerBlacklist []*regexp.Regexp,
	regexes []ParsedRegexReplace,
) (map[string]PlaybackStatus, error) {
	log.Debug().Msg("getting playback metadata using media-control")

	//nolint:gosec
	cmd := exec.Command(s.Command, s.Arguments...)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var outputParsed MediaControlInfo
	if err := json.Unmarshal(output, &outputParsed); err != nil {
		return nil, err
	}

	if outputParsed == (MediaControlInfo{}) {
		log.Debug().Msg("media-control did not find any active players")
		return map[string]PlaybackStatus{}, nil
	}

	if IsBlacklisted(playerBlacklist, outputParsed.BundleIdentifier) {
		return map[string]PlaybackStatus{}, nil
	}

	var state PlaybackState
	if outputParsed.Playing {
		state = PlaybackPlaying
	} else {
		state = PlaybackStopped
	}

	playbackStatus := PlaybackStatus{
		Scrobble: Scrobble{
			Artists:   []string{outputParsed.Artist},
			Track:     outputParsed.Title,
			Album:     outputParsed.Album,
			Duration:  time.Duration(outputParsed.Duration * float64(time.Second)),
			Timestamp: outputParsed.Timestamp,
		},
		State:    state,
		Position: time.Duration(outputParsed.ElapsedTimeNow * float64(time.Second)),
	}

	playbackStatus.RegexReplace(regexes)

	playerName := fmt.Sprintf("%s:%s", s.Name(), outputParsed.BundleIdentifier)
	return map[string]PlaybackStatus{playerName: playbackStatus}, nil
}

// https://github.com/ungive/media-control
// https://github.com/ungive/mediaremote-adapter
type MediaControlInfo struct {
	PlaybackRate          int       `json:"playbackRate"`
	Album                 string    `json:"album"`
	ElapsedTimeNow        float64   `json:"elapsedTimeNow"`
	ElapsedTime           float64   `json:"elapsedTime"`
	Timestamp             time.Time `json:"timestamp"`
	BundleIdentifier      string    `json:"bundleIdentifier"`
	ProcessIdentifier     int       `json:"processIdentifier"`
	ArtworkData           string    `json:"artworkData"`
	Title                 string    `json:"title"`
	ArtworkMimeType       string    `json:"artworkMimeType"`
	Duration              float64   `json:"duration"`
	Artist                string    `json:"artist"`
	ContentItemIdentifier string    `json:"contentItemIdentifier"`
	Playing               bool      `json:"playing"`
}
