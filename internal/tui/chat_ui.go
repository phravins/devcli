package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/phravins/devcli/internal/ai"
	"github.com/phravins/devcli/internal/ai/providers"
	"github.com/phravins/devcli/internal/config"
)

type ChatModel struct {
	viewport	viewport.Model
	textarea	textarea.Model
	spinner		spinner.Model
	provider	ai.Provider
	messages	[]ai.Message
	err		error
	loading		bool
	width		int
	height		int
	ready		bool
	showHelp	bool
	helpView	viewport.Model
}

func NewChatModel() ChatModel {
	ta := textarea.New()
	ta.Placeholder = "Ask the AI..."
	ta.Focus()
	ta.Prompt = " "
	ta.CharLimit = 2000
	ta.SetHeight(3)
	ta.ShowLineNumbers = false

	vp := viewport.New(0, 0)

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	hv := viewport.New(0, 0)
	hv.Style = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("62")).Padding(1, 2)

	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	out, err := renderer.Render(AIchatHelp)
	if err != nil {
		out = AIchatHelp
	}
	hv.SetContent(out)

	cfg, _ := config.LoadConfig()
	p, err := providers.GetProvider(cfg)
	if err != nil {
		fmt.Printf("Error initializing AI provider: %v\n", err)
	}

	return ChatModel{
		textarea:	ta,
		viewport:	vp,
		spinner:	sp,
		provider:	p,
		messages:	[]ai.Message{},
		helpView:	hv,
	}
}

func (m ChatModel) Init() tea.Cmd {
	return textarea.Blink
}

type errMsg error

func (m ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd	tea.Cmd
		vpCmd	tea.Cmd
		cmd	tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		headerHeight := 2
		footerHeight := 6

		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - headerHeight - footerHeight
		m.textarea.SetWidth(msg.Width - 4)
		m.ready = true

		m.helpView.Width = msg.Width - 6
		m.helpView.Height = msg.Height - 10

	case tea.MouseMsg:
		if m.showHelp {
			m.helpView, cmd = m.helpView.Update(msg)
			return m, cmd
		}

	case tea.KeyMsg:
		str := msg.String()

		if strings.Contains(str, "]11;") || strings.Contains(str, "rgb:") || strings.Contains(str, "\033") || strings.Contains(str, "\x1b") {
			return m, nil
		}

		if m.showHelp {
			switch msg.String() {
			case "esc", "?", "enter":
				m.showHelp = false
				return m, nil
			default:
				m.helpView, cmd = m.helpView.Update(msg)
				return m, cmd
			}
		}

		switch msg.Type {
		case tea.KeyRunes:
			if msg.String() == "?" {
				m.showHelp = true
				m.helpView.GotoTop()
				return m, nil
			}
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEsc:
			return m, func() tea.Msg { return BackMsg{} }
		case tea.KeyEnter:
			if m.loading {
				return m, nil
			}
			input := m.textarea.Value()
			if strings.TrimSpace(input) == "" {
				return m, nil
			}

			userMsg := ai.Message{Role: "user", Content: input}
			m.messages = append(m.messages, userMsg)
			m.renderMessages()

			m.textarea.Reset()
			m.loading = true

			return m, tea.Batch(m.spinner.Tick, m.sendToAI(m.messages))
		}

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case ai.Message:
		m.messages = append(m.messages, msg)
		m.loading = false
		m.renderMessages()
		return m, nil

	case errMsg:
		m.err = msg
		m.loading = false
		return m, nil
	}
	if !m.showHelp {
		m.textarea, tiCmd = m.textarea.Update(msg)
		m.viewport, vpCmd = m.viewport.Update(msg)
	}

	return m, tea.Batch(tiCmd, vpCmd, cmd)
}

func (m *ChatModel) renderMessages() {
	var sb strings.Builder

	userStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#90EE90")).
		Bold(true).
		Render

	aiContainerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF"))

	aiLabelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FFFF")).
		Bold(true)

	mdRenderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(m.width-10),
	)

	for _, msg := range m.messages {
		if msg.Role == "user" {

			content := fmt.Sprintf("You: %s", msg.Content)
			sb.WriteString(userStyle(content))
			sb.WriteString("\n\n")
		} else {

			rendered, err := mdRenderer.Render(msg.Content)
			if err != nil {
				rendered = msg.Content
			}

			label := aiLabelStyle.Render(m.provider.Name())
			sb.WriteString(label)
			sb.WriteString("\n")
			sb.WriteString(aiContainerStyle.Render(rendered))
			sb.WriteString("\n")
		}
	}

	m.viewport.SetContent(sb.String())
	m.viewport.GotoBottom()
}

func (m ChatModel) sendToAI(history []ai.Message) tea.Cmd {
	return func() tea.Msg {
		resp, err := m.provider.Send(history)
		if err != nil {
			return errMsg(err)
		}
		return ai.Message{Role: "assistant", Content: resp}
	}
}

func (m ChatModel) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	if m.showHelp {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			lipgloss.JoinVertical(lipgloss.Center,
				lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).MarginBottom(1).Render("AI Chat Help"),
				m.helpView.View(),
				lipgloss.NewStyle().Foreground(lipgloss.Color("240")).MarginTop(1).Render("Press [Esc] or [?] to go back"),
			),
		)
	}

	header := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#008069")).
		Bold(true).
		Render(fmt.Sprintf(" Devcli Chat :: %s (%s) ", m.provider.Name(), m.provider.Model()))

	chatView := m.viewport.View()

	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#AAAAAA")).
		Width(m.width-2).
		Padding(0, 1)

	var footerContent string
	if m.loading {
		footerContent = fmt.Sprintf("%s Generating response...", m.spinner.View())
	} else if m.err != nil {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")).Bold(true)
		helpHint := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(" [?] Help • [Esc] Quit")
		footerContent = fmt.Sprintf("%s\n%s\n%s", errStyle.Render("Error: "+m.err.Error()), m.textarea.View(), helpHint)
	} else {
		helpHint := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(" [?] Help • [Esc] Quit")
		footerContent = m.textarea.View() + "\n" + helpHint
	}

	footer := inputStyle.Render(footerContent)

	return lipgloss.JoinVertical(lipgloss.Left, header, chatView, footer)
}

func RunChat() {
	p := tea.NewProgram(Wrap(NewChatModel()), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running chat: %v\n", err)
		os.Exit(1)
	}
}

var ChatCmd = &cobra.Command{
	Use:	"chat",
	Short:	"Start AI chat session (TUI)",
	Run: func(cmd *cobra.Command, args []string) {
		RunChat()
	},
}
