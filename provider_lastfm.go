package main

import (
	"errors"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/shkh/lastfm-go/lastfm"
)

var lastFm *lastfm.Api

func (l *LastFmConfig) Name() string {
	return "last.fm"
}

func (l *LastFmConfig) NowPlaying(n NowPlaying) {
	if l == nil {
		return
	}

	if l.SessionKey == "" {
		log.Error().Msg("last.fm provider is configured, but not authenticated")
		return
	}

	if err := l.CreateSession(); err != nil {
		log.Error().Err(err).Msg("failed to create last.fm session")
		return
	}

	// https://www.last.fm/api/show/track.updateNowPlaying
	if _, err := lastFm.Track.UpdateNowPlaying(lastfm.P{
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

	if l.SessionKey == "" {
		log.Error().Msg("last.fm provider is configured, but not authenticated")
		return
	}

	if err := l.CreateSession(); err != nil {
		log.Error().Err(err).Msg("failed to create last.fm session")
		return
	}

	// https://www.last.fm/api/show/track.scrobble
	if _, err := lastFm.Track.Scrobble(lastfm.P{
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

func (l *LastFmConfig) CreateSession() error {
	if lastFm != nil {
		log.Debug().Msg("last.fm session already created")
	}

	if len(l.SessionKey) != 32 {
		return errors.New("invalid session key length")
	}

	log.Debug().Msg("creating last.fm session")

	api := lastfm.New(l.Key, l.Secret)
	api.SetSession(l.SessionKey)

	lastFm = api

	return nil
}
