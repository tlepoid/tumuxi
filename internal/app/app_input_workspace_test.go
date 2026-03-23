package app

import (
	"testing"

	"github.com/tlepoid/tumux/internal/data"
	"github.com/tlepoid/tumux/internal/messages"
	"github.com/tlepoid/tumux/internal/ui/center"
	"github.com/tlepoid/tumux/internal/ui/dashboard"
	"github.com/tlepoid/tumux/internal/ui/sidebar"
)

func TestHandleWorkspaceDeletedClearsDirtyWorkspaceMarker(t *testing.T) {
	ws := data.NewWorkspace("feature", "feature", "main", "/repo", "/repo/feature")
	wsID := string(ws.ID())

	app := &App{
		dashboard:            dashboard.New(),
		center:               center.New(nil),
		sidebar:              sidebar.NewTabbedSidebar(),
		sidebarTerminal:      sidebar.NewTerminalModel(),
		dirtyWorkspaces:      map[string]bool{wsID: true},
		deletingWorkspaceIDs: map[string]bool{wsID: true},
	}

	app.handleWorkspaceDeleted(messages.WorkspaceDeleted{Workspace: ws})

	if app.isWorkspaceDeleteInFlight(wsID) {
		t.Fatal("expected delete-in-flight marker to be cleared on delete success")
	}
	if app.dirtyWorkspaces[wsID] {
		t.Fatal("expected dirty workspace marker to be cleared on delete success")
	}
}
