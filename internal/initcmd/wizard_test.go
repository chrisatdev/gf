package initcmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chrisatdev/gf/internal/config"
)

func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("setup %v: %v\n%s", args, err, out)
		}
	}
	run("git", "init")
	run("git", "config", "user.email", "test@test.com")
	run("git", "config", "user.name", "Test")
	return dir
}

func chdir(t *testing.T, dir string) {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(old) })
}

func TestRun_notGitRepo(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	err := Run(strings.NewReader(""))
	if err == nil {
		t.Fatal("expected error for non-git repo, got nil")
	}
	if !strings.Contains(err.Error(), "Not a git repository") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRun_existingConfigDeclined(t *testing.T) {
	dir := initRepo(t)
	if err := os.WriteFile(filepath.Join(dir, ".git", "gf.toml"), []byte("[repo]\nplatform = \"gitlab\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	chdir(t, dir)

	if err := Run(strings.NewReader("N\n")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// config should remain unchanged
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load() error: %v", err)
	}
	if cfg.Repo.Platform != "gitlab" {
		t.Errorf("config should be unchanged, platform = %q", cfg.Repo.Platform)
	}
}

func TestRun_happyPath(t *testing.T) {
	dir := initRepo(t)
	chdir(t, dir)

	// No remote → DetectRemote fails → prompts for platform
	// Input: platform=github, branch=main (default via empty), mfa=n
	input := "github\n\nn\n"
	if err := Run(strings.NewReader(input)); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load() after init: %v", err)
	}
	if cfg.Repo.Platform != "github" {
		t.Errorf("platform = %q, want %q", cfg.Repo.Platform, "github")
	}
	if cfg.Repo.MainBranch != "main" {
		t.Errorf("main_branch = %q, want %q", cfg.Repo.MainBranch, "main")
	}
	if cfg.Flow.MFAActive {
		t.Error("mfa_active should be false")
	}
}

func TestRun_reinitialize(t *testing.T) {
	dir := initRepo(t)
	if err := os.WriteFile(filepath.Join(dir, ".git", "gf.toml"), []byte("[repo]\nplatform = \"gitlab\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	chdir(t, dir)

	// Input: reinit=y, platform=github, branch=develop, mfa=y
	input := "y\ngithub\ndevelop\ny\n"
	if err := Run(strings.NewReader(input)); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load() after reinit: %v", err)
	}
	if cfg.Repo.Platform != "github" {
		t.Errorf("platform = %q, want %q", cfg.Repo.Platform, "github")
	}
	if cfg.Repo.MainBranch != "develop" {
		t.Errorf("main_branch = %q, want %q", cfg.Repo.MainBranch, "develop")
	}
	if !cfg.Flow.MFAActive {
		t.Error("mfa_active should be true")
	}
}
