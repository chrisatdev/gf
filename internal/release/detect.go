package release

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/chrisatdev/gf/internal/gitexec"
)

// FileType identifies how a project stores its version.
type FileType int

const (
	FileTypeJSON  FileType = iota // package.json, composer.json
	FileTypeTOML                  // Cargo.toml
	FileTypePlain                 // VERSION, version.txt
	FileTypeGoTag                 // go.mod — version lives only in git tags
)

// VersionFile describes where and how a project's version is stored.
type VersionFile struct {
	Path     string
	FileType FileType
}

type candidate struct {
	name     string
	fileType FileType
}

var versionCandidates = []candidate{
	{"package.json", FileTypeJSON},
	{"composer.json", FileTypeJSON},
	{"Cargo.toml", FileTypeTOML},
	{"VERSION", FileTypePlain},
	{"version.txt", FileTypePlain},
}

// readGitTag returns the most recent annotated git tag.
// Returns "v0.0.0" when no tags exist.
var readGitTag = func() (string, error) {
	out, err := gitexec.Output("describe", "--tags", "--abbrev=0")
	if err != nil || out == "" {
		return "v0.0.0", nil
	}
	return out, nil
}

// Detect searches dir for a supported version file, returning the first match.
// Falls back to FileTypeGoTag when go.mod is present and no other file is found.
func Detect(dir string) (*VersionFile, error) {
	for _, c := range versionCandidates {
		if _, err := os.Stat(filepath.Join(dir, c.name)); err == nil {
			return &VersionFile{Path: filepath.Join(dir, c.name), FileType: c.fileType}, nil
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
		return &VersionFile{FileType: FileTypeGoTag}, nil
	}
	return nil, fmt.Errorf("release: no version file found; create VERSION, package.json, or composer.json")
}

// ReadVersion reads the current version from the version file.
func ReadVersion(vf *VersionFile) (string, error) {
	if vf.FileType == FileTypeGoTag {
		return readGitTag()
	}
	data, err := os.ReadFile(vf.Path)
	if err != nil {
		return "", fmt.Errorf("release: %w", err)
	}
	return parseVersion(string(data), vf.FileType)
}

// WriteVersion writes the new version into the version file.
// The "v" prefix is stripped for JSON and TOML files (which use bare semver).
func WriteVersion(vf *VersionFile, version string) error {
	if vf.FileType == FileTypeGoTag {
		return nil
	}
	data, err := os.ReadFile(vf.Path)
	if err != nil {
		return fmt.Errorf("release: %w", err)
	}
	updated, err := replaceVersion(string(data), version, vf.FileType)
	if err != nil {
		return err
	}
	return os.WriteFile(vf.Path, []byte(updated), 0644)
}

// parseVersion extracts the version string from raw file content.
func parseVersion(content string, ft FileType) (string, error) {
	switch ft {
	case FileTypeJSON:
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(content), &m); err != nil {
			return "", fmt.Errorf("release: invalid JSON: %w", err)
		}
		v, ok := m["version"].(string)
		if !ok {
			return "", fmt.Errorf("release: missing \"version\" field")
		}
		return normalizeVersion(v), nil
	case FileTypeTOML:
		re := regexp.MustCompile(`(?m)^version\s*=\s*"([^"]+)"`)
		m := re.FindStringSubmatch(content)
		if len(m) < 2 {
			return "", fmt.Errorf("release: version field not found in TOML file")
		}
		return normalizeVersion(m[1]), nil
	case FileTypePlain:
		return normalizeVersion(strings.TrimSpace(content)), nil
	}
	return "", fmt.Errorf("release: unsupported file type")
}

// replaceVersion substitutes the version value inside raw file content.
func replaceVersion(content, version string, ft FileType) (string, error) {
	bare := strings.TrimPrefix(version, "v") // JSON and TOML omit the "v"
	switch ft {
	case FileTypeJSON:
		re := regexp.MustCompile(`("version"\s*:\s*")([^"]+)(")`)
		if !re.MatchString(content) {
			return "", fmt.Errorf("release: \"version\" field not found in JSON file")
		}
		return re.ReplaceAllString(content, "${1}"+bare+"${3}"), nil
	case FileTypeTOML:
		re := regexp.MustCompile(`(?m)^version\s*=\s*"[^"]*"`)
		if !re.MatchString(content) {
			return "", fmt.Errorf("release: version field not found in TOML file")
		}
		return re.ReplaceAllString(content, `version = "`+bare+`"`), nil
	case FileTypePlain:
		return version + "\n", nil
	}
	return "", fmt.Errorf("release: unsupported file type")
}

// normalizeVersion ensures the version carries a "v" prefix.
func normalizeVersion(v string) string {
	if !strings.HasPrefix(v, "v") {
		return "v" + v
	}
	return v
}
