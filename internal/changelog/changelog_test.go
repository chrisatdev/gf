package changelog_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/chrisatdev/gf/internal/changelog"
)

// today returns the current date in YYYY-MM-DD format, mirroring the package.
func today() string {
	return time.Now().Format("2006-01-02")
}

// readFile reads a file and returns its content as a string.
func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("readFile: %v", err)
	}
	return string(data)
}

// writeFile writes content to a file at path.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writeFile: %v", err)
	}
}

// --- EnsureExists ---

func TestEnsureExists_createsFileWithHeader(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CHANGELOG.md")

	if err := changelog.EnsureExists(path); err != nil {
		t.Fatalf("EnsureExists: %v", err)
	}

	got := readFile(t, path)
	if !strings.Contains(got, "# Changelog") {
		t.Errorf("expected '# Changelog' header, got:\n%s", got)
	}
}

func TestEnsureExists_noopWhenFileExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CHANGELOG.md")

	original := "# My Custom Changelog\n\n- existing content\n"
	writeFile(t, path, original)

	if err := changelog.EnsureExists(path); err != nil {
		t.Fatalf("EnsureExists: %v", err)
	}

	got := readFile(t, path)
	if got != original {
		t.Errorf("EnsureExists modified existing file\ngot:  %q\nwant: %q", got, original)
	}
}

// --- Prepend ---

func TestPrepend_missingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CHANGELOG.md")

	if err := changelog.Prepend(path, "feat(commit): add commit generator"); err != nil {
		t.Fatalf("Prepend: %v", err)
	}

	got := readFile(t, path)
	wantHeader := "# Changelog"
	wantSection := "## " + today()
	wantEntry := "- feat(commit): add commit generator"

	for _, want := range []string{wantHeader, wantSection, wantEntry} {
		if !strings.Contains(got, want) {
			t.Errorf("Prepend output missing %q\ngot:\n%s", want, got)
		}
	}
}

func TestPrepend_addsToExistingTodaySection(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CHANGELOG.md")

	initial := "# Changelog\n\n## " + today() + "\n\n- first entry\n"
	writeFile(t, path, initial)

	if err := changelog.Prepend(path, "second entry"); err != nil {
		t.Fatalf("Prepend: %v", err)
	}

	got := readFile(t, path)
	if !strings.Contains(got, "- second entry") {
		t.Errorf("new entry not found in output:\n%s", got)
	}
	if !strings.Contains(got, "- first entry") {
		t.Errorf("existing entry removed from output:\n%s", got)
	}
	// second entry should appear before first entry (prepend semantics)
	secondIdx := strings.Index(got, "- second entry")
	firstIdx := strings.Index(got, "- first entry")
	if secondIdx > firstIdx {
		t.Errorf("new entry should appear before existing entry\ngot:\n%s", got)
	}
	// Section header should appear only once
	if count := strings.Count(got, "## "+today()); count != 1 {
		t.Errorf("today section header appears %d times, want 1\ngot:\n%s", count, got)
	}
}

func TestPrepend_createsNewDateSectionWhenNeeded(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CHANGELOG.md")

	// File has a yesterday section; today's section should be inserted above.
	initial := "# Changelog\n\n## 2020-01-01\n\n- old entry\n"
	writeFile(t, path, initial)

	if err := changelog.Prepend(path, "new entry"); err != nil {
		t.Fatalf("Prepend: %v", err)
	}

	got := readFile(t, path)
	if !strings.Contains(got, "## "+today()) {
		t.Errorf("today's section header not found:\n%s", got)
	}
	if !strings.Contains(got, "- new entry") {
		t.Errorf("new entry not found:\n%s", got)
	}
	if !strings.Contains(got, "## 2020-01-01") {
		t.Errorf("old section removed:\n%s", got)
	}
	// Today's section should come before the old section.
	todayIdx := strings.Index(got, "## "+today())
	oldIdx := strings.Index(got, "## 2020-01-01")
	if todayIdx > oldIdx {
		t.Errorf("today's section should precede old section\ngot:\n%s", got)
	}
}

func TestPrepend_deduplicatesSameDayEntry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CHANGELOG.md")

	entry := "feat(commit): add commit generator"
	initial := "# Changelog\n\n## " + today() + "\n\n- " + entry + "\n"
	writeFile(t, path, initial)

	if err := changelog.Prepend(path, entry); err != nil {
		t.Fatalf("Prepend: %v", err)
	}

	got := readFile(t, path)
	count := strings.Count(got, "- "+entry)
	if count != 1 {
		t.Errorf("duplicate entry: found %d occurrences, want 1\ngot:\n%s", count, got)
	}
}

// --- MergeConflict ---

func TestMergeConflict_removesMarkersAndPreservesBothEntries(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CHANGELOG.md")

	conflictContent := "# Changelog\n\n" +
		"<<<<<<< HEAD\n" +
		"## 2026-07-13\n\n" +
		"- feat(commit): add commit generator\n" +
		"=======\n" +
		"## 2026-07-13\n\n" +
		"- feat(changelog): add changelog writer\n" +
		">>>>>>> feature\n"

	writeFile(t, path, conflictContent)

	if err := changelog.MergeConflict(path); err != nil {
		t.Fatalf("MergeConflict: %v", err)
	}

	got := readFile(t, path)

	// No conflict markers remain.
	for _, marker := range []string{"<<<<<<<", "=======", ">>>>>>>"} {
		if strings.Contains(got, marker) {
			t.Errorf("conflict marker %q still present:\n%s", marker, got)
		}
	}
	// Both entries are present.
	if !strings.Contains(got, "- feat(commit): add commit generator") {
		t.Errorf("ours entry missing:\n%s", got)
	}
	if !strings.Contains(got, "- feat(changelog): add changelog writer") {
		t.Errorf("theirs entry missing:\n%s", got)
	}
	// Header preserved.
	if !strings.Contains(got, "# Changelog") {
		t.Errorf("# Changelog header missing:\n%s", got)
	}
}

func TestMergeConflict_deduplicatesIdenticalEntries(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CHANGELOG.md")

	// Both sides contain the same entry.
	conflictContent := "# Changelog\n\n" +
		"<<<<<<< HEAD\n" +
		"## 2026-07-13\n\n" +
		"- feat: same entry\n" +
		"=======\n" +
		"## 2026-07-13\n\n" +
		"- feat: same entry\n" +
		">>>>>>> branch\n"

	writeFile(t, path, conflictContent)

	if err := changelog.MergeConflict(path); err != nil {
		t.Fatalf("MergeConflict: %v", err)
	}

	got := readFile(t, path)
	count := strings.Count(got, "- feat: same entry")
	if count != 1 {
		t.Errorf("duplicate entry: found %d occurrences, want 1\ngot:\n%s", count, got)
	}
}

func TestMergeConflict_sortsSectionsNewestFirst(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CHANGELOG.md")

	conflictContent := "# Changelog\n\n" +
		"<<<<<<< HEAD\n" +
		"## 2026-07-13\n\n- new entry\n" +
		"=======\n" +
		"## 2026-07-12\n\n- old entry\n" +
		">>>>>>> branch\n"

	writeFile(t, path, conflictContent)

	if err := changelog.MergeConflict(path); err != nil {
		t.Fatalf("MergeConflict: %v", err)
	}

	got := readFile(t, path)
	newIdx := strings.Index(got, "## 2026-07-13")
	oldIdx := strings.Index(got, "## 2026-07-12")
	if newIdx < 0 || oldIdx < 0 {
		t.Fatalf("section headers missing in output:\n%s", got)
	}
	if newIdx > oldIdx {
		t.Errorf("newer section should appear before older section\ngot:\n%s", got)
	}
}

func TestMergeConflict_noopWhenNoMarkers(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CHANGELOG.md")

	original := "# Changelog\n\n## 2026-07-13\n\n- clean entry\n"
	writeFile(t, path, original)

	if err := changelog.MergeConflict(path); err != nil {
		t.Fatalf("MergeConflict: %v", err)
	}

	got := readFile(t, path)
	if got != original {
		t.Errorf("MergeConflict modified clean file\ngot:  %q\nwant: %q", got, original)
	}
}
