package platform

import "fmt"

// NoopPlatform performs a plain tracked push without any platform-specific
// behavior (no MR/PR creation, no browser open).
type NoopPlatform struct{}

// Name returns the platform identifier.
func (n *NoopPlatform) Name() string { return "noop" }

// Push pushes branch to the remote with upstream tracking set (-u origin).
func (n *NoopPlatform) Push(branch, title, mainBranch string, onlyPush bool) error {
	if err := gitRun("push", "-u", "origin", branch); err != nil {
		return fmt.Errorf("gf platform: %w", err)
	}
	return nil
}
