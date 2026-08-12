package release

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetect_packageJSON(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "package.json"), `{"version":"1.0.0"}`)
	vf, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect() error: %v", err)
	}
	if vf.FileType != FileTypeJSON {
		t.Errorf("FileType = %v, want FileTypeJSON", vf.FileType)
	}
}

func TestDetect_VERSION(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "VERSION"), "v1.2.3\n")
	vf, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect() error: %v", err)
	}
	if vf.FileType != FileTypePlain {
		t.Errorf("FileType = %v, want FileTypePlain", vf.FileType)
	}
}

func TestDetect_gomod(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "go.mod"), "module example.com/foo\ngo 1.21\n")
	vf, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect() error: %v", err)
	}
	if vf.FileType != FileTypeGoTag {
		t.Errorf("FileType = %v, want FileTypeGoTag", vf.FileType)
	}
}

func TestDetect_packageJSONTakesPriorityOverGoMod(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "package.json"), `{"version":"2.0.0"}`)
	write(t, filepath.Join(dir, "go.mod"), "module example.com/foo\ngo 1.21\n")
	vf, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect() error: %v", err)
	}
	if vf.FileType != FileTypeJSON {
		t.Errorf("FileType = %v, want FileTypeJSON", vf.FileType)
	}
}

func TestDetect_noVersionFile(t *testing.T) {
	dir := t.TempDir()
	_, err := Detect(dir)
	if err == nil {
		t.Fatal("expected error for empty directory, got nil")
	}
}

func TestParseVersion_JSON(t *testing.T) {
	v, err := parseVersion(`{"name":"foo","version":"1.2.3"}`, FileTypeJSON)
	if err != nil {
		t.Fatalf("parseVersion() error: %v", err)
	}
	if v != "v1.2.3" {
		t.Errorf("parseVersion() = %q, want %q", v, "v1.2.3")
	}
}

func TestParseVersion_TOML(t *testing.T) {
	content := "[package]\nname = \"foo\"\nversion = \"0.1.0\"\n"
	v, err := parseVersion(content, FileTypeTOML)
	if err != nil {
		t.Fatalf("parseVersion() error: %v", err)
	}
	if v != "v0.1.0" {
		t.Errorf("parseVersion() = %q, want %q", v, "v0.1.0")
	}
}

func TestParseVersion_Plain(t *testing.T) {
	v, err := parseVersion("v1.0.0\n", FileTypePlain)
	if err != nil {
		t.Fatalf("parseVersion() error: %v", err)
	}
	if v != "v1.0.0" {
		t.Errorf("parseVersion() = %q, want %q", v, "v1.0.0")
	}
}

func TestReplaceVersion_JSON(t *testing.T) {
	content := `{"name":"foo","version":"1.0.0","main":"index.js"}`
	got, err := replaceVersion(content, "v2.0.0", FileTypeJSON)
	if err != nil {
		t.Fatalf("replaceVersion() error: %v", err)
	}
	want := `{"name":"foo","version":"2.0.0","main":"index.js"}`
	if got != want {
		t.Errorf("replaceVersion() = %q, want %q", got, want)
	}
}

func TestReplaceVersion_TOML(t *testing.T) {
	content := "[package]\nversion = \"0.1.0\"\n"
	got, err := replaceVersion(content, "v0.2.0", FileTypeTOML)
	if err != nil {
		t.Fatalf("replaceVersion() error: %v", err)
	}
	want := "[package]\nversion = \"0.2.0\"\n"
	if got != want {
		t.Errorf("replaceVersion() = %q, want %q", got, want)
	}
}

func TestReplaceVersion_Plain(t *testing.T) {
	got, err := replaceVersion("v1.0.0\n", "v1.1.0", FileTypePlain)
	if err != nil {
		t.Fatalf("replaceVersion() error: %v", err)
	}
	if got != "v1.1.0\n" {
		t.Errorf("replaceVersion() = %q, want %q", got, "v1.1.0\n")
	}
}

func TestNormalizeVersion(t *testing.T) {
	if got := normalizeVersion("1.2.3"); got != "v1.2.3" {
		t.Errorf("normalizeVersion(%q) = %q, want %q", "1.2.3", got, "v1.2.3")
	}
	if got := normalizeVersion("v1.2.3"); got != "v1.2.3" {
		t.Errorf("normalizeVersion(%q) = %q, want %q", "v1.2.3", got, "v1.2.3")
	}
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
