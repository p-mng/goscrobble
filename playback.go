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

type PlaybackStatus struct {
	Artists   []string
	Track     string
	Album     string
	Duration  time.Duration
	Timestamp time.Time

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

func (p PlaybackStatus) JoinArtists() string {
	return strings.Join(p.Artists, ", ")
}

func (p PlaybackStatus) Equals(other PlaybackStatus) bool {
	return reflect.DeepEqual(p.Artists, other.Artists) &&
		p.Track == other.Track &&
		p.Album == other.Album
}

func (p PlaybackStatus) IsValid() bool {
	switch {
	case p.Album == "":
		return false
	case p.Track == "":
		return false
	case p.JoinArtists() == "":
		return false
	case p.Duration == 0:
		return false
	default:
		return true
	}
}

func (p *PlaybackStatus) RegexReplace(regexes []ParsedRegexReplace) {
	for _, r := range regexes {
		log.Debug().
			Str("expression", r.Match.String()).
			Str("replacement", r.Replace).
			Msg("running match/replace substitution")

		if r.Artist {
			var newArtists []string
			for _, artist := range p.Artists {
				newArtist := r.Match.ReplaceAllString(artist, r.Replace)
				newArtists = append(newArtists, newArtist)
			}
			p.Artists = newArtists
		}
		if r.Track {
			p.Track = r.Match.ReplaceAllString(p.Track, r.Replace)
		}
		if r.Album {
			p.Album = r.Match.ReplaceAllString(p.Album, r.Replace)
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
