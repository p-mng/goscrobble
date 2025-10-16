package main

type Provider interface {
	Name() string
	NowPlaying(NowPlayingInfo) error
	Scrobble(NowPlayingInfo) error
}
