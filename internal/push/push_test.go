package push

import (
	"strings"
	"testing"

	"github.com/chrisatdev/gf/internal/config"
	"github.com/chrisatdev/gf/internal/gitrepo"
	"github.com/chrisatdev/gf/internal/platform"
)

// --- mock helpers ---

type mockPlatform struct {
	name    string
	calls   []platformCall
	pushErr error
}

type platformCall struct {
	branch, title, mainBranch string
	onlyPush                  bool
}

func (m *mockPlatform) Name() string { return m.name }
func (m *mockPlatform) Push(branch, title, mainBranch string, onlyPush bool) error {
	m.calls = append(m.calls, platformCall{branch, title, mainBranch, onlyPush})
	return m.pushErr
}

func overrideInspect(t *testing.T, s *gitrepo.State) {
	t.Helper()
	orig := inspectRepo
	inspectRepo = func() (*gitrepo.State, error) { return s, nil }
	t.Cleanup(func() { inspectRepo = orig })
}

func overrideGitRun(t *testing.T) *[][]string {
	t.Helper()
	var calls [][]string
	orig := gitRun
	gitRun = func(args ...string) error {
		calls = append(calls, append([]string{}, args...))
		return nil
	}
	t.Cleanup(func() { gitRun = orig })
	return &calls
}

func overrideGitOutput(t *testing.T, fn func(...string) (string, error)) {
	t.Helper()
	orig := gitOutput
	gitOutput = fn
	t.Cleanup(func() { gitOutput = orig })
}

func overridePlatform(t *testing.T, plat platform.Platform) {
	t.Helper()
	orig := newPlatform
	newPlatform = func(_ *config.Config) platform.Platform { return plat }
	t.Cleanup(func() { newPlatform = orig })
}

func overrideChangelog(t *testing.T) *[][]string {
	t.Helper()
	var calls [][]string
	orig := prependChangelog
	prependChangelog = func(path, entry string) error {
		calls = append(calls, []string{path, entry})
		return nil
	}
	t.Cleanup(func() { prependChangelog = orig })
	return &calls
}

func gitLabCfg() *config.Config {
	return &config.Config{Repo: config.RepoConfig{Platform: "gitlab", MainBranch: "main"}}
}

func gitHubCfg() *config.Config {
	return &config.Config{
		Repo: config.RepoConfig{Platform: "github", MainBranch: "main", ProjectPath: "owner/repo"},
	}
}

// --- tests ---

func TestExecute_MissingConfig(t *testing.T) {
	err := Execute(nil, Options{})
	if err == nil {
		t.Fatal("expected error for nil config")
	}
	if !strings.Contains(err.Error(), "gf -i") {
		t.Errorf("error should mention 'gf -i', got: %v", err)
	}
}

func TestExecute_NothingToPush(t *testing.T) {
	overrideInspect(t, &gitrepo.State{
		Branch:      "feat/x",
		IsDirty:     false,
		AheadCount:  0,
		HasUpstream: true,
	})
	err := Execute(gitLabCfg(), Options{})
	if err == nil || err.Error() != "Nothing to push." {
		t.Errorf("expected 'Nothing to push.', got: %v", err)
	}
}

func TestExecute_CleanAhead_PushOnly(t *testing.T) {
	overrideInspect(t, &gitrepo.State{
		Branch:      "feat/x",
		IsDirty:     false,
		AheadCount:  2,
		HasUpstream: true,
	})
	overrideGitOutput(t, func(args ...string) (string, error) {
		if len(args) >= 1 && args[0] == "log" {
			return "feat(x): add feature x", nil
		}
		return "", nil
	})
	mock := &mockPlatform{name: "gitlab"}
	overridePlatform(t, mock)

	if err := Execute(gitLabCfg(), Options{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mock.calls) != 1 {
		t.Fatalf("expected 1 platform push, got %d", len(mock.calls))
	}
	c := mock.calls[0]
	if c.branch != "feat/x" {
		t.Errorf("branch = %q, want feat/x", c.branch)
	}
	if c.title != "feat(x): add feature x" {
		t.Errorf("title = %q, want 'feat(x): add feature x'", c.title)
	}
}

func TestExecute_DirtyGeneratedMessage_GitLab(t *testing.T) {
	overrideInspect(t, &gitrepo.State{
		Branch:  "feat/myfeature",
		IsDirty: true,
	})
	overrideGitOutput(t, func(args ...string) (string, error) {
		if len(args) >= 2 && args[0] == "status" && args[1] == "--porcelain" {
			return "M  internal/foo/foo.go\n", nil
		}
		return "", nil
	})
	gitCalls := overrideGitRun(t)
	clCalls := overrideChangelog(t)
	mock := &mockPlatform{name: "gitlab"}
	overridePlatform(t, mock)

	if err := Execute(gitLabCfg(), Options{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Changelog must be prepended.
	if len(*clCalls) == 0 {
		t.Fatal("expected changelog.Prepend to be called")
	}
	if (*clCalls)[0][0] != "CHANGELOG.md" {
		t.Errorf("changelog path = %q, want CHANGELOG.md", (*clCalls)[0][0])
	}

	// Must see git add then git commit.
	if len(*gitCalls) < 2 {
		t.Fatalf("expected at least 2 git calls, got %d: %v", len(*gitCalls), *gitCalls)
	}
	if (*gitCalls)[0][0] != "add" {
		t.Errorf("first git call should be 'add', got %q", (*gitCalls)[0][0])
	}
	if (*gitCalls)[1][0] != "commit" {
		t.Errorf("second git call should be 'commit', got %q", (*gitCalls)[1][0])
	}

	// Platform push called.
	if len(mock.calls) != 1 {
		t.Fatalf("expected 1 platform push call, got %d", len(mock.calls))
	}
	if mock.calls[0].branch != "feat/myfeature" {
		t.Errorf("push branch = %q, want feat/myfeature", mock.calls[0].branch)
	}
}

func TestExecute_DirtyExplicitMessage_GitHub(t *testing.T) {
	overrideInspect(t, &gitrepo.State{
		Branch:  "feat/myfeature",
		IsDirty: true,
	})
	overrideGitOutput(t, func(args ...string) (string, error) {
		return "", nil
	})
	gitCalls := overrideGitRun(t)
	clCalls := overrideChangelog(t)
	mock := &mockPlatform{name: "github"}
	overridePlatform(t, mock)

	if err := Execute(gitHubCfg(), Options{Message: "add login page"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Changelog entry should use the explicit message.
	if len(*clCalls) == 0 {
		t.Fatal("expected changelog.Prepend to be called")
	}
	if (*clCalls)[0][1] != "add login page" {
		t.Errorf("changelog entry = %q, want 'add login page'", (*clCalls)[0][1])
	}

	// Commit -m should carry the explicit message.
	if len(*gitCalls) < 2 {
		t.Fatalf("expected at least 2 git calls, got %d", len(*gitCalls))
	}
	commitCall := (*gitCalls)[1]
	if commitCall[0] != "commit" || len(commitCall) < 3 || commitCall[2] != "add login page" {
		t.Errorf("commit args = %v, expected ['commit', '-m', 'add login page']", commitCall)
	}
	// Must NOT have a second -m (no body for explicit message).
	if len(commitCall) > 3 {
		t.Errorf("explicit message commit should have no body paragraph, got args: %v", commitCall)
	}

	// Platform push with the explicit title.
	if len(mock.calls) != 1 {
		t.Fatalf("expected 1 platform push, got %d", len(mock.calls))
	}
	if mock.calls[0].title != "add login page" {
		t.Errorf("platform title = %q, want 'add login page'", mock.calls[0].title)
	}
}

func TestExecute_OnlyPush(t *testing.T) {
	overrideInspect(t, &gitrepo.State{
		Branch:      "feat/x",
		IsDirty:     false,
		AheadCount:  1,
		HasUpstream: true,
	})
	overrideGitOutput(t, func(args ...string) (string, error) {
		return "feat(x): some commit", nil
	})
	mock := &mockPlatform{name: "gitlab"}
	overridePlatform(t, mock)

	if err := Execute(gitLabCfg(), Options{OnlyPush: true}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mock.calls) != 1 {
		t.Fatalf("expected 1 platform push, got %d", len(mock.calls))
	}
	if !mock.calls[0].onlyPush {
		t.Error("onlyPush flag should be propagated to platform")
	}
}

func TestExecute_DirtyNoUpstream(t *testing.T) {
	overrideInspect(t, &gitrepo.State{
		Branch:      "feat/new-feature",
		IsDirty:     true,
		HasUpstream: false, // no remote tracking branch yet
	})
	overrideGitOutput(t, func(args ...string) (string, error) {
		if len(args) >= 2 && args[0] == "status" {
			return "A  new_file.go\n", nil
		}
		return "", nil
	})
	overrideGitRun(t)
	overrideChangelog(t)
	mock := &mockPlatform{name: "noop"}
	overridePlatform(t, mock)

	if err := Execute(gitLabCfg(), Options{}); err != nil {
		t.Fatalf("dirty+no-upstream should still succeed: %v", err)
	}
	if len(mock.calls) != 1 {
		t.Fatalf("expected 1 platform push, got %d", len(mock.calls))
	}
	if mock.calls[0].branch != "feat/new-feature" {
		t.Errorf("push branch = %q, want feat/new-feature", mock.calls[0].branch)
	}
}
