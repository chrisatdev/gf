package start

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/chrisatdev/gf/internal/config"
	"github.com/chrisatdev/gf/internal/gitexec"
)

// BranchTypes maps type name to prefix.
var BranchTypes = map[string]string{
	"feat": "feat",
	"bug":  "bug",
	"fix":  "fix",
	"iss":  "iss",
	"task": "task",
}

// PrintHelp prints the branch type table to stdout.
func PrintHelp() {
	fmt.Println("Branch types:")
	fmt.Println("  feat   feat/<name>   New feature")
	fmt.Println("  bug    bug/<name>    Bug fix")
	fmt.Println("  fix    fix/<name>    Hotfix")
	fmt.Println("  iss    iss/<name>    Issue")
	fmt.Println("  task   task/<name>   Task")
}

// Execute creates a new branch of the given type with the given name.
// Sequence: fetch origin → checkout main → pull → git checkout -b <type>/<name>
func Execute(cfg *config.Config, branchType, name string) error {
	prefix, ok := BranchTypes[branchType]
	if !ok {
		return fmt.Errorf("unknown branch type %q. Valid types: feat, bug, fix, iss, task", branchType)
	}

	if name == "" {
		fmt.Print("Branch name: ")
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			name = strings.TrimSpace(scanner.Text())
		}
	}

	if err := validateName(name); err != nil {
		return err
	}

	fullBranch := prefix + "/" + name

	out, _ := gitexec.Output("branch", "--list", fullBranch)
	if out != "" {
		return fmt.Errorf("branch %q already exists", fullBranch)
	}

	if err := gitexec.Run("fetch", "origin"); err != nil {
		return err
	}
	if err := gitexec.Run("checkout", cfg.Repo.MainBranch); err != nil {
		return err
	}
	if err := gitexec.Run("pull"); err != nil {
		return err
	}
	return gitexec.Run("checkout", "-b", fullBranch)
}

func validateName(name string) error {
	if name == "" {
		return fmt.Errorf("branch name cannot be empty")
	}
	for _, c := range []string{" ", "~", "^", ":", "?", "*", "[", "\\"} {
		if strings.Contains(name, c) {
			return fmt.Errorf("branch name %q contains invalid character %q", name, c)
		}
	}
	return nil
}
