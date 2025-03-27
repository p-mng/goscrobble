package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

func (f *FileConfig) NowPlaying(_ NowPlaying) {

}

func (f *FileConfig) Scrobble(n NowPlaying) {
	if f == nil {
		return
	}

	//nolint:gosec // goscrobble runs as the user who owns the config, so this is not an issue
	file, err := os.OpenFile(f.Filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Printf("[file] error opening %s: %v", f.Filename, err)
		return
	}
	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			fmt.Printf("error closing scrobbles file: %v", err)
		}
	}(file)

	line := strings.Join([]string{
		n.Track,
		n.Album,
		strings.Join(n.Artists, ", "),
		strconv.FormatInt(n.Timestamp, 10),
	}, "|")

	if _, err := fmt.Fprintf(file, "%s\n", line); err != nil {
		log.Printf("[file] error writing scrobble: %v", err)
		return
	}

	log.Println("[file] scrobbled ✅")
}
