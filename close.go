package main

import (
	"os"

	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog/log"
)

func CloseFile(file *os.File) {
	if err := file.Close(); err != nil {
		log.Error().Err(err).Msg("error closing scrobbles file")
	}
}

func CloseDBus(conn *dbus.Conn) {
	if err := conn.Close(); err != nil {
		log.Warn().Err(err).Msg("error closing dbus connection")
	}
}
