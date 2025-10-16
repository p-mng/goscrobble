package nowplaying

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
func GetNowPlaying(
	playerBlacklist []*regexp.Regexp,
	regexes []ParsedRegexEntry,
) (map[string]NowPlayingInfo, error) {
	log.Debug().Msg("getting media metadata using `media-control`")

	cmd := exec.Command("/usr/bin/env", "media-control", "get")
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var info mediaControlInfo
	if err := json.Unmarshal(output, &info); err != nil {
		return nil, err
	}

	if isBlacklisted(playerBlacklist, info.BundleIdentifier) {
		return map[string]NowPlayingInfo{}, nil
	}

	nowPlaying := NowPlayingInfo{
		Artists:        []string{info.Artist},
		Track:          info.Title,
		Album:          info.Album,
		Duration:       int64(info.Duration),
		Timestamp:      info.Timestamp.Unix(),
		PlaybackStatus: playbackStatus(info.Playing),
		Position:       int64(info.ElapsedTime),
	}

	nowPlaying.RegexReplace(regexes)

	return map[string]NowPlayingInfo{
		info.BundleIdentifier: nowPlaying,
	}, nil
}

func playbackStatus(playing bool) string {
	if playing {
		return PlaybackPlaying
	}
	return PlaybackStopped
}
