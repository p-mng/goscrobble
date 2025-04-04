package main

import (
	"log"
	"strings"

	"github.com/shkh/lastfm-go/lastfm"
)

func (l *LastFmConfig) NowPlaying(n NowPlaying) {
	if l == nil {
		return
	}

	api := lastfm.New(l.Key, l.Secret)
	err := api.Login(l.Username, l.Password)
	if err != nil {
		log.Printf("[lastfm] login error: %v", err)
		return
	}

	// https://www.last.fm/api/show/track.updateNowPlaying
	if _, err := api.Track.UpdateNowPlaying(lastfm.P{
		"artist":   strings.Join(n.Artists, ", "),
		"track":    n.Track,
		"album":    n.Album,
		"duration": n.Duration,
	}); err != nil {
		log.Printf("[lastfm] error updating now playing status: %v", err)
		return
	}

	log.Println("[lastfm] updated now playing status ✅")
}

func (l *LastFmConfig) Scrobble(n NowPlaying) {
	if l == nil {
		return
	}

	api := lastfm.New(l.Key, l.Secret)
	err := api.Login(l.Username, l.Password)
	if err != nil {
		log.Printf("[lastfm] login error: %v", err)
		return
	}

	// https://www.last.fm/api/show/track.scrobble
	if _, err := api.Track.Scrobble(lastfm.P{
		"artist":    strings.Join(n.Artists, ", "),
		"track":     n.Track,
		"album":     n.Album,
		"duration":  Duration(n.Duration),
		"timestamp": n.Timestamp,
	}); err != nil {
		log.Printf("[lastfm] error scrobbling: %v", err)
		return
	}

	log.Println("[lastfm] scrobbled ✅")
}

func Duration(d int64) int64 {
	if d < 30 {
		return 30
	}
	return d
}
