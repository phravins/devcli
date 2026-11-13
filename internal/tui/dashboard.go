package tui

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/phravins/devcli/internal/config"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type item struct {
	id, title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type customItemDelegate struct{}

func (d customItemDelegate) Height() int                             { return 2 }
func (d customItemDelegate) Spacing() int                            { return 1 }
func (d customItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d customItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	isSelected := index == m.Index()

	var bar, title, desc string

	if isSelected {
		bar = dashSelectedStyle.Render("│ ")
		title = dashSelectedStyle.Render(i.title)
		desc = dashSelectedStyle.Render(i.desc)
	} else {
		bar = "  "
		title = dashItemTitleStyle.Render(i.title)
		desc = dashItemDescStyle.Render(i.desc)
	}

	// Clip to the list width so long descriptions don't wrap and break the
	// fixed two-line item height on narrow terminals.
	if maxW := m.Width() - 2; maxW > 0 {
		if lipgloss.Width(title) > maxW {
			title = lipgloss.NewStyle().MaxWidth(maxW).Render(title)
		}
		if lipgloss.Width(desc) > maxW {
			desc = lipgloss.NewStyle().MaxWidth(maxW).Render(desc)
		}
	}

	if i.desc == "" {
		fmt.Fprintf(w, "%s%s", bar, title)
	} else {
		fmt.Fprintf(w, "%s%s\n  %s", bar, title, desc)
	}
}

const dashLogo = `  ____  _______     __   ____ _     ___
 |  _ \| ____\ \   / /  / ___| |   |_ _|
 | | | |  _|  \ \ / /  | |   | |    | |
 | |_| | |___  \ V /   | |___| |___ | |
 |____/|_____|  \_/     \____|_____|___|`

// Styles are built once: constructing lipgloss styles inside View() and the
// item delegate allocated on every frame and every list row.
var (
	dashSelectedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#BD93F9")).Bold(true)
	dashItemTitleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#EFE9E0")).Bold(true)
	dashItemDescStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	dashLogoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#0F9E99")).Bold(true)
	dashTitleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#EFE9E0"))
	dashVersionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Italic(true)
	dashFooterStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))

	dashLogoWidth = lipgloss.Width(dashLogo)
)

// renderDashboardHeader renders the centred logo block. The ASCII logo is
// rendered as one block (all lines padded to the same width) and then
// centred as a unit, so its letters stay aligned; on terminals narrower than
// the logo it falls back to a one-line wordmark instead of wrapping into
// garbage. On short terminals only the wordmark is shown so the menu fits.
func renderDashboardHeader(width, height int) string {
	if height > 0 && height < 20 {
		return lipgloss.PlaceHorizontal(max(width, 0), lipgloss.Center,
			dashLogoStyle.Render("DEVCLI")+"  "+dashVersionStyle.Render(config.Version))
	}
	var logo string
	if width >= dashLogoWidth+2 {
		logo = dashLogoStyle.Render(dashLogo)
	} else {
		logo = dashLogoStyle.Render("DEVCLI")
	}
	block := lipgloss.JoinVertical(lipgloss.Center,
		logo,
		dashTitleStyle.Render("Developer's CLI"),
		dashVersionStyle.Render(config.Version),
	)
	if width <= 0 {
		return block
	}
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, block)
}

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
	header       string // rendered on resize; the logo never changes per frame
}

func NewDashboard() DashboardModel {
	items := []list.Item{
		item{title: "Project Tools", desc: "Create projects, sync, clone, scan"},
		item{title: "AI Chat", desc: "Chat with AI models"},
		item{title: "Editor", desc: "Built-in code editor"},
		item{title: "File Manager", desc: "Explore, Search, and Manage Files (RW/Move)"},
		item{title: "Settings / Configuration", desc: "Configure AI backends and Keys"},
		item{title: "DevCLI Commands", desc: "List all available project commands"},
		item{title: "API & HTTP Playground", desc: "Built-in Postman alternative"},
		item{title: "Docker Dashboard", desc: "Inspect and manage containers"},
		item{title: "Auto-Update", desc: "Update Languages, AI Keys, and DevCLI"},
		item{title: "..", desc: ""},
	}

	delegate := customItemDelegate{}
	l := list.New(items, delegate, 0, 0)
	count := 0
	for _, it := range items {
		if it.(item).title != ".." {
			count++
		}
	}
	l.Title = fmt.Sprintf("%d items", count)
	l.SetShowTitle(true)
	l.SetShowStatusBar(false)
	l.SetShowHelp(true)

	l.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Padding(0, 0, 1, 0)

	m := DashboardModel{
		list:     l,
		settings: NewSettingsModel(),
	}

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
				m.settings.quitting = false
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
				if strings.Contains(i.title, "DevCLI Commands") {
					m.showCommands = true
					m.commandView.SetContent(generateCommandsHelp())
					m.commandView.GotoTop()
					return m, nil
				}
				if strings.Contains(i.title, "Settings / Configuration") {
					m.showSettings = true

					m.settings = NewSettingsModel()

					if m.width > 0 && m.height > 0 {
						updatedSettings, _ := m.settings.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
						m.settings = updatedSettings.(SettingsModel)
					}
					return m, m.settings.inputs[0].Focus()
				}
				if strings.Contains(i.title, "File Manager") {
					m.choice = i.title
					return m, func() tea.Msg { return SwitchViewMsg{TargetState: StateFileManager} }
				}
				if strings.Contains(i.title, "Project Tools") {
					m.choice = i.title
					return m, func() tea.Msg { return SwitchViewMsg{TargetState: StateProject} }
				}
				if strings.Contains(i.title, "AI Chat") {
					m.choice = i.title
					return m, func() tea.Msg { return SwitchViewMsg{TargetState: StateChat} }
				}
				if strings.Contains(i.title, "Editor") {
					m.choice = i.title
					return m, func() tea.Msg { return SwitchViewMsg{TargetState: StateEditor} }
				}
				if strings.Contains(i.title, "Auto-Update") {
					m.choice = i.title
					return m, func() tea.Msg { return SwitchViewMsg{TargetState: StateAutoUpdate} }
				}
				if strings.Contains(i.title, "Docker Dashboard") {
					m.choice = i.title
					return m, func() tea.Msg { return SwitchViewMsg{TargetState: StateDocker} }
				}
				if strings.Contains(i.title, "API & HTTP Playground") {
					m.choice = i.title
					return m, func() tea.Msg { return SwitchViewMsg{TargetState: StateAPIClient} }
				}
				if strings.Contains(i.title, "Docs") {
					m.choice = i.title
					return m, func() tea.Msg { return SwitchViewMsg{TargetState: StateDocs} }
				}

				m.choice = i.title
				m.quitting = true
				return m, tea.Quit
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
		// header + blank spacer line + footer line
		m.header = renderDashboardHeader(msg.Width, msg.Height)
		chrome := lipgloss.Height(m.header) + 2
		listHeight := msg.Height - v - chrome
		if listHeight < 3 {
			listHeight = 3
		}
		m.list.SetSize(msg.Width-h, listHeight)

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

	footer := lipgloss.PlaceHorizontal(m.width, lipgloss.Center, dashFooterStyle.Render("OPENDEV TOOLKIT"))

	if m.showCommands {
		commandsTitle := lipgloss.NewStyle().
			Width(m.width).
			Align(lipgloss.Center).
			Render(titleStyle.Render("DEVCLI COMMANDS"))

		content := lipgloss.JoinVertical(lipgloss.Center,
			"\n",
			commandsTitle,
			strings.Repeat("\n", 1),
			m.commandView.View(),
			strings.Repeat("\n", 1),
			footer,
		)

		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Top, content)
	}

	header := m.header
	if header == "" {
		header = renderDashboardHeader(m.width, m.height)
	}
	contentView := header + "\n\n" + m.list.View()

	// Pad so the footer sits on the last line without overflowing.
	gap := m.height - lipgloss.Height(contentView) - lipgloss.Height(footer)
	if gap < 0 {
		gap = 0
	}

	return docStyle.Render(contentView + strings.Repeat("\n", gap+1) + footer)
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
