package tui

import (
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/phravins/devcli/internal/timemachine"
)

type TimeMachineModel struct {
	timeline       *timemachine.Timeline
	viewport       viewport.Model
	blameViewport  viewport.Model
	detailViewport viewport.Model
	helpViewport   viewport.Model
	width          int
	height         int
	ready          bool
	showHelp       bool
	err            error
	bugSuspects    []timemachine.BugSuspect
	authorColors   map[string]lipgloss.Color
}

func NewTimeMachineModel(repoPath, filePath string) (*TimeMachineModel, error) {
	timeline, err := timemachine.NewTimeline(repoPath, filePath)
	if err != nil {
		return nil, err
	}
	suspects := timemachine.AnalyzeBugRisks(timeline.Commits)

	colors := generateAuthorColors(timeline.GetAuthors())
	blameVp := viewport.New(80, 20)
	detailVp := viewport.New(80, 15)
	helpVp := viewport.New(80, 30)
	helpVp.MouseWheelEnabled = true
	helpVp.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#0F9E99")).
		Padding(1, 2)
	helpVp.SetContent(TimeMachineHelp)

	model := &TimeMachineModel{
		timeline:       timeline,
		bugSuspects:    suspects,
		authorColors:   colors,
		blameViewport:  blameVp,
		detailViewport: detailVp,
		helpViewport:   helpVp,
		width:          160,
		height:         40,
		ready:          true, // Mark as ready immediately
	}

	// Set initial content
	model.updateViewports()

	return model, nil
}

// Init initializes the model
func (m *TimeMachineModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m *TimeMachineModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			// Go back to Bonus menu instead of quitting
			return m, func() tea.Msg { return SubFeatureBackMsg{} }

		case "ctrl+c":
			// Force quit
			return m, tea.Quit

		case "?":
			m.showHelp = !m.showHelp
			if m.showHelp {
				// Resize and update help viewport when showing
				m.helpViewport.Width = m.width - 8
				m.helpViewport.Height = m.height - 4
				m.helpViewport.GotoTop()
			}
			return m, nil

		case "left", "h":
			// Go to newer commit
			if err := m.timeline.Previous(); err == nil {
				m.updateViewports()
			}
			return m, nil

		case "right", "l":
			// Go to older commit
			if err := m.timeline.Next(); err == nil {
				m.updateViewports()
			}
			return m, nil

		case "home":
			// Go to newest (current) commit
			if err := m.timeline.MoveToIndex(0); err == nil {
				m.updateViewports()
			}
			return m, nil

		case "end":
			// Go to oldest commit
			if err := m.timeline.MoveToIndex(m.timeline.GetCommitCount() - 1); err == nil {
				m.updateViewports()
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizeViewports()
		m.updateViewports()
		// Also resize help viewport
		if m.showHelp {
			m.helpViewport.Width = m.width - 8
			m.helpViewport.Height = m.height - 4
		}
		return m, nil
	}

	// Update viewports
	var cmd tea.Cmd
	if m.showHelp {
		// Update help viewport when help is shown
		m.helpViewport, cmd = m.helpViewport.Update(msg)
	} else {
		// Update blame viewport when help is not shown
		m.blameViewport, cmd = m.blameViewport.Update(msg)
	}
	return m, cmd
}

// View renders the UI
func (m *TimeMachineModel) View() string {
	if !m.ready {
		return "Initializing Code Time Machine..."
	}

	if m.showHelp {
		return m.renderHelp()
	}

	// Build the main content box
	boxContent := m.renderBoxContent()

	// Wrap in a bordered box
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#4ECDC4")).
		Padding(1, 2).
		Width(m.width - 10).
		Height(m.height - 6)

	box := boxStyle.Render(boxContent)

	// Center the box on screen
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}
func (m *TimeMachineModel) setupViewports() {
	// Fixed height for top UI elements (Header + Timeline + Padding)
	topUIHeight := 6

	// Fixed height for bottom Detail Box - keep it very small
	detailHeight := 0 // Set to 0 to minimize empty space

	// Blame View (Tracking History) takes the rest
	// Total height - TopUI - DetailHeight - Margins (more generous)
	blameHeight := m.height - topUIHeight - detailHeight - 8 // Increased margin from 4 to 8
	if blameHeight < 10 {
		blameHeight = 10
	}

	// Width with better margins: 8 chars total (4 on each side)
	availableWidth := m.width - 12 // Increased from 10 to 12 for wider margins
	if availableWidth < 60 {
		availableWidth = 60
	}

	// Detail box should be narrower - use 25% of available width
	detailWidth := int(float64(availableWidth) * 0.25) // Make detail box very small and compact

	m.blameViewport = viewport.New(availableWidth, blameHeight)
	m.detailViewport = viewport.New(detailWidth, detailHeight)
}
func (m *TimeMachineModel) resizeViewports() {
	// Re-run setup logic with new dimensions
	m.setupViewports()
}
func (m *TimeMachineModel) updateViewports() {
	m.blameViewport.SetContent(m.renderBlameView())
	m.detailViewport.SetContent(m.renderCommitDetails())
}
func (m *TimeMachineModel) renderHeader() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FF6B6B")).
		Padding(0, 1)

	fileStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6BCF7F")).
		Padding(0, 1)

	title := titleStyle.Render("Code Time Machine")
	file := fileStyle.Render(m.timeline.FilePath)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		file,
	)
}

// renderBoxContent combines all content sections into a single layout
func (m *TimeMachineModel) renderBoxContent() string {
	header := m.renderHeader()
	timeline := m.renderTimeline()
	blameView := m.blameViewport.View()
	commitDetails := m.renderCommitDetailsCompact()
	footer := m.renderFooter()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		timeline,
		"",
		blameView,
		"",
		commitDetails,
		"",
		footer,
	)
}
func (m *TimeMachineModel) renderTimeline() string {
	if len(m.timeline.Commits) == 0 {
		return ""
	}

	progress := m.timeline.GetProgress()
	current := m.timeline.GetCurrentCommit()
	barWidth := m.width - 20 - 4
	if barWidth < 10 {
		barWidth = 10
	}
	filledWidth := int(float64(barWidth) * progress)

	filled := strings.Repeat("═", filledWidth)
	empty := strings.Repeat("─", barWidth-filledWidth)

	timelineBar := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4ECDC4")).
		Render(filled) +
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("#353535")).
			Render(empty)

	// Position info
	position := fmt.Sprintf("Commit %d/%d", m.timeline.CurrentIndex+1, len(m.timeline.Commits))

	// Date
	dateStr := ""
	if current != nil {
		dateStr = current.Date.Format("Jan 02, 2006")
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		"  ● ",
		timelineBar,
		" ●  ",
		position,
		"  ",
		dateStr,
	)
}
func (m *TimeMachineModel) renderBlameView() string {
	var lines []string

	suspiciousCommits := make(map[string]bool)
	for _, suspect := range m.bugSuspects {
		suspiciousCommits[suspect.Commit.Hash] = true
	}
	numStyle := lipgloss.NewStyle().Width(5).Align(lipgloss.Right).Foreground(lipgloss.Color("#666666"))
	sepStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#444444"))
	riskStyle := lipgloss.NewStyle().Width(3).Foreground(lipgloss.Color("#FF4444"))
	authorStyle := lipgloss.NewStyle().Width(15)
	dateStyle := lipgloss.NewStyle().Width(13).Foreground(lipgloss.Color("#888888"))
	codeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#E0E0E0"))

	overhead := 46 // Increased overhead for right border (was 43)
	availableCodeWidth := m.blameViewport.Width - overhead
	if availableCodeWidth < 20 {
		availableCodeWidth = 20
	}

	for _, line := range m.timeline.BlameData {
		// Line Number
		lNum := numStyle.Render(fmt.Sprintf("%d", line.LineNumber))

		// Separator
		sep := sepStyle.Render(" │ ")

		// Risk
		rStr := "   "
		if suspiciousCommits[line.CommitHash] {
			rStr = "!  "
		}
		risk := riskStyle.Render(rStr)
		aName := truncate(line.Author, 15)
		author := authorStyle.Foreground(m.authorColors[line.Author]).Render(aName)

		// Date
		dStr := line.Timestamp.Format("Jan 02 15:04")
		date := dateStyle.Render(dStr)

		// Code
		cStr := line.Content
		if len(cStr) > availableCodeWidth {
			cStr = truncate(cStr, availableCodeWidth)
		}
		code := codeStyle.Render(cStr)

		// Join horizontally to ensure width alignment is preserved
		// Added right separator
		fullLine := lipgloss.JoinHorizontal(lipgloss.Bottom, lNum, sep, risk, author, " ", date, sep, code, sep)
		lines = append(lines, fullLine)
	}

	return strings.Join(lines, "\n")
}

// renderCommitDetails creates the commit details panel
func (m *TimeMachineModel) renderCommitDetails() string {
	current := m.timeline.GetCurrentCommit()
	if current == nil {
		return "No commit selected"
	}

	var details []string

	// Commit hash with author and date on same line
	hashStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFD700")).
		Bold(true)

	authorStyle := lipgloss.NewStyle().Foreground(m.authorColors[current.Author])

	// Compact first line: Commit: hash
	details = append(details, hashStyle.Render("Commit: ")+current.ShortHash)

	// Second line: Author and Date together
	authorDateLine := "Author: " + authorStyle.Render(current.Author) +
		"    Date: " + current.Date.Format("Mon Jan 02, 2006 15:04")
	details = append(details, authorDateLine)
	details = append(details, "")

	// Message
	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true)

	details = append(details, messageStyle.Render("Message:"))
	for _, line := range strings.Split(current.Message, "\n") {
		details = append(details, "  "+line)
	}

	// Stats on same line
	statsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#4ECDC4"))
	statsLine := fmt.Sprintf("Files changed: %d    +%d -%d lines",
		len(current.FilesChanged), current.LinesAdded, current.LinesRemoved)
	details = append(details, statsStyle.Render(statsLine))

	// Bug risk if applicable
	for _, suspect := range m.bugSuspects {
		if suspect.Commit.Hash == current.Hash {
			riskStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(timemachine.GetRiskColor(suspect.Risk))).
				Bold(true)

			riskLevel := timemachine.GetRiskLevel(suspect.Risk)
			details = append(details, riskStyle.Render(fmt.Sprintf("⚠ Risk: %s (%.0f%%)", riskLevel, suspect.Risk*100)))
			details = append(details, "Reason: "+suspect.Reason)
		}
	}

	return strings.Join(details, "\n")
}

// renderCommitDetailsCompact creates a compact one-line commit summary
func (m *TimeMachineModel) renderCommitDetailsCompact() string {
	current := m.timeline.GetCurrentCommit()
	if current == nil {
		return ""
	}

	hashStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700")).Bold(true)
	authorStyle := lipgloss.NewStyle().Foreground(m.authorColors[current.Author])
	msgStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))

	// Get first line of message
	msgLines := strings.Split(current.Message, "\n")
	firstLineMsg := msgLines[0]
	if len(firstLineMsg) > 60 {
		firstLineMsg = firstLineMsg[:57] + "..."
	}

	return hashStyle.Render("● "+current.ShortHash) + " " +
		authorStyle.Render(current.Author) + " " +
		msgStyle.Render(firstLineMsg)
}

// renderMainContent returns the tracking history box (blame view)
func (m *TimeMachineModel) renderMainContent() string {
	// Tracking History Box with proper sizing and padding
	blameBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#4ECDC4")).
		Padding(0, 1).
		Width(m.blameViewport.Width + 4).   // Add padding to width
		Height(m.blameViewport.Height + 2). // Add padding to height
		Render(m.blameViewport.View())

	return blameBox
}

// renderDetailBox returns the commit details box
func (m *TimeMachineModel) renderDetailBox() string {
	// Commit Details Box - minimal padding for compact size
	detailBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF6B6B")).
		Padding(0, 0).                     // No padding for compact size
		Width(m.detailViewport.Width + 2). // Minimal border width
		Render(m.detailViewport.View())

	return detailBox
}

// renderFooter creates the footer with shortcuts
func (m *TimeMachineModel) renderFooter() string {
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Padding(1, 0)

	shortcuts := "←/→ Navigate │ Home/End Jump │ ? Help │ Q/Esc Back │ Ctrl+C Quit"
	return footerStyle.Render(shortcuts)
}

// renderHelp shows the help screen
func (m *TimeMachineModel) renderHelp() string {
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.helpViewport.View())
}

// generateAuthorColors creates consistent colors for authors
func generateAuthorColors(authors []string) map[string]lipgloss.Color {
	colors := map[string]lipgloss.Color{}

	// Predefined color palette
	palette := []string{
		"#FF6B6B", "#4ECDC4", "#45B7D1", "#FFA07A",
		"#98D8C8", "#F7DC6F", "#BB8FCE", "#85C1E2",
		"#F8B4D1", "#52D3AA", "#FDA7DF", "#87CEEB",
	}

	for i, author := range authors {
		if i < len(palette) {
			colors[author] = lipgloss.Color(palette[i])
		} else {
			// Generate color from hash
			colors[author] = hashToColor(author)
		}
	}

	return colors
}

// hashToColor generates a color from a string
func hashToColor(s string) lipgloss.Color {
	h := fnv.New32a()
	h.Write([]byte(s))
	hash := h.Sum32()

	// Generate RGB values
	r := (hash & 0xFF0000) >> 16
	g := (hash & 0x00FF00) >> 8
	b := hash & 0x0000FF

	// Ensure colors are bright enough
	r = (r % 156) + 100
	g = (g % 156) + 100
	b = (b % 156) + 100

	return lipgloss.Color(fmt.Sprintf("#%02X%02X%02X", r, g, b))
}

// truncate shortens a string to max length using runes (safe for emojis/unicode)
func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	if max < 3 {
		return string(runes[:max])
	}
	return string(runes[:max-3]) + "..."
}

// RunTimeMachine starts the Code Time Machine TUI
func RunTimeMachine(repoPath, filePath string) error {
	model, err := NewTimeMachineModel(repoPath, filePath)
	if err != nil {
		return fmt.Errorf("failed to create time machine: %w", err)
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running time machine: %w", err)
	}

	return nil
}
