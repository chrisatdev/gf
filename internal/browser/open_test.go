package browser

import (
	"strings"
	"testing"
)

func TestOpen(t *testing.T) {
	tests := []struct {
		name     string
		os       string
		wantProg string
		wantErr  bool
	}{
		{"linux selects xdg-open", "linux", "xdg-open", false},
		{"darwin selects open", "darwin", "open", false},
		{"windows returns error", "windows", "", true},
		{"unknown os returns error", "plan9", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origGOOS := goos
			goos = tt.os
			defer func() { goos = origGOOS }()

			var gotProg, gotURL string
			origFn := openFn
			openFn = func(program, rawURL string) error {
				gotProg = program
				gotURL = rawURL
				return nil
			}
			defer func() { openFn = origFn }()

			err := Open("https://example.com")

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if !strings.Contains(err.Error(), "unsupported OS") {
					t.Errorf("error should mention 'unsupported OS', got: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotProg != tt.wantProg {
				t.Errorf("program = %q, want %q", gotProg, tt.wantProg)
			}
			if gotURL != "https://example.com" {
				t.Errorf("url = %q, want https://example.com", gotURL)
			}
		})
	}
}
