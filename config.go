package main

import (
	"fmt"
	"os"
	"regexp"

	"github.com/godbus/dbus/v5"
	"github.com/pelletier/go-toml/v2"
	"github.com/rs/zerolog/log"
)

type Config struct {
	PollRate            int            `toml:"poll_rate"`
	MinPlaybackDuration int64          `toml:"min_playback_duration"`
	MinPlaybackPercent  int64          `toml:"min_playback_percent"`
	NotifyOnScrobble    bool           `toml:"notify_on_scrobble"`
	NotifyOnError       bool           `toml:"notify_on_error"`
	Blacklist           []string       `toml:"blacklist"`
	Regexes             []RegexReplace `toml:"regexes"`

	Sources SourcesConfig `toml:"sources"`
	Sinks   SinksConfig   `toml:"sinks"`
}

type RegexReplace struct {
	Match   string `toml:"match"`
	Replace string `toml:"replace"`
	Artist  bool   `toml:"artist"`
	Track   bool   `toml:"track"`
	Album   bool   `toml:"album"`
}

type SourcesConfig struct {
	DBus         *DBusConfig         `toml:"dbus"`
	MediaControl *MediaControlConfig `toml:"media-control"`
}

type SinksConfig struct {
	LastFm *LastFmConfig `toml:"lastfm"`
	CSV    *CSVConfig    `toml:"csv"`
}

type DBusConfig struct {
	Address string `toml:"address"`
}

type MediaControlConfig struct {
	Command   string   `toml:"command"`
	Arguments []string `toml:"arguments"`
}

type LastFmConfig struct {
	Key        string `toml:"key"`
	Secret     string `toml:"secret"`
	SessionKey string `toml:"session_key"`
}

type CSVConfig struct {
	Filename string `toml:"filename"`
}

func (c Config) GetSources() []Source {
	sources := make([]Source, 0)

	if c.Sources.DBus != nil {
		var conn *dbus.Conn
		var err error
		if c.Sources.DBus.Address == "" {
			conn, err = dbus.ConnectSessionBus()
		} else {
			conn, err = dbus.Connect(c.Sources.DBus.Address)
		}

		if err != nil {
			log.Error().
				Err(err).
				Str("address", c.Sources.DBus.Address).
				Msg("failed to connect to session bus")
		}
		sources = append(sources, DBusSource{Conn: conn})
	}
	if c.Sources.MediaControl != nil {
		sources = append(sources, MediaControlSource{
			Command:   c.Sources.MediaControl.Command,
			Arguments: c.Sources.MediaControl.Arguments,
		})
	}

	return sources
}

func (c Config) GetSinks() []Sink {
	var sinks []Sink

	if c.Sinks.LastFm != nil {
		sinks = append(sinks, c.Sinks.LastFm)
	}
	if c.Sinks.CSV != nil {
		sinks = append(sinks, c.Sinks.CSV)
	}

	return sinks
}

func (c Config) ParseRegexes() []ParsedRegexReplace {
	var parsed []ParsedRegexReplace

	for _, r := range c.Regexes {
		match, err := regexp.Compile(r.Match)
		if err != nil {
			log.Warn().
				Err(err).
				Str("expression", r.Match).
				Msg("error compiling match/repalce expression")
			continue
		}
		parsed = append(parsed, ParsedRegexReplace{
			Match:   match,
			Replace: r.Replace,
			Artist:  r.Artist,
			Track:   r.Track,
			Album:   r.Album,
		})
	}

	return parsed
}

func ReadConfig() (Config, error) {
	configDir := ConfigDir()
	if err := os.MkdirAll(configDir, 0700); err != nil && !os.IsExist(err) {
		return Config{}, fmt.Errorf("failed to create config directory: %w", err)
	}

	filename := fmt.Sprintf("%s/config.toml", configDir)
	log.Debug().Str("filename", filename).Msg("reading config file")

	var config Config
	var configErr error

	//nolint:gosec
	data, err := os.ReadFile(filename)
	switch {
	case err == nil:
		config, configErr = ParseConfig(data)
	case os.IsNotExist(err):
		config, configErr = CreateDefaultConfig(filename)
	default:
		configErr = err
	}

	if configErr != nil {
		return Config{}, configErr
	}

	config.Validate()

	return config, nil
}

func ParseConfig(data []byte) (Config, error) {
	var config Config
	return config, toml.Unmarshal(data, &config)
}

func CreateDefaultConfig(filename string) (Config, error) {
	log.Info().
		Str("filename", filename).
		Msg("config file does not exist, writing default config")

	config := Config{
		PollRate:            2,
		MinPlaybackDuration: 4 * 60,
		MinPlaybackPercent:  50,
		Blacklist:           []string{},
		Regexes:             []RegexReplace{},
		NotifyOnScrobble:    false,
		NotifyOnError:       true,
		Sources: SourcesConfig{
			DBus:         &DBusConfig{Address: ""},
			MediaControl: &MediaControlConfig{Command: "media-control", Arguments: []string{"get", "--now"}},
		},
		Sinks: SinksConfig{
			LastFm: &LastFmConfig{Key: "last.fm API key", Secret: "last.fm API secret", SessionKey: ""},
			CSV:    &CSVConfig{Filename: fmt.Sprintf("%s/scrobbles.csv", os.Getenv("HOME"))},
		},
	}

	data, err := toml.Marshal(config)
	if err != nil {
		return Config{}, fmt.Errorf("failed to marshal default config: %w", err)
	}

	if err := os.WriteFile(filename, data, 0600); err != nil {
		return Config{}, fmt.Errorf("failed to write default config: %w", err)
	}

	return config, nil
}

func (c *Config) Validate() {
	if c.PollRate <= 0 || c.PollRate > 60 {
		log.Warn().
			Int("poll_rate", c.PollRate).
			Msg("invalid poll rate, resetting to default value")
		c.PollRate = 2
	}
	if c.MinPlaybackDuration <= 0 || c.MinPlaybackDuration > 20*60 {
		log.Warn().
			Int64("min_playback_duration", c.MinPlaybackDuration).
			Msg("invalid minimum playback duration, resetting to default value")
		// https://www.last.fm/api/scrobbling#when-is-a-scrobble-a-scrobble
		c.MinPlaybackDuration = 4 * 60
	}
	if c.MinPlaybackPercent <= 0 || c.MinPlaybackPercent > 100 {
		log.Warn().
			Int64("min_playback_percent", c.MinPlaybackPercent).
			Msg("invalid minimum playback percentage, resetting to default value")
		c.MinPlaybackPercent = 50
	}

	if !c.NotifyOnError {
		log.Warn().Msg("goscrobble will not send desktop notifications on failed scrobbles")
	}
	if len(c.GetSinks()) == 0 {
		log.Warn().Msg("no scrobbling sinks configured, this is probably not what you want")
	}
	if c.Sinks.LastFm != nil && c.Sinks.LastFm.SessionKey == "" {
		log.Warn().Msg("last.fm sink is configured, but not authenticated: run goscrobble auth to generate a token")
	}
}

func (c Config) WriteConfig() error {
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
