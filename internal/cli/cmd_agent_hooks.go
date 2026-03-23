package cli

import (
	"github.com/tlepoid/tumux/internal/data"
	"github.com/tlepoid/tumux/internal/tmux"
)

// Test-seam variables: tests that mutate these must NOT use t.Parallel(),
// because they share this package-level mutable state.
var (
	tmuxSessionStateFor        = tmux.SessionStateFor
	tmuxKillSession            = tmux.KillSession
	tmuxSendKeys               = tmux.SendKeys
	tmuxSendInterrupt          = tmux.SendInterrupt
	tmuxSetSessionTag          = tmux.SetSessionTagValue
	tmuxCapturePaneTail        = tmux.CapturePaneTail
	tmuxStartSession           = tmuxNewSession
	startSendJobProcess        = launchSendJobProcessor
	appendWorkspaceOpenTabMeta = func(store *data.WorkspaceStore, wsID data.WorkspaceID, tab data.TabInfo) error {
		return store.AppendOpenTab(wsID, tab)
	}
)
