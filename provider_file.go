package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

func (f *FileConfig) Name() string {
	return "local file"
}

func (f *FileConfig) NowPlaying(_ NowPlaying) error {
	return nil
}

func (f *FileConfig) Scrobble(n NowPlaying) error {
	if f == nil {
		return nil
	}

	//nolint:gosec // goscrobble runs as the user who owns the config, so this is not an issue
	file, err := os.OpenFile(f.Filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			log.Err(err).Msg("error closing scrobbles file")
		}
	}(file)

	line := strings.Join([]string{
		n.Track,
		n.Album,
		strings.Join(n.Artists, ", "),
		strconv.FormatInt(n.Timestamp, 10),
	}, "|")

	_, err = fmt.Fprintf(file, "%s\n", line)
	return err
}
