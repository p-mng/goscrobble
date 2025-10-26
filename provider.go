package main

type Provider interface {
	Name() string
	NowPlaying(Info) error
	Scrobble(Info) error
}
