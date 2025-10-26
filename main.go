package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/shkh/lastfm-go/lastfm"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:  "goscrobble",
		Usage: "A simple, cross-platform music scrobbler daemon",
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
				Usage:  "Continuously checks music players and sends scrobbles to configured sinks",
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

	config, err := ReadConfig()
	if err != nil {
		log.Error().Err(err).Msg("error reading config file")
		return nil
	}
	log.Debug().Any("config", config).Msg("parsed config")

	RunMainLoop(config)

	return nil
}

func cmdAuth(_ context.Context, cmd *cli.Command) error {
	setupLogger(cmd)

	config, err := ReadConfig()
	if err != nil {
		log.Error().Err(err).Msg("error reading config file")
		return nil
	}

	if config.Sinks.LastFm == nil || config.Sinks.LastFm.Key == "" || config.Sinks.LastFm.Secret == "" {
		log.Error().Msg("last.fm sink is not configured")
		return nil
	}

	api := lastfm.New(config.Sinks.LastFm.Key, config.Sinks.LastFm.Secret)

	token, err := api.GetToken()
	if err != nil {
		log.Error().Err(err).Msg("failed to generate last.fm request token")
		return nil
	}

	authURL := api.GetAuthTokenUrl(token)

	//nolint:gosec
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

	config.Sinks.LastFm.SessionKey = sessionKey

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

	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	if json {
		log.Logger = log.Output(os.Stderr)
	} else {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
}
