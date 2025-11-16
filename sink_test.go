package main_test

import (
	"errors"
	"time"

	main "github.com/p-mng/goscrobble"
)

type FakeSink struct {
	NowPlayingLog []main.Scrobble
	ScrobbleLog   []main.Scrobble
	Error         bool
}

func (*FakeSink) Name() string {
	return "fake sink"
}

func (s *FakeSink) NowPlaying(p main.Scrobble) error {
	if s.Error {
		return errors.New("fake error")
	}
	s.NowPlayingLog = append(s.NowPlayingLog, p)
	return nil
}

func (s *FakeSink) Scrobble(p main.Scrobble) error {
	if s.Error {
		return errors.New("fake error")
	}
	s.ScrobbleLog = append(s.ScrobbleLog, p)
	return nil
}

func (s *FakeSink) GetScrobbles(_ int, _, _ time.Time) ([]main.Scrobble, error) {
	return []main.Scrobble{}, nil
}
