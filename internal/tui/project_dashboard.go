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
	"github.com/charmbracelet/lipgloss"
	"github.com/phravins/devcli/internal/history"
	"github.com/phravins/devcli/internal/project"
	"github.com/phravins/devcli/internal/templates"
)

type ProjectDashboardModel struct {
	menuList     list.Model // Top Level Menu
	projectList  list.Model // Project List (Sub Menu)
	templateList list.Model // Wizard Step 1
	input        textinput.Model
	pathInput    textinput.Model // New Input for Path
	spinner      spinner.Model
	historyList  list.Model // New History List

	// State
	state         int
	width, height int

	// Data
	manager *project.Manager

	// Sub-Models
	venvModel        VenvDashboardModel // Embedded Venv Model
	devServerModel   DevServerDashboardModel
	boilerplateModel BoilerplateDashboardModel
	bonusModel       BonusDashboardModel

	selectedTpl string
	err         error
	statusMsg   string

	// Installation Logging
	installOutput *strings.Builder
	installView   viewport.Model
	helpView      viewport.Model // New
}

const (
	StateMenu           = iota // Top level: "Project Creation & Management", etc.
	StateProjectList           // Spec: "My Projects" list with "+ New Project"
	StateSelectTemplate        // Wizard Step 1
	StateNameProject           // Wizard Step 2
	StateSelectPath            // New State
	StateCreating              // Wizard Step 3 (Processing)
	StateSuccess               // Completion Screen
	StateBackupInput           // New Backup State
	StateCleanupPrompt         // New: Ask to delete old logs
	StateHistoryList           // New: View History
	StateConfirmDelete         // New: Confirm Deletion
	StateProjectHelp           // Help screen

	StateVenvWizard  // Sub-feature 2 (Delegated to venvModel)
	StateDevServer   // Sub-feature 3 (Dev Server Launcher)
	StateBoilerplate // Sub-feature 4 (Boilerplate Generator)
	StateBonus       // Sub-feature 5 (Bonus Features)
)

func NewProjectDashboardModel() ProjectDashboardModel {
	mgr := project.NewManager("")

	// 1. Top Level Menu
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

	// 2. Project List (Sub-feature)
	items := loadProjects(mgr.Workspace)
	items = append([]list.Item{item{title: "+ New Project", desc: "Create a new project from template"}}, items...)
	pl := list.New(items, list.NewDefaultDelegate(), 0, 0)
	pl.Title = "My Projects"
	pl.SetShowHelp(false)

	// 3. Template List (Wizard)
	var tplItems []list.Item
	for _, t := range templates.List() {
		tplItems = append(tplItems, item{title: t.Name, desc: t.Description})
	}
	tplList := list.New(tplItems, list.NewDefaultDelegate(), 0, 0)
	tplList.Title = "Select a Template"
	tplList.Title = "Select Project Template (v2)"
	tplList.SetShowHelp(false)

	// 4. History List
	histList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	histList.Title = "Project History"
	histList.SetShowHelp(false)
	histList.SetShowTitle(false)

	// Input
	ti := textinput.New()
	ti.Placeholder = "Project Name"
	ti.CharLimit = 50
	ti.Width = 40

	// Path Input
	pi := textinput.New()
	pi.Placeholder = "Parent Directory (e.g. C:\\Projects or ~)"
	// Default to current Workspace
	pi.SetValue(mgr.Workspace)
	pi.CharLimit = 100
	pi.Width = 50

	// Spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	// Viewport for logs
	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("62")) // Purple border

	// Help Viewport
	hv := viewport.New(80, 20)
	hv.Style = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("62"))
	hv.SetContent(ProjectToolsHelp)

	return ProjectDashboardModel{
		menuList:         menu,
		projectList:      pl,
		templateList:     tplList,
		historyList:      histList,
		input:            ti,
		pathInput:        pi, // Add to struct
		spinner:          s,
		manager:          mgr,
		venvModel:        NewVenvDashboardModel(),                     // Init Venv Model
		devServerModel:   NewDevServerDashboardModel(mgr.Workspace),   // Init Dev Server Model
		boilerplateModel: NewBoilerplateDashboardModel(mgr.Workspace), // Init Boilerplate Model
		bonusModel:       NewBonusDashboardModel(mgr.Workspace),       // Init Bonus Model
		state:            StateMenu,                                   // Start at Top Level
		installOutput:    &strings.Builder{},
		installView:      vp,
		helpView:         hv,
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
			// Smart Filtering: Only list if it looks like a project
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

// isProject checks if a directory contains common project markers
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
	// Check for old history on startup
	old := history.GetOldEntries(30)
	if len(old) > 0 {
		return tea.Batch(
			func() tea.Msg { return cleanupPromptMsg{} },
			m.spinner.Tick,
			m.venvModel.Init(),
		)
	}
	return tea.Batch(m.spinner.Tick, m.venvModel.Init(), m.boilerplateModel.Init())
}

type cleanupPromptMsg struct{}

type projectCreatedMsg struct {
	installCmd string
	path       string
	err        error
}

type delayedSuccessMsg struct{}

type installDoneMsg struct{ err error }

// Actual implementation using "Next Line" command pattern
type cmdProcess struct {
	cmd    *exec.Cmd
	reader *bufio.Reader
}

type installStartedMsg struct {
	proc *cmdProcess
}
type installOutputMsg struct {
	line string
	proc *cmdProcess
}

func createProjectCmd(mgr *project.Manager, name, stack, path string) tea.Cmd {
	return func() tea.Msg {
		// Step 1: Generate Files (Fast)
		cmdStr, resolvedPath, err := mgr.CreateProject(name, stack, path)
		return projectCreatedMsg{installCmd: cmdStr, path: resolvedPath, err: err}
	}
}

func startInstallCmd(dir, cmdStr string) tea.Cmd {
	return func() tea.Msg {
		// Use explicit echoes to force output
		// We use 'call' to ensure batch files work.
		var c *exec.Cmd
		if runtime.GOOS == "windows" {
			fullCmd := fmt.Sprintf("@echo on & echo [DevCLI] Starting installation process... & echo [DevCLI] Directory: %s & echo [DevCLI] Running: %s & echo ---------------------------------------- & call %s & echo. & echo ---------------------------------------- & echo [DevCLI] Process Completed.", dir, cmdStr, cmdStr)
			c = exec.Command("cmd", "/c", fullCmd)
		} else {
			// Unix/Linux/Mac Buffer-friendly command chain
			fullCmd := fmt.Sprintf("echo '[DevCLI] Starting installation process...' && echo '[DevCLI] Directory: %s' && echo '[DevCLI] Running: %s' && echo '----------------------------------------' && %s && echo '' && echo '----------------------------------------' && echo '[DevCLI] Process Completed.'", dir, cmdStr, cmdStr)
			c = exec.Command("sh", "-c", fullCmd)
		}
		c.Dir = dir

		outPipe, _ := c.StdoutPipe()
		c.Stderr = c.Stdout // Merge stderr

		if err := c.Start(); err != nil {
			return installDoneMsg{err: err}
		}

		return installStartedMsg{
			proc: &cmdProcess{
				cmd:    c,
				reader: bufio.NewReader(outPipe),
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
			proc.cmd.Wait() // Cleanup
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

	// --- Navigation Messages ---
	switch msg.(type) {
	case VenvBackMsg, DevServerBackMsg, BoilerplateBackMsg, BonusBackMsg:
		m.state = StateMenu
		return m, nil
	case BackMsg:
		m.state = StateMenu
		return m, nil
	}

	// --- State Delegation ---
	// Check delegation first to ensure sub-models receive all events (including Mouse & Keys)

	if m.state == StateDevServer {
		var devCmd tea.Cmd
		// Intercept WindowSizeMsg to pass inner dimensions
		if wMsg, ok := msg.(tea.WindowSizeMsg); ok {
			h, v := AppBorderStyle.GetFrameSize()
			innerMsg := tea.WindowSizeMsg{Width: wMsg.Width - h, Height: wMsg.Height - v}
			m.devServerModel, devCmd = m.devServerModel.Update(innerMsg)
		} else {
			m.devServerModel, devCmd = m.devServerModel.Update(msg)
		}
		return m, devCmd
	}

	if m.state == StateVenvWizard {
		var venvCmd tea.Cmd
		// Intercept WindowSizeMsg to pass inner dimensions
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
		// Intercept WindowSizeMsg to pass inner dimensions
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
		// Intercept WindowSizeMsg to pass inner dimensions
		if wMsg, ok := msg.(tea.WindowSizeMsg); ok {
			h, v := AppBorderStyle.GetFrameSize()
			innerMsg := tea.WindowSizeMsg{Width: wMsg.Width - h, Height: wMsg.Height - v}
			m.bonusModel, bonusCmd = m.bonusModel.Update(innerMsg)
		} else {
			m.bonusModel, bonusCmd = m.bonusModel.Update(msg)
		}
		return m, bonusCmd
	}

	// --- Main Logic ---

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.state == StateCreating {
			return m, nil // Block input while creating
		}

		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}

	case tea.MouseMsg:
		// Manual scroll handling for reliability
		if msg.Type == tea.MouseWheelUp {
			m.menuList.CursorUp()
			return m, nil
		}
		if msg.Type == tea.MouseWheelDown {
			m.menuList.CursorDown()
			return m, nil
		}
		var cmd tea.Cmd

		switch m.state {
		case StateMenu:
			m.menuList, cmd = m.menuList.Update(msg)
		case StateProjectList:
			m.projectList, cmd = m.projectList.Update(msg)
		case StateSelectTemplate:
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
		// --- State Machine ---

		switch m.state {
		case StateSuccess:
			switch msg.String() {
			case "enter", "esc":
				m.state = StateProjectList
				return m, nil
			}

		case StateMenu:
			switch msg.String() {
			case "enter":
				i, ok := m.menuList.SelectedItem().(item)
				if ok {
					if i.title == "Project Creation & Management" {
						m.state = StateProjectList
						// Refresh projects?
						items := loadProjects(m.manager.Workspace)
						items = append([]list.Item{item{title: "+ New Project", desc: "Create a new project from template"}}, items...)
						m.projectList.SetItems(items)
						return m, nil
					}
					if i.title == "Virtual Environment Wizard" {
						m.state = StateVenvWizard
						m.venvModel = NewVenvDashboardModel()
						// Initialize with current dimensions
						h, v := AppBorderStyle.GetFrameSize()
						innerW := m.width - h - 2
						innerH := m.height - v
						m.venvModel, _ = m.venvModel.Update(tea.WindowSizeMsg{Width: innerW, Height: innerH})
						return m, m.venvModel.Init()
					}
					if i.title == "Dev Server" {
						m.state = StateDevServer
						m.devServerModel = NewDevServerDashboardModel(m.manager.Workspace)
						// Initialize with current dimensions to ensure correct layout (centering)
						h, v := AppBorderStyle.GetFrameSize()
						innerW := m.width - h - 2
						innerH := m.height - v
						m.devServerModel, _ = m.devServerModel.Update(tea.WindowSizeMsg{Width: innerW, Height: innerH})
						return m, m.devServerModel.Init()
					}
					if i.title == "Boilerplate Generator" {
						m.state = StateBoilerplate
						m.boilerplateModel = NewBoilerplateDashboardModel(m.manager.Workspace)
						// Pass dimensions
						h, v := AppBorderStyle.GetFrameSize()
						innerW := m.width - h - 2
						innerH := m.height - v
						m.boilerplateModel.resizeLists(innerW, innerH)
						return m, nil
					}
					if i.title == "Bonus Features" {
						m.state = StateBonus
						m.bonusModel = NewBonusDashboardModel(m.manager.Workspace)
						// Initialize with current dimensions
						h, v := AppBorderStyle.GetFrameSize()
						innerW := m.width - h - 2
						innerH := m.height - v
						m.bonusModel, _ = m.bonusModel.Update(tea.WindowSizeMsg{Width: innerW, Height: innerH})
						return m, m.bonusModel.Init()
					}
					if i.title == "Project History" {
						m.state = StateHistoryList
						// Load History items
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
				// Delete
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
				// Confirm Delete
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
				// Delete selected
				idx := m.historyList.Index()
				if idx >= 0 && len(m.historyList.Items()) > 0 {
					history.DeleteOne(idx)
					// Reload
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
				m.state = StateProjectList
				return m, nil
			}
			var cmd tea.Cmd
			m.helpView, cmd = m.helpView.Update(msg)
			return m, cmd

		case StateProjectList:
			switch msg.String() {
			case "?":
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
			case "b": // Backup
				if len(m.projectList.Items()) > 1 { // Assuming + New Project is item 0
					// Check if valid project selected
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
				// Back to Top Menu
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
					// Perform Backup
					i, _ := m.projectList.SelectedItem().(item)
					projectName := i.title
					srcPath := filepath.Join(m.manager.Workspace, projectName)

					m.state = StateCreating // Reuse creating screen for logs
					m.statusMsg = "Backing up project..."
					m.installOutput.Reset()
					m.installOutput.WriteString(fmt.Sprintf("Backing up '%s' to '%s'...\n", srcPath, dest))
					m.installView.SetContent(m.installOutput.String())

					// Run in goroutine/cmd
					return m, func() tea.Msg {
						err := m.manager.BackupProject(srcPath, dest)
						if err != nil {
							return installDoneMsg{err: err}
						}
						// Artificial delay
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
					// Smart Naming
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
					// Go to Next Step: Path Selection
					m.state = StateSelectPath
					// Ensure path input has latest workspace if they haven't edited it?
					// Or keep sticky. Let's keep sticky or default.
					m.pathInput.Focus()
					// Start blinking cursor
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
				// Validate Path
				pathVal := m.pathInput.Value()
				_, err := m.manager.ValidateParentDir(pathVal)
				if err != nil {
					m.pathInput.SetValue(pathVal)                                                 // Keep value
					m.pathInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // Red
					m.err = err
					return m, nil
				}
				m.err = nil                                 // Clear error
				m.pathInput.TextStyle = lipgloss.NewStyle() // Reset style

				// Create!
				m.state = StateCreating
				m.statusMsg = "Initializing Project..."
				m.installOutput.Reset()

				// Customizable Log Header
				timestamp := time.Now().Format("2006-01-02 15:04:05")
				header := fmt.Sprintf("PROJECT CREATION LOG\n========================\nName : %s\nPath : %s\nTime : %s\n========================\n\n", m.input.Value(), pathVal, timestamp)
				m.installOutput.WriteString(header)
				m.installOutput.WriteString("Starting Project Generation...\n")
				m.installView.SetContent(m.installOutput.String())

				// Record History
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
		// Append to existing logs
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
		// Start reading loop
		return m, readNextLine(msg.proc)

	case installOutputMsg:
		m.installOutput.WriteString(msg.line)
		m.installView.SetContent(m.installOutput.String())
		m.installView.GotoBottom()
		// Chain next read using the process reference passed from previous msg
		return m, readNextLine(msg.proc)

	case installDoneMsg:
		if msg.err != nil {
			m.err = msg.err
			// Don't fail completely, just show error?
			m.installOutput.WriteString(fmt.Sprintf("\n\nError: %v", msg.err))
			// Wait a bit so they see it?
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
		// Use AppBorderStyle to determine available inner space
		h, v := AppBorderStyle.GetFrameSize()
		// Subtract 2 explicitly to match View() and avoid overflow
		innerW := msg.Width - h - 2
		innerH := msg.Height - v

		// Resize Lists with appropriate offsets for headers/footers
		m.menuList.SetSize(innerW, innerH-14)    // Reserve space for Big Header + Spacing
		m.projectList.SetSize(innerW, innerH-4)  // Reserve space for Footer
		m.templateList.SetSize(innerW, innerH-4) // Reserve space for Header
		m.historyList.SetSize(innerW, innerH-4)  // Reserve space for Footer

		// Also update venvModel so it's consistent if we switch to it
		m.venvModel, _ = m.venvModel.Update(tea.WindowSizeMsg{Width: innerW, Height: innerH})
		m.boilerplateModel, _ = m.boilerplateModel.Update(tea.WindowSizeMsg{Width: innerW, Height: innerH})
		m.devServerModel, _ = m.devServerModel.Update(tea.WindowSizeMsg{Width: innerW, Height: innerH})
		m.bonusModel, _ = m.bonusModel.Update(tea.WindowSizeMsg{Width: innerW, Height: innerH})

		m.width = msg.Width
		m.height = msg.Height
		// Resize viewport
		m.installView.Width = innerW
		// Calculate available height: Total - Header (~3 lines) - Padding
		m.installView.Height = innerH - 3

		// Resize Help View
		m.helpView.Width = innerW
		m.helpView.Height = innerH - 3
	}

	return m, nil
}

func (m ProjectDashboardModel) View() string {
	// Calculate content size inside the global border
	// Border(1) + Padding(1) on each side -> 2 chars horizontal per side = 4 chars total?
	// Vertical: Border(1) + Padding(1) top/bottom = ?

	// Lipgloss GetFrameSize return horizontal, vertical usage.
	h, v := AppBorderStyle.GetFrameSize()

	// Subtract 2 explicitly to avoid Windows auto-wrapping issues at the exact edge
	contentWidth := m.width - h - 2
	contentHeight := m.height - v

	var innerContent string

	switch m.state {
	case StateMenu:
		// Standard List View (Like Dashboard)
		// We want the header centered, then the list below.

		// Header
		header := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center).Render(
			titleStyle.Render("Project Tools"),
		)

		footer := lipgloss.NewStyle().Align(lipgloss.Center).Width(contentWidth).Render(
			subtleStyle.Render("Use ↑/↓ to Navigate • Enter to Select • Q to Quit"),
		)

		// Join vertically
		innerContent = lipgloss.JoinVertical(lipgloss.Left,
			header,
			"\n",
			m.menuList.View(),
			"\n",
			footer,
		)

	case StateDevServer:
		// Dev Server dashboard
		innerContent = m.devServerModel.View()

	case StateVenvWizard:
		// Venv Wizard is already resized to fit inner content in Update()
		// So we just render it. It will be wrapped by valid border below.
		innerContent = m.venvModel.View()

	case StateBoilerplate:
		innerContent = m.boilerplateModel.View()

	case StateBonus:
		innerContent = m.bonusModel.View()

	case StateCreating:
		// Full Screen Installation View
		// Use almost full screen for logs, with a styled header
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
		// Centered Card Layout for Inputs
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

		// Calculate vertical center
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
		// Confirmation Dialog
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

		// Align with other list views style if needed, or simple render
		innerContent = docStyle.Render(lipgloss.JoinVertical(lipgloss.Left, header, listContent, footer))

	case StateProjectHelp:
		// Render help content
		innerContent = lipgloss.Place(contentWidth, contentHeight, lipgloss.Center, lipgloss.Center, m.helpView.View())

	default:
		// Default List View (Select Template)
		listContent := m.projectList.View()
		footer := subtleStyle.Render("\n [Enter] Select • [b] Backup Project • [?] Help • [Esc] Back")
		innerContent = docStyle.Render(lipgloss.JoinVertical(lipgloss.Left, listContent, footer))
	}

	// WRAP EVERYTHING IN GLOBAL BORDER
	// Removed AppBorderStyle as requested
	return innerContent
}

func RunProjectDashboard() {
	m := NewProjectDashboardModel()
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error:", err)
	}
}
