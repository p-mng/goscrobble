package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const DefaultTidalHifiEndpoint = "http://localhost:47836/current"

// http://localhost:47836/docs/#/current/get_current
type TidalHifiAPIResponse struct {
	Title             string  `json:"title"`
	Artists           string  `json:"artists"`
	ArtistsArray      []any   `json:"artistsArray"`
	Album             string  `json:"album"`
	PlayingFrom       string  `json:"playingFrom"`
	Status            string  `json:"status"`
	URL               string  `json:"url"`
	Current           string  `json:"current"`
	CurrentInSeconds  float64 `json:"currentInSeconds"`
	Duration          string  `json:"duration"`
	DurationInSeconds int     `json:"durationInSeconds"`
	Image             string  `json:"image"`
	Icon              string  `json:"icon"`
	LocalAlbumArt     string  `json:"localAlbumArt"`
	Favorite          bool    `json:"favorite"`
	TrackID           string  `json:"trackId"`
	Volume            int     `json:"volume"`
	Player            struct {
		Status  string `json:"status"`
		Shuffle bool   `json:"shuffle"`
		Repeat  string `json:"repeat"`
	} `json:"player"`
	Artist string `json:"artist"`
}

type TidalHifiSource struct {
	Client   http.Client
	Endpoint string
}

func (s TidalHifiSource) Name() string {
	return "tidal-hifi"
}

func (s TidalHifiSource) GetInfo() (map[string]PlaybackStatus, error) {
	response, err := s.Client.Get(s.Endpoint)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var body TidalHifiAPIResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		return nil, err
	}

	var status PlaybackState
	switch body.Status {
	case "playing":
		status = PlaybackPlaying
	case "paused":
		status = PlaybackPaused
	default:
		return nil, errors.New("invalid playback status returned by API")
	}

	info := PlaybackStatus{
		Scrobble: Scrobble{
			Artists:   strings.Split(body.Artist, ", "),
			Track:     body.Title,
			Album:     body.Album,
			Duration:  time.Duration(body.DurationInSeconds * int(time.Second)),
			Timestamp: time.Time{},
		},
		State:    status,
		Position: time.Duration(body.CurrentInSeconds * float64(time.Second)),
	}
	playerName := fmt.Sprintf("%s:%s", s.Name(), s.Endpoint)

	return map[string]PlaybackStatus{
		playerName: info,
	}, nil
}
