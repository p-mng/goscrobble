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

func (s CSVSink) Name() string {
	return "local CSV file"
}

func (s CSVSink) NowPlaying(_ PlaybackStatus) error {
	return nil
}

func (s CSVSink) Scrobble(p PlaybackStatus) error {
	var scrobbles [][]string

	file, err := os.Open(s.Filename)
	if err == nil {
		defer CloseFile(file)

		scrobbles, err = csv.NewReader(file).ReadAll()
		if err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	scrobbles = append(
		scrobbles,
		createRow(p.JoinArtists(), p.Track, p.Album, p.Duration, p.Timestamp),
	)

	newFile, err := os.Create(s.Filename)
	if err != nil {
		return err
	}
	defer CloseFile(newFile)

	return csv.NewWriter(newFile).WriteAll(scrobbles)
}

func createRow(artists, track, album string, duration time.Duration, timestamp time.Time) []string {
	return []string{
		artists,
		track,
		album,
		strconv.FormatFloat(duration.Seconds(), 'f', 2, 64),
		timestamp.Format(time.RFC1123),
	}
}
