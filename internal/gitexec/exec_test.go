package gitexec

import (
	"strings"
	"testing"
)

func TestRun_success(t *testing.T) {
	if err := Run("--version"); err != nil {
		t.Fatalf("Run(--version) unexpected error: %v", err)
	}
}

func TestRun_invalidCommand(t *testing.T) {
	err := Run("invalid-command-that-does-not-exist")
	if err == nil {
		t.Fatal("expected error for invalid git command, got nil")
	}
}

func TestOutput_success(t *testing.T) {
	out, err := Output("--version")
	if err != nil {
		t.Fatalf("Output(--version) unexpected error: %v", err)
	}
	if out == "" {
		t.Fatal("Output(--version) returned empty string")
	}
	if !strings.HasPrefix(out, "git") {
		t.Errorf("Output(--version) = %q, want prefix %q", out, "git")
	}
}

func TestOutput_invalidCommand(t *testing.T) {
	_, err := Output("invalid-command-that-does-not-exist")
	if err == nil {
		t.Fatal("expected error for invalid git command, got nil")
	}
}

func TestLines_success(t *testing.T) {
	lines, err := Lines("--version")
	if err != nil {
		t.Fatalf("Lines(--version) unexpected error: %v", err)
	}
	if len(lines) == 0 {
		t.Fatal("Lines(--version) returned empty slice")
	}
}

func TestRunInDir_success(t *testing.T) {
	dir := t.TempDir()
	if err := RunInDir(dir, "--version"); err != nil {
		t.Fatalf("RunInDir unexpected error: %v", err)
	}
}
