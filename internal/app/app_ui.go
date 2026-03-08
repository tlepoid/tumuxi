package app

import (
	tea "charm.land/bubbletea/v2"

	"github.com/tlepoid/tumuxi/internal/messages"
)

// focusPane changes focus to the specified pane
func (a *App) focusPane(pane messages.PaneType) tea.Cmd {
	a.focusedPane = pane
	// Keep focus transitions fail-safe for partially initialized App instances
	// used in lightweight tests.
	a.syncPaneFocusFlags()
	switch pane {
	case messages.PaneCenter:
		// Seamless UX: when center regains focus, attempt reattach for detached active tab.
		if a.center != nil {
			return a.center.ReattachActiveTabIfDetached()
		}
	case messages.PaneSidebarTerminal:
		// Lazy initialization: create terminal on focus if none exists.
		if a.sidebarTerminal != nil {
			return a.sidebarTerminal.EnsureTerminalTab()
		}
	}
	return nil
}

// focusPaneLeft moves focus one pane to the left, respecting layout visibility.
func (a *App) focusPaneLeft() tea.Cmd {
	switch a.focusedPane {
	case messages.PaneSidebar:
		if a.layout != nil && a.layout.ShowCenter() {
			return a.focusPane(messages.PaneCenter)
		}
		return a.focusPane(messages.PaneDashboard)
	case messages.PaneCenter, messages.PaneSidebarTerminal:
		return a.focusPane(messages.PaneDashboard)
	}
	return nil
}

// focusPaneDown moves focus from the agent pane to the terminal pane below it.
func (a *App) focusPaneDown() tea.Cmd {
	if a.focusedPane == messages.PaneCenter {
		return a.focusPane(messages.PaneSidebarTerminal)
	}
	return nil
}

// focusPaneUp moves focus from the terminal pane back to the agent pane above it.
func (a *App) focusPaneUp() tea.Cmd {
	if a.focusedPane == messages.PaneSidebarTerminal {
		return a.focusPane(messages.PaneCenter)
	}
	return nil
}

// focusPaneRight moves focus one pane to the right, respecting layout visibility.
func (a *App) focusPaneRight() tea.Cmd {
	switch a.focusedPane {
	case messages.PaneDashboard:
		if a.layout != nil && a.layout.ShowCenter() {
			return a.focusPane(messages.PaneCenter)
		}
		if a.layout != nil && a.layout.ShowSidebar() {
			return a.focusPane(messages.PaneSidebar)
		}
	case messages.PaneCenter, messages.PaneSidebarTerminal:
		if a.layout != nil && a.layout.ShowSidebar() {
			return a.focusPane(messages.PaneSidebar)
		}
	}
	return nil
}

// updateLayout updates component sizes based on window size
func (a *App) updateLayout() {
	leftGutter := a.layout.LeftGutter()
	topGutter := a.layout.TopGutter()

	// Left column: dashboard full height
	dashWidth := a.layout.DashboardWidth()
	a.dashboard.SetSize(dashWidth, a.layout.Height())

	// Center column: agent (top 3/4) + terminal (bottom 1/4)
	centerWidth := a.layout.CenterWidth()
	centerTopHeight, centerBottomHeight := centerPaneHeights(a.layout.Height())
	a.center.SetSize(centerWidth, centerTopHeight)
	gapX := 0
	if a.layout.ShowCenter() {
		gapX = a.layout.GapX()
	}
	centerX := leftGutter + dashWidth + gapX
	a.center.SetOffset(centerX) // Set X offset for mouse coordinate conversion
	a.center.SetCanFocusRight(a.layout.ShowSidebar())
	a.dashboard.SetCanFocusRight(a.layout.ShowCenter())

	termContentWidth := centerWidth - 4
	if termContentWidth < 1 {
		termContentWidth = 1
	}
	termContentHeight := centerBottomHeight - 2
	if termContentHeight < 1 {
		termContentHeight = 1
	}
	a.sidebarTerminal.SetSize(termContentWidth, termContentHeight)
	// Terminal offset: inside center column border+padding, below agent pane
	a.sidebarTerminal.SetOffset(centerX+2, topGutter+centerTopHeight+1)

	// Right sidebar: full height (no split)
	sidebarWidth := a.layout.SidebarWidth()
	sidebarContentWidth := sidebarWidth - 4
	if sidebarContentWidth < 1 {
		sidebarContentWidth = 1
	}
	sidebarContentHeight := a.layout.Height() - 2
	if sidebarContentHeight < 1 {
		sidebarContentHeight = 1
	}
	a.sidebar.SetSize(sidebarContentWidth, sidebarContentHeight)

	if a.dialog != nil {
		a.dialog.SetSize(a.width, a.height)
	}
	if a.filePicker != nil {
		a.filePicker.SetSize(a.width, a.height)
	}
	if a.settingsDialog != nil {
		a.settingsDialog.SetSize(a.width, a.height)
	}
}

func (a *App) setKeymapHintsEnabled(enabled bool) {
	if a.config != nil {
		a.config.UI.ShowKeymapHints = enabled
	}
	a.dashboard.SetShowKeymapHints(enabled)
	a.center.SetShowKeymapHints(enabled)
	a.sidebar.SetShowKeymapHints(enabled)
	a.sidebarTerminal.SetShowKeymapHints(enabled)
	if a.dialog != nil {
		a.dialog.SetShowKeymapHints(enabled)
	}
	if a.filePicker != nil {
		a.filePicker.SetShowKeymapHints(enabled)
	}
}

// centerPaneHeights splits the center column: ~3/4 for the agent, ~1/4 for the terminal.
func centerPaneHeights(total int) (int, int) {
	if total <= 0 {
		return 0, 0
	}
	bottom := total / 4
	if bottom < 3 {
		bottom = 3
	}
	top := total - bottom
	if top < 3 {
		top = 3
		bottom = total - top
		if bottom < 0 {
			bottom = 0
		}
	}
	return top, bottom
}
