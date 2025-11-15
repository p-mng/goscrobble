package main_test

import (
	"testing"
	"time"

	main "github.com/p-mng/goscrobble"
	"github.com/stretchr/testify/require"
)

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
