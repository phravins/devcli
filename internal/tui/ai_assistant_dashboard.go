package tui

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/phravins/devcli/internal/ai"
	"github.com/phravins/devcli/internal/ai/providers"
	"github.com/phravins/devcli/internal/config"
)

type AIAssistantModel struct {
	input       textarea.Model
	output      viewport.Model
	spinner     spinner.Model
	provider    ai.Provider
	state       int // 0: input, 1: generating, 2: result
	activeAgent int // 0: CodeGen, 1: Architect, 2: Debugger
	prompt      string
	result      string
	width       int
	height      int

	// Help
	helpView viewport.Model
}

const (
	aiStateInput = iota
	aiStateGenerating
	aiStateResult
	aiStateHelp
)

func NewAIAssistantModel() AIAssistantModel {
	ta := textarea.New()
	ta.Placeholder = "Enter your requirements (e.g., 'Generate a Go CLI for file processing')..."
	ta.SetWidth(80)
	ta.SetHeight(10)
	ta.Focus()

	vp := viewport.New(80, 20)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	cfg, _ := config.LoadConfig()
	p, _ := providers.GetProvider(cfg)

	return AIAssistantModel{
		input:    ta,
		output:   vp,
		spinner:  s,
		provider: p,
		helpView: viewport.New(80, 20),
		state:    aiStateInput,
	}
}

func (m AIAssistantModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m AIAssistantModel) Update(msg tea.Msg) (AIAssistantModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global Agent Switching (Always works in any state)
		switch msg.String() {
		case "tab", "[", "ctrl+n":
			m.activeAgent = (m.activeAgent + 1) % 3
			return m, nil
		case "shift+tab", "]", "ctrl+p":
			m.activeAgent = (m.activeAgent - 1 + 3) % 3
			return m, nil
		case "?":
			if m.state != aiStateHelp {
				m.state = aiStateHelp
				sidebarWidth := 20
				mainAreaWidth := m.width - sidebarWidth - 4
				m.helpView.SetContent(RenderHelp(AIAssistantHelp, mainAreaWidth, m.height))
				return m, nil
			}
		}

		if m.state == aiStateHelp {
			if msg.String() == "esc" || msg.String() == "q" || msg.String() == "?" {
				m.state = aiStateInput
				return m, nil
			}
			var cmd tea.Cmd
			m.helpView, cmd = m.helpView.Update(msg)
			return m, cmd
		}

		switch m.state {
		case aiStateInput:
			switch msg.String() {
			case "esc":
				if m.input.Value() == "" {
					return m, func() tea.Msg { return SubFeatureBackMsg{} }
				}
				m.input.SetValue("")
				return m, nil
			case "ctrl+d":
				// Send prompt
				m.prompt = m.input.Value()
				if m.prompt != "" && m.provider != nil {
					m.state = aiStateGenerating
					return m, tea.Batch(
						m.spinner.Tick,
						m.sendToAI(m.prompt),
					)
				}
				return m, nil
			}
			m.input, cmd = m.input.Update(msg)
			return m, cmd

		case aiStateResult:
			switch msg.String() {
			case "esc", "q":
				m.state = aiStateInput
				m.input.SetValue("")
				m.input.Focus()
				return m, textarea.Blink
			case "n":
				// New prompt
				m.state = aiStateInput
				m.input.SetValue("")
				m.input.Focus()
				return m, textarea.Blink
			}
			m.output, cmd = m.output.Update(msg)
			return m, cmd
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case aiResponseMsg:
		m.result = string(msg)

		// Render results with Glamour for premium look
		renderer, _ := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(m.output.Width-4),
		)
		out, err := renderer.Render(m.result)
		if err == nil {
			m.output.SetContent(out)
		} else {
			m.output.SetContent(m.result)
		}

		m.state = aiStateResult
		m.output.GotoTop()
		return m, nil

	case tea.MouseMsg:
		if m.state == aiStateResult {
			m.output, cmd = m.output.Update(msg)
			return m, cmd
		}
		m.input, cmd = m.input.Update(msg)
		return m, cmd

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		sidebarWidth := 20
		mainWidth := m.width - sidebarWidth - 4
		m.input.SetWidth(mainWidth)
		m.input.SetHeight(m.height - 6)
		m.output.Width = mainWidth
		m.output.Height = m.height - 6
		m.helpView.Width = mainWidth
		m.helpView.Height = msg.Height
	}

	return m, nil
}

func (m AIAssistantModel) View() string {
	sidebarWidth := 20
	mainAreaWidth := m.width - sidebarWidth - 4

	// Sync heights to ensure the vertical boundary reaches the bottom
	workspaceHeight := m.height

	sidebarStyle := lipgloss.NewStyle().
		Width(sidebarWidth).
		Height(workspaceHeight).
		Border(lipgloss.NormalBorder(), false, true, false, false).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1)

	// Sidebar Content
	agentNames := []string{"CodeGen-v1", "Architect", "Debugger"}
	var agentItems []string
	for i, name := range agentNames {
		indicator := " ○ "
		if i == m.activeAgent {
			indicator = " ● "
		}

		item := lipgloss.NewStyle().Foreground(lipgloss.Color(func() string {
			if i == m.activeAgent {
				return "46"
			}
			return "240"
		}())).Render(indicator) +
			lipgloss.NewStyle().Foreground(lipgloss.Color(func() string {
				if i == m.activeAgent {
					return "255"
				}
				return "240"
			}())).Bold(i == m.activeAgent).Render(name)

		agentItems = append(agentItems, item)
	}

	// Correctly render sidebar once
	sidebar := sidebarStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
		"\n",
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")).Render(" AGENTS"),
		"\n",
		lipgloss.JoinVertical(lipgloss.Left, agentItems...),
		"\n\n",
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")).Render(" STATUS"),
		"\n",
		func() string {
			if m.state == aiStateGenerating {
				return lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Render(" Thinking")
			}
			return lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(" Ready")
		}(),
	))

	renderHeader := func(text string, color string) string {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color(color)).
			Bold(true).
			Padding(0, 4).
			Render(text)
	}

	var mainContent string
	switch m.state {
	case aiStateInput:
		header := renderHeader(" AGENT AI IDE :: COMPOSER ", "#7D56F4")
		inputBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 2).
			Width(mainAreaWidth).
			Height(m.height - 6).
			Render(m.input.View())
		footer := subtleStyle.Render("Ctrl+D: Dispatch • Ctrl+P/Tab: Switch Agent")
		mainContent = lipgloss.JoinVertical(lipgloss.Center, header, inputBox, footer)

	case aiStateGenerating:
		msg := lipgloss.JoinVertical(lipgloss.Center,
			m.spinner.View(),
			"\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).Render("THE AGENT IS SOLVING YOUR REQUEST"),
			"\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Performing multi-model inference..."),
		)
		mainContent = lipgloss.Place(mainAreaWidth, m.height-6, lipgloss.Center, lipgloss.Center, msg)

	case aiStateResult:
		header := renderHeader(" EXECUTION RESULT ", "#059669")
		viewportBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#059669")).
			Padding(1, 2).
			Width(mainAreaWidth).
			Height(m.height - 6).
			Render(m.output.View())
		footer := subtleStyle.Render("N: New • Esc: Back • Ctrl+P/Tab: Switch Agent")
		mainContent = lipgloss.JoinVertical(lipgloss.Center, header, viewportBox, footer)

	case aiStateHelp:
		mainContent = m.helpView.View()
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, sidebar, mainContent)
}

type aiResponseMsg string

func (m AIAssistantModel) sendToAI(prompt string) tea.Cmd {
	return func() tea.Msg {
		systemPrompts := []string{
			"You are an expert AI software engineer specialized in code generation. Provide high-quality, efficient code directly.",
			"You are a Senior System Architect. Provide high-level design patterns, architecture diagrams (markdown), and structural advice.",
			"You are an expert Debugger. Focus on identifying potential bugs, performance bottlenecks, and security vulnerabilities in the provided context.",
		}

		messages := []ai.Message{
			{Role: "system", Content: systemPrompts[m.activeAgent]},
			{Role: "user", Content: prompt},
		}
		resp, err := m.provider.Send(messages)
		if err != nil {
			return aiResponseMsg("Error: " + err.Error())
		}
		return aiResponseMsg(resp)
	}
}
