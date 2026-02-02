package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/phravins/devcli/internal/devserver"
)

type DevServerDashboardModel struct {
	width, height       int
	state               int
	projectPath         string
	projectInfo         devserver.ProjectInfo
	runner              *devserver.Runner
	logView             viewport.Model
	helpView            viewport.Model // New: scrollable help
	searchInput         textinput.Model
	pathInput           textinput.Model // New: for customizable path
	logs                []logEntry
	filterMode          string // "all", "errors", "warnings"
	serverFilter        string // "all", "backend", "frontend"
	autoScroll          bool
	showHelp            bool
	err                 error
	pendingAction       string // Stores the action waiting for confirmation
	confirmationMessage string // Message to display in confirmation dialog
}

type logEntry struct {
	timestamp  string
	serverName string
	line       string
	isError    bool
	isWarning  bool
}

const (
	StateDevServerPathInput = iota // New: Path selection
	StateDevServerDetecting
	StateDevServerReady
	StateDevServerRunning
	StateDevServerConfirmation // Confirmation dialog state
	StateDevServerStopping     // Server stopping state
	StateDevServerHelp
)

type detectDoneMsg struct {
	info devserver.ProjectInfo
	err  error
}

type logReceivedMsg struct {
	log devserver.LogLine
}

type serverStoppedMsg struct{}

func NewDevServerDashboardModel(projectPath string) DevServerDashboardModel {
	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62"))

	ti := textinput.New()
	ti.Placeholder = "Search logs..."
	ti.Width = 30

	// Path input for customizable detection
	pi := textinput.New()
	pi.Placeholder = "Type path here or press Enter for current"
	pi.SetValue(projectPath)
	pi.CharLimit = 200
	pi.Width = 62
	pi.Width = 62
	pi.Focus() // Focus the input immediately

	// Initialize help viewport
	hv := viewport.New(80, 20)
	hv.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#0F9E99")).
		Padding(1, 2)
	hv.SetContent(DevServerHelp)

	return DevServerDashboardModel{
		state:        StateDevServerPathInput, // Start with path input
		projectPath:  projectPath,
		logView:      vp,
		helpView:     hv,
		searchInput:  ti,
		pathInput:    pi,
		logs:         make([]logEntry, 0),
		filterMode:   "all",
		serverFilter: "all",
		autoScroll:   true,
		showHelp:     false,
	}
}

func (m DevServerDashboardModel) Init() tea.Cmd {
	return textinput.Blink
}

func detectProjectCmd(path string) tea.Cmd {
	return func() tea.Msg {
		info := devserver.Detect(path)
		if info.Type == devserver.TypeUnknown {
			return detectDoneMsg{
				info: info,
				err:  fmt.Errorf("unable to detect project type"),
			}
		}
		return detectDoneMsg{info: info, err: nil}
	}
}
func stopServerCmd(runner *devserver.Runner) tea.Cmd {
	return func() tea.Msg {
		runner.Stop()
		return serverStoppedMsg{}
	}
}
func waitForLogCmd(runner *devserver.Runner) tea.Cmd {
	return func() tea.Msg {
		logChan := runner.GetLogChannel()
		select {
		case log, ok := <-logChan:
			if !ok {
				// Channel closed, server stopped
				return nil
			}
			return logReceivedMsg{log: log}
		case <-time.After(100 * time.Millisecond):
			// Timeout - return a tick message to check again
			return tickMsg{}
		}
	}
}

type tickMsg struct{}

func (m DevServerDashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.showHelp {
			switch msg.String() {
			case "esc", "?":
				m.showHelp = false
				return m, nil
			default:
				var cmd tea.Cmd
				m.helpView, cmd = m.helpView.Update(msg)
				return m, cmd
			}
		}

		// Handle path input state
		if m.state == StateDevServerPathInput {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "esc":
				return m, func() tea.Msg { return DevServerBackMsg{} }
			case "enter":
				// Use the entered path or default
				path := m.pathInput.Value()
				if path == "" {
					path = m.projectPath
				}
				m.projectPath = path
				m.state = StateDevServerDetecting
				return m, detectProjectCmd(path)
			}
			m.pathInput, cmd = m.pathInput.Update(msg)
			return m, cmd
		}

		// Handle search input when focused - but allow Esc to unfocus
		if m.searchInput.Focused() {
			switch msg.String() {
			case "esc":
				m.searchInput.Blur()
				return m, nil
			case "enter":
				m.searchInput.Blur()
				m.updateLogView()
				return m, nil
			default:
				m.searchInput, cmd = m.searchInput.Update(msg)
				return m, cmd
			}
		}

		// Handle confirmation state - MUST be checked first to prevent other handlers from consuming keys
		if m.state == StateDevServerConfirmation {
			switch msg.String() {
			case "y", "Y":
				// User confirmed - execute the pending action
				return m.executePendingAction()
			case "n", "N", "esc":
				// User cancelled - return to running state
				m.state = StateDevServerRunning
				m.pendingAction = ""
				m.confirmationMessage = ""
				// Continue waiting for logs
				if m.runner != nil {
					return m, waitForLogCmd(m.runner)
				}
				return m, nil
			default:
				return m, nil
			}
		}

		// Handle main keyboard shortcuts
		switch msg.String() {
		case "ctrl+c", "q":
			if m.runner != nil {
				m.runner.Stop()
			}
			return m, tea.Quit
		case "esc":
			if m.state == StateDevServerRunning && m.runner != nil {
				// Ask for confirmation before stopping and going back
				m.state = StateDevServerConfirmation
				m.pendingAction = "back"
				m.confirmationMessage = "Stop the server and go back?"
				m.confirmationMessage = "Stop the server and go back?"
				return m, nil
			} else {
				return m, func() tea.Msg { return DevServerBackMsg{} }
			}
		case "?":
			if m.state == StateDevServerRunning && m.runner != nil {
				// Ask for confirmation before showing help
				m.state = StateDevServerConfirmation
				m.pendingAction = "help"
				m.confirmationMessage = "Show help?"
				m.confirmationMessage = "Show help?"
				return m, nil
			}
			m.showHelp = !m.showHelp
			return m, nil
		case "s":
			if m.state == StateDevServerReady {
				m.runner = devserver.NewRunner()
				if err := m.runner.Start(m.projectInfo); err != nil {
					m.err = err
					return m, nil
				} else {
					m.state = StateDevServerRunning
					return m, waitForLogCmd(m.runner)
				}
			} else if m.state == StateDevServerRunning && m.runner != nil {
				// Ask for confirmation before stopping
				m.state = StateDevServerConfirmation
				m.pendingAction = "stop"
				m.confirmationMessage = "Stop the server?"
				m.pendingAction = "stop"
				m.confirmationMessage = "Stop the server?"
				return m, nil
			}
			return m, nil
		case "f":
			if m.state == StateDevServerRunning && m.runner != nil {
				// Ask for confirmation before changing filter
				m.state = StateDevServerConfirmation
				m.pendingAction = "filter"
				m.confirmationMessage = "Change filter mode?"
				m.pendingAction = "filter"
				m.confirmationMessage = "Change filter mode?"
				return m, nil
			}
			return m, nil
		case "b":
			if m.state == StateDevServerRunning && m.runner != nil {
				// Ask for confirmation before changing server filter
				if m.projectInfo.Type == devserver.TypeFullstack {
					m.state = StateDevServerConfirmation
					m.pendingAction = "source"
					m.confirmationMessage = "Change server source filter?"
					m.pendingAction = "source"
					m.confirmationMessage = "Change server source filter?"
					return m, nil
				}
			}
			return m, nil
		case "a":
			if m.state == StateDevServerRunning && m.runner != nil {
				// Ask for confirmation before toggling auto-scroll
				m.state = StateDevServerConfirmation
				m.pendingAction = "autoscroll"
				m.confirmationMessage = "Toggle auto-scroll?"
				m.pendingAction = "autoscroll"
				m.confirmationMessage = "Toggle auto-scroll?"
				return m, nil
			}
			return m, nil
		case "c":
			if m.state == StateDevServerRunning && m.runner != nil {
				// Ask for confirmation before clearing logs
				m.state = StateDevServerConfirmation
				m.pendingAction = "clear"
				m.confirmationMessage = "Clear all logs?"
				m.pendingAction = "clear"
				m.confirmationMessage = "Clear all logs?"
				return m, nil
			}
			return m, nil
		case "/":
			if m.state == StateDevServerRunning && m.runner != nil {
				// Ask for confirmation before opening search
				m.state = StateDevServerConfirmation
				m.pendingAction = "search"
				m.confirmationMessage = "Open search?"
				m.pendingAction = "search"
				m.confirmationMessage = "Open search?"
				return m, nil
			}
			return m, nil
		case "up", "down", "pgup", "pgdown", "home", "end":
			// These keys are for viewport scrolling only when running
			if m.state == StateDevServerRunning && m.runner != nil {
				m.logView, cmd = m.logView.Update(msg)
				return m, cmd
			}
			return m, nil
		}

	case detectDoneMsg:
		m.projectInfo = msg.info
		m.err = msg.err
		if msg.err == nil {
			m.state = StateDevServerReady
		}

	case serverStoppedMsg:
		m.state = StateDevServerReady
		m.runner = nil
		return m, nil

	case logReceivedMsg:
		timestamp := time.Now().Format("15:04:05")
		isWarning := strings.Contains(strings.ToLower(msg.log.Line), "warn")

		m.logs = append(m.logs, logEntry{
			timestamp:  timestamp,
			serverName: msg.log.ServerName,
			line:       msg.log.Line,
			isError:    msg.log.IsError,
			isWarning:  isWarning,
		})

		m.updateLogView()
		if m.autoScroll {
			m.logView.GotoBottom()
		}

		// Only continue waiting if runner is still valid and server is running/stopping/confirming
		if (m.state == StateDevServerRunning || m.state == StateDevServerConfirmation || m.state == StateDevServerStopping) && m.runner != nil {
			return m, waitForLogCmd(m.runner)
		}
		return m, nil

	case tickMsg:
		// Timeout occurred while waiting for logs, continue waiting if still active
		if (m.state == StateDevServerRunning || m.state == StateDevServerConfirmation || m.state == StateDevServerStopping) && m.runner != nil {
			return m, waitForLogCmd(m.runner)
		}
		return m, nil

	case tea.MouseMsg:
		if m.showHelp {
			var cmd tea.Cmd
			m.helpView, cmd = m.helpView.Update(msg)
			return m, cmd
		}
		// Pass to log view mainly, or list logic if we implemented a list
		var cmd tea.Cmd
		m.logView, cmd = m.logView.Update(msg)
		return m, cmd

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		m.logView.Width = msg.Width - 4    // Full width minus small padding
		m.logView.Height = msg.Height - 14 // Increased padding for header

		// Resize help view
		m.helpView.Width = msg.Width - 8
		m.helpView.Height = msg.Height - 4
	}

	// For non-key messages (like mouse events), pass to viewport
	if _, ok := msg.(tea.KeyMsg); !ok {
		m.logView, cmd = m.logView.Update(msg)
		return m, cmd
	}

	return m, nil

}

func (m *DevServerDashboardModel) updateLogView() {
	var content strings.Builder
	searchTerm := strings.ToLower(m.searchInput.Value())

	for _, log := range m.logs {
		// Apply filters
		if m.filterMode == "errors" && !log.isError {
			continue
		}
		if m.filterMode == "warnings" && !log.isWarning {
			continue
		}

		// Apply server filter
		if m.serverFilter != "all" {
			if m.serverFilter == "backend" && !strings.Contains(strings.ToLower(log.serverName), "backend") {
				continue
			}
			if m.serverFilter == "frontend" && !strings.Contains(strings.ToLower(log.serverName), "frontend") {
				continue
			}
		}

		// Apply search filter
		if searchTerm != "" && !strings.Contains(strings.ToLower(log.line), searchTerm) {
			continue
		}

		// Format log line
		var lineStyle lipgloss.Style
		if log.isError {
			lineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // Red
		} else if log.isWarning {
			lineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("226")) // Yellow
		} else {
			lineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("255")) // White
		}

		serverStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("141")).Bold(true) // Purple
		timeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))              // Gray

		formattedLine := fmt.Sprintf("%s [%s] %s\n",
			timeStyle.Render(log.timestamp),
			serverStyle.Render(log.serverName),
			lineStyle.Render(log.line),
		)

		// Highlight search term
		if searchTerm != "" {
			highlightStyle := lipgloss.NewStyle().Background(lipgloss.Color("226")).Foreground(lipgloss.Color("0"))
			formattedLine = strings.ReplaceAll(formattedLine, searchTerm, highlightStyle.Render(searchTerm))
		}

		content.WriteString(formattedLine)
	}

	m.logView.SetContent(content.String())
}

// executePendingAction executes the action that was confirmed by the user
func (m DevServerDashboardModel) executePendingAction() (DevServerDashboardModel, tea.Cmd) {
	// Store and clear confirmation state
	action := m.pendingAction
	m.pendingAction = ""
	m.confirmationMessage = ""
	// Don't set state here - let each action handle its own state transition

	// Execute the action
	switch action {
	case "stop":
		// Stop the server asynchronously
		m.state = StateDevServerStopping
		return m, stopServerCmd(m.runner)

	case "filter":
		// Cycle through filter modes
		m.state = StateDevServerRunning
		switch m.filterMode {
		case "all":
			m.filterMode = "errors"
		case "errors":
			m.filterMode = "warnings"
		case "warnings":
			m.filterMode = "all"
		}
		m.updateLogView()
		return m, nil

	case "source":
		// Toggle server filter (for fullstack)
		m.state = StateDevServerRunning
		if m.projectInfo.Type == devserver.TypeFullstack {
			switch m.serverFilter {
			case "all":
				m.serverFilter = "backend"
			case "backend":
				m.serverFilter = "frontend"
			case "frontend":
				m.serverFilter = "all"
			}
			m.updateLogView()
		}
		return m, nil

	case "search":
		// Open search input
		m.state = StateDevServerRunning
		m.searchInput.Focus()
		return m, textinput.Blink

	case "clear":
		// Clear all logs
		m.state = StateDevServerRunning
		m.logs = make([]logEntry, 0)
		m.updateLogView()
		return m, nil

	case "autoscroll":
		// Toggle auto-scroll
		m.state = StateDevServerRunning
		m.autoScroll = !m.autoScroll
		return m, nil

	case "help":
		// Show help
		m.state = StateDevServerRunning
		m.showHelp = true
		m.helpView.GotoTop()
		return m, nil

	case "back":
		// Stop server and go back - do this asynchronously
		m.state = StateDevServerStopping
		return m, stopServerCmd(m.runner)

	default:
		// Unknown action, just return to running state
		m.state = StateDevServerRunning
		return m, nil
	}
}

func (m DevServerDashboardModel) View() string {
	if m.showHelp {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.helpView.View())
	}

	// Header
	header := lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).PaddingTop(1).Render(
		titleStyle.Render("Dev Server Dashboard"),
	)

	var content string
	switch m.state {
	case StateDevServerPathInput:
		content = m.renderPathInput()
	case StateDevServerDetecting:
		content = m.renderDetecting()
	case StateDevServerReady:
		content = m.renderReady()
	case StateDevServerRunning:
		content = m.renderRunning()
	case StateDevServerStopping:
		content = m.renderRunning() // Reuse running view, status will show stopping
	case StateDevServerConfirmation:
		content = m.renderConfirmation()
	default:
		content = "Unknown state"
	}

	combined := lipgloss.JoinVertical(lipgloss.Center, header, "\n", content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, combined)
}

func (m DevServerDashboardModel) renderPathInput() string {
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("141")).
		Bold(true).
		Render("Auto-Detect Framework")

	instruction := lipgloss.NewStyle().
		Foreground(lipgloss.Color("226")).
		Render("Enter the path to your project folder:")

	pathLabel := lipgloss.NewStyle().
		Foreground(lipgloss.Color("141")).
		Bold(true).
		Render("Path:")
	inputBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("141")).
		Padding(0, 1).
		Width(64).
		Render(m.pathInput.View())

	tip := lipgloss.NewStyle().
		Foreground(lipgloss.Color("46")).
		Render("Tip: Press Enter without typing to use current path")

	// Current path box at the bottom
	currentPathBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 1).
		Width(64).
		Foreground(lipgloss.Color("255")).
		Render(fmt.Sprintf("Current: %s", m.projectPath))

	helpText := subtleStyle.Render("[Enter] Scan This Path • [Esc] Back")

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		instruction,
		"",
		pathLabel,
		inputBox,
		"",
		tip,
		"",
		currentPathBox,
		"",
		helpText,
	)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("141")).
		Padding(2, 4).
		Width(75)

	return boxStyle.Render(content)
}

func (m DevServerDashboardModel) renderDetecting() string {
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("141")).
		Bold(true).
		Render("Auto-Detect Framework")

	scanText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("226")).
		Render("Scanning project folder...")

	detailText := subtleStyle.Render("Looking for: manage.py, package.json, pom.xml, vite.config.js, and more...")

	content := lipgloss.JoinVertical(lipgloss.Center,
		title,
		"\n",
		scanText,
		"\n",
		detailText,
	)

	return content
}

func (m DevServerDashboardModel) renderReady() string {
	if m.err != nil {
		content := lipgloss.JoinVertical(lipgloss.Center,
			lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true).Render(" Detection Failed"),
			"\n",
			m.err.Error(),
			"\n",
			subtleStyle.Render("Press [Esc] to go back"),
		)
		return content
	}

	// Title
	titleText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("141")).
		Bold(true).
		Render("Auto-Detect Framework")

	// Detected Framework - Large and prominent
	frameworkStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("46")). // Green
		Bold(true).
		Render(string(m.projectInfo.Type))

	detectedLine := fmt.Sprintf("Detected: %s", frameworkStyle)

	// Show what was found (detection method)
	var detectionMethod string
	if len(m.projectInfo.Servers) > 0 {
		switch m.projectInfo.Type {
		case devserver.TypeDjango:
			detectionMethod = "Found: manage.py"
		case devserver.TypeFastAPI:
			detectionMethod = "Found: main.py + fastapi import"
		case devserver.TypeReact:
			detectionMethod = "Found: package.json (React)"
		case devserver.TypeVite:
			detectionMethod = "Found: vite.config.js"
		case devserver.TypeWebpack:
			detectionMethod = "Found: webpack.config.js"
		case devserver.TypeSpring:
			detectionMethod = "Found: pom.xml"
		case devserver.TypeNode:
			detectionMethod = "Found: package.json"
		case devserver.TypePython:
			detectionMethod = "Found: Python project files"
		case devserver.TypeGo:
			detectionMethod = "Found: go.mod"
		case devserver.TypeFullstack:
			detectionMethod = "Found: backend/ + frontend/ folders"
		default:
			detectionMethod = "Project detected"
		}
	}

	methodStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(detectionMethod)

	// Command that will run
	var commandInfo strings.Builder
	commandInfo.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("141")).
		Render("Command to run:") + "\n\n")

	for i, srv := range m.projectInfo.Servers {
		cmdStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")). // Yellow
			Bold(true)

		if len(m.projectInfo.Servers) > 1 {
			commandInfo.WriteString(fmt.Sprintf("  %s: %s\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("141")).Render(srv.Name),
				cmdStyle.Render(fmt.Sprintf("%s %s", srv.Cmd, strings.Join(srv.Args, " "))),
			))
		} else {
			commandInfo.WriteString(fmt.Sprintf("  %s\n",
				cmdStyle.Render(fmt.Sprintf("%s %s", srv.Cmd, strings.Join(srv.Args, " "))),
			))
		}

		if i < len(m.projectInfo.Servers)-1 {
			commandInfo.WriteString("\n")
		}
	}

	// Big "Just press Start" instruction
	startInstruction := lipgloss.NewStyle().
		Foreground(lipgloss.Color("46")).
		Bold(true).
		Render("Just press [s] to Start!")

	// Help text
	helpText := subtleStyle.Render("[s] Start • [?] Help • [Esc] Back")

	// Assemble content
	content := lipgloss.JoinVertical(lipgloss.Left,
		titleText,
		"",
		detectedLine,
		methodStyle,
		"",
		"",
		commandInfo.String(),
		"",
		"",
		startInstruction,
		"",
		helpText,
	)

	// Create a nice box around it
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("141")).
		Padding(2, 4).
		Width(60)

	return boxStyle.Render(content)
}

func (m DevServerDashboardModel) renderRunning() string {
	// Header
	statusIcon := ""
	statusColor := lipgloss.Color("46") // Green
	header := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Left).
		Foreground(lipgloss.Color("141")).
		Bold(true).
		MarginBottom(1).
		Render(fmt.Sprintf("Dev Server - %s", m.projectInfo.Type))

	status := lipgloss.NewStyle().
		Foreground(statusColor).
		Bold(true).
		Render(fmt.Sprintf("Status: %s Running", statusIcon))

	if m.state == StateDevServerStopping {
		status = lipgloss.NewStyle().
			Foreground(lipgloss.Color("208")). // Orange
			Bold(true).
			Render("Status:  Stopping...")
	}

	// Filters
	filterStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	activeFilterStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("141")).Bold(true)

	var filterButtons []string
	filters := []string{"All", "Errors", "Warnings"}
	for _, f := range filters {
		if strings.ToLower(f) == m.filterMode {
			filterButtons = append(filterButtons, activeFilterStyle.Render("[ "+f+" ]"))
		} else {
			filterButtons = append(filterButtons, filterStyle.Render("[ "+f+" ]"))
		}
	}

	filterLine := lipgloss.NewStyle().
		MarginTop(1).
		Render(fmt.Sprintf("Filters:  %s", strings.Join(filterButtons, "  ")))

	// Server filter (only for fullstack)
	var serverFilterLine string
	if m.projectInfo.Type == devserver.TypeFullstack {
		var serverButtons []string
		servers := []string{"All", "Backend", "Frontend"}
		for _, s := range servers {
			if strings.ToLower(s) == m.serverFilter {
				serverButtons = append(serverButtons, activeFilterStyle.Render("[ "+s+" ]"))
			} else {
				serverButtons = append(serverButtons, filterStyle.Render("[ "+s+" ]"))
			}
		}
		serverFilterLine = fmt.Sprintf("Source:   %s", strings.Join(serverButtons, "  "))
	}

	// Search
	searchLine := fmt.Sprintf("Search:   %s", m.searchInput.View())

	// Auto-scroll indicator
	scrollIndicator := ""
	if m.autoScroll {
		scrollIndicator = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Bold(true).
			Render("Auto-scroll ON")
	} else {
		scrollIndicator = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("Auto-scroll OFF")
	}

	// Footer
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		MarginTop(1).
		Render("[s] Stop • [f] Filter • [b] Source • [/] Search • [a] Auto-scroll • [c] Clear • [?] Help • [Esc] Back")

	// Assemble
	var content string
	if serverFilterLine != "" {
		content = lipgloss.JoinVertical(lipgloss.Left,
			"",
			header,
			status,
			"",
			filterLine,
			serverFilterLine,
			searchLine,
			scrollIndicator,
			"",
			"",
			m.logView.View(),
			"",
			footer,
			"",
		)
	} else {
		content = lipgloss.JoinVertical(lipgloss.Left,
			"",
			header,
			status,
			"",
			filterLine,
			searchLine,
			scrollIndicator,
			"",
			"",
			m.logView.View(),
			"",
			footer,
			"",
		)
	}

	return docStyle.Render(content)
}

func (m DevServerDashboardModel) renderConfirmation() string {
	// Create confirmation dialog overlay
	confirmTitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("226")).
		Bold(true).
		Render("Confirmation Required")

	confirmMessage := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255")).
		Render(m.confirmationMessage)

	yesOption := lipgloss.NewStyle().
		Foreground(lipgloss.Color("46")).
		Bold(true).
		Render("[y] Yes")

	noOption := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true).
		Render("[n] No")

	options := fmt.Sprintf("%s  •  %s", yesOption, noOption)

	dialogContent := lipgloss.JoinVertical(lipgloss.Center,
		confirmTitle,
		"",
		confirmMessage,
		"",
		options,
	)

	// Create a box for the dialog
	dialogBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("226")).
		Padding(2, 4).
		Width(50).
		Render(dialogContent)

	return dialogBox
}
