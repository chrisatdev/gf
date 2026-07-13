package commit

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// --- helpers ---

// advance drives m through a single Update and returns the typed result.
func advance(m model, msg tea.Msg) model {
	next, _ := m.Update(msg)
	return next.(model)
}

func keyEnter() tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyEnter} }
func keyDown() tea.KeyMsg  { return tea.KeyMsg{Type: tea.KeyDown} }
func keyRune(r rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
}

// driveToScope advances a fresh model past stepType.
func driveToScope(m model) model { return advance(m, keyEnter()) }

// driveToBreaking advances past stepType and stepScope (empty scope).
func driveToBreaking(m model) model { return advance(driveToScope(m), keyEnter()) }

// driveToDescription advances past stepType, stepScope, and stepBreaking (N).
func driveToDescription(m model) model { return advance(driveToBreaking(m), keyRune('n')) }

// driveToPreview advances to stepPreview with desc already set.
func driveToPreview(m model, desc string) model {
	m = driveToDescription(m)
	m.input.SetValue(desc)
	return advance(m, keyEnter())
}

// --- TestModel_TypeSelectionAdvances ---

func TestModel_TypeSelectionAdvances(t *testing.T) {
	m := newModel()
	// initial state: stepType, cursor at feat (index 0)
	m = advance(m, keyEnter())

	if m.step != stepScope {
		t.Errorf("step: want stepScope(%d), got %d", stepScope, m.step)
	}
	if m.cursor != 0 {
		t.Errorf("cursor: want 0 (feat), got %d", m.cursor)
	}
}

func TestModel_TypeSelectionMovesWithArrows(t *testing.T) {
	m := newModel()
	m = advance(m, keyDown()) // feat → fix
	m = advance(m, keyDown()) // fix  → chore

	if m.cursor != 2 {
		t.Errorf("cursor after 2 down: want 2 (chore), got %d", m.cursor)
	}

	m = advance(m, keyEnter())
	if m.step != stepScope {
		t.Errorf("step: want stepScope, got %d", m.step)
	}
	if m.cursor != 2 {
		t.Errorf("cursor preserved: want 2, got %d", m.cursor)
	}
}

// --- TestModel_ScopeCanBeSkipped ---

func TestModel_ScopeCanBeSkipped(t *testing.T) {
	m := newModel()
	m = driveToScope(m)
	m = advance(m, keyEnter()) // skip — empty Enter

	if m.step != stepBreaking {
		t.Errorf("step: want stepBreaking(%d), got %d", stepBreaking, m.step)
	}
	if m.scope != "" {
		t.Errorf("scope: want empty, got %q", m.scope)
	}
}

// --- TestModel_BreakingSelection ---

func TestModel_BreakingSelection(t *testing.T) {
	cases := []struct {
		name     string
		key      tea.KeyMsg
		wantBrk  bool
	}{
		{"y sets breaking", keyRune('y'), true},
		{"Y sets breaking", keyRune('Y'), true},
		{"n keeps false", keyRune('n'), false},
		{"N keeps false", keyRune('N'), false},
		{"enter defaults to N", keyEnter(), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := newModel()
			m = driveToBreaking(m)
			m = advance(m, tc.key)

			if m.step != stepDescription {
				t.Errorf("step: want stepDescription(%d), got %d", stepDescription, m.step)
			}
			if m.breaking != tc.wantBrk {
				t.Errorf("breaking: want %v, got %v", tc.wantBrk, m.breaking)
			}
		})
	}
}

// --- TestModel_DescriptionValidation ---

func TestModel_DescriptionValidation(t *testing.T) {
	m := newModel()
	m = driveToDescription(m)

	// too short — must stay in stepDescription with an error
	m.input.SetValue("ab")
	m = advance(m, keyEnter())

	if m.step != stepDescription {
		t.Errorf("short desc: step: want stepDescription, got %d", m.step)
	}
	if m.err == "" {
		t.Error("short desc: want non-empty error message")
	}

	// valid description — must advance to stepPreview and clear error
	m.input.SetValue("add wizard")
	m = advance(m, keyEnter())

	if m.step != stepPreview {
		t.Errorf("valid desc: step: want stepPreview, got %d", m.step)
	}
	if m.desc != "add wizard" {
		t.Errorf("valid desc: desc: want 'add wizard', got %q", m.desc)
	}
	if m.err != "" {
		t.Errorf("valid desc: want no error, got %q", m.err)
	}
}

// --- TestModel_PreviewConfirmBuildsCommit ---

func TestModel_PreviewConfirmBuildsCommit(t *testing.T) {
	// feat, no scope, no breaking → "feat: add wizard"
	m := newModel()
	m = driveToPreview(m, "add wizard")
	m = advance(m, keyRune('y'))

	if m.step != stepDone {
		t.Errorf("step: want stepDone, got %d", m.step)
	}
	if !m.confirmed {
		t.Error("want confirmed=true")
	}
	want := "feat: add wizard"
	if got := m.buildTitle(); got != want {
		t.Errorf("title: want %q, got %q", want, got)
	}
}

func TestModel_PreviewConfirmEnterKey(t *testing.T) {
	m := newModel()
	m = driveToPreview(m, "add wizard")
	m = advance(m, keyEnter()) // Enter also confirms (default Y)

	if !m.confirmed {
		t.Error("Enter key: want confirmed=true")
	}
	if m.step != stepDone {
		t.Errorf("Enter key: want stepDone, got %d", m.step)
	}
}

// --- TestModel_PreviewCancel ---

func TestModel_PreviewCancel(t *testing.T) {
	m := newModel()
	m = driveToPreview(m, "add wizard")
	m = advance(m, keyRune('n'))

	if m.step != stepDone {
		t.Errorf("step: want stepDone, got %d", m.step)
	}
	if !m.cancelled {
		t.Error("want cancelled=true")
	}
	if m.confirmed {
		t.Error("want confirmed=false")
	}
}

// --- TestRunTUI_CommitCalledOnConfirm ---

func TestRunTUI_CommitCalledOnConfirm(t *testing.T) {
	var gotArgs []string
	orig := gitRun
	gitRun = func(args ...string) error {
		gotArgs = append(gotArgs, args...)
		return nil
	}
	defer func() { gitRun = orig }()

	m := model{
		types:     []string{"feat", "fix", "chore", "docs", "refactor", "test", "ci", "build"},
		cursor:    0,
		desc:      "add wizard",
		confirmed: true,
		step:      stepDone,
	}
	if err := commitIfConfirmed(m); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(gotArgs) < 3 {
		t.Fatalf("gitRun: want >=3 args, got %v", gotArgs)
	}
	if gotArgs[0] != "commit" || gotArgs[1] != "-m" {
		t.Errorf("gitRun args[0:2]: want [commit -m], got %v", gotArgs[:2])
	}
	wantTitle := "feat: add wizard"
	if gotArgs[2] != wantTitle {
		t.Errorf("commit title: want %q, got %q", wantTitle, gotArgs[2])
	}
}

// --- TestRunTUI_NoCommitOnCancel ---

func TestRunTUI_NoCommitOnCancel(t *testing.T) {
	called := false
	orig := gitRun
	gitRun = func(_ ...string) error {
		called = true
		return nil
	}
	defer func() { gitRun = orig }()

	m := model{
		types:     []string{"feat"},
		cancelled: true,
		step:      stepDone,
	}
	if err := commitIfConfirmed(m); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("gitRun must not be called when cancelled")
	}
}

// --- breaking change title format ---

func TestModel_BreakingTitleFormat(t *testing.T) {
	// feat(api)!: remove v1 endpoint
	m := newModel()
	m.cursor = 0 // feat
	m.scope = "api"
	m.breaking = true
	m.desc = "remove v1 endpoint"

	want := "feat(api)!: remove v1 endpoint"
	if got := m.buildTitle(); got != want {
		t.Errorf("breaking with scope: want %q, got %q", want, got)
	}

	// feat!: breaking without scope
	m2 := newModel()
	m2.cursor = 0
	m2.breaking = true
	m2.desc = "breaking change"

	want2 := "feat!: breaking change"
	if got := m2.buildTitle(); got != want2 {
		t.Errorf("breaking no scope: want %q, got %q", want2, got)
	}
}
