package tui

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

type DocsModel struct {
	viewport viewport.Model
	width    int
	height   int
	ready    bool
	renderer *glamour.TermRenderer
}

func NewDocsModel() DocsModel {
	// Initialize glamour renderer
	r, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(100), // Default, will be updated on resize
	)

	return DocsModel{
		width:    100, // Default, will be resized
		height:   30,
		renderer: r,
	}
}

func (m DocsModel) Init() tea.Cmd {
	return nil
}

func (m DocsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, func() tea.Msg { return BackMsg{} }
		case "ctrl+c":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizeViewport()
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m DocsModel) View() string {
	if !m.ready {
		return "Loading Docs..."
	}

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#0F9E99")).
		Padding(1, 0).
		Align(lipgloss.Center).
		Width(m.width).
		Render("DevCLI Documentation")

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Padding(1, 0).
		Align(lipgloss.Center).
		Width(m.width).
		Render("Esc/Q: Back • ↑/↓: Scroll")

	return lipgloss.JoinVertical(
		lipgloss.Center,
		header,
		m.viewport.View(),
		footer,
	)
}

func (m *DocsModel) resizeViewport() {
	headerHeight := 4 // Header + Padding
	footerHeight := 3 // Footer + Padding
	verticalMarginHeight := headerHeight + footerHeight

	// Re-create renderer with new width
	contentWidth := m.width - 4 // Padding
	if contentWidth < 20 {
		contentWidth = 20
	}

	m.renderer, _ = glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(contentWidth),
	)

	renderedContent, err := m.renderer.Render(getDocsContent())
	if err != nil {
		renderedContent = "Error rendering documentation: " + err.Error()
	}

	if !m.ready {
		// First time initialization
		m.viewport = viewport.New(m.width, m.height-verticalMarginHeight)
		m.viewport.SetContent(renderedContent)
		m.ready = true
	} else {
		m.viewport.Width = m.width
		m.viewport.Height = m.height - verticalMarginHeight
		m.viewport.SetContent(renderedContent) // Re-render content for wrapping
	}
}

func getDocsContent() string {
	return `
# DevCLI - Unified Developer Workspace

DevCLI is a terminal-based power tool designed to consolidate your entire development workflow into a single, keyboard-driven interface. It replaces scattered scripts and context switching with a unified dashboard for project management, coding, and AI assistance.

> **Philosophy**: "Stay in the flow." DevCLI brings your tools to you, right in your terminal.

---

## 🚀 Key Features

### 1. Project Management
*   **Project Dashboard**: Get a bird's-eye view of all your projects (status, tech stack, last modified).
*   **One-Click Scaffolding**: Create production-ready projects in Go, Python, Node.js, React, and more.
*   **Task Runner**: Auto-detects ` + "`package.json`" + `, ` + "`Makefile`" + `, ` + "`go.mod`" + `, etc., and lets you run build/test commands instantly.
*   **Smart File Creator**: Generate ` + "`.gitignore`" + `, ` + "`Dockerfile`" + `, ` + "`README.md`" + `, or CI/CD configs in seconds.

### 2. Development Environment
*   **Dev Server Launcher**: Automatically detects your web framework (Next.js, Flask, Django) and launches the dev server with live log streaming.
*   **Virtual Environment Wizard**: Centralized management for Python ` + "`venvs`" + ` and Node ` + "`node_modules`" + `. scan, sync, and clean up to save disk space.
*   **Built-in Editor**: A lightweight, nano-like editor with syntax highlighting for quick edits without leaving DevCLI.
*   **File Manager**: A fully functional file explorer with fuzzy search and file operations.

### 3. AI & Analysis
*   **AI Assistant**: Chat with LLMs (Ollama, OpenAI, Claude, Gemini) directly in your terminal. Context-aware code generation and debugging.
*   **Code Time Machine**: A visual interface for Git history. Step through commits, see blame annotations, and get AI-powered bug risk analysis.
*   **Snippet Library**: Save and organize your favorite code blocks for instant reuse.

---

## ⚙️ Configuration

DevCLI stores its configuration in ` + "`~/.devcli/config.yaml`" + ` (or ` + "`%USERPROFILE%\\.devcli\\config.yaml`" + ` on Windows).

### AI Providers
You can configure multiple AI backends. Go to **Settings** in the main menu or edit the config file directly.

` + "```yaml" + `
ai:
  provider: "ollama" # or "openai", "anthropic", "gemini"
  model: "llama3"    # Model name
  api_key: ""        # Required for cloud providers
  base_url: ""       # Optional custom endpoint
` + "```" + `

### Customizing Styles
DevCLI supports themes. Currently, it defaults to an adaptive theme based on your terminal's background color.

---

## ⌨️ Global Shortcuts

| Key | Action |
| :--- | :--- |
| **Ctrl+C** | Quit Application |
| **Esc / Q** | Go Back / Close View |
| **Arrow Keys** | Navigate Menus & Lists |
| **Enter** | Select / Confirm |
| **?** | Show Help (Context Sensitive) |
| **Ctrl+L** | Clear Screen / Redraw |

---

## ❓ FAQ & Troubleshooting

**Q: DevCLI doesn't detect my project type.**
A: Ensure your project has standard marker files like ` + "`package.json`" + `, ` + "`go.mod`" + `, ` + "`requirements.txt`" + `, or ` + "`pom.xml`" + `.

**Q: The AI Assistant isn't working.**
A: 
1. Check your internet connection.
2. If using **Ollama**, ensure the service is running (` + "`ollama serve`" + `).
3. If using **OpenAI/Claude**, verify your API key in **Settings**.

**Q: How do I update DevCLI?**
A: Go to **Bonus Features** -> **Check for Updates**. DevCLI can self-update by pulling the latest code and rebuilding itself.

---

## 🤝 Contributing

DevCLI is open source! We welcome contributions.
*   **Repo**: https://github.com/phravins/devcli
*   **Issues**: Report bugs or request features on GitHub issues.

*Built with ❤️ using Go, Bubble Tea, and Lip Gloss.*
`
}
