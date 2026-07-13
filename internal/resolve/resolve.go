package resolve

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/chrisatdev/gf/internal/changelog"
	"github.com/chrisatdev/gf/internal/gitexec"
)

// injectable dependencies — override in tests.
var (
	linesFn    = gitexec.Lines
	runFn      = gitexec.Run
	mergeCLFn  = changelog.MergeConflict
	editorRunFn = openEditor
)

// Execute detects conflicted files and interactively resolves them.
//
// Scenarios:
//
//	no conflicts        → "No conflicts detected." + return nil
//	CHANGELOG.md        → auto-resolve via changelog.MergeConflict + git add
//	other files         → prompt [o]urs / [t]heirs / [e]dit / [s]kip per file
//	all resolved        → "All conflicts resolved. Run 'git commit' to finish."
//	any unresolved      → "Some conflicts remain unresolved."
func Execute(r io.Reader) error {
	conflicts, err := linesFn("diff", "--name-only", "--diff-filter=U")
	if err != nil {
		return fmt.Errorf("gf resolve: %w", err)
	}

	if len(conflicts) == 0 {
		fmt.Println("No conflicts detected.")
		return nil
	}

	scanner := bufio.NewScanner(r)
	var unresolved []string

	for _, file := range conflicts {
		if file == "CHANGELOG.md" {
			if err := mergeCLFn("CHANGELOG.md"); err != nil {
				return fmt.Errorf("gf resolve: %w", err)
			}
			if err := runFn("add", "CHANGELOG.md"); err != nil {
				return fmt.Errorf("gf resolve: %w", err)
			}
			fmt.Println("CHANGELOG.md resolved automatically.")
			continue
		}

		fmt.Printf("%s: [o]urs / [t]heirs / [e]dit / [s]kip: ", file)
		if !scanner.Scan() {
			unresolved = append(unresolved, file)
			continue
		}
		choice := strings.TrimSpace(scanner.Text())

		switch choice {
		case "o":
			if err := runFn("checkout", "--ours", file); err != nil {
				return fmt.Errorf("gf resolve: %w", err)
			}
			if err := runFn("add", file); err != nil {
				return fmt.Errorf("gf resolve: %w", err)
			}
		case "t":
			if err := runFn("checkout", "--theirs", file); err != nil {
				return fmt.Errorf("gf resolve: %w", err)
			}
			if err := runFn("add", file); err != nil {
				return fmt.Errorf("gf resolve: %w", err)
			}
		case "e":
			if err := editorRunFn(file); err != nil {
				return fmt.Errorf("gf resolve: %w", err)
			}
			if err := runFn("add", file); err != nil {
				return fmt.Errorf("gf resolve: %w", err)
			}
		case "s":
			unresolved = append(unresolved, file)
		default:
			// Unrecognised input — treat as skip.
			unresolved = append(unresolved, file)
		}
	}

	if len(unresolved) > 0 {
		fmt.Println("Some conflicts remain unresolved.")
	} else {
		fmt.Println("All conflicts resolved. Run 'git commit' to finish.")
	}

	return nil
}

// openEditor launches the user's preferred editor for the given file.
// Falls back to "vi" when $EDITOR is unset.
func openEditor(file string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	cmd := exec.Command(editor, file)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
