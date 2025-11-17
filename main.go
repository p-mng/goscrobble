package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
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
				Name:  "scrobbles",
				Usage: "Print scrobbles for the given sink",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:    "limit",
						Aliases: []string{"l"},
						Value:   10,
						Usage:   "maximum number of scrobbles to display",
					},
					&cli.TimestampFlag{
						Name:        "from",
						Aliases:     []string{"f"},
						Value:       time.Now().Add(-14 * 24 * time.Hour),
						DefaultText: "current datetime minus 14 days",
						Usage:       "only display scrobbles after this time",
					},
					&cli.TimestampFlag{
						Name:        "to",
						Aliases:     []string{"t"},
						Value:       time.Now(),
						DefaultText: "current datetime",
						Usage:       "only display scrobbles before this time",
					},
				},
				Arguments: []cli.Argument{
					&cli.StringArg{Name: "sink"},
				},
				Action: ActionScrobbles,
			},
			{
				Name:   "list-sinks",
				Usage:  "Print names of all configured sinks",
				Action: ActionListSinks,
			},
			{
				Name:   "lastfm-auth",
				Usage:  "Authenticate last.fm sand save session key and username",
				Action: ActionAuthLastFm,
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Println("Error:", err.Error())
	}
}

func ActionRun(_ context.Context, cmd *cli.Command) error {
	SetupLogger(cmd)

	config, err := ReadConfig()
	if err != nil {
		log.Error().
			Err(err).
			Msg("error reading config file")
		return nil
	}
	log.Debug().
		Interface("config", config).
		Msg("parsed config")

	RunMainLoop(config)

	return nil
}

func ActionScrobbles(_ context.Context, cmd *cli.Command) error {
	SetupLogger(cmd)

	limit := cmd.Int("limit")
	from := cmd.Timestamp("from")
	to := cmd.Timestamp("to")

	sinkName := cmd.StringArg("sink")

	if sinkName == "" {
		fmt.Println("No sink provided. Run `goscrobble list-sinks` to list all configured sinks.")
		return nil
	}

	config, err := ReadConfig()
	if err != nil {
		fmt.Println("Error reading config file:", err.Error())
		return nil
	}

	var sink Sink
	for _, s := range config.SetupSinks() {
		if s.Name() == sinkName {
			sink = s
			break
		}
	}

	if sink == nil {
		fmt.Println("Invalid sink name. Run `goscrobble list-sinks` to list all configured sinks.")
		return nil
	}

	scrobbles, err := sink.GetScrobbles(limit, from, to)
	if err != nil {
		fmt.Println("Error fetching scrobbles:", err.Error())
		return nil
	}

	tbl := table.NewWriter()
	tbl.SetOutputMirror(os.Stdout)

	tbl.AppendHeader(table.Row{"Artists", "Track", "Album", "Duration", "Timestamp"})
	for _, s := range scrobbles {
		tbl.AppendRow(table.Row{s.JoinArtists(), s.Track, s.Album, s.Duration, s.Timestamp.Format(time.RFC1123)})
	}

	tbl.Render()

	return nil
}

func ActionListSinks(_ context.Context, cmd *cli.Command) error {
	SetupLogger(cmd)

	config, err := ReadConfig()
	if err != nil {
		fmt.Println("Error reading config file:", err.Error())
	}

	for _, sink := range config.SetupSinks() {
		fmt.Println(sink.Name())
	}

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

	if config.Sinks.LastFm.SessionKey != "" && config.Sinks.LastFm.Username != "" {
		fmt.Println("last.fm is already authenticated")
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
		log.Error().
			Err(err).
			Msg("error writing updated config file")
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
