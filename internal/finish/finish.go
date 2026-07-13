package finish

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/chrisatdev/gf/internal/config"
	"github.com/chrisatdev/gf/internal/gitexec"
	"github.com/chrisatdev/gf/internal/gitrepo"
)

// Execute finishes the current branch:
// aborts if dirty, confirms with user, then deletes locally and remotely.
func Execute(cfg *config.Config, r io.Reader) error {
	state, err := gitrepo.Inspect()
	if err != nil {
		return fmt.Errorf("gf finish: %w", err)
	}

	if state.IsDirty {
		return fmt.Errorf("cannot finish: uncommitted changes. Stage and commit or stash first.")
	}

	branch := state.Branch
	if branch == cfg.Repo.MainBranch {
		return fmt.Errorf("cannot finish the main branch")
	}

	fmt.Printf("Finish branch %s? [y/N]: ", branch)
	scanner := bufio.NewScanner(r)
	if !scanner.Scan() {
		return fmt.Errorf("aborted")
	}
	if answer := strings.TrimSpace(scanner.Text()); answer != "y" && answer != "Y" {
		return fmt.Errorf("aborted")
	}

	if err := gitexec.Run("checkout", cfg.Repo.MainBranch); err != nil {
		return err
	}
	if err := gitexec.Run("branch", "-d", branch); err != nil {
		return err
	}
	if err := gitexec.Run("push", "origin", "--delete", branch); err != nil {
		return err
	}

	fmt.Printf("Branch %s deleted locally and remotely.\n", branch)
	return nil
}
