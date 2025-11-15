package main_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/jinzhu/copier"
	main "github.com/p-mng/goscrobble"
	"github.com/stretchr/testify/require"
)

var (
	defaultScrobble = main.Scrobble{
		Artists:   []string{"Placebo", "David Bowie"},
		Track:     "Without You I'm Nothing",
		Album:     "A Place For Us To Dream",
		Duration:  time.Duration(time.Second * 251),
		Timestamp: time.Unix(1699225080, 0),
	}
	defaultPlaybackStatus = main.PlaybackStatus{
		Scrobble: defaultScrobble,
		State:    main.PlaybackPlaying,
		Position: time.Duration(time.Second * 110),
	}
)

func TestScrobbleJoinArtists(t *testing.T) {
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

func TestScrobbleIsValid(t *testing.T) {
	copied := main.Scrobble{}
	err := copier.Copy(&copied, &defaultScrobble)
	require.NoError(t, err)

	require.True(t, defaultScrobble.IsValid())

	copied.Album = ""
	require.False(t, copied.IsValid())
}

func TestScrobbleRegexReplace(t *testing.T) {
	copied := main.Scrobble{}
	err := copier.Copy(&copied, &defaultScrobble)
	require.NoError(t, err)

	copied.RegexReplace([]main.ParsedRegexReplace{
		{
			Match:   regexp.MustCompile("David Bowie"),
			Replace: "DavidBowie",
			Artist:  true,
			Track:   false,
			Album:   false,
		},
		{
			Match:   regexp.MustCompile("^Without You"),
			Replace: "With You",
			Artist:  false,
			Track:   true,
			Album:   true,
		},
	})
	require.Equal(t, "DavidBowie", copied.Artists[1])
	require.Equal(t, "With You I'm Nothing", copied.Track)
	require.Equal(t, "A Place For Us To Dream", copied.Album)
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
		t.Run(k, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, main.IsBlacklisted(blacklist, k), v)
		})
	}
}
