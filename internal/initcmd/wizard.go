package initcmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/chrisatdev/gf/internal/config"
	"github.com/chrisatdev/gf/internal/gitrepo"
	"github.com/chrisatdev/gf/internal/platformdetect"
)

// ErrNotGitRepo is returned when the wizard is invoked outside a git repository.
var ErrNotGitRepo = errors.New("not a git repository, run 'git init' first")

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
			fmt.Print("Platform (github/gitlab): ")
			if !scanner.Scan() {
				return fmt.Errorf("gf init: unexpected end of input")
			}
			platform = strings.ToLower(strings.TrimSpace(scanner.Text()))
			if platform == "github" || platform == "gitlab" {
				break
			}
			fmt.Printf("Invalid platform %q, must be github or gitlab\n", platform)
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

	fmt.Print("MFA active? [y/N]: ")
	if !scanner.Scan() {
		return fmt.Errorf("gf init: unexpected end of input")
	}
	mfaInput := strings.TrimSpace(scanner.Text())
	mfaActive := mfaInput == "y" || mfaInput == "Y"

	if err := config.Write(&config.Config{
		Repo: config.RepoConfig{
			Platform:    platform,
			MainBranch:  mainBranch,
			ProjectPath: projectPath,
		},
		Flow: config.FlowConfig{
			MFAActive: mfaActive,
		},
	}); err != nil {
		return fmt.Errorf("gf init: %w", err)
	}

	fmt.Println("Initialized gf config at .git/gf.toml")
	return nil
}
