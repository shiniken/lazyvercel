package ui

import "github.com/charmbracelet/lipgloss"

var (
	baseStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	chromeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("255"))

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("248"))

	mutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Bold(true)

	selectedMutedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("250"))

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
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

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("153")).
			Bold(true)

	sectionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Bold(true)

	activeBorder = lipgloss.Color("105")
	quietBorder  = lipgloss.Color("238")
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
		Padding(0, 1).
		MarginRight(1)
}

func badgeStyle(state string, selected bool) lipgloss.Style {
	foreground := lipgloss.Color("252")
	switch state {
	case "READY":
		foreground = lipgloss.Color("120")
	case "ERROR", "CANCELED":
		foreground = lipgloss.Color("203")
	case "BUILDING", "INITIALIZING", "QUEUED":
		foreground = lipgloss.Color("222")
	default:
		foreground = lipgloss.Color("248")
	}
	return lipgloss.NewStyle().
		Foreground(foreground).
		Bold(selected)
}
