package sidebar

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tlepoid/tumux/internal/logging"
	"github.com/tlepoid/tumux/internal/messages"
	"github.com/tlepoid/tumux/internal/ui/common"
	"github.com/tlepoid/tumux/internal/ui/compositor"
)

// Update handles messages.
func (m *LazygitModel) Update(msg tea.Msg) (*LazygitModel, tea.Cmd) {
	switch msg := msg.(type) {
	case LazygitStarted:
		return m, m.handleStarted(msg)

	case messages.LazygitPTYOutput:
		return m, m.handlePTYOutput(msg)

	case messages.LazygitPTYFlush:
		return m, m.handlePTYFlush(msg)

	case messages.LazygitPTYStopped:
		m.handlePTYStopped(msg)
		return m, nil

	case tea.KeyPressMsg:
		if !m.focused {
			return m, nil
		}
		m.mu.Lock()
		term := m.term
		vt := m.vt
		if vt != nil && vt.IsScrolled() {
			vt.ScrollViewToBottom()
		}
		m.mu.Unlock()
		if term != nil {
			input := common.KeyToBytes(msg)
			if len(input) > 0 {
				if err := term.SendString(string(input)); err != nil {
					logging.Warn("lazygit input failed: %v", err)
				}
			}
		}
		return m, nil

	case tea.PasteMsg:
		if !m.focused {
			return m, nil
		}
		m.mu.Lock()
		term := m.term
		m.mu.Unlock()
		if term != nil {
			text := "\x1b[200~" + msg.Content + "\x1b[201~"
			if err := term.SendString(text); err != nil {
				logging.Warn("lazygit paste failed: %v", err)
			}
		}
		return m, nil

	case tea.MouseWheelMsg:
		if !m.focused {
			return m, nil
		}
		m.mu.Lock()
		if m.vt != nil {
			delta := common.ScrollDeltaForHeight(m.vt.Height, 8)
			if msg.Button == tea.MouseWheelUp {
				m.vt.ScrollView(delta)
			} else if msg.Button == tea.MouseWheelDown {
				m.vt.ScrollView(-delta)
			}
		}
		m.mu.Unlock()
		return m, nil

	case tea.MouseClickMsg:
		if !m.focused {
			return m, nil
		}
		m.mu.Lock()
		term := m.term
		m.mu.Unlock()
		if term != nil {
			// Forward mouse click as escape sequence
			btn := 0
			switch msg.Button {
			case tea.MouseRight:
				btn = 2
			case tea.MouseMiddle:
				btn = 1
			}
			seq := fmt.Sprintf("\x1b[M%c%c%c", byte(32+btn), byte(32+msg.X+1), byte(32+msg.Y+1))
			_ = term.SendString(seq)
		}
		return m, nil
	}
	return m, nil
}

func (m *LazygitModel) handlePTYOutput(msg messages.LazygitPTYOutput) tea.Cmd {
	if msg.RunGen != m.runGen {
		return nil
	}
	m.pendingOutput = append(m.pendingOutput, msg.Data...)
	if len(m.pendingOutput) > lazygitMaxBufferedBytes {
		overflow := len(m.pendingOutput) - lazygitMaxBufferedBytes
		m.pendingOutput = append([]byte(nil), m.pendingOutput[overflow:]...)
	}
	m.lastOutputAt = time.Now()
	if !m.flushScheduled {
		m.flushScheduled = true
		m.flushPendingSince = m.lastOutputAt
		wsID := msg.WorkspaceID
		gen := msg.RunGen
		return common.SafeTick(lazygitFlushQuiet, func(time.Time) tea.Msg {
			return messages.LazygitPTYFlush{WorkspaceID: wsID, RunGen: gen}
		})
	}
	return nil
}

func (m *LazygitModel) handlePTYFlush(msg messages.LazygitPTYFlush) tea.Cmd {
	if msg.RunGen != m.runGen {
		return nil
	}
	now := time.Now()
	quietFor := now.Sub(m.lastOutputAt)
	pendingFor := time.Duration(0)
	if !m.flushPendingSince.IsZero() {
		pendingFor = now.Sub(m.flushPendingSince)
	}
	if quietFor < lazygitFlushQuiet && pendingFor < lazygitFlushMax {
		delay := lazygitFlushQuiet - quietFor
		if delay < time.Millisecond {
			delay = time.Millisecond
		}
		m.flushScheduled = true
		wsID := msg.WorkspaceID
		gen := msg.RunGen
		return common.SafeTick(delay, func(time.Time) tea.Msg {
			return messages.LazygitPTYFlush{WorkspaceID: wsID, RunGen: gen}
		})
	}
	m.flushScheduled = false
	m.flushPendingSince = time.Time{}
	if len(m.pendingOutput) > 0 {
		chunkSize := len(m.pendingOutput)
		if chunkSize > lazygitFlushChunkSize {
			chunkSize = lazygitFlushChunkSize
		}
		chunk := append([]byte(nil), m.pendingOutput[:chunkSize]...)
		copy(m.pendingOutput, m.pendingOutput[chunkSize:])
		m.pendingOutput = m.pendingOutput[:len(m.pendingOutput)-chunkSize]

		m.mu.Lock()
		if m.vt != nil {
			filtered := common.FilterKnownPTYNoiseStream(chunk, &m.ptyNoiseTrailing)
			if len(filtered) > 0 {
				m.vt.Write(filtered)
			}
			m.cachedSnap = nil
			m.cachedVersion = 0
		}
		m.mu.Unlock()

		if len(m.pendingOutput) > 0 {
			m.flushScheduled = true
			m.flushPendingSince = time.Now()
			wsID := msg.WorkspaceID
			gen := msg.RunGen
			return common.SafeTick(lazygitFlushQuiet, func(time.Time) tea.Msg {
				return messages.LazygitPTYFlush{WorkspaceID: wsID, RunGen: gen}
			})
		}
	}
	return nil
}

func (m *LazygitModel) handlePTYStopped(msg messages.LazygitPTYStopped) {
	if msg.RunGen != m.runGen {
		logging.Info("lazygit: PTY stopped (stale gen=%d current=%d) workspace=%s err=%v", msg.RunGen, m.runGen, msg.WorkspaceID, msg.Err)
		return
	}
	m.stopPTYReader()
	m.mu.Lock()
	m.running = false
	m.mu.Unlock()
	logging.Warn("lazygit: PTY stopped workspace=%s err=%v", msg.WorkspaceID, msg.Err)
}

// TerminalLayer returns a VTermLayer for compositor rendering.
func (m *LazygitModel) TerminalLayer() *compositor.VTermLayer {
	return m.TerminalLayerWithCursorOwner(true)
}

// TerminalLayerWithCursorOwner returns a VTermLayer while enforcing whether
// this pane currently owns cursor rendering.
func (m *LazygitModel) TerminalLayerWithCursorOwner(cursorOwner bool) *compositor.VTermLayer {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.vt == nil {
		return nil
	}
	version := m.vt.Version()
	showCursor := m.focused && cursorOwner
	if m.cachedSnap != nil && m.cachedVersion == version && m.cachedShowCursor == showCursor {
		return compositor.NewVTermLayer(m.cachedSnap)
	}
	snap := compositor.NewVTermSnapshot(m.vt, showCursor)
	if snap == nil {
		return nil
	}
	m.cachedSnap = snap
	m.cachedVersion = version
	m.cachedShowCursor = showCursor
	return compositor.NewVTermLayer(snap)
}

// View renders the lazygit pane.
func (m *LazygitModel) View() string {
	var b strings.Builder

	m.mu.Lock()
	vt := m.vt
	running := m.running
	if vt != nil {
		vt.ShowCursor = m.focused
	}
	m.mu.Unlock()

	if vt == nil {
		if m.workspace == nil {
			b.WriteString(m.styles.Muted.Render("No workspace selected"))
		} else if !running {
			b.WriteString(m.styles.Muted.Render("lazygit not running"))
		} else {
			b.WriteString(m.styles.Muted.Render("Starting lazygit..."))
		}
		return b.String()
	}

	m.mu.Lock()
	content := vt.Render()
	isScrolled := vt.IsScrolled()
	var scrollInfo string
	if isScrolled {
		offset, total := vt.GetScrollInfo()
		scrollInfo = fmt.Sprintf("%d/%d lines up", offset, total)
	}
	m.mu.Unlock()

	b.WriteString(content)
	if isScrolled {
		b.WriteString("\n")
		scrollStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(common.ColorBackground()).
			Background(common.ColorInfo())
		b.WriteString(scrollStyle.Render(" SCROLL: " + scrollInfo + " "))
	}

	result := b.String()
	if m.height > 0 {
		lines := strings.Split(result, "\n")
		if len(lines) > m.height {
			lines = lines[:m.height]
			result = strings.Join(lines, "\n")
		}
	}
	return result
}
