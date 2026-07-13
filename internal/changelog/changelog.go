package changelog

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

// EnsureExists creates the changelog file with a minimal header if it does not
// already exist. It is a no-op when the file is present.
func EnsureExists(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	return os.WriteFile(path, []byte("# Changelog\n"), 0644)
}

// Prepend adds entry (without the leading "- ") under today's date section.
// Creates the file if it does not exist. Deduplicates exact duplicate entries
// within the same date section.
func Prepend(path, entry string) error {
	if err := EnsureExists(path); err != nil {
		return fmt.Errorf("gf changelog: %w", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("gf changelog: %w", err)
	}
	today := time.Now().Format("2006-01-02")
	updated := insertEntry(string(data), today, "- "+entry)
	return os.WriteFile(path, []byte(updated), 0644)
}

// MergeConflict resolves git conflict markers in the changelog at path.
// Both sides are merged, exact-duplicate entries are removed, and date sections
// are sorted newest-first. The file is rewritten cleanly without markers.
// Returns nil immediately when no conflict markers are present.
func MergeConflict(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("gf changelog: %w", err)
	}
	if !strings.Contains(string(data), "<<<<<<<") {
		return nil
	}
	resolved, err := resolveConflict(string(data))
	if err != nil {
		return fmt.Errorf("gf changelog: %w", err)
	}
	return os.WriteFile(path, []byte(resolved), 0644)
}

// --- internal helpers ---

// insertEntry inserts entryLine into the changelog content under the given date
// section header ("## YYYY-MM-DD"). When today's section already exists, the
// entry is placed at the top of its entry list. When it does not exist, a new
// section is inserted right after "# Changelog". Duplicate entry lines within
// the same date section are silently dropped.
func insertEntry(content, today, entryLine string) string {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(content, "\n")

	// Normalize: trim a single trailing empty line produced by Split on a
	// newline-terminated file so our rebuilds don't accumulate extra blanks.
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	todayHeader := "## " + today

	// Search for an existing today section.
	for i, line := range lines {
		if line != todayHeader {
			continue
		}
		// Locate insertion point: skip the optional blank line after the header.
		j := i + 1
		if j < len(lines) && lines[j] == "" {
			j++
		}
		// Scan entries in this section to detect duplicates.
		for k := j; k < len(lines); k++ {
			if strings.HasPrefix(lines[k], "## ") {
				break // next section boundary
			}
			if lines[k] == entryLine {
				// Exact duplicate — return content unchanged.
				return strings.Join(append(lines, ""), "\n")
			}
		}
		// Insert at j (top of the entry list, below the optional blank line).
		out := make([]string, 0, len(lines)+2)
		out = append(out, lines[:j]...)
		out = append(out, entryLine)
		out = append(out, lines[j:]...)
		out = append(out, "")
		return strings.Join(out, "\n")
	}

	// No today section: insert a new one right after "# Changelog".
	for i, line := range lines {
		if line != "# Changelog" {
			continue
		}
		newSec := []string{"", todayHeader, "", entryLine}
		out := make([]string, 0, len(lines)+len(newSec)+1)
		out = append(out, lines[:i+1]...)
		out = append(out, newSec...)
		out = append(out, lines[i+1:]...)
		out = append(out, "")
		return strings.Join(out, "\n")
	}

	// No "# Changelog" header at all: create the file content from scratch.
	return "# Changelog\n\n" + todayHeader + "\n\n" + entryLine + "\n"
}

// section represents a dated changelog section with its entries.
type section struct {
	date    string
	entries []string
}

// parseSections extracts ## YYYY-MM-DD sections and their "- …" entries.
func parseSections(text string) []section {
	var result []section
	var cur *section
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.HasPrefix(line, "## ") {
			result = append(result, section{date: strings.TrimSpace(line[3:])})
			cur = &result[len(result)-1]
		} else if cur != nil && strings.HasPrefix(line, "- ") {
			cur.entries = append(cur.entries, line)
		}
	}
	return result
}

// mergeSections combines two sets of sections, preserving order of first
// appearance for each date.
func mergeSections(a, b []section) []section {
	byDate := make(map[string][]string)
	var order []string
	seen := make(map[string]bool)

	add := func(s section) {
		if !seen[s.date] {
			seen[s.date] = true
			order = append(order, s.date)
		}
		byDate[s.date] = append(byDate[s.date], s.entries...)
	}
	for _, s := range a {
		add(s)
	}
	for _, s := range b {
		add(s)
	}

	out := make([]section, 0, len(order))
	for _, date := range order {
		out = append(out, section{date: date, entries: dedupeSlice(byDate[date])})
	}
	return out
}

// dedupeSlice removes exact duplicate strings while preserving first-seen order.
func dedupeSlice(in []string) []string {
	seen := make(map[string]bool, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

// resolveConflict parses a single-block git conflict in content, merges both
// sides, deduplicates, sorts sections newest-first, and returns clean content.
func resolveConflict(content string) (string, error) {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(content, "\n")

	state := "before"
	var before, ours, theirs, after []string

	for _, line := range lines {
		switch {
		case (strings.HasPrefix(line, "<<<<<<< ") || line == "<<<<<<<") && state == "before":
			state = "ours"
		case line == "=======" && state == "ours":
			state = "theirs"
		case (strings.HasPrefix(line, ">>>>>>> ") || line == ">>>>>>>") && state == "theirs":
			state = "after"
		default:
			switch state {
			case "before":
				before = append(before, line)
			case "ours":
				ours = append(ours, line)
			case "theirs":
				theirs = append(theirs, line)
			case "after":
				after = append(after, line)
			}
		}
	}

	if state != "after" {
		return "", fmt.Errorf("malformed conflict markers in changelog")
	}

	oSecs := parseSections(strings.Join(ours, "\n"))
	tSecs := parseSections(strings.Join(theirs, "\n"))
	merged := mergeSections(oSecs, tSecs)

	sort.Slice(merged, func(i, j int) bool {
		return merged[i].date > merged[j].date // lexicographic: newer ISO date is greater
	})

	var sb strings.Builder

	// before typically ends with "# Changelog" possibly followed by a blank line.
	beforeStr := strings.TrimRight(strings.Join(before, "\n"), "\n")
	sb.WriteString(beforeStr)

	for _, sec := range merged {
		sb.WriteString("\n\n## " + sec.date + "\n\n")
		for _, e := range sec.entries {
			sb.WriteString(e + "\n")
		}
	}

	// Append any content after the conflict block (rare in a changelog).
	afterStr := strings.TrimLeft(strings.Join(after, "\n"), "\n")
	if afterStr != "" {
		sb.WriteString("\n" + afterStr)
	} else {
		sb.WriteString("\n")
	}

	return sb.String(), nil
}
