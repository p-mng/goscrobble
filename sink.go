package main

type Sink interface {
	Name() string
	NowPlaying(PlaybackStatus) error
	Scrobble(PlaybackStatus) error
}
