package pull

import (
	"fmt"

	"github.com/chrisatdev/gf/internal/gitexec"
)

// Execute runs git pull on the current branch.
func Execute() error {
	if err := gitexec.Run("pull"); err != nil {
		return fmt.Errorf("gf pull: %w", err)
	}
	return nil
}
