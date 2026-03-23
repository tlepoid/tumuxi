package app

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/tlepoid/tumux/internal/data"
	"github.com/tlepoid/tumux/internal/messages"
)

var githubSlugNonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

// fetchGitHubIssuesCmd returns a Cmd that lists open GitHub issues via the gh CLI.
func fetchGitHubIssuesCmd(project *data.Project) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("gh", "issue", "list",
			"--state", "open",
			"--json", "number,title,url,body",
			"--limit", "50",
		)
		if project != nil && project.Path != "" {
			cmd.Dir = project.Path
		}
		out, err := cmd.Output()
		if err != nil {
			return messages.GitHubIssuesLoaded{
				Project: project,
				Err:     fmt.Errorf("gh issue list: %w", err),
			}
		}
		var issues []*data.GitHubIssue
		if err := json.Unmarshal(out, &issues); err != nil {
			return messages.GitHubIssuesLoaded{
				Project: project,
				Err:     fmt.Errorf("parse gh output: %w", err),
			}
		}
		return messages.GitHubIssuesLoaded{Project: project, Issues: issues}
	}
}

// issueLabel formats a GitHub issue for display in the picker.
func issueLabel(issue *data.GitHubIssue) string {
	return fmt.Sprintf("#%d: %s", issue.Number, issue.Title)
}

// issueWorkspaceName generates a workspace name from an issue.
func issueWorkspaceName(issue *data.GitHubIssue) string {
	slug := githubSlugifyTitle(issue.Title)
	if slug == "" {
		return fmt.Sprintf("issue-%d", issue.Number)
	}
	const maxSlugLen = 40
	if len(slug) > maxSlugLen {
		slug = strings.TrimRight(slug[:maxSlugLen], "-")
	}
	return fmt.Sprintf("issue-%d-%s", issue.Number, slug)
}

func githubSlugifyTitle(title string) string {
	s := strings.ToLower(strings.TrimSpace(title))
	s = githubSlugNonAlnum.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}
