package main

import (
	"os/exec"

	"github.com/rs/zerolog/log"
)

func SendNotification(_ uint32, summary, body string) (uint32, error) {
	log.Debug().
		Str("summary", summary).
		Str("body", body).
		Msg("sending desktop notification via terminal-notifier")

	// https://github.com/julienXX/terminal-notifier
	cmd := exec.Command("/usr/bin/env", "terminal-notifier", "-title", "goscrobble", "-subtitle", summary, "-message", body)
	err := cmd.Run()

	log.Debug().Msg("sent desktop notification")

	return 0, err
}
