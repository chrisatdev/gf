package platform

import (
	"fmt"
	"strings"
)

// GitLabPlatform handles push operations for GitLab repositories.
type GitLabPlatform struct{}

// Name returns the platform identifier.
func (g *GitLabPlatform) Name() string { return "gitlab" }

// Push pushes branch to GitLab.
//
// When onlyPush is true, a plain tracked push is performed.
// Otherwise, GitLab push options create an MR and the URL printed by GitLab
// in the push output is opened in the system browser.
func (g *GitLabPlatform) Push(branch, title, mainBranch string, onlyPush bool) error {
	if onlyPush {
		if err := gitRun("push", "-u", "origin", branch); err != nil {
			return fmt.Errorf("gf platform: %w", err)
		}
		return nil
	}
	out, err := gitCombinedOutput(
		"push", "-u", "origin", branch,
		"-o", "merge_request.create",
		"-o", "merge_request.title="+title,
		"-o", "merge_request.target="+mainBranch,
	)
	if err != nil {
		return fmt.Errorf("gf platform: %w", err)
	}
	if mrURL := parseRemoteURL(out); mrURL != "" {
		if err := openURL(mrURL); err != nil {
			return fmt.Errorf("gf platform: %w", err)
		}
	}
	return nil
}

// parseRemoteURL extracts the MR URL from git push remote output.
// GitLab prints the URL indented under "remote:" lines after the MR create push option.
func parseRemoteURL(output string) string {
	for _, line := range strings.Split(output, "\n") {
		rest := strings.TrimPrefix(line, "remote:")
		if rest == line {
			continue
		}
		rest = strings.TrimSpace(rest)
		if strings.HasPrefix(rest, "https://") || strings.HasPrefix(rest, "http://") {
			return rest
		}
	}
	return ""
}
