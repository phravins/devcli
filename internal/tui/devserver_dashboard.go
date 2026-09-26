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
	width, height		int
	state			int
	projectPath		string
	projectInfo		devserver.ProjectInfo
	runner			*devserver.Runner
	logView			viewport.Model
	helpView		viewport.Model
	searchInput		textinput.Model
	pathInput		textinput.Model
	logs			[]logEntry
	filterMode		string
	serverFilter		string
	autoScroll		bool
	showHelp		bool
	err			error
	pendingAction		string
	confirmationMessage	string
}

type logEntry struct {
	timestamp	string
	serverName	string
	line		string
	isError		bool
	isWarning	bool
}

const (
	StateDevServerPathInput	= iota
	StateDevServerDetecting
	StateDevServerReady
	StateDevServerRunning
	StateDevServerConfirmation
	StateDevServerStopping
	StateDevServerHelp
)

type detectDoneMsg struct {
	info	devserver.ProjectInfo
	err	error
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

	pi := textinput.New()
	pi.Placeholder = "Type path here or press Enter for current"
	pi.SetValue(projectPath)
	pi.CharLimit = 200
	pi.Width = 62
	pi.Focus()

	hv := viewport.New(80, 20)
	hv.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#0F9E99")).
		Padding(1, 2)
	hv.SetContent(DevServerHelp)

	return DevServerDashboardModel{
		state:		StateDevServerPathInput,
		projectPath:	projectPath,
		logView:	vp,
		helpView:	hv,
		searchInput:	ti,
		pathInput:	pi,
		logs:		make([]logEntry, 0),
		filterMode:	"all",
		serverFilter:	"all",
		autoScroll:	true,
		showHelp:	false,
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
				info:	info,
				err:	fmt.Errorf("unable to detect project type"),
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

				return nil
			}
			return logReceivedMsg{log: log}
		case <-time.After(100 * time.Millisecond):

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

		if m.state == StateDevServerPathInput {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "esc":
				return m, func() tea.Msg { return DevServerBackMsg{} }
			case "enter":

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

		if m.state == StateDevServerConfirmation {
			switch msg.String() {
			case "y", "Y":

				return m.executePendingAction()
			case "n", "N", "esc":

				m.state = StateDevServerRunning
				m.pendingAction = ""
				m.confirmationMessage = ""

				if m.runner != nil {
					return m, waitForLogCmd(m.runner)
				}
				return m, nil
			default:
				return m, nil
			}
		}

		switch msg.String() {
		case "ctrl+c", "q":
			if m.runner != nil {
				m.runner.Stop()
			}
			return m, tea.Quit
		case "esc":
			if m.state == StateDevServerRunning && m.runner != nil {

				m.state = StateDevServerConfirmation
				m.pendingAction = "back"
				m.confirmationMessage = "Stop the server and go back?"
				return m, nil
			} else {
				return m, func() tea.Msg { return DevServerBackMsg{} }
			}
		case "?":
			if m.state == StateDevServerRunning && m.runner != nil {

				m.state = StateDevServerConfirmation
				m.pendingAction = "help"
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

				m.state = StateDevServerConfirmation
				m.pendingAction = "stop"
				m.confirmationMessage = "Stop the server?"
				return m, nil
			}
			return m, nil
		case "f":
			if m.state == StateDevServerRunning && m.runner != nil {

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

				m.state = StateDevServerConfirmation
				m.pendingAction = "search"
				m.confirmationMessage = "Open search?"
				m.pendingAction = "search"
				m.confirmationMessage = "Open search?"
				return m, nil
			}
			return m, nil
		case "up", "down", "pgup", "pgdown", "home", "end":

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
			timestamp:	timestamp,
			serverName:	msg.log.ServerName,
			line:		msg.log.Line,
			isError:	msg.log.IsError,
			isWarning:	isWarning,
		})

		m.updateLogView()
		if m.autoScroll {
			m.logView.GotoBottom()
		}

		if (m.state == StateDevServerRunning || m.state == StateDevServerConfirmation || m.state == StateDevServerStopping) && m.runner != nil {
			return m, waitForLogCmd(m.runner)
		}
		return m, nil

	case tickMsg:

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

		var cmd tea.Cmd
		m.logView, cmd = m.logView.Update(msg)
		return m, cmd

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		m.logView.Width = msg.Width - 4
		m.logView.Height = msg.Height - 14

		m.helpView.Width = msg.Width - 8
		m.helpView.Height = msg.Height - 4
	}

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

		if m.filterMode == "errors" && !log.isError {
			continue
		}
		if m.filterMode == "warnings" && !log.isWarning {
			continue
		}

		if m.serverFilter != "all" {
			if m.serverFilter == "backend" && !strings.Contains(strings.ToLower(log.serverName), "backend") {
				continue
			}
			if m.serverFilter == "frontend" && !strings.Contains(strings.ToLower(log.serverName), "frontend") {
				continue
			}
		}

		if searchTerm != "" && !strings.Contains(strings.ToLower(log.line), searchTerm) {
			continue
		}

		var lineStyle lipgloss.Style
		if log.isError {
			lineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		} else if log.isWarning {
			lineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
		} else {
			lineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
		}

		serverStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("141")).Bold(true)
		timeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

		formattedLine := fmt.Sprintf("%s [%s] %s\n",
			timeStyle.Render(log.timestamp),
			serverStyle.Render(log.serverName),
			lineStyle.Render(log.line),
		)

		if searchTerm != "" {
			highlightStyle := lipgloss.NewStyle().Background(lipgloss.Color("226")).Foreground(lipgloss.Color("0"))
			formattedLine = strings.ReplaceAll(formattedLine, searchTerm, highlightStyle.Render(searchTerm))
		}

		content.WriteString(formattedLine)
	}

	m.logView.SetContent(content.String())
}

func (m DevServerDashboardModel) executePendingAction() (DevServerDashboardModel, tea.Cmd) {

	action := m.pendingAction
	m.pendingAction = ""
	m.confirmationMessage = ""

	switch action {
	case "stop":

		m.state = StateDevServerStopping
		return m, stopServerCmd(m.runner)

	case "filter":

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

		m.state = StateDevServerRunning
		m.searchInput.Focus()
		return m, textinput.Blink

	case "clear":

		m.state = StateDevServerRunning
		m.logs = make([]logEntry, 0)
		m.updateLogView()
		return m, nil

	case "autoscroll":

		m.state = StateDevServerRunning
		m.autoScroll = !m.autoScroll
		return m, nil

	case "help":

		m.state = StateDevServerRunning
		m.showHelp = true
		m.helpView.GotoTop()
		return m, nil

	case "back":

		m.state = StateDevServerStopping
		return m, stopServerCmd(m.runner)

	default:

		m.state = StateDevServerRunning
		return m, nil
	}
}

func (m DevServerDashboardModel) View() string {
	if m.showHelp {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.helpView.View())
	}

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
		content = m.renderRunning()
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

	titleText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("141")).
		Bold(true).
		Render("Auto-Detect Framework")

	frameworkStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("46")).
		Bold(true).
		Render(string(m.projectInfo.Type))

	detectedLine := fmt.Sprintf("Detected: %s", frameworkStyle)

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

	var commandInfo strings.Builder
	commandInfo.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("141")).
		Render("Command to run:"))
	commandInfo.WriteString("\n\n")

	for i, srv := range m.projectInfo.Servers {
		cmdStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")).
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

	startInstruction := lipgloss.NewStyle().
		Foreground(lipgloss.Color("46")).
		Bold(true).
		Render("Just press [s] to Start!")

	helpText := subtleStyle.Render("[s] Start • [?] Help • [Esc] Back")

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

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("141")).
		Padding(2, 4).
		Width(60)

	return boxStyle.Render(content)
}

func (m DevServerDashboardModel) renderRunning() string {

	statusIcon := ""
	statusColor := lipgloss.Color("46")
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
			Foreground(lipgloss.Color("208")).
			Bold(true).
			Render("Status:  Stopping...")
	}

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

	searchLine := fmt.Sprintf("Search:   %s", m.searchInput.View())

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

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		MarginTop(1).
		Render("[s] Stop • [f] Filter • [b] Source • [/] Search • [a] Auto-scroll • [c] Clear • [?] Help • [Esc] Back")

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

	dialogBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("226")).
		Padding(2, 4).
		Width(50).
		Render(dialogContent)

	return dialogBox
}
