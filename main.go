package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	lastfm "github.com/p-mng/lastfm-go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
				Usage:  "Watch sources and send scrobbles to configured sinks",
				Action: ActionRun,
			},
			{
				Name:   "lastfm-auth",
				Usage:  "Authenticate last.fm sand save session key and username",
				Action: ActionAuthLastFm,
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		panic(err)
	}
}

func ActionRun(_ context.Context, cmd *cli.Command) error {
	SetupLogger(cmd)

	config, err := ReadConfig()
	if err != nil {
		log.Error().Err(err).Msg("error reading config file")
		return nil
	}
	log.Debug().Any("config", config).Msg("parsed config")

	RunMainLoop(config)

	return nil
}

func ActionAuthLastFm(_ context.Context, cmd *cli.Command) error {
	SetupLogger(cmd)

	config, err := ReadConfig()
	if err != nil {
		fmt.Println("Error reading config file:", err.Error())
		return nil
	}

	if config.Sinks.LastFm == nil {
		fmt.Println("Error: last.fm sink is not configured")
		return nil
	}

	client, err := lastfm.NewDesktopClient(lastfm.BaseURL, config.Sinks.LastFm.Key, config.Sinks.LastFm.Secret)
	if err != nil {
		fmt.Println("Error setting up last.fm client:", err.Error())
		return nil
	}

	token, err := client.AuthGetToken()
	if err != nil {
		fmt.Println("Error getting authorization token:", err.Error())
		return nil
	}

	authURL := client.DesktopAuthorizationURL(token.Token)

	//nolint:gosec
	openBrowserCmd := exec.Command("/usr/bin/env", "xdg-open", authURL)
	if err := openBrowserCmd.Run(); err != nil {
		fmt.Println("Error opening URL in default browser:", err.Error())
	}

	fmt.Println("Please open the following URL in your browser and authorize the application:", authURL)
	fmt.Print("Finished authorization? [Y/n] ")

	input := bufio.NewScanner(os.Stdin)
	input.Scan()

	response := strings.ToLower(strings.TrimSpace(input.Text()))
	if response != "y" && response != "" {
		return nil
	}

	session, err := client.AuthGetSession(token.Token)
	if err != nil {
		fmt.Println("Error fetching session key from last.fm API:", err.Error())
		return nil
	}

	fmt.Println("Logged in with user:", session.Session.Name)

	config.Sinks.LastFm.SessionKey = session.Session.Key
	config.Sinks.LastFm.Username = session.Session.Name

	filename := fmt.Sprintf("%s/%s", ConfigDir(), DefaultConfigFileName)

	if err := config.Write(filename); err != nil {
		log.Error().Err(err).Msg("failed to write updated config file")
		return nil
	}

	return nil
}

func SetupLogger(cmd *cli.Command) {
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
