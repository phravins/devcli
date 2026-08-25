package tui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/phravins/devcli/internal/timemachine"
)

type BonusDashboardModel struct {
	menuList list.Model
	state    int
	width    int
	height   int

	// Sub-features
	taskRunnerModel  TaskRunnerModel
	smartFileModel   SmartFileModel
	snippetsModel    SnippetsModel
	projectDashModel ProjectDashModel
	aiAssistantModel AIAssistantModel
	timeMachineModel interface{} // Will hold *TimeMachineModel
	helpView         viewport.Model

	// File picker for Code Time Machine
	filePickerList  list.Model
	filePickerReady bool
	filePickerCwd   string
}

const (
	StateBonusMenu = iota
	StateBonusTaskRunner
	StateBonusSmartFile
	StateBonusSnippets
	StateBonusProjectDash
	StateBonusAIAssistant
	StateBonusTimeMachinePicker // NEW: file picker screen
	StateBonusTimeMachine
	StateBonusHelp // Help Screen
)

func NewBonusDashboardModel(workspace string) BonusDashboardModel {
	// Always resolve to actual CWD if workspace is empty (e.g. globally installed binary)
	if workspace == "" {
		if cwd, err := os.Getwd(); err == nil {
			workspace = cwd
		}
	}
	items := []list.Item{
		item{title: "Task Runner", desc: "One-click build, test, format, and lint"},
		item{title: "Smart File Creator", desc: "Generate config files (.env, Dockerfile, etc.)"},
		item{title: "Project Dashboard", desc: "Overview of all your projects and stats"},
		item{title: "Snippet Library", desc: "Personal vault of reusable code"},
		item{title: "AI Assistant", desc: "AI-powered code generation and assistance"},
		item{title: "Code Time Machine", desc: "Track code evolution, find bugs, and analyze history"},
	}

	menu := list.New(items, list.NewDefaultDelegate(), 60, 14)
	menu.Title = "Bonus Features"
	menu.SetShowHelp(true)
	menu.SetShowTitle(true)

	return BonusDashboardModel{
		menuList:         menu,
		state:            StateBonusMenu,
		taskRunnerModel:  NewTaskRunnerModel(workspace),
		smartFileModel:   NewSmartFileModel(workspace),
		snippetsModel:    NewSnippetsModel(),
		projectDashModel: NewProjectDashModel(workspace),
		aiAssistantModel: NewAIAssistantModel(),
		helpView:         viewport.New(80, 20),
	}
}

func (m BonusDashboardModel) Init() tea.Cmd {
	return nil
}

// buildFilePicker creates list.Model with all git-tracked files in cwd.
func buildFilePicker(cwd string, w, h int) (list.Model, bool) {
	files, err := timemachine.GetTrackedFiles(cwd)
	if err != nil || len(files) == 0 {
		return list.Model{}, false
	}

	listItems := make([]list.Item, 0, len(files))
	for _, f := range files {
		listItems = append(listItems, item{title: f, desc: ""})
	}

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false

	l := list.New(listItems, delegate, w-4, h-10)
	l.Title = "Select a file to explore its history"
	l.SetShowHelp(true)
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	return l, true
}

func (m BonusDashboardModel) Update(msg tea.Msg) (BonusDashboardModel, tea.Cmd) {
	var cmd tea.Cmd

	// Priority: Handle global messages like "back" regardless of state
	switch msg.(type) {
	case BonusBackMsg:
		m.state = StateBonusMenu
		return m, nil
	case SubFeatureBackMsg:
		// Coming back from Time Machine — go back to the file picker, not the menu
		if m.state == StateBonusTimeMachine && m.filePickerReady {
			m.state = StateBonusTimeMachinePicker
			return m, nil
		}
		m.state = StateBonusMenu
		return m, nil
	}

	// Global Help Toggle
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "?" && m.state == StateBonusMenu {
			m.state = StateBonusHelp
			m.helpView.SetContent(RenderHelp(BonusFeaturesHelp, m.width-2, m.height))
			return m, nil
		}
	}

	// Handle delegation to sub-features based on current state
	switch m.state {
	case StateBonusTaskRunner:
		var trCmd tea.Cmd
		m.taskRunnerModel, trCmd = m.taskRunnerModel.Update(msg)
		return m, trCmd

	case StateBonusSmartFile:
		var sfCmd tea.Cmd
		m.smartFileModel, sfCmd = m.smartFileModel.Update(msg)
		return m, sfCmd

	case StateBonusSnippets:
		var snCmd tea.Cmd
		m.snippetsModel, snCmd = m.snippetsModel.Update(msg)
		return m, snCmd
	case StateBonusProjectDash:
		var pdCmd tea.Cmd
		m.projectDashModel, pdCmd = m.projectDashModel.Update(msg)
		return m, pdCmd

	case StateBonusAIAssistant:
		var aiCmd tea.Cmd
		m.aiAssistantModel, aiCmd = m.aiAssistantModel.Update(msg)
		return m, aiCmd

	case StateBonusTimeMachine:
		if m.timeMachineModel != nil {
			if tm, ok := m.timeMachineModel.(*TimeMachineModel); ok {
				var tmCmd tea.Cmd
				var updatedModel tea.Model
				updatedModel, tmCmd = tm.Update(msg)
				if updated, ok := updatedModel.(*TimeMachineModel); ok {
					m.timeMachineModel = updated
				}
				return m, tmCmd
			}
		}
		return m, nil

	case StateBonusTimeMachinePicker:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc", "q":
				// Back to bonus menu
				m.state = StateBonusMenu
				return m, nil
			case "enter":
				if sel, ok := m.filePickerList.SelectedItem().(item); ok {
					filePath := sel.title
					cwd := m.filePickerCwd
					if tm, err := NewTimeMachineModel(cwd, filePath); err == nil {
						m.timeMachineModel = tm
						m.state = StateBonusTimeMachine
						return m, tm.Init()
					} else {
						// Stay in picker; show error in title
						m.filePickerList.Title = fmt.Sprintf("❌ No history for %s — pick another", filePath)
						return m, nil
					}
				}
			}
		case tea.WindowSizeMsg:
			m.width = msg.Width
			m.height = msg.Height
			m.filePickerList.SetWidth(msg.Width - 4)
			m.filePickerList.SetHeight(msg.Height - 10)
		}
		// Forward everything else (typing filter, scrolling) to the file picker list
		m.filePickerList, cmd = m.filePickerList.Update(msg)
		return m, cmd

	case StateBonusHelp:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "esc" || msg.String() == "q" || msg.String() == "enter" || msg.String() == "?" {
				m.state = StateBonusMenu
				return m, nil
			}
		// Mouse wheel
		case tea.MouseMsg:
			if msg.Type == tea.MouseWheelUp {
				m.helpView.LineUp(3)
			}
			if msg.Type == tea.MouseWheelDown {
				m.helpView.LineDown(3)
			}
		}
		var cmd tea.Cmd
		m.helpView, cmd = m.helpView.Update(msg)
		return m, cmd
	}

	// Handle menu navigation
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case StateBonusMenu:
			switch msg.String() {
			case "esc", "q":
				return m, func() tea.Msg { return BonusBackMsg{} }

			case "enter":
				i, ok := m.menuList.SelectedItem().(item)
				if ok {
					switch i.title {
					case "Task Runner":
						m.state = StateBonusTaskRunner
						return m, m.taskRunnerModel.Init()
					case "Smart File Creator":
						m.state = StateBonusSmartFile
						return m, m.smartFileModel.Init()
					case "Snippet Library":
						m.state = StateBonusSnippets
						return m, m.snippetsModel.Init()
					case "Project Dashboard":
						m.state = StateBonusProjectDash
						return m, m.projectDashModel.Init()
					case "AI Assistant":
						m.state = StateBonusAIAssistant
						return m, m.aiAssistantModel.Init()
					case "Code Time Machine":
						// Show file picker for all git-tracked files
						cwd := m.taskRunnerModel.workspace
						if cwd == "" {
							if c, err := os.Getwd(); err == nil {
								cwd = c
							} else {
								cwd = "."
							}
						}
						w := m.width
						if w == 0 {
							w = 80
						}
						h := m.height
						if h == 0 {
							h = 40
						}
						if picker, ok := buildFilePicker(cwd, w, h); ok {
							m.filePickerList = picker
							m.filePickerReady = true
							m.filePickerCwd = cwd
							m.state = StateBonusTimeMachinePicker
						}
						// If not a git repo or no files — stay in menu
						return m, nil
					}
				}
			}
			m.menuList, cmd = m.menuList.Update(msg)
			return m, cmd
		}

	case tea.MouseMsg:
		// Forward to active sub-models first
		switch m.state {
		case StateBonusTaskRunner:
			m.taskRunnerModel, cmd = m.taskRunnerModel.Update(msg)
			return m, cmd
		case StateBonusSmartFile:
			m.smartFileModel, cmd = m.smartFileModel.Update(msg)
			return m, cmd
		case StateBonusSnippets:
			m.snippetsModel, cmd = m.snippetsModel.Update(msg)
			return m, cmd
		case StateBonusProjectDash:
			m.projectDashModel, cmd = m.projectDashModel.Update(msg)
			return m, cmd
		case StateBonusAIAssistant:
			m.aiAssistantModel, cmd = m.aiAssistantModel.Update(msg)
			return m, cmd
		}

		// If we are in Menu state, handle list scroll
		if m.state == StateBonusMenu {
			if msg.Type == tea.MouseWheelUp {
				m.menuList.CursorUp()
				return m, nil
			}
			if msg.Type == tea.MouseWheelDown {
				m.menuList.CursorDown()
				return m, nil
			}
			m.menuList, cmd = m.menuList.Update(msg)
			return m, cmd
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.menuList.SetSize(msg.Width-4, msg.Height-8)

		// Resize sub-models
		m.taskRunnerModel, _ = m.taskRunnerModel.Update(msg)
		m.smartFileModel, _ = m.smartFileModel.Update(msg)
		m.snippetsModel, _ = m.snippetsModel.Update(msg)
		m.projectDashModel, _ = m.projectDashModel.Update(msg)
		m.aiAssistantModel, _ = m.aiAssistantModel.Update(msg)

		m.helpView.Width = msg.Width
		m.helpView.Height = msg.Height
		if m.state == StateBonusHelp {
			m.helpView.SetContent(RenderHelp(BonusFeaturesHelp, m.width-2, m.height))
		}

		// Resize file picker too
		if m.filePickerReady {
			m.filePickerList.SetWidth(msg.Width - 4)
			m.filePickerList.SetHeight(msg.Height - 10)
		}
	}

	return m, nil
}

func (m BonusDashboardModel) View() string {
	switch m.state {
	case StateBonusTaskRunner:
		return m.taskRunnerModel.View()
	case StateBonusSmartFile:
		return m.smartFileModel.View()
	case StateBonusSnippets:
		return m.snippetsModel.View()
	case StateBonusProjectDash:
		return m.projectDashModel.View()
	case StateBonusAIAssistant:
		return m.aiAssistantModel.View()

	case StateBonusTimeMachinePicker:
		header := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF6B6B")).
			Background(lipgloss.Color("#1A1A2E")).
			Width(m.width).
			Padding(0, 2).
			Render("Code Time Machine  -  Choose a file")

		footer := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#555555")).
			Background(lipgloss.Color("#0D0D1A")).
			Width(m.width).
			Render("  ↑/↓ Navigate  │  / Filter  │  Enter: Open  │  Q/Esc: Back")

		return lipgloss.JoinVertical(lipgloss.Left,
			header,
			m.filePickerList.View(),
			footer,
		)

	case StateBonusTimeMachine:
		if m.timeMachineModel != nil {
			if tm, ok := m.timeMachineModel.(*TimeMachineModel); ok {
				return tm.View()
			}
		}
		return "Code Time Machine not initialized. Press ESC to return."

	case StateBonusHelp:
		helpWithBorder := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#0F9E99")).
			Render(m.helpView.View())
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, helpWithBorder)
	}

	// Menu view
	footer := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(m.width).
		Render(subtleStyle.Render("↑/↓: Navigate • Enter: Select • Q/Esc: Back"))

	// Header
	header := lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(
		titleStyle.Render("Bonus Features"),
	)

	content := lipgloss.JoinVertical(lipgloss.Left,
		header,
		"\n",
		m.menuList.View(),
		"\n",
		footer,
	)

	return docStyle.Render(content)
}
