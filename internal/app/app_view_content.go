package app

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/tlepoid/tumuxi/internal/ui/common"
)

func (a *App) centerPaneStyle() lipgloss.Style {
	width := a.layout.CenterWidth()
	height := a.layout.Height()

	return lipgloss.NewStyle().
		Width(width-2).
		Height(height-2).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(common.ColorBorder()).
		Padding(0, 1)
}

// renderCenterPaneContent renders the center pane content when no tabs (raw content, no borders)
func (a *App) renderCenterPaneContent() string {
	if a.showWelcome {
		return a.renderWelcome()
	}

	if a.activeWorkspace != nil {
		return a.renderWorkspaceInfo()
	}

	return "Select a workspace from the dashboard"
}

func (a *App) centerPaneContentOrigin() (x, y int) {
	if a.layout == nil {
		return 0, 0
	}
	frameX, frameY := a.centerPaneStyle().GetFrameSize()
	gapX := 0
	if a.layout.ShowCenter() {
		gapX = a.layout.GapX()
	}
	return a.layout.LeftGutter() + a.layout.DashboardWidth() + gapX + frameX/2, a.layout.TopGutter() + frameY/2
}

func (a *App) goHome() {
	a.showWelcome = true
	a.activeWorkspace = nil
	if a.center != nil {
		a.center.SetWorkspace(nil)
	}
	if a.sidebar != nil {
		_ = a.sidebar.SetWorkspace(nil)
	}
	if a.sidebarTerminal != nil {
		_ = a.sidebarTerminal.SetWorkspace(nil)
	}
	if a.dashboard != nil {
		a.dashboard.ClearActiveRoot()
	}
	a.centerBtnFocused = false
	a.centerBtnIndex = 0
}

// renderWorkspaceInfo renders information about the active workspace
func (a *App) renderWorkspaceInfo() string {
	ws := a.activeWorkspace

	title := a.styles.Title.Render(ws.Name)
	content := title + "\n\n"
	content += fmt.Sprintf("Branch: %s\n", ws.Branch)
	content += fmt.Sprintf("Path: %s\n", ws.Root)

	if a.activeProject != nil {
		content += fmt.Sprintf("Project: %s\n", a.activeProject.Name)
	}

	activeStyle := lipgloss.NewStyle().Foreground(common.ColorForeground()).Bold(true)
	inactiveStyle := lipgloss.NewStyle().Foreground(common.ColorMuted())

	btnStyle := inactiveStyle
	if a.centerBtnFocused && a.centerBtnIndex == 0 {
		btnStyle = activeStyle
	}
	agentBtn := btnStyle.Render("[New agent]")
	content += "\n" + agentBtn
	if a.config.UI.ShowKeymapHints {
		content += "\n" + a.styles.Help.Render("C-Spc t a:new agent")
	}

	return content
}

// renderWelcome renders the welcome screen with the logo centered and buttons pinned to the bottom.
func (a *App) renderWelcome() string {
	width := a.layout.CenterWidth() - 4 // Account for borders/padding
	topHeight, _ := centerPaneHeights(a.layout.Height())
	height := topHeight - 2

	centerStyle := lipgloss.NewStyle().Width(width).AlignHorizontal(lipgloss.Center)

	logo, logoStyle := a.welcomeLogo()
	logoStr := centerStyle.Render(logoStyle.Render(logo))
	buttonsStr := centerStyle.Render(a.welcomeButtons())

	logoAreaHeight := height - lipgloss.Height(buttonsStr)
	if logoAreaHeight < 1 {
		logoAreaHeight = 1
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.Place(width, logoAreaHeight, lipgloss.Center, lipgloss.Center, logoStr),
		buttonsStr,
	)
}

func (a *App) welcomeButtons() string {
	activeStyle := lipgloss.NewStyle().Foreground(common.ColorForeground()).Bold(true)
	inactiveStyle := lipgloss.NewStyle().Foreground(common.ColorMuted())

	addProjectStyle := inactiveStyle
	settingsStyle := inactiveStyle
	if a.centerBtnFocused {
		if a.centerBtnIndex == 0 {
			addProjectStyle = activeStyle
		} else if a.centerBtnIndex == 1 {
			settingsStyle = activeStyle
		}
	}
	addProject := addProjectStyle.Render("[Add project]")
	settingsBtn := settingsStyle.Render("[Settings]")

	var b strings.Builder
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Center, addProject, "  ", settingsBtn))
	if a.config.UI.ShowKeymapHints {
		b.WriteString("\n")
		b.WriteString(a.styles.Help.Render("Dashboard: j/k to move • Enter to select"))
	}
	return b.String()
}

func (a *App) welcomeLogo() (string, lipgloss.Style) {
	logo := `
_________          _______                   _________
\__   __/|\     /|(       )|\     /||\     /|\__   __/
   ) (   | )   ( || () () || )   ( |( \   / )   ) (
   | |   | |   | || || || || |   | | \ (_) /    | |
   | |   | |   | || |(_)| || |   | |  ) _ (     | |
   | |   | |   | || |   | || |   | | / ( ) \    | |
   | |   | (___) || )   ( || (___) |( /   \ )___) (___
   )_(   (_______)|/     \|(_______)|/     \|\_______/`

	logoStyle := lipgloss.NewStyle().
		Foreground(common.ColorPrimary()).
		Bold(true)
	return logo, logoStyle
}
