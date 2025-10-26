package main

import (
	"errors"

	"github.com/rs/zerolog/log"
	"github.com/shkh/lastfm-go/lastfm"
)

const ErrLastFmNotAuthenticated = "last.fm sink is configured, but not authenticated"

var lastFm *lastfm.Api

func (l *LastFmConfig) Name() string {
	return "last.fm"
}

func (l *LastFmConfig) NowPlaying(p PlaybackStatus) error {
	if l.SessionKey == "" {
		return errors.New(ErrLastFmNotAuthenticated)
	}

	if err := l.CreateSession(); err != nil {
		log.Error().Err(err).Msg("failed to create last.fm session")
		return errors.New(ErrLastFmNotAuthenticated)
	}

	// https://www.last.fm/api/show/track.updateNowPlaying
	_, err := lastFm.Track.UpdateNowPlaying(lastfm.P{
		"artist":   p.JoinArtists(),
		"track":    p.Track,
		"album":    p.Album,
		"duration": max(int(p.Duration.Seconds()), 30),
	})
	return err
}

func (l *LastFmConfig) Scrobble(p PlaybackStatus) error {
	if l.SessionKey == "" {
		return errors.New(ErrLastFmNotAuthenticated)
	}

	if err := l.CreateSession(); err != nil {
		return err
	}

	// https://www.last.fm/api/show/track.scrobble
	_, err := lastFm.Track.Scrobble(lastfm.P{
		"artist":    p.JoinArtists(),
		"track":     p.Track,
		"album":     p.Album,
		"duration":  max(int(p.Duration.Seconds()), 30),
		"timestamp": p.Timestamp.Unix(),
	})
	return err
}

func (l *LastFmConfig) CreateSession() error {
	if lastFm != nil {
		log.Debug().Msg("last.fm session already created")
		return nil
	}

	if len(l.SessionKey) != 32 {
		return errors.New("invalid last.fm session key length")
	}

	log.Debug().Msg("creating last.fm session")

	api := lastfm.New(l.Key, l.Secret)
	api.SetSession(l.SessionKey)

	lastFm = api

	return nil
}
