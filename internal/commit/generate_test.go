package commit_test

import (
	"strings"
	"testing"

	"github.com/chrisatdev/gf/internal/commit"
)

// --- InferType ---

func TestInferType(t *testing.T) {
	mixed := "M  internal/auth/auth.go\nA  internal/user/user.go"
	docOnly := "M  README.md\nA  docs/guide.md"
	testOnly := "A  internal/auth/auth_test.go\nM  internal/user/user_test.go"

	cases := []struct {
		name      string
		branch    string
		porcelain string
		want      string
	}{
		{"feat branch mixed", "feat/user-auth", mixed, "feat"},
		{"bug branch mixed", "bug/login-crash", mixed, "fix"},
		{"fix branch mixed", "fix/null-ptr", mixed, "fix"},
		{"iss branch mixed", "iss/42-timeout", mixed, "fix"},
		{"task branch mixed", "task/cleanup", mixed, "chore"},
		{"no prefix fallback", "hotfix-branch", mixed, "chore"},
		{"docs-only overrides feat", "feat/update-docs", docOnly, "docs"},
		{"docs-only overrides task", "task/readme", docOnly, "docs"},
		{"test-only overrides chore", "task/add-tests", testOnly, "test"},
		{"test-only overrides feat", "feat/auth", testOnly, "test"},
		{"empty porcelain uses branch", "feat/auth", "", "feat"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := commit.InferType(tc.branch, tc.porcelain)
			if got != tc.want {
				t.Errorf("InferType(%q, ...) = %q, want %q", tc.branch, got, tc.want)
			}
		})
	}
}

// --- InferScope ---

func TestInferScope(t *testing.T) {
	cases := []struct {
		name   string
		branch string
		want   string
	}{
		{"feat with scope", "feat/user-auth", "user-auth"},
		{"bug with scope", "bug/login-crash", "login-crash"},
		{"no slash", "main", ""},
		{"empty branch", "", ""},
		{"scope exactly 30 chars", "feat/" + strings.Repeat("a", 30), strings.Repeat("a", 30)},
		{"scope truncated to 30", "feat/" + strings.Repeat("b", 40), strings.Repeat("b", 30)},
		{"hyphens preserved", "feat/user-profile-v2", "user-profile-v2"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := commit.InferScope(tc.branch)
			if got != tc.want {
				t.Errorf("InferScope(%q) = %q, want %q", tc.branch, got, tc.want)
			}
		})
	}
}

// --- BuildBody ---

func TestBuildBody(t *testing.T) {
	cases := []struct {
		name      string
		porcelain string
		want      []string
	}{
		{
			name:      "single add",
			porcelain: "A  internal/auth/auth.go",
			want:      []string{"add: internal/auth/auth.go"},
		},
		{
			name:      "single modify",
			porcelain: "M  internal/config/config.go",
			want:      []string{"modify: internal/config/config.go"},
		},
		{
			name:      "single delete",
			porcelain: "D  old/file.go",
			want:      []string{"remove: old/file.go"},
		},
		{
			name:      "rename preserves arrow notation",
			porcelain: "R  old_name.go -> new_name.go",
			want:      []string{"rename: old_name.go -> new_name.go"},
		},
		{
			name:      "mixed changes",
			porcelain: "M  internal/auth/auth.go\nA  internal/auth/auth_test.go\nD  internal/old/old.go",
			want: []string{
				"modify: internal/auth/auth.go",
				"add: internal/auth/auth_test.go",
				"remove: internal/old/old.go",
			},
		},
		{
			name:      "untracked ignored",
			porcelain: "?? scratch.go",
			want:      nil,
		},
		{
			name:      "empty porcelain",
			porcelain: "",
			want:      nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := commit.BuildBody(tc.porcelain)
			if len(got) != len(tc.want) {
				t.Fatalf("BuildBody() len=%d, want %d\ngot:  %v\nwant: %v", len(got), len(tc.want), got, tc.want)
			}
			for i, line := range got {
				if line != tc.want[i] {
					t.Errorf("BuildBody()[%d] = %q, want %q", i, line, tc.want[i])
				}
			}
		})
	}
}

// --- inferDescription (via InferFromBranch) ---

func TestInferFromBranch_description(t *testing.T) {
	cases := []struct {
		name      string
		porcelain string
		wantDesc  string
	}{
		{"single add", "A  internal/auth/auth.go", "add auth.go"},
		{"single modify", "M  internal/config/config.go", "update config.go"},
		{"single delete", "D  internal/old/old.go", "remove old.go"},
		{"two files", "M  a.go\nA  b.go", "update 2 files"},
		{"three files", "M  a.go\nA  b.go\nD  c.go", "update 3 files"},
		{"empty porcelain", "", "update files"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			msg := commit.InferFromBranch("feat/x", tc.porcelain)
			if msg.Description != tc.wantDesc {
				t.Errorf("Description = %q, want %q", msg.Description, tc.wantDesc)
			}
		})
	}
}

// --- CommitMessage.String ---

func TestCommitMessage_String(t *testing.T) {
	cases := []struct {
		name string
		msg  commit.CommitMessage
		want string
	}{
		{
			name: "with scope",
			msg: commit.CommitMessage{
				Type: "feat", Scope: "user-auth", Description: "update 3 files",
				Body: []string{"add: internal/auth/auth.go", "modify: internal/user/user.go"},
			},
			want: "feat(user-auth): update 3 files\n\nadd: internal/auth/auth.go\nmodify: internal/user/user.go",
		},
		{
			name: "without scope",
			msg: commit.CommitMessage{
				Type: "chore", Scope: "", Description: "update 2 files",
				Body: []string{"add: foo.go", "modify: bar.go"},
			},
			want: "chore: update 2 files\n\nadd: foo.go\nmodify: bar.go",
		},
		{
			name: "no body",
			msg:  commit.CommitMessage{Type: "docs", Scope: "", Description: "update README.md"},
			want: "docs: update README.md",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.msg.String()
			if got != tc.want {
				t.Errorf("String() =\n%q\nwant:\n%q", got, tc.want)
			}
		})
	}
}

// --- docs-only classification edge cases ---

func TestInferType_docsEdgeCases(t *testing.T) {
	cases := []struct {
		name      string
		branch    string
		porcelain string
		want      string
	}{
		{
			name:      "README.sh is not docs",
			branch:    "feat/docs",
			porcelain: "M  README.sh",
			want:      "feat",
		},
		{
			name:      "CHANGELOG.md is docs",
			branch:    "feat/release",
			porcelain: "M  CHANGELOG.md",
			want:      "docs",
		},
		{
			name:      "mixed doc and code files is not docs-only",
			branch:    "feat/auth",
			porcelain: "M  README.md\nA  internal/auth/auth.go",
			want:      "feat",
		},
		{
			name:      "docs directory counts as doc",
			branch:    "task/docs",
			porcelain: "A  docs/guide.md\nM  docs/reference.md",
			want:      "docs",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := commit.InferType(tc.branch, tc.porcelain)
			if got != tc.want {
				t.Errorf("InferType(%q, %q) = %q, want %q", tc.branch, tc.porcelain, got, tc.want)
			}
		})
	}
}
