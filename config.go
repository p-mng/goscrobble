package main

import (
	"fmt"
	"log"
	"os"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	PollRate            int      `toml:"poll_rate" comment:"track position update frequency in seconds"`
	MinPlaybackDuration int64    `toml:"min_playback_duration" comment:"minimum playback duration in seconds"`
	MinPlaybackPercent  int64    `toml:"min_playback_percent" comment:"minimum playback percentage"`
	Blacklist           []string `toml:"blacklist" comment:"MPRIS player blacklist"`

	LastFm *LastFmConfig `toml:"lastfm" comment:"last.fm configuration"`
	File   *FileConfig   `toml:"file" comment:"local file configuration"`
}

func (c Config) Providers() []Provider {
	providers := make([]Provider, 0)

	providers = append(providers, c.LastFm)
	providers = append(providers, c.File)

	return providers
}

type LastFmConfig struct {
	Username string `toml:"username" comment:"username"`
	Password string `toml:"password" comment:"password"`
	Key      string `toml:"key" comment:"API key"`
	Secret   string `toml:"secret" comment:"shared secret"`
}

type FileConfig struct {
	Filename string `toml:"filename" comment:"file to write scrobbles to"`
}

func ReadConfig() (*Config, error) {
	dir := fmt.Sprintf("%s/goscrobble", ConfigDir())

	err := os.MkdirAll(dir, 0700)
	if err != nil && !os.IsExist(err) {
		return nil, err
	}

	fileName := fmt.Sprintf("%s/config.toml", dir)
	//nolint:gosec // goscrobble runs as the user who owns the config, so this is not an issue
	data, err := os.ReadFile(fileName)
	if os.IsNotExist(err) {
		defaultConfig := Config{
			PollRate:            2,
			MinPlaybackDuration: 4 * 60,
			MinPlaybackPercent:  50,
			Blacklist:           []string{},
			LastFm:              nil,
			File:                nil,
		}
		defaultMarshalled, err := toml.Marshal(defaultConfig)
		if err != nil {
			return nil, err
		}

		if err := os.WriteFile(fileName, defaultMarshalled, 0600); err != nil {
			return nil, err
		}

		data = defaultMarshalled
	} else if err != nil {
		return nil, err
	}

	config := Config{}
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	if config.PollRate <= 0 || config.PollRate > 60 {
		config.PollRate = 2
	}

	// https://www.last.fm/api/scrobbling#when-is-a-scrobble-a-scrobble
	if config.MinPlaybackDuration <= 0 || config.MinPlaybackDuration > 20*60 {
		config.MinPlaybackDuration = 4 * 60
	}
	if config.MinPlaybackPercent <= 0 || config.MinPlaybackPercent > 100 {
		config.MinPlaybackPercent = 50
	}

	if config.LastFm == nil && config.File == nil {
		log.Println("no scrobbling providers configured, this is probably not what you want")
	}

	return &config, nil
}

// https://wiki.archlinux.org/title/XDG_Base_Directory
func ConfigDir() string {
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome != "" {
		return xdgConfigHome
	}

	home := os.Getenv("HOME")
	if home == "" {
		log.Fatalf("HOME environment variable is not set")
	}

	return fmt.Sprintf("%s/.config", home)
}
