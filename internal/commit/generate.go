package commit

import (
	"fmt"
	"path/filepath"
	"strings"
)

// CommitMessage holds the parsed parts of a conventional commit.
type CommitMessage struct {
	Type        string
	Scope       string
	Description string
	Body        []string
	Breaking    bool
}

// String formats the commit message in Conventional Commits style.
// Header: type(scope): description
// Body separated by a blank line when present.
func (m CommitMessage) String() string {
	var header string
	if m.Scope != "" {
		header = fmt.Sprintf("%s(%s): %s", m.Type, m.Scope, m.Description)
	} else {
		header = fmt.Sprintf("%s: %s", m.Type, m.Description)
	}
	if len(m.Body) == 0 {
		return header
	}
	return header + "\n\n" + strings.Join(m.Body, "\n")
}

// InferFromBranch builds a CommitMessage from a branch name and the output of
// git status --porcelain.
func InferFromBranch(branch, porcelain string) CommitMessage {
	return CommitMessage{
		Type:        InferType(branch, porcelain),
		Scope:       InferScope(branch),
		Description: inferDescription(porcelain),
		Body:        BuildBody(porcelain),
	}
}

// InferType returns the conventional commit type.
// Priority: docs-only > test-only > branch prefix > chore fallback.
func InferType(branch, porcelain string) string {
	if isDocsOnly(porcelain) {
		return "docs"
	}
	if isTestOnly(porcelain) {
		return "test"
	}
	for _, p := range branchPrefixes {
		if strings.HasPrefix(branch, p.prefix) {
			return p.typ
		}
	}
	return "chore"
}

// InferScope extracts the scope from the branch name: the part after the first
// '/', truncated to 30 characters. Returns an empty string if no slash is present.
func InferScope(branch string) string {
	idx := strings.Index(branch, "/")
	if idx < 0 {
		return ""
	}
	scope := branch[idx+1:]
	if len(scope) > 30 {
		scope = scope[:30]
	}
	return scope
}

// BuildBody parses git status --porcelain lines into conventional body lines.
//
//	A  path       → "add: path"
//	M  path       → "modify: path"
//	D  path       → "remove: path"
//	R  old -> new → "rename: old -> new"
func BuildBody(porcelain string) []string {
	var lines []string
	for _, raw := range strings.Split(porcelain, "\n") {
		raw = strings.TrimRight(raw, "\r")
		if len(raw) < 3 {
			continue
		}
		xy := raw[:2]
		rest := strings.TrimSpace(raw[2:])
		if rest == "" {
			continue
		}
		s := primaryStatus(xy)
		if s == ' ' || s == '?' || s == '!' {
			continue
		}
		switch s {
		case 'A':
			lines = append(lines, "add: "+rest)
		case 'M':
			lines = append(lines, "modify: "+rest)
		case 'D':
			lines = append(lines, "remove: "+rest)
		case 'R':
			lines = append(lines, "rename: "+rest)
		}
	}
	return lines
}

// --- internal helpers ---

var branchPrefixes = []struct {
	prefix string
	typ    string
}{
	{"feat/", "feat"},
	{"bug/", "fix"},
	{"fix/", "fix"},
	{"iss/", "fix"},
	{"task/", "chore"},
}

// primaryStatus returns the effective status character from an XY pair.
// Index status (X) is preferred; falls back to worktree status (Y).
func primaryStatus(xy string) byte {
	if len(xy) == 0 {
		return ' '
	}
	if xy[0] != ' ' {
		return xy[0]
	}
	if len(xy) > 1 {
		return xy[1]
	}
	return ' '
}

// pathsFromPorcelain extracts the affected file paths from porcelain output.
// For renames, returns the new (destination) path.
func pathsFromPorcelain(porcelain string) []string {
	var paths []string
	for _, raw := range strings.Split(porcelain, "\n") {
		raw = strings.TrimRight(raw, "\r")
		if len(raw) < 3 {
			continue
		}
		xy := raw[:2]
		rest := strings.TrimSpace(raw[2:])
		if rest == "" {
			continue
		}
		s := primaryStatus(xy)
		if s == ' ' || s == '?' || s == '!' {
			continue
		}
		if strings.Contains(rest, " -> ") {
			parts := strings.SplitN(rest, " -> ", 2)
			rest = strings.TrimSpace(parts[1])
		}
		paths = append(paths, rest)
	}
	return paths
}

// isDocsOnly reports whether every changed file is a documentation file.
func isDocsOnly(porcelain string) bool {
	paths := pathsFromPorcelain(porcelain)
	if len(paths) == 0 {
		return false
	}
	for _, p := range paths {
		if !isDocFile(p) {
			return false
		}
	}
	return true
}

// isDocFile reports whether path is a documentation file.
// Executable extensions (.sh, .go, etc.) are never treated as docs even when
// the filename starts with README or CHANGELOG.
func isDocFile(path string) bool {
	base := filepath.Base(path)
	dir := filepath.Dir(path)
	ext := strings.ToLower(filepath.Ext(base))

	// Explicit doc extensions.
	switch ext {
	case ".md", ".rst", ".txt", ".adoc":
		return true
	}

	// Explicit non-doc extensions always win.
	switch ext {
	case ".sh", ".bash", ".py", ".go", ".js", ".ts", ".rb",
		".java", ".c", ".cpp", ".h", ".cs", ".php", ".swift":
		return false
	}

	// README / CHANGELOG without an extension.
	if ext == "" {
		upper := strings.ToUpper(base)
		if upper == "README" || upper == "CHANGELOG" {
			return true
		}
	}

	// Files inside a docs/ directory (any extension not blocked above).
	if dir == "docs" || strings.HasPrefix(dir, "docs/") {
		return true
	}

	return false
}

// isTestOnly reports whether every changed file is a Go test file.
func isTestOnly(porcelain string) bool {
	paths := pathsFromPorcelain(porcelain)
	if len(paths) == 0 {
		return false
	}
	for _, p := range paths {
		if !isTestFile(p) {
			return false
		}
	}
	return true
}

// isTestFile reports whether path is a Go test file.
func isTestFile(path string) bool {
	base := filepath.Base(path)
	return strings.HasSuffix(base, "_test.go") || strings.HasPrefix(base, "test_")
}

// inferDescription builds a short human-readable description from porcelain.
//
//	1 file  → "add <base>", "remove <base>", or "update <base>"
//	N files → "update N files"
func inferDescription(porcelain string) string {
	type fileChange struct {
		status byte
		path   string
	}
	var changes []fileChange

	for _, raw := range strings.Split(porcelain, "\n") {
		raw = strings.TrimRight(raw, "\r")
		if len(raw) < 3 {
			continue
		}
		xy := raw[:2]
		rest := strings.TrimSpace(raw[2:])
		if rest == "" {
			continue
		}
		s := primaryStatus(xy)
		if s == ' ' || s == '?' || s == '!' {
			continue
		}
		p := rest
		if strings.Contains(rest, " -> ") {
			parts := strings.SplitN(rest, " -> ", 2)
			p = strings.TrimSpace(parts[1])
		}
		changes = append(changes, fileChange{status: s, path: p})
	}

	if len(changes) == 0 {
		return "update files"
	}
	if len(changes) == 1 {
		base := filepath.Base(changes[0].path)
		switch changes[0].status {
		case 'A':
			return "add " + base
		case 'D':
			return "remove " + base
		default:
			return "update " + base
		}
	}
	return fmt.Sprintf("update %d files", len(changes))
}
