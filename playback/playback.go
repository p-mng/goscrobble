package playback

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
)

// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Enum:Playback_Status
const (
	PlaybackPlaying = "Playing"
	PlaybackPaused  = "Paused"
	PlaybackStopped = "Stopped"
)

type Info struct {
	Artists   []string
	Track     string
	Album     string
	Duration  int64
	Timestamp int64

	PlaybackStatus string
	Position       int64
}

func (p Info) JoinArtists() string {
	return strings.Join(p.Artists, ", ")
}

type ParsedRegexEntry struct {
	Match   *regexp.Regexp
	Replace string
	Artist  bool
	Track   bool
	Album   bool
}

func (p Info) Equals(other Info) bool {
	return reflect.DeepEqual(p.Artists, other.Artists) &&
		p.Track == other.Track &&
		p.Album == other.Album
}

func (p Info) Valid() bool {
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

func (p *Info) RegexReplace(regexes []ParsedRegexEntry) {
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

func isBlacklisted(blacklist []*regexp.Regexp, player string) bool {
	for _, re := range blacklist {
		if re.MatchString(player) {
			return true
		}
	}
	return false
}
