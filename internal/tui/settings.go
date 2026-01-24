package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"github.com/phravins/devcli/internal/config"
)

type SettingsModel struct {
	inputs     []textinput.Model
	focusedIdx int
	err        error
	successMsg string
	quitting   bool
	showHelp   bool
	width      int
	height     int
	helpView   viewport.Model
	mainView   viewport.Model
}

func NewSettingsModel() SettingsModel {
	cfg, _ := config.LoadConfig()

	inputs := make([]textinput.Model, 4)

	// AI Backend
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "ollama / gemini / openai / claude"
	inputs[0].Focus()
	inputs[0].Prompt = "AI Backend: "
	inputs[0].SetValue(cfg.AIBackend)
	inputs[0].CharLimit = 30
	inputs[0].Width = 30

	// Model Name
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "gemini-1.5-flash / gpt-3.5-turbo"
	inputs[1].Prompt = "AI Model: "
	inputs[1].SetValue(cfg.AIModel)
	inputs[1].CharLimit = 50
	inputs[1].Width = 30

	// API Key
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "sk-..."
	inputs[2].Prompt = "API Key: "
	inputs[2].EchoMode = textinput.EchoPassword

	// Pre-fill key based on backend
	currentKey := cfg.AIAPIKey
	switch strings.ToLower(cfg.AIBackend) {
	case "huggingface":
		if cfg.HFAccessToken != "" {
			currentKey = cfg.HFAccessToken
		}
	case "gemini", "google":
		if cfg.GeminiAPIKey != "" {
			currentKey = cfg.GeminiAPIKey
		}
	}
	inputs[2].SetValue(currentKey)
	inputs[2].CharLimit = 100
	inputs[2].Width = 30

	// Base URL
	inputs[3] = textinput.New()
	inputs[3].Placeholder = "Optional (e.g. http://localhost:1234/v1)"
	inputs[3].Prompt = "Base URL: "
	inputs[3].SetValue(cfg.AIBaseURL)
	inputs[3].CharLimit = 100
	inputs[3].Width = 50

	// Help Viewport
	hv := viewport.New(100, 40)
	hv.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2)

	// Render Markdown Help
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	out, err := renderer.Render(SettingsHelp)
	if err != nil {
		out = SettingsHelp
	}
	hv.SetContent(out)

	// Main Viewport
	mv := viewport.New(100, 40)
	mv.Style = lipgloss.NewStyle().Padding(1, 2)

	m := SettingsModel{
		inputs:     inputs,
		focusedIdx: 0,
		width:      100,
		height:     40,
		helpView:   hv,
		mainView:   mv,
	}
	m.updateMainViewContent()
	return m
}

func (m SettingsModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m SettingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.helpView.Width = msg.Width - 6
		m.helpView.Height = msg.Height - 10
		m.mainView.Width = msg.Width
		m.mainView.Height = msg.Height
		m.updateMainViewContent()
		return m, nil

	case tea.KeyMsg:
		// Help screen handler
		if m.showHelp {
			switch msg.String() {
			case "esc", "?", "enter":
				m.showHelp = false
				return m, nil
			default:
				var cmd tea.Cmd
				m.helpView, cmd = m.helpView.Update(msg)
				return m, cmd
			}
		}

		switch msg.String() {
		case "?":
			m.showHelp = true
			m.helpView.GotoTop()
			return m, nil
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit // Return to main dashboard logic
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			// Enter on last field = Save
			if s == "enter" && m.focusedIdx == len(m.inputs)-1 {
				m.saveConfig()
				m.updateMainViewContent() // Show success/error
				return m, nil
			}

			// Navigation
			if s == "up" || s == "shift+tab" {
				m.focusedIdx--
			} else {
				m.focusedIdx++
			}

			if m.focusedIdx > len(m.inputs)-1 {
				m.focusedIdx = 0
			} else if m.focusedIdx < 0 {
				m.focusedIdx = len(m.inputs) - 1
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i < len(m.inputs); i++ {
				if i == m.focusedIdx {
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205")) // Pink
					m.inputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
				} else {
					m.inputs[i].Blur()
					m.inputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
					m.inputs[i].TextStyle = lipgloss.NewStyle()
				}
			}
			m.updateMainViewContent() // CRITICAL: Update view to show new focus
			return m, tea.Batch(cmds...)
		}

	case tea.MouseMsg:
		if m.showHelp {
			var cmd tea.Cmd
			m.helpView, cmd = m.helpView.Update(msg)
			return m, cmd
		}

		// Fixed: Mouse wheel now changes focus instead of scrolling viewport
		switch msg.Type {
		case tea.MouseWheelUp:
			m.focusedIdx--
			if m.focusedIdx < 0 {
				m.focusedIdx = len(m.inputs) - 1
			}
		case tea.MouseWheelDown:
			m.focusedIdx++
			if m.focusedIdx > len(m.inputs)-1 {
				m.focusedIdx = 0
			}
		}

		// Update focus state based on new index
		for i := 0; i < len(m.inputs); i++ {
			if i == m.focusedIdx {
				m.inputs[i].Focus()
				m.inputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
				m.inputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
			} else {
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
				m.inputs[i].TextStyle = lipgloss.NewStyle()
			}
		}

		m.updateMainViewContent() // CRITICAL: Update view to show new focus
		return m, nil
	}

	// Handle Input Updates
	inputCmd := m.updateInputs(msg)
	cmds = append(cmds, inputCmd)

	return m, tea.Batch(cmds...)
}

func (m *SettingsModel) updateInputs(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd = make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	m.updateMainViewContent() // Update content whenever inputs change
	return tea.Batch(cmds...)
}

func (m *SettingsModel) updateMainViewContent() {
	// Create a centralized card style
	card := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")). // Purple/Blurple
		Padding(1, 3).
		Width(60).
		Align(lipgloss.Left)

	var b strings.Builder

	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("212")). // Pink
		Bold(true).
		Render("CONFIGURATION")

	b.WriteString(lipgloss.NewStyle().Align(lipgloss.Center).Width(54).Render(title))
	b.WriteString("\n\n")

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteString("\n\n") // More spacing
		}
	}

	// Button Logic
	buttonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255")).
		Background(lipgloss.Color("62")).
		Padding(0, 3).
		Bold(true)

	inactiveButton := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		Render("Submit (Enter)")

	button := "\n\n"
	if m.focusedIdx == len(m.inputs)-1 {
		// Active Button
		button += lipgloss.PlaceHorizontal(54, lipgloss.Center, buttonStyle.Render("SAVE CHANGES"))
	} else {
		button += lipgloss.PlaceHorizontal(54, lipgloss.Center, inactiveButton)
	}
	b.WriteString(button)

	if m.successMsg != "" {
		b.WriteString("\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Align(lipgloss.Center).Width(54).Render(m.successMsg))
	}
	if m.err != nil {
		b.WriteString("\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Align(lipgloss.Center).Width(54).Render(m.err.Error()))
	}

	help := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Align(lipgloss.Center).Width(54).Render("Esc to Cancel • Tab to Navigate • [?] Help")
	b.WriteString("\n\n" + help)

	// Wrap everything in a nice centered box
	view := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		card.Render(b.String()),
	)

	m.mainView.SetContent(view)
}

func (m *SettingsModel) saveConfig() {
	if err := m.validateInputs(); err != nil {
		m.err = err
		m.successMsg = ""
		return
	}

	config.Set("ai_backend", strings.TrimSpace(m.inputs[0].Value()))
	config.Set("ai_model", strings.TrimSpace(m.inputs[1].Value()))

	apiKey := strings.TrimSpace(m.inputs[2].Value())
	config.Set("ai_api_key", apiKey) // Set default/active key

	// Also save to specific provider keys for persistence
	backend := strings.ToLower(strings.TrimSpace(m.inputs[0].Value()))
	switch backend {
	case "huggingface":
		config.Set("hf_access_token", apiKey)
	case "gemini", "google":
		config.Set("gemini_api_key", apiKey)
	}

	config.Set("ai_base_url", strings.TrimSpace(m.inputs[3].Value()))

	if err := config.Write(); err != nil {
		m.err = err
		m.successMsg = ""
	} else {
		m.successMsg = "Configuration Saved Successfully!"
		m.err = nil
	}
}

func (m *SettingsModel) validateInputs() error {
	backend := strings.ToLower(strings.TrimSpace(m.inputs[0].Value()))
	apiKey := strings.TrimSpace(m.inputs[2].Value())

	if backend == "" {
		return fmt.Errorf("backend cannot be empty")
	}

	// List of backends that require an API key
	needsKey := []string{"openai", "gemini", "google", "claude", "anthropic", "mistral", "groq", "huggingface", "kimi"}

	for _, b := range needsKey {
		if backend == b && apiKey == "" {
			return fmt.Errorf("API Key is required for %s", b)
		}
	}
	// Base URL validation
	baseURL := strings.TrimSpace(m.inputs[3].Value())
	if baseURL != "" {
		if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
			return fmt.Errorf("base URL must start with http:// or https://")
		}
	}

	return nil
}

func (m SettingsModel) View() string {
	if m.quitting {
		return ""
	}

	// Show help screen
	// Show help screen
	if m.showHelp {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			lipgloss.JoinVertical(lipgloss.Center,
				lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).MarginBottom(1).Render("Settings Help"),
				m.helpView.View(),
				lipgloss.NewStyle().Foreground(lipgloss.Color("240")).MarginTop(1).Render("Press [Esc] or [?] to go back"),
			),
		)
	}

	// Return the viewport view instead of the raw string
	if m.mainView.Width == 0 {
		m.updateMainViewContent() // Fallback init
	}
	return m.mainView.View()
}

// Wrap for standalone run if needed, but we will call from dashboard
func RunSettings() {
	p := tea.NewProgram(NewSettingsModel())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error:", err)
	}
}
