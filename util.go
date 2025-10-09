package main

import (
	"os"

	"github.com/rs/zerolog/log"
)

func closeFileLogged(file *os.File) {
	if err := file.Close(); err != nil {
		log.Error().Err(err).Msg("error closing scrobbles file")
	}
}
