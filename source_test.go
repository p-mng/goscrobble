package main_test

import (
	"errors"

	main "github.com/p-mng/goscrobble"
)

type FakeSource struct {
	PlayerName     string
	Empty          bool
	Error          bool
	PlaybackStatus main.PlaybackStatus
}

func (m FakeSource) Name() string {
	return "fake source"
}

func (m FakeSource) GetInfo() (map[string]main.PlaybackStatus, error) {
	if m.Empty {
		return map[string]main.PlaybackStatus{}, nil
	}

	var playerName string
	if m.PlayerName == "" {
		playerName = "fake player"
	} else {
		playerName = m.PlayerName
	}
	status := map[string]main.PlaybackStatus{playerName: m.PlaybackStatus}

	if m.Error {
		return status, errors.New("fake error")
	}
	return status, nil
}
