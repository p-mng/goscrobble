package main

import "os/exec"

func SendNotification(_ uint32, summary, body string) (uint32, error) {
	// https://github.com/julienXX/terminal-notifier
	cmd := exec.Command("/usr/bin/env", "terminal-notifier", "-title", "goscrobble", "-subtitle", summary, "-message", body)
	return 0, cmd.Run()
}
