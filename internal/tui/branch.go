package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/chrisatdev/gf/internal/git"
)

type BranchModel struct {
	branches []string
	current  int
	filter   string
	selected string
	done     bool
}

func PickBranch(prompt string) (string, error) {
	runner := git.NewRunner()
	output, err := runner.Command("branch")
	if err != nil {
		return "", err
	}

	branches := strings.Split(output, "\n")
	var cleanBranches []string

	for _, b := range branches {
		b = strings.TrimSpace(b)
		if b == "" {
			continue
		}
		if strings.HasPrefix(b, "*") {
			cleanBranches = append(cleanBranches, strings.TrimSpace(b[1:]))
		} else {
			cleanBranches = append(cleanBranches, b)
		}
	}

	model := BranchModel{
		branches: cleanBranches,
		current:  0,
	}

	p := tea.NewProgram(model)
	result, err := p.Run()
	if err != nil {
		return "", err
	}

	return result.(BranchModel).selected, nil
}

func (m BranchModel) Init() tea.Cmd {
	return nil
}

func (m BranchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.done = true
			return m, tea.Quit

		case "up", "k":
			if m.current > 0 {
				m.current--
			}

		case "down", "j":
			if m.current < len(m.branches)-1 {
				m.current++
			}

		case "enter":
			if m.current < len(m.branches) {
				m.selected = m.branches[m.current]
				m.done = true
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m BranchModel) View() string {
	s := TitleStyle.Render("BRANCH SELECTOR")
	s += HelpStyle.Render("\n\n")

	for i, branch := range m.branches {
		marker := "○"
		if strings.HasPrefix(branch, "* ") || (i == 0 && m.current == 0) {
			marker = "●"
		}

		if i == m.current {
			s += SelectedStyle.Render(fmt.Sprintf(" %s %s\n", marker, branch))
		} else {
			s += fmt.Sprintf(" %s %s\n", marker, FileStyle.Render(branch))
		}
	}

	s += HelpStyle.Render("\n[up/down] Navigate  [Enter] Select  [Q] Quit\n")
	return s
}
