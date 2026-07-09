package ui

import "github.com/charmbracelet/lipgloss"

var (
	baseStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Background(lipgloss.Color("235"))

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("255"))

	mutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Background(lipgloss.Color("238")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("203")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("120")).
			Bold(true)

	warnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("222")).
			Bold(true)

	activeBorder = lipgloss.Color("252")
	quietBorder  = lipgloss.Color("240")
)

func panelStyle(width, height int, focused bool) lipgloss.Style {
	border := quietBorder
	if focused {
		border = activeBorder
	}
	return baseStyle.
		Width(width).
		Height(height).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(border).
		Padding(0, 1)
}
