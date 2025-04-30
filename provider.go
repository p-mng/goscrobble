package main

type Provider interface {
	Name() string
	NowPlaying(NowPlaying)
	Scrobble(NowPlaying)
}
