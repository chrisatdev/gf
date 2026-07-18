package push

import (
	"fmt"
	"strings"

	"github.com/chrisatdev/gf/internal/changelog"
	"github.com/chrisatdev/gf/internal/commit"
	"github.com/chrisatdev/gf/internal/config"
	"github.com/chrisatdev/gf/internal/gitexec"
	"github.com/chrisatdev/gf/internal/gitrepo"
	"github.com/chrisatdev/gf/internal/platform"
)

// Options configures push command behavior.
type Options struct {
	Message  string // explicit commit message; empty means auto-infer
	OnlyPush bool   // skip MR/PR creation; use plain push on all platforms
}

// injectable dependencies — override in tests.
var (
	inspectRepo      = gitrepo.Inspect
	gitOutput        = gitexec.Output
	gitRun           = gitexec.Run
	newPlatform      = platform.New
	prependChangelog = changelog.Prepend
)

// Execute runs the push state machine.
//
// Scenarios:
//
//	nil cfg                          → error: run 'gf -i'
//	dirty tree                       → changelog + stage + commit + platform push
//	clean + ahead of upstream        → platform push only
//	clean + not ahead + has upstream → "Nothing to push."
//	dirty or no upstream             → dirty path sets -u origin via platform
func Execute(cfg *config.Config, opts Options) error {
	if cfg == nil {
		return fmt.Errorf("No gf config found. Run 'gf -i' to initialize.")
	}

	state, err := inspectRepo()
	if err != nil {
		return fmt.Errorf("gf push: %w", err)
	}

	branch := state.Branch

	// Never create MR/PR when pushing directly from the main branch
	if branch == cfg.Repo.MainBranch {
		opts.OnlyPush = true
	}

	// Only bail out when we have a confirmed upstream AND nothing is ahead/dirty.
	// When there is no upstream tracking, let the push attempt set it via -u.
	if !state.IsDirty && state.AheadCount == 0 && state.HasUpstream {
		return fmt.Errorf("Nothing to push.")
	}

	var mrTitle string

	if state.IsDirty {
		mrTitle, err = prepareCommit(branch, opts)
		if err != nil {
			return err
		}
	} else {
		// Clean + ahead (or no upstream): derive a title from the last commit
		// subject so the platform can use it as the MR/PR title.
		if t, logErr := gitOutput("log", "-1", "--pretty=%s"); logErr == nil && t != "" {
			mrTitle = t
		} else {
			mrTitle = branch
		}
	}

	plat := newPlatform(cfg)
	if pushErr := plat.Push(branch, mrTitle, cfg.Repo.MainBranch, opts.OnlyPush); pushErr != nil {
		return fmt.Errorf("gf push: %w", pushErr)
	}
	return nil
}

// prepareCommit stages files, prepends the changelog, and commits for a dirty
// working tree. The changelog is updated BEFORE git-add so it is folded into
// the main commit — no separate "docs: update changelog" commit is created.
//
// Returns the commit title (used as MR/PR title on the platform).
func prepareCommit(branch string, opts Options) (string, error) {
	var (
		mrTitle    string
		commitArgs []string
	)

	if opts.Message != "" {
		mrTitle = opts.Message
		commitArgs = []string{"commit", "-m", mrTitle}
	} else {
		porcelain, err := gitOutput("status", "--porcelain")
		if err != nil {
			return "", fmt.Errorf("gf push: %w", err)
		}
		msg := commit.InferFromBranch(branch, porcelain)
		mrTitle = commitHeader(msg)
		commitArgs = []string{"commit", "-m", mrTitle}
		if body := strings.Join(msg.Body, "\n"); body != "" {
			commitArgs = append(commitArgs, "-m", body)
		}
	}

	// Prepend changelog BEFORE staging so that it is included in git add --all.
	if err := prependChangelog("CHANGELOG.md", mrTitle); err != nil {
		return "", fmt.Errorf("gf push: %w", err)
	}

	// Stage all files (source changes + CHANGELOG.md).
	if err := gitRun("add", "--all"); err != nil {
		return "", fmt.Errorf("gf push: %w", err)
	}

	if err := gitRun(commitArgs...); err != nil {
		return "", fmt.Errorf("gf push: %w", err)
	}

	return mrTitle, nil
}

// commitHeader formats the first line of a conventional commit message.
func commitHeader(msg commit.CommitMessage) string {
	if msg.Scope != "" {
		return fmt.Sprintf("%s(%s): %s", msg.Type, msg.Scope, msg.Description)
	}
	return fmt.Sprintf("%s: %s", msg.Type, msg.Description)
}
