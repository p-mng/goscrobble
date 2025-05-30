package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/shkh/lastfm-go/lastfm"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:  "goscrobble",
		Usage: "A simple music scrobbler daemon for MPRIS-based music players",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"d"},
				Value:   false,
				Usage:   "print debug log messages",
			},
			&cli.BoolFlag{
				Name:    "json",
				Aliases: []string{"j"},
				Value:   false,
				Usage:   "print log messages in JSON format",
			},
		},
		Commands: []*cli.Command{
			{
				Name:   "run",
				Usage:  "Continuously checks music players and sends scrobbles to configured providers",
				Action: cmdRun,
			},
			{
				Name:   "auth",
				Usage:  "Authenticates last.fm session",
				Action: cmdAuth,
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		panic(err)
	}
}

func cmdRun(_ context.Context, cmd *cli.Command) error {
	setupLogger(cmd)

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
		return nil
	}

	RunMainLoop(conn, config)

	return nil
}

func cmdAuth(_ context.Context, cmd *cli.Command) error {
	setupLogger(cmd)

	config, err := ReadConfig()
	if err != nil {
		log.Error().Err(err).Msg("error reading config file")
		return nil
	}

	if config.LastFm == nil || config.LastFm.Key == "" || config.LastFm.Secret == "" {
		log.Error().Msg("last.fm provider is not configured")
		return nil
	}

	api := lastfm.New(config.LastFm.Key, config.LastFm.Secret)

	token, err := api.GetToken()
	if err != nil {
		log.Error().Err(err).Msg("failed to generate last.fm request token")
		return nil
	}

	authURL := api.GetAuthTokenUrl(token)

	//nolint:gosec // authURL comes from shkh/lastfm-go and cannot be set by an attacker
	openBrowserCmd := exec.Command("/usr/bin/env", "xdg-open", authURL)
	if err := openBrowserCmd.Run(); err != nil {
		log.Warn().Err(err).Msg("failed to open auth URL in web browser")
	}

	fmt.Println("please open the following URL in your browser and authorize the application:", authURL)
	fmt.Print("finished authorization? [Y/n] ")

	input := bufio.NewScanner(os.Stdin)
	input.Scan()

	response := strings.ToLower(strings.TrimSpace(input.Text()))
	if response != "y" && response != "" {
		return nil
	}

	if err := api.LoginWithToken(token); err != nil {
		log.Error().Err(err).Msg("failed to authenticate using request token")
		return nil
	}

	sessionKey := api.GetSessionKey()

	config.LastFm.SessionKey = sessionKey

	if err := config.WriteConfig(); err != nil {
		log.Error().Err(err).Msg("failed to write updated config file")
		return nil
	}

	return nil
}

func setupLogger(cmd *cli.Command) {
	debug := cmd.Bool("debug")
	json := cmd.Bool("json")

	log.Logger = log.With().Caller().Logger()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	log.Logger = log.Output(os.Stderr)
	if !json {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
}
