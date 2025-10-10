package main

import (
	"fmt"
	"os"
	"regexp"

	"github.com/pelletier/go-toml/v2"
	"github.com/rs/zerolog/log"
)

type Config struct {
	PollRate            int          `toml:"poll_rate" comment:"track position update frequency in seconds"`
	MinPlaybackDuration int64        `toml:"min_playback_duration" comment:"minimum playback duration in seconds"`
	MinPlaybackPercent  int64        `toml:"min_playback_percent" comment:"minimum playback percentage"`
	Blacklist           []string     `toml:"blacklist" comment:"MPRIS player blacklist"`
	Regexes             []RegexEntry `toml:"regexes" comment:"regex match/replace"`
	NotifyOnScrobble    bool         `toml:"notify_on_scrobble" comment:"send a desktop notification when a scrobble is saved"`
	NotifyOnError       bool         `toml:"notify_on_error" comment:"send a desktop notification when a scrobble cannot be saved"`

	LastFm *LastFmConfig `toml:"lastfm" comment:"last.fm configuration"`
	File   *FileConfig   `toml:"file" comment:"local file configuration"`
	CSV    *CSVConfig    `toml:"csv" comment:"CSV file configuration"`
}

type RegexEntry struct {
	Match   string `toml:"match"`
	Replace string `toml:"replace"`
	Artist  bool   `toml:"artist"`
	Track   bool   `toml:"track"`
	Album   bool   `toml:"album"`
}

type LastFmConfig struct {
	Key        string `toml:"key" comment:"API key"`
	Secret     string `toml:"secret" comment:"shared secret"`
	SessionKey string `toml:"session_key" comment:"session key (automatically generated using goscrobble auth)"`
}

type FileConfig struct {
	Filename string `toml:"filename" comment:"file to write scrobbles to"`
}

type CSVConfig struct {
	Filename string `toml:"filename" comment:"file to write scrobbles to"`
}

func (c Config) Providers() []Provider {
	providers := make([]Provider, 0)

	if c.LastFm != nil {
		providers = append(providers, c.LastFm)
	}
	if c.File != nil {
		providers = append(providers, c.File)
	}
	if c.CSV != nil {
		providers = append(providers, c.CSV)
	}

	return providers
}

func (c Config) ParseRegexes() []ParsedRegexEntry {
	var parsed []ParsedRegexEntry

	for _, r := range c.Regexes {
		match, err := regexp.Compile(r.Match)
		if err != nil {
			log.Warn().
				Err(err).
				Str("expression", r.Match).
				Msg("error compiling match/repalce expression")
			continue
		}
		parsed = append(parsed, ParsedRegexEntry{
			Match:   match,
			Replace: r.Replace,
			Artist:  r.Artist,
			Track:   r.Track,
			Album:   r.Album,
		})
	}

	return parsed
}

func ReadConfig() (*Config, error) {
	configDir := ConfigDir()

	if err := os.MkdirAll(configDir, 0700); err != nil && !os.IsExist(err) {
		return nil, err
	}

	fileName := fmt.Sprintf("%s/config.toml", configDir)

	//nolint:gosec
	data, err := os.ReadFile(fileName)
	if os.IsNotExist(err) {
		defaultConfig := Config{
			PollRate:            2,
			MinPlaybackDuration: 4 * 60,
			MinPlaybackPercent:  50,
			Blacklist:           []string{},
			Regexes:             []RegexEntry{},
			NotifyOnScrobble:    false,
			NotifyOnError:       true,
			LastFm:              nil,
			File:                nil,
			CSV:                 nil,
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

	if !config.NotifyOnError {
		log.Warn().Msg("goscrobble will not send desktop notifications on failed scrobbles")
	}

	if config.LastFm == nil && config.File == nil {
		log.Warn().Msg("no scrobbling providers configured, this is probably not what you want")
	}

	if config.LastFm != nil && config.LastFm.SessionKey == "" {
		log.Warn().Msg("last.fm provider is configured, but not authenticated: run goscrobble auth to generate a token")
	}

	return &config, nil
}

func (c *Config) WriteConfig() error {
	configDir := ConfigDir()
	fileName := fmt.Sprintf("%s/config.toml", configDir)

	data, err := toml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(fileName, data, 0600)
}

func ConfigDir() string {
	// https://wiki.archlinux.org/title/XDG_Base_Directory
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome != "" {
		return fmt.Sprintf("%s/goscrobble", xdgConfigHome)
	}

	home := os.Getenv("HOME")
	if home == "" {
		// $HOME environment variable is required to be set at all times
		// https://unix.stackexchange.com/a/123859
		panic("HOME environment variable is not set")
	}

	return fmt.Sprintf("%s/.config/goscrobble", home)
}
