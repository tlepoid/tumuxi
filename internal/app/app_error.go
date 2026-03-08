package app

import (
	tea "charm.land/bubbletea/v2"

	"github.com/tlepoid/tumuxi/internal/logging"
	"github.com/tlepoid/tumuxi/internal/messages"
)

func (a *App) handleErrorMessage(msg messages.Error) tea.Cmd {
	if msg.Err == nil {
		return nil
	}
	a.err = msg.Err
	if !msg.Logged {
		logging.Error("Error in %s: %v", msg.Context, msg.Err)
	}
	return nil
}
