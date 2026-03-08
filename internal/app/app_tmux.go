package app

import (
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/tlepoid/tumuxi/internal/data"
	"github.com/tlepoid/tumuxi/internal/logging"
	"github.com/tlepoid/tumuxi/internal/messages"
	"github.com/tlepoid/tumuxi/internal/tmux"
)

func (a *App) cleanupWorkspaceTmuxSessions(ws *data.Workspace) tea.Cmd {
	if ws == nil {
		return nil
	}
	wsID := string(ws.ID())
	opts := a.tmuxOptions
	svc := a.tmuxService
	return func() tea.Msg {
		if svc == nil {
			return nil
		}
		tags := map[string]string{
			"@tumuxi":           "1",
			"@tumuxi_workspace": wsID,
		}
		cleaned, err := svc.KillSessionsMatchingTags(tags, opts)
		if err != nil {
			logging.Warn("Failed to cleanup tmux sessions for workspace %s: %v", ws.Name, err)
		}
		if cleaned {
			logging.Info("Cleaned up @tumuxi tmux sessions for workspace %s", ws.Name)
		}
		if err := svc.KillWorkspaceSessions(wsID, opts); err != nil {
			logging.Warn("Failed to cleanup tmux sessions for workspace %s: %v", ws.Name, err)
		}
		return nil
	}
}

func (a *App) cleanupAllTmuxSessions() tea.Cmd {
	opts := a.tmuxOptions
	svc := a.tmuxService
	return func() tea.Msg {
		if svc == nil {
			return messages.Toast{Message: "tmux cleanup unavailable", Level: messages.ToastWarning}
		}
		cleanedTagged, err := svc.KillSessionsMatchingTags(map[string]string{"@tumuxi": "1"}, opts)
		if err != nil {
			logging.Warn("Failed to cleanup tmux sessions by tag: %v", err)
		} else if cleanedTagged {
			logging.Info("Cleaned up @tumuxi tmux sessions")
		}
		prefix := tmux.SessionName("tumuxi") + "-"
		if err := svc.KillSessionsWithPrefix(prefix, opts); err != nil {
			return messages.Toast{Message: fmt.Sprintf("tmux cleanup failed: %v", err), Level: messages.ToastWarning}
		}
		if cleanedTagged {
			return messages.Toast{Message: fmt.Sprintf("Cleaned up @tumuxi and %s* tmux sessions", prefix), Level: messages.ToastSuccess}
		}
		return messages.Toast{Message: fmt.Sprintf("Cleaned up %s* tmux sessions", prefix), Level: messages.ToastSuccess}
	}
}

// CleanupTmuxOnExit is a no-op since sessions are always persisted across restarts.
func (a *App) CleanupTmuxOnExit() {
}
