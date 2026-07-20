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
	//nolint:gosec
	cmd := exec.Command("/usr/bin/env", "terminal-notifier", "-title", "goscrobble", "-subtitle", summary, "-message", body)
	err := cmd.Run()
	if err != nil {
		log.Error().
			Int("exit code", cmd.ProcessState.ExitCode()).
			Msg("terminal-notifier exited with error")
		return 0, err
	}

	log.Debug().
		Int("exit code", cmd.ProcessState.ExitCode()).
		Msg("sent desktop notification using terminal-notifier")
	return 0, nil
}
