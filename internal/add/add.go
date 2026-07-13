package add

import "github.com/chrisatdev/gf/internal/gitexec"

// Execute stages files for commit. If paths is empty, stages all tracked and untracked files.
func Execute(paths []string) error {
	if len(paths) == 0 {
		return gitexec.Run("add", "--all")
	}
	args := append([]string{"add"}, paths...)
	return gitexec.Run(args...)
}
