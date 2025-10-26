package main

import "regexp"

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
	return nil, nil
}
