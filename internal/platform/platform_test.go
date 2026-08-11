package platform

import (
	"testing"

	"github.com/chrisatdev/gf/internal/config"
)

// captureGitRun replaces gitRun with a recorder and restores the original
// via t.Cleanup. Returns a pointer to the captured call log.
func captureGitRun(t *testing.T) *[][]string {
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

// captureGitCombinedOutput replaces gitCombinedOutput with a recorder that returns output.
func captureGitCombinedOutput(t *testing.T, output string) *[][]string {
	t.Helper()
	var calls [][]string
	orig := gitCombinedOutput
	gitCombinedOutput = func(args ...string) (string, error) {
		calls = append(calls, append([]string{}, args...))
		return output, nil
	}
	t.Cleanup(func() { gitCombinedOutput = orig })
	return &calls
}

// captureOpenURL replaces openURL with a recorder and restores it.
func captureOpenURL(t *testing.T) *string {
	t.Helper()
	var opened string
	orig := openURL
	openURL = func(u string) error { opened = u; return nil }
	t.Cleanup(func() { openURL = orig })
	return &opened
}

func equalArgs(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// --- New() factory ---

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *config.Config
		wantName string
	}{
		{
			"github",
			&config.Config{Repo: config.RepoConfig{Platform: "github", ProjectPath: "owner/repo"}},
			"github",
		},
		{
			"gitlab",
			&config.Config{Repo: config.RepoConfig{Platform: "gitlab"}},
			"gitlab",
		},
		{
			"unknown falls back to noop",
			&config.Config{Repo: config.RepoConfig{Platform: "bitbucket"}},
			"noop",
		},
		{
			"empty platform falls back to noop",
			&config.Config{},
			"noop",
		},
		{
			"nil cfg falls back to noop",
			nil,
			"noop",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(tt.cfg)
			if p.Name() != tt.wantName {
				t.Errorf("Name() = %q, want %q", p.Name(), tt.wantName)
			}
		})
	}
}

// --- GitLabPlatform ---

func TestGitLabPush(t *testing.T) {
	t.Run("only_push performs plain tracked push", func(t *testing.T) {
		calls := captureGitRun(t)
		p := &GitLabPlatform{}
		if err := p.Push("feat/x", "feat(x): add feature", "main", true); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		wantArgs := []string{"push", "-u", "origin", "feat/x"}
		if len(*calls) != 1 || !equalArgs((*calls)[0], wantArgs) {
			t.Errorf("git args = %v, want %v", *calls, [][]string{wantArgs})
		}
	})

	t.Run("mr_push opens MR URL from remote output", func(t *testing.T) {
		const fakeURL = "https://gitlab.com/owner/repo/-/merge_requests/new"
		fakeOutput := "remote: \nremote: To create a merge request for feat/x, visit:\nremote:   " + fakeURL + "\nremote: \n"
		calls := captureGitCombinedOutput(t, fakeOutput)
		openedURL := captureOpenURL(t)

		p := &GitLabPlatform{}
		if err := p.Push("feat/x", "feat(x): add feature", "main", false); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		wantArgs := []string{
			"push", "-u", "origin", "feat/x",
			"-o", "merge_request.create",
			"-o", "merge_request.title=feat(x): add feature",
			"-o", "merge_request.target=main",
		}
		if len(*calls) != 1 || !equalArgs((*calls)[0], wantArgs) {
			t.Errorf("git args = %v, want %v", *calls, [][]string{wantArgs})
		}
		if *openedURL != fakeURL {
			t.Errorf("browser URL = %q, want %q", *openedURL, fakeURL)
		}
	})

	t.Run("mr_push with no URL in output skips browser", func(t *testing.T) {
		captureGitCombinedOutput(t, "remote: \nremote: Everything up-to-date\nremote: \n")
		openedURL := captureOpenURL(t)

		p := &GitLabPlatform{}
		if err := p.Push("feat/x", "feat(x): add feature", "main", false); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if *openedURL != "" {
			t.Errorf("expected no browser open, got URL %q", *openedURL)
		}
	})
}

// --- GitHubPlatform ---

func TestGitHubPush(t *testing.T) {
	t.Run("only_push skips browser open", func(t *testing.T) {
		calls := captureGitRun(t)
		openedURL := captureOpenURL(t)

		p := &GitHubPlatform{projectPath: "owner/repo"}
		if err := p.Push("feat/x", "feat(x): add feature", "main", true); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		wantArgs := []string{"push", "-u", "origin", "feat/x"}
		if len(*calls) != 1 || !equalArgs((*calls)[0], wantArgs) {
			t.Errorf("git args = %v, want %v", *calls, [][]string{wantArgs})
		}
		if *openedURL != "" {
			t.Errorf("expected no browser open, got URL %q", *openedURL)
		}
	})

	t.Run("mr_push opens compare URL in browser", func(t *testing.T) {
		captureGitRun(t)
		openedURL := captureOpenURL(t)

		p := &GitHubPlatform{projectPath: "owner/repo"}
		if err := p.Push("feat/x", "feat(x): add feature", "main", false); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		wantURL := "https://github.com/owner/repo/compare/feat/x?expand=1"
		if *openedURL != wantURL {
			t.Errorf("browser URL = %q, want %q", *openedURL, wantURL)
		}
	})
}

// --- NoopPlatform ---

func TestNoopPush(t *testing.T) {
	tests := []struct {
		name     string
		onlyPush bool
	}{
		{"onlyPush=false still plain push", false},
		{"onlyPush=true still plain push", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := captureGitRun(t)
			p := &NoopPlatform{}
			if err := p.Push("feat/x", "title", "main", tt.onlyPush); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			wantArgs := []string{"push", "-u", "origin", "feat/x"}
			if len(*calls) != 1 || !equalArgs((*calls)[0], wantArgs) {
				t.Errorf("git args = %v, want %v", *calls, [][]string{wantArgs})
			}
		})
	}
}
