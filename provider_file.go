package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func (f *FileConfig) NowPlaying(_ NowPlaying) {

}

func (f *FileConfig) Scrobble(n NowPlaying) {
	if f == nil {
		return
	}

	filename := f.Filename
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("[file] error opening %s: %v", filename, err)
		return
	}
	defer file.Close()

	line := strings.Join([]string{
		n.Track,
		n.Album,
		strings.Join(n.Artists, ", "),
		strconv.FormatInt(time.Now().Unix(), 10),
	}, "|")

	if _, err := file.WriteString(fmt.Sprintf("%s\n", line)); err != nil {
		log.Printf("[file] error writing scrobble: %v", err)
		return
	}

	log.Println("[file] scrobbled ✅")
}
