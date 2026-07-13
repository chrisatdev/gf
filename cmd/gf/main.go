package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/chrisatdev/gf/internal/add"
	"github.com/chrisatdev/gf/internal/config"
	"github.com/chrisatdev/gf/internal/finish"
	"github.com/chrisatdev/gf/internal/initcmd"
	"github.com/chrisatdev/gf/internal/merge"
	"github.com/chrisatdev/gf/internal/push"
	"github.com/chrisatdev/gf/internal/resolve"
	"github.com/chrisatdev/gf/internal/start"
	"github.com/chrisatdev/gf/internal/switchcmd"
	gfsync "github.com/chrisatdev/gf/internal/sync"
)

var version = "dev"

var (
	flagInit     bool
	flagStart    bool
	flagPush     bool
	flagAdd      bool
	flagFinish   bool
	flagMerge    bool
	flagStatus   bool
	flagSwitch   bool
	flagUpdate   bool
	flagVersion  bool
	flagConfig   bool
	flagMessage  string
	flagOnlyPush bool
)

var rootCmd = &cobra.Command{
	Use:          "gf",
	Short:        "Git Flow CLI — streamlined git workflow tool",
	SilenceUsage: true,
	RunE:         dispatch,
}

func init() {
	rootCmd.Flags().BoolVarP(&flagInit, "init", "i", false, "initialize gf config in current repo")
	rootCmd.Flags().BoolVarP(&flagStart, "start", "s", false, "start a new branch: gf -s <type> <name>")
	rootCmd.Flags().BoolVarP(&flagPush, "push", "p", false, "commit, update changelog, and push")
	rootCmd.Flags().BoolVarP(&flagAdd, "add", "a", false, "stage files (default: all)")
	rootCmd.Flags().BoolVarP(&flagFinish, "finish", "f", false, "delete current branch locally and remotely")
	rootCmd.Flags().BoolVarP(&flagMerge, "merge", "m", false, "merge origin main into current branch")
	rootCmd.Flags().BoolVarP(&flagStatus, "status", "S", false, "show git status")
	rootCmd.Flags().BoolVarP(&flagSwitch, "switch", "w", false, "interactively switch branches")
	rootCmd.Flags().BoolVarP(&flagUpdate, "update", "u", false, "self-update gf binary")
	rootCmd.Flags().BoolVarP(&flagVersion, "version", "v", false, "print gf version")
	rootCmd.Flags().BoolVarP(&flagConfig, "config", "c", false, "print current gf config")
	rootCmd.Flags().StringVarP(&flagMessage, "message", "M", "", "commit message for push")
	rootCmd.Flags().BoolVar(&flagOnlyPush, "only-push", false, "push without creating PR/MR")

	rootCmd.AddCommand(commitCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(resolveCmd)
}

func dispatch(cmd *cobra.Command, args []string) error {
	switch {
	case flagVersion:
		return runVersion()
	case flagInit:
		return runInit()
	case flagStart:
		return runStart(args)
	case flagPush:
		return runPush()
	case flagAdd:
		return runAdd(args)
	case flagFinish:
		return runFinish()
	case flagMerge:
		return runMerge()
	case flagStatus:
		return runStatus()
	case flagSwitch:
		return runSwitch(args)
	case flagUpdate:
		return runUpdate()
	case flagConfig:
		return runConfig()
	default:
		return cmd.Help()
	}
}

func runVersion() error {
	fmt.Printf("gf version %s\n", version)
	return nil
}

func runInit() error {
	if err := initcmd.Run(os.Stdin); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	return nil
}

func runStart(args []string) error {
	if len(args) == 0 {
		start.PrintHelp()
		return nil
	}
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "No config found. Run 'gf -i' to initialize.")
		os.Exit(1)
	}
	branchType := args[0]
	name := ""
	if len(args) > 1 {
		name = args[1]
	}
	if err := start.Execute(cfg, branchType, name); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	return nil
}

func runPush() error {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "No gf config found. Run 'gf -i' to initialize.")
		os.Exit(1)
	}
	opts := push.Options{
		Message:  flagMessage,
		OnlyPush: flagOnlyPush,
	}
	if err := push.Execute(cfg, opts); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	return nil
}

func runAdd(args []string) error {
	if err := add.Execute(args); err != nil {
		return err
	}
	cmd := exec.Command("git", "diff", "--cached", "--stat")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
	return nil
}

func runFinish() error {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "No config found. Run 'gf -i' to initialize.")
		os.Exit(1)
	}
	if err := finish.Execute(cfg, os.Stdin); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	return nil
}

func runMerge() error {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "No config found. Run 'gf -i' to initialize.")
		os.Exit(1)
	}
	if err := merge.Execute(cfg); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	return nil
}

func runStatus() error {
	cmd := exec.Command("git", "status")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runSwitch(args []string) error {
	branch := ""
	if len(args) > 0 {
		branch = args[0]
	}
	return switchcmd.Execute(branch, os.Stdin)
}

func runUpdate() error {
	fmt.Println("[update]: not implemented yet")
	return nil
}

func runConfig() error {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "No config found. Run 'gf -i' to initialize.")
		os.Exit(1)
	}
	fmt.Println("gf configuration:")
	fmt.Printf("  platform:     %s\n", cfg.Repo.Platform)
	fmt.Printf("  main_branch:  %s\n", cfg.Repo.MainBranch)
	fmt.Printf("  project_path: %s\n", cfg.Repo.ProjectPath)
	fmt.Printf("  mfa_active:   %v\n", cfg.Flow.MFAActive)
	return nil
}

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "interactive conventional commit wizard",
	RunE: func(_ *cobra.Command, _ []string) error {
		fmt.Println("[commit]: not implemented yet")
		return nil
	},
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "sync current branch with origin main",
	RunE: func(_ *cobra.Command, _ []string) error {
		cfg, _ := config.Load()
		if err := gfsync.Execute(cfg); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		return nil
	},
}

var resolveCmd = &cobra.Command{
	Use:   "resolve",
	Short: "interactively resolve merge conflicts",
	RunE: func(_ *cobra.Command, _ []string) error {
		if err := resolve.Execute(os.Stdin); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		return nil
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
