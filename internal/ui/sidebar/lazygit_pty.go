package sidebar

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/tlepoid/tumuxi/internal/logging"
	"github.com/tlepoid/tumuxi/internal/messages"
	"github.com/tlepoid/tumuxi/internal/pty"
	"github.com/tlepoid/tumuxi/internal/safego"
	"github.com/tlepoid/tumuxi/internal/ui/common"
	"github.com/tlepoid/tumuxi/internal/vterm"
)

// lazygitCommand returns the shell command used to launch lazygit, including
// a --use-config-file flag that merges the user's existing config with an
// tumuxi-generated theme overlay (comma-separated, later files win).
// Falls back to plain "lazygit" if writing the overlay fails.
func lazygitCommand() string {
	hex := common.HexColor
	yaml := fmt.Sprintf(`gui:
  nerdFontsVersion: "3"
  theme:
    activeBorderColor:
      - '%s'
      - bold
    inactiveBorderColor:
      - '%s'
    searchingActiveBorderColor:
      - '%s'
      - bold
    optionsTextColor:
      - '%s'
    selectedLineBgColor:
      - '%s'
    selectedRangeBgColor:
      - '%s'
    cherryPickedCommitBgColor:
      - '%s'
    cherryPickedCommitFgColor:
      - '%s'
    unstagedChangesColor:
      - '%s'
    defaultFgColor:
      - '%s'
`,
		hex(common.ColorBorderFocused()),
		hex(common.ColorBorder()),
		hex(common.ColorInfo()),
		hex(common.ColorSecondary()),
		hex(common.ColorSelection()),
		hex(common.ColorSelection()),
		hex(common.ColorSurface2()),
		hex(common.ColorPrimary()),
		hex(common.ColorError()),
		hex(common.ColorForeground()),
	)

	themePath := filepath.Join(os.TempDir(), "tumuxi-lazygit-theme.yml")
	if err := os.WriteFile(themePath, []byte(yaml), 0o600); err != nil {
		logging.Warn("lazygit: could not write theme config: %v", err)
		return "lazygit"
	}

	// Build comma-separated config list: user config (if any) then our overlay.
	var configs []string
	if existing := os.Getenv("LG_CONFIG_FILE"); existing != "" {
		configs = append(configs, existing)
	} else if configDir, err := os.UserConfigDir(); err == nil {
		defaultCfg := filepath.Join(configDir, "lazygit", "config.yml")
		if _, err := os.Stat(defaultCfg); err == nil {
			configs = append(configs, defaultCfg)
		}
	}
	configs = append(configs, themePath)

	return "lazygit --use-config-file " + strings.Join(configs, ",")
}

// startLazygit launches lazygit in a PTY.
func (m *LazygitModel) startLazygit() tea.Cmd {
	ws := m.workspace
	if ws == nil {
		return nil
	}
	termWidth, termHeight := m.contentSize()
	wsID := string(ws.ID())
	root := ws.Root

	m.runGen++
	gen := m.runGen

	return func() tea.Msg {
		env := []string{"COLORTERM=truecolor"}
		cmd := lazygitCommand()
		logging.Info("lazygit: launching PTY workspace=%s root=%s size=%dx%d gen=%d", wsID, root, termWidth, termHeight, gen)
		term, err := pty.NewWithSize(cmd, root, env, uint16(termHeight), uint16(termWidth))
		if err != nil {
			logging.Warn("lazygit: PTY launch failed: %v", err)
			return LazygitStarted{WorkspaceID: wsID, RunGen: gen, Err: err}
		}
		logging.Info("lazygit: PTY launched OK workspace=%s gen=%d", wsID, gen)
		return LazygitStarted{WorkspaceID: wsID, RunGen: gen, Terminal: term}
	}
}

// LazygitStarted is returned when lazygit starts (or fails to start).
type LazygitStarted struct {
	WorkspaceID string
	RunGen      uint64
	Terminal    *pty.Terminal
	Err         error
}

// handleStarted processes a LazygitStarted message.
func (m *LazygitModel) handleStarted(msg LazygitStarted) tea.Cmd {
	if msg.RunGen != m.runGen {
		// Stale - workspace or gen changed before lazygit started
		if msg.Terminal != nil {
			msg.Terminal.Close()
		}
		return nil
	}
	if msg.Err != nil {
		logging.Warn("lazygit failed to start: %v", msg.Err)
		return nil
	}

	termWidth, termHeight := m.contentSize()
	term := msg.Terminal

	vt := vterm.New(termWidth, termHeight)
	vt.AllowAltScreenScrollback = true
	vt.ShowCursor = m.focused
	vt.SetResponseWriter(func(data []byte) {
		if term != nil {
			if _, err := term.Write(data); err != nil {
				logging.Debug("lazygit vterm response write failed: %v", err)
			}
		}
	})

	if err := term.SetSize(uint16(termHeight), uint16(termWidth)); err != nil {
		logging.Debug("lazygit initial resize failed: %v", err)
	}

	m.mu.Lock()
	m.term = term
	m.vt = vt
	m.running = true
	m.lastWidth = termWidth
	m.lastHeight = termHeight
	m.pendingOutput = m.pendingOutput[:0]
	m.cachedSnap = nil
	m.cachedVersion = 0
	m.mu.Unlock()

	return m.startPTYReader(msg.WorkspaceID, msg.RunGen)
}

// stopLazygit stops the running lazygit PTY.
func (m *LazygitModel) stopLazygit() {
	m.stopPTYReader()
	m.mu.Lock()
	term := m.term
	m.term = nil
	m.running = false
	m.vt = nil
	m.pendingOutput = m.pendingOutput[:0]
	m.cachedSnap = nil
	m.cachedVersion = 0
	m.mu.Unlock()
	// Reset flush state so the next workspace's output gets a fresh flush schedule.
	m.flushScheduled = false
	m.flushPendingSince = time.Time{}
	if term != nil {
		term.Close()
	}
}

// startPTYReader starts the goroutine that reads from the lazygit PTY.
func (m *LazygitModel) startPTYReader(wsID string, gen uint64) tea.Cmd {
	m.mu.Lock()
	if m.readerActive {
		m.mu.Unlock()
		return nil
	}
	if m.term == nil || !m.running {
		m.mu.Unlock()
		return nil
	}
	if m.readerCancel != nil {
		common.SafeClose(m.readerCancel)
	}
	m.readerCancel = make(chan struct{})
	m.ptyMsgCh = make(chan tea.Msg, lazygitReadQueueSize)
	m.readerActive = true
	atomic.StoreInt64(&m.ptyHeartbeat, time.Now().UnixNano())
	term := m.term
	cancel := m.readerCancel
	msgCh := m.ptyMsgCh
	m.mu.Unlock()

	sink := m.msgSink

	safego.Go("sidebar.lazygit_reader", func() {
		defer m.markPTYReaderStopped()
		common.RunPTYReader(term, msgCh, cancel, &m.ptyHeartbeat, common.PTYReaderConfig{
			Label:           "sidebar.lazygit_read_loop",
			ReadBufferSize:  lazygitReadBufferSize,
			ReadQueueSize:   lazygitReadQueueSize,
			FrameInterval:   lazygitFrameInterval,
			MaxPendingBytes: lazygitMaxPendingBytes,
		}, common.PTYMsgFactory{
			Output: func(data []byte) tea.Msg {
				return messages.LazygitPTYOutput{WorkspaceID: wsID, RunGen: gen, Data: data}
			},
			Stopped: func(err error) tea.Msg {
				return messages.LazygitPTYStopped{WorkspaceID: wsID, RunGen: gen, Err: err}
			},
		})
	})
	safego.Go("sidebar.lazygit_forward", func() {
		m.forwardPTYMsgs(msgCh, sink)
	})
	return nil
}

func (m *LazygitModel) forwardPTYMsgs(msgCh <-chan tea.Msg, sink func(tea.Msg)) {
	for msg := range msgCh {
		if sink != nil {
			sink(msg)
		}
	}
}

func (m *LazygitModel) stopPTYReader() {
	m.mu.Lock()
	if m.readerCancel != nil {
		common.SafeClose(m.readerCancel)
		m.readerCancel = nil
	}
	m.readerActive = false
	m.ptyMsgCh = nil
	m.mu.Unlock()
	atomic.StoreInt64(&m.ptyHeartbeat, 0)
}

func (m *LazygitModel) markPTYReaderStopped() {
	m.mu.Lock()
	m.readerActive = false
	m.ptyMsgCh = nil
	m.mu.Unlock()
	atomic.StoreInt64(&m.ptyHeartbeat, 0)
}
