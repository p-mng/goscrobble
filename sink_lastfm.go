package main

import (
	"errors"
	"strings"
	"time"

	lastfm "github.com/p-mng/lastfm-go"
	"github.com/rs/zerolog/log"
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

func (s LastFmSink) NowPlaying(scrobble Scrobble) error {
	_, err := s.Client.TrackUpdateNowPlaying(lastfm.P{
		"artist":   scrobble.JoinArtists(),
		"track":    scrobble.Track,
		"album":    scrobble.Album,
		"duration": max(int(scrobble.Duration.Seconds()), 30),
		"sk":       s.SessionKey,
	})
	return err
}

func (s LastFmSink) Scrobble(scrobble Scrobble) error {
	_, err := s.Client.TrackScrobble(lastfm.P{
		"artist":    scrobble.JoinArtists(),
		"track":     scrobble.Track,
		"album":     scrobble.Album,
		"duration":  max(int(scrobble.Duration.Seconds()), 30),
		"timestamp": scrobble.Timestamp.Unix(),
		"sk":        s.SessionKey,
	})
	return err
}

func (s LastFmSink) GetScrobbles(limit int, from, to time.Time) ([]Scrobble, error) {
	currentPage := 1
	totalPages := int64(1)

	var scrobbles []Scrobble

outer:
	for {
		params := lastfm.P{
			"limit":    min(limit, 200),
			"user":     s.Username,
			"page":     currentPage,
			"from":     from.Unix(),
			"extended": 1,
			"to":       to.Unix(),
		}

		log.Debug().Int("current page", currentPage).Interface("params", params).Msg("fetching page from last.fm API")
		page, err := s.Client.UserGetRecentTracks(params)
		if err != nil {
			return nil, err
		}

		if page.RecentTracks.TotalPages != totalPages {
			log.Debug().Int64("total pages", totalPages).Msg("updated number of total pages")
			totalPages = page.RecentTracks.TotalPages
		}

		log.Debug().Int("length", len(page.RecentTracks.Tracks)).Msg("converting scrobbles")

		for _, track := range page.RecentTracks.Tracks {
			scrobbles = append(scrobbles, Scrobble{
				// FIXME: this does not work in some cases (e.g., "Tyler, the Creator")
				Artists:   strings.Split(track.Artist.Name, ", "),
				Track:     track.Name,
				Album:     track.Album.Name,
				Duration:  time.Duration(0),
				Timestamp: time.Unix(track.Date.UTS, 0),
			})
			if len(scrobbles) >= limit {
				log.Debug().Int("limit", limit).Msg("reached limit")
				break outer
			}
		}

		if int64(currentPage) >= totalPages {
			log.Debug().Msg("reached last page")
			break outer
		}
		currentPage++
	}

	return scrobbles, nil
}
