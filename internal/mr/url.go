package mr

import (
	"fmt"
	"strings"

	"github.com/chrisatdev/gf/internal/git"
)

type Platform int

const (
	PlatformGitLab Platform = iota
	PlatformGitHub
	PlatformUnknown
)

func DetectPlatform(remoteURL string) Platform {
	if strings.Contains(remoteURL, "gitlab") {
		return PlatformGitLab
	}
	if strings.Contains(remoteURL, "github") {
		return PlatformGitHub
	}
	return PlatformUnknown
}

func GetMRURL(runner *git.Runner, branch string) (string, error) {
	remoteURL, err := runner.RemoteURL()
	if err != nil {
		return "", err
	}

	// Clean up URL
	remoteURL = strings.TrimSuffix(remoteURL, ".git")
	remoteURL = strings.ReplaceAll(remoteURL, "git@", "")
	remoteURL = strings.ReplaceAll(remoteURL, "https://", "")
	remoteURL = strings.ReplaceAll(remoteURL, ":", "/")

	platform := DetectPlatform(remoteURL)

	switch platform {
	case PlatformGitLab:
		return fmt.Sprintf("https://%s/-/merge_requests/new?merge_request[source_branch]=%s", remoteURL, branch), nil
	case PlatformGitHub:
		return fmt.Sprintf("https://%s/compare/%s?expand=1", remoteURL, branch), nil
	default:
		return "", fmt.Errorf("unsupported platform")
	}
}
