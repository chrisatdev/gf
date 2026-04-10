package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/chrisatdev/gf/internal/changelog"
	"github.com/chrisatdev/gf/internal/commit"
	"github.com/chrisatdev/gf/internal/git"
	"github.com/chrisatdev/gf/internal/mr"
	"github.com/chrisatdev/gf/internal/tui"
)

// Version info
const Version = "2.0.0"

// CLI represents the main application structure with git-style flags
var CLI struct {
	// Version flag
	Version bool `short:"v" help:"Show version information"`

	// Main flags
	Init      bool   `short:"i" help:"Initialize new Git repository"`
	Configure bool   `short:"c" help:"Configure existing repository (save main branch)"`
	Status    bool   `short:"s" help:"Show git status"`
	Feature   string `short:"f" help:"Create feature branch (-f name)"`
	Hotfix    string `short:"h" help:"Create hotfix branch (-h name)"`
	Bugfix    string `short:"b" help:"Create bugfix branch (-b name)"`
	Release   string `short:"r" help:"Create release branch (-r name)"`
	Add       bool   `short:"a" help:"Stage changes (-a or -a file1 file2)"`
	Commit    bool   `short:"p" help:"Commit and push (-p or -p \"message\")"`
	Merge     bool   `short:"m" help:"Merge main into current branch"`
	Finish    bool   `name:"finish" short:"F" help:"Finish and delete current branch (-f shortcut)"`
	MR        bool   `name:"mr" short:"M" help:"Create MR"`
	Help      bool   `short:"?" help:"Show this help"`

	// Positional args (for add and mr)
	Args []string `arg optional name:"args" help:"Files to stage or MR source/target"`
}

// gfDir is the hidden directory for gf config
const gfDir = ".gf"

// mainBranchFile stores the main branch name
const mainBranchFile = "main_branch"

func Execute() error {
	_ = kong.Parse(&CLI,
		kong.Name("gf"),
		kong.Description("Git Flow Enhanced - A powerful Git workflow automation tool"),
		kong.HelpOptions{
			Compact: true,
		},
	)

	// Version flag
	if CLI.Version {
		fmt.Printf("gf version %s\n", Version)
		return nil
	}

	// Help flag
	if CLI.Help || len(os.Args) == 1 {
		showHelp()
		return nil
	}

	runner := git.NewRunner()

	// Initialize
	if CLI.Init {
		return runInit(runner)
	}

	// Check if inside git repo
	if !runner.IsRepo() {
		return fmt.Errorf("not a git repository. Run 'gf -i' to initialize")
	}

	// Configure
	if CLI.Configure {
		return runConfigure(runner)
	}

	// Status / Branch creation
	if CLI.Status {
		if CLI.Feature != "" {
			return runBranch(runner, "feature", CLI.Feature)
		}
		if CLI.Hotfix != "" {
			return runBranch(runner, "hotfix", CLI.Hotfix)
		}
		if CLI.Bugfix != "" {
			return runBranch(runner, "bugfix", CLI.Bugfix)
		}
		if CLI.Release != "" {
			return runBranch(runner, "release", CLI.Release)
		}
		// Just status
		return runStatus(runner)
	}

	// Add (when -a flag is used)
	if CLI.Add {
		return runAdd(runner, CLI.Args)
	}

	// Commit & Push
	if CLI.Commit || flagProvided("-p") {
		msg := ""
		if len(CLI.Args) > 0 {
			msg = CLI.Args[0]
		}
		return runCommit(runner, msg)
	}

	// Merge
	if CLI.Merge {
		return runMerge(runner)
	}

	// Finish
	if CLI.Finish {
		return runFinish(runner)
	}

	// MR
	if CLI.MR {
		var source, target string
		if len(CLI.Args) >= 1 {
			source = CLI.Args[0]
		}
		if len(CLI.Args) >= 2 {
			target = CLI.Args[1]
		}
		return runMR(runner, source, target)
	}

	// No flags provided, show help
	showHelp()
	return nil
}

func flagProvided(flag string) bool {
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, flag) && !strings.HasPrefix(arg, "--") {
			return true
		}
	}
	return false
}

func showHelp() {
	cyan := "\033[36m"
	green := "\033[32m"
	red := "\033[31m"
	yellow := "\033[33m"
	blue := "\033[34m"
	purple := "\033[35m"
	nc := "\033[0m"

	fmt.Printf("%s🚀 Git Flow Enhanced (gf)%s\n", green, nc)
	fmt.Printf("%s Version: %s - by Christian Benítez%s\n", green, Version, nc)
	fmt.Printf("%s GitHub: https://github.com/chrisatdev%s\n\n", green, nc)

	fmt.Printf("%s  Usage:%s\n", green, nc)
	fmt.Printf("  %sgf -i%s                        %s🆕%s Initialize new Git repository\n", cyan, nc, green, nc)
	fmt.Printf("  %sgf -c%s                        %s🔧%s Configure repository (save main branch)\n", cyan, nc, green, nc)
	fmt.Printf("  %sgf -s%s                        %s✅%s Alias to git status\n", cyan, nc, green, nc)
	fmt.Printf("  %sgf -s -f [name]%s              %s✨%s Create feature branch (feature/name)\n", cyan, nc, green, nc)
	fmt.Printf("  %sgf -s -h [name]%s              %s🐛%s Create hotfix branch (hotfix/name)\n", cyan, nc, red, nc)
	fmt.Printf("  %sgf -s -b [name]%s              %s🚑%s Create bugfix branch (bugfix/name)\n", cyan, nc, yellow, nc)
	fmt.Printf("  %sgf -s -r [name]%s              %s🚀%s Create release branch (release/name)\n", cyan, nc, blue, nc)
	fmt.Printf("  %sgf -a%s                        %s📦%s Stage changes (stage all if no files specified)\n", cyan, nc, green, nc)
	fmt.Printf("  %sgf -p \"[msg]\"%s                %s💾%s Commit (with message) and push, then open MR/PR\n", cyan, nc, green, nc)
	fmt.Printf("  %sgf -m%s                        %s🔀%s Merge main into current branch (handle conflicts)\n", cyan, nc, green, nc)
	fmt.Printf("  %sgf -F%s                        %s🗑️%s Finish and delete current branch (local & remote)\n", cyan, nc, red, nc)
	fmt.Printf("  %sgf -M [source] [target]%s      %s🔄%s Create MR from source to target branch (GitLab)\n", cyan, nc, purple, nc)
	fmt.Printf("  %sgf -h%s                        %sℹ️%s Show this help\n", cyan, nc, blue, nc)

	fmt.Printf("\n%s📚 Examples:%s\n", purple, nc)
	fmt.Printf("  %sgf -i%s\n", cyan, nc)
	fmt.Printf("  %sgf -c%s\n", cyan, nc)
	fmt.Printf("  %sgf -s -f ticket-1000%s\n", cyan, nc)
	fmt.Printf("  %sgf -a%s\n", cyan, nc)
	fmt.Printf("  %sgf -a file.txt%s\n", cyan, nc)
	fmt.Printf("  %sgf -p \"feat: add new API endpoint\"%s\n", cyan, nc)
	fmt.Printf("  %sgf -m%s\n", cyan, nc)
	fmt.Printf("  %sgf -F%s\n", cyan, nc)
	fmt.Printf("  %sgf -M main dev%s\n", cyan, nc)
}

// runInit initializes a new git repository
func runInit(runner *git.Runner) error {
	green := "\033[32m"
	yellow := "\033[33m"
	red := "\033[31m"
	nc := "\033[0m"

	if runner.IsRepo() {
		fmt.Printf("%s⚠️  Already a git repository%s\n", yellow, nc)
		return nil
	}

	fmt.Printf("%s🆕 Initializing new Git repository...%s\n", green, nc)
	if err := runner.Init(); err != nil {
		return fmt.Errorf("%s❌ Error initializing repository: %v%s\n", red, err, nc)
	}

	// Initialize changelog
	changelog.Init()

	// Stage changelog
	runner.Add("CHANGELOG.md")

	// Create initial commit
	commitMsg := "🔧 chore: initial commit"
	if err := runner.Commit(commitMsg); err != nil {
		fmt.Printf("%s⚠️  Warning: Could not create initial commit: %v%s\n", yellow, err, nc)
	}

	// Save main branch info
	mainBranch := detectMainBranch()
	if err := saveMainBranch(mainBranch); err != nil {
		fmt.Printf("%s⚠️  Warning: Could not save main branch config: %v%s\n", yellow, err, nc)
	}

	fmt.Printf("%s✅ Repository initialized with empty commit%s\n", green, nc)
	fmt.Printf("%s📁 Main branch saved as: %s%s\n", green, mainBranch, nc)
	return nil
}

// runConfigure configures an existing repository
func runConfigure(runner *git.Runner) error {
	green := "\033[32m"
	yellow := "\033[33m"
	red := "\033[31m"
	cyan := "\033[36m"
	nc := "\033[0m"

	// Detect common main branch names
	branches := []string{"main", "master", "develop", "dev"}
	var detected string

	for _, branch := range branches {
		if _, err := runner.Command("rev-parse", "--verify", "origin/"+branch); err == nil {
			detected = branch
			break
		}
	}

	if detected == "" {
		detected = "main"
		fmt.Printf("%s⚠️  Could not detect main branch. Defaulting to: %s%s\n", yellow, detected, nc)
	}

	fmt.Printf("%s🔧 Configuring gf for this repository%s\n\n", cyan, nc)
	fmt.Printf("%sDetected main branch: %s%s\n", green, detected, nc)

	// Save to git config
	mainBranch := detected
	if err := saveMainBranchToGit(mainBranch); err != nil {
		return fmt.Errorf("%s❌ Failed to save configuration: %v%s\n", red, err, nc)
	}

	fmt.Printf("%s✅ Configuration saved successfully%s\n", green, nc)
	fmt.Printf("%s📁 Main branch: %s%s\n", green, mainBranch, nc)
	fmt.Printf("\n%s💡 This configuration will be used for MRs and merge operations%s\n", cyan, nc)

	return nil
}

// saveMainBranchToGit saves the main branch to git config
func saveMainBranchToGit(branch string) error {
	// Save to .gf directory
	if err := os.MkdirAll(gfDir, 0755); err != nil {
		return err
	}
	path := filepath.Join(gfDir, mainBranchFile)
	return os.WriteFile(path, []byte(branch), 0644)
}

// detectMainBranch detects if the default branch is main or master
func detectMainBranch() string {
	runner := git.NewRunner()

	// Check if main exists
	if _, err := runner.Command("rev-parse", "--verify", "origin/main"); err == nil {
		return "main"
	}

	// Check if master exists
	if _, err := runner.Command("rev-parse", "--verify", "origin/master"); err == nil {
		return "master"
	}

	// Default to main
	return "main"
}

// saveMainBranch saves the main branch name to .gf/main_branch
func saveMainBranch(branch string) error {
	// Create .gf directory
	if err := os.MkdirAll(gfDir, 0755); err != nil {
		return err
	}

	// Write main branch file
	path := filepath.Join(gfDir, mainBranchFile)
	return os.WriteFile(path, []byte(branch), 0644)
}

// GetMainBranch reads the saved main branch name
func GetMainBranch() string {
	path := filepath.Join(gfDir, mainBranchFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return "main" // Default
	}
	return strings.TrimSpace(string(data))
}

// runStatus shows git status
func runStatus(runner *git.Runner) error {
	output, err := runner.Status()
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

// runBranch creates a new branch
func runBranch(runner *git.Runner, branchType, name string) error {
	green := "\033[32m"
	yellow := "\033[33m"
	red := "\033[31m"
	cyan := "\033[36m"
	nc := "\033[0m"

	if name == "" {
		return fmt.Errorf("%s❌ Branch name is required%s\n", red, nc)
	}

	// Get main branch
	mainBranch := GetMainBranch()

	// Update main branch first
	fmt.Printf("%s🔄 Updating %s branch...%s\n", green, mainBranch, nc)
	runner.Command("checkout", mainBranch)
	runner.Fetch("origin")

	// Try to pull
	if _, err := runner.Command("pull", "origin", mainBranch); err != nil {
		fmt.Printf("%s⚠️  Couldn't pull from origin/%s. Using local %s branch%s\n", yellow, mainBranch, mainBranch, nc)
	}

	// Create new branch
	fullName := fmt.Sprintf("%s/%s", branchType, name)
	emoji := getBranchEmoji(branchType)

	fmt.Printf("%s🌱 Creating branch: %s %s%s\n", green, cyan, fullName, emoji, nc)

	if err := runner.CreateBranch(fullName); err != nil {
		return fmt.Errorf("%s❌ Error creating branch: %v%s\n", red, err, nc)
	}

	fmt.Printf("%s✅ Branch %s %screated%s\n", green, cyan, fullName, green, nc)
	return nil
}

func getBranchEmoji(branchType string) string {
	switch branchType {
	case "feature":
		return "✨"
	case "hotfix":
		return "🐛"
	case "bugfix":
		return "🚑"
	case "release":
		return "🚀"
	default:
		return ""
	}
}

// runAdd stages changes
func runAdd(runner *git.Runner, files []string) error {
	green := "\033[32m"
	red := "\033[31m"
	nc := "\033[0m"

	if len(files) == 0 {
		fmt.Printf("%s📦 Staging all changes...%s\n", green, nc)
		if err := runner.Add("."); err != nil {
			return fmt.Errorf("%s❌ Error staging changes: %v%s\n", red, err, nc)
		}
	} else {
		fmt.Printf("%s📦 Staging specified files...%s\n", green, nc)
		if err := runner.Add(files...); err != nil {
			return fmt.Errorf("%s❌ Error staging changes: %v%s\n", red, err, nc)
		}
	}

	fmt.Printf("%s✅ Changes staged%s\n", green, nc)
	return nil
}

// runCommit commits and pushes
func runCommit(runner *git.Runner, message string) error {
	green := "\033[32m"
	yellow := "\033[33m"
	red := "\033[31m"
	cyan := "\033[36m"
	nc := "\033[0m"

	currentBranch, _ := runner.CurrentBranch()
	mainBranch := GetMainBranch()
	isMain := currentBranch == mainBranch

	// Check for staged files
	staged, _ := runner.StagedFiles()
	if len(staged) == 0 {
		// Try to auto-stage
		runner.Add(".")
		staged, _ = runner.StagedFiles()
		if len(staged) == 0 {
			return fmt.Errorf("%s❌ No changes to commit%s\n", red, nc)
		}
	}

	// Generate or use provided message
	if message == "" {
		// Auto-generate message
		gen := commit.NewGenerator()

		// Get files by status
		newFiles, modFiles, delFiles, renamedFiles, _ := runner.StagedFilesByStatus()
		diff, _ := runner.GetStagedDiff()

		info := gen.GenerateFromFileStatus(newFiles, modFiles, delFiles, renamedFiles, diff)
		message = gen.FormatMessage(info)

		fmt.Printf("%s📝 Auto-generated commit message:%s\n\n%s\n", yellow, nc, message)
		fmt.Println()
	} else {
		// Add gitmoji if not present
		if !containsGitmoji(message) {
			message = addGitmoji(message)
		}
	}

	// Update changelog
	commitType := extractCommitType(message)
	changelog.AddEntry(commitType, message)

	// Commit
	runner.Command("commit", "-m", message)
	fmt.Printf("%s💾 Commit created%s\n", green, nc)

	// Push
	fmt.Printf("%s📤 Pushing to %s%s%s\n", green, cyan, currentBranch, nc)
	if err := runner.Push(currentBranch, true); err != nil {
		fmt.Printf("%s❌ Push failed: %v%s\n", red, err, nc)
		fmt.Printf("%s⚠️  If conflicts exist, run: %sgf -m%s\n", yellow, cyan, nc)
		return nil
	}
	fmt.Printf("%s✅ Push successful%s\n", green, nc)

	// Open MR for non-main branches
	if !isMain {
		url, _ := mr.GetMRURL(runner, currentBranch)
		if url != "" {
			fmt.Printf("%s🔗 Opening Merge Request...%s\n", cyan, nc)
			openBrowser(url)
		}
	} else {
		fmt.Printf("%s⚠️  No MR will be created for %s branch%s\n", yellow, mainBranch, nc)
	}

	return nil
}

func containsGitmoji(msg string) bool {
	emojis := []string{"✨", "🐛", "📝", "💄", "♻️", "✅", "🔧", "👷", "⚙️", "⚡", "⏪"}
	for _, e := range emojis {
		if strings.Contains(msg, e) {
			return true
		}
	}
	return false
}

func addGitmoji(msg string) string {
	msgLower := strings.ToLower(msg)
	var emoji string
	switch {
	case strings.HasPrefix(msgLower, "feat"):
		emoji = "✨"
	case strings.HasPrefix(msgLower, "fix"):
		emoji = "🐛"
	case strings.HasPrefix(msgLower, "docs"):
		emoji = "📝"
	case strings.HasPrefix(msgLower, "style"):
		emoji = "💄"
	case strings.HasPrefix(msgLower, "refactor"):
		emoji = "♻️"
	case strings.HasPrefix(msgLower, "test"):
		emoji = "✅"
	case strings.HasPrefix(msgLower, "build"):
		emoji = "👷"
	case strings.HasPrefix(msgLower, "ci"):
		emoji = "⚙️"
	case strings.HasPrefix(msgLower, "perf"):
		emoji = "⚡"
	default:
		emoji = "🔧"
	}
	return emoji + " " + msg
}

func openBrowser(url string) {
	cmds := [][]string{
		{"xdg-open", url},
		{"open", url},
		{"cmd", "/c", "start", url},
	}
	for _, cmd := range cmds {
		if exec.Command(cmd[0], cmd[1:]...).Run() == nil {
			return
		}
	}
}

func extractCommitType(message string) string {
	msgLower := strings.ToLower(message)
	if strings.HasPrefix(msgLower, "feat") {
		return "feat"
	}
	if strings.HasPrefix(msgLower, "fix") {
		return "fix"
	}
	if strings.HasPrefix(msgLower, "docs") {
		return "docs"
	}
	if strings.HasPrefix(msgLower, "style") {
		return "style"
	}
	if strings.HasPrefix(msgLower, "refactor") {
		return "refactor"
	}
	if strings.HasPrefix(msgLower, "test") {
		return "test"
	}
	if strings.HasPrefix(msgLower, "build") {
		return "build"
	}
	if strings.HasPrefix(msgLower, "ci") {
		return "ci"
	}
	if strings.HasPrefix(msgLower, "perf") {
		return "perf"
	}
	return "chore"
}

// runMerge merges main into current branch
func runMerge(runner *git.Runner) error {
	green := "\033[32m"
	yellow := "\033[33m"
	red := "\033[31m"
	cyan := "\033[36m"
	nc := "\033[0m"

	currentBranch, _ := runner.CurrentBranch()
	mainBranch := GetMainBranch()

	if currentBranch == mainBranch {
		return fmt.Errorf("%s❌ Cannot merge %s into itself%s\n", red, mainBranch, nc)
	}

	// Check for uncommitted changes
	if runner.HasChanges() {
		fmt.Printf("%s⚠️  Uncommitted changes detected%s\n", yellow, nc)
		fmt.Println("Please commit or stash your changes before merging")
		return fmt.Errorf("uncommitted changes")
	}

	fmt.Printf("%s🔄 Fetching origin/%s...%s\n", green, mainBranch, nc)
	runner.Fetch("origin")

	fmt.Printf("%s🔀 Merging origin/%s into %s%s%s\n", green, mainBranch, cyan, currentBranch, nc)

	if err := runner.Merge("origin/"+mainBranch, false); err != nil {
		// Check for conflicts
		if runner.HasConflicts() {
			fmt.Printf("%s❌ Merge conflicts detected!%s\n", red, nc)
			conflicts, _ := runner.GetConflictingFiles()
			fmt.Printf("%s⚠️  %d conflicting file(s)%s\n", yellow, len(conflicts), nc)
			return tui.ResolveConflicts()
		}
		return fmt.Errorf("%s❌ Merge failed: %v%s\n", red, err, nc)
	}

	fmt.Printf("%s✅ Merge successful%s\n", green, nc)
	return nil
}

// runFinish finishes and cleans up the current branch
func runFinish(runner *git.Runner) error {
	green := "\033[32m"
	red := "\033[31m"
	cyan := "\033[36m"
	nc := "\033[0m"

	currentBranch, _ := runner.CurrentBranch()
	mainBranch := GetMainBranch()

	if currentBranch == mainBranch {
		return fmt.Errorf("%s❌ Cannot delete %s branch%s\n", red, mainBranch, nc)
	}

	fmt.Printf("%s🔄 Switching to %s...%s\n", green, mainBranch, nc)
	runner.Checkout(mainBranch, false)

	fmt.Printf("%s📥 Pulling latest %s...%s\n", green, mainBranch, nc)
	runner.Pull()

	fmt.Printf("%s🗑️  Deleting local branch %s%s%s\n", green, cyan, currentBranch, nc)
	runner.DeleteBranch(currentBranch, true)

	fmt.Printf("%s♻️  Deleting remote branch...%s\n", green, nc)
	runner.DeleteRemoteBranch(currentBranch)

	fmt.Printf("%s✅ Branch %s %scleaned up%s\n", green, cyan, currentBranch, nc)
	return nil
}

// runMR creates a merge request
func runMR(runner *git.Runner, source, target string) error {
	green := "\033[32m"
	red := "\033[31m"
	cyan := "\033[36m"
	nc := "\033[0m"

	if source == "" || target == "" {
		return fmt.Errorf("%s❌ Both source and target branches are required%s\n", red, nc)
	}

	// Verify GitLab
	remoteURL, _ := runner.RemoteURL()
	if !strings.Contains(remoteURL, "gitlab") {
		return fmt.Errorf("%s❌ MR creation is only supported for GitLab repositories%s\n", red, nc)
	}

	// Save current branch
	currentBranch, _ := runner.CurrentBranch()

	fmt.Printf("%s🔄 Creating MR from %s to %s...%s\n", green, cyan, source, cyan, target, nc)

	// Checkout and push source
	runner.Checkout(source, false)
	runner.Push(source, true)

	// Get MR URL
	url, err := mr.GetMRURL(runner, source)
	if err != nil {
		return fmt.Errorf("%s❌ Failed to generate MR URL: %v%s\n", red, err, nc)
	}

	fmt.Printf("%s✅ MR created: %s%s\n", green, cyan, url, nc)
	fmt.Printf("%s🔗 Opening in browser...%s\n", cyan, nc)
	openBrowser(url)

	// Return to original branch
	runner.Checkout(currentBranch, false)

	return nil
}
