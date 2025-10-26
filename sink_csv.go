package main

import (
	"encoding/csv"
	"os"
	"strconv"
)

func (c *CSVConfig) Name() string {
	return "local CSV file"
}

func (c *CSVConfig) NowPlaying(_ PlaybackStatus) error {
	return nil
}

func (c *CSVConfig) Scrobble(p PlaybackStatus) error {
	var scrobbles [][]string

	file, err := os.Open(c.Filename)
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

	newFile, err := os.Create(c.Filename)
	if err != nil {
		return err
	}
	defer CloseFile(newFile)

	return csv.NewWriter(newFile).WriteAll(scrobbles)
}

func createRow(artists, track, album string, duration, timestamp int64) []string {
	return []string{
		artists,
		track,
		album,
		strconv.FormatInt(duration, 10),
		strconv.FormatInt(timestamp, 10),
	}
}
