package main

import (
	"encoding/json"
	"os/exec"
	"regexp"
	"time"

	"github.com/rs/zerolog/log"
)

type mediaControlInfo struct {
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

// https://github.com/ungive/media-control
// https://github.com/ungive/mediaremote-adapter
func GetInfo(
	playerBlacklist []*regexp.Regexp,
	regexes []ParsedRegexReplace,
) (map[string]PlaybackStatus, error) {
	log.Debug().Msg("getting playback metadata using media-control")

	cmd := exec.Command("/usr/bin/env", "media-control", "get", "--now")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var outputParsed mediaControlInfo
	if err := json.Unmarshal(output, &outputParsed); err != nil {
		return nil, err
	}

	if outputParsed == (mediaControlInfo{}) {
		log.Debug().Msg("media-control did not find any active players")
		return map[string]PlaybackStatus{}, nil
	}

	if isBlacklisted(playerBlacklist, outputParsed.BundleIdentifier) {
		return map[string]PlaybackStatus{}, nil
	}

	playbackStatus := PlaybackStatus{
		Artists:        []string{outputParsed.Artist},
		Track:          outputParsed.Title,
		Album:          outputParsed.Album,
		Duration:       int64(outputParsed.Duration),
		Timestamp:      outputParsed.Timestamp.Unix(),
		PlaybackStatus: playbackStatus(outputParsed.Playing),
		Position:       int64(outputParsed.ElapsedTimeNow),
	}

	playbackStatus.RegexReplace(regexes)

	return map[string]PlaybackStatus{
		outputParsed.BundleIdentifier: playbackStatus,
	}, nil
}

func playbackStatus(playing bool) string {
	if playing {
		return PlaybackPlaying
	}
	return PlaybackStopped
}
