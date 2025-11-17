package main

import (
	"bufio"
	"encoding/csv"
	"os"
	"slices"
	"time"

	"github.com/rs/zerolog/log"
)

type CSVSink struct {
	Filename string
}

func CSVSinkFromConfig(c CSVConfig) CSVSink {
	return CSVSink(c)
}

func (s CSVSink) Name() string {
	return "csv"
}

func (s CSVSink) NowPlaying(_ Scrobble) error {
	return nil
}

func (s CSVSink) Scrobble(scrobble Scrobble) error {
	var scrobbles [][]string

	file, err := os.Open(s.Filename)
	if err == nil {
		defer CloseLogged(file)

		scrobbles, err = csv.NewReader(file).ReadAll()
		if err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	scrobbles = append(scrobbles, scrobble.ToStringSlice())

	newFile, err := os.Create(s.Filename)
	if err != nil {
		return err
	}
	defer CloseLogged(newFile)

	return csv.NewWriter(newFile).WriteAll(scrobbles)
}

func (s CSVSink) GetScrobbles(limit int, from, to time.Time) ([]Scrobble, error) {
	log.Debug().
		Str("filename", s.Filename).
		Msg("opening scrobbles file")

	file, err := os.Open(s.Filename)
	if err != nil {
		return nil, err
	}
	defer CloseLogged(file)

	scanner := bufio.NewScanner(file)

	log.Debug().
		Str("filename", file.Name()).
		Msg("reading lines")

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if scanner.Err() != nil {
		return nil, err
	}

	log.Debug().
		Int("length", len(lines)).
		Msg("reversing slice")
	slices.Reverse(lines)

	var scrobbles []Scrobble

	log.Debug().
		Int("length", len(lines)).
		Msg("processing lines")

	for _, line := range lines {
		scrobble, err := ScrobbleFromCSV(line)
		if err != nil {
			return nil, err
		}

		if scrobble.Timestamp.Before(from) || scrobble.Timestamp.After(to) {
			log.Debug().
				Time("timestamp", scrobble.Timestamp).
				Time("from", from).
				Time("to", to).
				Msg("skipping scrobble with invalid timestamp")
			continue
		}

		scrobbles = append(scrobbles, scrobble)

		if len(scrobbles) >= limit {
			log.Debug().
				Int("limit", limit).
				Msg("reached limit")
			break
		}
	}

	return scrobbles, nil
}
