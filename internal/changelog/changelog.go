package changelog

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const changelogDir = "changelogs"
const changelogFile = "CHANGELOG.md"

// Init initializes the changelog directory (NOT the file)
func Init() {
	// Create changelogs directory if it doesn't exist
	if err := os.MkdirAll(changelogDir, 0755); err != nil {
		fmt.Printf("Warning: Could not create changelog directory: %v\n", err)
	}
}

// EnsureExists creates CHANGELOG.md only if it doesn't exist
func EnsureExists() {
	if _, err := os.Stat(changelogFile); os.IsNotExist(err) {
		currentDate := time.Now().Format("January 2006")
		content := fmt.Sprintf("# CHANGELOG\n\n## [Unreleased] - %s\n\n### Added\n\n", currentDate)
		if err := os.WriteFile(changelogFile, []byte(content), 0644); err != nil {
			fmt.Printf("Warning: Could not create CHANGELOG.md: %v\n", err)
		}
	}
}

// AddEntry adds a new entry to the changelog
func AddEntry(commitType, message string) error {
	// Simple implementation - just append to file
	if _, err := os.Stat(changelogFile); os.IsNotExist(err) {
		return nil
	}

	// Extract the short message
	shortMsg := extractShortMessage(message)
	entry := fmt.Sprintf("- %s\n", shortMsg)

	// Determine section
	section := getSection(commitType)

	// Read file
	content, err := os.ReadFile(changelogFile)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")

	// Find or create section
	sectionFound := false
	for i, line := range lines {
		if strings.TrimSpace(line) == section {
			sectionFound = true
			// Insert entry after section header and any existing entries
			// Find the next section or end of file
			for j := i + 1; j < len(lines); j++ {
				if strings.HasPrefix(strings.TrimSpace(lines[j]), "### ") {
					// Insert before next section
					newLines := make([]string, len(lines)+1)
					copy(newLines[:i+1], lines[:i+1])
					newLines[i+1] = entry
					copy(newLines[i+2:], lines[i+1:])
					os.WriteFile(changelogFile, []byte(strings.Join(newLines, "\n")), 0644)
					return nil
				}
			}
			// Section exists but no next section found, append at end
			break
		}
	}

	if !sectionFound {
		// Add new section
		content = append(content, []byte("\n"+section+"\n"+entry)...)
		return os.WriteFile(changelogFile, content, 0644)
	}

	// Default: append to section
	content = append(content, []byte(entry)...)
	return os.WriteFile(changelogFile, content, 0644)
}

func extractShortMessage(message string) string {
	// Remove emoji and clean up
	msg := strings.TrimSpace(message)
	parts := strings.SplitN(msg, " ", 2)
	if len(parts) > 1 {
		msg = parts[1]
	}
	// Take first line only
	parts = strings.SplitN(msg, "\n", 2)
	return strings.TrimSpace(parts[0])
}

func getSection(commitType string) string {
	sections := map[string]string{
		"feat":     "### Added",
		"fix":      "### Fixed",
		"docs":     "### Documentation",
		"style":    "### Style",
		"refactor": "### Refactored",
		"test":     "### Testing",
		"build":    "### Build",
		"ci":       "### CI_CD",
		"perf":     "### Performance",
		"revert":   "### Reverted",
	}
	if section, ok := sections[commitType]; ok {
		return section
	}
	return "### Changed"
}

// StageChangelog stages the changelog file
func StageChangelog() {
	exec.Command("git", "add", changelogFile).Run()
}

// GetPath returns the path to the changelog directory
func GetPath() string {
	dir, _ := filepath.Abs(changelogDir)
	return dir
}
