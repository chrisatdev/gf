package release

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/chrisatdev/gf/internal/gitexec"
)

// Options configures the release command.
type Options struct {
	Bump string // "major", "minor", or "patch"
}

// injectable dependencies — override in tests.
var (
	detectFn   = Detect
	gitRunFn   = gitexec.Run
	lookPathFn = exec.LookPath
)

// Execute runs the full release flow:
// detect version file → read current version → bump → write file → commit → tag → push → goreleaser.
func Execute(opts Options) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("release: %w", err)
	}

	vf, err := detectFn(cwd)
	if err != nil {
		return err
	}

	current, err := ReadVersion(vf)
	if err != nil {
		return err
	}

	next, err := BumpVersion(current, opts.Bump)
	if err != nil {
		return err
	}

	fmt.Printf("%s → %s\n", current, next)

	if vf.FileType != FileTypeGoTag {
		if err := WriteVersion(vf, next); err != nil {
			return err
		}
		if err := gitRunFn("add", vf.Path); err != nil {
			return fmt.Errorf("release: %w", err)
		}
		if err := gitRunFn("commit", "-m", "chore: bump version to "+next); err != nil {
			return fmt.Errorf("release: %w", err)
		}
		if err := gitRunFn("push", "origin", "HEAD"); err != nil {
			return fmt.Errorf("release: %w", err)
		}
	}

	if err := gitRunFn("tag", "-a", next, "-m", "Release "+next); err != nil {
		return fmt.Errorf("release: %w", err)
	}
	if err := gitRunFn("push", "origin", next); err != nil {
		return fmt.Errorf("release: %w", err)
	}

	fmt.Printf("Tag %s pushed.\n", next)

	if bin, err := lookPathFn("goreleaser"); err == nil {
		fmt.Println("Running goreleaser...")
		cmd := exec.Command(bin, "release", "--clean")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if runErr := cmd.Run(); runErr != nil {
			fmt.Fprintf(os.Stderr, "goreleaser: %v\n", runErr)
		}
	}

	return nil
}
