package add

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func setupRepo(t *testing.T) string {
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

func TestExecute_all(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	dir := setupRepo(t)
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	old, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(old) })
	os.Chdir(dir)

	if err := Execute(nil); err != nil {
		t.Fatalf("Execute(nil) error: %v", err)
	}
	out, _ := exec.Command("git", "-C", dir, "diff", "--cached", "--name-only").Output()
	if !strings.Contains(string(out), "a.txt") {
		t.Errorf("expected a.txt staged, got: %s", out)
	}
}

func TestExecute_specificPath(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	dir := setupRepo(t)
	for _, name := range []string{"a.txt", "b.txt"} {
		os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644)
	}
	old, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(old) })
	os.Chdir(dir)

	if err := Execute([]string{"a.txt"}); err != nil {
		t.Fatalf("Execute(paths) error: %v", err)
	}
	out, _ := exec.Command("git", "-C", dir, "diff", "--cached", "--name-only").Output()
	staged := string(out)
	if !strings.Contains(staged, "a.txt") {
		t.Errorf("expected a.txt staged, got: %s", staged)
	}
	if strings.Contains(staged, "b.txt") {
		t.Errorf("expected b.txt NOT staged, got: %s", staged)
	}
}
