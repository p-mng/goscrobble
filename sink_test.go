package main_test

import (
	"errors"

	main "github.com/p-mng/goscrobble"
)

type MockSink struct {
	NowPlayingLog []main.PlaybackStatus
	ScrobbleLog   []main.PlaybackStatus
	Error         bool
}

func (*MockSink) Name() string {
	return "mock-sink"
}

func (s *MockSink) NowPlaying(p main.PlaybackStatus) error {
	if s.Error {
		return errors.New("mock error")
	}
	s.NowPlayingLog = append(s.NowPlayingLog, p)
	return nil
}

func (s *MockSink) Scrobble(p main.PlaybackStatus) error {
	if s.Error {
		return errors.New("mock error")
	}
	s.ScrobbleLog = append(s.ScrobbleLog, p)
	return nil
}
