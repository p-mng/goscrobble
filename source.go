package main

import "regexp"

type Source interface {
	Name() string
	GetInfo(
		playerBlacklist []*regexp.Regexp,
		regexes []ParsedRegexReplace,
	) (map[string]PlaybackStatus, error)
}
