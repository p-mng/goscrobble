package main_test

import (
	"errors"

	main "github.com/p-mng/goscrobble"
)

type FakeSink struct {
	NowPlayingLog []main.PlaybackStatus
	ScrobbleLog   []main.PlaybackStatus
	Error         bool
}

func (*FakeSink) Name() string {
	return "fake sink"
}

func (s *FakeSink) NowPlaying(p main.PlaybackStatus) error {
	if s.Error {
		return errors.New("fake error")
	}
	s.NowPlayingLog = append(s.NowPlayingLog, p)
	return nil
}

func (s *FakeSink) Scrobble(p main.PlaybackStatus) error {
	if s.Error {
		return errors.New("fake error")
	}
	s.ScrobbleLog = append(s.ScrobbleLog, p)
	return nil
}
