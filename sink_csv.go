package main

import (
	"encoding/csv"
	"os"
	"strconv"
	"time"
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

	scrobbles = append(
		scrobbles,
		CreateRow(scrobble.JoinArtists(), scrobble.Track, scrobble.Album, scrobble.Duration, scrobble.Timestamp),
	)

	newFile, err := os.Create(s.Filename)
	if err != nil {
		return err
	}
	defer CloseLogged(newFile)

	return csv.NewWriter(newFile).WriteAll(scrobbles)
}

func CreateRow(artists, track, album string, duration time.Duration, timestamp time.Time) []string {
	return []string{
		artists,
		track,
		album,
		strconv.FormatFloat(duration.Seconds(), 'f', 2, 64),
		timestamp.Format(time.RFC1123),
	}
}
