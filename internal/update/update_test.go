package update

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// makeRelease builds a Release with asset download URLs rooted at baseURL.
func makeRelease(tag string, assetNames []string, baseURL string) Release {
	assets := make([]Asset, len(assetNames))
	for i, name := range assetNames {
		assets[i] = Asset{
			Name:               name,
			BrowserDownloadURL: baseURL + "/download/" + name,
		}
	}
	return Release{TagName: tag, Assets: assets}
}

// serveRelease creates a test server that always returns the given release JSON.
func serveRelease(t *testing.T, rel Release) *httptest.Server {
	t.Helper()
	data, _ := json.Marshal(rel)
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}))
}

// makeTarGz builds a minimal tar.gz archive containing a single file named "gf".
func makeTarGz(t *testing.T, content string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)

	data := []byte(content)
	_ = tw.WriteHeader(&tar.Header{Name: "gf", Mode: 0755, Size: int64(len(data))})
	_, _ = tw.Write(data)
	_ = tw.Close()
	_ = gz.Close()
	return buf.Bytes()
}

// patchPlatform overrides GOOS and GOARCH for the duration of a test.
func patchPlatform(t *testing.T, goos, goarch string) {
	t.Helper()
	origOS, origArch := currentGOOS, currentGOARCH
	t.Cleanup(func() { currentGOOS, currentGOARCH = origOS, origArch })
	currentGOOS, currentGOARCH = goos, goarch
}

// patchAPIBase redirects CheckLatest to srv for the duration of a test.
func patchAPIBase(t *testing.T, srv *httptest.Server) {
	t.Helper()
	orig := githubAPIBase
	t.Cleanup(func() { githubAPIBase = orig })
	githubAPIBase = srv.URL
}

// ---------------------------------------------------------------------------
// AssetName
// ---------------------------------------------------------------------------

func TestAssetName(t *testing.T) {
	tests := []struct {
		name   string
		goos   string
		goarch string
		want   string
	}{
		{"linux amd64", "linux", "amd64", "gf_linux_amd64.tar.gz"},
		{"linux arm64", "linux", "arm64", "gf_linux_arm64.tar.gz"},
		{"darwin amd64", "darwin", "amd64", "gf_darwin_amd64.tar.gz"},
		{"darwin arm64", "darwin", "arm64", "gf_darwin_arm64.tar.gz"},
		{"windows unsupported", "windows", "amd64", ""},
		{"freebsd unsupported", "freebsd", "arm64", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patchPlatform(t, tt.goos, tt.goarch)
			got := AssetName()
			if got != tt.want {
				t.Errorf("AssetName() = %q, want %q", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// CheckLatest
// ---------------------------------------------------------------------------

func TestCheckLatest(t *testing.T) {
	t.Run("success returns release", func(t *testing.T) {
		want := makeRelease("v1.2.3", []string{"gf_linux_amd64.tar.gz"}, "http://fake")
		srv := serveRelease(t, want)
		defer srv.Close()
		patchAPIBase(t, srv)

		got, err := CheckLatest("test/repo")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.TagName != "v1.2.3" {
			t.Errorf("TagName = %q, want v1.2.3", got.TagName)
		}
		if len(got.Assets) != 1 {
			t.Errorf("Assets len = %d, want 1", len(got.Assets))
		}
	})

	t.Run("non-200 returns error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer srv.Close()
		patchAPIBase(t, srv)

		_, err := CheckLatest("test/repo")
		if err == nil {
			t.Fatal("expected error for 404, got nil")
		}
		if !strings.Contains(err.Error(), "404") {
			t.Errorf("error %q does not mention 404", err.Error())
		}
	})
}

// ---------------------------------------------------------------------------
// Execute
// ---------------------------------------------------------------------------

func TestExecute(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		goos           string
		goarch         string
		releaseTag     string
		assetNames     []string
		// when true, downloadReplaceFn is expected to be called
		expectDownload bool
		wantErr        bool
		wantErrContain string
	}{
		{
			name:           "already up to date",
			currentVersion: "v1.0.0",
			goos:           "linux",
			goarch:         "amd64",
			releaseTag:     "v1.0.0",
			assetNames:     []string{"gf_linux_amd64.tar.gz"},
			expectDownload: false,
		},
		{
			name:           "newer version with matching asset triggers download",
			currentVersion: "v1.0.0",
			goos:           "linux",
			goarch:         "amd64",
			releaseTag:     "v1.1.0",
			assetNames:     []string{"gf_linux_amd64.tar.gz"},
			expectDownload: true,
		},
		{
			name:           "asset missing returns error naming expected asset",
			currentVersion: "v1.0.0",
			goos:           "linux",
			goarch:         "amd64",
			releaseTag:     "v1.1.0",
			assetNames:     []string{"gf_darwin_arm64.tar.gz"},
			wantErr:        true,
			wantErrContain: "gf_linux_amd64.tar.gz",
		},
		{
			name:           "unsupported platform returns error",
			currentVersion: "v1.0.0",
			goos:           "windows",
			goarch:         "amd64",
			releaseTag:     "v1.1.0",
			assetNames:     []string{"gf_windows_amd64.zip"},
			wantErr:        true,
			wantErrContain: "unsupported platform",
		},
		{
			name:           "dev version reports latest and returns nil",
			currentVersion: "dev",
			goos:           "linux",
			goarch:         "amd64",
			releaseTag:     "v1.0.0",
			assetNames:     []string{"gf_linux_amd64.tar.gz"},
			expectDownload: false,
		},
		{
			name:           "empty version treated same as dev",
			currentVersion: "",
			goos:           "linux",
			goarch:         "amd64",
			releaseTag:     "v1.0.0",
			assetNames:     []string{"gf_linux_amd64.tar.gz"},
			expectDownload: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the API server for this test case.
			rel := makeRelease(tt.releaseTag, tt.assetNames, "http://fake-download")
			srv := serveRelease(t, rel)
			defer srv.Close()
			patchAPIBase(t, srv)

			// Patch platform.
			patchPlatform(t, tt.goos, tt.goarch)

			// Mock downloadReplaceFn to avoid real network/filesystem work.
			downloaded := false
			origDl := downloadReplaceFn
			t.Cleanup(func() { downloadReplaceFn = origDl })
			downloadReplaceFn = func(_ string) error {
				downloaded = true
				return nil
			}

			err := Execute(tt.currentVersion)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrContain != "" && !strings.Contains(err.Error(), tt.wantErrContain) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.wantErrContain)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.expectDownload && !downloaded {
				t.Error("expected downloadReplaceFn to be called, but it was not")
			}
			if !tt.expectDownload && downloaded {
				t.Error("downloadReplaceFn was called but should not have been")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// extractBinary
// ---------------------------------------------------------------------------

func TestExtractBinary(t *testing.T) {
	t.Run("extracts gf from tar.gz", func(t *testing.T) {
		content := "#!/bin/sh\necho gf-test"
		archive := makeTarGz(t, content)
		dst := filepath.Join(t.TempDir(), "gf-out")

		if err := extractBinary(bytes.NewReader(archive), dst); err != nil {
			t.Fatalf("extractBinary: %v", err)
		}

		got, err := os.ReadFile(dst)
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}
		if string(got) != content {
			t.Errorf("got %q, want %q", got, content)
		}
	})

	t.Run("error when binary not in archive", func(t *testing.T) {
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		tw := tar.NewWriter(gz)
		data := []byte("not-gf")
		_ = tw.WriteHeader(&tar.Header{Name: "other-file", Mode: 0755, Size: int64(len(data))})
		_, _ = tw.Write(data)
		_ = tw.Close()
		_ = gz.Close()

		dst := filepath.Join(t.TempDir(), "gf-out")
		err := extractBinary(bytes.NewReader(buf.Bytes()), dst)
		if err == nil {
			t.Fatal("expected error when gf not in archive, got nil")
		}
	})
}

// ---------------------------------------------------------------------------
// isNewer (semver-aware comparison)
// ---------------------------------------------------------------------------

func TestIsNewer(t *testing.T) {
	tests := []struct {
		current string
		latest  string
		want    bool
	}{
		{"v1.0.0", "v1.0.0", false},
		{"v1.0.0", "v1.1.0", true},
		{"v1.1.0", "v1.0.0", false},
		{"v1.9.0", "v1.10.0", true}, // double-digit minor
		{"v2.0.0", "v1.9.9", false},
		{"1.0.0", "1.0.1", true},    // no leading v
	}
	for _, tt := range tests {
		got := isNewer(tt.current, tt.latest)
		if got != tt.want {
			t.Errorf("isNewer(%q, %q) = %v, want %v", tt.current, tt.latest, got, tt.want)
		}
	}
}
