package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/chrisatdev/gf/internal/config"
)

func makeGitDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func chdirAndRestore(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })
}

func TestWriteLoadRoundtrip(t *testing.T) {
	dir := makeGitDir(t)
	chdirAndRestore(t, dir)

	want := &config.Config{
		Repo: config.RepoConfig{
			Platform:    "github",
			MainBranch:  "main",
			ProjectPath: "owner/repo",
		},
		Flow: config.FlowConfig{MFAActive: true},
	}

	if err := config.Write(want); err != nil {
		t.Fatalf("Write: %v", err)
	}

	got, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.Repo != want.Repo {
		t.Errorf("Repo mismatch: got %+v, want %+v", got.Repo, want.Repo)
	}
	if got.Flow != want.Flow {
		t.Errorf("Flow mismatch: got %+v, want %+v", got.Flow, want.Flow)
	}
}

func TestExistsReturnsFalseWithoutConfig(t *testing.T) {
	dir := makeGitDir(t)
	chdirAndRestore(t, dir)

	if config.Exists() {
		t.Error("Exists() should return false when config file has not been written")
	}
}

func TestExistsReturnsTrueAfterWrite(t *testing.T) {
	dir := makeGitDir(t)
	chdirAndRestore(t, dir)

	cfg := &config.Config{Repo: config.RepoConfig{MainBranch: "main"}}
	if err := config.Write(cfg); err != nil {
		t.Fatal(err)
	}

	if !config.Exists() {
		t.Error("Exists() should return true after Write")
	}
}

func TestFindGitDirWalksUpward(t *testing.T) {
	dir := makeGitDir(t)
	subdir := filepath.Join(dir, "subdir", "subdir2")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatal(err)
	}

	chdirAndRestore(t, subdir)

	gitDir, err := config.FindGitDir()
	if err != nil {
		t.Fatalf("FindGitDir: %v", err)
	}

	expected := filepath.Join(dir, ".git")
	if gitDir != expected {
		t.Errorf("FindGitDir = %q, want %q", gitDir, expected)
	}
}
