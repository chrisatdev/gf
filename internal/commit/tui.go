package commit

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/chrisatdev/gf/internal/gitexec"
)

// gitRun is injectable for tests.
var gitRun = gitexec.Run

type step int

const (
	stepType        step = iota
	stepScope
	stepBreaking
	stepDescription
	stepPreview
	stepDone
)

type model struct {
	step      step
	types     []string
	cursor    int
	scope     string
	breaking  bool
	desc      string
	err       string
	confirmed bool
	cancelled bool
	input     textinput.Model
}

func newModel() model {
	ti := textinput.New()
	ti.CharLimit = 72
	return model{
		types: []string{"feat", "fix", "chore", "docs", "refactor", "test", "ci", "build"},
		input: ti,
	}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok && key.Type == tea.KeyCtrlC {
		m.cancelled = true
		m.step = stepDone
		return m, tea.Quit
	}
	switch m.step {
	case stepType:
		return m.updateType(msg)
	case stepScope:
		return m.updateScope(msg)
	case stepBreaking:
		return m.updateBreaking(msg)
	case stepDescription:
		return m.updateDescription(msg)
	case stepPreview:
		return m.updatePreview(msg)
	}
	return m, nil
}

func (m model) updateType(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.Type {
	case tea.KeyUp:
		if m.cursor > 0 {
			m.cursor--
		}
	case tea.KeyDown:
		if m.cursor < len(m.types)-1 {
			m.cursor++
		}
	case tea.KeyEnter:
		m.step = stepScope
		m.input.Reset()
		m.input.Placeholder = "optional – press Enter to skip"
		cmd := m.input.Focus()
		return m, cmd
	}
	return m, nil
}

func (m model) updateScope(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	if key.Type == tea.KeyEnter {
		m.scope = m.input.Value()
		m.step = stepBreaking
		m.input.Blur()
		return m, nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m model) updateBreaking(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch {
	case key.Type == tea.KeyEnter:
		// default N — no breaking change
	case key.Type == tea.KeyRunes && (string(key.Runes) == "y" || string(key.Runes) == "Y"):
		m.breaking = true
	case key.Type == tea.KeyRunes && (string(key.Runes) == "n" || string(key.Runes) == "N"):
		// keep false
	default:
		return m, nil
	}
	m.step = stepDescription
	m.input.Reset()
	m.input.Placeholder = "description (required, min 3 chars)"
	cmd := m.input.Focus()
	return m, cmd
}

func (m model) updateDescription(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	if key.Type == tea.KeyEnter {
		val := m.input.Value()
		if len(val) < 3 {
			m.err = "description must be at least 3 characters"
			return m, nil
		}
		m.desc = val
		m.err = ""
		m.step = stepPreview
		m.input.Blur()
		return m, nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m model) updatePreview(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch {
	case key.Type == tea.KeyEnter ||
		(key.Type == tea.KeyRunes && (string(key.Runes) == "y" || string(key.Runes) == "Y")):
		m.confirmed = true
		m.step = stepDone
		return m, tea.Quit
	case key.Type == tea.KeyRunes && (string(key.Runes) == "n" || string(key.Runes) == "N"):
		m.cancelled = true
		m.step = stepDone
		return m, tea.Quit
	}
	return m, nil
}

func (m model) View() string {
	switch m.step {
	case stepType:
		return m.viewType()
	case stepScope:
		return "Enter scope (optional, press Enter to skip):\n> " + m.input.View() + "\n"
	case stepBreaking:
		return "Breaking change? [y/N]: "
	case stepDescription:
		s := "Enter description (required, min 3 chars):\n> " + m.input.View() + "\n"
		if m.err != "" {
			s += "  " + m.err + "\n"
		}
		return s
	case stepPreview:
		return "Preview: " + m.buildTitle() + "\nConfirm commit? [Y/n]: "
	case stepDone:
		return ""
	}
	return ""
}

func (m model) viewType() string {
	s := "Select commit type (↑/↓ to move, Enter to select):\n"
	for i, t := range m.types {
		prefix := "  "
		if i == m.cursor {
			prefix = "> "
		}
		s += prefix + t + "\n"
	}
	return s
}

// buildTitle formats the conventional commit title with optional breaking marker.
func (m model) buildTitle() string {
	breaking := ""
	if m.breaking {
		breaking = "!"
	}
	t := m.types[m.cursor]
	if m.scope != "" {
		return fmt.Sprintf("%s(%s)%s: %s", t, m.scope, breaking, m.desc)
	}
	return fmt.Sprintf("%s%s: %s", t, breaking, m.desc)
}

// commitIfConfirmed handles post-TUI commit logic. Extracted for testability.
func commitIfConfirmed(m model) error {
	if m.cancelled {
		fmt.Println("Commit cancelled.")
		return nil
	}
	if m.confirmed {
		title := m.buildTitle()
		if err := gitRun("commit", "-m", title); err != nil {
			return fmt.Errorf("gf commit: %w", err)
		}
	}
	return nil
}

// RunTUI launches the interactive conventional commit wizard.
// It commits only if the user confirms at the preview step.
func RunTUI() error {
	p := tea.NewProgram(newModel())
	fm, err := p.Run()
	if err != nil {
		return fmt.Errorf("gf commit: %w", err)
	}
	m, ok := fm.(model)
	if !ok {
		return fmt.Errorf("gf commit: unexpected model type")
	}
	return commitIfConfirmed(m)
}
