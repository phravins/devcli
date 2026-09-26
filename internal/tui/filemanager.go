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
	currentPath	string
	files		[]fs.DirEntry
	filtered	[]fs.DirEntry

	cursor		int
	width		int
	height		int
	quitting	bool
	searchInput	textinput.Model
	err		error

	selectedFile	string

	moveMode	bool
	moveInput	textinput.Model
	selectedForMove	string

	copyMode	bool
	copyInput	textinput.Model
	selectedForCopy	string

	pathMode	bool
	pathInput	textinput.Model

	allFilePaths	[]string

	history	[]string

	globalSearch	bool

	loading	bool

	searchID	int

	scanChan	chan string

	ready	bool

	showHelp	bool
	helpView	viewport.Model
}

type searchDebounceMsg struct {
	id int
}

type filterFinishedMsg struct {
	results []fs.DirEntry
}

const maxSearchResults = 1000
const fastSearchThreshold = 5000

func containsIgnoreCase(s, lowerSubstr string) bool {
	if len(lowerSubstr) == 0 {
		return true
	}
	if len(s) < len(lowerSubstr) {
		return false
	}

	hasNonASCII := false
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 {
			hasNonASCII = true
			break
		}
	}

	if hasNonASCII {
		return strings.Contains(strings.ToLower(s), lowerSubstr)
	}

	n := len(lowerSubstr)
	for i := 0; i <= len(s)-n; i++ {
		match := true
		for j := 0; j < n; j++ {
			b := s[i+j]
			if b >= 'A' && b <= 'Z' {
				b += 32
			}
			if b != lowerSubstr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func matchPaths(paths []string, query string) []fs.DirEntry {
	if query == "" {
		return nil
	}

	var matches []string
	useFastPath := len(paths) > fastSearchThreshold

	if useFastPath {
		lowerQuery := strings.ToLower(query)
		for _, path := range paths {
			if len(matches) >= maxSearchResults {
				break
			}
			if containsIgnoreCase(path, lowerQuery) {
				matches = append(matches, path)
			}
		}
	} else {
		fuzzyMatches := fuzzy.Find(query, paths)
		for _, m := range fuzzyMatches {
			if len(matches) >= maxSearchResults {
				break
			}
			matches = append(matches, m.Str)
		}
	}

	results := make([]fs.DirEntry, 0, len(matches))
	for _, matchPath := range matches {
		results = append(results, dummyEntry{path: matchPath})
	}
	return results
}

func performSearchCmd(paths []string, query string) tea.Cmd {
	return func() tea.Msg {
		results := matchPaths(paths, query)
		return filterFinishedMsg{results: results}
	}
}

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
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
	ti.Cursor.Style = lipgloss.NewStyle().Background(lipgloss.Color("255"))
	ti.Focus()

	hv := viewport.New(80, 20)
	hv.Style = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("62")).Padding(1, 2)

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
		currentPath:	startPath,
		searchInput:	ti,
		moveInput:	mi,
		copyInput:	ci,
		pathInput:	pi,
		globalSearch:	true,
		loading:	true,
		scanChan:	make(chan string, 1000),

		helpView:	hv,
	}

	m.reloadAllFiles()

	m.loadFiles()
	return m
}

type scanStartedMsg struct{}

type searchResultMsg struct {
	paths []string
}

type scanFinishedMsg struct{}

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

func waitForSearchResults(ch chan string) tea.Cmd {
	return func() tea.Msg {
		var batch []string
		const maxBatch = 5000
		const batchTimeout = 200 * time.Millisecond

		path, ok := <-ch
		if !ok {
			return scanFinishedMsg{}
		}
		batch = append(batch, path)

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

	case scanStartedMsg:
		m.loading = true
		return m, waitForSearchResults(m.scanChan)

	case searchResultMsg:
		m.allFilePaths = append(m.allFilePaths, msg.paths...)

		if m.searchInput.Value() != "" {
			query := strings.ToLower(m.searchInput.Value())
			for _, p := range msg.paths {
				if len(m.filtered) >= maxSearchResults {
					break
				}
				if containsIgnoreCase(p, query) {
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
		m.ready = true

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
				m.cursor -= 3
				if m.cursor < 0 {
					m.cursor = 0
				}
			}
		case tea.MouseWheelDown:
			if m.cursor < len(m.filtered)-1 {
				m.cursor += 3
				if m.cursor >= len(m.filtered) {
					m.cursor = len(m.filtered) - 1
				}
			}
		case tea.MouseLeft:

			headerHeight := 3
			spacerHeight := 1
			footerHeight := 2
			listStartY := headerHeight + spacerHeight

			listHeight := m.height - headerHeight - footerHeight - 2
			if listHeight < 1 {
				listHeight = 1
			}

			if msg.Y >= listStartY && msg.Y < listStartY+listHeight {

				start := 0
				if m.cursor >= listHeight {
					start = m.cursor - listHeight + 1
				}

				clickOffset := msg.Y - listStartY
				clickedIndex := start + clickOffset

				if clickedIndex >= 0 && clickedIndex < len(m.filtered) {
					m.cursor = clickedIndex

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
						m.pathInput.SetValue(fullPath)
						m.currentPath = fullPath
						m.searchInput.Reset()
						m.globalSearch = false
						m.loadFiles()
						m.cursor = 0
					} else {
						m.selectedFile = fullPath

						return m, func() tea.Msg { return SwitchViewMsg{TargetState: StateEditor, Args: fullPath} }
					}
				}
			}
		}

		return m, nil

	case tea.KeyMsg:

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
				m.pathInput.SetValue(m.currentPath)
				m.searchInput.Focus()
				return m, nil
			}
			m.pathInput, cmd = m.pathInput.Update(msg)
			return m, cmd
		}

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

			if m.searchInput.Value() != "" {
				m.searchInput.Reset()
				m.filterFiles("")
				return m, nil
			}

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

			parent := filepath.Dir(m.currentPath)

			if parent != "." && parent != m.currentPath {

				m.currentPath = parent
				m.loadFiles()
				m.cursor = 0
				m.pathInput.SetValue(m.currentPath)
				return m, nil
			}

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

			info, err := os.Stat(fullPath)
			isDir := false
			if err == nil && info.IsDir() {
				isDir = true
			} else if selected.IsDir() {

				isDir = true
			}

			if isDir {
				m.history = append(m.history, m.currentPath)
				m.currentPath = fullPath

				m.searchInput.Reset()

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

				}
				m.searchInput.Placeholder = "SEARCHING ALL DRIVES..."
			} else {
				m.searchInput.Placeholder = "Type to search current dir..."
			}
			m.filterFiles(m.searchInput.Value())
			return m, nil

		case "left_arrow_placeholder":

		}

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

			if m.searchInput.Value() == "" {
				m.currentPath = filepath.Dir(m.currentPath)
				m.loadFiles()
				m.cursor = 0
				m.pathInput.SetValue(m.currentPath)
				m.searchInput.Reset()
				return m, nil
			}
		}

		m.searchInput.Focus()
		oldValue := m.searchInput.Value()
		var nextCmd tea.Cmd
		m.searchInput, nextCmd = m.searchInput.Update(msg)

		if m.searchInput.Value() != oldValue {
			m.searchID++

			if m.searchInput.Value() == "" {
				m.filtered = m.files

				m.filterFiles("")
				return m, nextCmd
			}

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

	w, h := m.width, m.height
	if w <= 0 {
		w = 80
	}
	if h <= 0 {
		h = 24
	}

	if m.showHelp {
		return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center,
			lipgloss.JoinVertical(lipgloss.Center,
				lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).MarginBottom(1).Render("File Manager Help"),
				m.helpView.View(),
				lipgloss.NewStyle().Foreground(lipgloss.Color("240")).MarginTop(1).Render("Press [Esc] or [?] to go back"),
			),
		)
	}

	if !m.ready {

		return lipgloss.NewStyle().Width(w).Height(h).Align(lipgloss.Center, lipgloss.Center).Render("Loading File Manager...")
	}

	searchBorder := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#0F9E99")).
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

	grey := lipgloss.Color("240")
	infoStyle := lipgloss.NewStyle().Foreground(grey)

	pathBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#BD93F9")).
		Padding(0, 1).
		Foreground(lipgloss.Color("#50FA7B"))

	pathContent := m.pathInput.View()
	if !m.pathMode {
		pathContent = m.currentPath
	}
	pathBox := pathBoxStyle.Render(pathContent)

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

	listHeight := h - headerHeight - footerHeight
	if listHeight < 0 {
		listHeight = 0
	}

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

			var nameStyle, iconStyle lipgloss.Style
			var rowRendered string

			if isCursor {
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

			list.WriteString(rowRendered)
			list.WriteString("\n")
		}
	}

	listContent := list.String()

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

	return lipgloss.NewStyle().MaxHeight(h).Render(viewContent)
}

func (m *FileManagerModel) loadFiles() {
	entries, err := os.ReadDir(m.currentPath)
	if err != nil {
		m.err = err
		return
	}

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

	m.filterFiles(m.searchInput.Value())
}

func (m *FileManagerModel) filterFiles(query string) {
	if query == "" {
		m.filtered = m.files
		return
	}

	if !m.globalSearch && m.allFilePaths == nil {

		m.reloadAllFiles()
	}

	m.filtered = matchPaths(m.allFilePaths, query)
	m.cursor = 0
}

func (m *FileManagerModel) reloadAllFiles() {
	if m.globalSearch {
		return
	}

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

type dummyEntry struct {
	path string
}

func (d dummyEntry) Name() string		{ return d.path }
func (d dummyEntry) IsDir() bool		{ return false }
func (d dummyEntry) Type() fs.FileMode		{ return 0 }
func (d dummyEntry) Info() (fs.FileInfo, error)	{ return nil, nil }

func (m FileManagerModel) Init() tea.Cmd {
	var cmds []tea.Cmd
	cmds = append(cmds, tea.EnableMouseCellMotion)

	if len(m.allFilePaths) == 0 {
		cmds = append(cmds, startGlobalScanCmd(m.scanChan))
	}
	return tea.Batch(cmds...)
}
