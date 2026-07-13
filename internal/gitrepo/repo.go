package gitrepo

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/chrisatdev/gf/internal/gitexec"
)

type State struct {
	Branch      string
	IsDirty     bool
	AheadCount  int
	BehindCount int
	HasUpstream bool
	IsDetached  bool
}

// IsGitRepo returns true if the current working directory is inside a git repository.
func IsGitRepo() bool {
	return exec.Command("git", "rev-parse", "--git-dir").Run() == nil
}

// CurrentBranch returns the current branch name. Returns empty string if HEAD is detached.
func CurrentBranch() (string, error) {
	out, err := gitexec.Output("symbolic-ref", "--short", "HEAD")
	if err != nil {
		return "", fmt.Errorf("gf gitrepo: %w", err)
	}
	return out, nil
}

// Inspect returns the current worktree state.
func Inspect() (*State, error) {
	s := &State{}

	_, err := gitexec.Output("symbolic-ref", "HEAD")
	s.IsDetached = err != nil

	if !s.IsDetached {
		branch, err := gitexec.Output("symbolic-ref", "--short", "HEAD")
		if err != nil {
			return nil, fmt.Errorf("gf gitrepo: %w", err)
		}
		s.Branch = branch
	}

	out, err := gitexec.Output("status", "--porcelain")
	if err != nil {
		return nil, fmt.Errorf("gf gitrepo: %w", err)
	}
	s.IsDirty = out != ""

	upstream, err := gitexec.Output("rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}")
	s.HasUpstream = err == nil && upstream != ""

	if s.HasUpstream && !s.IsDetached {
		counts, err := gitexec.Output("rev-list", "--left-right", "--count", "@{u}...HEAD")
		if err == nil {
			parts := strings.Fields(counts)
			if len(parts) == 2 {
				s.BehindCount, _ = strconv.Atoi(parts[0])
				s.AheadCount, _ = strconv.Atoi(parts[1])
			}
		}
	}

	return s, nil
}
