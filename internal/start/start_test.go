package start_test

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chrisatdev/gf/internal/config"
	"github.com/chrisatdev/gf/internal/start"
)

func TestPrintHelp(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w

	start.PrintHelp()

	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)

	out := buf.String()
	for _, want := range []string{"feat", "bug", "fix", "iss", "task"} {
		if !strings.Contains(out, want) {
			t.Errorf("PrintHelp output missing %q\ngot: %s", want, out)
		}
	}
}

func TestExecute_invalidType(t *testing.T) {
	cfg := &config.Config{Repo: config.RepoConfig{MainBranch: "main"}}
	err := start.Execute(cfg, "invalid", "my-branch")
	if err == nil {
		t.Fatal("expected error for invalid branch type")
	}
	if !strings.Contains(err.Error(), "unknown branch type") {
		t.Errorf("unexpected error message: %s", err)
	}
}

func TestExecute_invalidName(t *testing.T) {
	cfg := &config.Config{Repo: config.RepoConfig{MainBranch: "main"}}
	cases := []struct {
		desc string
		name string
	}{
		{"space", "branch with space"},
		{"tilde", "branch~1"},
		{"caret", "branch^bad"},
		{"colon", "branch:name"},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			err := start.Execute(cfg, "feat", tc.name)
			if err == nil {
				t.Errorf("expected error for name %q", tc.name)
			}
		})
	}
}

func TestExecute_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	work := setupRepo(t)
	orig, _ := os.Getwd()
	if err := os.Chdir(work); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })

	cfg := &config.Config{Repo: config.RepoConfig{MainBranch: "main"}}
	if err := start.Execute(cfg, "feat", "user-auth"); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	out, err := exec.Command("git", "branch", "--list", "feat/user-auth").Output()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "feat/user-auth") {
		t.Error("branch feat/user-auth not found after Execute")
	}
}

func TestExecute_branchAlreadyExists(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	work := setupRepo(t)
	orig, _ := os.Getwd()
	if err := os.Chdir(work); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })

	must(t, exec.Command("git", "checkout", "-b", "feat/existing").Run())
	must(t, exec.Command("git", "checkout", "main").Run())

	cfg := &config.Config{Repo: config.RepoConfig{MainBranch: "main"}}
	err := start.Execute(cfg, "feat", "existing")
	if err == nil {
		t.Fatal("expected error for existing branch")
	}
	if !strings.Contains(err.Error(), "already exists") {
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
