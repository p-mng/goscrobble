package main

type Provider interface {
	Name() string
	NowPlaying(PlaybackStatus) error
	Scrobble(PlaybackStatus) error
}
