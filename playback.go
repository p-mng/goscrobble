package main

import (
	"reflect"
	"regexp"
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

func IsBlacklisted(blacklist []*regexp.Regexp, player string) bool {
	for _, re := range blacklist {
		if re.MatchString(player) {
			return true
		}
	}
	return false
}
