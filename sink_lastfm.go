package main

import (
	"errors"

	lastfm "github.com/p-mng/lastfm-go"
)

type LastFmSink struct {
	Client     lastfm.Client
	SessionKey string
	Username   string
}

func LastFmSinkFromConfig(c LastFmConfig) (LastFmSink, error) {
	var sink LastFmSink

	if c.SessionKey == "" || c.Username == "" {
		return sink, errors.New("last.fm sink is configured, but not authenticated")
	}

	client, err := lastfm.NewDesktopClient(lastfm.BaseURL, c.Key, c.Secret)
	if err != nil {
		return sink, err
	}

	return LastFmSink{Client: client, SessionKey: c.SessionKey, Username: c.Username}, nil
}

func (s LastFmSink) Name() string {
	return "last.fm"
}

func (s LastFmSink) NowPlaying(p PlaybackStatus) error {
	_, err := s.Client.TrackUpdateNowPlaying(lastfm.P{
		"artist":   p.JoinArtists(),
		"track":    p.Track,
		"album":    p.Album,
		"duration": max(int(p.Duration.Seconds()), 30),
		"sk":       s.SessionKey,
	})
	return err
}

func (s LastFmSink) Scrobble(p PlaybackStatus) error {
	_, err := s.Client.TrackScrobble(lastfm.P{
		"artist":    p.JoinArtists(),
		"track":     p.Track,
		"album":     p.Album,
		"duration":  max(int(p.Duration.Seconds()), 30),
		"timestamp": p.Timestamp.Unix(),
		"sk":        s.SessionKey,
	})
	return err
}
