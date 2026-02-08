package tui

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type DocsModel struct {
	viewport viewport.Model
	width    int
	height   int
	ready    bool
}

func NewDocsModel() DocsModel {
	return DocsModel{
		width:  100, // Default, will be resized
		height: 30,
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
		// Initialize viewport on first view if not ready (or wait for Resize)
		// But better to simulate a resize or just return empty/loading
		return "Loading Docs..."
	}

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#0F9E99")).
		Padding(1, 0).
		Render("DevCLI Documentation")

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Padding(1, 0).
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

	if !m.ready {
		// First time initialization
		m.viewport = viewport.New(m.width, m.height-verticalMarginHeight)
		m.viewport.SetContent(getDocsContent())
		m.ready = true
	} else {
		m.viewport.Width = m.width
		m.viewport.Height = m.height - verticalMarginHeight
	}
}

func getDocsContent() string {
	return `
# DevCLI - Developer's Command Line Interface

DevCLI is a comprehensive terminal-based toolkit designed to streamline your development workflow. It integrates essential tools, AI assistance, project management, and more into a single, unified interface.

## 🚀 Key Features

### 1. Project Management
- **Create Projects**: Instantly generate boilerplate for Go, Python, Node.js, and more.
- **Manage Projects**: Track active projects, view history, and backup important work.
- **Bonus Features**: Task Runner, Smart File Creator, and Snippet Library.

### 2. AI Assistance
- **Chat**: Interact with top AI models (Ollama, Gemini, OpenAI, Claude).
- **Context Aware**: The AI knows about your project structure and languages.
- **Agents**: Specialized agents for Code Generation, Architecture, and Debugging.

### 3. Integrated Tools
- **Editor**: A lightweight, terminal-based IDE with syntax highlighting.
- **File Manager**: Navigate your file system, manage files, and search globally.
- **Dev Server**: Auto-detect and run local dev servers (npm start, go run, etc.) with log filtering.
- **Git Time Machine**: Visualize code history and blame in an interactive TUI.

### 4. Configuration & Updates
- **Settings**: Easily configure API keys and preferences.
- **Auto-Update**: Keep DevCLI and your language runtimes up to date.

## ⌨️ Global Shortcuts

| Key | Action |
| :--- | :--- |
| **Ctrl+C** | Quit Application (Force Exit) |
| **Esc / Q** | Go Back / Previous Menu |
| **Up / Down** | Navigate Menus & Lists |
| **Enter** | Select Item / Confirm |
| **?** | Toggle Context-Specific Help |

## 🛠️ Getting Started

1.  **Explore the Dashboard**: Use Arrow keys to navigate the main menu.
2.  **Configure AI**: Go to **Settings** to set up your preferred AI provider (e.g., free local Ollama or Gemini API).
3.  **Start a Project**: Use **Project Tools** to scaffold a new application.
4.  **Need Help?**: Press **?** in any specific tool to see its usage guide.

## 📦 Supported Languages & Stacks

- **Go**: Full support for modules, building, and web servers.
- **Python**: venv management, pip requirements, Flask/Django/FastAPI.
- **JavaScript/TypeScript**: Node.js, React, Vue, Vite, Express.
- **Rust, Java, C/C++**: Basic project templates and compilation support.

---
*DevCLI v1.0.0 - Built for efficiency.*
`
}
