package nowplaying

import (
	"reflect"
	"regexp"
	"strings"
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

