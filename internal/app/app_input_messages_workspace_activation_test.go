package app

import (
	"testing"

	"github.com/tlepoid/tumux/internal/data"
	"github.com/tlepoid/tumux/internal/messages"
	"github.com/tlepoid/tumux/internal/ui/center"
	"github.com/tlepoid/tumux/internal/ui/dashboard"
	"github.com/tlepoid/tumux/internal/ui/layout"
	"github.com/tlepoid/tumux/internal/ui/sidebar"
)

func TestHandleWorkspaceActivated_AutoFocusCenterQueuesSingleReattach(t *testing.T) {
	ws := data.NewWorkspace("feature", "feature", "main", "/repo", "/repo")
	project := data.NewProject("/repo")
	project.AddWorkspace(*ws)

	centerModel := center.New(nil)
	centerModel.SetWorkspace(ws)
	centerModel.AddTab(&center.Tab{
		ID:          center.TabID("tab-1"),
		Name:        "Claude",
		Assistant:   "claude",
		SessionName: "tumux-test-session",
		Workspace:   ws,
		Detached:    true,
	})

	layoutManager := layout.NewManager()
	layoutManager.Resize(140, 40)

	app := &App{
		layout:          layoutManager,
		dashboard:       dashboard.New(),
		center:          centerModel,
		sidebar:         sidebar.NewTabbedSidebar(),
		sidebarTerminal: sidebar.NewTerminalModel(),
	}

	cmds := app.handleWorkspaceActivated(messages.WorkspaceActivated{
		Project:   project,
		Workspace: ws,
	})

	toastCount := 0
	for _, cmd := range cmds {
		if cmd == nil {
			continue
		}
		if toast, ok := cmd().(messages.Toast); ok && toast.Message == "Tab cannot be reattached" {
			toastCount++
		}
	}

	if toastCount != 1 {
		t.Fatalf("expected exactly one reattach toast command, got %d", toastCount)
	}
}
