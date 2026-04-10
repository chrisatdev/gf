package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbletea"
	"github.com/chrisatdev/gf/internal/git"
)

type Model struct {
	conflicts []string
	current   int
	screen    int
	resolved  map[string]bool
}

const (
	ScreenList = iota
	ScreenDiff
	ScreenDone
)

func ResolveConflicts() error {
	runner := git.NewRunner()
	conflicts, err := runner.GetConflictingFiles()
	if err != nil || len(conflicts) == 0 {
		if err != nil {
			return err
		}
		return fmt.Errorf("no conflicts found")
	}

	model := Model{
		conflicts: conflicts,
		resolved:  make(map[string]bool),
	}

	p := tea.NewProgram(model)
	if err := p.Start(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	return nil
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "up", "k":
			if m.current > 0 {
				m.current--
			}

		case "down", "j":
			if m.current < len(m.conflicts)-1 {
				m.current++
			}

		case "enter":
			if m.screen == ScreenList {
				m.screen = ScreenDiff
			}

		case "o":
			if m.screen == ScreenDiff {
				file := m.conflicts[m.current]
				git.AcceptOurs(file)
				git.StageFile(file)
				m.resolved[file] = true
				m.screen = ScreenList
				if allResolved(m.conflicts, m.resolved) {
					m.screen = ScreenDone
				}
			}

		case "t":
			if m.screen == ScreenDiff {
				file := m.conflicts[m.current]
				git.AcceptTheirs(file)
				git.StageFile(file)
				m.resolved[file] = true
				m.screen = ScreenList
				if allResolved(m.conflicts, m.resolved) {
					m.screen = ScreenDone
				}
			}

		case "e":
			if m.screen == ScreenDiff {
				// Open mergetool
				file := m.conflicts[m.current]
				runMergetool(file)
				m.screen = ScreenList
			}

		case "a":
			if m.screen == ScreenList {
				// Accept all ours
				for _, file := range m.conflicts {
					git.AcceptOurs(file)
					git.StageFile(file)
					m.resolved[file] = true
				}
				m.screen = ScreenDone
			}

		case "esc":
			if m.screen > ScreenList {
				m.screen--
			}

		case " ":
			if m.screen == ScreenDone {
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	switch m.screen {
	case ScreenList:
		return renderConflictList(m)
	case ScreenDiff:
		return renderDiffView(m)
	case ScreenDone:
		return renderDone(m)
	default:
		return ""
	}
}

func renderConflictList(m Model) string {
	s := TitleStyle.Render("MERGE CONFLICTS DETECTED")
	s += fmt.Sprintf("\n\n%d files with conflicts:\n\n", len(m.conflicts))

	for i, file := range m.conflicts {
		marker := "○"
		status := "[  CONFLICTED  ]"

		if m.resolved[file] {
			marker = "●"
			status = "[  RESOLVED   ]"
		}

		if i == m.current {
			s += SelectedStyle.Render(fmt.Sprintf(" %s %s %s\n", marker, file, status))
		} else {
			s += fmt.Sprintf(" %s %s %s\n", marker, FileStyle.Render(file), HelpStyle.Render(status))
		}
	}

	s += HelpStyle.Render("\n[up/down] Navigate  [Enter] View  [A] Accept All  [Q] Quit\n")
	return s
}

func renderDiffView(m Model) string {
	file := m.conflicts[m.current]

	s := TitleStyle.Render("FILE: " + file)
	s += HelpStyle.Render("\n\n")

	// Simple diff view - read file content and show markers
	content2, _ := git.NewRunner().Command("show", ":2:"+file) // ours
	content3, _ := git.NewRunner().Command("show", ":3:"+file) // theirs

	s += PanelStyle.Render(fmt.Sprintf("OURS (HEAD)\n%s", content2))
	s += "\n\n"
	s += PanelStyle.Render(fmt.Sprintf("THEIRS (incoming)\n%s", content3))

	s += HelpStyle.Render("\n[O] Accept Ours  [T] Accept Theirs  [E] Edit  [Esc] Back\n")
	return s
}

func renderDone(m Model) string {
	s := SuccessStyle.Render("ALL CONFLICTS RESOLVED")
	s += HelpStyle.Render("\n\nRun `git commit` to complete the merge\n")
	s += HelpStyle.Render("\n[Space/Enter] Quit\n")
	return s
}

func allResolved(conflicts []string, resolved map[string]bool) bool {
	for _, f := range conflicts {
		if !resolved[f] {
			return false
		}
	}
	return true
}

func runMergetool(file string) {
	git.NewRunner().Command("mergetool", "--no-prompt", file)
}
