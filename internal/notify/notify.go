package notify

import (
	"os/exec"

	"github.com/tlepoid/tumuxi/internal/logging"
)

// Send sends a desktop notification using notify-send.
// It runs asynchronously and logs errors without blocking.
func Send(title, body string) {
	go func() {
		cmd := exec.Command("notify-send", "--app-name=tumuxi", title, body)
		if err := cmd.Run(); err != nil {
			logging.Warn("notify-send failed: %v", err)
		}
	}()
}
