package main

import (
	"flag"
	"os"

	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Logger = log.With().Caller().Logger()

	flagDebug := flag.Bool("debug", false, "set log level to debug")
	flagJSON := flag.Bool("json", false, "print log messages in JSON format")
	flag.Parse()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *flagDebug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	log.Logger = log.Output(os.Stderr)
	if !*flagJSON {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		log.Info().Err(err).Msg("failed to connect to session bus")
		os.Exit(1)
	}
	defer func(conn *dbus.Conn) {
		if err := conn.Close(); err != nil {
			log.Warn().Err(err).Msg("error closing dbus connection")
		}
	}(conn)

	config, err := ReadConfig()
	if err != nil {
		log.Error().Err(err).Msg("error reading config file")
		return
	}

	RunMainLoop(conn, config)
}
