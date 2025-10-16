package nowplaying

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

type NowPlayingInfo struct {
	Artists   []string
	Track     string
	Album     string
	Duration  int64
	Timestamp int64

	PlaybackStatus string
	Position       int64
}

func (n NowPlayingInfo) JoinArtists() string {
	return strings.Join(n.Artists, ", ")
}

type ParsedRegexEntry struct {
	Match   *regexp.Regexp
	Replace string
	Artist  bool
	Track   bool
	Album   bool
}

func (n NowPlayingInfo) Equals(other NowPlayingInfo) bool {
	return reflect.DeepEqual(n.Artists, other.Artists) &&
		n.Track == other.Track &&
		n.Album == other.Album
}

func (n NowPlayingInfo) Valid() bool {
	switch {
	case n.Album == "":
		return false
	case n.Track == "":
		return false
	case n.JoinArtists() == "":
		return false
	case n.Duration == 0:
		return false
	default:
		return true
	}
}

func (n *NowPlayingInfo) RegexReplace(regexes []ParsedRegexEntry) {
	for _, r := range regexes {
		log.Debug().
			Str("expression", r.Match.String()).
			Str("replacement", r.Replace).
			Msg("running match/replace substitution")

		if r.Artist {
			var newArtists []string
			for _, artist := range n.Artists {
				newArtist := r.Match.ReplaceAllString(artist, r.Replace)
				newArtists = append(newArtists, newArtist)
			}
			n.Artists = newArtists
		}
		if r.Track {
			n.Track = r.Match.ReplaceAllString(n.Track, r.Replace)
		}
		if r.Album {
			n.Album = r.Match.ReplaceAllString(n.Album, r.Replace)
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
