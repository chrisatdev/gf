package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

// UpdateURL is the GitHub repo for updates
const UpdateURL = "https://github.com/chrisatdev/gf"

// CLI represents the main application structure with git-style flags
var CLI struct {
	// Version flag
	Version bool `short:"v" help:"Show version information"`
	Update  bool `long:"update" help:"Update gf to latest version"`

	// Main flags
	Init      bool   `short:"i" help:"Initialize new Git repository"`
	Configure bool   `short:"c" help:"Configure existing repository (save main branch)"`
	Status    bool   `short:"s" help:"Show git status"`
	Feature   string `short:"f" help:"Create feature branch (use with -s: gf -s -f name)"`
	Hotfix    string `short:"h" help:"Create hotfix branch (use with -s: gf -s -h name)"`
	Bugfix    string `short:"b" help:"Create bugfix branch (use with -s: gf -s -b name)"`
	Release   string `short:"r" help:"Create release branch (use with -s: gf -s -r name)"`
	Add       bool   `short:"a" help:"Stage changes (-a or -a file1 file2)"`
	Commit    bool   `short:"p" help:"Commit and push (-p or -p \"message\")"`
	Merge     bool   `short:"m" help:"Merge main into current branch"`
	Finish    bool   `name:"finish" short:"F" help:"Finish and delete current branch (-f shortcut)"`
	MR        bool   `name:"mr" short:"M" help:"Create MR"`
	Tag       bool   `short:"t" help:"Create a tag (-t <version> [-m <message>] [-p])"`
	Switch    string `short:"w" help:"Switch to a branch (-w name)"`
	Help      bool   `short:"?" help:"Show this help"`
	PushOnly  bool   `short:"u" name:"push-only" help:"Push without creating MR/PR (use with -p)"`
	Pull      bool   `short:"P" help:"Pull current branch from origin"`

	// Positional args (for add and mr)
	Args []string `arg optional name:"args" help:"Files to stage or MR source/target"`
}

// TagCmd for tag operations
type TagCmd struct {
	Version string `arg optional name:"version" help:"Tag version (e.g., v1.0.0)"`
	Message string `short:"m" help:"Tag message"`
	Push    bool   `short:"p" help:"Push tag to remote after creation"`
}

// gfDir is the hidden directory for gf config
const gfDir = ".git/gf"

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

	// Update flag
	if CLI.Update {
		return runUpdate()
	}

	// Help flag - Kong handles this automatically, so just return to let Kong show help
	if CLI.Help || len(os.Args) == 1 {
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
	if CLI.Commit || CLI.PushOnly || flagProvided("-p") {
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

	// Tag
	if CLI.Tag {
		return runTag()
	}

	// Switch to branch
	if CLI.Switch != "" {
		return runSwitch(runner, CLI.Switch)
	}

	// Pull current branch
	if CLI.Pull {
		return runPull(runner)
	}

	// No flags provided, let Kong show help
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

	fmt.Printf("%s🚀 Git Flow Enhanced (gf) v%s%s\n", green, Version, nc)
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
	fmt.Printf("  %sgf -u%s                        %s⏭️ %s Push only (no MR creation)\n", cyan, nc, green, nc)
	fmt.Printf("  %sgf -m%s                        %s🔀%s Merge main into current branch (handle conflicts)\n", cyan, nc, green, nc)
	fmt.Printf("  %sgf -P%s                        %s📥%s Pull current branch from origin\n", cyan, nc, green, nc)
	fmt.Printf("  %sgf -F%s                        %s🗑️%s Finish and delete current branch (local & remote)\n", cyan, nc, red, nc)
	fmt.Printf("  %sgf -M [source] [target]%s      %s🔄%s Create MR from source to target branch (GitLab)\n", cyan, nc, purple, nc)
	fmt.Printf("  %sgf -t <version>%s             %s🏷️%s Create tag (e.g., gf -t v1.0.0)\n", cyan, nc, green, nc)
	fmt.Printf("  %sgf --update%s                  %s⬆️%s Update to latest version\n", cyan, nc, green, nc)
	fmt.Printf("  %sgf -h%s                        %sℹ️%s Show this help\n", cyan, nc, blue, nc)

	fmt.Printf("\n%s📚 Examples:%s\n", purple, nc)
	fmt.Printf("  %sgf -i%s\n", cyan, nc)
	fmt.Printf("  %sgf -c%s\n", cyan, nc)
	fmt.Printf("  %sgf -s -f ticket-1000%s\n", cyan, nc)
	fmt.Printf("  %sgf -a%s\n", cyan, nc)
	fmt.Printf("  %sgf -a file.txt%s\n", cyan, nc)
	fmt.Printf("  %sgf -p \"feat: add new API endpoint\"%s\n", cyan, nc)
	fmt.Printf("  %sgf -u%s\n", cyan, nc)
	fmt.Printf("  %sgf -P%s\n", cyan, nc)
	fmt.Printf("  %sgf -m%s\n", cyan, nc)
	fmt.Printf("  %sgf -F%s\n", cyan, nc)
	fmt.Printf("  %sgf -M main dev%s\n", cyan, nc)
	fmt.Printf("  %sgf -t v1.0.0 -p%s\n", cyan, nc)
	fmt.Printf("  %sgf --update%s\n", cyan, nc)
}

// runTag creates a git tag
func runTag() error {
	green := "\033[32m"
	red := "\033[31m"
	cyan := "\033[36m"
	yellow := "\033[33m"
	nc := "\033[0m"

	runner := git.NewRunner()

	// Parse tag arguments manually
	args := os.Args
	var tagVersion, tagMessage string
	var pushTag bool

	for i, arg := range args {
		if arg == "-t" || arg == "--tag" {
			if i+1 < len(args) {
				tagVersion = args[i+1]
				if strings.HasPrefix(tagVersion, "-") {
					tagVersion = ""
				}
			}
		}
		if arg == "-m" || arg == "--message" {
			if i+1 < len(args) {
				tagMessage = args[i+1]
			}
		}
		if arg == "-p" || arg == "--push" {
			pushTag = true
		}
	}

	if tagVersion == "" {
		return fmt.Errorf("%s❌ Tag version required. Usage: gf -t v1.0.0 [-m \"message\"] [-p]%s\n", red, nc)
	}

	// Ensure version starts with 'v'
	if !strings.HasPrefix(tagVersion, "v") {
		tagVersion = "v" + tagVersion
	}

	// Get current branch
	currentBranch, _ := runner.CurrentBranch()
	mainBranch := GetMainBranch()

	// Only allow tags from main branch
	if currentBranch != mainBranch {
		fmt.Printf("%s⚠️  You are on branch '%s'. Tags should be created from '%s'.%s\n", yellow, currentBranch, mainBranch, nc)
		fmt.Printf("%s💡 Switch to %s and pull latest first.%s\n", cyan, mainBranch, nc)
	}

	// Get tag message
	if tagMessage == "" {
		tagMessage = fmt.Sprintf("Release %s", tagVersion)
	}

	// Create tag
	fmt.Printf("%s🏷️ Creating tag: %s%s\n", cyan, tagVersion, nc)

	// Create annotated tag
	cmd := exec.Command("git", "tag", "-a", tagVersion, "-m", tagMessage)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s❌ Failed to create tag: %v%s\n", red, err, nc)
	}

	fmt.Printf("%s✅ Tag %s created%s\n", green, tagVersion, nc)

	// Push tag if requested
	if pushTag {
		fmt.Printf("%s📤 Pushing tag to origin...%s\n", cyan, nc)
		cmd = exec.Command("git", "push", "origin", tagVersion)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%s❌ Failed to push tag: %v%s\n", red, err, nc)
		}
		fmt.Printf("%s✅ Tag %s pushed to origin%s\n", green, tagVersion, nc)
		fmt.Printf("%s🔗 %s/releases/tag/%s%s\n", cyan, UpdateURL, tagVersion, nc)
	}

	return nil
}

// runUpdate updates gf to the latest version
func runUpdate() error {
	green := "\033[32m"
	yellow := "\033[33m"
	red := "\033[31m"
	cyan := "\033[36m"
	nc := "\033[0m"

	fmt.Printf("%s⬆️ Checking for updates...%s\n", cyan, nc)

	// Get current binary path
	currentPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("%s❌ Could not find current executable: %v%s\n", red, err, nc)
	}

	// Get OS and architecture
	osName := getOS()
	arch := getArch()
	extension := getExtension(osName)

	// Get latest release tag
	latestTag, err := getLatestTag()
	if err != nil {
		return fmt.Errorf("%s❌ Could not check for updates: %v%s\n", red, err, nc)
	}

	// Compare versions
	if latestTag == Version {
		fmt.Printf("%s✅ You're running the latest version: %s%s\n", green, Version, nc)
		return nil
	}

	fmt.Printf("%s🔔 New version available: %s%s\n", yellow, latestTag, nc)
	fmt.Printf("%s📦 Current version: %s%s\n", cyan, Version, nc)

	// Download URL
	downloadName := fmt.Sprintf("gf_%s_%s_%s%s", latestTag, osName, arch, extension)
	downloadURL := fmt.Sprintf("%s/releases/download/%s/%s", UpdateURL, latestTag, downloadName)

	fmt.Printf("%s⬇️ Downloading from: %s%s\n", cyan, downloadURL, nc)

	// Create temp file
	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, "gf_new"+extension)

	// Download
	if err := downloadFile(downloadURL, tempFile); err != nil {
		return fmt.Errorf("%s❌ Download failed: %v%s\n", red, err, nc)
	}

	// Make executable
	if err := os.Chmod(tempFile, 0755); err != nil {
		return fmt.Errorf("%s❌ Could not make executable: %v%s\n", red, err, nc)
	}

	// Try to replace binary
	if err := os.Rename(tempFile, currentPath); err != nil {
		// Binary is in use or no permissions, provide instructions
		fmt.Printf("%s⚠️  Could not replace running binary (in use or no permissions)%s\n", yellow, nc)
		fmt.Printf("%s📦 New version downloaded to: %s%s\n", cyan, tempFile, nc)
		fmt.Println()
		fmt.Printf("%s💡 To complete the update, run:%s\n", green, nc)
		fmt.Printf("   mv %s %s%s\n", tempFile, currentPath, nc)
		fmt.Println()
		fmt.Printf("%s   Or (if above fails):%s\n", cyan, nc)
		fmt.Printf("   sudo mv %s %s%s\n", tempFile, currentPath, nc)
		return nil
	}

	fmt.Printf("%s✅ Successfully updated to version %s%s\n", green, latestTag, nc)
	fmt.Printf("%s💡 Restart gf to use the new version%s\n", cyan, nc)

	return nil
}

func getOS() string {
	switch runtime.GOOS {
	case "darwin":
		return "darwin"
	case "windows":
		return "windows"
	default:
		return "linux"
	}
}

func getArch() string {
	switch runtime.GOARCH {
	case "arm64":
		return "arm64"
	case "arm":
		return "arm"
	default:
		return "amd64"
	}
}

func getExtension(osName string) string {
	if osName == "windows" {
		return ".exe"
	}
	return ""
}

func getLatestTag() (string, error) {
	// Try GitHub API first
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", UpdateURL)
	cmd := exec.Command("curl", "-s", "-m", "10", url)
	output, err := cmd.Output()
	if err == nil && len(output) > 10 {
		var result struct {
			TagName string `json:"tag_name"`
		}
		if err := json.Unmarshal(output, &result); err == nil && result.TagName != "" {
			return strings.TrimPrefix(result.TagName, "v"), nil
		}
	}

	// Fallback: get from git ls-remote
	cmd = exec.Command("git", "ls-remote", "--tags", "--refs", UpdateURL)
	output, err = cmd.Output()
	if err != nil {
		return Version, fmt.Errorf("could not connect to GitHub")
	}

	lines := strings.Split(string(output), "\n")
	var latestTag string
	for _, line := range lines {
		if strings.Contains(line, "refs/tags/") {
			parts := strings.Split(line, "refs/tags/")
			if len(parts) > 1 {
				tag := strings.Split(parts[1], "^")[0]
				if !strings.Contains(tag, "^{}") && tag != "" && strings.HasPrefix(tag, "v") {
					latestTag = tag
				}
			}
		}
	}
	if latestTag == "" {
		return Version, fmt.Errorf("no version tags found")
	}
	return strings.TrimPrefix(latestTag, "v"), nil
}

func downloadFile(url, dest string) error {
	cmd := exec.Command("curl", "-sL", "-o", dest, url)
	return cmd.Run()
}

func copyFile(src, dst string) error {
	from, err := os.Open(src)
	if err != nil {
		return err
	}
	defer from.Close()

	to, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer to.Close()

	_, err = from.Stat()
	if err != nil {
		return err
	}

	_, err = to.Stat()
	if err != nil {
		return err
	}

	return nil
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

	// Initialize changelog directory (but not the file itself)
	changelog.Init()

	// Create initial commit
	commitMsg := "🔧 chore: initial commit"
	if err := runner.Commit(commitMsg); err != nil {
		fmt.Printf("%s⚠️  Warning: Could not create initial commit: %v%s\n", yellow, err, nc)
	}

	// Also run configure to save main branch
	fmt.Println()
	if err := runConfigure(runner); err != nil {
		fmt.Printf("%s⚠️  Warning: Could not save main branch config: %v%s\n", yellow, err, nc)
	}

	fmt.Printf("%s✅ Repository initialized with empty commit%s\n", green, nc)
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

	// Pull current branch if we're on a feature branch
	currentBranch, _ := runner.CurrentBranch()
	if currentBranch != mainBranch && currentBranch != "" {
		fmt.Printf("%s📥 Pulling current branch %s...%s\n", green, currentBranch, nc)
		runner.Fetch("origin")
		if _, err := runner.Command("pull", "origin", currentBranch); err != nil {
			fmt.Printf("%s⚠️  Couldn't pull from origin/%s%s\n", yellow, currentBranch, nc)
		}
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
	yellow := "\033[33m"
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

	// Automatically update CHANGELOG
	changelog.EnsureExists()

	// Get staged files info
	newFiles, modFiles, delFiles, _, _ := runner.StagedFilesByStatus()

	// Generate commit message for changelog
	var allFiles []string
	allFiles = append(allFiles, newFiles...)
	allFiles = append(allFiles, modFiles...)
	allFiles = append(allFiles, delFiles...)

	if len(allFiles) > 0 {
		gen := commit.NewGenerator()
		info := gen.Generate(allFiles, "")
		shortMsg := fmt.Sprintf("%s %s: %s", info.Emoji, info.Type, info.Description)

		commitType := info.Type
		if err := changelog.AddEntry(commitType, shortMsg); err != nil {
			fmt.Printf("%s⚠️  Warning: Could not update CHANGELOG: %v%s\n", yellow, err, nc)
		} else {
			fmt.Printf("%s✅ CHANGELOG.md updated automatically%s\n", green, nc)

			// Stage the changelog
			runner.Add("CHANGELOG.md")
		}
	}

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

	// Open MR for non-main branches (unless PushOnly is set)
	if !isMain && !CLI.PushOnly {
		url, _ := mr.GetMRURL(runner, currentBranch)
		if url != "" {
			fmt.Printf("%s🔗 Opening Merge Request...%s\n", cyan, nc)
			openBrowser(url)
		}
	} else if CLI.PushOnly {
		fmt.Printf("%s⏭️  Push only mode - skipping MR creation%s\n", yellow, nc)
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

// runSwitch switches to a different branch
func runSwitch(runner *git.Runner, branch string) error {
	red := "\033[31m"
	cyan := "\033[36m"
	green := "\033[32m"
	yellow := "\033[33m"
	nc := "\033[0m"

	if runner.HasChanges() {
		fmt.Printf("%s⚠️  You have uncommitted changes. Stashing before switching...%s\n", yellow, nc)
		runner.Command("stash", "push", "-u", "-m", "Auto-stash by gf before branch switch")
	}

	localBranches, _ := runner.LocalBranches()
	branchExists := false
	for _, b := range localBranches {
		if b == branch {
			branchExists = true
			break
		}
	}

	if !branchExists {
		fmt.Printf("%s📥 Branch '%s' not found locally. Fetching from origin...%s\n", cyan, branch, nc)
		cmd := exec.Command("git", "fetch", "origin", branch)
		if err := cmd.Run(); err != nil {
			if runner.HasChanges() {
				fmt.Printf("%s💾 Restoring stashed changes...%s\n", cyan, nc)
				runner.Command("stash", "pop")
			}
			return fmt.Errorf("%s❌ Branch '%s' not found%s\n", red, branch, nc)
		}
		if err := runner.Checkout(branch, false); err != nil {
			if runner.HasChanges() {
				runner.Command("stash", "pop")
			}
			return fmt.Errorf("%s❌ Failed to checkout branch '%s'%s\n", red, branch, nc)
		}
		fmt.Printf("%s✅ Switched to branch '%s'%s\n", green, branch, nc)
	} else {
		if err := runner.Checkout(branch, false); err != nil {
			if runner.HasChanges() {
				runner.Command("stash", "pop")
			}
			return fmt.Errorf("%s❌ Failed to switch to branch '%s'%s\n", red, branch, nc)
		}
		fmt.Printf("%s✅ Switched to branch '%s'%s\n", green, branch, nc)
	}

	behind, _ := runner.IsBehindRemote(branch)
	if behind {
		fmt.Printf("%s⚠️  Branch '%s' is behind origin. Run 'git pull'%s\n", yellow, branch, nc)
	}

	return nil
}

// runPull pulls the current branch from origin
func runPull(runner *git.Runner) error {
	green := "\033[32m"
	red := "\033[31m"
	cyan := "\033[36m"
	nc := "\033[0m"

	currentBranch, err := runner.CurrentBranch()
	if err != nil {
		return fmt.Errorf("%s❌ Could not determine current branch: %v%s\n", red, err, nc)
	}

	fmt.Printf("%s📥 Pulling latest changes from origin/%s...%s\n", green, currentBranch, nc)

	_, err = runner.Pull("origin", currentBranch)
	if err != nil {
		return fmt.Errorf("%s❌ Pull failed: %v%s\n", red, err, nc)
	}

	fmt.Printf("%s✅ Successfully pulled latest changes%s\n", green, nc)
	fmt.Printf("%s💡 Current branch '%s' is now up to date%s\n", cyan, currentBranch, nc)

	return nil
}
