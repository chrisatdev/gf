package git

import (
	"fmt"
	"os/exec"
	"strings"
)

type Runner struct {
	dryRun bool
}

func NewRunner() *Runner {
	return &Runner{}
}

func (r *Runner) Command(args ...string) (string, error) {
	if r.dryRun {
		return "", nil
	}

	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %w\n%s", args[0], err, output)
	}
	return strings.TrimSpace(string(output)), nil
}

func (r *Runner) IsRepo() bool {
	_, err := r.Command("rev-parse", "--is-inside-work-tree")
	return err == nil
}

func (r *Runner) CurrentBranch() (string, error) {
	return r.Command("rev-parse", "--abbrev-ref", "HEAD")
}

func (r *Runner) Init() error {
	_, err := r.Command("init")
	return err
}

func (r *Runner) Status() (string, error) {
	return r.Command("status")
}

func (r *Runner) Add(files ...string) error {
	args := append([]string{"add"}, files...)
	_, err := r.Command(args...)
	return err
}

func (r *Runner) Commit(message string) error {
	_, err := r.Command("commit", "-m", message)
	return err
}

func (r *Runner) Push(branch string, setUpstream bool) error {
	args := []string{"push"}
	if setUpstream {
		args = append(args, "-u")
	}
	args = append(args, "origin", branch)
	_, err := r.Command(args...)
	return err
}

func (r *Runner) Pull(args ...string) (string, error) {
	fullArgs := append([]string{"pull"}, args...)
	return r.Command(fullArgs...)
}

func (r *Runner) Fetch(remote string) error {
	_, err := r.Command("fetch", remote)
	return err
}

func (r *Runner) Checkout(branch string, create bool) error {
	args := []string{"checkout"}
	if create {
		args = append(args, "-b")
	}
	args = append(args, branch)
	_, err := r.Command(args...)
	return err
}

func (r *Runner) CreateBranch(name string) error {
	_, err := r.Command("checkout", "-b", name)
	return err
}

func (r *Runner) DeleteBranch(name string, force bool) error {
	args := []string{"branch"}
	if force {
		args = append(args, "-D")
	} else {
		args = append(args, "-d")
	}
	args = append(args, name)
	_, err := r.Command(args...)
	return err
}

func (r *Runner) DeleteRemoteBranch(name string) error {
	_, err := r.Command("push", "origin", "--delete", name)
	return err
}

func (r *Runner) Merge(branch string, noCommit bool) error {
	args := []string{"merge"}
	if noCommit {
		args = append(args, "--no-commit")
	}
	args = append(args, branch)
	_, err := r.Command(args...)
	return err
}

func (r *Runner) HasChanges() bool {
	output, _ := r.Command("status", "--porcelain")
	return len(output) > 0
}

func (r *Runner) StagedFiles() ([]string, error) {
	output, err := r.Command("diff", "--name-only", "--cached")
	if err != nil {
		return nil, err
	}
	if output == "" {
		return []string{}, nil
	}
	return strings.Split(output, "\n"), nil
}

func (r *Runner) RemoteURL() (string, error) {
	return r.Command("remote", "get-url", "origin")
}
