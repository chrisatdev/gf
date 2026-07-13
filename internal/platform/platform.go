package platform

import (
	"strings"

	"github.com/chrisatdev/gf/internal/config"
	"github.com/chrisatdev/gf/internal/gitexec"
)

// gitRun is the function used to execute git commands.
// Shared by all platform implementations; override in tests.
var gitRun = gitexec.Run

// Platform abstracts push-related interactions with a remote hosting service.
type Platform interface {
	// Push pushes branch to the remote. title is used as the MR/PR title when
	// onlyPush is false. mainBranch is the merge target.
	Push(branch, title, mainBranch string, onlyPush bool) error

	// Name returns the platform identifier ("github", "gitlab", "noop").
	Name() string
}

// New returns the Platform implementation for the repository described in cfg.
//
// Design note: the constructor accepts *config.Config (not a plain name string)
// so that GitHubPlatform can receive project_path for compare-URL generation
// without coupling to a global config read. Unknown or empty platform values
// fall back to NoopPlatform.
func New(cfg *config.Config) Platform {
	if cfg == nil {
		return &NoopPlatform{}
	}
	switch strings.ToLower(cfg.Repo.Platform) {
	case "github":
		return &GitHubPlatform{projectPath: cfg.Repo.ProjectPath}
	case "gitlab":
		return &GitLabPlatform{}
	default:
		return &NoopPlatform{}
	}
}
