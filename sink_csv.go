package main

import (
	"encoding/csv"
	"os"
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
