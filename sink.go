package main

type Sink interface {
	Name() string
	NowPlaying(Scrobble) error
	Scrobble(Scrobble) error
}
