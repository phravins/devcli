package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// Global States
const (
	StateDashboard = iota
	StateProject
	StateFileManager
	StateChat
	StateEditor
	StateAutoUpdate
	StateDocs
	StateDocker
	StateAPIClient
)

// Messages
type SwitchViewMsg struct {
	TargetState int
	Args        interface{} // Generic args (e.g., initial path)
}

type BackMsg struct{}

// Feature-specific Back Messages for nested navigation
type VenvBackMsg struct{}
type DevServerBackMsg struct{}
type BoilerplateBackMsg struct{}
type BonusBackMsg struct{}
type SubFeatureBackMsg struct{} // Intermediate navigation to parent menu

type RootModel struct {
	state  int
	width  int
	height int

	// Sub-models
	dashboard   DashboardModel
	project     ProjectDashboardModel
	fileManager FileManagerModel
	chat        ChatModel
	editor      model // Using the struct 'model' from editor.go
	autoupdate  AutoUpdateModel
	docs        DocsModel
	docker      DockerDashboardModel
	apiClient   APIClientModel

	// Generic error
	err error
}

func NewRootModel() RootModel {
	return RootModel{
		state:     StateDashboard,
		dashboard: NewDashboard(),
		project:   NewProjectDashboardModel(),
	}
}

func (m RootModel) Init() tea.Cmd {
	return tea.Batch(
		m.dashboard.Init(),
	)
}

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// We do NOT manually propagate here. The active model will receive it in the switch below.

	case SwitchViewMsg:
		m.state = msg.TargetState

		// Initialize the target model and apply current dimensions
		switch m.state {
		case StateFileManager:
			path := ""
			if p, ok := msg.Args.(string); ok {
				path = p
			}
			m.fileManager = NewFileManagerModel(path)
			// Resize immediately
			var fm tea.Model
			fm, cmd = m.fileManager.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
			m.fileManager = fm.(FileManagerModel)
			cmds = append(cmds, cmd, m.fileManager.Init())

		case StateChat:
			m.chat = NewChatModel()
			var cm tea.Model
			cm, cmd = m.chat.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
			m.chat = cm.(ChatModel)
			cmds = append(cmds, cmd, m.chat.Init())

		case StateEditor:
			filename := ""
			if f, ok := msg.Args.(string); ok {
				filename = f
			}
			m.editor = initialModel(filename)
			var em tea.Model
			em, cmd = m.editor.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
			m.editor = em.(model)
			cmds = append(cmds, cmd, m.editor.Init())

		case StateProject:
			m.project = NewProjectDashboardModel()
			var pm tea.Model
			pm, cmd = m.project.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
			m.project = pm.(ProjectDashboardModel)
			cmds = append(cmds, cmd, m.project.Init())

		case StateAutoUpdate:
			m.autoupdate = NewAutoUpdateModel()
			var am tea.Model
			am, cmd = m.autoupdate.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
			m.autoupdate = am.(AutoUpdateModel)
			cmds = append(cmds, cmd, m.autoupdate.Init())

		case StateDocs:
			m.docs = NewDocsModel()
			var dm tea.Model
			dm, cmd = m.docs.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
			m.docs = dm.(DocsModel)
			cmds = append(cmds, cmd, m.docs.Init())

		case StateDocker:
			m.docker = NewDockerDashboardModel()
			var dockM tea.Model
			dockM, cmd = m.docker.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
			m.docker = dockM.(DockerDashboardModel)
			cmds = append(cmds, cmd, m.docker.Init())

		case StateAPIClient:
			m.apiClient = NewAPIClientModel()
			var apiM tea.Model
			apiM, cmd = m.apiClient.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
			m.apiClient = apiM.(APIClientModel)
			cmds = append(cmds, cmd, m.apiClient.Init())
		}

	case BackMsg:
		if m.state == StateDashboard {
			return m, tea.Quit
		}
		m.state = StateDashboard
	}

	switch m.state {
	case StateDashboard:
		newM, newCmd := m.dashboard.Update(msg)
		m.dashboard = newM.(DashboardModel)
		cmds = append(cmds, newCmd)
	case StateProject:
		newM, newCmd := m.project.Update(msg)
		m.project = newM.(ProjectDashboardModel)
		cmds = append(cmds, newCmd)
	case StateFileManager:
		newM, newCmd := m.fileManager.Update(msg)
		m.fileManager = newM.(FileManagerModel)
		cmds = append(cmds, newCmd)
	case StateChat:
		newM, newCmd := m.chat.Update(msg)
		m.chat = newM.(ChatModel)
		cmds = append(cmds, newCmd)
	case StateEditor:
		newM, newCmd := m.editor.Update(msg)
		m.editor = newM.(model)
		cmds = append(cmds, newCmd)
	case StateAutoUpdate:
		newM, newCmd := m.autoupdate.Update(msg)
		m.autoupdate = newM.(AutoUpdateModel)
		cmds = append(cmds, newCmd)
	case StateDocs:
		newM, newCmd := m.docs.Update(msg)
		m.docs = newM.(DocsModel)
		cmds = append(cmds, newCmd)
	case StateDocker:
		newM, newCmd := m.docker.Update(msg)
		m.docker = newM.(DockerDashboardModel)
		cmds = append(cmds, newCmd)
	case StateAPIClient:
		newM, newCmd := m.apiClient.Update(msg)
		m.apiClient = newM.(APIClientModel)
		cmds = append(cmds, newCmd)
	}

	return m, tea.Batch(cmds...)
}

func (m RootModel) View() string {
	switch m.state {
	case StateDashboard:
		return m.dashboard.View()
	case StateProject:
		return m.project.View()
	case StateFileManager:
		return m.fileManager.View()
	case StateChat:
		return m.chat.View()
	case StateEditor:
		return m.editor.View()
	case StateAutoUpdate:
		return m.autoupdate.View()
	case StateDocs:
		return m.docs.View()
	case StateDocker:
		return m.docker.View()
	case StateAPIClient:
		return m.apiClient.View()
	}
	return "Unknown State"
}

func RunRoot() {
	p := tea.NewProgram(NewRootModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running devcli: %v\n", err)
		os.Exit(1)
	}
}
