package cli

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/tlepoid/tumux/internal/data"
)

// githubIssueResult wraps a fetched issue so callers can distinguish "no issue requested" from "issue fetched".
type githubIssueResult struct {
	issue *data.GitHubIssue
}

var slugNonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

// fetchGitHubIssue retrieves issue metadata from GitHub using the gh CLI.
func fetchGitHubIssue(issueNumber int) (*data.GitHubIssue, error) {
	out, err := exec.Command("gh", "issue", "view", strconv.Itoa(issueNumber),
		"--json", "number,title,url,body").Output()
	if err != nil {
		return nil, fmt.Errorf("gh issue view %d: %w", issueNumber, err)
	}
	var issue data.GitHubIssue
	if err := json.Unmarshal(out, &issue); err != nil {
		return nil, fmt.Errorf("parse gh output: %w", err)
	}
	return &issue, nil
}

// issueDefaultName returns a short workspace name derived from the issue number.
func issueDefaultName(issue *data.GitHubIssue) string {
	slug := slugifyTitle(issue.Title)
	if slug == "" {
		return fmt.Sprintf("issue-%d", issue.Number)
	}
	// Keep slug reasonably short for branch names.
	const maxSlugLen = 40
	if len(slug) > maxSlugLen {
		slug = strings.TrimRight(slug[:maxSlugLen], "-")
	}
	return fmt.Sprintf("issue-%d-%s", issue.Number, slug)
}

func slugifyTitle(title string) string {
	s := strings.ToLower(strings.TrimSpace(title))
	s = slugNonAlnum.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}
