package sync

import (
	"errors"
	"strings"
	"testing"

	"github.com/chrisatdev/gf/internal/config"
	"github.com/chrisatdev/gf/internal/gitrepo"
)

// --- mock helpers ---

func overrideInspect(t *testing.T, s *gitrepo.State) {
	t.Helper()
	orig := inspectFn
	inspectFn = func() (*gitrepo.State, error) { return s, nil }
	t.Cleanup(func() { inspectFn = orig })
}

// overrideRun replaces runFn with a no-op that records all calls.
func overrideRun(t *testing.T) *[][]string {
	t.Helper()
	var calls [][]string
	orig := runFn
	runFn = func(args ...string) error {
		calls = append(calls, append([]string{}, args...))
		return nil
	}
	t.Cleanup(func() { runFn = orig })
	return &calls
}

// overrideRunFailMerge replaces runFn so that a "merge" call returns an error.
// All other calls succeed. Records every call.
func overrideRunFailMerge(t *testing.T) *[][]string {
	t.Helper()
	var calls [][]string
	orig := runFn
	runFn = func(args ...string) error {
		calls = append(calls, append([]string{}, args...))
		if len(args) > 0 && args[0] == "merge" {
			return errors.New("merge: conflict")
		}
		return nil
	}
	t.Cleanup(func() { runFn = orig })
	return &calls
}

// overrideLines replaces linesFn to return a fixed file list.
func overrideLines(t *testing.T, files []string) {
	t.Helper()
	orig := linesFn
	linesFn = func(_ ...string) ([]string, error) { return files, nil }
	t.Cleanup(func() { linesFn = orig })
}

// overrideOutput replaces outputFn to return a fixed string (used for rev-list).
// Format: "<behind> <ahead>" e.g. "3 1"
func overrideOutput(t *testing.T, out string) {
	t.Helper()
	orig := outputFn
	outputFn = func(_ ...string) (string, error) { return out, nil }
	t.Cleanup(func() { outputFn = orig })
}

func syncCfg() *config.Config {
	return &config.Config{Repo: config.RepoConfig{MainBranch: "main"}}
}

// --- tests ---

func TestExecute_NilConfig(t *testing.T) {
	err := Execute(nil)
	if err == nil {
		t.Fatal("expected error for nil config")
	}
	if !strings.Contains(err.Error(), "gf -i") {
		t.Errorf("error should mention 'gf -i', got: %v", err)
	}
}

func TestExecute_BehindZero_AlreadyUpToDate(t *testing.T) {
	overrideInspect(t, &gitrepo.State{
		Branch:      "feat/x",
		AheadCount:  1,
		BehindCount: 0,
	})
	overrideOutput(t, "0 1") // behind=0, ahead=1
	gitCalls := overrideRun(t)

	if err := Execute(syncCfg()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// fetch must be called; merge must NOT be called.
	var sawFetch, sawMerge bool
	for _, c := range *gitCalls {
		if len(c) > 0 && c[0] == "fetch" {
			sawFetch = true
		}
		if len(c) > 0 && c[0] == "merge" {
			sawMerge = true
		}
	}
	if !sawFetch {
		t.Error("expected 'git fetch origin' to be called")
	}
	if sawMerge {
		t.Error("merge should not be called when already up to date")
	}
}

func TestExecute_BehindPositive_CleanMerge(t *testing.T) {
	overrideInspect(t, &gitrepo.State{
		Branch:      "feat/x",
		AheadCount:  1,
		BehindCount: 3,
	})
	overrideOutput(t, "3 1") // behind=3, ahead=1
	gitCalls := overrideRun(t)

	if err := Execute(syncCfg()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the commit was called with the expected message.
	var commitArgs []string
	for _, c := range *gitCalls {
		if len(c) > 0 && c[0] == "commit" {
			commitArgs = c
			break
		}
	}
	if commitArgs == nil {
		t.Fatal("expected 'git commit' to be called")
	}
	msg := strings.Join(commitArgs[2:], " ") // skip "commit" "-m"
	if !strings.Contains(msg, "sync main into feat/x") {
		t.Errorf("commit message %q should contain 'sync main into feat/x'", msg)
	}
}

func TestExecute_BehindPositive_WithConflicts(t *testing.T) {
	overrideInspect(t, &gitrepo.State{
		Branch:      "feat/x",
		AheadCount:  1,
		BehindCount: 2,
	})
	overrideOutput(t, "2 1") // behind=2, ahead=1
	overrideRunFailMerge(t)
	overrideLines(t, []string{"internal/foo/foo.go", "README.md"})

	err := Execute(syncCfg())
	if err == nil {
		t.Fatal("expected error for merge conflicts")
	}
	if !strings.Contains(err.Error(), "conflict") {
		t.Errorf("error should mention conflict, got: %v", err)
	}
}
