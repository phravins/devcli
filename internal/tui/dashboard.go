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
		item{title: "📂 Project Tools", desc: "Create projects, sync, clone, scan"},
		item{title: "🤖 AI Chat", desc: "Chat with AI models"},
		item{title: "✏️ Editor", desc: "Built-in code editor"},
		item{title: "🗂️ File Manager", desc: "Explore, Search, and Manage Files (RW/Move)"},
		item{title: "🐳 Docker Dashboard", desc: "Manage Containers, Inspect Logs, Start/Stop"},
		item{title: "🌐 API & HTTP Playground", desc: "Test REST API Endpoints with HTTP Client"},
		item{title: "⚙️ Settings / Configuration", desc: "Configure AI backends and Keys"},
		item{title: "💻 DevCLI Commands", desc: "List all available project commands"},
		item{title: "🔄 Auto-Update", desc: "Update Languages, AI Keys, and DevCLI"},
		item{title: "📚 Docs", desc: "Read DevCLI Documentation"},
		item{title: "🚪 Exit", desc: "Quit DevCLI"},
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.NormalTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("#F8F8F2")).Bold(true)
	delegate.Styles.NormalDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("#A0A0A0"))
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")).Bold(true)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("#8BE9FD"))

	m := DashboardModel{
		list:     list.New(items, delegate, 80, 20),
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
			var cmd tea.Cmd
			m.commandView, cmd = m.commandView.Update(msg)
			return m, cmd
		}

		if m.showSettings {
			var cmd tea.Cmd
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
				if i.title == "💻 DevCLI Commands" {
					m.showCommands = true
					m.commandView.SetContent(generateCommandsHelp())
					m.commandView.GotoTop()
					return m, nil
				}
				if i.title == "⚙️ Settings / Configuration" {
					m.showSettings = true
					// Re-init settings to read fresh config
					m.settings = NewSettingsModel()
					// Immediately resize to current dimensions
					if m.width > 0 && m.height > 0 {
						updatedSettings, _ := m.settings.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
						m.settings = updatedSettings.(SettingsModel)
					}
					return m, m.settings.inputs[0].Focus()
				}
				if i.title == "🗂️ File Manager" {
					m.choice = i.title
					return m, func() tea.Msg { return SwitchViewMsg{TargetState: StateFileManager} }
				}
				if i.title == "📂 Project Tools" {
					m.choice = i.title
					return m, func() tea.Msg { return SwitchViewMsg{TargetState: StateProject} }
				}
				if i.title == "🤖 AI Chat" {
					m.choice = i.title
					return m, func() tea.Msg { return SwitchViewMsg{TargetState: StateChat} }
				}
				if i.title == "✏️ Editor" {
					m.choice = i.title
					return m, func() tea.Msg { return SwitchViewMsg{TargetState: StateEditor} }
				}
				if i.title == "🔄 Auto-Update" {
					m.choice = i.title
					return m, func() tea.Msg { return SwitchViewMsg{TargetState: StateAutoUpdate} }
				}
				if i.title == "🐳 Docker Dashboard" {
					m.choice = i.title
					return m, func() tea.Msg { return SwitchViewMsg{TargetState: StateDocker} }
				}
				if i.title == "🌐 API & HTTP Playground" {
					m.choice = i.title
					return m, func() tea.Msg { return SwitchViewMsg{TargetState: StateAPIClient} }
				}
				if i.title == "📚 Docs" {
					m.choice = i.title
					return m, func() tea.Msg { return SwitchViewMsg{TargetState: StateDocs} }
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
			updatedModel, cmd := m.settings.Update(msg)
			m.settings = updatedModel.(SettingsModel)
			return m, cmd
		}
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
		listHeight := msg.Height - v - 12
		if listHeight < 10 {
			listHeight = 10
		}
		m.list.SetSize(msg.Width-h, listHeight)

		// Resize Settings
		if m.showSettings {
			updatedSettings, _ := m.settings.Update(msg)
			m.settings = updatedSettings.(SettingsModel)
		}
		availableHeight := m.height - 8
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

	// 1. Render Top Live Status Bar
	topBar := RenderStatusBar(m.width)

	// 2. Render ASCII Logo & Header Banner
	logoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#8BE9FD")). // Glowing Cyan
		Bold(true)

	logo := logoStyle.Render(`
  ____  _______     __   ____ _     ___ 
 |  _ \| ____\ \   / /  / ___| |   |_ _|
 | | | |  _|  \ \ / /  | |   | |    | | 
 | |_| | |___  \ V /   | |___| |___ | | 
 |____/|_____|  \_/     \____|_____|___|`)

	versionBadge := lipgloss.NewStyle().
		Background(lipgloss.Color("#50FA7B")).
		Foreground(lipgloss.Color("#282a36")).
		Bold(true).
		Padding(0, 1).
		Render(config.Version)

	headerText := lipgloss.JoinHorizontal(lipgloss.Center,
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#BD93F9")).Render("DevCLI - Unified Developer Workspace "),
		versionBadge,
	)

	headerBanner := lipgloss.Place(m.width, 6, lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center, logo, headerText),
	)

	// --- COMMANDS VIEW ---
	if m.showCommands {
		commandsTitle := lipgloss.NewStyle().
			Width(m.width).
			Align(lipgloss.Center).
			Render(titleStyle.Render("DEVCLI COMMANDS"))

		footer := lipgloss.NewStyle().
			Width(m.width).
			Align(lipgloss.Center).
			Foreground(lipgloss.Color("#666666")).
			Render("Press [Esc] to return")

		content := lipgloss.JoinVertical(lipgloss.Center,
			topBar,
			"\n",
			commandsTitle,
			"\n",
			m.commandView.View(),
			"\n",
			footer,
		)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Top, content)
	}

	// 3. Dual-Pane Layout Construction
	leftWidth := (m.width * 45) / 100
	if leftWidth < 38 {
		leftWidth = 38
	}
	rightWidth := m.width - leftWidth - 6
	if rightWidth < 30 {
		rightWidth = 30
	}

	listHeight := m.height - 14
	if listHeight < 10 {
		listHeight = 10
	}
	m.list.SetSize(leftWidth-4, listHeight)

	paneHeight := m.height - 12
	if paneHeight < 10 {
		paneHeight = 10
	}

	leftBox := LeftPaneStyle.
		Width(leftWidth).
		Height(paneHeight).
		Render(m.list.View())

	// Dynamic Right Pane Card
	selectedTitle := "Quick Info"
	selectedDesc := "Select an option from the menu to see details."
	if i, ok := m.list.SelectedItem().(item); ok {
		selectedTitle = i.title
		selectedDesc = i.desc
	}

	rightContentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F1FA8C"))
	infoTitleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF79C6")).Bold(true)

	var rightInfo strings.Builder
	rightInfo.WriteString(infoTitleStyle.Render("📌 "+selectedTitle) + "\n\n")
	rightInfo.WriteString(rightContentStyle.Render(selectedDesc) + "\n\n")
	rightInfo.WriteString("------------------------------------\n\n")

	// Dynamic Help/Preview content based on selected menu item
	switch {
	case strings.Contains(selectedTitle, "Project"):
		rightInfo.WriteString("• Scaffold Go, Python, Node, React & FastAPI projects\n")
		rightInfo.WriteString("• Automatically initializes Git repositories\n")
		rightInfo.WriteString("• Generates structured README & build configs\n")
	case strings.Contains(selectedTitle, "AI Chat"):
		rightInfo.WriteString("• Multi-provider AI assistant\n")
		rightInfo.WriteString("• Supports Ollama, Gemini, OpenAI, Claude & HuggingFace\n")
		rightInfo.WriteString("• Code explanation, debugging & refactoring\n")
	case strings.Contains(selectedTitle, "Editor"):
		rightInfo.WriteString("• Built-in terminal code editor with line numbers\n")
		rightInfo.WriteString("• Syntax highlighting for multiple languages\n")
	case strings.Contains(selectedTitle, "File Manager"):
		rightInfo.WriteString("• Full-featured terminal file manager\n")
		rightInfo.WriteString("• Read, write, move, rename, delete files\n")
	case strings.Contains(selectedTitle, "Docker"):
		rightInfo.WriteString("• View active & stopped containers\n")
		rightInfo.WriteString("• Start, stop & restart containers\n")
		rightInfo.WriteString("• Stream live container logs in viewport\n")
	case strings.Contains(selectedTitle, "API"):
		rightInfo.WriteString("• Built-in Postman alternative in TUI\n")
		rightInfo.WriteString("• Test GET, POST, PUT, DELETE, PATCH endpoints\n")
		rightInfo.WriteString("• Pretty JSON response formatting & latency counter\n")
	case strings.Contains(selectedTitle, "Settings"):
		rightInfo.WriteString("• Configure AI API keys & custom base URLs\n")
		rightInfo.WriteString("• Customize compiler execution paths\n")
	default:
		rightInfo.WriteString("Press [Enter] to open the selected feature.\n")
		rightInfo.WriteString("Press [?] for help or [q] to exit.\n")
	}

	rightBox := RightPaneStyle.
		Width(rightWidth).
		Height(paneHeight).
		Render(rightInfo.String())

	dualPane := lipgloss.JoinHorizontal(lipgloss.Top, leftBox, " ", rightBox)

	// 4. Bottom Keybinding Bar
	keyBar := lipgloss.NewStyle().
		Background(lipgloss.Color("#44475a")).
		Foreground(lipgloss.Color("#F8F8F2")).
		Width(m.width).
		Align(lipgloss.Center).
		Render("  [↑/↓] Navigate  •  [Enter] Select  •  [?] Commands  •  [q] Quit  ")

	// Combine all elements into final sleek UI layout
	return lipgloss.JoinVertical(lipgloss.Left,
		topBar,
		headerBanner,
		dualPane,
		keyBar,
	)
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
