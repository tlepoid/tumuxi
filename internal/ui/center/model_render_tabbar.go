package center

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tlepoid/tumux/internal/messages"
	"github.com/tlepoid/tumux/internal/ui/common"
)

// renderTabBar renders the tab bar with activity indicators
func (m *Model) renderTabBar() string {
	m.tabHits = m.tabHits[:0]
	currentTabs := m.getTabs()
	activeIdx := m.getActiveTabIdx()

	if len(currentTabs) == 0 {
		empty := m.styles.TabPlus.Render("New agent")
		emptyWidth := lipgloss.Width(empty)
		if emptyWidth > 0 {
			m.tabHits = append(m.tabHits, tabHit{
				kind:  tabHitPlus,
				index: -1,
				region: common.HitRegion{
					X:      0,
					Y:      0,
					Width:  emptyWidth,
					Height: 1,
				},
			})
		}
		return empty
	}

	var renderedTabs []string
	x := 0

	for i, tab := range currentTabs {
		name := tab.Name
		if name == "" {
			name = tab.Assistant
		}

		// Determine conversation status for chat tabs
		var indicator string
		isChat := m.isChatTab(tab)
		var status ConversationStatus
		if isChat {
			status = m.TabConversationStatus(tab)
			// Upgrade Waiting→Running when the tab has recent visible output.
			if status == ConvStatusWaiting && m.IsTabActive(tab) {
				status = ConvStatusRunning
			}
			switch status {
			case ConvStatusRunning:
				indicator = common.Icons.Running + " "
			case ConvStatusWaiting:
				indicator = common.Icons.Waiting + " "
			case ConvStatusError:
				indicator = common.Icons.Error + " "
			case ConvStatusComplete:
				indicator = common.Icons.Complete + " "
			default: // ConvStatusIdle
				indicator = common.Icons.Idle + " "
			}
		}

		agentStyle := lipgloss.NewStyle().Foreground(common.AgentColor(tab.Assistant))

		// indicatorColor returns the status-appropriate foreground color for the indicator.
		indicatorColor := func() interface {
			RGBA() (uint32, uint32, uint32, uint32)
		} {
			switch status {
			case ConvStatusRunning:
				return common.AgentColor(tab.Assistant)
			case ConvStatusWaiting:
				return common.ColorWarning()
			case ConvStatusError:
				return common.ColorError()
			case ConvStatusComplete:
				return common.ColorInfo()
			default:
				return common.ColorMuted()
			}
		}

		// Build tab content with close affordance
		closeLabel := m.styles.Muted.Render("×")
		var rendered string
		var style lipgloss.Style
		if i == activeIdx {
			// Active tab - each part styled with same background
			bg := common.ColorSurface2()
			pad := lipgloss.NewStyle().Background(bg).Render(" ")
			indicatorPart := lipgloss.NewStyle().Foreground(indicatorColor()).Background(bg).Render(indicator)
			// Name color reflects status: running→primary+bold, waiting→warning, error→error, idle→muted
			nameStyle := lipgloss.NewStyle().Background(bg)
			switch status {
			case ConvStatusRunning:
				nameStyle = nameStyle.Foreground(common.ColorPrimary()).Bold(true)
			case ConvStatusWaiting:
				nameStyle = nameStyle.Foreground(common.ColorWarning())
			case ConvStatusError:
				nameStyle = nameStyle.Foreground(common.ColorError())
			case ConvStatusComplete:
				nameStyle = nameStyle.Foreground(common.ColorInfo())
			default:
				nameStyle = nameStyle.Foreground(common.ColorMuted())
			}
			if !isChat {
				nameStyle = nameStyle.Foreground(common.ColorForeground())
			}
			namePart := nameStyle.Render(name)
			space := lipgloss.NewStyle().Background(bg).Render(" ")
			closePart := lipgloss.NewStyle().Foreground(common.ColorMuted()).Background(bg).Render("×")
			rendered = pad + indicatorPart + namePart + space + closePart + pad
			style = m.styles.ActiveTab
		} else {
			// Inactive tab
			var nameStyled string
			switch status {
			case ConvStatusRunning:
				nameStyled = lipgloss.NewStyle().Foreground(common.ColorPrimary()).Bold(true).Render(name)
			case ConvStatusWaiting:
				nameStyled = lipgloss.NewStyle().Foreground(common.ColorWarning()).Render(name)
			case ConvStatusError:
				nameStyled = lipgloss.NewStyle().Foreground(common.ColorError()).Render(name)
			case ConvStatusComplete:
				nameStyled = lipgloss.NewStyle().Foreground(common.ColorInfo()).Render(name)
			default:
				nameStyled = m.styles.Muted.Render(name)
			}
			if !isChat {
				nameStyled = m.styles.Muted.Render(name)
			}
			indicatorStyled := lipgloss.NewStyle().Foreground(indicatorColor()).Render(indicator)
			content := indicatorStyled + nameStyled + " " + closeLabel
			rendered = m.styles.Tab.Render(content)
			style = m.styles.Tab
		}
		renderedWidth := lipgloss.Width(rendered)
		if renderedWidth > 0 {
			m.tabHits = append(m.tabHits, tabHit{
				kind:  tabHitTab,
				index: i,
				region: common.HitRegion{
					X:      x,
					Y:      0,
					Width:  renderedWidth,
					Height: 1,
				},
			})

			frameX, _ := style.GetFrameSize()
			leftFrame := frameX / 2
			prefixWidth := lipgloss.Width(agentStyle.Render(indicator) + name + " ")
			closeWidth := lipgloss.Width(closeLabel)
			closeX := x + leftFrame + prefixWidth
			if closeWidth > 0 {
				// Expand close button hit region for easier clicking
				expandedCloseX := closeX - 1
				expandedCloseWidth := renderedWidth - leftFrame - prefixWidth + 1
				m.tabHits = append(m.tabHits, tabHit{
					kind:  tabHitClose,
					index: i,
					region: common.HitRegion{
						X:      expandedCloseX,
						Y:      0,
						Width:  expandedCloseWidth,
						Height: 1,
					},
				})
			}
		}
		x += renderedWidth
		renderedTabs = append(renderedTabs, rendered)
	}

	// Add control buttons with matching border style
	btn := m.styles.TabPlus.Render("+ New")
	btnWidth := lipgloss.Width(btn)
	if btnWidth > 0 {
		m.tabHits = append(m.tabHits, tabHit{
			kind:  tabHitPlus,
			index: -1,
			region: common.HitRegion{
				X:      x,
				Y:      0,
				Width:  btnWidth,
				Height: 1,
			},
		})
	}
	renderedTabs = append(renderedTabs, btn)

	// Join tabs horizontally at the bottom so borders align
	return lipgloss.JoinHorizontal(lipgloss.Bottom, renderedTabs...)
}

func (m *Model) handleTabBarClick(msg tea.MouseClickMsg) tea.Cmd {
	// Tab bar is at screen Y=1: Y=0 is pane border, Y=1 is tab content (compact, no tab border)
	// Account for border (1) and padding (1) on the left side when converting X coordinates
	const (
		borderTop   = 1
		borderLeft  = 1
		paddingLeft = 1
	)
	if msg.Y != borderTop {
		return nil
	}
	// Convert screen X to content X (subtract pane offset, border, and padding)
	localX := msg.X - m.offsetX - borderLeft - paddingLeft
	if localX < 0 {
		return nil
	}
	// Convert screen Y to local Y within tab bar content (all tab hits are at Y=0)
	localY := msg.Y - borderTop
	// Check close buttons first (they overlap with tab regions)
	for _, hit := range m.tabHits {
		if hit.kind == tabHitClose && hit.region.Contains(localX, localY) {
			return m.closeTabAt(hit.index)
		}
	}
	// Then check tabs and other buttons
	for _, hit := range m.tabHits {
		if hit.region.Contains(localX, localY) {
			switch hit.kind {
			case tabHitPlus:
				return func() tea.Msg { return messages.ShowSelectAssistantDialog{} }
			case tabHitTab:
				before := m.getActiveTabIdx()
				m.setActiveTabIdx(hit.index)
				return m.tabSelectionChangedCmd(hit.index != before)
			}
		}
	}
	return nil
}
