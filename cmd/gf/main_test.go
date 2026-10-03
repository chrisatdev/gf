package main

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestVersionFlagUsesLinkerInjectedVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("builds and executes the CLI")
	}

	binary := filepath.Join(t.TempDir(), "gf")
	build := exec.Command("go", "build", "-ldflags=-X=main.version=v0.2.0", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build CLI with version injection: %v\n%s", err, output)
	}

	output, err := exec.Command(binary, "-v").CombinedOutput()
	if err != nil {
		t.Fatalf("run CLI version flag: %v\n%s", err, output)
	}
	if got, want := strings.TrimSpace(string(output)), "gf version v0.2.0"; got != want {
		t.Errorf("version output = %q, want %q", got, want)
	}
}
