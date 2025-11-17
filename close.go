package main

import (
	"github.com/rs/zerolog/log"
	"io"
)

func CloseLogged(closer io.Closer) {
	if err := closer.Close(); err != nil {
		log.Error().
			Err(err).
			Interface("closer", closer).
			Msg("error calling `Close`")
	}
}
