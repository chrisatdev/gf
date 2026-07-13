package finish_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chrisatdev/gf/internal/config"
	"github.com/chrisatdev/gf/internal/finish"
)

func TestExecute_dirtyTree(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	work := setupRepo(t)
	orig, _ := os.Getwd()
	must(t, os.Chdir(work))
	t.Cleanup(func() { os.Chdir(orig) })

	must(t, os.WriteFile(filepath.Join(work, "dirty.txt"), []byte("dirty"), 0644))

	cfg := &config.Config{Repo: config.RepoConfig{MainBranch: "main"}}
	err := finish.Execute(cfg, strings.NewReader("y\n"))
	if err == nil {
		t.Fatal("expected error for dirty tree")
	}
	if !strings.Contains(err.Error(), "uncommitted changes") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestExecute_mainBranch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	work := setupRepo(t)
	orig, _ := os.Getwd()
	must(t, os.Chdir(work))
	t.Cleanup(func() { os.Chdir(orig) })

	cfg := &config.Config{Repo: config.RepoConfig{MainBranch: "main"}}
	err := finish.Execute(cfg, strings.NewReader("y\n"))
	if err == nil {
		t.Fatal("expected error for main branch")
	}
	if !strings.Contains(err.Error(), "cannot finish the main branch") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestExecute_aborted(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	work := setupRepo(t)
	orig, _ := os.Getwd()
	must(t, os.Chdir(work))
	t.Cleanup(func() { os.Chdir(orig) })

	must(t, exec.Command("git", "checkout", "-b", "feat/my-feature").Run())

	cfg := &config.Config{Repo: config.RepoConfig{MainBranch: "main"}}
	err := finish.Execute(cfg, strings.NewReader("n\n"))
	if err == nil {
		t.Fatal("expected error for aborted finish")
	}
	if !strings.Contains(err.Error(), "aborted") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestExecute_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	work := setupRepo(t)
	orig, _ := os.Getwd()
	must(t, os.Chdir(work))
	t.Cleanup(func() { os.Chdir(orig) })

	must(t, exec.Command("git", "checkout", "-b", "feat/done").Run())
	must(t, exec.Command("git", "push", "--set-upstream", "origin", "feat/done").Run())

	cfg := &config.Config{Repo: config.RepoConfig{MainBranch: "main"}}
	if err := finish.Execute(cfg, strings.NewReader("y\n")); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	out, _ := exec.Command("git", "branch", "--list", "feat/done").Output()
	if strings.TrimSpace(string(out)) != "" {
		t.Error("branch feat/done still exists locally after finish")
	}

	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(out)) != "main" {
		t.Errorf("expected main, got %s", strings.TrimSpace(string(out)))
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
