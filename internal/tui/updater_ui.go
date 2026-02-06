package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/phravins/devcli/internal/updater"
)

// UpdaterModel represents the update checker UI state
type UpdaterModel struct {
	width   int
	height  int
	info    *updater.UpdateInfo
	err     error
	status  string
	updated bool
}

// UpdateCheckMsg contains the result of checking for updates
type UpdateCheckMsg struct {
	info *updater.UpdateInfo
	err  error
}

// UpdateCompleteMsg indicates the update completed
type UpdateCompleteMsg struct {
	err error
}

// NewUpdaterModel creates a new updater model
func NewUpdaterModel() UpdaterModel {
	return UpdaterModel{
		status: "Checking for updates...",
	}
}

// Init initializes the updater model
func (m UpdaterModel) Init() tea.Cmd {
	return checkForUpdatesCmd
}

// checkForUpdatesCmd checks for updates
func checkForUpdatesCmd() tea.Msg {
	info, err := updater.CheckForUpdates()
	return UpdateCheckMsg{info: info, err: err}
}

// performUpdateCmd performs the update
func performUpdateCmd() tea.Msg {
	err := updater.PerformUpdate()
	return UpdateCompleteMsg{err: err}
}

// Update handles messages
func (m UpdaterModel) Update(msg tea.Msg) (UpdaterModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			return m, func() tea.Msg { return SubFeatureBackMsg{} }
		case "ctrl+c":
			return m, tea.Quit
		case "u":
			// User pressed 'u' to update
			if m.info != nil && m.info.IsUpdateAvailable && !m.updated {
				m.status = "Downloading and installing update..."
				return m, performUpdateCmd
			}
		}

	case UpdateCheckMsg:
		m.info = msg.info
		m.err = msg.err
		if msg.err != nil {
			m.status = fmt.Sprintf("Error: %v", msg.err)
		} else if msg.info.IsUpdateAvailable {
			m.status = "Update available! Press 'u' to update"
		} else {
			m.status = "You are up to date!"
		}
		return m, nil

	case UpdateCompleteMsg:
		if msg.err != nil {
			m.status = fmt.Sprintf("Update failed: %v", msg.err)
		} else {
			m.status = "Update successful! Please restart DevCLI."
			m.updated = true
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

	return m, nil
}

// View renders the updater UI
func (m UpdaterModel) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FF6B6B")).
		Padding(0, 1)

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4ECDC4")).
		Padding(1, 2)

	versionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFD700")).
		Padding(0, 2)

	notesStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#AAAAAA")).
		Padding(1, 2)

	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Padding(1, 0)

	var content string

	title := titleStyle.Render("🔄 DevCLI Update Checker")
	content = title + "\n\n"

	// Show status
	content += statusStyle.Render(m.status) + "\n\n"

	// Show version info if available
	if m.info != nil {
		content += versionStyle.Render(fmt.Sprintf("Current Version: %s", m.info.CurrentVersion)) + "\n"
		content += versionStyle.Render(fmt.Sprintf("Latest Version:  %s", m.info.LatestVersion)) + "\n\n"

		if m.info.IsUpdateAvailable && m.info.ReleaseNotes != "" {
			content += lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#6BCF7F")).
				Render("📝 Release Notes:") + "\n"
			content += notesStyle.Render(m.info.ReleaseNotes) + "\n"
		}
	}

	// Footer with instructions
	var footer string
	if m.info != nil && m.info.IsUpdateAvailable && !m.updated {
		footer = "U: Update • Q/Esc: Back • Ctrl+C: Quit"
	} else {
		footer = "Q/Esc: Back • Ctrl+C: Quit"
	}

	content += footerStyle.Render(footer)

	// Center content
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#4ECDC4")).
		Padding(2, 4).
		Width(m.width - 4).
		Render(content)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}
