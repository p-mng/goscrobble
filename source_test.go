package main_test

import (
	"errors"
	"regexp"

	main "github.com/p-mng/goscrobble"
)

type FakeSource struct {
	Empty          bool
	Error          bool
	PlaybackStatus main.PlaybackStatus
}

func (m FakeSource) Name() string {
	return "fake source"
}

func (m FakeSource) GetInfo(
	_ []*regexp.Regexp,
	_ []main.ParsedRegexReplace,
) (map[string]main.PlaybackStatus, error) {
	if m.Empty {
		return map[string]main.PlaybackStatus{}, nil
	}

	status := map[string]main.PlaybackStatus{"fake player": m.PlaybackStatus}

	if m.Error {
		return status, errors.New("fake error")
	}
	return status, nil
}
