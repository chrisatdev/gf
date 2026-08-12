package initcmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/chrisatdev/gf/internal/config"
	"github.com/chrisatdev/gf/internal/gitrepo"
	"github.com/chrisatdev/gf/internal/platformdetect"
)

// ErrNotGitRepo is returned when the wizard is invoked outside a git repository.
var ErrNotGitRepo = errors.New("Not a git repository. Run 'git init' first.")

// Run executes the interactive init wizard.
// r is the input reader (os.Stdin in production, strings.NewReader in tests).
func Run(r io.Reader) error {
	if !gitrepo.IsGitRepo() {
		return ErrNotGitRepo
	}

	scanner := bufio.NewScanner(r)

	if config.Exists() {
		fmt.Print("Config already exists. Reinitialize? [y/N]: ")
		if !scanner.Scan() {
			return nil
		}
		answer := strings.TrimSpace(scanner.Text())
		if answer != "y" && answer != "Y" {
			return nil
		}
	}

	var platform, projectPath string

	remote, err := platformdetect.DetectRemote()
	if err != nil || remote.Platform == "unknown" {
		if err == nil {
			fmt.Printf("Detected host: %s (platform unknown)\n", remote.Host)
			projectPath = remote.ProjectPath
		}
		for {
			fmt.Print("  Platform:\n    1) GitHub\n    2) GitLab\n  Choice [1]: ")
			if !scanner.Scan() {
				return fmt.Errorf("gf init: unexpected end of input")
			}
			choice := strings.TrimSpace(scanner.Text())
			switch choice {
			case "1", "":
				platform = "github"
			case "2":
				platform = "gitlab"
			default:
				fmt.Printf("Invalid choice %q, enter 1 or 2\n", choice)
				continue
			}
			break
		}
	} else {
		fmt.Printf("Detected platform: %s (%s)\n", remote.Platform, remote.ProjectPath)
		platform = remote.Platform
		projectPath = remote.ProjectPath
	}

	fmt.Print("Main branch [main]: ")
	if !scanner.Scan() {
		return fmt.Errorf("gf init: unexpected end of input")
	}
	mainBranch := strings.TrimSpace(scanner.Text())
	if mainBranch == "" {
		mainBranch = "main"
	}

	if err := config.Write(&config.Config{
		Repo: config.RepoConfig{
			Platform:    platform,
			MainBranch:  mainBranch,
			ProjectPath: projectPath,
		},
		Flow: config.FlowConfig{
			MFAActive: false,
		},
	}); err != nil {
		return fmt.Errorf("gf init: %w", err)
	}

	fmt.Println("Initialized gf config at .git/gf.toml")
	if err := ensureGitignore(".atl/"); err != nil {
		fmt.Printf("Warning: could not update .gitignore: %v\n", err)
	}
	return nil
}

// ensureGitignore appends entry to .gitignore if not already present.
func ensureGitignore(entry string) error {
	const path = ".gitignore"
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == entry {
			return nil
		}
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	prefix := ""
	if len(data) > 0 && data[len(data)-1] != '\n' {
		prefix = "\n"
	}
	_, err = fmt.Fprintf(f, "%s%s\n", prefix, entry)
	return err
}
