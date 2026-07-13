package sync

import (
	"fmt"

	"github.com/chrisatdev/gf/internal/config"
	"github.com/chrisatdev/gf/internal/gitexec"
	"github.com/chrisatdev/gf/internal/gitrepo"
)

// injectable dependencies — override in tests.
var (
	inspectFn = gitrepo.Inspect
	runFn     = gitexec.Run
	linesFn   = gitexec.Lines
)

// Execute fetches origin, reports ahead/behind counts, and merges main when behind.
//
// Scenarios:
//
//	nil cfg                  → error: run 'gf -i'
//	behind == 0              → "Already up to date." + return nil
//	behind > 0, clean merge  → commit "merge: sync <main> into <branch>" + "Synced with <main>."
//	behind > 0, conflicts    → print file list + "Resolve conflicts then run 'gf resolve'." + error
func Execute(cfg *config.Config) error {
	if cfg == nil {
		return fmt.Errorf("No gf config found. Run 'gf -i' to initialize.")
	}

	if err := runFn("fetch", "origin"); err != nil {
		return fmt.Errorf("gf sync: %w", err)
	}

	state, err := inspectFn()
	if err != nil {
		return fmt.Errorf("gf sync: %w", err)
	}

	fmt.Printf("Your branch is %d commit(s) ahead and %d behind %s.\n",
		state.AheadCount, state.BehindCount, cfg.Repo.MainBranch)

	if state.BehindCount == 0 {
		fmt.Println("Already up to date.")
		return nil
	}

	target := "origin/" + cfg.Repo.MainBranch
	if mergeErr := runFn("merge", target, "--no-commit", "--no-ff"); mergeErr != nil {
		conflicts, lErr := linesFn("diff", "--name-only", "--diff-filter=U")
		if lErr == nil && len(conflicts) > 0 {
			for _, f := range conflicts {
				fmt.Println(f)
			}
			fmt.Println("Resolve conflicts then run 'gf resolve'.")
			return fmt.Errorf("gf sync: merge conflicts detected")
		}
		return fmt.Errorf("gf sync: %w", mergeErr)
	}

	msg := fmt.Sprintf("merge: sync %s into %s", cfg.Repo.MainBranch, state.Branch)
	if err := runFn("commit", "-m", msg); err != nil {
		return fmt.Errorf("gf sync: %w", err)
	}

	fmt.Printf("Synced with %s.\n", cfg.Repo.MainBranch)
	return nil
}
