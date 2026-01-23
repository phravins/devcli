package tui

import "github.com/charmbracelet/lipgloss"

// Color Palette (Dracula-inspired)
var (
	colorPurple = lipgloss.Color("#BD93F9")
	colorCyan   = lipgloss.Color("#8BE9FD")
	colorGreen  = lipgloss.Color("#50FA7B")
	colorRed    = lipgloss.Color("#FF5555")
	colorPink   = lipgloss.Color("#FF79C6") // Dracula Pink

	colorGray   = lipgloss.Color("#6272A4")
	colorYellow = lipgloss.Color("#F1FA8C")
)

// Shared Styles
var (
	// Main container style - removed margin to let border handle it, or keep for spacing
	docStyle = lipgloss.NewStyle().Margin(0, 0) // Reset to 0, strict sizing manually

	// Global App Border
	AppBorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPurple).
			Padding(1, 2)

	// Titles
	titleStyle = lipgloss.NewStyle().
			Foreground(colorPurple).
			Bold(true).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPurple)

	// Input boxes
	inputBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorGray).
			Padding(1, 3).
			Align(lipgloss.Center)

	focusedInputBoxStyle = inputBoxStyle.Copy().
				BorderForeground(colorPurple)

	// Success/Error boxes
	successBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorGreen).
			Padding(1, 4).
			Align(lipgloss.Center)

	errorBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorRed).
			Padding(1, 4).
			Align(lipgloss.Center)

	// Helpers
	subtleStyle = lipgloss.NewStyle().Foreground(colorGray)

	loadingStyle = lipgloss.NewStyle().
			Foreground(colorYellow).
			Bold(true).
			Align(lipgloss.Center)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorRed).
			Bold(true)

	// Venv Wizard Styles
	venvTitleStyle = lipgloss.NewStyle().
			Foreground(colorCyan).
			Bold(true).
			Padding(0, 1).
			Border(lipgloss.DoubleBorder(), false, false, true, false).
			BorderForeground(colorPurple)

	venvCardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPurple).
			Padding(1, 2).
			Width(50).
			Align(lipgloss.Center)

	venvSelectedStyle = lipgloss.NewStyle().
				Foreground(colorGreen).
				Bold(true).
				PaddingLeft(1)

	// --- Smart File Premium Styles ---

	// Wizard Card for inputs
	WizardCardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPurple).
			Padding(1, 2).
			Width(65).
			Align(lipgloss.Center)

	// Step text like "Step 1/3"
	StepStyle = lipgloss.NewStyle().
			Foreground(colorPink).
			Bold(true).
			MarginBottom(1)

	// Preview Window Header
	PreviewHeaderStyle = lipgloss.NewStyle().
				Background(colorCyan).
				Foreground(lipgloss.Color("#282a36")). // Dark text
				Bold(true).
				Padding(0, 2)
)
