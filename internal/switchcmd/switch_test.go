package switchcmd_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chrisatdev/gf/internal/switchcmd"
)

func TestExecute_byName(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	work := setupRepo(t)
	orig, _ := os.Getwd()
	must(t, os.Chdir(work))
	t.Cleanup(func() { os.Chdir(orig) })

	must(t, exec.Command("git", "checkout", "-b", "feat/test-branch").Run())
	must(t, exec.Command("git", "push", "--set-upstream", "origin", "feat/test-branch").Run())
	must(t, exec.Command("git", "checkout", "main").Run())

	if err := switchcmd.Execute("feat/test-branch", strings.NewReader("")); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(out)) != "feat/test-branch" {
		t.Errorf("expected feat/test-branch, got %s", strings.TrimSpace(string(out)))
	}
}

func TestExecute_selectByNumber(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	work := setupRepo(t)
	orig, _ := os.Getwd()
	must(t, os.Chdir(work))
	t.Cleanup(func() { os.Chdir(orig) })

	must(t, exec.Command("git", "checkout", "-b", "feat/pick-me").Run())
	must(t, exec.Command("git", "checkout", "main").Run())

	// branches listed alphabetically: feat/pick-me (1), main (2)
	if err := switchcmd.Execute("", strings.NewReader("1\n")); err != nil {
		t.Fatalf("Execute (interactive): %v", err)
	}

	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(out)) != "feat/pick-me" {
		t.Errorf("expected feat/pick-me, got %s", strings.TrimSpace(string(out)))
	}
}

func TestExecute_invalidSelection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	work := setupRepo(t)
	orig, _ := os.Getwd()
	must(t, os.Chdir(work))
	t.Cleanup(func() { os.Chdir(orig) })

	err := switchcmd.Execute("", strings.NewReader("999\n"))
	if err == nil {
		t.Fatal("expected error for out-of-range selection")
	}
	if !strings.Contains(err.Error(), "invalid selection") {
		t.Errorf("unexpected error: %v", err)
	}
}

func setupRepo(t *testing.T) string {
	t.Helper()

	origin := t.TempDir()
	must(t, exec.Command("git", "init", "--bare", origin).Run())

	work := t.TempDir()
	must(t, exec.Command("git", "clone", origin, work).Run())
	must(t, exec.Command("git", "-C", work, "config", "user.email", "test@example.com").Run())
	must(t, exec.Command("git", "-C", work, "config", "user.name", "Test").Run())

	readme := filepath.Join(work, "README.md")
	must(t, os.WriteFile(readme, []byte("# test"), 0644))
	must(t, exec.Command("git", "-C", work, "add", ".").Run())
	must(t, exec.Command("git", "-C", work, "commit", "-m", "init").Run())
	must(t, exec.Command("git", "-C", work, "push", "origin", "HEAD:main").Run())

	return work
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
