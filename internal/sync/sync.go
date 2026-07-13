package sync

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/chrisatdev/gf/internal/config"
	"github.com/chrisatdev/gf/internal/gitexec"
	"github.com/chrisatdev/gf/internal/gitrepo"
)

// injectable dependencies — override in tests.
var (
	inspectFn    = gitrepo.Inspect
	runFn        = gitexec.Run
	linesFn      = gitexec.Lines
	outputFn     = gitexec.Output
)

// aheadBehind returns (ahead, behind) commit counts of HEAD relative to target ref.
func aheadBehind(target string) (ahead, behind int, err error) {
	out, err := outputFn("rev-list", "--left-right", "--count", target+"...HEAD")
	if err != nil {
		return 0, 0, err
	}
	parts := strings.Fields(strings.TrimSpace(out))
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("unexpected rev-list output: %q", out)
	}
	b, err1 := strconv.Atoi(parts[0])
	a, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return 0, 0, fmt.Errorf("could not parse rev-list output: %q", out)
	}
	return a, b, nil
}

// Execute fetches origin, reports ahead/behind counts against origin/<main>, and merges main when behind.
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

	// Compute ahead/behind explicitly against origin/<main> (not @{u})
	target := "origin/" + cfg.Repo.MainBranch
	ahead, behind, err := aheadBehind(target)
	if err != nil {
		return fmt.Errorf("gf sync: %w", err)
	}

	fmt.Printf("Your branch is %d commit(s) ahead and %d behind %s.\n",
		ahead, behind, cfg.Repo.MainBranch)

	if behind == 0 {
		fmt.Println("Already up to date.")
		return nil
	}

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
