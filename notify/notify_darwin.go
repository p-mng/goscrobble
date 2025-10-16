package notify

import "os/exec"

// https://github.com/julienXX/terminal-notifier
func SendNotification(
	_ uint32,
	summary,
	body string,
) (uint32, error) {
		cmd := exec.Command(
			"/usr/bin/env",
			"terminal-notifier",
			"-title",
			"goscrobble",
			"-subtitle",
			summary,
			"-message",
			body,
		)
		return 0, cmd.Run()
}
