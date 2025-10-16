package main

import "github.com/p-mng/goscrobble/playback"

type Provider interface {
	Name() string
	NowPlaying(playback.Info) error
	Scrobble(playback.Info) error
}
