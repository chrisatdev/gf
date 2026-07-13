package merge

import (
	"fmt"
	"strings"

	"github.com/chrisatdev/gf/internal/config"
	"github.com/chrisatdev/gf/internal/gitexec"
	"github.com/chrisatdev/gf/internal/gitrepo"
)

// Execute merges the main branch into the current branch.
// Uses --no-commit --no-ff so the caller can inspect before committing.
func Execute(cfg *config.Config) error {
	current, err := gitrepo.CurrentBranch()
	if err != nil {
		return fmt.Errorf("gf merge: %w", err)
	}

	if err := gitexec.Run("fetch", "origin"); err != nil {
		return err
	}

	target := "origin/" + cfg.Repo.MainBranch
	if mergeErr := gitexec.Run("merge", target, "--no-commit", "--no-ff"); mergeErr != nil {
		conflicts, err := gitexec.Lines("diff", "--name-only", "--diff-filter=U")
		if err == nil && len(conflicts) > 0 {
			fmt.Println("Merge conflicts in:")
			for _, f := range conflicts {
				fmt.Printf("  %s\n", f)
			}
			return fmt.Errorf("merge conflicts in: %s. Resolve conflicts then run 'git merge --continue' or 'gf resolve'.", strings.Join(conflicts, ", "))
		}
		return mergeErr
	}

	fmt.Printf("Merged %s into %s.\n", cfg.Repo.MainBranch, current)
	return nil
}
