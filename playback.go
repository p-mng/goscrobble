package main

import (
	"encoding/csv"
	"errors"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type PlaybackState string

// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Enum:Playback_Status
const (
	PlaybackPlaying = PlaybackState("Playing")
	PlaybackPaused  = PlaybackState("Paused")
	PlaybackStopped = PlaybackState("Stopped")
)

type Scrobble struct {
	Artists   []string
	Track     string
	Album     string
	Duration  time.Duration
	Timestamp time.Time
}

type PlaybackStatus struct {
	Scrobble
	State    PlaybackState
	Position time.Duration
}

type ParsedRegexReplace struct {
	Match   *regexp.Regexp
	Replace string
	Artist  bool
	Track   bool
	Album   bool
}

func (s Scrobble) JoinArtists() string {
	return strings.Join(s.Artists, ", ")
}

func (p PlaybackStatus) Equals(other PlaybackStatus) bool {
	return reflect.DeepEqual(p.Artists, other.Artists) &&
		p.Track == other.Track &&
		p.Album == other.Album &&
		p.Duration == other.Duration
}

func (s Scrobble) IsValid() bool {
	switch {
	case s.JoinArtists() == "":
		return false
	case s.Track == "":
		return false
	case s.Album == "":
		return false
	case s.Duration == 0:
		return false
	default:
		return true
	}
}

func (s *Scrobble) RegexReplace(regexes []ParsedRegexReplace) {
	for _, r := range regexes {
		log.Debug().
			Str("expression", r.Match.String()).
			Str("replacement", r.Replace).
			Msg("running match/replace substitution")

		if r.Artist {
			var newArtists []string
			for _, artist := range s.Artists {
				newArtist := r.Match.ReplaceAllString(artist, r.Replace)
				newArtists = append(newArtists, newArtist)
			}
			s.Artists = newArtists
		}
		if r.Track {
			s.Track = r.Match.ReplaceAllString(s.Track, r.Replace)
		}
		if r.Album {
			s.Album = r.Match.ReplaceAllString(s.Album, r.Replace)
		}
	}
}

func (s Scrobble) ToStringSlice() []string {
	return []string{
		s.JoinArtists(),
		s.Track,
		s.Album,
		strconv.FormatInt(s.Duration.Milliseconds(), 10),
		s.Timestamp.Format(time.RFC1123),
	}
}

func ScrobbleFromCSV(input string) (Scrobble, error) {
	if strings.ContainsRune(input, '\n') {
		return Scrobble{}, errors.New("input must be a single line")
	}

	inputReader := strings.NewReader(input)
	csvReader := csv.NewReader(inputReader)

	parts, err := csvReader.Read()
	if err != nil {
		return Scrobble{}, err
	}

	if len(parts) != 5 {
		return Scrobble{}, errors.New("input has invalid number of columns")
	}

	millis, err := strconv.ParseInt(parts[3], 10, 64)
	if err != nil {
		return Scrobble{}, err
	}
	duration := time.Millisecond * time.Duration(millis)

	timestamp, err := time.Parse(time.RFC1123, parts[4])
	if err != nil {
		return Scrobble{}, err
	}

	return Scrobble{
		// FIXME: this does not work in some cases (e.g., "Tyler, the Creator")
		Artists:   strings.Split(parts[0], ", "),
		Track:     parts[1],
		Album:     parts[2],
		Duration:  duration,
		Timestamp: timestamp.In(time.Local),
	}, nil
}

func IsBlacklisted(blacklist []*regexp.Regexp, player string) bool {
	for _, re := range blacklist {
		if re.MatchString(player) {
			return true
		}
	}
	return false
}
