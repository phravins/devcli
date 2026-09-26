package tui

import "github.com/charmbracelet/lipgloss"

var (
	colorPurple	= lipgloss.Color("#BD93F9")
	colorCyan	= lipgloss.Color("#8BE9FD")
	colorGreen	= lipgloss.Color("#50FA7B")
	colorRed	= lipgloss.Color("#FF5555")
	colorPink	= lipgloss.Color("#FF79C6")

	colorGray	= lipgloss.Color("#6272A4")
	colorYellow	= lipgloss.Color("#F1FA8C")
)

var (
	docStyle	= lipgloss.NewStyle().Margin(0, 0)

	AppBorderStyle	= lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPurple).
			Padding(1, 2)

	LeftPaneStyle	= lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPurple).
			Padding(1, 1)

	RightPaneStyle	= lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorCyan).
			Padding(1, 2)

	HeaderBadgeStyle	= lipgloss.NewStyle().
				Background(colorGreen).
				Foreground(lipgloss.Color("#282a36")).
				Bold(true).
				Padding(0, 1)

	titleStyle	= lipgloss.NewStyle().
			Foreground(colorPurple).
			Bold(true).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPurple)

	inputBoxStyle	= lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorGray).
			Padding(1, 3).
			Align(lipgloss.Center)

	focusedInputBoxStyle	= inputBoxStyle.BorderForeground(colorPurple)

	successBoxStyle	= lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorGreen).
			Padding(1, 4).
			Align(lipgloss.Center)

	errorBoxStyle	= lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorRed).
			Padding(1, 4).
			Align(lipgloss.Center)

	subtleStyle	= lipgloss.NewStyle().Foreground(colorGray)

	loadingStyle	= lipgloss.NewStyle().
			Foreground(colorYellow).
			Bold(true).
			Align(lipgloss.Center)

	errorStyle	= lipgloss.NewStyle().
			Foreground(colorRed).
			Bold(true)

	venvTitleStyle	= lipgloss.NewStyle().
			Foreground(colorCyan).
			Bold(true).
			Padding(0, 1).
			Border(lipgloss.DoubleBorder(), false, false, true, false).
			BorderForeground(colorPurple)

	venvCardStyle	= lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPurple).
			Padding(1, 2).
			Width(50).
			Align(lipgloss.Center)

	venvSelectedStyle	= lipgloss.NewStyle().
				Foreground(colorGreen).
				Bold(true).
				PaddingLeft(1)

	WizardCardStyle	= lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPurple).
			Padding(1, 2).
			Width(65).
			Align(lipgloss.Center)

	StepStyle	= lipgloss.NewStyle().
			Foreground(colorPink).
			Bold(true).
			MarginBottom(1)

	PreviewHeaderStyle	= lipgloss.NewStyle().
				Background(colorCyan).
				Foreground(lipgloss.Color("#282a36")).
				Bold(true).
				Padding(0, 2)
)
