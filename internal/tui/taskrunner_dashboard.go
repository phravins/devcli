package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/phravins/devcli/internal/taskrunner"
)

type TaskRunnerModel struct {
	workspace   string
	tasks       []taskrunner.Task
	list        list.Model
	running     bool
	output      *strings.Builder
	outputView  viewport.Model
	helpView    viewport.Model
	currentTask *taskrunner.Task
	ctx         context.Context
	cancel      context.CancelFunc
	width       int
	height      int
	state       int // 0: list, 1: running, 2: completed, 3: help
}

const (
	trStateList = iota
	trStateRunning
	trStateCompleted
	trStateHelp
)

type tasksDetectedMsg struct {
	tasks []taskrunner.Task
}

type taskOutputMsg string
type taskCompleteMsg struct {
	err error
}

func NewTaskRunnerModel(workspace string) TaskRunnerModel {
	return TaskRunnerModel{
		workspace:  workspace,
		list:       list.New([]list.Item{}, list.NewDefaultDelegate(), 60, 14),
		output:     &strings.Builder{},
		outputView: viewport.New(80, 20),
		helpView:   viewport.New(80, 20),
		state:      trStateList,
	}
}

func (m TaskRunnerModel) Init() tea.Cmd {
	return func() tea.Msg {
		tasks := taskrunner.DetectTasks(m.workspace)
		return tasksDetectedMsg{tasks: tasks}
	}
}

func (m TaskRunnerModel) Update(msg tea.Msg) (TaskRunnerModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tasksDetectedMsg:
		m.tasks = msg.tasks
		items := make([]list.Item, len(m.tasks))
		for i, task := range m.tasks {
			items[i] = item{
				title: fmt.Sprintf("%s %s", task.Icon, task.Name),
				desc:  task.Description,
			}
		}
		m.list.SetItems(items)
		return m, nil

	case taskOutputMsg:
		// Limit buffer size to prevent memory issues with long processes
		const maxOutputLen = 50000 // characters

		newStr := string(msg) + "\n"

		if m.output.Len()+len(newStr) > maxOutputLen {
			// Truncate old output
			fullStr := m.output.String() + newStr
			keepStart := len(fullStr) - maxOutputLen
			if keepStart < 0 {
				keepStart = 0
			}
			m.output.Reset()
			m.output.WriteString("... (truncated output) ...\n")
			m.output.WriteString(fullStr[keepStart:])
		} else {
			m.output.WriteString(newStr)
		}

		m.outputView.SetContent(m.output.String())
		m.outputView.GotoBottom()
		return m, nil

	case taskCompleteMsg:
		m.running = false
		m.state = trStateCompleted
		if msg.err != nil {
			m.output.WriteString(fmt.Sprintf("\n Error: %v\n", msg.err))
		} else {
			m.output.WriteString("\n Task completed successfully!\n")
		}
		m.outputView.SetContent(m.output.String())
		m.outputView.GotoBottom()
		return m, nil

	case tea.KeyMsg:
		if m.state == trStateHelp {
			if msg.String() == "esc" || msg.String() == "q" || msg.String() == "?" {
				m.state = trStateList
				return m, nil
			}
			m.helpView, cmd = m.helpView.Update(msg)
			return m, cmd
		}

		switch m.state {
		case trStateList:
			switch msg.String() {
			case "esc", "q":
				return m, func() tea.Msg { return SubFeatureBackMsg{} }

			case "r":
				return m, m.Init()
			case "?":
				m.state = trStateHelp
				m.helpView.SetContent(TaskRunnerHelp)
				m.helpView.Width = m.width - 10
				m.helpView.Height = m.height - 10
				m.helpView.GotoTop()
				return m, nil
			case "enter":
				if len(m.tasks) == 0 {
					return m, nil
				}
				idx := m.list.Index()
				if idx >= 0 && idx < len(m.tasks) {
					m.currentTask = &m.tasks[idx]
					m.state = trStateRunning
					m.running = true
					m.output.Reset()
					m.output.WriteString(fmt.Sprintf("Running: %s\n\n", m.currentTask.Name))
					m.outputView.SetContent(m.output.String())

					m.ctx, m.cancel = context.WithCancel(context.Background())
					outputChan := make(chan string, 100)

					go func() {
						err := taskrunner.ExecuteTask(m.ctx, *m.currentTask, m.workspace, outputChan)
						if err != nil {
							tea.Printf("Task error: %v\n", err)
						}
					}()

					return m, func() tea.Msg {
						for line := range outputChan {
							return taskOutputMsg(line)
						}
						return taskCompleteMsg{}
					}
				}
			}
			m.list, cmd = m.list.Update(msg)
			return m, cmd

		case trStateRunning:
			switch msg.String() {
			case "ctrl+c":
				if m.cancel != nil {
					m.cancel()
				}
				m.running = false
				m.state = trStateCompleted
				m.output.WriteString("\n Task cancelled\n")
				m.outputView.SetContent(m.output.String())
				return m, nil
			case "?":
				m.state = trStateHelp
				m.helpView.SetContent(TaskRunnerHelp)
				m.helpView.Width = m.width - 10
				m.helpView.Height = m.height - 10
				m.helpView.GotoTop()
				return m, nil
			}
			m.outputView, cmd = m.outputView.Update(msg)
			return m, cmd

		case trStateCompleted:
			switch msg.String() {
			case "esc", "enter", "q":
				m.state = trStateList
				return m, nil
			case "?":
				m.state = trStateHelp
				m.helpView.SetContent(TaskRunnerHelp)
				m.helpView.Width = m.width - 10
				m.helpView.Height = m.height - 10
				m.helpView.GotoTop()
				return m, nil
			}
			m.outputView, cmd = m.outputView.Update(msg)
			return m, cmd
		}

	case tea.MouseMsg:
		switch m.state {
		case trStateHelp:
			if msg.Type == tea.MouseWheelUp {
				m.helpView.LineUp(3)
				return m, nil
			}
			if msg.Type == tea.MouseWheelDown {
				m.helpView.LineDown(3)
				return m, nil
			}
			m.helpView, cmd = m.helpView.Update(msg)
			return m, cmd
		case trStateList:
			if msg.Type == tea.MouseWheelUp {
				m.list.CursorUp()
				return m, nil
			}
			if msg.Type == tea.MouseWheelDown {
				m.list.CursorDown()
				return m, nil
			}
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		case trStateRunning, trStateCompleted:
			if msg.Type == tea.MouseWheelUp {
				m.outputView.LineUp(3)
				return m, nil
			}
			if msg.Type == tea.MouseWheelDown {
				m.outputView.LineDown(3)
				return m, nil
			}
			m.outputView, cmd = m.outputView.Update(msg)
			return m, cmd
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width-4, msg.Height-10)
		m.outputView.Width = msg.Width - 6
		m.outputView.Height = msg.Height - 8
		m.helpView.Width = msg.Width - 10
		m.helpView.Height = msg.Height - 10
	}

	return m, nil
}

func (m TaskRunnerModel) View() string {
	switch m.state {
	case trStateList:
		if len(m.tasks) == 0 {
			empty := " No tasks detected in this project\n\n" +
				"Supported project types:\n" +
				"• Node.js (package.json)\n" +
				"• Python (requirements.txt, *.py)\n" +
				"• Java (Maven, Gradle, Main.java)\n" +
				"• Go (go.mod)\n" +
				"• Rust (Cargo.toml)\n" +
				"• C/C++ (CMake, gcc/g++)\n" +
				"• Makefile\n\n" +
				"Press R to refresh • ? for help • Esc to go back"

			return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
				lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(empty))
		}

		header := lipgloss.NewStyle().
			Width(m.width).
			Align(lipgloss.Center).
			PaddingTop(1).
			Render(titleStyle.Render("Task Runner"))

		footer := lipgloss.NewStyle().
			Width(m.width).
			Align(lipgloss.Left).
			PaddingLeft(4).
			PaddingTop(1).
			Render(subtleStyle.Render("↑/↓: Navigate • Enter: Run Task • R: Refresh • ?: Help • Esc: Back"))

		// Left aligned content area
		listView := lipgloss.NewStyle().
			PaddingLeft(4).
			Render(m.list.View())

		content := lipgloss.JoinVertical(lipgloss.Left,
			header,
			"\n",
			listView,
			"\n",
			footer,
		)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Top, content)

	case trStateRunning:
		header := lipgloss.NewStyle().
			Width(m.width).
			Align(lipgloss.Center).
			PaddingTop(1).
			Render(titleStyle.Render(fmt.Sprintf("Running: %s", m.currentTask.Name)))

		status := lipgloss.NewStyle().
			Width(m.width).
			Align(lipgloss.Center).
			Foreground(lipgloss.Color("226")).
			Render("Task in progress...")

		footer := lipgloss.NewStyle().
			Width(m.width).
			Align(lipgloss.Left).
			PaddingLeft(4).
			Render(subtleStyle.Render("Ctrl+C: Cancel Task • ?: Help"))

		content := lipgloss.JoinVertical(lipgloss.Left,
			header,
			status,
			"\n",
			lipgloss.NewStyle().PaddingLeft(4).Render(m.outputView.View()),
			"\n",
			footer,
		)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Top, content)

	case trStateCompleted:
		header := lipgloss.NewStyle().
			Width(m.width).
			Align(lipgloss.Center).
			PaddingTop(1).
			Render(titleStyle.Render(fmt.Sprintf("Task Complete: %s", m.currentTask.Name)))

		footer := lipgloss.NewStyle().
			Width(m.width).
			Align(lipgloss.Left).
			PaddingLeft(4).
			Render(subtleStyle.Render("Press Enter or Esc to continue • ?: Help"))

		content := lipgloss.JoinVertical(lipgloss.Left,
			header,
			"\n",
			lipgloss.NewStyle().PaddingLeft(4).Render(m.outputView.View()),
			"\n",
			footer,
		)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Top, content)

	case trStateHelp:
		header := lipgloss.NewStyle().
			Width(m.width).
			Align(lipgloss.Center).
			PaddingTop(1).
			Render(titleStyle.Render("Task Runner Help"))

		footer := lipgloss.NewStyle().
			Width(m.width).
			Align(lipgloss.Center).
			Render(subtleStyle.Render("↑/↓ or Mouse Wheel: Scroll • Esc: Back"))

		helpBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#0F9E99")).
			Padding(0, 1).
			Render(m.helpView.View())

		content := lipgloss.JoinVertical(lipgloss.Center,
			header,
			"\n",
			helpBox,
			"\n",
			footer,
		)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}

	return ""
}
