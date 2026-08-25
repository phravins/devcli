package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/phravins/devcli/internal/auth"
)

type AuthSuccessMsg struct {
	Username string
}

// ----------------------------------------------------
// 1. Auth Setup Model (First-Run Account Creation)
// ----------------------------------------------------

type AuthSetupModel struct {
	inputs     []textinput.Model
	focusedIdx int
	err        error
	width      int
	height     int
}

func NewAuthSetupModel() AuthSetupModel {
	inputs := make([]textinput.Model, 3)

	// Username
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "e.g. dev_user"
	inputs[0].Prompt = "Username         : "
	inputs[0].Focus()
	inputs[0].CharLimit = 32
	inputs[0].Width = 30

	// Password
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Min 8 chars (Upper, Lower, Number)"
	inputs[1].Prompt = "Master Password  : "
	inputs[1].EchoMode = textinput.EchoPassword
	inputs[1].CharLimit = 64
	inputs[1].Width = 30

	// Confirm Password
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Re-enter password"
	inputs[2].Prompt = "Confirm Password : "
	inputs[2].EchoMode = textinput.EchoPassword
	inputs[2].CharLimit = 64
	inputs[2].Width = 30

	return AuthSetupModel{
		inputs:     inputs,
		focusedIdx: 0,
		width:      80,
		height:     24,
	}
}

func (m AuthSetupModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m AuthSetupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "tab", "shift+tab", "down", "up":
			if msg.String() == "up" || msg.String() == "shift+tab" {
				m.focusedIdx--
			} else {
				m.focusedIdx++
			}

			if m.focusedIdx > len(m.inputs)-1 {
				m.focusedIdx = 0
			} else if m.focusedIdx < 0 {
				m.focusedIdx = len(m.inputs) - 1
			}

			for i := 0; i < len(m.inputs); i++ {
				if i == m.focusedIdx {
					m.inputs[i].Focus()
				} else {
					m.inputs[i].Blur()
				}
			}
			return m, nil

		case "enter":
			if m.focusedIdx < len(m.inputs)-1 {
				m.focusedIdx++
				for i := 0; i < len(m.inputs); i++ {
					if i == m.focusedIdx {
						m.inputs[i].Focus()
					} else {
						m.inputs[i].Blur()
					}
				}
				return m, nil
			}

			// Submit
			username := strings.TrimSpace(m.inputs[0].Value())
			password := m.inputs[1].Value()
			confirm := m.inputs[2].Value()

			if username == "" {
				m.err = fmt.Errorf("username cannot be empty")
				return m, nil
			}

			if password != confirm {
				m.err = fmt.Errorf("passwords do not match")
				return m, nil
			}

			if err := auth.ValidatePasswordStrength(password); err != nil {
				m.err = err
				return m, nil
			}

			if err := auth.SetupUser(username, password); err != nil {
				m.err = err
				return m, nil
			}

			return m, func() tea.Msg {
				return AuthSuccessMsg{Username: username}
			}
		}
	}

	for i := range m.inputs {
		var cmd tea.Cmd
		m.inputs[i], cmd = m.inputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m AuthSetupModel) View() string {
	card := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#0F9E99")).
		Padding(1, 3).
		Width(65)

	var b strings.Builder

	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#0F9E99")).
		Bold(true).
		Render("🔒 DevCLI Production Security - Setup Account")

	sub := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("First-time setup required: Create master credentials to secure your workspace")

	b.WriteString(title)
	b.WriteString("\n")
	b.WriteString(sub)
	b.WriteString("\n\n")

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		b.WriteString("\n\n")
	}

	btn := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255")).
		Background(lipgloss.Color("#0F9E99")).
		Padding(0, 3).
		Bold(true).
		Render("CREATE ACCOUNT (Enter)")

	b.WriteString("\n")
	b.WriteString(btn)
	b.WriteString("\n")

	if m.err != nil {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
		b.WriteString("\n❌ ")
		b.WriteString(errStyle.Render(m.err.Error()))
	}

	help := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("\nTab / Up / Down to navigate • Enter to submit")
	b.WriteString(help)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, card.Render(b.String()))
}

// ----------------------------------------------------
// 2. Auth Login Model (Password Lock Screen)
// ----------------------------------------------------

type AuthLoginModel struct {
	passwordInput textinput.Model
	err           error
	username      string
	width         int
	height        int
}

func NewAuthLoginModel() AuthLoginModel {
	data, _ := auth.GetAuthData()
	username := "Developer"
	if data != nil && data.Username != "" {
		username = data.Username
	}

	ti := textinput.New()
	ti.Placeholder = "Enter master password"
	ti.Prompt = "Password: "
	ti.EchoMode = textinput.EchoPassword
	ti.Focus()
	ti.CharLimit = 64
	ti.Width = 30

	return AuthLoginModel{
		passwordInput: ti,
		username:      username,
		width:         80,
		height:        24,
	}
}

func (m AuthLoginModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m AuthLoginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			pass := m.passwordInput.Value()
			valid, err := auth.VerifyPassword(pass)
			if err != nil || !valid {
				m.err = fmt.Errorf("invalid password, please try again")
				m.passwordInput.SetValue("")
				return m, nil
			}

			return m, func() tea.Msg {
				return AuthSuccessMsg{Username: m.username}
			}
		}
	}

	var cmd tea.Cmd
	m.passwordInput, cmd = m.passwordInput.Update(msg)
	return m, cmd
}

func (m AuthLoginModel) View() string {
	card := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#0F9E99")).
		Padding(1, 3).
		Width(60)

	var b strings.Builder

	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#0F9E99")).
		Bold(true).
		Render("🔒 DevCLI Locked - Authenticate Session")

	userLabel := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255")).
		Bold(true).
		Render(fmt.Sprintf("User: %s", m.username))

	b.WriteString(title)
	b.WriteString("\n\n")
	b.WriteString(userLabel)
	b.WriteString("\n\n")
	b.WriteString(m.passwordInput.View())
	b.WriteString("\n\n")

	btn := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255")).
		Background(lipgloss.Color("#0F9E99")).
		Padding(0, 3).
		Bold(true).
		Render("UNLOCK (Enter)")

	b.WriteString(btn)
	b.WriteString("\n")

	if m.err != nil {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
		b.WriteString("\n❌ ")
		b.WriteString(errStyle.Render(m.err.Error()))
	}

	help := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("\nEnter password to unlock workspace • Ctrl+C to exit")
	b.WriteString("\n")
	b.WriteString(help)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, card.Render(b.String()))
}
