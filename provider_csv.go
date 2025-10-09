package main

import (
	"encoding/csv"
	"os"
	"strconv"
)

func (c *CSVConfig) Name() string {
	return "local CSV file"
}

func (c *CSVConfig) NowPlaying(_ NowPlaying) error {
	return nil
}

func (c *CSVConfig) Scrobble(n NowPlaying) error {
	readExisting := true

	file, err := os.Open(c.Filename)
	if os.IsNotExist(err) {
		readExisting = false
	} else if err != nil {
		return err
	}

	var scrobbles [][]string

	if readExisting {
		defer closeFileLogged(file)

		readScrobbles, err := csv.NewReader(file).ReadAll()
		if err != nil {
			return err
		}

		readScrobbles = append(
			readScrobbles,
			createRow(n.JoinArtists(), n.Track, n.Album, n.Duration, n.Timestamp),
		)
		scrobbles = readScrobbles
	} else {
		scrobbles = append(
			scrobbles,
			createRow(n.JoinArtists(), n.Track, n.Album, n.Duration, n.Timestamp),
		)
	}

	newFile, err := os.Create(c.Filename)
	if err != nil {
		return err
	}

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
