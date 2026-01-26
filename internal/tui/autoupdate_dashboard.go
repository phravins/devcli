package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"github.com/phravins/devcli/internal/ai"
	"github.com/phravins/devcli/internal/ai/providers"
	"github.com/phravins/devcli/internal/config"
	"github.com/phravins/devcli/pkg/utils"
)

var autoUpdateMenuItems = []list.Item{
	item{title: "Check Language Versions", desc: "View installed versions of Go, Python, Node, etc."},
	item{title: "Update AI Keys", desc: "Update API keys for AI providers"},
	item{title: "Check DevCLI Updates", desc: "Check for new versions of DevCLI"},
}

const (
	StateAutoUpdateMenu = iota
	StateAutoUpdateLanguages
	StateAutoUpdateKeys
	StateAutoUpdateCheck       // Checking git
	StateAutoUpdateSummarizing // Generating AI summary
	StateAutoUpdateReview      // Reviewing AI Summary
	StateAutoUpdateInstalling
	StateAutoUpdateKeyInput
	StateAutoUpdateDone
	StateAutoUpdateHelp
)

type AutoUpdateModel struct {
	state         int
	list          list.Model
	width, height int
	spinner       spinner.Model
	input         textinput.Model
	keyProvider   string

	// Review / Output View
	outputView viewport.Model

	// Internal data
	updateLog     string // Raw git log
	updateSummary string // The AI generated summary
	provider      ai.Provider

	// Error handling
	err       error
	statusMsg string
}

// Msg types
type updateCheckMsg struct {
	hasUpdates bool
	log        string
	err        error
}

type summaryMsg struct {
	content string
	err     error
}

type installMsg struct {
	err error
}

func showMainMenu(m *AutoUpdateModel) {
	m.list.SetItems(autoUpdateMenuItems)
	m.list.Title = "Auto-Update Center"
	m.state = StateAutoUpdateMenu
}

func showKeyProviderMenu(m *AutoUpdateModel) {
	items := []list.Item{
		item{title: "Google Gemini", desc: "Update Gemini API Key"},
		item{title: "OpenAI", desc: "Update OpenAI / ChatGPT API Key"},
		item{title: "Anthropic Claude", desc: "Update Claude API Key"},
		item{title: "Ollama", desc: "No key required usually, but can set base URL"},
		item{title: "HuggingFace", desc: "Update HF Access Token"},
	}
	m.list.SetItems(items)
	m.list.Title = "Select AI Provider"
	m.state = StateAutoUpdateKeys
}

func NewAutoUpdateModel() AutoUpdateModel {
	l := list.New(autoUpdateMenuItems, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Auto-Update Center"
	l.SetShowTitle(true)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	ti := textinput.New()
	ti.Placeholder = "Enter API Key..."
	ti.CharLimit = 100
	ti.Width = 50

	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2)

	// Initialize provider (Copied from ChatModel)
	cfg, _ := config.LoadConfig()
	var p ai.Provider

	backend := strings.TrimSpace(strings.ToLower(cfg.AIBackend))
	if backend == "" {
		backend = "ollama"
	}

	switch backend {
	case "ollama":
		p = &providers.OllamaProvider{}
	case "huggingface":
		p = &providers.HFProvider{}
	case "local":
		p = &providers.LocalHFProvider{}
	case "claude", "anthropic":
		p = &providers.AnthropicProvider{}
	case "gemini", "google":
		p = &providers.GeminiProvider{}
	case "mistral":
		p = &providers.OpenAIProvider{BaseURL: "https://api.mistral.ai/v1"}
	case "kimi", "moonshot":
		p = &providers.OpenAIProvider{BaseURL: "https://api.moonshot.cn/v1"}
	case "groq":
		p = &providers.OpenAIProvider{BaseURL: "https://api.groq.com/openai/v1"}
	case "deepseek":
		p = &providers.OpenAIProvider{BaseURL: "https://api.deepseek.com/v1"}
	case "lmstudio":
		p = &providers.OpenAIProvider{}
	default:
		p = &providers.OpenAIProvider{}
	}

	if err := p.Configure(cfg); err != nil {
		fmt.Printf("Error configuring provider: %v\n", err)
	}

	return AutoUpdateModel{
		state:      StateAutoUpdateMenu,
		list:       l,
		spinner:    s,
		outputView: vp,
		provider:   p,
		input:      ti,
	}
}

func (m AutoUpdateModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, textinput.Blink)
}

func (m AutoUpdateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v-10) // Leave space for header/footer

		// Resize Output View
		m.outputView.Width = msg.Width - 4
		m.outputView.Height = msg.Height - 10

	case tea.MouseMsg:
		if msg.Type == tea.MouseWheelUp {
			if m.state == StateAutoUpdateMenu || m.state == StateAutoUpdateKeys {
				m.list.CursorUp()
			} else {
				m.outputView.LineUp(3)
			}
			return m, nil
		}
		if msg.Type == tea.MouseWheelDown {
			if m.state == StateAutoUpdateMenu || m.state == StateAutoUpdateKeys {
				m.list.CursorDown()
			} else {
				m.outputView.LineDown(3)
			}
			return m, nil
		}

	case tea.KeyMsg:
		if m.state == StateAutoUpdateMenu {
			switch msg.String() {
			case "esc":
				return m, func() tea.Msg { return BackMsg{} }
			case "enter":
				i, ok := m.list.SelectedItem().(item)
				if ok {
					switch i.title {
					case "Check Language Versions":
						m.state = StateAutoUpdateLanguages
						m.statusMsg = "Checking versions..."
						return m, tea.Batch(m.spinner.Tick, checkLanguageVersionsCmd())
					case "Update AI Keys":
						showKeyProviderMenu(&m)
						return m, nil
					case "Check DevCLI Updates":
						m.state = StateAutoUpdateCheck
						m.statusMsg = "Checking for updates..."
						return m, tea.Batch(m.spinner.Tick, checkDevCLIUpdatesCmd())
					}
				}
			case "?":
				showHelp(&m)
				return m, nil
			}
		} else if m.state == StateAutoUpdateKeys {
			switch msg.String() {
			case "esc":
				showMainMenu(&m)
				return m, nil
			case "enter":
				i, ok := m.list.SelectedItem().(item)
				if ok {
					m.keyProvider = i.title
					m.state = StateAutoUpdateKeyInput
					m.input.Reset()
					m.input.Placeholder = fmt.Sprintf("Enter API Key for %s", i.title)
					m.input.Focus()
					return m, textinput.Blink
				}
			}
		} else if m.state == StateAutoUpdateKeyInput {
			switch msg.String() {
			case "esc":
				m.input.Blur()
				showKeyProviderMenu(&m)
				return m, nil
			case "enter":
				key := m.input.Value()
				m.input.Blur()
				// Save key
				saveKeyCmd(m.keyProvider, key)
				m.statusMsg = fmt.Sprintf("Updated key for %s!", m.keyProvider)
				m.state = StateAutoUpdateDone
				return m, nil
			}
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd

		} else if m.state == StateAutoUpdateReview {
			// In review mode, handle confirmation or cancel
			switch msg.String() {
			case "y", "Y":
				m.state = StateAutoUpdateInstalling
				m.statusMsg = "Updating DevCLI..."
				return m, tea.Batch(m.spinner.Tick, installDevCLIUpdatesCmd())
			case "n", "N", "esc":
				m.state = StateAutoUpdateMenu
				return m, nil
			}
		} else if m.state == StateAutoUpdateDone || m.err != nil {
			if msg.String() == "esc" || msg.String() == "enter" {
				m.state = StateAutoUpdateMenu
				m.err = nil
				m.statusMsg = ""
				return m, nil
			}
		} else if m.state == StateAutoUpdateHelp {
			if msg.String() == "esc" || msg.String() == "q" {
				m.state = StateAutoUpdateMenu
				return m, nil
			}
		} else {
			// Helper to cancel any operation (e.g. Languages check)
			if msg.String() == "esc" {
				m.state = StateAutoUpdateMenu
				return m, nil
			}
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case updateCheckMsg:
		if msg.err != nil {
			m.err = msg.err
			m.state = StateAutoUpdateDone // Show error state
		} else if !msg.hasUpdates {
			m.statusMsg = "DevCLI is up to date!"
			m.state = StateAutoUpdateDone
		} else {
			m.updateLog = msg.log
			m.state = StateAutoUpdateSummarizing
			m.statusMsg = "Found updates! Generating AI summary..."
			return m, tea.Batch(m.spinner.Tick, summarizeUpdatesCmd(m.provider, msg.log))
		}

	case summaryMsg:
		switch m.state {
		case StateAutoUpdateCheck, StateAutoUpdateSummarizing:
			// This was from the AI summary
			if msg.err != nil {
				// Fallback to raw log if AI fails
				m.updateSummary = "Failed to generate AI summary. Raw logs:\n" + m.updateLog
			} else {
				m.updateSummary = msg.content
			}
			m.state = StateAutoUpdateReview

			// Render markdown
			renderer, _ := glamour.NewTermRenderer(
				glamour.WithAutoStyle(),
				glamour.WithWordWrap(m.width-10),
			)
			out, err := renderer.Render(m.updateSummary)
			if err != nil {
				out = m.updateSummary
			}
			m.outputView.SetContent(out)

		case StateAutoUpdateLanguages:
			// This was from language check
			m.updateSummary = msg.content
			// We manually styled this with lipgloss, so bypass glamour
			m.outputView.SetContent(m.updateSummary)
		}

	case installMsg:
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.statusMsg = "Update Complete! Please restart DevCLI."
		}
		m.state = StateAutoUpdateDone
	}

	// Update list only in menu or keys select
	if m.state == StateAutoUpdateMenu || m.state == StateAutoUpdateKeys {
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Update viewport in review/done/langs/help
	if m.state == StateAutoUpdateReview || m.state == StateAutoUpdateDone || m.state == StateAutoUpdateLanguages || m.state == StateAutoUpdateHelp {
		m.outputView, cmd = m.outputView.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m AutoUpdateModel) View() string {
	switch m.state {
	case StateAutoUpdateMenu, StateAutoUpdateKeys:
		view := m.list.View()
		footer := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("\nPress [?] for Help")
		return docStyle.Render(view + footer)

	case StateAutoUpdateHelp:
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			lipgloss.JoinVertical(lipgloss.Center,
				lipgloss.NewStyle().
					Foreground(lipgloss.Color("205")).
					Bold(true).
					MarginBottom(1).
					Render("Auto-Update Help"),
				m.outputView.View(),
				lipgloss.NewStyle().
					Foreground(lipgloss.Color("240")).
					MarginTop(1).
					Render("Press [Esc] or [q] to go back [Wheel to Scroll]"),
			),
		)

	case StateAutoUpdateKeyInput:
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			lipgloss.JoinVertical(lipgloss.Center,
				lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).Render("Update API Key"),
				fmt.Sprintf("\nProvider: %s\n", m.keyProvider),
				m.input.View(),
				"\nPress [Enter] to Save • [Esc] to Cancel",
			),
		)

	case StateAutoUpdateLanguages:
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			lipgloss.JoinVertical(lipgloss.Center,
				lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).Render("Programming Language Versions"),
				m.outputView.View(),
				"\nPress [Esc] to go back",
			),
		)

	case StateAutoUpdateCheck, StateAutoUpdateSummarizing, StateAutoUpdateInstalling:
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			lipgloss.JoinVertical(lipgloss.Center,
				m.spinner.View(),
				"\n"+m.statusMsg,
			),
		)

	case StateAutoUpdateReview:
		header := lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).Render("New Updates Available!")
		footer := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Press [y] to Install • [n] to Cancel")

		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			lipgloss.JoinVertical(lipgloss.Center,
				header,
				m.outputView.View(),
				footer,
			),
		)

	case StateAutoUpdateDone:
		if m.err != nil {
			return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
				lipgloss.JoinVertical(lipgloss.Center,
					lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true).Render(" Error"),
					m.err.Error(),
					"\nPress [Esc] to go back",
				),
			)
		}
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			lipgloss.JoinVertical(lipgloss.Center,
				lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true).Render(" "+m.statusMsg),
				"\nPress [Esc] to go back",
			),
		)

	default:
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, "Feature coming soon...")
	}
}

func showHelp(m *AutoUpdateModel) {
	var sb strings.Builder

	sb.WriteString("# Auto-Update Features Guide\n\n")

	sb.WriteString("## 1. Language Version Check\n")
	sb.WriteString("Scans your system for installed programming languages like **Go, Python, Node, Java, Rust, Zig, and C/C++**. It checks common installation paths and your system's PATH variable to provide version info and absolute paths.\n\n")

	sb.WriteString("## 2. AI API Key Management\n")
	sb.WriteString("Allows you to securely update API keys for various AI providers (**Gemini, OpenAI, Claude, HuggingFace**). Keys are saved locally in the DevCLI configuration file for use in the AI Assistant and Chat features.\n\n")

	sb.WriteString("## 3. DevCLI Self-Update\n")
	sb.WriteString("Checks the official DevCLI repository for updates. If updates are found, it:\n")
	sb.WriteString("- **Pulls** the latest changes via Git.\n")
	sb.WriteString("- **Generates** an AI-powered summary of the release notes.\n")
	sb.WriteString("- **Rebuilds** the DevCLI executable automatically if you confirm.\n\n")

	sb.WriteString("---\n")
	sb.WriteString("### How to Use\n")
	sb.WriteString("- **Arrow Keys / Mouse Wheel**: Navigate menus and scroll content.\n")
	sb.WriteString("- **Enter**: Select an option or confirm.\n")
	sb.WriteString("- **?**: Open this help screen from the main menu.\n")
	sb.WriteString("- **Esc / q**: Go back or cancel operations.\n")

	m.updateSummary = sb.String()

	// Render markdown
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(m.width-12),
	)
	out, err := renderer.Render(m.updateSummary)
	if err != nil {
		out = m.updateSummary
	}
	m.outputView.SetContent(out)
	m.outputView.YOffset = 0 // Reset scroll
	m.state = StateAutoUpdateHelp
}

// Commands

func checkLanguageVersionsCmd() tea.Cmd {
	return func() tea.Msg {
		var sb strings.Builder

		pinky := lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
		header := lipgloss.NewStyle().Foreground(lipgloss.Color("62")).Bold(true).Render("# Installed Languages & Tools")

		sb.WriteString(header + "\n\n")

		check := func(name, cmdName string, args []string, fallbacks []string) {
			path := utils.FindExecutable(cmdName, fallbacks)

			pathStr := path
			if pathStr == "" {
				pathStr = "Not Found (Install or add to PATH)"
			}

			// Add Language Name Header
			nameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Bold(true)
			sb.WriteString(fmt.Sprintf("%s\n", nameStyle.Render("## "+name)))

			if path != "" {
				vCmd := exec.Command(path, args...)
				vOut, err := vCmd.CombinedOutput()
				if err == nil {
					// Clean version string
					outStr := string(vOut)
					lines := strings.Split(outStr, "\n")
					version := strings.TrimSpace(lines[0])
					sb.WriteString(fmt.Sprintf("• Version: %s\n", pinky.Render(version)))
				} else {
					sb.WriteString(fmt.Sprintf("• Version: %s\n", pinky.Render("Detected but version check failed")))
				}
			} else {
				sb.WriteString(fmt.Sprintf("• Version: %s\n", pinky.Render("Unknown (Check triggered error)")))
			}
			sb.WriteString(fmt.Sprintf("• Path:    %s\n", pinky.Render(pathStr)))
			sb.WriteString("\n")
		}

		// Fallback Definitions
		userHome, _ := os.UserHomeDir()

		// 1. Go
		check("Go", "go", []string{"version"}, []string{
			`C:\Program Files\Go\bin\go.exe`,
			`C:\Go\bin\go.exe`,
		})

		// 2. Python
		check("Python", "python", []string{"--version"}, []string{
			`C:\Python*\python.exe`,
			`C:\Program Files\Python*\python.exe`,
			filepath.Join(userHome, `AppData\Local\Programs\Python\Python*\python.exe`),
		})

		// 3. Node.js
		check("Node.js", "node", []string{"--version"}, []string{
			`C:\Program Files\nodejs\node.exe`,
		})

		// 4. Java
		check("Java", "java", []string{"-version"}, []string{
			`C:\Program Files\Java\jdk*\bin\java.exe`,
			`C:\Program Files\Eclipse Adoptium\jdk*\bin\java.exe`,
		})

		// 5. Rust
		check("Rust", "rustc", []string{"--version"}, []string{
			filepath.Join(userHome, `.cargo\bin\rustc.exe`),
		})

		// 6. Zig
		check("Zig", "zig", []string{"version"}, []string{
			`C:\Program Files\Zig*\zig.exe`,
			`C:\zig*\zig.exe`,
		})

		// 7. C/C++ (GCC) - Special focus on Code::Blocks/MinGW
		gccFallbacks := []string{
			`C:\Program Files\CodeBlocks\MinGW\bin\gcc.exe`,
			`C:\Program Files (x86)\CodeBlocks\MinGW\bin\gcc.exe`,
			`C:\MinGW\bin\gcc.exe`,
			`C:\TDM-GCC-64\bin\gcc.exe`,
		}
		check("C (GCC)", "gcc", []string{"--version"}, gccFallbacks)

		// G++ usually in same place as gcc, reuse fallbacks looking for g++
		gppFallbacks := make([]string, len(gccFallbacks))
		for i, p := range gccFallbacks {
			gppFallbacks[i] = strings.Replace(p, "gcc.exe", "g++.exe", 1)
		}
		check("C++ (G++)", "g++", []string{"--version"}, gppFallbacks)

		noteStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
		sb.WriteString(noteStyle.Render("> Note: Checked system PATH and common installation directories."))

		return summaryMsg{content: sb.String()}
	}
}

func checkDevCLIUpdatesCmd() tea.Cmd {
	return func() tea.Msg {
		// 0. Check if git repo
		if _, err := exec.Command("git", "rev-parse", "--is-inside-work-tree").Output(); err != nil {
			return updateCheckMsg{err: fmt.Errorf("not a git repository. Please initialize git to use this feature")}
		}

		// 1. Fetch
		fetch := exec.Command("git", "fetch")
		if output, err := fetch.CombinedOutput(); err != nil {
			// Check for common errors
			outStr := string(output)
			if strings.Contains(outStr, "fatal: no remote") {
				return updateCheckMsg{err: fmt.Errorf("no git remote configured")}
			}
			return updateCheckMsg{err: fmt.Errorf("git fetch failed: %s", outStr)}
		}
		branchCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
		branchOut, err := branchCmd.Output()
		if err != nil {
			return updateCheckMsg{err: fmt.Errorf("git rev-parse failed: %w", err)}
		}
		branch := strings.TrimSpace(string(branchOut))

		// Log
		logCmd := exec.Command("git", "log", fmt.Sprintf("HEAD..origin/%s", branch), "--oneline")
		out, err := logCmd.Output()
		if err != nil {
			// Maybe no upstream configured?
			return updateCheckMsg{err: fmt.Errorf("check failed (no upstream for branch '%s'?)", branch)}
		}

		logStr := strings.TrimSpace(string(out))
		if logStr == "" {
			return updateCheckMsg{hasUpdates: false}
		}

		return updateCheckMsg{hasUpdates: true, log: logStr}
	}
}

func summarizeUpdatesCmd(p ai.Provider, log string) tea.Cmd {
	return func() tea.Msg {
		if p == nil {
			return summaryMsg{err: fmt.Errorf("no AI provider configured")}
		}

		prompt := fmt.Sprintf("Visualize these git commit logs into a nice, human-readable release note summary. Highlight new features and fixes. Keep it concise.\n\nLogs:\n%s", log)

		msgs := []ai.Message{{Role: "user", Content: prompt}}
		resp, err := p.Send(msgs)
		return summaryMsg{content: resp, err: err}
	}
}

func installDevCLIUpdatesCmd() tea.Cmd {
	return func() tea.Msg {
		// Get current branch
		branchCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
		branchOut, err := branchCmd.Output()
		if err != nil {
			return installMsg{err: fmt.Errorf("failed to get current branch: %w", err)}
		}
		branch := strings.TrimSpace(string(branchOut))

		// Check if there are uncommitted changes
		statusCmd := exec.Command("git", "status", "--porcelain")
		statusOut, err := statusCmd.Output()
		hasChanges := err == nil && len(statusOut) > 0

		// Stash changes if any exist
		if hasChanges {
			stash := exec.Command("git", "stash", "push", "-m", "DevCLI auto-update backup")
			if output, err := stash.CombinedOutput(); err != nil {
				return installMsg{err: fmt.Errorf("git stash failed: %s", string(output))}
			}
		}

		// git pull with explicit remote and branch
		pull := exec.Command("git", "pull", "origin", branch)
		if output, err := pull.CombinedOutput(); err != nil {
			// If pull fails, restore stashed changes
			if hasChanges {
				exec.Command("git", "stash", "pop").Run()
			}
			return installMsg{err: fmt.Errorf("git pull failed: %s", string(output))}
		}

		// Restore stashed changes after successful pull
		if hasChanges {
			pop := exec.Command("git", "stash", "pop")
			if output, err := pop.CombinedOutput(); err != nil {
				// Don't fail the update if stash pop has conflicts, just warn
				return installMsg{err: fmt.Errorf("update succeeded but stash restore had conflicts: %s\nYour changes are in 'git stash list'", string(output))}
			}
		}

		// go build
		// Assuming we run this from the project root or we can find it
		build := exec.Command("go", "build", "-o", "devcli.exe", ".")
		if output, err := build.CombinedOutput(); err != nil {
			return installMsg{err: fmt.Errorf("go build failed: %s", string(output))}
		}

		return installMsg{err: nil}
	}
}

func saveKeyCmd(provider, key string) {
	// We run this synchronously as it is fast
	key = strings.TrimSpace(key)

	switch strings.ToLower(provider) {
	case "google gemini":
		config.Set("gemini_api_key", key)
	case "openai":
		config.Set("ai_api_key", key) // Default usually
	case "anthropic claude":
		config.Set("anthropic_api_key", key)
	case "huggingface":
		config.Set("hf_access_token", key)
	}
	// Also set the main key if generic
	if strings.ToLower(provider) == "openai" {
		config.Set("ai_api_key", key)
	}

	config.Write()
}
