package gitexec

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Run executes a git command and returns an error if the exit code is non-zero.
// Stderr is captured and included in the error message.
func Run(args ...string) error {
	cmd := exec.Command("git", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gf gitexec: git %s: %s", strings.Join(args, " "), strings.TrimSpace(stderr.String()))
	}
	return nil
}

// Output executes a git command and returns trimmed stdout.
// Returns an error if the exit code is non-zero.
func Output(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("gf gitexec: git %s: %s", strings.Join(args, " "), strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(stdout.String()), nil
}

// Lines executes a git command and returns stdout split by newlines with empty lines removed.
func Lines(args ...string) ([]string, error) {
	out, err := Output(args...)
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}
	var lines []string
	for _, l := range strings.Split(out, "\n") {
		if l != "" {
			lines = append(lines, l)
		}
	}
	return lines, nil
}

// CombinedOutput executes a git command and returns trimmed stdout+stderr combined.
// Useful for commands like git push that write remote messages (including MR URLs) to stderr.
func CombinedOutput(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	out, err := cmd.CombinedOutput()
	trimmed := strings.TrimSpace(string(out))
	if err != nil {
		return "", fmt.Errorf("gf gitexec: git %s: %s", strings.Join(args, " "), trimmed)
	}
	return trimmed, nil
}

// RunInDir executes a git command in the specified directory.
func RunInDir(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gf gitexec: git %s: %s", strings.Join(args, " "), strings.TrimSpace(stderr.String()))
	}
	return nil
}
