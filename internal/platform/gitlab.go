package platform

import "fmt"

// GitLabPlatform handles push operations for GitLab repositories.
type GitLabPlatform struct{}

// Name returns the platform identifier.
func (g *GitLabPlatform) Name() string { return "gitlab" }

// Push pushes branch to GitLab.
//
// When onlyPush is true, a plain tracked push is performed.
// Otherwise, GitLab push options are added to open a merge request targeting
// mainBranch with the given title.
func (g *GitLabPlatform) Push(branch, title, mainBranch string, onlyPush bool) error {
	if onlyPush {
		if err := gitRun("push", "-u", "origin", branch); err != nil {
			return fmt.Errorf("gf platform: %w", err)
		}
		return nil
	}
	if err := gitRun(
		"push", "-u", "origin", branch,
		"-o", "merge_request.create",
		"-o", "merge_request.title="+title,
		"-o", "merge_request.target="+mainBranch,
	); err != nil {
		return fmt.Errorf("gf platform: %w", err)
	}
	return nil
}
