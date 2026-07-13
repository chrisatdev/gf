package platformdetect_test

import (
	"strings"
	"testing"

	"github.com/chrisatdev/gf/internal/platformdetect"
)

func TestParseRemoteURL(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		wantPlatform    string
		wantProjectPath string
		wantNoToken     string // substring that must NOT appear in HTTPSURL
		wantNoSSH       bool   // HTTPSURL must not contain "git@"
	}{
		{
			name:            "SSH github",
			input:           "git@github.com:chrisatdev/gf.git",
			wantPlatform:    "github",
			wantProjectPath: "chrisatdev/gf",
			wantNoSSH:       true,
		},
		{
			name:            "HTTPS github",
			input:           "https://github.com/chrisatdev/gf.git",
			wantPlatform:    "github",
			wantProjectPath: "chrisatdev/gf",
		},
		{
			name:            "SSH gitlab",
			input:           "git@gitlab.com:team/project.git",
			wantPlatform:    "gitlab",
			wantProjectPath: "team/project",
			wantNoSSH:       true,
		},
		{
			name:            "HTTPS gitlab with credentials",
			input:           "https://user:s3cr3t@gitlab.edge.com.py/team/project.git",
			wantPlatform:    "gitlab",
			wantProjectPath: "team/project",
			wantNoToken:     "s3cr3t",
		},
		{
			name:            "SSH gitlab internal host",
			input:           "git@gitlab.internal:team/project.git",
			wantPlatform:    "gitlab",
			wantProjectPath: "team/project",
			wantNoSSH:       true,
		},
		{
			name:            "SSH bitbucket unknown",
			input:           "git@bitbucket.org:team/project.git",
			wantPlatform:    "unknown",
			wantProjectPath: "team/project",
		},
		{
			name:            "HTTPS gitlab company host",
			input:           "https://gitlab.mycompany.com/org/repo.git",
			wantPlatform:    "gitlab",
			wantProjectPath: "org/repo",
		},
		{
			name:            "SSH github no .git suffix",
			input:           "git@github.com:owner/repo",
			wantPlatform:    "github",
			wantProjectPath: "owner/repo",
			wantNoSSH:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := platformdetect.ParseRemoteURL(tt.input)

			if got.Platform != tt.wantPlatform {
				t.Errorf("Platform = %q, want %q", got.Platform, tt.wantPlatform)
			}
			if got.ProjectPath != tt.wantProjectPath {
				t.Errorf("ProjectPath = %q, want %q", got.ProjectPath, tt.wantProjectPath)
			}
			if tt.wantNoSSH && strings.Contains(got.HTTPSURL, "git@") {
				t.Errorf("HTTPSURL contains SSH prefix: %q", got.HTTPSURL)
			}
			if tt.wantNoToken != "" && strings.Contains(got.HTTPSURL, tt.wantNoToken) {
				t.Errorf("HTTPSURL contains credential %q: %q", tt.wantNoToken, got.HTTPSURL)
			}
		})
	}
}
