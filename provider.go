package main

type Provider interface {
	Name() string
	NowPlaying(NowPlaying) error
	Scrobble(NowPlaying) error
}
