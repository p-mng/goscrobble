package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func (f *FileConfig) Name() string {
	return "local file"
}

func (f *FileConfig) NowPlaying(_ NowPlayingInfo) error {
	return nil
}

func (f *FileConfig) Scrobble(n NowPlayingInfo) error {
	// https://pkg.go.dev/os#pkg-constants
	file, err := os.OpenFile(f.Filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer closeFileLogged(file)

	line := strings.Join([]string{
		n.Track,
		n.Album,
		n.JoinArtists(),
		strconv.FormatInt(n.Timestamp, 10),
	}, "|")

	_, err = fmt.Fprintf(file, "%s\n", line)
	return err
}
