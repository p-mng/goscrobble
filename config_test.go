package main_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BurntSushi/toml"
	main "github.com/p-mng/goscrobble"
	"github.com/stretchr/testify/require"
)

func TestReadConfig(t *testing.T) {
	t.Run("read existing config", func(t *testing.T) {
		filename := filepath.Join(t.TempDir(), main.DefaultConfigFileName)

		//nolint:gosec
		file, err := os.Create(filename)
		require.NoError(t, err)
		defer main.CloseLogged(file)

		err = toml.NewEncoder(file).Encode(main.DefaultConfig)
		require.NoError(t, err)

		config, err := main.ReadConfig(filename)
		require.NoError(t, err)
		require.Equal(t, main.DefaultConfig, config)
	})
	t.Run("create new config", func(t *testing.T) {
		filename := filepath.Join(t.TempDir(), main.DefaultConfigFileName)

		config, err := main.ReadConfig(filename)
		require.NoError(t, err)
		require.Equal(t, main.DefaultConfig, config)
	})
	t.Run("create new config (subdirectory)", func(t *testing.T) {
		filename := filepath.Join(
			t.TempDir(),
			"subdirectory",
			main.DefaultConfigFileName,
		)

		config, err := main.ReadConfig(filename)
		require.NoError(t, err)
		require.Equal(t, main.DefaultConfig, config)
	})
}

func TestConfigValidate(t *testing.T) {
	//nolint:exhaustruct_v5
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
	filename := filepath.Join(t.TempDir(), main.DefaultConfigFileName)

	err := main.DefaultConfig.Write(filename)
	require.NoError(t, err)

	//nolint:gosec
	file, err := os.Open(filename)
	require.NoError(t, err)

	stat, err := file.Stat()
	require.NoError(t, err)

	require.Greater(t, stat.Size(), int64(100))
	require.False(t, stat.IsDir())
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
