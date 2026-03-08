package sidebar

import (
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/andyrewlee/amux/internal/data"
	"github.com/andyrewlee/amux/internal/logging"
	"github.com/andyrewlee/amux/internal/pty"
	"github.com/andyrewlee/amux/internal/ui/common"
	"github.com/andyrewlee/amux/internal/ui/compositor"
	"github.com/andyrewlee/amux/internal/vterm"
)

const (
	lazygitFlushQuiet       = 20 * time.Millisecond
	lazygitFlushMax         = 80 * time.Millisecond
	lazygitFlushChunkSize   = 32 * 1024
	lazygitReadBufferSize   = 32 * 1024
	lazygitReadQueueSize    = 32
	lazygitFrameInterval    = time.Second / 24
	lazygitMaxPendingBytes  = 256 * 1024
	lazygitMaxBufferedBytes = 4 * 1024 * 1024
)

// LazygitModel is the Bubbletea model for the lazygit sidebar pane.
type LazygitModel struct {
	workspace       *data.Workspace
	focused         bool
	width           int
	height          int
	styles          common.Styles
	showKeymapHints bool

	msgSink func(tea.Msg)

	mu               sync.Mutex
	term             *pty.Terminal
	vt               *vterm.VTerm
	running          bool
	lastWidth        int
	lastHeight       int
	ptyNoiseTrailing []byte

	// PTY buffering
	pendingOutput     []byte
	flushScheduled    bool
	lastOutputAt      time.Time
	flushPendingSince time.Time

	// PTY reader state
	readerActive bool
	readerCancel chan struct{}
	ptyMsgCh     chan tea.Msg
	ptyHeartbeat int64
	runGen       uint64

	// Snapshot cache
	cachedSnap       *compositor.VTermSnapshot
	cachedVersion    uint64
	cachedShowCursor bool
}

// NewLazygitModel creates a new lazygit model.
func NewLazygitModel() *LazygitModel {
	return &LazygitModel{
		styles: common.DefaultStyles(),
	}
}

// SetShowKeymapHints controls whether helper text is rendered.
func (m *LazygitModel) SetShowKeymapHints(show bool) {
	m.showKeymapHints = show
}

// SetStyles updates the component's styles (for theme changes).
func (m *LazygitModel) SetStyles(styles common.Styles) {
	m.styles = styles
}

// SetMsgSink sets a callback for PTY messages.
func (m *LazygitModel) SetMsgSink(sink func(tea.Msg)) {
	m.msgSink = sink
}

// Init initializes the lazygit model.
func (m *LazygitModel) Init() tea.Cmd {
	return nil
}

// Focus sets focus.
func (m *LazygitModel) Focus() {
	m.focused = true
	m.mu.Lock()
	if m.vt != nil {
		m.vt.ShowCursor = true
	}
	m.cachedSnap = nil
	m.cachedVersion = 0
	m.mu.Unlock()
}

// Blur removes focus.
func (m *LazygitModel) Blur() {
	m.focused = false
	m.mu.Lock()
	if m.vt != nil {
		m.vt.ShowCursor = false
	}
	m.cachedSnap = nil
	m.cachedVersion = 0
	m.mu.Unlock()
}

// Focused returns whether the pane is focused.
func (m *LazygitModel) Focused() bool {
	return m.focused
}

// SetWorkspace sets the active workspace and restarts lazygit if needed.
func (m *LazygitModel) SetWorkspace(ws *data.Workspace) tea.Cmd {
	if sameWorkspaceByCanonicalPaths(m.workspace, ws) {
		m.workspace = ws
		return nil
	}
	m.stopLazygit()
	m.workspace = ws
	if ws == nil {
		logging.Info("lazygit: workspace cleared")
		return nil
	}
	logging.Info("lazygit: starting for workspace=%s root=%s size=%dx%d", ws.ID(), ws.Root, m.width, m.height)
	return m.startLazygit()
}

// SetSize sets the component size.
func (m *LazygitModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	termWidth, termHeight := m.contentSize()
	m.mu.Lock()
	if m.vt != nil && (m.lastWidth != termWidth || m.lastHeight != termHeight) {
		m.lastWidth = termWidth
		m.lastHeight = termHeight
		m.vt.Resize(termWidth, termHeight)
		if m.term != nil {
			_ = m.term.SetSize(uint16(termHeight), uint16(termWidth))
		}
	}
	m.mu.Unlock()
}

// contentSize returns the usable terminal dimensions.
func (m *LazygitModel) contentSize() (int, int) {
	w := m.width
	h := m.height
	if w < 10 {
		w = 10
	}
	if h < 3 {
		h = 3
	}
	return w, h
}

// Close shuts down the lazygit PTY.
func (m *LazygitModel) Close() {
	m.stopLazygit()
}
