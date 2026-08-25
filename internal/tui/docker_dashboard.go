package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/phravins/devcli/internal/devtools"
)

const (
	StateDockerList = iota
	StateDockerLogs
	StateDockerError
)

type containerItem struct {
	container devtools.Container
}

func (i containerItem) Title() string {
	statusBadge := "[RUNNING]"
	if !strings.Contains(strings.ToLower(i.container.Status), "up") {
		statusBadge = "[STOPPED]"
	}
	return fmt.Sprintf("%s  %s (%s)", statusBadge, i.container.Names, i.container.Image)
}

func (i containerItem) Description() string {
	return fmt.Sprintf("ID: %s • Status: %s", i.container.ID[:12], i.container.Status)
}

func (i containerItem) FilterValue() string { return i.container.Names + " " + i.container.Image }

type DockerDashboardModel struct {
	state     int
	list      list.Model
	viewport  viewport.Model
	spinner   spinner.Model
	width     int
	height    int
	selected  devtools.Container
	err       error
	statusMsg string
}

type dockerLoadedMsg struct {
	containers []devtools.Container
	err        error
}

type dockerActionMsg struct {
	message string
	err     error
}

type dockerLogsMsg struct {
	logs string
	err  error
}

func NewDockerDashboardModel() DockerDashboardModel {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "🐳 Docker Container Dashboard"
	l.SetShowTitle(true)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2)

	return DockerDashboardModel{
		state:    StateDockerList,
		list:     l,
		viewport: vp,
		spinner:  s,
	}
}

func (m DockerDashboardModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, fetchDockerContainersCmd())
}

func fetchDockerContainersCmd() tea.Cmd {
	return func() tea.Msg {
		if !devtools.IsDockerAvailable() {
			return dockerLoadedMsg{err: fmt.Errorf("docker daemon is not running or not installed")}
		}
		containers, err := devtools.ListContainers()
		return dockerLoadedMsg{containers: containers, err: err}
	}
}

func (m DockerDashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width-4, msg.Height-8)
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 8

	case dockerLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.state = StateDockerError
			return m, nil
		}
		var items []list.Item
		for _, c := range msg.containers {
			items = append(items, containerItem{container: c})
		}
		m.list.SetItems(items)
		m.err = nil

	case dockerActionMsg:
		if msg.err != nil {
			m.statusMsg = "Error: " + msg.err.Error()
		} else {
			m.statusMsg = msg.message
		}
		return m, fetchDockerContainersCmd()

	case dockerLogsMsg:
		if msg.err != nil {
			m.viewport.SetContent("Error fetching logs: " + msg.err.Error())
		} else {
			m.viewport.SetContent(msg.logs)
		}
		m.state = StateDockerLogs

	case tea.KeyMsg:
		switch m.state {
		case StateDockerList:
			switch msg.String() {
			case "esc":
				return m, func() tea.Msg { return BackMsg{} }
			case "r":
				m.statusMsg = "Refreshing containers..."
				return m, fetchDockerContainersCmd()
			case "s":
				if i, ok := m.list.SelectedItem().(containerItem); ok {
					m.statusMsg = "Toggling container status..."
					return m, toggleContainerCmd(i.container)
				}
			case "l", "enter":
				if i, ok := m.list.SelectedItem().(containerItem); ok {
					m.selected = i.container
					return m, fetchContainerLogsCmd(i.container.ID)
				}
			}

		case StateDockerLogs, StateDockerError:
			if msg.String() == "esc" || msg.String() == "q" {
				m.state = StateDockerList
				return m, nil
			}
		}
	}

	if m.state == StateDockerList {
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	} else if m.state == StateDockerLogs {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func toggleContainerCmd(c devtools.Container) tea.Cmd {
	return func() tea.Msg {
		if strings.Contains(strings.ToLower(c.Status), "up") {
			err := devtools.StopContainer(c.ID)
			return dockerActionMsg{message: "Stopped container " + c.Names, err: err}
		}
		err := devtools.StartContainer(c.ID)
		return dockerActionMsg{message: "Started container " + c.Names, err: err}
	}
}

func fetchContainerLogsCmd(id string) tea.Cmd {
	return func() tea.Msg {
		logs, err := devtools.GetContainerLogs(id)
		return dockerLogsMsg{logs: logs, err: err}
	}
}

func (m DockerDashboardModel) View() string {
	statusBar := RenderStatusBar(m.width)

	switch m.state {
	case StateDockerList:
		actions := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("\n[Enter/l] Logs • [s] Start/Stop • [r] Refresh • [Esc] Back")
		content := m.list.View() + "\n" + m.statusMsg + actions
		return lipgloss.JoinVertical(lipgloss.Left, statusBar, content)

	case StateDockerLogs:
		header := lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).Render(fmt.Sprintf("Logs: %s (%s)", m.selected.Names, m.selected.Image))
		footer := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("\nPress [Esc] or [q] to return to containers")
		return lipgloss.JoinVertical(lipgloss.Left, statusBar, header, m.viewport.View(), footer)

	case StateDockerError:
		errText := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true).Render("Docker Connection Error:\n" + m.err.Error())
		footer := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("\nEnsure Docker Desktop or dockerd service is active.\nPress [Esc] to return.")
		return lipgloss.JoinVertical(lipgloss.Left, statusBar, errText, footer)
	}

	return "Unknown Docker view"
}
