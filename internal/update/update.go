package update

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// Release is the GitHub API releases/latest response shape.
type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

// Asset is a single release artefact listed in a GitHub release.
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// HTTPClient is the HTTP transport interface, injectable for tests.
type HTTPClient interface {
	Get(url string) (*http.Response, error)
}

// injectable dependencies — override in tests.
var (
	githubAPIBase     = "https://api.github.com"
	httpDoer          HTTPClient             = &http.Client{}
	currentGOOS                              = runtime.GOOS
	currentGOARCH                            = runtime.GOARCH
	executableFn      func() (string, error) = os.Executable
	downloadReplaceFn func(url string) error  = downloadAndReplace
)

const defaultRepo = "chrisatdev/gf"

// Execute checks for a newer release and updates the binary when one is found.
//
//   - dev/empty version: prints latest available, returns nil (no update applied)
//   - same version:      prints "gf is up to date (vX.Y.Z)."
//   - newer version:     downloads archive, extracts binary, replaces executable
//   - unsupported OS:    returns clear error
//   - asset not found:   returns clear error naming the expected asset
func Execute(currentVersion string) error {
	release, err := CheckLatest(defaultRepo)
	if err != nil {
		return err
	}

	latestTag := release.TagName

	// dev or empty build: report what is available but do not update.
	if currentVersion == "" || currentVersion == "dev" {
		fmt.Printf("gf latest available: %s (running dev build — no update applied).\n", latestTag)
		return nil
	}

	if !isNewer(currentVersion, latestTag) {
		fmt.Printf("gf is up to date (%s).\n", currentVersion)
		return nil
	}

	assetName := AssetName()
	if assetName == "" {
		return fmt.Errorf("gf update: unsupported platform %s/%s", currentGOOS, currentGOARCH)
	}

	var downloadURL string
	for _, a := range release.Assets {
		if a.Name == assetName {
			downloadURL = a.BrowserDownloadURL
			break
		}
	}
	if downloadURL == "" {
		return fmt.Errorf("gf update: asset %q not found in release %s", assetName, latestTag)
	}

	if err := downloadReplaceFn(downloadURL); err != nil {
		return fmt.Errorf("gf update: %w", err)
	}

	fmt.Printf("Updated gf to %s.\n", latestTag)
	return nil
}

// CheckLatest fetches the latest release metadata from the GitHub Releases API.
func CheckLatest(repo string) (*Release, error) {
	url := fmt.Sprintf("%s/repos/%s/releases/latest", githubAPIBase, repo)
	resp, err := httpDoer.Get(url)
	if err != nil {
		return nil, fmt.Errorf("gf update: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gf update: GitHub API returned status %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("gf update: %w", err)
	}
	return &release, nil
}

// AssetName returns the expected archive file name for the current OS/arch,
// or an empty string when the platform is not supported.
//
// Supported combinations: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64.
func AssetName() string {
	switch currentGOOS {
	case "linux", "darwin":
	default:
		return ""
	}
	switch currentGOARCH {
	case "amd64", "arm64":
	default:
		return ""
	}
	return fmt.Sprintf("gf_%s_%s.tar.gz", currentGOOS, currentGOARCH)
}

// downloadAndReplace downloads the .tar.gz archive at url, extracts the gf
// binary, and atomically replaces the currently running executable.
func downloadAndReplace(url string) error {
	exePath, err := executableFn()
	if err != nil {
		return fmt.Errorf("could not locate executable: %w", err)
	}

	resp, err := httpDoer.Get(url)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned HTTP %d", resp.StatusCode)
	}

	tmpDir, err := os.MkdirTemp("", "gf-update-*")
	if err != nil {
		return fmt.Errorf("temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	binPath := filepath.Join(tmpDir, "gf")
	if err := extractBinary(resp.Body, binPath); err != nil {
		return err
	}

	// Atomic replace: write to a temp file in the same directory, then rename.
	exeDir := filepath.Dir(exePath)
	tmp, err := os.CreateTemp(exeDir, ".gf-update-*")
	if err != nil {
		return fmt.Errorf("temp file: %w", err)
	}
	tmpPath := tmp.Name()

	removeTemp := true
	defer func() {
		if removeTemp {
			os.Remove(tmpPath)
		}
	}()

	src, err := os.Open(binPath)
	if err != nil {
		tmp.Close()
		return err
	}

	if _, err := io.Copy(tmp, src); err != nil {
		src.Close()
		tmp.Close()
		return fmt.Errorf("copy: %w", err)
	}
	src.Close()

	if err := tmp.Chmod(0755); err != nil {
		tmp.Close()
		return fmt.Errorf("chmod: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp: %w", err)
	}

	removeTemp = false
	return os.Rename(tmpPath, exePath)
}

// extractBinary reads a .tar.gz from r and writes the entry whose base name is
// "gf" to dst. Returns an error when no such entry is found.
func extractBinary(r io.Reader, dst string) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("gzip: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar: %w", err)
		}
		if filepath.Base(hdr.Name) != "gf" {
			continue
		}

		out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
		if err != nil {
			return fmt.Errorf("create: %w", err)
		}
		defer out.Close()

		if _, err := io.Copy(out, tr); err != nil {
			return fmt.Errorf("extract: %w", err)
		}
		return nil
	}
	return fmt.Errorf("gf binary not found in archive")
}

// isNewer reports whether latestTag represents a strictly greater semver than
// currentTag. Both accept an optional leading "v".
func isNewer(current, latest string) bool {
	c := parseSemver(strings.TrimPrefix(current, "v"))
	l := parseSemver(strings.TrimPrefix(latest, "v"))
	for i := 0; i < len(c) && i < len(l); i++ {
		if l[i] > c[i] {
			return true
		}
		if l[i] < c[i] {
			return false
		}
	}
	return len(l) > len(c)
}

// parseSemver splits a dot-separated version string into integer components.
func parseSemver(v string) []int {
	parts := strings.SplitN(v, ".", 3)
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		n, _ := strconv.Atoi(p)
		out = append(out, n)
	}
	return out
}
