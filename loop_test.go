package main_test

import (
	"regexp"
	"testing"
	"time"

	main "github.com/p-mng/goscrobble"
	"github.com/stretchr/testify/require"
)

func TestMainLoop(t *testing.T) {
	previouslyPlaying := map[string]main.PlaybackStatus{}
	scrobbledPrevious := map[string]bool{}

	playerBlacklist := []*regexp.Regexp{}
	parsedRegexes := []main.ParsedRegexReplace{}

	fakeSource := &FakeSource{
		Empty:          true,
		Error:          false,
		PlaybackStatus: defaultPlaybackStatus,
	}
	sources := []main.Source{fakeSource}

	fakeSink := &FakeSink{}
	sinks := []main.Sink{fakeSink}

	minPlaybackDuration := 4 * 60
	minPlaybackPercent := 50

	notifyOnScrobble := true
	notifyOnError := true

	mockNotifier := FakeNotifier{}

	runLoop := func() {
		main.RunMainLoopOnce(
			previouslyPlaying,
			scrobbledPrevious,
			playerBlacklist,
			parsedRegexes,
			sources,
			sinks,
			minPlaybackDuration,
			minPlaybackPercent,
			notifyOnScrobble,
			notifyOnError,
			mockNotifier.SendNotification,
		)
	}

	runLoop()
	require.Len(t, fakeSink.NowPlayingLog, 0)
	require.Len(t, fakeSink.ScrobbleLog, 0)
	require.Equal(t, mockNotifier.Notifications, 0)

	fakeSource.Empty = false

	runLoop()
	require.Len(t, fakeSink.NowPlayingLog, 1)
	require.Len(t, fakeSink.ScrobbleLog, 0)
	require.Equal(t, mockNotifier.Notifications, 1)

	fakeSource.PlaybackStatus.Position = time.Duration(time.Second * 241)

	runLoop()
	require.Len(t, fakeSink.NowPlayingLog, 1)
	require.Len(t, fakeSink.ScrobbleLog, 1)
	require.Equal(t, mockNotifier.Notifications, 2)

	newPlaybackStatus := main.PlaybackStatus{
		Artists:   []string{"Placebo"},
		Track:     "Every You Every Me",
		Album:     "Without You I'm Nothing",
		Duration:  time.Duration(time.Minute*3 + time.Second*34),
		Timestamp: defaultPlaybackStatus.Timestamp.Add(defaultPlaybackStatus.Duration),
		State:     main.PlaybackPlaying,
		Position:  time.Duration(0),
	}

	fakeSource.PlaybackStatus = newPlaybackStatus

	runLoop()
	require.Len(t, fakeSink.NowPlayingLog, 2)
	require.Len(t, fakeSink.ScrobbleLog, 1)
	require.Equal(t, mockNotifier.Notifications, 3)

	fakeSource.PlaybackStatus.Position = time.Duration(time.Minute * 2)

	runLoop()
	require.Len(t, fakeSink.NowPlayingLog, 2)
	require.Len(t, fakeSink.ScrobbleLog, 2)
	require.Equal(t, mockNotifier.Notifications, 4)

	fakeSource.PlaybackStatus = defaultPlaybackStatus
	fakeSink.Error = true

	runLoop()
	require.Len(t, fakeSink.NowPlayingLog, 2)
	require.Len(t, fakeSink.ScrobbleLog, 2)
	require.Equal(t, mockNotifier.Notifications, 6)

	fakeSource.Empty = true
	fakeSink.Error = false

	runLoop()
	require.Len(t, fakeSink.NowPlayingLog, 2)
	require.Len(t, fakeSink.ScrobbleLog, 2)
	require.Equal(t, mockNotifier.Notifications, 6)
}

func TestCompilePlayerBlacklist(t *testing.T) {
	blacklist := []string{"[", "test"}
	compiled := main.CompilePlayerBlacklist(blacklist)

	require.Len(t, compiled, 1)
	require.Equal(t, regexp.MustCompile("test"), compiled[0])
}

func TestSendNowPlaying(t *testing.T) {
	mockSink := FakeSink{}
	mockNotifier := FakeNotifier{}

	main.SendNowPlaying(
		"mock player",
		&mockSink,
		defaultPlaybackStatus,
		true,
		mockNotifier.SendNotification,
	)
	require.Len(t, mockSink.NowPlayingLog, 1)
	require.Equal(t, 0, mockNotifier.Notifications)

	mockSink.Error = true

	main.SendNowPlaying(
		"mock player",
		&mockSink,
		defaultPlaybackStatus,
		true,
		mockNotifier.SendNotification,
	)
	require.Len(t, mockSink.NowPlayingLog, 1)
	require.Equal(t, 1, mockNotifier.Notifications)
}

func TestSendScrobble(t *testing.T) {
	mockSink := FakeSink{}
	mockNotifier := FakeNotifier{}

	main.SendScrobble(
		"mock player",
		&mockSink,
		defaultPlaybackStatus,
		true,
		mockNotifier.SendNotification,
	)
	require.Len(t, mockSink.ScrobbleLog, 1)
	require.Equal(t, 0, mockNotifier.Notifications)

	mockSink.Error = true

	main.SendScrobble(
		"mock player",
		&mockSink,
		defaultPlaybackStatus,
		true,
		mockNotifier.SendNotification,
	)
	require.Len(t, mockSink.ScrobbleLog, 1)
	require.Equal(t, 1, mockNotifier.Notifications)
}

func TestMinPlayTime(t *testing.T) {
	minPlaybackDuration := 4 * 60
	minPlaybackPercent := 50

	durations := map[time.Duration]time.Duration{
		time.Duration(time.Minute*17 + time.Second*46): time.Duration(time.Minute * 4),
		time.Duration(time.Minute*5 + time.Second*7):   time.Duration(time.Minute*2 + time.Second*33 + time.Millisecond*500),
	}

	for k, v := range durations {
		actual, err := main.MinPlayTime(k, minPlaybackDuration, minPlaybackPercent)
		require.NoError(t, err)
		require.Equal(t, v, actual)
	}

	_, err := main.MinPlayTime(time.Duration(-time.Second), minPlaybackDuration, minPlaybackPercent)
	require.Error(t, err)
}
