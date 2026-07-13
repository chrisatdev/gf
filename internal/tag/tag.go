package tag

import (
	"fmt"

	"github.com/chrisatdev/gf/internal/gitexec"
)

// Create creates a git tag. If message is provided, creates an annotated tag.
// If push is true, pushes the tag to origin.
func Create(name, message string, push bool) error {
	if name == "" {
		return fmt.Errorf("gf tag: tag name is required")
	}

	var err error
	if message != "" {
		err = gitexec.Run("tag", "-a", name, "-m", message)
	} else {
		err = gitexec.Run("tag", name)
	}
	if err != nil {
		return fmt.Errorf("gf tag: %w", err)
	}

	fmt.Printf("Tag %s created.\n", name)

	if push {
		if err := gitexec.Run("push", "origin", name); err != nil {
			return fmt.Errorf("gf tag: push failed: %w", err)
		}
		fmt.Printf("Tag %s pushed to origin.\n", name)
	}
	return nil
}
