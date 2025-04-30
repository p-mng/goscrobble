package main

import (
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/shkh/lastfm-go/lastfm"
)

func (l *LastFmConfig) Name() string {
	return "last.fm"
}

func (l *LastFmConfig) NowPlaying(n NowPlaying) {
	if l == nil {
		return
	}

	api := lastfm.New(l.Key, l.Secret)
	err := api.Login(l.Username, l.Password)
	if err != nil {
		log.Error().
			Str("provider", l.Name()).
			Err(err).
			Msg("login error")
		return
	}

	// https://www.last.fm/api/show/track.updateNowPlaying
	if _, err := api.Track.UpdateNowPlaying(lastfm.P{
		"artist":   strings.Join(n.Artists, ", "),
		"track":    n.Track,
		"album":    n.Album,
		"duration": n.Duration,
	}); err != nil {
		log.Error().
			Str("provider", l.Name()).
			Err(err).
			Msg("error updating now playing status")
		return
	}

	log.Info().
		Str("provider", l.Name()).
		Interface("status", n).
		Msg("updated now playing status ✅")
}

func (l *LastFmConfig) Scrobble(n NowPlaying) {
	if l == nil {
		return
	}

	api := lastfm.New(l.Key, l.Secret)
	err := api.Login(l.Username, l.Password)
	if err != nil {
		log.Error().
			Str("provider", l.Name()).
			Err(err).
			Msg("login error")
		return
	}

	// https://www.last.fm/api/show/track.scrobble
	if _, err := api.Track.Scrobble(lastfm.P{
		"artist":    strings.Join(n.Artists, ", "),
		"track":     n.Track,
		"album":     n.Album,
		"duration":  max(n.Duration, 30),
		"timestamp": n.Timestamp,
	}); err != nil {
		log.Error().
			Str("provider", l.Name()).
			Err(err).
			Msg("error sending scrobble")
		return
	}

	log.Info().
		Str("provider", l.Name()).
		Interface("status", n).
		Msg("saved scrobble ✅")
}
