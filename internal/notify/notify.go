package notify

import (
	"os/exec"
	"strings"

	"github.com/tlepoid/tumuxi/internal/logging"
)

// Send sends a desktop notification using notify-send.
// If onAction is non-nil, the notification includes a "Switch" action button.
// When clicked, onAction is called from a background goroutine.
// The goroutine blocks until the notification is dismissed or acted on.
func Send(title, body string, onAction func()) {
	go func() {
		args := []string{"--app-name=tumuxi"}
		if onAction != nil {
			args = append(args, "--action=switch=Switch")
		}
		args = append(args, title, body)

		cmd := exec.Command("notify-send", args...)
		out, err := cmd.Output()
		if err != nil {
			logging.Warn("notify-send failed: %v", err)
			return
		}
		if onAction != nil && strings.TrimSpace(string(out)) == "switch" {
			onAction()
		}
	}()
}
