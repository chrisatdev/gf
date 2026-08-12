package release

import (
	"os"
	"path/filepath"
	"testing"
)

func captureGitRun(t *testing.T) *[][]string {
	t.Helper()
	var calls [][]string
	orig := gitRunFn
	gitRunFn = func(args ...string) error {
		calls = append(calls, append([]string{}, args...))
		return nil
	}
	t.Cleanup(func() { gitRunFn = orig })
	return &calls
}

func TestExecute_plainVersionFile(t *testing.T) {
	dir := t.TempDir()
	versionFile := filepath.Join(dir, "VERSION")
	write(t, versionFile, "v1.0.0\n")

	calls := captureGitRun(t)

	origDetect := detectFn
	detectFn = func(_ string) (*VersionFile, error) {
		return &VersionFile{Path: versionFile, FileType: FileTypePlain}, nil
	}
	t.Cleanup(func() { detectFn = origDetect })

	origLook := lookPathFn
	lookPathFn = func(_ string) (string, error) { return "", os.ErrNotExist }
	t.Cleanup(func() { lookPathFn = origLook })

	if err := Execute(Options{Bump: "patch"}); err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	data, _ := os.ReadFile(versionFile)
	if string(data) != "v1.0.1\n" {
		t.Errorf("VERSION = %q, want %q", string(data), "v1.0.1\n")
	}

	// Expect: git add, git commit, git push HEAD, git tag -a, git push tag
	if len(*calls) != 5 {
		t.Fatalf("expected 5 git calls, got %d: %v", len(*calls), *calls)
	}
	assertArgs(t, (*calls)[0], "add", versionFile)
	assertArgs(t, (*calls)[1], "commit", "-m", "chore: bump version to v1.0.1")
	assertArgs(t, (*calls)[2], "push", "origin", "HEAD")
	assertArgs(t, (*calls)[3], "tag", "-a", "v1.0.1", "-m", "Release v1.0.1")
	assertArgs(t, (*calls)[4], "push", "origin", "v1.0.1")
}

func TestExecute_goTagNoFileUpdate(t *testing.T) {
	calls := captureGitRun(t)

	origDetect := detectFn
	detectFn = func(_ string) (*VersionFile, error) {
		return &VersionFile{FileType: FileTypeGoTag}, nil
	}
	t.Cleanup(func() { detectFn = origDetect })

	origReadGitTag := readGitTag
	readGitTag = func() (string, error) { return "v2.1.0", nil }
	t.Cleanup(func() { readGitTag = origReadGitTag })

	origLook := lookPathFn
	lookPathFn = func(_ string) (string, error) { return "", os.ErrNotExist }
	t.Cleanup(func() { lookPathFn = origLook })

	if err := Execute(Options{Bump: "minor"}); err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	// No file write → no add/commit/push-branch; only tag -a and push tag
	if len(*calls) != 2 {
		t.Fatalf("expected 2 git calls, got %d: %v", len(*calls), *calls)
	}
	assertArgs(t, (*calls)[0], "tag", "-a", "v2.2.0", "-m", "Release v2.2.0")
	assertArgs(t, (*calls)[1], "push", "origin", "v2.2.0")
}

func TestExecute_invalidBump(t *testing.T) {
	origDetect := detectFn
	detectFn = func(_ string) (*VersionFile, error) {
		return &VersionFile{FileType: FileTypeGoTag}, nil
	}
	t.Cleanup(func() { detectFn = origDetect })

	origReadGitTag := readGitTag
	readGitTag = func() (string, error) { return "v1.0.0", nil }
	t.Cleanup(func() { readGitTag = origReadGitTag })

	if err := Execute(Options{Bump: "oops"}); err == nil {
		t.Fatal("expected error for invalid bump type, got nil")
	}
}

func assertArgs(t *testing.T, got []string, want ...string) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("git call args = %v, want %v", got, want)
		return
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("git call args[%d] = %q, want %q (full: %v)", i, got[i], want[i], got)
		}
	}
}
