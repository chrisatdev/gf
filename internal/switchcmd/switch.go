package switchcmd

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/chrisatdev/gf/internal/gitexec"
)

// Execute switches to the given branch and pulls.
// If branch is empty, lists local branches and prompts for selection.
func Execute(branch string, r io.Reader) error {
	if branch == "" {
		var err error
		branch, err = selectBranch(r)
		if err != nil {
			return err
		}
	}

	if err := gitexec.Run("switch", branch); err != nil {
		return err
	}
	// pull is best-effort; ignore if no upstream is configured
	_ = gitexec.Run("pull")
	return nil
}

func selectBranch(r io.Reader) (string, error) {
	branches, err := gitexec.Lines("branch", "--format=%(refname:short)")
	if err != nil {
		return "", fmt.Errorf("gf switch: %w", err)
	}
	if len(branches) == 0 {
		return "", fmt.Errorf("gf switch: no local branches found")
	}

	for i, b := range branches {
		fmt.Printf("  %d) %s\n", i+1, b)
	}
	fmt.Print("Branch number: ")

	scanner := bufio.NewScanner(r)
	if !scanner.Scan() {
		return "", fmt.Errorf("gf switch: no input")
	}
	n, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
	if err != nil || n < 1 || n > len(branches) {
		return "", fmt.Errorf("gf switch: invalid selection")
	}
	return branches[n-1], nil
}
