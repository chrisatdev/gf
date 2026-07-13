package gitrepo

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
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

func seedCommit(t *testing.T, dir, file, msg string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, file), []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("seedCommit %v: %v\n%s", args, err, out)
		}
	}
	run("git", "add", file)
	run("git", "commit", "-m", msg)
}

func TestIsGitRepo_inside(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	dir := initRepo(t)
	old, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(old) })
	os.Chdir(dir)

	if !IsGitRepo() {
		t.Fatal("IsGitRepo() returned false inside a git repo")
	}
}

func TestIsGitRepo_outside(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	dir := t.TempDir()
	old, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(old) })
	os.Chdir(dir)

	if IsGitRepo() {
		t.Fatal("IsGitRepo() returned true outside a git repo")
	}
}

func TestCurrentBranch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	dir := initRepo(t)
	seedCommit(t, dir, "a.txt", "initial")
	old, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(old) })
	os.Chdir(dir)

	branch, err := CurrentBranch()
	if err != nil {
		t.Fatalf("CurrentBranch() error: %v", err)
	}
	if branch == "" {
		t.Fatal("CurrentBranch() returned empty string")
	}
}

func TestInspect_cleanRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	dir := initRepo(t)
	seedCommit(t, dir, "a.txt", "initial")
	old, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(old) })
	os.Chdir(dir)

	s, err := Inspect()
	if err != nil {
		t.Fatalf("Inspect() error: %v", err)
	}
	if s.IsDirty {
		t.Error("expected IsDirty=false on clean repo")
	}
	if s.IsDetached {
		t.Error("expected IsDetached=false on normal branch")
	}
	if s.Branch == "" {
		t.Error("expected non-empty Branch")
	}
}

func TestInspect_dirtyRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	dir := initRepo(t)
	seedCommit(t, dir, "a.txt", "initial")
	// add untracked file to make repo dirty
	if err := os.WriteFile(filepath.Join(dir, "dirty.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	old, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(old) })
	os.Chdir(dir)

	s, err := Inspect()
	if err != nil {
		t.Fatalf("Inspect() error: %v", err)
	}
	if !s.IsDirty {
		t.Error("expected IsDirty=true for repo with untracked file")
	}
}

func TestInspect_noUpstream(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	dir := initRepo(t)
	seedCommit(t, dir, "a.txt", "initial")
	old, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(old) })
	os.Chdir(dir)

	s, err := Inspect()
	if err != nil {
		t.Fatalf("Inspect() error: %v", err)
	}
	if s.HasUpstream {
		t.Error("expected HasUpstream=false for repo with no remote")
	}
	if s.AheadCount != 0 || s.BehindCount != 0 {
		t.Errorf("expected AheadCount=0 BehindCount=0, got ahead=%d behind=%d", s.AheadCount, s.BehindCount)
	}
}
