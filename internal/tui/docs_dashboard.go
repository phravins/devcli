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

DevCLI is a terminal-based development workspace that consolidates essential developer tools into a single unified interface. It manages projects, files, virtual environments, and provides AI-powered assistance without requiring you to leave the command line.

The application is built using Go and the Bubble Tea framework, providing a fast and responsive terminal user interface that works across all major operating systems.

## 🚀 Core Features

DevCLI serves as a central hub for common development tasks, avoiding the need for scattered scripts.

### 1. Project Management
- **Project Manager**: Scaffolding, templates, and history tracking.
- **Task Runner**: One-click execution of build, test, and lint commands (Go, Python, Node, Rust, C++).
- **Boilerplate Generator**: Instant code snippets and architectural patterns.
- **Smart File Creator**: Instant generation of Dockerfiles, .env, Makefiles, and CI/CD configs.

### 2. Development Tools
- **Virtual Environment Wizard**: Centralized management of Python venvs and Node modules.
- **Dev Server**: Auto-detecting live reload servers for web development.
- **File Manager & Editor**: Keyboard-driven filesystem navigation and quick editing.
- **Auto-Update System**: Keeps your languages and tools current.

### 3. AI & Analysis
- **AI Assistant**: Built-in chat for coding help, debugging, and explanations.
- **Code Time Machine**: Git-powered code evolution tracker with bug detection and blame visualization.
- **Snippet Library**: Your personal vault for reusable code blocks.

---

## 📦 Installation

**Note**: When installed via 'go install', dependencies are managed automatically.

### Method 1: Automated (Windows)
1. Download 'setup_devcli.bat'.
2. Right-click -> "Run as administrator".
3. Automatically installs Go, DevCLI, and sets PATH.

### Method 2: Automated (Linux/macOS)
1. Download 'install.sh'.
2. Run:
   'chmod +x install.sh'
   './install.sh'

### Method 3: Go Install
If Go is already installed:
'go install github.com/phravins/devcli@latest'

### Method 4: Build from Source
'git clone https://github.com/phravins/devcli.git'
'cd devcli'
'go build -o devcli.exe .'

---

## 🛠️ Usage

### Interactive Mode
Launch the main dashboard:
'devcli'

Use **Arrow Keys** to navigate, **Enter** to select, and **Esc/Q** to go back.

### Direct Subcommands
- 'devcli dev'          # Open project management tools
- 'devcli file'         # Launch file manager
- 'devcli ai'           # Start AI chat session
- 'devcli editor FILE'  # Open file in built-in editor

---

## 🔍 Feature Deep Dive

### Project Creation & Management
Create projects from predefined templates (Go, Python, Node.js, React, etc.). Features smart naming, history tracking, backups, and automatic dependency installation (npm install, pip install).

### Virtual Environment Wizard
Manage Python venvs and Node modules globally. Recursively scans workspace, syncs requirements.txt, clones environments, and cleans up unused deps.

### Dev Server Launcher
Intelligent server manager. Detects framework (Next.js, Flask, Go, etc.) and runs the dev server. Features live log streaming, filtering, and search.

### Code Time Machine
Interactive Git history.
- **Time Travel**: Step through commits to see code evolution.
- **Bug Radar**: formatting detects risky commits (late-night, large refactors).
- **Blame**: Line-by-line author tracking.

### AI Integration
Chat with Ollama, Gemini, OpenAI, Claude. Supports system prompts, context-aware suggestions, and works offline with local models.

### Built-in Editor
Lightweight editor for quick fixes. Syntax highlighting for Python/Go, direct execution (Ctrl+R), and basic IDE features.

---
*DevCLI - The unified workspace for modern developers.*
`
}
