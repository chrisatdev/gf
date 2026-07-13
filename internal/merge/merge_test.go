package merge_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chrisatdev/gf/internal/config"
	"github.com/chrisatdev/gf/internal/merge"
)

func TestExecute_cleanMerge(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	work := setupDivergedRepo(t, false)
	orig, _ := os.Getwd()
	must(t, os.Chdir(work))
	t.Cleanup(func() { os.Chdir(orig) })

	cfg := &config.Config{Repo: config.RepoConfig{MainBranch: "main"}}
	if err := merge.Execute(cfg); err != nil {
		t.Fatalf("Execute: %v", err)
	}
}

func TestExecute_conflictDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	work := setupDivergedRepo(t, true)
	orig, _ := os.Getwd()
	must(t, os.Chdir(work))
	t.Cleanup(func() { os.Chdir(orig) })

	cfg := &config.Config{Repo: config.RepoConfig{MainBranch: "main"}}
	err := merge.Execute(cfg)
	if err == nil {
		t.Fatal("expected conflict error")
	}
	if !strings.Contains(err.Error(), "merge conflicts") {
		t.Errorf("unexpected error: %v", err)
	}
}

// setupDivergedRepo creates a clone with a feature branch that diverges from main.
// If withConflict is true, both branches modify the same file to force a conflict.
func setupDivergedRepo(t *testing.T, withConflict bool) string {
	t.Helper()

	origin := t.TempDir()
	must(t, exec.Command("git", "init", "--bare", origin).Run())

	work := t.TempDir()
	must(t, exec.Command("git", "clone", origin, work).Run())
	must(t, exec.Command("git", "-C", work, "config", "user.email", "test@example.com").Run())
	must(t, exec.Command("git", "-C", work, "config", "user.name", "Test").Run())

	readme := filepath.Join(work, "README.md")
	must(t, os.WriteFile(readme, []byte("# initial\n"), 0644))
	must(t, exec.Command("git", "-C", work, "add", ".").Run())
	must(t, exec.Command("git", "-C", work, "commit", "-m", "init").Run())
	must(t, exec.Command("git", "-C", work, "push", "origin", "HEAD:main").Run())

	// create feature branch
	must(t, exec.Command("git", "-C", work, "checkout", "-b", "feat/diverge").Run())

	if withConflict {
		// feature branch modifies README
		must(t, os.WriteFile(readme, []byte("# feature branch\n"), 0644))
		must(t, exec.Command("git", "-C", work, "add", ".").Run())
		must(t, exec.Command("git", "-C", work, "commit", "-m", "feature change").Run())

		// main also modifies README (different content → conflict)
		must(t, exec.Command("git", "-C", work, "checkout", "main").Run())
		must(t, os.WriteFile(readme, []byte("# main change\n"), 0644))
		must(t, exec.Command("git", "-C", work, "add", ".").Run())
		must(t, exec.Command("git", "-C", work, "commit", "-m", "main change").Run())
		must(t, exec.Command("git", "-C", work, "push", "origin", "main").Run())

		must(t, exec.Command("git", "-C", work, "checkout", "feat/diverge").Run())
	} else {
		// feature branch adds a different file
		other := filepath.Join(work, "feature.txt")
		must(t, os.WriteFile(other, []byte("feature\n"), 0644))
		must(t, exec.Command("git", "-C", work, "add", ".").Run())
		must(t, exec.Command("git", "-C", work, "commit", "-m", "feature change").Run())

		// main adds an unrelated file
		must(t, exec.Command("git", "-C", work, "checkout", "main").Run())
		extra := filepath.Join(work, "extra.txt")
		must(t, os.WriteFile(extra, []byte("extra\n"), 0644))
		must(t, exec.Command("git", "-C", work, "add", ".").Run())
		must(t, exec.Command("git", "-C", work, "commit", "-m", "extra").Run())
		must(t, exec.Command("git", "-C", work, "push", "origin", "main").Run())

		must(t, exec.Command("git", "-C", work, "checkout", "feat/diverge").Run())
	}

	return work
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
