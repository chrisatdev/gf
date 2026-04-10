package commit

import (
	"strings"
)

type Detector struct{}

func NewDetector() *Detector {
	return &Detector{}
}

func (d *Detector) Detect(files []string, diffContent string) string {
	// Priority 1: File path patterns
	for _, file := range files {
		switch {
		case matchesAny(file, "src/", "lib/", "components/", "pages/", "views/", "app/"):
			return "feat"
		case matchesAny(file, "fix/", "bug/", "*fix*", "*bug*"):
			return "fix"
		case matchesAny(file, "README", "docs/", ".md"):
			return "docs"
		case matchesAny(file, ".css", ".scss", ".less"):
			return "style"
		case matchesAny(file, "test/", "spec/", ".test.", ".spec."):
			return "test"
		case matchesAny(file, "package.json", "tsconfig", "webpack", "vite", "vite.config"):
			return "build"
		}
	}

	// Priority 2: Diff content keywords
	diffLower := strings.ToLower(diffContent)
	if strings.Contains(diffLower, "fix:") || strings.Contains(diffLower, "bug:") {
		return "fix"
	}
	if strings.Contains(diffLower, "perf:") || strings.Contains(diffLower, "optimize:") {
		return "perf"
	}

	return "chore"
}

func matchesAny(file string, patterns ...string) bool {
	for _, pattern := range patterns {
		if len(pattern) > 0 && pattern[0] == '*' {
			// Glob pattern
			if len(pattern) > 1 && strings.Contains(file, pattern[1:]) {
				return true
			}
		} else if strings.Contains(file, pattern) {
			return true
		}
	}
	return false
}
