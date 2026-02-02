package tui

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
)

type FileManagerModel struct {
	currentPath string
	files       []fs.DirEntry
	filtered    []fs.DirEntry

	cursor      int
	width       int
	height      int
	quitting    bool
	searchInput textinput.Model
	err         error

	selectedFile string

	// Move Implementation
	moveMode        bool
	moveInput       textinput.Model
	selectedForMove string

	// Copy Implementation
	copyMode        bool
	copyInput       textinput.Model
	selectedForCopy string

	// Path Edit Implementation
	pathMode  bool
	pathInput textinput.Model

	// Search Cache
	allFilePaths []string

	// Navigation History
	history []string

	// Global Search
	globalSearch bool

	// Loading State
	loading bool

	// Performance
	searchID int

	// Concurrency
	scanChan chan string

	// Layout
	ready bool

	// Help
	// Help
	showHelp bool
	helpView viewport.Model // New
}

type searchDebounceMsg struct {
	id int
}

type filterFinishedMsg struct {
	results []fs.DirEntry
}

// Async Search Command
func performSearchCmd(paths []string, query string) tea.Cmd {
	return func() tea.Msg {
		if query == "" {
			// Special case: usually handled before calling this, but safe fallback
			return filterFinishedMsg{results: nil}
		}

		// Cap results to prevent UI lag on huge result sets
		const maxResults = 1000
		var matches []string

		useFastPath := len(paths) > 5000

		lowerQuery := strings.ToLower(query)

		if useFastPath {
			for _, path := range paths {
				if len(matches) >= maxResults {
					break
				}
				if strings.Contains(strings.ToLower(path), lowerQuery) {
					matches = append(matches, path)
				}
			}
		} else {
			// Fuzzy match for smaller sets
			fuzzyMatches := fuzzy.Find(query, paths)
			for _, m := range fuzzyMatches {
				if len(matches) >= maxResults {
					break
				}
				matches = append(matches, m.Str)
			}
		}

		var results []fs.DirEntry
		for _, matchPath := range matches {
			results = append(results, dummyEntry{path: matchPath})
		}
		return filterFinishedMsg{results: results}
	}
}

// Get available drives (Windows specific simplified)
func getDrives() []string {
	drives := []string{}
	for _, drive := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		path := string(drive) + ":\\"
		_, err := os.Stat(path)
		if err == nil {
			drives = append(drives, path)
		}
	}
	return drives
}

func NewFileManagerModel(startPath string) FileManagerModel {
	if startPath == "" {
		startPath, _ = os.Getwd()
	}

	ti := textinput.New()
	ti.Placeholder = "Type to search ALL DRIVES... (Scanning)"
	ti.CharLimit = 156
	ti.Width = 60
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("33"))
	ti.Prompt = "> "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("86"))   // Cyan
	ti.Cursor.Style = lipgloss.NewStyle().Background(lipgloss.Color("255")) // White Block Cursor
	ti.Cursor.Style = lipgloss.NewStyle().Background(lipgloss.Color("255")) // White Block Cursor
	ti.Focus()                                                              // Ensure focused at start

	// Help Viewport
	hv := viewport.New(80, 20)
	hv.Style = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("62")).Padding(1, 2)

	// Render Markdown Help
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	out, err := renderer.Render(FileManagerHelp)
	if err != nil {
		out = FileManagerHelp
	}
	hv.SetContent(out)

	mi := textinput.New()
	mi.Placeholder = "New path/name..."
	mi.CharLimit = 256
	mi.Width = 50

	ci := textinput.New()
	ci.Placeholder = "Copy to new path/name..."
	ci.CharLimit = 256
	ci.Width = 50

	pi := textinput.New()
	pi.Placeholder = "/path/to/folder"
	pi.CharLimit = 256
	pi.Width = 60
	pi.SetValue(startPath)

	m := FileManagerModel{
		currentPath:  startPath,
		searchInput:  ti,
		moveInput:    mi,
		copyInput:    ci,
		pathInput:    pi,
		globalSearch: true, // Default to Global
		loading:      true, // Start loading
		scanChan:     make(chan string, 1000),
		// width/height default to 0, waiting for WindowSizeMsg
		helpView: hv,
	}

	// Pre-load current directory recursively so search works immediately for local files
	m.reloadAllFiles() // This fills m.allFilePaths with local files first

	m.loadFiles()
	return m
}

// Msg when scanning starts
type scanStartedMsg struct{}

// Msg for incremental results
type searchResultMsg struct {
	paths []string
}

// Msg when scanning is complete
type scanFinishedMsg struct{}

// Command to start background scanning
func startGlobalScanCmd(ch chan string) tea.Cmd {
	return func() tea.Msg {
		go func() {
			drives := getDrives()
			var wg sync.WaitGroup

			for _, drive := range drives {
				wg.Add(1)
				go func(d string) {
					defer wg.Done()
					filepath.WalkDir(d, func(path string, de fs.DirEntry, err error) error {
						if err != nil {
							if de != nil && de.IsDir() {
								return filepath.SkipDir
							}
							return nil
						}
						// Non-blocking send or block? If buffer full, we block.
						ch <- path
						return nil
					})
				}(drive)
			}

			wg.Wait()
			close(ch)
		}()
		return scanStartedMsg{}
	}
}

// Command to listen for results (Batched with Time Buffer)
func waitForSearchResults(ch chan string) tea.Cmd {
	return func() tea.Msg {
		var batch []string
		const maxBatch = 5000
		const batchTimeout = 200 * time.Millisecond

		// 1. Blocking wait for at least one item
		path, ok := <-ch
		if !ok {
			return scanFinishedMsg{}
		}
		batch = append(batch, path)

		// 2. Try to collect more items
		timer := time.NewTimer(batchTimeout)
		defer timer.Stop()

	loop:
		for len(batch) < maxBatch {
			select {
			case p, open := <-ch:
				if !open {
					break loop
				}
				batch = append(batch, p)
			case <-timer.C:
				break loop
			}
		}

		return searchResultMsg{paths: batch}
	}
}

func (m FileManagerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	// Start listening when file load starts
	case scanStartedMsg:
		m.loading = true
		return m, waitForSearchResults(m.scanChan)

	// Handle Streamed Result
	case searchResultMsg:
		m.allFilePaths = append(m.allFilePaths, msg.paths...)

		// Performance Optimization: Incremental Filter
		// Use simple substring match for real-time updates to avoid lag.
		// Fuzzy is too slow for high-frequency updates.
		if m.searchInput.Value() != "" {
			query := strings.ToLower(m.searchInput.Value())
			for _, p := range msg.paths {
				if strings.Contains(strings.ToLower(p), query) {
					m.filtered = append(m.filtered, dummyEntry{path: p})
				}
			}
		}
		return m, waitForSearchResults(m.scanChan)

	case scanFinishedMsg:
		m.loading = false
		m.searchInput.Placeholder = fmt.Sprintf("Search %d files across all drives...", len(m.allFilePaths))
		if m.searchInput.Value() == "" {
			return m, nil
		}
		return m, performSearchCmd(m.allFilePaths, m.searchInput.Value())

	case searchDebounceMsg:
		if msg.id == m.searchID {
			m.searchID++
			return m, performSearchCmd(m.allFilePaths, m.searchInput.Value())
		}
		return m, nil

	case filterFinishedMsg:
		m.filtered = msg.results
		m.cursor = 0
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.height = msg.Height
		m.ready = true

		// Resize Help View
		m.helpView.Width = msg.Width - 6
		m.helpView.Height = msg.Height - 10
		return m, nil

	case tea.MouseMsg:
		if m.showHelp {
			var cmd tea.Cmd
			m.helpView, cmd = m.helpView.Update(msg)
			return m, cmd
		}

		switch msg.Type {
		case tea.MouseWheelUp:
			if m.cursor > 0 {
				m.cursor -= 3 // Scroll 3 lines
				if m.cursor < 0 {
					m.cursor = 0
				}
			}
		case tea.MouseWheelDown:
			if m.cursor < len(m.filtered)-1 {
				m.cursor += 3 // Scroll 3 lines
				if m.cursor >= len(m.filtered) {
					m.cursor = len(m.filtered) - 1
				}
			}
		case tea.MouseLeft:
			// Calculate layout metrics (Must match View)
			headerHeight := 3 // Border(2) + Input(1)
			spacerHeight := 1
			footerHeight := 2
			listStartY := headerHeight + spacerHeight

			// Available height for list
			listHeight := m.height - headerHeight - footerHeight - 2
			if listHeight < 1 {
				listHeight = 1
			}

			// Check if click is within list area
			// msg.Y is 0-indexed row
			if msg.Y >= listStartY && msg.Y < listStartY+listHeight {
				// Calculate Scroll Offset
				start := 0
				if m.cursor >= listHeight {
					start = m.cursor - listHeight + 1
				}

				// Determine clicked index
				clickOffset := msg.Y - listStartY
				clickedIndex := start + clickOffset

				if clickedIndex >= 0 && clickedIndex < len(m.filtered) {
					m.cursor = clickedIndex

					// Trigger Select Action (Same as KeyEnter)
					selected := m.filtered[m.cursor]
					pathName := selected.Name()
					var fullPath string
					if filepath.IsAbs(pathName) {
						fullPath = pathName
					} else {
						fullPath = filepath.Join(m.currentPath, pathName)
					}

					info, err := os.Stat(fullPath)
					isDir := false
					if err == nil && info.IsDir() {
						isDir = true
					} else if selected.IsDir() {
						isDir = true
					}

					if isDir {
						m.history = append(m.history, m.currentPath)
						m.pathInput.SetValue(fullPath) // Update path input
						m.currentPath = fullPath
						m.searchInput.Reset()
						m.globalSearch = false
						m.loadFiles()
						m.cursor = 0
					} else {
						m.selectedFile = fullPath
						// Switch to Editor
						return m, func() tea.Msg { return SwitchViewMsg{TargetState: StateEditor, Args: fullPath} }
					}
				}
			}
		}
		// Since we changed cursor, update pagination if needed? View handles it.
		return m, nil

	case tea.KeyMsg:
		// Modal Inputs (Move/Copy Prompt)
		if m.moveMode {
			switch msg.Type {
			case tea.KeyEnter:
				newName := m.moveInput.Value()
				if newName != "" && m.selectedForMove != "" {
					oldPath := filepath.Join(m.currentPath, m.selectedForMove)
					newPath := filepath.Join(m.currentPath, newName)
					if filepath.IsAbs(newName) {
						newPath = newName
					}

					if err := os.Rename(oldPath, newPath); err != nil {
						if strings.Contains(err.Error(), "cross-device link") || strings.Contains(err.Error(), "different drive") {
							if cpErr := copyFile(oldPath, newPath); cpErr != nil {
								m.err = fmt.Errorf("move failed: %w", cpErr)
							} else {
								if rmErr := os.RemoveAll(oldPath); rmErr != nil {
									m.err = fmt.Errorf("move completed but failed to delete original: %w", rmErr)
								} else {
									m.err = nil
									m.loadFiles()
								}
							}
						} else {
							m.err = err
						}
					} else {
						m.err = nil
						m.loadFiles()
					}
				}
				m.moveMode = false
				m.moveInput.Blur()
				return m, nil

			case tea.KeyEsc:
				m.moveMode = false
				m.moveInput.Blur()
				m.err = nil
				return m, nil
			}
			m.moveInput, cmd = m.moveInput.Update(msg)
			return m, cmd
		}

		if m.copyMode {
			switch msg.Type {
			case tea.KeyEnter:
				newName := m.copyInput.Value()
				if newName != "" && m.selectedForCopy != "" {
					oldPath := filepath.Join(m.currentPath, m.selectedForCopy)
					newPath := filepath.Join(m.currentPath, newName)
					if filepath.IsAbs(newName) {
						newPath = newName
					}

					if err := copyFile(oldPath, newPath); err != nil {
						m.err = err
					} else {
						m.err = nil
						m.loadFiles()
					}
				}
				m.copyMode = false
				m.copyInput.Blur()
				return m, nil

			case tea.KeyEsc:
				m.copyMode = false
				m.copyInput.Blur()
				return m, nil
			}
			m.copyInput, cmd = m.copyInput.Update(msg)
			return m, cmd
		}

		if m.pathMode {
			switch msg.Type {
			case tea.KeyEnter:
				newPath := m.pathInput.Value()
				if newPath != "" {
					// Verify path exists and is dir
					info, err := os.Stat(newPath)
					if err == nil && info.IsDir() {
						m.currentPath = newPath
						m.loadFiles()
						m.cursor = 0
						m.pathMode = false
						m.pathInput.Blur()
						m.searchInput.Focus()
						m.err = nil
					} else {
						m.err = fmt.Errorf("invalid directory: %s", newPath)
					}
				} else {
					m.pathMode = false
					m.pathInput.Blur()
					m.searchInput.Focus()
				}
				return m, nil

			case tea.KeyEsc:
				m.pathMode = false
				m.pathInput.Blur()
				m.pathInput.SetValue(m.currentPath) // Reset
				m.searchInput.Focus()
				return m, nil
			}
			m.pathInput, cmd = m.pathInput.Update(msg)
			return m, cmd
		}

		// Main "Always Search" Mode

		// Help Screen Handler
		if m.showHelp {
			switch msg.String() {
			case "esc", "?":
				m.showHelp = false
				return m, nil
			default:
				var cmd tea.Cmd
				m.helpView, cmd = m.helpView.Update(msg)
				return m, cmd
			}
		}

		// 1. Navigation & Search Control
		switch msg.String() {
		case "?":
			m.showHelp = true
			m.helpView.GotoTop()
			return m, nil

		case "ctrl+l":
			m.pathMode = true
			m.pathInput.SetValue(m.currentPath)
			m.pathInput.Focus()
			m.searchInput.Blur()
			return m, textinput.Blink

		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case "down", "j":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}
			return m, nil

		case "left", "esc":
			// 1. If searching with text/filter active, clear it first
			if m.searchInput.Value() != "" {
				m.searchInput.Reset()
				m.filterFiles("")
				return m, nil
			}

			// 2. If we have navigation history, pop it (Back button behavior)
			if len(m.history) > 0 {
				items := len(m.history)
				prev := m.history[items-1]
				m.history = m.history[:items-1]
				m.currentPath = prev
				m.loadFiles()
				m.cursor = 0
				m.pathInput.SetValue(m.currentPath)
				return m, nil
			}

			// 3. If no history, try to go UP a directory (Parent behavior)
			// e.g. if we started in C:\Users, Esc should go to C:\
			parent := filepath.Dir(m.currentPath)
			// Check if we are at root (parent is same as current or "." on some OS, or volume root)
			// On Windows "C:\" parent is "C:\". On Linux "/" parent is "/".
			if parent != "." && parent != m.currentPath {
				// Determine if we are effectively at a drive root
				// Windows: filepath.Dir("C:\") -> "C:\"
				// So if parent == currentPath, we are at root. (Handled by condition above)
				m.currentPath = parent
				m.loadFiles()
				m.cursor = 0
				m.pathInput.SetValue(m.currentPath)
				return m, nil
			}

			// 4. If at root and nothing else to do, Go Back
			return m, func() tea.Msg { return BackMsg{} }

		case "enter":
			if len(m.filtered) == 0 {
				return m, nil
			}
			selected := m.filtered[m.cursor]

			pathName := selected.Name()
			var fullPath string
			if filepath.IsAbs(pathName) {
				fullPath = pathName
			} else {
				fullPath = filepath.Join(m.currentPath, pathName)
			}

			// FIX: Check if it's a directory using os.Stat
			// This handles dummyEntry (from search) which reports false for IsDir()
			info, err := os.Stat(fullPath)
			isDir := false
			if err == nil && info.IsDir() {
				isDir = true
			} else if selected.IsDir() {
				// Fallback to entry info if Stat fails (rare) or strict
				isDir = true
			}

			if isDir {
				m.history = append(m.history, m.currentPath)
				m.currentPath = fullPath

				// Entering a folder:
				// 1. Reset Search (so we see contents, not the old query)
				m.searchInput.Reset()
				// 2. Disable Global Search (we are now scoping to this folder)
				m.globalSearch = false

				m.loadFiles()
				m.cursor = 0
			} else {
				m.selectedFile = fullPath
				return m, func() tea.Msg { return SwitchViewMsg{TargetState: StateEditor, Args: fullPath} }
			}
			return m, nil

		case "tab":
			m.globalSearch = !m.globalSearch
			m.searchInput.Focus()
			if m.globalSearch {
				if m.allFilePaths == nil {
					m.loading = true
					// Start the scanner if not already done?
					// Ideally we only scan once. If nil, scan.
					// If we previously scanned and m.allFilePaths exists, we don't need to re-scan unless requested.
					// But init scan handles it.
					// If user toggles OFF then ON, we already have data.
					// If init hasn't run yet? Init runs on startup.
				}
				m.searchInput.Placeholder = "SEARCHING ALL DRIVES..."
			} else {
				m.searchInput.Placeholder = "Type to search current dir..."
			}
			m.filterFiles(m.searchInput.Value())
			return m, nil

		case "left_arrow_placeholder":
			// Consolidated above
		}

		// Handle explicit KEY TYPES if not caught by string
		// actually bubbletea msg.String() handles "left" correctly.
		// We can remove the separated blocks.

		// 2. Actions (Alt+...)
		switch msg.String() {
		case "alt+m":
			if len(m.filtered) > 0 {
				selected := m.filtered[m.cursor]
				m.selectedForMove = selected.Name()
				m.moveInput.SetValue(selected.Name())
				m.moveMode = true
				m.moveInput.Focus()
				return m, textinput.Blink
			}
		case "alt+c":
			if len(m.filtered) > 0 {
				selected := m.filtered[m.cursor]
				m.selectedForCopy = selected.Name()
				m.copyInput.SetValue(selected.Name())
				m.copyMode = true
				m.copyInput.Focus()
				return m, textinput.Blink
			}
		case "alt+e":
			if len(m.filtered) > 0 {
				selected := m.filtered[m.cursor]
				if !selected.IsDir() {
					pathName := selected.Name()
					var fullPath string
					if filepath.IsAbs(pathName) {
						fullPath = pathName
					} else {
						fullPath = filepath.Join(m.currentPath, pathName)
					}
					m.selectedFile = fullPath
					return m, func() tea.Msg { return SwitchViewMsg{TargetState: StateEditor, Args: fullPath} }
				}
			}
		case "backspace":
			// Special case: if input empty, go up?
			if m.searchInput.Value() == "" {
				m.currentPath = filepath.Dir(m.currentPath)
				m.loadFiles()
				m.cursor = 0
				m.pathInput.SetValue(m.currentPath)
				m.searchInput.Reset()
				return m, nil
			}
		}

		// 3. Search Input (Default)
		// Ensure focus
		m.searchInput.Focus()
		oldValue := m.searchInput.Value()
		var nextCmd tea.Cmd
		m.searchInput, nextCmd = m.searchInput.Update(msg)

		if m.searchInput.Value() != oldValue {
			m.searchID++
			// If empty, reset immediately
			if m.searchInput.Value() == "" {
				m.filtered = m.files // Show local files? Or clear?
				// "Type to search ALL DRIVES" implies we show nothing or everything?
				// Previously we showed m.files (local dir) when empty
				m.filterFiles("")
				return m, nextCmd
			}
			// Trigger Debounce
			nextCmd = tea.Batch(nextCmd, tea.Tick(200*time.Millisecond, func(_ time.Time) tea.Msg {
				return searchDebounceMsg{id: m.searchID}
			}))
		}
		return m, nextCmd
	}
	return m, nil
}

func (m FileManagerModel) View() string {
	if m.quitting {
		return ""
	}

	// Fallback if dimensions incorrectly 0
	w, h := m.width, m.height
	if w <= 0 {
		w = 80
	}
	if h <= 0 {
		h = 24
	}

	// Show help screen
	if m.showHelp {
		return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center,
			lipgloss.JoinVertical(lipgloss.Center,
				lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).MarginBottom(1).Render("File Manager Help"),
				m.helpView.View(),
				lipgloss.NewStyle().Foreground(lipgloss.Color("240")).MarginTop(1).Render("Press [Esc] or [?] to go back"),
			),
		)
	}

	// If not ready (WindowSizeMsg not received), show Loading or Safe Default
	if !m.ready {
		// Return a simple loading screen to avoid artifacts
		return lipgloss.NewStyle().Width(w).Height(h).Align(lipgloss.Center, lipgloss.Center).Render("Loading File Manager...")
	}

	// 1. Render Search Bar (Header) First to measure
	searchBorder := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#0F9E99")). // Tropical Teal (Matches Dashboard)
		Padding(0, 1).
		Width(w - 4)

	loading := ""
	if m.loading && m.searchInput.Value() != "" {
		loading = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF79C6")).Render("  Scanning...")
	} else if m.loading {
		loading = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(" (Indexing...)")
	}

	searchBar := searchBorder.Render(m.searchInput.View() + loading)
	headerHeight := lipgloss.Height(searchBar)

	// 2. Render Footer to measure (moved up)
	grey := lipgloss.Color("240")
	infoStyle := lipgloss.NewStyle().Foreground(grey)

	// Truncate path if too long
	// dispPath := m.currentPath
	// if len(dispPath) > w/2 {
	// 	dispPath = "..." + dispPath[len(dispPath)-(w/2):]
	// }

	// Path Box Style
	pathBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#BD93F9")). // Purple
		Padding(0, 1).
		Foreground(lipgloss.Color("#50FA7B")) // Green text

	pathContent := m.pathInput.View()
	if !m.pathMode {
		pathContent = m.currentPath // Just show text if not editing
	}
	pathBox := pathBoxStyle.Render(pathContent)

	// Status Bar (Top of Footer)
	status := fmt.Sprintf("  Files: %d  Global: %v", len(m.filtered), m.globalSearch)
	infoBar := lipgloss.JoinHorizontal(lipgloss.Left, pathBox, infoStyle.Render(status))

	keyFooter := ""
	if m.moveMode {
		keyFooter = fmt.Sprintf("Rename/Move '%s' to: %s", m.selectedForMove, m.moveInput.View())
	} else if m.copyMode {
		keyFooter = fmt.Sprintf("Copy '%s' to: %s", m.selectedForCopy, m.copyInput.View())
	} else {
		drives := getDrives()
		keyFooter = infoStyle.Render(fmt.Sprintf("Esc: Back • Tab: Global • [Ctrl+L] Edit Path • [?] Help • Drives: %v", drives))
	}

	totalFilesStr := fmt.Sprintf("Total files : %d", len(m.filtered))
	leftText := keyFooter
	rightText := infoStyle.Render(totalFilesStr)

	gap := w - lipgloss.Width(leftText) - lipgloss.Width(rightText) - 2
	if gap < 1 {
		gap = 1
	}
	if gap > w {
		gap = 1
	}

	combinedFooter := leftText + strings.Repeat(" ", gap) + rightText
	fullFooter := lipgloss.JoinVertical(lipgloss.Left, infoBar, combinedFooter)
	footerHeight := lipgloss.Height(fullFooter)

	// 3. Calculate List Height
	// Available H = Window - Header - Footer
	listHeight := h - headerHeight - footerHeight
	if listHeight < 0 {
		listHeight = 0
	}

	// 4. Render File List (with calculated height)
	var list strings.Builder

	start := 0
	end := len(m.filtered)

	if m.cursor >= listHeight {
		start = m.cursor - listHeight + 1
	}
	if end > start+listHeight {
		end = start + listHeight
	}

	if len(m.filtered) == 0 {
		list.WriteString("\n  (No matches found)")
		// Ensure even empty msg doesn't overflow if listHeight is tiny
	} else {
		for i := start; i < end; i++ {
			f := m.filtered[i]
			isCursor := m.cursor == i

			name := f.Name()
			icon := "  "

			if f.IsDir() {
				icon = ""
			} else {
				if strings.HasSuffix(name, ".go") {
					icon = ""
				} else if strings.HasSuffix(name, ".py") {
					icon = ""
				} else if strings.HasSuffix(name, ".js") || strings.HasSuffix(name, ".ts") {
					icon = ""
				} else if strings.HasSuffix(name, ".md") {
					icon = ""
				} else if strings.HasSuffix(name, ".json") || strings.HasSuffix(name, ".yaml") {
					icon = "️ "
				} else {
					icon = ""
				}
			}

			// Styling
			var nameStyle, iconStyle lipgloss.Style
			var rowRendered string

			if isCursor {
				nameStyle = lipgloss.NewStyle().
					Background(lipgloss.Color("#5A4E8C")).
					Foreground(lipgloss.Color("#FFFFFF")).
					Bold(true).
					Padding(0, 1)

				iconStyle = lipgloss.NewStyle().
					Background(lipgloss.Color("#5A4E8C")).
					Foreground(lipgloss.Color("#FFFFFF")).
					Padding(0, 0, 0, 1)

				rowContent := fmt.Sprintf("%s %s", icon, name)
				rowRendered = lipgloss.NewStyle().
					Background(lipgloss.Color("#5A4E8C")).
					Width(w - 2).
					Render(rowContent)
			} else {
				if f.IsDir() {
					nameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#44A8F0"))
				} else {
					nameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#E0E0E0"))
				}
				iconStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

				rowRendered = fmt.Sprintf(" %s %s", iconStyle.Render(icon), nameStyle.Render(name))
				rowRendered = lipgloss.NewStyle().Width(w - 2).Render(rowRendered)
			}

			list.WriteString(rowRendered + "\n")
		}
	}

	listContent := list.String()

	// 5. Final Assembly (Lipgloss)
	var scrollbar strings.Builder
	totalFiles := len(m.filtered)
	if totalFiles > listHeight {
		scrollThumbHeight := int(float64(listHeight) * float64(listHeight) / float64(totalFiles))
		if scrollThumbHeight < 1 {
			scrollThumbHeight = 1
		}
		scrollOffset := int(float64(m.cursor) * float64(listHeight-scrollThumbHeight) / float64(totalFiles-1))

		for i := 0; i < listHeight; i++ {
			if i >= scrollOffset && i < scrollOffset+scrollThumbHeight {
				scrollbar.WriteString("\n")
			} else {
				scrollbar.WriteString("\n")
			}
		}
	} else {
		for i := 0; i < listHeight; i++ {
			scrollbar.WriteString(" \n")
		}
	}

	listWithScroll := lipgloss.JoinHorizontal(lipgloss.Top, listContent, lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(scrollbar.String()))
	currentHeight := lipgloss.Height(listWithScroll)

	if currentHeight < listHeight {
		gap := listHeight - currentHeight
		filler := strings.Repeat("\n", gap)
		listWithScroll = listWithScroll + filler
	} else if currentHeight > listHeight {
	}
	viewContent := lipgloss.JoinVertical(lipgloss.Left,
		searchBar,
		listWithScroll,
		fullFooter,
	)

	// Strict safety clamp
	return lipgloss.NewStyle().MaxHeight(h).Render(viewContent)
}

func (m *FileManagerModel) loadFiles() {
	entries, err := os.ReadDir(m.currentPath)
	if err != nil {
		m.err = err
		return
	}
	// ... sort ...
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir() && !entries[j].IsDir() {
			return true
		}
		if !entries[i].IsDir() && entries[j].IsDir() {
			return false
		}
		return entries[i].Name() < entries[j].Name()
	})

	m.files = entries
	// FIX: Always filter to update view when files are loaded, even if background scan is running.
	m.filterFiles(m.searchInput.Value())
}

func (m *FileManagerModel) filterFiles(query string) {
	if query == "" {
		m.filtered = m.files
		return
	}

	if !m.globalSearch && m.allFilePaths == nil {
		// Lazy load local
		m.reloadAllFiles()
	}

	const fastSearchThreshold = 20000

	var matches []string

	if len(m.allFilePaths) > fastSearchThreshold {
		// FAST PATH: Simple Case-Insensitive Substring Match
		lowerQuery := strings.ToLower(query)
		for _, path := range m.allFilePaths {
			if strings.Contains(strings.ToLower(path), lowerQuery) {
				matches = append(matches, path)
			}
		}
	} else {
		// SLOW PATH: Fuzzy Match
		fuzzyMatches := fuzzy.Find(query, m.allFilePaths)
		for _, m := range fuzzyMatches {
			matches = append(matches, m.Str)
		}
	}

	var results []fs.DirEntry
	for _, matchPath := range matches {
		results = append(results, dummyEntry{path: matchPath})
	}
	m.filtered = results
	m.cursor = 0
}

func (m *FileManagerModel) reloadAllFiles() {
	if m.globalSearch {
		return // Should be handled by async loader
	}
	// Local recursive load (sync)
	m.allFilePaths = []string{}
	filepath.WalkDir(m.currentPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if path == m.currentPath {
			return nil
		}
		rel, _ := filepath.Rel(m.currentPath, path)
		m.allFilePaths = append(m.allFilePaths, rel)
		return nil
	})
}

// Dummy entry for search results
type dummyEntry struct {
	path string
}

func (d dummyEntry) Name() string               { return d.path }
func (d dummyEntry) IsDir() bool                { return false } // Assume file for search results mainly? Or check ext? using stat is better but slow.
func (d dummyEntry) Type() fs.FileMode          { return 0 }
func (d dummyEntry) Info() (fs.FileInfo, error) { return nil, nil }

func (m FileManagerModel) Init() tea.Cmd {
	var cmds []tea.Cmd
	cmds = append(cmds, tea.EnableMouseCellMotion) // Enable Mouse

	// Only start global scan if we haven't already loaded files or if explicitly requested.
	if len(m.allFilePaths) == 0 {
		cmds = append(cmds, startGlobalScanCmd(m.scanChan))
	}
	return tea.Batch(cmds...)
}
