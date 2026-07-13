package resolve

import (
	"strings"
	"testing"
)

// --- mock helpers ---

// overrideLines replaces linesFn with a fixed file list.
func overrideLines(t *testing.T, files []string) {
	t.Helper()
	orig := linesFn
	linesFn = func(_ ...string) ([]string, error) { return files, nil }
	t.Cleanup(func() { linesFn = orig })
}

// overrideRun replaces runFn with a no-op that records all calls.
func overrideRun(t *testing.T) *[][]string {
	t.Helper()
	var calls [][]string
	orig := runFn
	runFn = func(args ...string) error {
		calls = append(calls, append([]string{}, args...))
		return nil
	}
	t.Cleanup(func() { runFn = orig })
	return &calls
}

// overrideMergeCL replaces mergeCLFn and counts invocations.
func overrideMergeCL(t *testing.T, retErr error) *int {
	t.Helper()
	count := 0
	orig := mergeCLFn
	mergeCLFn = func(_ string) error {
		count++
		return retErr
	}
	t.Cleanup(func() { mergeCLFn = orig })
	return &count
}

// overrideEditor replaces editorRunFn and records files it was asked to open.
func overrideEditor(t *testing.T) *[]string {
	t.Helper()
	var opened []string
	orig := editorRunFn
	editorRunFn = func(file string) error {
		opened = append(opened, file)
		return nil
	}
	t.Cleanup(func() { editorRunFn = orig })
	return &opened
}

// --- tests ---

func TestExecute_NoConflicts(t *testing.T) {
	overrideLines(t, nil)

	if err := Execute(strings.NewReader("")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecute_ChangelogAutoResolve(t *testing.T) {
	overrideLines(t, []string{"CHANGELOG.md"})
	clCount := overrideMergeCL(t, nil)
	gitCalls := overrideRun(t)

	if err := Execute(strings.NewReader("")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if *clCount != 1 {
		t.Errorf("changelog.MergeConflict called %d times, want 1", *clCount)
	}

	var sawAdd bool
	for _, c := range *gitCalls {
		if len(c) == 2 && c[0] == "add" && c[1] == "CHANGELOG.md" {
			sawAdd = true
		}
	}
	if !sawAdd {
		t.Error("expected 'git add CHANGELOG.md'")
	}
}

func TestExecute_Choices(t *testing.T) {
	cases := []struct {
		name         string
		input        string
		wantCheckout string // "--ours" or "--theirs" or ""
		wantAdd      bool
	}{
		{"ours", "o\n", "--ours", true},
		{"theirs", "t\n", "--theirs", true},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			overrideLines(t, []string{"conflict.go"})
			gitCalls := overrideRun(t)

			if err := Execute(strings.NewReader(tt.input)); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var sawCheckout, sawAdd bool
			for _, c := range *gitCalls {
				if len(c) == 3 && c[0] == "checkout" && c[1] == tt.wantCheckout && c[2] == "conflict.go" {
					sawCheckout = true
				}
				if len(c) == 2 && c[0] == "add" && c[1] == "conflict.go" {
					sawAdd = true
				}
			}
			if !sawCheckout {
				t.Errorf("expected 'git checkout %s conflict.go'", tt.wantCheckout)
			}
			if tt.wantAdd && !sawAdd {
				t.Error("expected 'git add conflict.go'")
			}
		})
	}
}

func TestExecute_SkipLeavesUnresolved(t *testing.T) {
	overrideLines(t, []string{"conflict.go"})
	gitCalls := overrideRun(t)

	if err := Execute(strings.NewReader("s\n")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// No checkout or add should have happened for the skipped file.
	for _, c := range *gitCalls {
		if len(c) > 0 && c[0] == "checkout" {
			t.Errorf("checkout should not be called for skipped file, got: %v", c)
		}
	}
}

func TestExecute_EditChoiceUsesInjectedEditor(t *testing.T) {
	overrideLines(t, []string{"conflict.go"})
	opened := overrideEditor(t)
	gitCalls := overrideRun(t)

	if err := Execute(strings.NewReader("e\n")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(*opened) != 1 || (*opened)[0] != "conflict.go" {
		t.Errorf("editor opened %v, want [conflict.go]", *opened)
	}

	var sawAdd bool
	for _, c := range *gitCalls {
		if len(c) == 2 && c[0] == "add" && c[1] == "conflict.go" {
			sawAdd = true
		}
	}
	if !sawAdd {
		t.Error("expected 'git add conflict.go' after edit")
	}
}
