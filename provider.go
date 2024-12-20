package main

type Provider interface {
	NowPlaying(NowPlaying)
	Scrobble(NowPlaying)
}
