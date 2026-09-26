package tui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/phravins/devcli/internal/history"
	"github.com/phravins/devcli/internal/project"
	"github.com/phravins/devcli/internal/templates"
)

type ProjectDashboardModel struct {
	menuList	list.Model
	projectList	list.Model
	templateList	list.Model
	input		textinput.Model
	pathInput	textinput.Model
	spinner		spinner.Model
	historyList	list.Model

	state		int
	previousState	int
	width, height	int

	manager	*project.Manager

	venvModel		VenvDashboardModel
	devServerModel		DevServerDashboardModel
	boilerplateModel	BoilerplateDashboardModel
	bonusModel		BonusDashboardModel

	selectedTpl	string
	err		error
	statusMsg	string

	installOutput	*strings.Builder
	installView	viewport.Model
	helpView	viewport.Model
}

const (
	StateMenu	= iota
	StateProjectList
	StateSelectTemplate
	StateNameProject
	StateSelectPath
	StateCreating
	StateSuccess
	StateBackupInput
	StateCleanupPrompt
	StateHistoryList
	StateConfirmDelete
	StateProjectHelp

	StateVenvWizard
	StateDevServer
	StateBoilerplate
	StateBonus
)

func NewProjectDashboardModel() ProjectDashboardModel {
	mgr := project.NewManager("")

	menuItems := []list.Item{
		item{title: "Project Creation & Management", desc: "Create, list, and manage local projects"},
		item{title: "Virtual Environment Wizard", desc: "Manage Python/Node environments, sync packages, etc."},
		item{title: "Dev Server", desc: "Auto-detect & launch development servers with live logs"},
		item{title: "Boilerplate Generator", desc: "Code presets, templates, and architecture generation"},
		item{title: "Bonus Features", desc: "Project Dashboard, Task Runner, Smart Files, Snippets, AI Assistant"},
		item{title: "Project History", desc: "View creation logs (Auto-cleanup > 30 days)"},
	}
	menu := list.New(menuItems, list.NewDefaultDelegate(), 0, 0)
	menu.Title = "Project Tools"
	menu.SetShowHelp(false)
	menu.SetShowTitle(false)

	items := []list.Item{item{title: "Loading projects...", desc: "Please wait"}}
	items = append([]list.Item{item{title: "+ New Project", desc: "Create a new project from template"}}, items...)
	pl := list.New(items, list.NewDefaultDelegate(), 0, 0)
	pl.Title = "My Projects"
	pl.SetShowHelp(false)

	var tplItems []list.Item
	for _, t := range templates.List() {
		tplItems = append(tplItems, item{title: t.Name, desc: t.Description})
	}
	tplList := list.New(tplItems, list.NewDefaultDelegate(), 0, 0)
	tplList.Title = "Select Project Template (v2)"
	tplList.SetShowHelp(false)

	histList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	histList.Title = "Project History"
	histList.SetShowHelp(false)
	histList.SetShowTitle(false)

	ti := textinput.New()
	ti.Placeholder = "Project Name"
	ti.CharLimit = 50
	ti.Width = 40

	pi := textinput.New()
	pi.Placeholder = "Parent Directory (e.g. C:\\Projects or ~)"

	pi.SetValue(mgr.Workspace)
	pi.CharLimit = 100
	pi.Width = 50

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("62"))

	hv := viewport.New(80, 20)
	hv.Style = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("62")).Padding(1, 2)

	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	out, err := renderer.Render(ProjectToolsHelp)
	if err != nil {
		out = ProjectToolsHelp
	}
	hv.SetContent(out)

	return ProjectDashboardModel{
		menuList:		menu,
		projectList:		pl,
		templateList:		tplList,
		historyList:		histList,
		input:			ti,
		pathInput:		pi,
		spinner:		s,
		manager:		mgr,
		venvModel:		NewVenvDashboardModel(),
		devServerModel:		NewDevServerDashboardModel(mgr.Workspace),
		boilerplateModel:	NewBoilerplateDashboardModel(mgr.Workspace),
		bonusModel:		NewBonusDashboardModel(mgr.Workspace),
		state:			StateMenu,
		installOutput:		&strings.Builder{},
		installView:		vp,
		helpView:		hv,
	}
}

func loadProjects(workspace string) []list.Item {
	entries, err := os.ReadDir(workspace)
	if err != nil {
		return []list.Item{}
	}
	var items []list.Item
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {

			fullPath := filepath.Join(workspace, e.Name())
			if isProject(fullPath) {
				info, err := e.Info()
				desc := "Existing Project"
				if err == nil {
					modTime := info.ModTime().Format("2006-01-02 15:04")
					desc = fmt.Sprintf("Path: %s | Modified: %s", fullPath, modTime)
				}
				items = append(items, item{title: e.Name(), desc: desc})
			}
		}
	}
	return items
}

func isProject(dir string) bool {
	markers := []string{
		"go.mod",
		"package.json",
		"requirements.txt",
		".git",
		"main.py",
		"main.go",
		"index.js",
		"README.md",
	}
	for _, m := range markers {
		if _, err := os.Stat(filepath.Join(dir, m)); err == nil {
			return true
		}
	}
	return false
}

func (m ProjectDashboardModel) Init() tea.Cmd {

	old := history.GetOldEntries(30)
	if len(old) > 0 {
		return tea.Batch(
			func() tea.Msg { return cleanupPromptMsg{} },
			m.spinner.Tick,
			m.venvModel.Init(),
		)
	}
	return tea.Batch(
		tea.Tick(10*time.Millisecond, func(t time.Time) tea.Msg {
			return loadProjectsMsg(loadProjects(m.manager.Workspace))
		}),
		m.spinner.Tick,
		m.venvModel.Init(),
		m.boilerplateModel.Init(),
	)
}

type cleanupPromptMsg struct{}

type projectCreatedMsg struct {
	installCmd	string
	path		string
	err		error
}

type delayedSuccessMsg struct{}

type loadProjectsMsg []list.Item

type installDoneMsg struct{ err error }

type cmdProcess struct {
	cmd	*exec.Cmd
	reader	*bufio.Reader
}

type installStartedMsg struct {
	proc *cmdProcess
}
type installOutputMsg struct {
	line	string
	proc	*cmdProcess
}

func createProjectCmd(mgr *project.Manager, name, stack, path string) tea.Cmd {
	return func() tea.Msg {

		cmdStr, resolvedPath, err := mgr.CreateProject(name, stack, path)
		return projectCreatedMsg{installCmd: cmdStr, path: resolvedPath, err: err}
	}
}

func startInstallCmd(dir, cmdStr string) tea.Cmd {
	return func() tea.Msg {

		var c *exec.Cmd
		if runtime.GOOS == "windows" {
			fullCmd := fmt.Sprintf("@echo on & echo [DevCLI] Starting installation process... & echo [DevCLI] Directory: %s & echo [DevCLI] Running: %s & echo ---------------------------------------- & call %s & echo. & echo ---------------------------------------- & echo [DevCLI] Process Completed.", dir, cmdStr, cmdStr)
			c = exec.Command("cmd", "/c", fullCmd)
		} else {

			fullCmd := fmt.Sprintf("echo '[DevCLI] Starting installation process...' && echo '[DevCLI] Directory: %s' && echo '[DevCLI] Running: %s' && echo '----------------------------------------' && %s && echo '' && echo '----------------------------------------' && echo '[DevCLI] Process Completed.'", dir, cmdStr, cmdStr)
			c = exec.Command("sh", "-c", fullCmd)
		}
		c.Dir = dir

		outPipe, _ := c.StdoutPipe()
		c.Stderr = c.Stdout

		if err := c.Start(); err != nil {
			return installDoneMsg{err: err}
		}

		return installStartedMsg{
			proc: &cmdProcess{
				cmd:	c,
				reader:	bufio.NewReader(outPipe),
			},
		}
	}
}

func readNextLine(proc *cmdProcess) tea.Cmd {
	return func() tea.Msg {
		if proc == nil || proc.reader == nil {
			return installDoneMsg{err: nil}
		}
		line, err := proc.reader.ReadString('\n')
		if err != nil {
			proc.cmd.Wait()
			if err == io.EOF {
				return installDoneMsg{err: nil}
			}
			return installDoneMsg{err: err}
		}
		return installOutputMsg{line: line, proc: proc}
	}
}

func (m ProjectDashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if items, ok := msg.(loadProjectsMsg); ok {

		finalItems := append([]list.Item{item{title: "+ New Project", desc: "Create a new project from template"}}, items...)
		m.projectList.SetItems(finalItems)
		return m, nil
	}

	switch msg.(type) {
	case VenvBackMsg, DevServerBackMsg, BoilerplateBackMsg, BonusBackMsg:
		m.state = StateMenu
		return m, nil
	case BackMsg:
		m.state = StateMenu
		return m, nil
	}

	if m.state == StateDevServer {
		var devCmd tea.Cmd
		var devModel tea.Model

		if wMsg, ok := msg.(tea.WindowSizeMsg); ok {
			h, v := AppBorderStyle.GetFrameSize()
			innerMsg := tea.WindowSizeMsg{Width: wMsg.Width - h, Height: wMsg.Height - v}
			devModel, devCmd = m.devServerModel.Update(innerMsg)
		} else {
			devModel, devCmd = m.devServerModel.Update(msg)
		}
		m.devServerModel = devModel.(DevServerDashboardModel)
		return m, devCmd
	}

	if m.state == StateVenvWizard {
		var venvCmd tea.Cmd

		if wMsg, ok := msg.(tea.WindowSizeMsg); ok {
			h, v := AppBorderStyle.GetFrameSize()
			innerMsg := tea.WindowSizeMsg{Width: wMsg.Width - h, Height: wMsg.Height - v}
			m.venvModel, venvCmd = m.venvModel.Update(innerMsg)
		} else {
			m.venvModel, venvCmd = m.venvModel.Update(msg)
		}
		return m, venvCmd
	}

	if m.state == StateBoilerplate {
		var bpCmd tea.Cmd

		if wMsg, ok := msg.(tea.WindowSizeMsg); ok {
			h, v := AppBorderStyle.GetFrameSize()
			innerMsg := tea.WindowSizeMsg{Width: wMsg.Width - h, Height: wMsg.Height - v}
			m.boilerplateModel, bpCmd = m.boilerplateModel.Update(innerMsg)
		} else {
			m.boilerplateModel, bpCmd = m.boilerplateModel.Update(msg)
		}
		return m, bpCmd
	}

	if m.state == StateBonus {
		var bonusCmd tea.Cmd

		if wMsg, ok := msg.(tea.WindowSizeMsg); ok {
			m.bonusModel, bonusCmd = m.bonusModel.Update(wMsg)
		} else {
			m.bonusModel, bonusCmd = m.bonusModel.Update(msg)
		}
		return m, bonusCmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.state == StateCreating {
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}

	case tea.MouseMsg:
		var cmd tea.Cmd

		switch m.state {
		case StateMenu:

			if msg.Type == tea.MouseWheelUp {
				m.menuList.CursorUp()
				return m, nil
			}
			if msg.Type == tea.MouseWheelDown {
				m.menuList.CursorDown()
				return m, nil
			}
			m.menuList, cmd = m.menuList.Update(msg)
		case StateProjectList:

			if msg.Type == tea.MouseWheelUp {
				m.projectList.CursorUp()
				return m, nil
			}
			if msg.Type == tea.MouseWheelDown {
				m.projectList.CursorDown()
				return m, nil
			}
			m.projectList, cmd = m.projectList.Update(msg)

		case StateSelectTemplate:
			if msg.Type == tea.MouseWheelUp {
				m.templateList.CursorUp()
				return m, nil
			}
			if msg.Type == tea.MouseWheelDown {
				m.templateList.CursorDown()
				return m, nil
			}
			m.templateList, cmd = m.templateList.Update(msg)
		case StateHistoryList:
			m.historyList, cmd = m.historyList.Update(msg)
		case StateProjectHelp:
			m.helpView, cmd = m.helpView.Update(msg)
		case StateCreating:
			m.installView, cmd = m.installView.Update(msg)
		}
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:

		switch m.state {
		case StateSuccess:
			switch msg.String() {
			case "enter", "esc":
				m.state = StateProjectList
				return m, nil
			}

		case StateMenu:
			switch msg.String() {
			case "?":
				m.state = StateProjectHelp
				m.helpView.GotoTop()
				return m, nil
			case "enter":
				i, ok := m.menuList.SelectedItem().(item)
				if ok {
					if i.title == "Project Creation & Management" {
						m.state = StateProjectList

						items := loadProjects(m.manager.Workspace)
						items = append([]list.Item{item{title: "+ New Project", desc: "Create a new project from template"}}, items...)
						m.projectList.SetItems(items)
						return m, nil
					}
					if i.title == "Virtual Environment Wizard" {
						m.state = StateVenvWizard
						m.venvModel = NewVenvDashboardModel()

						h, v := AppBorderStyle.GetFrameSize()
						innerW := m.width - h - 2
						innerH := m.height - v
						m.venvModel, _ = m.venvModel.Update(tea.WindowSizeMsg{Width: innerW, Height: innerH})
						return m, m.venvModel.Init()
					}
					if i.title == "Dev Server" {
						m.state = StateDevServer
						m.devServerModel = NewDevServerDashboardModel(m.manager.Workspace)

						h, v := AppBorderStyle.GetFrameSize()
						innerW := m.width - h - 2
						innerH := m.height - v
						devModel, _ := m.devServerModel.Update(tea.WindowSizeMsg{Width: innerW, Height: innerH})
						m.devServerModel = devModel.(DevServerDashboardModel)
						return m, m.devServerModel.Init()
					}
					if i.title == "Boilerplate Generator" {
						m.state = StateBoilerplate
						m.boilerplateModel = NewBoilerplateDashboardModel(m.manager.Workspace)

						h, v := AppBorderStyle.GetFrameSize()
						innerW := m.width - h - 2
						innerH := m.height - v
						m.boilerplateModel.resizeLists(innerW, innerH)
						return m, nil
					}
					if i.title == "Bonus Features" {
						m.state = StateBonus

						cwd, err := os.Getwd()
						if err != nil || cwd == "" {
							cwd = m.manager.Workspace
						}
						m.bonusModel = NewBonusDashboardModel(cwd)

						h, v := AppBorderStyle.GetFrameSize()
						innerW := m.width - h - 2
						innerH := m.height - v
						m.bonusModel, _ = m.bonusModel.Update(tea.WindowSizeMsg{Width: innerW, Height: innerH})
						return m, m.bonusModel.Init()
					}
					if i.title == "Project History" {
						m.state = StateHistoryList

						entries, _ := history.Load()
						var items []list.Item
						for _, e := range entries {
							desc := fmt.Sprintf("Path: %s | Time: %s", e.Path, e.CreatedAt.Format("2006-01-02 15:04"))
							items = append(items, item{title: e.Name, desc: desc})
						}
						m.historyList.SetItems(items)
						return m, nil
					}
				}
			case "q", "esc":
				return m, func() tea.Msg { return BackMsg{} }
			}
			m.menuList, cmd = m.menuList.Update(msg)
			return m, cmd

		case StateCleanupPrompt:
			if msg.String() == "enter" {

				history.DeleteOld(30)
				m.state = StateMenu
			} else if msg.String() == "esc" {
				m.state = StateMenu
			}
			return m, nil

		case StateHistoryList:
			switch msg.String() {
			case "esc":
				m.state = StateMenu
				return m, nil
			case "d":

				if len(m.historyList.Items()) > 0 {
					m.state = StateConfirmDelete
				}
				return m, nil
			}
			m.historyList, cmd = m.historyList.Update(msg)
			return m, cmd

		case StateConfirmDelete:
			switch msg.String() {
			case "esc", "n":
				m.state = StateHistoryList
				return m, nil
			case "enter", "y":

				idx := m.historyList.Index()
				if idx >= 0 && len(m.historyList.Items()) > 0 {
					history.DeleteOne(idx)

					entries, _ := history.Load()
					var items []list.Item
					for _, e := range entries {
						desc := fmt.Sprintf("Path: %s | Time: %s", e.Path, e.CreatedAt.Format("2006-01-02 15:04"))
						items = append(items, item{title: e.Name, desc: desc})
					}
					m.historyList.SetItems(items)
				}
				m.state = StateHistoryList
				return m, nil
			}

		case StateProjectHelp:
			switch msg.String() {
			case "esc", "enter", "?":
				m.state = m.previousState
				return m, nil
			}
			var cmd tea.Cmd
			m.helpView, cmd = m.helpView.Update(msg)
			return m, cmd

		case StateProjectList:
			switch msg.String() {
			case "?":
				m.previousState = StateProjectList
				m.state = StateProjectHelp
				m.helpView.GotoTop()
				return m, nil
			case "enter":
				i, ok := m.projectList.SelectedItem().(item)
				if ok && i.title == "+ New Project" {
					m.state = StateSelectTemplate
					m.templateList.ResetSelected()
					return m, nil
				}
			case "b":
				if len(m.projectList.Items()) > 1 {

					i, ok := m.projectList.SelectedItem().(item)
					if ok && i.title != "+ New Project" && i.desc == "Existing Project" {
						m.state = StateBackupInput
						m.pathInput.Placeholder = "Backup Destination (e.g. D:\\Backups)"
						m.pathInput.SetValue("")
						m.pathInput.Focus()
						return m, nil
					}
				}
			case "esc":

				m.state = StateMenu
				return m, nil
			}
			m.projectList, cmd = m.projectList.Update(msg)
			return m, cmd

		case StateBackupInput:
			switch msg.String() {
			case "esc":
				m.state = StateProjectList
			case "enter":
				dest := m.pathInput.Value()
				if dest != "" {

					i, _ := m.projectList.SelectedItem().(item)
					projectName := i.title
					srcPath := filepath.Join(m.manager.Workspace, projectName)

					m.state = StateCreating
					m.statusMsg = "Backing up project..."
					m.installOutput.Reset()
					m.installOutput.WriteString(fmt.Sprintf("Backing up '%s' to '%s'...\n", srcPath, dest))
					m.installView.SetContent(m.installOutput.String())

					return m, func() tea.Msg {
						err := m.manager.BackupProject(srcPath, dest)
						if err != nil {
							return installDoneMsg{err: err}
						}

						time.Sleep(1 * time.Second)
						return installDoneMsg{err: nil}
					}
				}
			}
			m.pathInput, cmd = m.pathInput.Update(msg)
			return m, cmd

		case StateSelectTemplate:
			switch msg.String() {
			case "enter":
				i, ok := m.templateList.SelectedItem().(item)
				if ok {
					m.selectedTpl = i.title

					suggestion := m.manager.SuggestProjectName(m.selectedTpl)
					m.input.SetValue(suggestion)

					m.state = StateNameProject
					m.input.Focus()
					return m, nil
				}
			case "esc":
				m.state = StateProjectList
				return m, nil
			}
			m.templateList, cmd = m.templateList.Update(msg)
			return m, cmd

		case StateNameProject:
			switch msg.String() {
			case "enter":
				if m.input.Value() != "" {

					m.state = StateSelectPath

					m.pathInput.Focus()

					return m, textinput.Blink
				}
			case "esc":
				m.state = StateSelectTemplate
				return m, nil
			}
			m.input, cmd = m.input.Update(msg)
			return m, cmd

		case StateSelectPath:
			switch msg.String() {
			case "enter":

				pathVal := m.pathInput.Value()
				_, err := m.manager.ValidateParentDir(pathVal)
				if err != nil {
					m.pathInput.SetValue(pathVal)
					m.pathInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
					m.err = err
					return m, nil
				}
				m.err = nil
				m.pathInput.TextStyle = lipgloss.NewStyle()

				m.state = StateCreating
				m.statusMsg = "Initializing Project..."
				m.installOutput.Reset()

				timestamp := time.Now().Format("2006-01-02 15:04:05")
				header := fmt.Sprintf("PROJECT CREATION LOG\n========================\nName : %s\nPath : %s\nTime : %s\n========================\n\n", m.input.Value(), pathVal, timestamp)
				m.installOutput.WriteString(header)
				m.installOutput.WriteString("Starting Project Generation...\n")
				m.installView.SetContent(m.installOutput.String())

				history.Add(m.input.Value(), pathVal)
				return m, createProjectCmd(m.manager, m.input.Value(), m.selectedTpl, pathVal)
			case "esc":
				m.state = StateNameProject
				m.input.Focus()
				return m, nil
			}
			m.pathInput, cmd = m.pathInput.Update(msg)
			return m, cmd
		}

	case cleanupPromptMsg:
		m.state = StateCleanupPrompt
		return m, nil

	case projectCreatedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.state = StateSelectPath
			return m, nil
		}

		m.installOutput.WriteString(fmt.Sprintf("Project files generated at %s\n", msg.path))
		m.installOutput.WriteString("Preparing to install dependencies...\n")
		m.installView.SetContent(m.installOutput.String())

		m.statusMsg = "Starting installation..."

		if msg.installCmd != "" {
			return m, startInstallCmd(msg.path, msg.installCmd)
		}
		m.statusMsg = "Project Created Successfully!"
		return m, func() tea.Msg { return delayedSuccessMsg{} }

	case installStartedMsg:
		m.statusMsg = "Installing packages..."

		return m, readNextLine(msg.proc)

	case installOutputMsg:
		m.installOutput.WriteString(msg.line)
		m.installView.SetContent(m.installOutput.String())
		m.installView.GotoBottom()

		return m, readNextLine(msg.proc)

	case installDoneMsg:
		if msg.err != nil {
			m.err = msg.err

			m.installOutput.WriteString(fmt.Sprintf("\n\nError: %v", msg.err))

			return m, tea.Tick(5*time.Second, func(_ time.Time) tea.Msg { return delayedSuccessMsg{} })
		}
		m.statusMsg = "Project Created Successfully!"
		m.installOutput.WriteString("\n\n[SUCCESS] Installation Completed.\nWaiting 3 seconds...")
		m.installView.SetContent(m.installOutput.String())
		m.installView.GotoBottom()
		return m, tea.Tick(3*time.Second, func(_ time.Time) tea.Msg { return delayedSuccessMsg{} })

	case delayedSuccessMsg:
		m.state = StateSuccess
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.WindowSizeMsg:

		h, v := AppBorderStyle.GetFrameSize()

		innerW := msg.Width - h - 2
		innerH := msg.Height - v

		m.menuList.SetSize(innerW, innerH-4)
		m.projectList.SetSize(innerW, innerH-4)
		m.templateList.SetSize(innerW, innerH-4)
		m.historyList.SetSize(innerW, innerH-4)

		m.venvModel, _ = m.venvModel.Update(tea.WindowSizeMsg{Width: innerW, Height: innerH})
		m.boilerplateModel, _ = m.boilerplateModel.Update(tea.WindowSizeMsg{Width: innerW, Height: innerH})
		devModel, _ := m.devServerModel.Update(tea.WindowSizeMsg{Width: innerW, Height: innerH})
		m.devServerModel = devModel.(DevServerDashboardModel)
		m.bonusModel, _ = m.bonusModel.Update(msg)

		m.width = msg.Width
		m.height = msg.Height

		m.installView.Width = innerW

		m.installView.Height = innerH - 3

		m.helpView.Width = innerW
		m.helpView.Height = innerH - 3
	}

	return m, nil
}

func (m ProjectDashboardModel) View() string {

	h, v := AppBorderStyle.GetFrameSize()

	contentWidth := m.width - h - 2
	contentHeight := m.height - v

	var innerContent string

	switch m.state {
	case StateMenu:

		header := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center).Render(
			titleStyle.Render("Project Tools"),
		)

		footer := lipgloss.NewStyle().Align(lipgloss.Center).Width(contentWidth).Render(
			subtleStyle.Render("Use ↑/↓ to Navigate • Enter to Select • ? Help • Q to Quit"),
		)

		innerContent = lipgloss.JoinVertical(lipgloss.Left,
			header,
			"\n",
			m.menuList.View(),
			"\n",
			footer,
		)

	case StateDevServer:

		innerContent = m.devServerModel.View()

	case StateVenvWizard:

		innerContent = m.venvModel.View()

	case StateBoilerplate:
		innerContent = m.boilerplateModel.View()

	case StateBonus:
		innerContent = m.bonusModel.View()

	case StateCreating:

		header := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center).Render(
			titleStyle.Render("Project Creation & Management"),
		)
		innerContent = docStyle.Render(lipgloss.JoinVertical(lipgloss.Left, header, m.installView.View()))

	case StateSuccess:
		title := lipgloss.NewStyle().Foreground(colorGreen).Bold(true).Render(" PROJECT CREATED ")
		msg := fmt.Sprintf("Your project is ready at:\n%s\n\n(Press Enter to Exit)", m.pathInput.Value())

		content := lipgloss.JoinVertical(lipgloss.Center, title, "\n", msg)
		innerContent = lipgloss.Place(contentWidth, contentHeight, lipgloss.Center, lipgloss.Center,
			successBoxStyle.Render(content),
		)

	case StateNameProject, StateSelectPath, StateBackupInput:

		var title, inputView, footer string

		switch m.state {
		case StateNameProject:
			title = "Step 1: Project Name"
			inputView = m.input.View()
			footer = "(Enter to Next, Esc to Back)"
		case StateSelectPath:
			title = "Step 2: Project Path"
			inputView = m.pathInput.View()
			footer = "(Enter to Create, Esc to Back)"
		case StateBackupInput:
			title = "Backup Project"
			inputView = m.pathInput.View()
			footer = "(Enter Path to Backup, Esc to Cancel)"
		}

		content := lipgloss.JoinVertical(lipgloss.Center,
			titleStyle.Render(title),
			"\n",
			focusedInputBoxStyle.Render(inputView),
			"\n",
			subtleStyle.Render(footer),
		)

		innerContent = lipgloss.Place(contentWidth, contentHeight, lipgloss.Center, lipgloss.Center, content)

	case StateSelectTemplate:
		header := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center).Render(
			titleStyle.Render("Select Project Template"),
		)
		innerContent = docStyle.Render(
			lipgloss.JoinVertical(lipgloss.Left,
				header,
				m.templateList.View(),
			),
		)

	case StateCleanupPrompt:
		content := lipgloss.JoinVertical(lipgloss.Center,
			titleStyle.Render("Cleanup Old Logs?"),
			"\n",
			"We found project logs older than 30 days.",
			"Do you want to delete them from history?",
			"\n",
			subtleStyle.Render("[Enter] Yes, Delete • [Esc] Keep"),
		)
		innerContent = lipgloss.Place(contentWidth, contentHeight, lipgloss.Center, lipgloss.Center, content)

	case StateConfirmDelete:

		content := lipgloss.JoinVertical(lipgloss.Center,
			titleStyle.Render("Confirm Deletion"),
			"\n",
			"Are you sure you want to delete this entry?",
			"\n",
			subtleStyle.Render("[Enter] Yes, Delete • [Esc] Cancel"),
		)
		innerContent = lipgloss.Place(contentWidth, contentHeight, lipgloss.Center, lipgloss.Center, content)

	case StateHistoryList:
		header := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center).Render(
			titleStyle.Render("Project History"),
		)
		listContent := m.historyList.View()
		footer := subtleStyle.Render("\n [d] Delete Entry • [?] Help • [Esc] Back")

		innerContent = docStyle.Render(lipgloss.JoinVertical(lipgloss.Left, header, listContent, footer))

	case StateProjectHelp:

		innerContent = lipgloss.Place(contentWidth, contentHeight, lipgloss.Center, lipgloss.Center,
			lipgloss.JoinVertical(lipgloss.Center,
				lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).MarginBottom(1).Render("Project Tools Help"),
				m.helpView.View(),
				lipgloss.NewStyle().Foreground(lipgloss.Color("240")).MarginTop(1).Render("Press [Esc] or [?] to go back"),
			),
		)

	default:

		listContent := m.projectList.View()
		footer := subtleStyle.Render("\n [Enter] Select • [b] Backup Project • [?] Help • [Esc] Back")
		innerContent = docStyle.Render(lipgloss.JoinVertical(lipgloss.Left, listContent, footer))
	}
	return innerContent
}

func RunProjectDashboard() {
	m := NewProjectDashboardModel()
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error:", err)
	}
}
