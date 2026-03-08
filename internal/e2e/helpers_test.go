package e2e

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/tlepoid/tumuxi/internal/data"
	"github.com/tlepoid/tumuxi/internal/tmux"
)

func createWorkspaceFromDashboard(t *testing.T, session *PTYSession, name string) {
	t.Helper()
	if err := session.SendString("G"); err != nil {
		t.Fatalf("jump to create row: %v", err)
	}
	if err := session.SendString("\r"); err != nil {
		t.Fatalf("open create dialog: %v", err)
	}
	waitForUIContains(t, session, "Create Workspace", workspaceAgentTimeout)
	if err := session.SendString(name); err != nil {
		t.Fatalf("enter workspace name: %v", err)
	}
	if err := session.SendString("\r"); err != nil {
		t.Fatalf("confirm workspace name: %v", err)
	}
}

func waitForAgentSessions(t *testing.T, opts tmux.Options, timeout time.Duration) []string {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		sessions, err := tmux.ListSessionsMatchingTags(map[string]string{
			"@tumuxi":      "1",
			"@tumuxi_type": "agent",
		}, opts)
		if err == nil && len(sessions) > 0 {
			return sessions
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for agent sessions\n%s", tmuxSessionDebug(opts))
	return nil
}

func assertAgentSessionsStayLive(t *testing.T, opts tmux.Options, duration time.Duration) {
	t.Helper()
	deadline := time.Now().Add(duration)
	for time.Now().Before(deadline) {
		sessions, err := tmux.ListSessionsMatchingTags(map[string]string{
			"@tumuxi":      "1",
			"@tumuxi_type": "agent",
		}, opts)
		if err != nil {
			time.Sleep(200 * time.Millisecond)
			continue
		}
		if len(sessions) == 0 {
			t.Fatalf("expected at least one agent session to stay alive")
		}
		live := false
		for _, name := range sessions {
			state, err := tmux.SessionStateFor(name, opts)
			if err != nil {
				continue
			}
			if state.Exists && state.HasLivePane {
				live = true
				break
			}
		}
		if !live {
			t.Fatalf("agent sessions not live: %v", sessions)
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func assertScreenNeverContains(t *testing.T, session *PTYSession, needles []string, duration time.Duration) {
	t.Helper()
	deadline := time.Now().Add(duration)
	for time.Now().Before(deadline) {
		screen := session.ScreenASCII()
		for _, needle := range needles {
			if stringsContains(screen, needle) {
				t.Fatalf("unexpected screen text %q\n\nScreen:\n%s", needle, screen)
			}
		}
		time.Sleep(150 * time.Millisecond)
	}
}

func createSidebarTerminalTab(t *testing.T, session *PTYSession) {
	t.Helper()
	sendPrefixSequence(t, session, "t", "t")
	waitForUIContains(t, session, "Terminal 2", 10*time.Second)
}

func workspaceIDForRepo(repo string) string {
	ws := data.NewWorkspace("ws", "main", "main", repo, repo)
	return string(ws.ID())
}

func waitForTerminalSessionCount(t *testing.T, opts tmux.Options, wsID string, count int, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		sessions, err := tmux.ListSessionsMatchingTags(map[string]string{
			"@tumuxi":           "1",
			"@tumuxi_type":      "terminal",
			"@tumuxi_workspace": wsID,
		}, opts)
		if err == nil && len(sessions) == count {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for %d terminal sessions for workspace %s", count, wsID)
}

func waitForAssistantSessions(t *testing.T, opts tmux.Options, want map[string]bool, timeout time.Duration) map[string][]string {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		rows, err := tmux.SessionsWithTags(map[string]string{
			"@tumuxi":      "1",
			"@tumuxi_type": "agent",
		}, []string{"@tumuxi_assistant"}, opts)
		if err != nil {
			time.Sleep(200 * time.Millisecond)
			continue
		}
		byAssistant := make(map[string][]string)
		for _, row := range rows {
			assistant := strings.TrimSpace(row.Tags["@tumuxi_assistant"])
			if assistant == "" {
				continue
			}
			byAssistant[assistant] = append(byAssistant[assistant], row.Name)
		}
		ok := true
		for assistant := range want {
			if len(byAssistant[assistant]) == 0 {
				ok = false
				break
			}
		}
		if ok {
			return byAssistant
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for assistant sessions: %v\n%s", want, tmuxSessionDebug(opts))
	return nil
}

func tmuxSessionDebug(opts tmux.Options) string {
	rows, err := tmux.SessionsWithTags(map[string]string{}, []string{
		"@tumuxi",
		"@tumuxi_type",
		"@tumuxi_assistant",
		"@tumuxi_workspace",
		"@tumuxi_tab",
	}, opts)
	if err != nil {
		return fmt.Sprintf("tmux sessions: error=%v", err)
	}
	if len(rows) == 0 {
		return "tmux sessions: none"
	}
	lines := make([]string, 0, len(rows))
	for _, row := range rows {
		lines = append(lines, fmt.Sprintf(
			"%s tumuxi=%q type=%q assistant=%q workspace=%q tab=%q",
			row.Name,
			row.Tags["@tumuxi"],
			row.Tags["@tumuxi_type"],
			row.Tags["@tumuxi_assistant"],
			row.Tags["@tumuxi_workspace"],
			row.Tags["@tumuxi_tab"],
		))
	}
	return "tmux sessions:\n" + strings.Join(lines, "\n")
}
