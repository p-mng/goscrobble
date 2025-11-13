package main_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/jinzhu/copier"
	"github.com/p-mng/goscrobble"
	"github.com/stretchr/testify/require"
)

var defaultPlaybackStatus = main.PlaybackStatus{
	Artists:   []string{"Placebo", "David Bowie"},
	Track:     "Without You I'm Nothing",
	Album:     "A Place For Us To Dream",
	Duration:  time.Duration(time.Second * 251),
	Timestamp: time.Unix(1699225080, 0),
	State:     main.PlaybackPlaying,
	Position:  time.Duration(time.Second * 110),
}

func TestPlaybackStatusJoinArtists(t *testing.T) {
	require.Equal(t, "Placebo, David Bowie", defaultPlaybackStatus.JoinArtists())
}

func TestPlaybackStatusEquals(t *testing.T) {
	copied := main.PlaybackStatus{}
	err := copier.Copy(&copied, &defaultPlaybackStatus)
	require.NoError(t, err)

	require.True(t, defaultPlaybackStatus.Equals(copied))

	copied.Album = "Meds"
	require.False(t, defaultPlaybackStatus.Equals(copied))
}

func TestPlaybackStatusIsValid(t *testing.T) {
	copied := main.PlaybackStatus{}
	err := copier.Copy(&copied, &defaultPlaybackStatus)
	require.NoError(t, err)

	require.True(t, defaultPlaybackStatus.IsValid())

	copied.Album = ""
	require.False(t, copied.IsValid())
}

func TestPlaybackStatusRegexReplace(t *testing.T) {
	copied := main.PlaybackStatus{}
	err := copier.Copy(&copied, &defaultPlaybackStatus)
	require.NoError(t, err)

	require.Equal(t, "Without You I'm Nothing", defaultPlaybackStatus.Track)

	copied.RegexReplace([]main.ParsedRegexReplace{
		{
			Match:   regexp.MustCompile("^Without You"),
			Replace: "With You",
			Artist:  false,
			Track:   true,
			Album:   false,
		},
	})
	require.Equal(t, "With You I'm Nothing", copied.Track)
}

func TestIsBlacklisted(t *testing.T) {
	blacklist := []*regexp.Regexp{
		regexp.MustCompile("firefox"),
		regexp.MustCompile("chromium"),
	}
	expected := map[string]bool{
		"org.mpris.MediaPlayer2.chromium.instance10670": true,
		"org.mpris.MediaPlayer2.firefox.instance_1_84":  true,
		"org.mozilla.firefox":                           true,
		"com.tidal.desktop":                             false,
	}

	for k, v := range expected {
		require.Equal(t, main.IsBlacklisted(blacklist, k), v)
	}
}
