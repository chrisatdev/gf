package tui

import "github.com/charmbracelet/lipgloss"

var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7C3AED")).
			Padding(0, 1)

	WarningStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#F59E0B"))

	SuccessStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#10B981"))

	ErrorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#EF4444"))

	FileStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E5E7EB"))

	SelectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#374151")).
			Foreground(lipgloss.Color("#FAFAFA"))

	PanelStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#6B7280")).
			Padding(1)

	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF"))
)
