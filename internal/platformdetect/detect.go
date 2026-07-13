package platformdetect

import (
	"fmt"
	"net/url"
	"os/exec"
	"strings"
)

// RemoteInfo holds normalized information extracted from a git remote URL.
type RemoteInfo struct {
	Platform    string // "github" | "gitlab" | "unknown"
	Host        string
	ProjectPath string // "owner/repo"
	HTTPSURL    string // credential-free https URL
}

// ParseRemoteURL normalizes raw (SSH or HTTPS) into a RemoteInfo.
// Credentials are always stripped from HTTPSURL.
func ParseRemoteURL(raw string) RemoteInfo {
	raw = strings.TrimSpace(raw)

	var host, path string

	if strings.HasPrefix(raw, "git@") {
		// git@github.com:owner/repo.git
		trimmed := strings.TrimPrefix(raw, "git@")
		parts := strings.SplitN(trimmed, ":", 2)
		if len(parts) == 2 {
			host = parts[0]
			path = strings.TrimSuffix(parts[1], ".git")
		}
	} else {
		u, err := url.Parse(raw)
		if err == nil {
			host = u.Hostname()
			path = strings.TrimPrefix(u.Path, "/")
			path = strings.TrimSuffix(path, ".git")
		}
	}

	return RemoteInfo{
		Platform:    ClassifyHost(host),
		Host:        host,
		ProjectPath: path,
		HTTPSURL:    fmt.Sprintf("https://%s/%s", host, path),
	}
}

// NormalizeURL returns a credential-free HTTPS URL for any git remote URL format.
func NormalizeURL(raw string) string {
	return ParseRemoteURL(raw).HTTPSURL
}

// ClassifyHost returns "github", "gitlab", or "unknown" for the given host string.
func ClassifyHost(host string) string {
	lower := strings.ToLower(host)
	if lower == "github.com" || strings.HasSuffix(lower, ".github.com") {
		return "github"
	}
	if lower == "gitlab.com" || strings.Contains(lower, "gitlab") {
		return "gitlab"
	}
	return "unknown"
}

// DetectRemote shells out to git to read the origin remote URL and parses it.
func DetectRemote() (RemoteInfo, error) {
	out, err := exec.Command("git", "remote", "get-url", "origin").Output()
	if err != nil {
		return RemoteInfo{}, fmt.Errorf("gf platformdetect: %w", err)
	}
	return ParseRemoteURL(string(out)), nil
}
