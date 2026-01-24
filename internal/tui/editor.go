package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/phravins/devcli/internal/config"
	"github.com/phravins/devcli/internal/web"
	"github.com/phravins/devcli/pkg/utils"
	"github.com/spf13/cobra"
)

var EditorCmd = &cobra.Command{
	Use:   "editor [file]",
	Short: "Launch the built-in multi-language IDE",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filename := ""
		if len(args) > 0 {
			filename = args[0]
		}
		RunEditor(filename)
	},
}

func RunEditor(filename string) {
	p := tea.NewProgram(initialModel(filename), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running editor: %v\n", err)
		os.Exit(1)
	}
}

type blinkMsg struct{}

func blinkCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		return blinkMsg{}
	})
}

type sessionState int

const (
	stateSelection sessionState = iota
	stateEditor
	stateWebServer
	stateSavePrompt
	stateCommandPrompt
)

const (
	viewEditor = iota
	viewOutput
)

// Custom Editor Model
type editorModel struct {
	content string
	cursor  int // Linear index
	// We use the viewport for rendering
	viewport viewport.Model
}

type model struct {
	state    sessionState
	choices  []string
	cursor   int // Menu cursor
	filename string
	language string // New: explicitly track language mode

	// Custom Editor
	editor editorModel

	status         string
	showHelp       bool
	running        bool
	output         string
	saveInput      textinput.Model
	commandInput   string
	width          int
	height         int
	helpView       viewport.Model // New
	showCursorLine bool

	// Output View
	outputView      viewport.Model
	activeView      int // 0=Editor, 1=Output
	outputMaximized bool
	lastLanguage    string // Track for buffer clearing
}

func initialModel(filename string) model {
	ti := textinput.New()
	ti.Placeholder = "Enter path..."
	ti.CharLimit = 156
	ti.Width = 50

	initialContent := ""
	if filename != "" {
		if content, err := os.ReadFile(filename); err == nil {
			initialContent = string(content)
		}
	}

	// Output Viewport
	outVp := viewport.New(80, 10)

	vp := viewport.New(80, 20)

	// Help Viewport
	hv := viewport.New(80, 20)
	hv.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2)

	// Render Markdown Help
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	out, err := renderer.Render(EditorHelp)
	if err != nil {
		out = EditorHelp
	}
	hv.SetContent(out)

	startState := stateSelection
	if filename != "" {
		startState = stateEditor
	}

	return model{
		state:           startState,
		choices:         []string{"TUI Py (Python)", "TUI Java", "TUI C++", "TUI C", "TUI C#", "TUI Rust", "TUI Zig", "TUI G (Web Compiler)"},
		cursor:          0,
		filename:        filename,
		language:        detectLanguage(filename),
		editor:          editorModel{content: initialContent, cursor: 0, viewport: vp},
		status:          "Select an editor mode to begin",
		showHelp:        false,
		helpView:        hv,
		running:         false,
		output:          "",
		saveInput:       ti,
		width:           80,
		height:          40,
		outputView:      outVp,
		activeView:      viewEditor,
		outputMaximized: false,
	}
}

func (m *model) resolveExecutable(cmdName string, fallbacks []string) string {
	cacheKey := "compilers." + cmdName
	if cached := config.GetString(cacheKey); cached != "" {
		if utils.FileExists(cached) {
			return cached
		}
	}
	path := utils.FindExecutable(cmdName, fallbacks)
	if path != "" {
		config.SaveConfig(cacheKey, path)
		return path
	}
	userHome, _ := os.UserHomeDir()
	searchRoots := []string{
		`C:\Program Files`,
		`C:\Program Files (x86)`,
		filepath.Join(userHome, "Downloads"),
		`C:\`,
	}

	// Filter roots that exist
	validRoots := []string{}
	for _, r := range searchRoots {
		if utils.DirExists(r) {
			validRoots = append(validRoots, r)
		}
	}

	path = utils.DeepSearchExecutable(cmdName, validRoots)
	if path != "" {
		config.SaveConfig(cacheKey, path)
		return path
	}

	return ""
}

func highlightCode(code, language string) string {
	b := new(strings.Builder)
	// Map our internal lang names to Chroma lexers if needed, usually they match well
	lexer := language
	if lexer == "csharp" {
		lexer = "c#"
	}
	if lexer == "" || lexer == "text" {
		// Try to highlight as plain text or just return original if it fails
		lexer = "text"
	}
	err := quick.Highlight(b, code, lexer, "terminal256", "dracula")
	if err != nil {
		return code
	}
	return b.String()
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, blinkCmd())
}

func (m *model) updateLayout() {
	headerHeight := 4
	statusHeight := 3
	helpHeight := 0
	if m.showHelp {
		helpHeight = 7
	}

	availableHeight := m.height - headerHeight - statusHeight - helpHeight

	// Calculate Width
	width := m.width - 4
	if width < 20 {
		width = 20
	}

	// Calculate Heights
	if m.outputMaximized {
		// Output Maximized: Editor gets minimum, Output gets rest
		m.editor.viewport.Height = 5
		m.outputView.Height = availableHeight - 5
	} else if m.output != "" {
		// Split 50/50
		half := availableHeight / 2
		m.editor.viewport.Height = half
		m.outputView.Height = availableHeight - half
	} else {
		// Full Editor
		m.editor.viewport.Height = availableHeight
		m.outputView.Height = 0
	}

	m.editor.viewport.Width = width
	m.outputView.Width = width

	// Resize Help View

	// Resize Help View
	m.helpView.Width = m.width - 8
	m.helpView.Height = m.height - 4

	m.syncEditorView()
}

// Helper to highlight text AND insert a visual cursor
func (m *model) syncEditorView() {
	val := m.editor.content
	cursorPos := m.editor.cursor

	// Safety check for bounds
	if cursorPos > len(val) {
		cursorPos = len(val)
	}
	head := val[:cursorPos]
	tail := val[cursorPos:]
	currentLineIndex := strings.Count(head, "\n")
	cursorChar := "|"
	codeWithCursor := head + cursorChar + tail
	highlighted := highlightCode(codeWithCursor, m.language)

	rawLines := strings.Split(highlighted, "\n")
	var finalOutput strings.Builder
	lineNumStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4")) // Muted purple from theme

	vpWidth := m.editor.viewport.Width
	if vpWidth == 0 {
		vpWidth = 80 // Fallback
	}

	for i, line := range rawLines {
		// 1. Render Line Number with Margin
		var numStr string
		if i == currentLineIndex && m.showCursorLine {
			// Active Line: Yellow Bar
			numStr = fmt.Sprintf(" %s %3d ", cursorBarStyle.Render("|"), i+1)
		} else {
			// Inactive Line: Space instead of Bar
			// We render a space with the SAME style structure if needed, or just hardcode spaces
			numStr = fmt.Sprintf("   %3d ", i+1)
		}
		renderedNum := lineNumStyle.Render(numStr)

		// 2. Render Line Content
		// If this is the active line, apply the background style
		if i == currentLineIndex && m.showCursorLine {
			// We need to calculate the visible width of the line to pad it correctly
			// lipgloss.Width handles ANSI codes correctly
			contentWidth := lipgloss.Width(renderedNum) + lipgloss.Width(line)

			paddingNeeded := vpWidth - contentWidth
			if paddingNeeded < 0 {
				paddingNeeded = 0
			}

			// Construct the full line string
			fullLine := renderedNum + line + strings.Repeat(" ", paddingNeeded)

			// Apply the highlighter
			finalOutput.WriteString(cursorLineStyle.Render(fullLine))
		} else {
			finalOutput.WriteString(renderedNum)
			finalOutput.WriteString(line)
		}

		if i < len(rawLines)-1 {
			finalOutput.WriteString("\n")
		}
	}

	m.editor.viewport.SetContent(finalOutput.String())

	// Sync Scrolling (Keep cursor visible)
	viewportHeight := m.editor.viewport.Height
	currentOffset := m.editor.viewport.YOffset

	if currentLineIndex < currentOffset {
		m.editor.viewport.SetYOffset(currentLineIndex)
	} else if currentLineIndex >= currentOffset+viewportHeight {
		m.editor.viewport.SetYOffset(currentLineIndex - viewportHeight + 1)
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateLayout()

	case tea.MouseMsg:
		var cmd tea.Cmd
		if m.showHelp {
			m.helpView, cmd = m.helpView.Update(msg)
			return m, cmd
		}

		// Handle Editor Mode Selection Scroll
		if m.state == stateSelection {
			switch msg.Type {
			case tea.MouseWheelUp:
				if m.cursor > 0 {
					m.cursor--
				}
			case tea.MouseWheelDown:
				if m.cursor < len(m.choices)-1 {
					m.cursor++
				}
			}
			return m, nil
		}

		// Handle Output Scrolling if Focused
		if m.activeView == viewOutput {
			m.outputView, cmd = m.outputView.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		}

		// Only scroll editor viewport if we are IN the editor AND it is focused
		if m.state == stateEditor && m.activeView == viewEditor {
			m.editor.viewport, cmd = m.editor.viewport.Update(msg)
			cmds = append(cmds, cmd)
		}

	case tea.KeyMsg:
		// Global Shortcuts (Always active in Editor state)
		if m.state == stateEditor {
			switch msg.String() {
			case "ctrl+o":
				m.activeView = viewOutput
				m.status = "Focused: Output Terminal"
				m.updateLayout()
				return m, nil
			case "ctrl+e":
				m.activeView = viewEditor
				m.status = "Focused: Code Editor"
				m.updateLayout()
				return m, nil
			case "ctrl+m":
				if m.output != "" {
					m.outputMaximized = !m.outputMaximized
					m.updateLayout()
				}
				return m, nil
			}
		}

		if m.showHelp {
			switch msg.String() {
			case "esc", "ctrl+h", "?":
				m.showHelp = false
				m.updateLayout()
				return m, nil
			default:
				var cmd tea.Cmd
				m.helpView, cmd = m.helpView.Update(msg)
				return m, cmd
			}
		}

		switch m.state {
		case stateSelection:
			// Reset cursor visibility when selecting
			m.showCursorLine = true
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.choices)-1 {
					m.cursor++
				}
			case "enter":
				choice := m.choices[m.cursor]
				if strings.Contains(choice, "Web Compiler") {
					m.state = stateWebServer
					m.status = "Web Server Running..."
					go web.StartServer("8080")
					utils.OpenBrowser("http://127.0.0.1:8080")
				} else {
					m.state = stateEditor
					m.status = "Ready"
					// Set Language Mode based on selection
					newLang := ""
					switch {
					case strings.Contains(choice, "Py"):
						newLang = "python"
					case strings.Contains(choice, "Java"):
						newLang = "java"
					case strings.Contains(choice, "C++"):
						newLang = "cpp"
					case strings.Contains(choice, "C#"):
						newLang = "csharp"
					case strings.Contains(choice, "C"):
						if !strings.Contains(choice, "C++") && !strings.Contains(choice, "C#") {
							newLang = "c"
						}
					case strings.Contains(choice, "Rust"):
						newLang = "rust"
					case strings.Contains(choice, "Zig"):
						newLang = "zig"
					}

					// Buffer Isolation: Clear and inject boilerplate if switching languages on unsaved file
					if newLang != m.language && m.filename == "" {
						m.editor.content = getBoilerplate(newLang)
						m.editor.cursor = len(m.editor.content)
					}

					m.language = newLang
					m.updateLayout()
				}
			case "ctrl+c":
				return m, tea.Quit
			case "q", "esc":
				return m, func() tea.Msg { return BackMsg{} }
			case "?":
				m.showHelp = true
				m.helpView.GotoTop()
				m.updateLayout()
				return m, nil
			}

		case stateEditor:
			// Always show cursor line on input
			m.showCursorLine = true

			switch msg.Type {
			case tea.KeyCtrlC:
				return m, tea.Quit
			case tea.KeyEsc:
				// Go back to selection menu instead of exiting editor completely
				m.state = stateSelection
				m.status = "Select an editor mode to begin"
				m.updateLayout()
				return m, nil
			case tea.KeyCtrlS:
				m.state = stateSavePrompt
				if m.filename == "" {
					m.saveInput.SetValue("")
				} else {
					cwd, _ := os.Getwd()
					absPath, _ := filepath.Abs(m.filename)
					relPath, err := filepath.Rel(cwd, absPath)
					if err == nil && !strings.HasPrefix(relPath, "..") && !filepath.IsAbs(relPath) {
						m.saveInput.SetValue(relPath)
					} else {
						m.saveInput.SetValue(absPath)
					}
				}
				m.saveInput.Focus()
				m.status = "Enter filename (or full path) to save..."

			case tea.KeyCtrlR:
				if m.running {
					m.status = "Already running"
				} else {
					m.running = true
					m.status = fmt.Sprintf("Running %s code...", m.language)
					return m, m.runCode()
				}

			case tea.KeyCtrlH:
				m.showHelp = !m.showHelp
				m.helpView.GotoTop()
				m.updateLayout()

			case tea.KeyCtrlN:
				m.filename = ""
				m.editor.content = ""
				m.editor.cursor = 0
				m.syncEditorView()
				m.status = "New file created"

			case tea.KeyCtrlP:
				m.state = stateCommandPrompt
				m.status = "Enter shell command..."

			// Editor Input Handling
			case tea.KeyRunes:
				// Check for "?" key to toggle help
				if msg.String() == "?" && !m.running {
					m.showHelp = true
					m.helpView.GotoTop()
					m.updateLayout()
					return m, nil
				}

				val := m.editor.content
				pos := m.editor.cursor
				if pos > len(val) {
					pos = len(val)
				}

				char := msg.String()

				// Auto-closing logic
				var toInsert string
				var moveCursor int = 1

				switch char {
				case "{":
					toInsert = "{}"
				case "[":
					toInsert = "[]"
				case "(":
					toInsert = "()"
				case "\"":
					toInsert = "\"\""
				case "'":
					toInsert = "''"
				default:
					toInsert = char
				}

				m.editor.content = val[:pos] + toInsert + val[pos:]
				m.editor.cursor += moveCursor
				m.syncEditorView()

			case tea.KeySpace:
				val := m.editor.content
				pos := m.editor.cursor
				if pos > len(val) {
					pos = len(val)
				}
				m.editor.content = val[:pos] + " " + val[pos:]
				m.editor.cursor++
				m.syncEditorView()

			case tea.KeyEnter:
				val := m.editor.content
				pos := m.editor.cursor
				if pos > len(val) {
					pos = len(val)
				}

				// Smart Enter: Check if between brackets e.g. "{" | "}"
				isBetweenBraces := false
				if pos > 0 && pos < len(val) {
					prev := val[pos-1]
					next := val[pos]
					if prev == '{' && next == '}' {
						isBetweenBraces = true
					}
				}

				if isBetweenBraces {
					// Insert: \n    \n
					// Cursor: \n    | \n
					// Basic 4-space indentation
					indent := "    "
					toInsert := "\n" + indent + "\n"
					m.editor.content = val[:pos] + toInsert + val[pos:]
					m.editor.cursor += 1 + len(indent) // Move to indent position
				} else {
					m.editor.content = val[:pos] + "\n" + val[pos:]
					m.editor.cursor++
				}
				m.syncEditorView()

			case tea.KeyBackspace:
				val := m.editor.content
				pos := m.editor.cursor
				if pos > len(val) {
					pos = len(val)
				}

				if pos > 0 {
					// UTF-8 aware backspace
					// Decode the rune BEFORE the cursor
					r, size := utf8.DecodeLastRuneInString(val[:pos])
					if r != utf8.RuneError {
						m.editor.content = val[:pos-size] + val[pos:]
						m.editor.cursor -= size
						m.syncEditorView()
					}
				}

			case tea.KeyLeft:
				val := m.editor.content
				pos := m.editor.cursor
				if pos > 0 {
					_, size := utf8.DecodeLastRuneInString(val[:pos])
					m.editor.cursor -= size
					m.syncEditorView()
				}

			case tea.KeyRight:
				val := m.editor.content
				pos := m.editor.cursor
				if pos < len(val) {
					_, size := utf8.DecodeRuneInString(val[pos:])
					m.editor.cursor += size
					m.syncEditorView()
				}

			case tea.KeyUp, tea.KeyDown:
				m.moveCursorVertical(msg.Type)
				m.syncEditorView()
			}

		case stateSavePrompt:
			switch msg.Type {
			case tea.KeyEnter:
				filename := m.saveInput.Value()
				if filename != "" {
					m.filename = filename
					if err := os.WriteFile(m.filename, []byte(m.editor.content), 0644); err != nil {
						m.status = fmt.Sprintf("Error saving: %v", err)
					} else {
						m.status = fmt.Sprintf("Saved: %s", m.filename)
					}
					m.state = stateEditor
				}
			case tea.KeyEsc, tea.KeyCtrlC:
				m.saveInput.Reset()
				m.status = "Save cancelled"
				m.state = stateEditor
			}
			var cmd tea.Cmd
			m.saveInput, cmd = m.saveInput.Update(msg)
			cmds = append(cmds, cmd)

		case stateCommandPrompt:
			switch msg.Type {
			case tea.KeyEnter:
				if m.commandInput != "" {
					cmdStr := m.commandInput
					m.commandInput = ""
					m.status = "Running: " + cmdStr
					m.state = stateEditor
					return m, runShellCommand(cmdStr)
				}
			case tea.KeyEsc, tea.KeyCtrlC:
				m.commandInput = ""
				m.status = "Command cancelled"
				m.state = stateEditor
			case tea.KeyBackspace, tea.KeyDelete:
				if len(m.commandInput) > 0 {
					m.commandInput = m.commandInput[:len(m.commandInput)-1]
				}
			default:
				if msg.Type == tea.KeyRunes {
					m.commandInput += msg.String()
				}
			}
		case stateWebServer:
			// Allow quitting from web server mode
			switch msg.String() {
			case "ctrl+c", "q", "esc":
				return m, tea.Quit
			}
		}

	case blinkMsg:
		if m.state == stateEditor {
			m.showCursorLine = !m.showCursorLine
			m.syncEditorView()
		}
		return m, blinkCmd()

	case execResult:
		m.running = false
		m.output = msg.output
		m.outputView.SetContent(m.output) // Update viewport content
		m.activeView = viewOutput         // Auto-focus output
		m.outputView.GotoBottom()         // Auto-scroll to bottom

		if msg.err != nil {
			m.status = fmt.Sprintf("Error: %v", msg.err)
		} else {
			m.status = "Execution completed"
		}
		m.updateLayout()
		return m, nil
	}

	return m, tea.Batch(cmds...)
}

func (m *model) moveCursorVertical(key tea.KeyType) {
	// 1. Get lines
	lines := strings.Split(m.editor.content, "\n")

	// 2. Find visual row/col
	currentPos := 0
	currentRow := 0
	currentCol := 0 // byte offset in line

	found := false
	for r, line := range lines {
		lineLen := len(line) + 1 // newline
		if r == len(lines)-1 {
			lineLen--
		} // last line no newline

		if m.editor.cursor < currentPos+lineLen {
			currentRow = r
			currentCol = m.editor.cursor - currentPos
			found = true
			break
		}
		currentPos += lineLen
	}
	if !found {
		currentRow = len(lines) - 1
		currentCol = len(lines[currentRow])
	}

	// 3. Target Row
	targetRow := currentRow
	if key == tea.KeyUp {
		targetRow--
	} else {
		targetRow++
	}

	if targetRow < 0 {
		targetRow = 0
	}
	if targetRow >= len(lines) {
		targetRow = len(lines) - 1
	}

	// 4. Calculate new index
	targetLine := lines[targetRow]
	if currentCol > len(targetLine) {
		currentCol = len(targetLine) // Snap to end
	}

	// Sum length of previous lines
	newIndex := 0
	for i := 0; i < targetRow; i++ {
		newIndex += len(lines[i]) + 1
	}
	newIndex += currentCol

	m.editor.cursor = newIndex
}

// Styles
var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")). // Vivid Purple
			Padding(0, 1).
			Width(80)

	fileStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A8A8A8")). // Grey
			MarginLeft(1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#2E1065")). // Dark Purple
			Padding(0, 1)

	// Cursor Line Highlighting
	cursorLineStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#44475a")) // Dracula Selection Color

	// Vertical Bar Style (Yellow)
	cursorBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF00")). // Bright Yellow
			Bold(true)

	outputTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#0F9E99")). // Teal
				Bold(true)

	outputContentStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#0F9E99")). // Teal
				Padding(0, 1)

	// Selection Menu Styles
	selectionTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#EC4899")). // Pink-500
				Padding(1, 3).
				MarginBottom(1).
				Align(lipgloss.Center)

	selectionBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#EC4899")).
				Padding(1, 4).
				Margin(1, 0)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#EC4899")).
				Bold(true).
				Padding(0, 2).
				MarginLeft(1)

	unselectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#A1A1AA")). // Zinc-400
				PaddingLeft(3)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#71717A")). // Zinc-500
			Italic(true).
			MarginTop(1)
)

func (m model) View() string {
	if m.showHelp {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			lipgloss.JoinVertical(lipgloss.Center,
				lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).MarginBottom(1).Render("Editor Help"),
				m.helpView.View(),
				lipgloss.NewStyle().Foreground(lipgloss.Color("240")).MarginTop(1).Render("Press [Esc] or [?] to go back"),
			),
		)
	}

	if m.state == stateSelection {
		var choices strings.Builder

		for i, choice := range m.choices {
			if m.cursor == i {
				choices.WriteString(selectedItemStyle.Render("> "+choice) + "\n\n")
			} else {
				choices.WriteString(unselectedItemStyle.Render(choice) + "\n\n")
			}
		}

		title := selectionTitleStyle.Render("DEVCLI EDITOR")
		menuBox := selectionBoxStyle.Render(
			lipgloss.JoinVertical(lipgloss.Center,
				title,
				"\nChoose your development environment\n",
				choices.String(),
				helpStyle.Render("↑/↓: Navigate • Enter: Select • ?: Help • q: Back"),
			),
		)

		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, menuBox)
	}

	if m.state == stateWebServer {
		return fmt.Sprintf("\n=== TUI G (Web Compiler) ===\n\n" +
			"Server running at http://localhost:8080\n" +
			"The browser should have opened automatically.\n\n" +
			"Press Esc or Ctrl+C to stop server and exit.\n")
	}

	if m.state == stateSavePrompt {
		cwd, _ := os.Getwd()
		return fmt.Sprintf("\n=== Save As ===\n\n"+
			"Current Directory: %s\n"+
			"Enter filename/path: %s\n\n"+
			"Press Enter to save, Esc to cancel.", cwd, m.saveInput.View())
	}

	var s strings.Builder

	// Dynamic Header Config
	var title string
	var bgColor string

	switch m.language {
	case "python":
		title = "Python Mini-IDE (TUI Py)"
		bgColor = "#7D56F4" // Vivid Purple
	case "java":
		title = "Java IDE (TUI Java)"
		bgColor = "#b45309" // Amber/Orange
	case "cpp":
		title = "C++ IDE (TUI C++)"
		bgColor = "#0369a1" // Sky Blue
	case "c":
		title = "C IDE (TUI C)"
		bgColor = "#15803d" // Green
	case "csharp":
		title = "C# IDE (TUI C#)"
		bgColor = "#7e22ce" // Purple
	case "rust":
		title = "Rust IDE (TUI Rust)"
		bgColor = "#c2410c" // Rust Orange
	case "zig":
		title = "Zig IDE (TUI Zig)"
		bgColor = "#a21caf" // Fuchsia
	default:
		title = "Code Editor (Multi-Lang)"
		bgColor = "#44475a" // Muted Grey/Selection Color
	}

	// Header: Dynamic Title and Color
	header := headerStyle.Width(m.width).
		Background(lipgloss.Color(bgColor)).
		Render(title)

	fileInfo := fileStyle.Render(fmt.Sprintf("File: %s", m.filename))

	s.WriteString(header + "\n")
	s.WriteString(fileInfo + "\n\n")

	// Code Editor (Original Simple View)
	s.WriteString(m.editor.viewport.View())
	s.WriteString("\n")

	// Output section (Styled)
	if m.output != "" {
		cwd, _ := os.Getwd()
		title := fmt.Sprintf("Output (Executed in: %s) [Ctrl+E: Editor | Ctrl+M: Maximize]", cwd)

		// Change border color based on focus
		borderColor := "#0F9E99" // Teal (Default)
		if m.activeView == viewOutput {
			borderColor = "#FFFF00" // Yellow (Generic Focus)
			title = " >> " + title + " << "
		}

		outTitle := outputTitleStyle.Render(title)

		outView := m.outputView.View()
		outBox := outputContentStyle.
			Width(m.width - 2).
			BorderForeground(lipgloss.Color(borderColor)).
			Render(outView)

		s.WriteString("\n" + outTitle + "\n")
		s.WriteString(outBox + "\n")
	}

	// Status Bar
	// Calculate line number manually
	currentLine := strings.Count(m.editor.content[:m.editor.cursor], "\n") + 1

	statusText := fmt.Sprintf(" Status: %s | Line: %d ", m.status, currentLine)
	bar := statusStyle.Width(m.width).Render(statusText)

	s.WriteString("\n" + bar)

	return s.String()
}

// detectLanguage attempts to infer language from filename
func detectLanguage(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".py":
		return "python"
	case ".java":
		return "java"
	case ".cpp", ".cxx", ".cc":
		return "cpp"
	case ".c":
		return "c"
	case ".rs":
		return "rust"
	case ".zig":
		return "zig"
	case ".cs":
		return "csharp"
	case ".js":
		return "javascript"
	case ".ts":
		return "typescript"
	case ".html":
		return "html"
	case ".go":
		return "go"
	case ".json":
		return "json"
	case ".md":
		return "markdown"
	case ".h":
		return "c"
	default:
		return "text" // Default fallback
	}
}

// runCode dispatches execution based on language mode
func (m *model) runCode() tea.Cmd {
	code := m.editor.content
	language := m.language

	return func() tea.Msg {
		// SANITIZATION
		cleanCode := strings.Map(func(r rune) rune {
			if r == '\n' || r == '\t' {
				return r
			}
			if r < 32 || r == 127 {
				return -1
			}
			return r
		}, code)

		// Create a specific temp directory for this run to avoid collisions
		tmpDir, err := os.MkdirTemp("", "devcli_run_*")
		if err != nil {
			return execResult{"", fmt.Errorf("failed to create temp dir: %v", err)}
		}
		defer os.RemoveAll(tmpDir) // Cleanup everything after run

		var cmd *exec.Cmd

		switch language {
		case "python":
			tmpFile := filepath.Join(tmpDir, "script.py")
			if err := os.WriteFile(tmpFile, []byte(cleanCode), 0644); err != nil {
				return execResult{"", err}
			}

			pyFallbacks := []string{
				`C:\Python*\python.exe`,
				`C:\Program Files\Python*\python.exe`,
			}
			pyPath := m.resolveExecutable("python", pyFallbacks)
			if pyPath == "" {
				pyPath = m.resolveExecutable("python3", pyFallbacks)
			}

			if pyPath == "" {
				return execResult{"", fmt.Errorf("python not found. Please install Python or add to PATH")}
			}
			cmd = exec.Command(pyPath, "-u", tmpFile)

		case "java":
			// Attempt to find class name to name file correctly
			className := "Main"
			// Simple regex to find "public class X"
			lines := strings.Split(cleanCode, "\n")
			for _, line := range lines {
				if strings.Contains(line, "class ") {
					parts := strings.Fields(line)
					for i, p := range parts {
						if p == "class" && i+1 < len(parts) {
							// Strip braces if present
							name := strings.Trim(parts[i+1], "{")
							if name != "" {
								className = name
								break
							}
						}
					}
				}
			}
			srcFile := filepath.Join(tmpDir, className+".java")
			if err := os.WriteFile(srcFile, []byte(cleanCode), 0644); err != nil {
				return execResult{"", err}
			}

			// Find Compiler
			javaFallbacks := []string{
				`C:\Program Files\Java\jdk*\bin\java.exe`,
				`C:\Program Files\Eclipse Adoptium\jdk*\bin\java.exe`,
			}
			javacFallbacks := make([]string, len(javaFallbacks))
			for i, p := range javaFallbacks {
				javacFallbacks[i] = strings.Replace(p, "java.exe", "javac.exe", 1)
			}

			javaPath := m.resolveExecutable("java", javaFallbacks)
			javacPath := m.resolveExecutable("javac", javacFallbacks)

			if javacPath == "" || javaPath == "" {
				return execResult{"", fmt.Errorf("Java/Javac not found. Please install JDK or add to PATH")}
			}

			// Compile
			compileCmd := exec.Command(javacPath, "-d", ".", className+".java")
			compileCmd.Dir = tmpDir
			if out, err := compileCmd.CombinedOutput(); err != nil {
				return execResult{string(out), fmt.Errorf("compilation failed: %v", err)}
			}

			// Run
			cmd = exec.Command(javaPath, "-cp", ".", className)

		case "cpp":
			srcFile := filepath.Join(tmpDir, "main.cpp")
			exeFile := filepath.Join(tmpDir, "main.exe")
			if runtime.GOOS != "windows" {
				exeFile = filepath.Join(tmpDir, "main")
			}
			if err := os.WriteFile(srcFile, []byte(cleanCode), 0644); err != nil {
				return execResult{"", err}
			}

			// Find Compiler
			gppFallbacks := []string{
				`C:\Program Files\CodeBlocks\MinGW\bin\g++.exe`,
				`C:\Program Files (x86)\CodeBlocks\MinGW\bin\g++.exe`,
				`C:\MinGW\bin\g++.exe`,
				`C:\TDM-GCC-64\bin\g++.exe`,
			}
			gppPath := m.resolveExecutable("g++", gppFallbacks)
			if gppPath == "" {
				return execResult{"", fmt.Errorf("g++ compiler not found. Please install MinGW or add to PATH")}
			}

			// Compile
			compileCmd := exec.Command(gppPath, "main.cpp", "-o", exeFile)
			compileCmd.Dir = tmpDir
			if out, err := compileCmd.CombinedOutput(); err != nil {
				return execResult{string(out), fmt.Errorf("compilation failed: %v", err)}
			}

			// Run
			cmd = exec.Command(exeFile)

		case "c":
			srcFile := filepath.Join(tmpDir, "main.c")
			exeFile := filepath.Join(tmpDir, "main.exe")
			if runtime.GOOS != "windows" {
				exeFile = filepath.Join(tmpDir, "main")
			}
			if err := os.WriteFile(srcFile, []byte(cleanCode), 0644); err != nil {
				return execResult{"", err}
			}

			// Find Compiler
			gccFallbacks := []string{
				`C:\Program Files\CodeBlocks\MinGW\bin\gcc.exe`,
				`C:\Program Files (x86)\CodeBlocks\MinGW\bin\gcc.exe`,
				`C:\MinGW\bin\gcc.exe`,
				`C:\TDM-GCC-64\bin\gcc.exe`,
			}
			gccPath := m.resolveExecutable("gcc", gccFallbacks)
			if gccPath == "" {
				return execResult{"", fmt.Errorf("gcc compiler not found. Please install MinGW or add to PATH")}
			}

			// Compile
			compileCmd := exec.Command(gccPath, "main.c", "-o", exeFile)
			compileCmd.Dir = tmpDir
			if out, err := compileCmd.CombinedOutput(); err != nil {
				return execResult{string(out), fmt.Errorf("compilation failed: %v", err)}
			}

			// Run
			cmd = exec.Command(exeFile)

		case "rust":
			srcFile := filepath.Join(tmpDir, "main.rs")
			exeFile := filepath.Join(tmpDir, "main.exe")
			if runtime.GOOS != "windows" {
				exeFile = filepath.Join(tmpDir, "main")
			}
			if err := os.WriteFile(srcFile, []byte(cleanCode), 0644); err != nil {
				return execResult{"", err}
			}
			// Find Compiler
			userHome, _ := os.UserHomeDir()
			rustFallbacks := []string{
				filepath.Join(userHome, `.cargo\bin\rustc.exe`),
			}
			rustcPath := m.resolveExecutable("rustc", rustFallbacks)
			if rustcPath == "" {
				return execResult{"", fmt.Errorf("rustc not found. Please install Rust or add to PATH")}
			}

			// Compile
			compileCmd := exec.Command(rustcPath, "main.rs", "-o", exeFile)
			compileCmd.Dir = tmpDir
			if out, err := compileCmd.CombinedOutput(); err != nil {
				return execResult{string(out), fmt.Errorf("compilation failed: %v", err)}
			}

			// Run
			cmd = exec.Command(exeFile)

		case "zig":
			srcFile := filepath.Join(tmpDir, "main.zig")
			if err := os.WriteFile(srcFile, []byte(cleanCode), 0644); err != nil {
				return execResult{"", err}
			}
			// Find Zig
			zigFallbacks := []string{
				`C:\Program Files\Zig*\zig.exe`,
				`C:\zig*\zig.exe`,
			}
			zigPath := m.resolveExecutable("zig", zigFallbacks)
			if zigPath == "" {
				return execResult{"", fmt.Errorf("zig not found. Please install Zig or add to PATH")}
			}

			// zig run
			cmd = exec.Command(zigPath, "run", srcFile)

		case "csharp":
			// C# is tricky without a project. We will try to use 'dotnet-script' if available, or create a temp project.
			// Simplest robust way: dotnet new console, replace Program.cs, dotnet run.

			// 1. dotnet new console
			setupCmd := exec.Command("dotnet", "new", "console", "-o", tmpDir, "--force")
			if out, err := setupCmd.CombinedOutput(); err != nil {
				return execResult{string(out), fmt.Errorf("failed to init dotnet project: %v", err)}
			}

			// 2. Overwrite Program.cs
			mainFile := filepath.Join(tmpDir, "Program.cs")
			if err := os.WriteFile(mainFile, []byte(cleanCode), 0644); err != nil {
				return execResult{"", err}
			}

			// 3. dotnet run
			cmd = exec.Command("dotnet", "run", "--project", tmpDir)

		default:
			return execResult{"", fmt.Errorf("no runner defined for language: %s", language)}
		}

		cmd.Dir = tmpDir
		// If using 'cmd /C', we set dir to tmpDir so relative paths work?
		// Actually for compiled languages we generated commands assuming we are in tmpDir.

		output, err := cmd.CombinedOutput()
		outStr := string(output)

		if outStr == "" && err == nil {
			outStr = "[Success] (No output)"
		} else if err != nil && outStr == "" {
			outStr = fmt.Sprintf("[Error] %v", err)
		}

		return execResult{outStr, err}
	}
}

func runShellCommand(command string) tea.Cmd {
	return func() tea.Msg {
		cmd := utils.GetShellCommand(command)

		if cwd, err := os.Getwd(); err == nil {
			cmd.Dir = cwd
		}

		output, err := cmd.CombinedOutput()
		return execResult{string(output), err}
	}
}

func getBoilerplate(lang string) string {
	switch lang {
	case "python":
		return "print(\"Hello from Python!\")\n"
	case "java":
		return "public class Main {\n    public static void main(String[] args) {\n        System.out.println(\"Hello from Java!\");\n    }\n}\n"
	case "cpp":
		return "#include <iostream>\n\nint main() {\n    std::cout << \"Hello from C++!\" << std::endl;\n    return 0;\n}\n"
	case "c":
		return "#include <stdio.h>\n\nint main() {\n    printf(\"Hello from C!\\n\");\n    return 0;\n}\n"
	case "rust":
		return "fn main() {\n    println!(\"Hello from Rust!\");\n}\n"
	case "zig":
		return "const std = @import(\"std\");\n\npub fn main() !void {\n    std.debug.print(\"Hello from Zig!\\n\", .{});\n}\n"
	case "csharp":
		return "using System;\n\nclass Program {\n    static void Main() {\n        Console.WriteLine(\"Hello from C#!\");\n    }\n}\n"
	default:
		return ""
	}
}
