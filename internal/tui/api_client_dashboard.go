package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/phravins/devcli/internal/devtools"
)

const (
	StateAPIMethodSelect = iota
	StateAPIURLInput
	StateAPIBodyInput
	StateAPISending
	StateAPIResponseView
)

type APIClientModel struct {
	state          int
	methodIndex    int
	methods        []string
	urlInput       textinput.Model
	bodyInput      textinput.Model
	viewport       viewport.Model
	spinner        spinner.Model
	width, height  int
	lastResponse   devtools.APIResponse
	statusMsg      string
}

type apiExecMsg struct {
	response devtools.APIResponse
}

func NewAPIClientModel() APIClientModel {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

	tiURL := textinput.New()
	tiURL.Placeholder = "https://httpbin.org/get or http://127.0.0.1:8080/auth/login"
	tiURL.Width = 60
	tiURL.SetValue("https://httpbin.org/get")

	tiBody := textinput.New()
	tiBody.Placeholder = `{"key": "value"}`
	tiBody.Width = 60

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2)

	return APIClientModel{
		state:       StateAPIMethodSelect,
		methodIndex: 0,
		methods:     methods,
		urlInput:    tiURL,
		bodyInput:   tiBody,
		viewport:    vp,
		spinner:     s,
	}
}

func (m APIClientModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, textinput.Blink)
}

func (m APIClientModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 12

	case apiExecMsg:
		m.lastResponse = msg.response
		m.state = StateAPIResponseView

		// Render response
		resp := msg.response
		var sb strings.Builder

		if resp.Err != nil {
			sb.WriteString(fmt.Sprintf("Error: %v\n", resp.Err))
		} else {
			sb.WriteString(fmt.Sprintf("Status  : %s\n", resp.Status))
			sb.WriteString(fmt.Sprintf("Latency : %d ms\n", resp.LatencyMs))
			sb.WriteString("----------------------------------------\n\n")
			sb.WriteString(resp.Formatted)
		}

		m.viewport.SetContent(sb.String())

	case tea.KeyMsg:
		switch m.state {
		case StateAPIMethodSelect:
			switch msg.String() {
			case "esc":
				return m, func() tea.Msg { return BackMsg{} }
			case "up", "k":
				if m.methodIndex > 0 {
					m.methodIndex--
				}
			case "down", "j":
				if m.methodIndex < len(m.methods)-1 {
					m.methodIndex++
				}
			case "enter":
				m.state = StateAPIURLInput
				m.urlInput.Focus()
				return m, textinput.Blink
			}

		case StateAPIURLInput:
			switch msg.String() {
			case "esc":
				m.urlInput.Blur()
				m.state = StateAPIMethodSelect
				return m, nil
			case "enter":
				m.urlInput.Blur()
				if m.methods[m.methodIndex] == "GET" || m.methods[m.methodIndex] == "DELETE" {
					m.state = StateAPISending
					return m, tea.Batch(m.spinner.Tick, executeAPIReqCmd(m.methods[m.methodIndex], m.urlInput.Value(), ""))
				}
				m.state = StateAPIBodyInput
				m.bodyInput.Focus()
				return m, textinput.Blink
			}
			m.urlInput, cmd = m.urlInput.Update(msg)
			return m, cmd

		case StateAPIBodyInput:
			switch msg.String() {
			case "esc":
				m.bodyInput.Blur()
				m.state = StateAPIURLInput
				m.urlInput.Focus()
				return m, nil
			case "enter":
				m.bodyInput.Blur()
				m.state = StateAPISending
				return m, tea.Batch(m.spinner.Tick, executeAPIReqCmd(m.methods[m.methodIndex], m.urlInput.Value(), m.bodyInput.Value()))
			}
			m.bodyInput, cmd = m.bodyInput.Update(msg)
			return m, cmd

		case StateAPIResponseView:
			if msg.String() == "esc" || msg.String() == "q" {
				m.state = StateAPIMethodSelect
				return m, nil
			}
			if msg.String() == "r" {
				m.state = StateAPISending
				return m, tea.Batch(m.spinner.Tick, executeAPIReqCmd(m.methods[m.methodIndex], m.urlInput.Value(), m.bodyInput.Value()))
			}
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}
	}

	return m, tea.Batch(cmds...)
}

func executeAPIReqCmd(method, url, body string) tea.Cmd {
	return func() tea.Msg {
		req := devtools.APIRequest{
			Method:  method,
			URL:     url,
			Headers: map[string]string{"Accept": "application/json"},
			Body:    body,
		}
		resp := devtools.ExecuteAPIRequest(req)
		return apiExecMsg{response: resp}
	}
}

func (m APIClientModel) View() string {
	statusBar := RenderStatusBar(m.width)

	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).Render("🌐 DevCLI API & HTTP Playground")

	switch m.state {
	case StateAPIMethodSelect:
		var sb strings.Builder
		sb.WriteString(headerStyle + "\n\nSelect HTTP Method:\n\n")

		for i, method := range m.methods {
			cursor := "  "
			style := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
			if i == m.methodIndex {
				cursor = "👉 "
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("50FA7B")).Bold(true)
			}
			sb.WriteString(fmt.Sprintf("%s%s\n", cursor, style.Render(method)))
		}

		sb.WriteString("\n[Enter] Select Method • [Esc] Back")
		return lipgloss.JoinVertical(lipgloss.Left, statusBar, sb.String())

	case StateAPIURLInput:
		method := m.methods[m.methodIndex]
		card := lipgloss.JoinVertical(lipgloss.Left,
			headerStyle,
			fmt.Sprintf("\nMethod: %s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("50FA7B")).Bold(true).Render(method)),
			"Enter Target URL:",
			m.urlInput.View(),
			"\n[Enter] Next/Send • [Esc] Back",
		)
		return lipgloss.JoinVertical(lipgloss.Left, statusBar, card)

	case StateAPIBodyInput:
		method := m.methods[m.methodIndex]
		card := lipgloss.JoinVertical(lipgloss.Left,
			headerStyle,
			fmt.Sprintf("\nMethod: %s  URL: %s\n", method, m.urlInput.Value()),
			"Enter Request Body JSON (Optional):",
			m.bodyInput.View(),
			"\n[Enter] Send Request • [Esc] Back",
		)
		return lipgloss.JoinVertical(lipgloss.Left, statusBar, card)

	case StateAPISending:
		return lipgloss.JoinVertical(lipgloss.Left, statusBar,
			lipgloss.Place(m.width, m.height-2, lipgloss.Center, lipgloss.Center,
				lipgloss.JoinVertical(lipgloss.Center,
					m.spinner.View(),
					fmt.Sprintf("\nSending %s request to %s...", m.methods[m.methodIndex], m.urlInput.Value()),
				),
			),
		)

	case StateAPIResponseView:
		resp := m.lastResponse
		badgeColor := "46" // Green for 2xx
		if resp.StatusCode >= 400 {
			badgeColor = "196" // Red for 4xx/5xx
		}

		statusBadge := lipgloss.NewStyle().Background(lipgloss.Color(badgeColor)).Foreground(lipgloss.Color("0")).Bold(true).Padding(0, 1).Render(resp.Status)
		timerBadge := lipgloss.NewStyle().Background(lipgloss.Color("62")).Foreground(lipgloss.Color("255")).Padding(0, 1).Render(fmt.Sprintf("%d ms", resp.LatencyMs))

		topInfo := lipgloss.JoinHorizontal(lipgloss.Left, statusBadge, " ", timerBadge, fmt.Sprintf("  URL: %s", m.urlInput.Value()))
		footer := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("\n[r] Resend • [Esc/q] Back")

		return lipgloss.JoinVertical(lipgloss.Left, statusBar, topInfo, "\n", m.viewport.View(), footer)
	}

	return "Unknown API Client state"
}
