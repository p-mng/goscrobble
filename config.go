package main

import (
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/BurntSushi/toml"
	"github.com/godbus/dbus/v5"
	lastfm "github.com/p-mng/lastfm-go"
	"github.com/rs/zerolog/log"
)

const DefaultConfigFileName = "config.toml"

var DefaultConfig = Config{
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
		LastFm: map[string]LastFmConfig{"default": {
			BaseURL:    lastfm.BaseURL,
			Key:        "last.fm API key",
			Secret:     "last.fm API secret",
			SessionKey: "",
			Username:   "",
		}},
		CSV: map[string]CSVConfig{"default": {
			Filename: filepath.Join(os.Getenv("HOME"), "scrobbles.csv"),
		}},
	},
}

type Config struct {
	PollRate            int            `toml:"poll_rate"`
	MinPlaybackDuration int            `toml:"min_playback_duration"`
	MinPlaybackPercent  int            `toml:"min_playback_percent"`
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
	TidalHifi    *TidalHifiConfig    `toml:"tidal-hifi"`
}

type SinksConfig struct {
	LastFm map[string]LastFmConfig `toml:"lastfm"`
	CSV    map[string]CSVConfig    `toml:"csv"`
}

type DBusConfig struct {
	Address string `toml:"address"`
}

type MediaControlConfig struct {
	Command   string   `toml:"command"`
	Arguments []string `toml:"arguments"`
}

type TidalHifiConfig struct {
	Endpoint string `toml:"endpoint"`
}

type LastFmConfig struct {
	BaseURL    string `toml:"base_url"`
	Key        string `toml:"key"`
	Secret     string `toml:"secret"`
	SessionKey string `toml:"session_key"`
	Username   string `toml:"username"`
}

type CSVConfig struct {
	Filename string `toml:"filename"`
}

func (c Config) SetupSources() []Source {
	var sources []Source

	if c.Sources.DBus != nil {
		log.Debug().Msg("setting up dbus source")

		var conn *dbus.Conn
		var err error
		if c.Sources.DBus.Address == "" {
			log.Debug().Msg("connecting to session bus")
			conn, err = dbus.ConnectSessionBus()
		} else {
			log.Debug().Str("address", c.Sources.DBus.Address).Msg("connecting to bus")
			conn, err = dbus.Connect(c.Sources.DBus.Address)
		}

		if err != nil {
			log.Error().
				Err(err).
				Str("address", c.Sources.DBus.Address).
				Msg("failed to connect to bus")
		} else {
			sources = append(sources, DBusSource{Conn: conn})
		}
	}

	if c.Sources.MediaControl != nil {
		log.Debug().Msg("setting up media-control source")
		sources = append(sources, MediaControlSource{
			Command:   c.Sources.MediaControl.Command,
			Arguments: c.Sources.MediaControl.Arguments,
		})
	}

	if c.Sources.TidalHifi != nil {
		var endpoint string
		if c.Sources.TidalHifi.Endpoint == "" {
			log.Debug().Str("endpoint", DefaultTidalHifiEndpoint).Msg("using default endpoint")
			endpoint = DefaultTidalHifiEndpoint
		} else {
			endpoint = c.Sources.TidalHifi.Endpoint
		}

		log.Debug().Msg("setting up tidal-hifi API source")
		sources = append(sources, TidalHifiSource{
			Client:   http.Client{},
			Endpoint: endpoint,
		})
	}

	if len(sources) == 0 {
		log.Warn().Msg("no sources configured")
	} else {
		log.Debug().Msg("set up sources")
	}

	return sources
}

func (c Config) SetupSinks() []Sink {
	var sinks []Sink

	for _, sinkConfig := range c.Sinks.LastFm {
		log.Debug().Msg("setting up last.fm sink")

		sink, err := LastFmSinkFromConfig(sinkConfig)
		if err != nil {
			log.Error().
				Err(err).
				Msg("error setting up last.fm sink")
		} else {
			sinks = append(sinks, sink)
		}
	}

	for _, sinkConfig := range c.Sinks.CSV {
		log.Debug().Msg("setting up CSV sink")

		sink := CSVSinkFromConfig(sinkConfig)
		sinks = append(sinks, sink)
	}

	if len(sinks) == 0 {
		log.Warn().Msg("no sinks configured")
	} else {
		log.Debug().Msg("set up sinks")
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

	log.Debug().Msg("parsed match/replace expressions")
	return parsed
}

func ReadConfig(filename string) (Config, error) {
	log.Debug().Msg("creating config directory")
	directory := filepath.Dir(filename)
	if err := os.MkdirAll(directory, 0700); err != nil {
		return Config{}, err
	}

	log.Debug().Msg("reading config")
	var config Config
	_, err := toml.DecodeFile(filename, &config)

	switch {
	case os.IsNotExist(err):
		config = DefaultConfig

		log.Info().
			Str("filename", filename).
			Msg("creating default configuration file")
		if err := DefaultConfig.Write(filename); err != nil {
			return Config{}, err
		}
	case err != nil:
		return Config{}, err
	}

	log.Debug().Msg("successfully read configuration")

	config.Validate()

	return config, nil
}

func (c *Config) Validate() {
	log.Debug().Msg("validating configuration")

	if c.PollRate <= 0 || c.PollRate > 60 {
		log.Warn().
			Int("poll_rate", c.PollRate).
			Msg("invalid poll rate, using default value")
		c.PollRate = 2
	}
	if c.MinPlaybackDuration <= 0 || c.MinPlaybackDuration > 20*60 {
		log.Warn().
			Int("min_playback_duration", c.MinPlaybackDuration).
			Msg("invalid minimum playback duration, using default value")
		// https://www.last.fm/api/scrobbling#when-is-a-scrobble-a-scrobble
		c.MinPlaybackDuration = 4 * 60
	}
	if c.MinPlaybackPercent <= 0 || c.MinPlaybackPercent > 100 {
		log.Warn().
			Int("min_playback_percent", c.MinPlaybackPercent).
			Msg("invalid minimum playback percentage, using default value")
		c.MinPlaybackPercent = 50
	}

	if !c.NotifyOnError {
		log.Warn().Msg("goscrobble will not send desktop notifications on failed scrobbles")
	}

	if c.Sources.MediaControl != nil && len(c.Sources.MediaControl.Arguments) == 0 {
		log.Warn().Msg("no arguments for media-control specified, using `get --now`")
		c.Sources.MediaControl.Arguments = []string{"get", "--now"}
	}

	log.Debug().Msg("validated configuration")
}

func (c Config) Write(filename string) error {
	log.Debug().
		Str("filename", filename).
		Msg("writing config file")

	//nolint:gosec
	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	encoder := toml.NewEncoder(file)
	encoder.Indent = ""

	return encoder.Encode(c)
}

func ConfigDir() string {
	// https://specifications.freedesktop.org/basedir-spec/latest/
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome != "" {
		return filepath.Join(configHome, "goscrobble")
	}
	return filepath.Join(os.Getenv("HOME"), ".config", "goscrobble")
}
