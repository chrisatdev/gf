package platform

import (
	"fmt"

	"github.com/chrisatdev/gf/internal/browser"
)

// GitHubPlatform handles push operations for GitHub repositories.
type GitHubPlatform struct {
	projectPath string // e.g. "owner/repo"
}

// Name returns the platform identifier.
func (g *GitHubPlatform) Name() string { return "github" }

// openURL is the function used to open compare URLs in the browser.
// Override in tests to capture calls without launching a real browser.
var openURL = browser.Open

// Push pushes branch to GitHub.
//
// A plain tracked push is always performed first. When onlyPush is false,
// the system browser is opened at the compare URL for the branch so the user
// can create a pull request.
func (g *GitHubPlatform) Push(branch, title, mainBranch string, onlyPush bool) error {
	if err := gitRun("push", "-u", "origin", branch); err != nil {
		return fmt.Errorf("gf platform: %w", err)
	}
	if onlyPush {
		return nil
	}
	url := fmt.Sprintf(
		"https://github.com/%s/compare/%s?expand=1",
		g.projectPath, branch,
	)
	if err := openURL(url); err != nil {
		return fmt.Errorf("gf platform: %w", err)
	}
	return nil
}
