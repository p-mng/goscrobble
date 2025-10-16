package main

import np "github.com/p-mng/goscrobble/nowplaying"

type Provider interface {
	Name() string
	NowPlaying(np.NowPlayingInfo) error
	Scrobble(np.NowPlayingInfo) error
}
