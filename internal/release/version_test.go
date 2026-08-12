package release

import "testing"

func TestBumpVersion(t *testing.T) {
	tests := []struct {
		name    string
		current string
		bump    string
		want    string
		wantErr bool
	}{
		{"patch bump", "v1.2.3", "patch", "v1.2.4", false},
		{"minor bump resets patch", "v1.2.3", "minor", "v1.3.0", false},
		{"major bump resets minor and patch", "v1.2.3", "major", "v2.0.0", false},
		{"no v prefix accepted", "1.2.3", "patch", "v1.2.4", false},
		{"zero to one", "v0.0.0", "patch", "v0.0.1", false},
		{"unknown bump type", "v1.2.3", "invalid", "", true},
		{"non-numeric version", "not-a-version", "patch", "", true},
		{"missing patch segment", "v1.2", "patch", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BumpVersion(tt.current, tt.bump)
			if (err != nil) != tt.wantErr {
				t.Fatalf("BumpVersion(%q, %q) error = %v, wantErr %v", tt.current, tt.bump, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("BumpVersion(%q, %q) = %q, want %q", tt.current, tt.bump, got, tt.want)
			}
		})
	}
}
