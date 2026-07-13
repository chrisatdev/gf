package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
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
		return runSwitch()
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
	fmt.Println("[init]: not implemented yet")
	return nil
}

func runStart(args []string) error {
	if len(args) == 0 {
		fmt.Println("Usage: gf -s <type> <name>")
		fmt.Println()
		fmt.Println("Branch types:")
		fmt.Println("  feat   new feature")
		fmt.Println("  bug    bug fix")
		fmt.Println("  fix    hotfix")
		fmt.Println("  iss    issue")
		fmt.Println("  task   chore/task")
		return nil
	}
	fmt.Println("[start]: not implemented yet")
	return nil
}

func runPush() error {
	fmt.Println("[push]: not implemented yet")
	return nil
}

func runAdd(_ []string) error {
	fmt.Println("[add]: not implemented yet")
	return nil
}

func runFinish() error {
	fmt.Println("[finish]: not implemented yet")
	return nil
}

func runMerge() error {
	fmt.Println("[merge]: not implemented yet")
	return nil
}

func runStatus() error {
	fmt.Println("[status]: not implemented yet")
	return nil
}

func runSwitch() error {
	fmt.Println("[switch]: not implemented yet")
	return nil
}

func runUpdate() error {
	fmt.Println("[update]: not implemented yet")
	return nil
}

func runConfig() error {
	fmt.Println("[config]: not implemented yet")
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
		fmt.Println("[sync]: not implemented yet")
		return nil
	},
}

var resolveCmd = &cobra.Command{
	Use:   "resolve",
	Short: "interactively resolve merge conflicts",
	RunE: func(_ *cobra.Command, _ []string) error {
		fmt.Println("[resolve]: not implemented yet")
		return nil
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
