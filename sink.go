package main

import "time"

type Sink interface {
	Name() string
	NowPlaying(Scrobble) error
	Scrobble(Scrobble) error
	GetScrobbles(limit int, from, to time.Time) ([]Scrobble, error)
}
