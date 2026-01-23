package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/phravins/devcli/internal/config"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type item struct {
	id, title, desc string // Added id field
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type DashboardModel struct {
	list         list.Model
	settings     SettingsModel
	choice       string
	quitting     bool
	showCommands bool
	showSettings bool
	width        int
	height       int
	commandView  viewport.Model
}

func NewDashboard() DashboardModel {
	items := []list.Item{
		item{title: "Project Tools", desc: "Create projects, sync, clone, scan"},
		item{title: "AI Chat", desc: "Chat with AI models"},
		item{title: "Editor", desc: "Built-in code editor"},
		item{title: "File Manager", desc: "Explore, Search, and Manage Files (RW/Move)"},
		item{title: "Settings / Configuration", desc: "Configure AI backends and Keys"},
		item{title: "DevCLI Commands", desc: "List all available project commands"},
		item{title: "Auto-Update", desc: "Update Languages, AI Keys, and DevCLI"},
		item{title: "Exit", desc: "Quit DevCLI"},
	}

	m := DashboardModel{
		list:     list.New(items, list.NewDefaultDelegate(), 0, 0),
		settings: NewSettingsModel(),
	}
	m.list.SetShowTitle(false)

	// Initialize viewport
	m.commandView = viewport.New(0, 0)
	m.commandView.SetContent(generateCommandsHelp())

	return m
}

func (m DashboardModel) Init() tea.Cmd {
	return nil
}

func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.showCommands {
			if msg.String() == "esc" || msg.String() == "q" {
				m.showCommands = false
				return m, nil
			}

			// Handle viewport scrolling
			var cmd tea.Cmd
			m.commandView, cmd = m.commandView.Update(msg)
			return m, cmd
		}

		if m.showSettings {
			var cmd tea.Cmd
			// We need to cast back to SettingsModel to access internal fields if needed,
			// but Update returns tea.Model.
			// Check if quitting from settings
			updatedModel, cmd := m.settings.Update(msg)
			m.settings = updatedModel.(SettingsModel)

			if m.settings.quitting {
				m.showSettings = false
				m.settings.quitting = false // Reset for next time
				return m, nil
			}
			return m, cmd
		}

		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				if i.title == "DevCLI Commands" {
					m.showCommands = true
					// Ensure content is fresh and viewport resized if necessary
					m.commandView.SetContent(generateCommandsHelp())
					m.commandView.GotoTop()
					return m, nil
				}
				if i.title == "Settings / Configuration" {
					m.showSettings = true
					// Re-init settings to read fresh config?
					m.settings = NewSettingsModel()
					// Immediately resize to current dimensions
					if m.width > 0 && m.height > 0 {
						updatedSettings, _ := m.settings.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
						m.settings = updatedSettings.(SettingsModel)
					}
					return m, m.settings.inputs[0].Focus()
				}
				if i.title == "File Manager" {
					m.choice = i.title
					return m, func() tea.Msg { return SwitchViewMsg{TargetState: StateFileManager} }
				}
				if i.title == "Project Tools" {
					m.choice = i.title
					return m, func() tea.Msg { return SwitchViewMsg{TargetState: StateProject} }
				}
				if i.title == "AI Chat" {
					m.choice = i.title
					return m, func() tea.Msg { return SwitchViewMsg{TargetState: StateChat} }
				}
				if i.title == "Editor" {
					m.choice = i.title
					return m, func() tea.Msg { return SwitchViewMsg{TargetState: StateEditor} }
				}
				if i.title == "Auto-Update" {
					m.choice = i.title
					return m, func() tea.Msg { return SwitchViewMsg{TargetState: StateAutoUpdate} }
				}

				m.choice = i.title
				return m, tea.Quit // Exit for "Exit" option or unknown
			}
		}
	case tea.MouseMsg:
		if m.showCommands {
			var cmd tea.Cmd
			m.commandView, cmd = m.commandView.Update(msg)
			return m, cmd
		}
		if m.showSettings {
			var cmd tea.Cmd
			// Cast to tea.Model because m.settings.Update returns tea.Model
			updatedModel, cmd := m.settings.Update(msg)
			m.settings = updatedModel.(SettingsModel)
			return m, cmd
		}

		// Handle Main List Scrolling
		if msg.Type == tea.MouseWheelUp {
			m.list.CursorUp()
			return m, nil
		}
		if msg.Type == tea.MouseWheelDown {
			m.list.CursorDown()
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v-16)

		// Resize Settings
		if m.showSettings {
			updatedSettings, _ := m.settings.Update(msg)
			m.settings = updatedSettings.(SettingsModel)
		}

		// Resize Command Viewport
		// Calculate available space similar to View() logic
		availableHeight := m.height - 2 - lipgloss.Height(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("Crafted & Engineered by phravins")) - 2 // Approximate header/footer
		if availableHeight < 0 {
			availableHeight = 0
		}
		m.commandView.Width = msg.Width - 4
		m.commandView.Height = availableHeight
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m DashboardModel) View() string {
	if m.quitting {
		return "Bye!"
	}

	if m.showSettings {
		return docStyle.Render(m.settings.View())
	}

	headerStyle := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center)

	logo := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#0F9E99")). // Tropical Teal
		Bold(true).
		Render(`
  ____  _______     __   ____ _     ___ 
 |  _ \| ____\ \   / /  / ___| |   |_ _|
 | | | |  _|  \ \ / /  | |   | |    | | 
 | |_| | |___  \ V /   | |___| |___ | | 
 |____/|_____|  \_/     \____|_____|___|`)

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#EFE9E0")). // Soft Ivory
		Render("Developer's CLI")              // Removed Padding(0,1) to save logical height

	footer := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("#666666")). // Grey for "smaller" feel
		Render("OPENDEV TOOLKIT")

	version := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Italic(true).
		Render(config.Version)

	centeredHeader := headerStyle.Render(logo + "\n" + title + "\n" + version)

	// --- COMMANDS VIEW ---
	if m.showCommands {
		return docStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
			m.commandView.View(),
			strings.Repeat("\n", 1), // Small gap
			footer,
		))
	}

	// --- DASHBOARD LIST VIEW ---

	// Create the main content stack (Header + List) first to measure its height
	contentView := lipgloss.JoinVertical(lipgloss.Left,
		centeredHeader,
		"\n",
		m.list.View(),
	)

	// Calculate how much space we have left for the footer to be at the absolute bottom
	// -2 for margings (docStyle padding)
	availableHeight := m.height - 2
	contentHeight := lipgloss.Height(contentView)
	footerHeight := lipgloss.Height(footer)

	gapHeight := availableHeight - contentHeight - footerHeight
	if gapHeight < 0 {
		gapHeight = 0
	}

	spacer := strings.Repeat("\n", gapHeight)

	// Combine: Content + Spacer + Footer
	return docStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
		contentView,
		spacer,
		footer,
	))
}

func RunDashboard() string {
	m := NewDashboard()
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Println("Error running dashboard:", err)
		os.Exit(1)
	}

	if dashModel, ok := finalModel.(DashboardModel); ok {
		return dashModel.choice
	}
	return ""
}
