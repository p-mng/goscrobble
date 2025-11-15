package main_test

import (
	"fmt"
	"os"
	"testing"

	main "github.com/p-mng/goscrobble"
	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/require"
)

func TestParseConfig(t *testing.T) {
	data, err := toml.Marshal(main.DefaultConfig)
	require.NoError(t, err)

	config, err := main.ParseConfig(data)
	require.NoError(t, err)
	require.Equal(t, main.DefaultConfig, config)
}

func TestConfigValidate(t *testing.T) {
	//nolint:exhaustruct
	invalidConfig := main.Config{
		PollRate:            -20,
		MinPlaybackDuration: -20,
		MinPlaybackPercent:  200,
		// ...
	}
	invalidConfig.Validate()

	require.Equal(t, 2, invalidConfig.PollRate)
	require.Equal(t, 4*60, invalidConfig.MinPlaybackDuration)
	require.Equal(t, 50, invalidConfig.MinPlaybackPercent)
}

func TestConfigWrite(t *testing.T) {
	filename := fmt.Sprintf("%s/%s", t.TempDir(), main.DefaultConfigFileName)

	err := main.DefaultConfig.Write(filename)
	require.NoError(t, err)

	//nolint:gosec
	file, err := os.Open(filename)
	require.NoError(t, err)

	stat, err := file.Stat()
	require.NoError(t, err)

	require.Greater(t, stat.Size(), int64(100))
	require.False(t, stat.IsDir())
	require.Equal(t, "-rw-------", stat.Mode().String())
}

func TestConfigDir(t *testing.T) {
	t.Run("$XDG_CONFIG_HOME", func(t *testing.T) {
		t.Setenv("HOME", "/home/user")
		t.Setenv("XDG_CONFIG_HOME", "/home/user/my-config-dir")
		configDir := main.ConfigDir()
		require.Equal(t, "/home/user/my-config-dir/goscrobble", configDir)
	})
	t.Run("$HOME", func(t *testing.T) {
		t.Setenv("HOME", "/home/user")
		t.Setenv("XDG_CONFIG_HOME", "")
		configDir := main.ConfigDir()
		require.Equal(t, "/home/user/.config/goscrobble", configDir)
	})
}
