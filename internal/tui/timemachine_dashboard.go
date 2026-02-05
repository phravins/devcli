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

// TimeMachineModel represents the Code Time Machine TUI state
type TimeMachineModel struct {
	timeline       *timemachine.Timeline
	viewport       viewport.Model
	blameViewport  viewport.Model
	detailViewport viewport.Model
	width          int
	height         int
	ready          bool
	showHelp       bool
	err            error
	bugSuspects    []timemachine.BugSuspect
	authorColors   map[string]lipgloss.Color
}

// NewTimeMachineModel creates a new Time Machine model
func NewTimeMachineModel(repoPath, filePath string) (*TimeMachineModel, error) {
	timeline, err := timemachine.NewTimeline(repoPath, filePath)
	if err != nil {
		return nil, err
	}

	// Analyze bug risks
	suspects := timemachine.AnalyzeBugRisks(timeline.Commits)

	// Generate author colors
	colors := generateAuthorColors(timeline.GetAuthors())

	// Create viewports with default size (will be resized on WindowSizeMsg)
	blameVp := viewport.New(80, 20)
	detailVp := viewport.New(80, 15)

	model := &TimeMachineModel{
		timeline:       timeline,
		bugSuspects:    suspects,
		authorColors:   colors,
		blameViewport:  blameVp,
		detailViewport: detailVp,
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
		return m, nil
	}

	// Update viewports
	var cmd tea.Cmd
	m.blameViewport, cmd = m.blameViewport.Update(msg)
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

	// Build the layout
	header := m.renderHeader()
	timeline := m.renderTimeline()
	mainContent := m.renderMainContent()
	footer := m.renderFooter()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		timeline,
		mainContent,
		footer,
	)
}

// setupViewports initializes the viewports
func (m *TimeMachineModel) setupViewports() {
	headerHeight := 3
	timelineHeight := 3
	footerHeight := 2

	availableHeight := m.height - headerHeight - timelineHeight - footerHeight - 4

	// Split available height: 60% blame, 40% details
	blameHeight := int(float64(availableHeight) * 0.6)
	detailHeight := availableHeight - blameHeight

	// Split width: 50/50
	halfWidth := (m.width - 3) / 2

	m.blameViewport = viewport.New(halfWidth, blameHeight)
	m.detailViewport = viewport.New(halfWidth, detailHeight)
}

// resizeViewports adjusts viewport sizes
func (m *TimeMachineModel) resizeViewports() {
	headerHeight := 3
	timelineHeight := 3
	footerHeight := 2

	// For vertical stacking: 2 boxes with borders and padding
	// Each box has: top border (1) + bottom border (1) + top padding (1) + bottom padding (1) = 4 lines
	// Total for 2 boxes = 8 lines
	availableHeight := m.height - headerHeight - timelineHeight - footerHeight - 8

	// Split height: 85% for blame view (tracking history), 15% for commit details
	blameHeight := int(float64(availableHeight) * 0.85)
	detailHeight := availableHeight - blameHeight

	// Use full width minus borders and padding
	// Account for: left border (1) + right border (1) + left padding (1) + right padding (1) = 4
	availableWidth := m.width - 4

	m.blameViewport.Width = availableWidth
	m.blameViewport.Height = blameHeight
	m.detailViewport.Width = availableWidth
	m.detailViewport.Height = detailHeight
}

// updateViewports refreshes viewport content
func (m *TimeMachineModel) updateViewports() {
	m.blameViewport.SetContent(m.renderBlameView())
	m.detailViewport.SetContent(m.renderCommitDetails())
}

// renderHeader creates the header section
func (m *TimeMachineModel) renderHeader() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FF6B6B")).
		Padding(0, 1)

	fileStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6BCF7F"))

	title := titleStyle.Render("⏱ Code Time Machine")
	file := fileStyle.Render(m.timeline.FilePath)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		file,
	)
}

// renderTimeline creates the timeline visualization
func (m *TimeMachineModel) renderTimeline() string {
	if len(m.timeline.Commits) == 0 {
		return ""
	}

	progress := m.timeline.GetProgress()
	current := m.timeline.GetCurrentCommit()

	// Timeline bar
	barWidth := m.width - 20
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

	return lipgloss.NewStyle().Padding(1, 0).Render(
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			"●",
			timelineBar,
			"●  ",
			position,
			"  ",
			dateStr,
		),
	)
}

// renderBlameView creates the blame/code view
func (m *TimeMachineModel) renderBlameView() string {
	var lines []string

	suspiciousCommits := make(map[string]bool)
	for _, suspect := range m.bugSuspects {
		suspiciousCommits[suspect.Commit.Hash] = true
	}

	for _, line := range m.timeline.BlameData {
		// Get author color
		color := m.authorColors[line.Author]

		// Line number
		lineNumStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Width(5).
			Align(lipgloss.Right)

		lineNum := lineNumStyle.Render(fmt.Sprintf("%d", line.LineNumber))

		// Author info
		authorStyle := lipgloss.NewStyle().
			Foreground(color).
			Width(15)

		author := authorStyle.Render(truncate(line.Author, 13))

		// Date
		dateStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Width(12)

		date := dateStyle.Render(line.Timestamp.Format("Jan 02 15:04"))

		// Risk indicator
		risk := ""
		if suspiciousCommits[line.CommitHash] {
			risk = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF4444")).
				Render("⚠  ")
		} else {
			risk = "  "
		}

		// Code content
		codeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#E0E0E0"))
		code := codeStyle.Render(line.Content)

		fullLine := fmt.Sprintf("%s │ %s%s %s │ %s", lineNum, risk, author, date, code)
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

	// Commit hash
	hashStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFD700")).
		Bold(true)

	details = append(details, hashStyle.Render("Commit: ")+current.ShortHash)
	details = append(details, "")

	// Author
	authorStyle := lipgloss.NewStyle().Foreground(m.authorColors[current.Author])
	details = append(details, "Author: "+authorStyle.Render(current.Author))

	// Date
	details = append(details, "Date:   "+current.Date.Format("Mon Jan 02, 2006 15:04"))
	details = append(details, "")

	// Message
	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true)

	details = append(details, messageStyle.Render("Message:"))
	for _, line := range strings.Split(current.Message, "\n") {
		details = append(details, "  "+line)
	}
	details = append(details, "")

	// Stats
	statsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#4ECDC4"))
	files := fmt.Sprintf("Files changed: %d", len(current.FilesChanged))
	changes := fmt.Sprintf("+%d -%d lines", current.LinesAdded, current.LinesRemoved)
	details = append(details, statsStyle.Render(files))
	details = append(details, statsStyle.Render(changes))

	// Bug risk if applicable
	for _, suspect := range m.bugSuspects {
		if suspect.Commit.Hash == current.Hash {
			details = append(details, "")
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

// renderMainContent combines blame and details panels
func (m *TimeMachineModel) renderMainContent() string {
	blameBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#4ECDC4")).
		Padding(1).
		Render(m.blameViewport.View())

	detailBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF6B6B")).
		Padding(1).
		Render(m.detailViewport.View())

	// Stack vertically: tracking history (blame) on top, commit details below
	return lipgloss.JoinVertical(lipgloss.Left, blameBox, detailBox)
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
	helpStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#4ECDC4")).
		Padding(2).
		Width(m.width - 4)

	help := `
Code Time Machine - Help

NAVIGATION:
  ← / H       Previous commit (newer)
  → / L       Next commit (older)
  Home        Jump to newest commit
  End         Jump to oldest commit

VIEW:
  ?           Toggle this help
  Q / Esc     Go back
  Ctrl+C      Force quit

FEATURES:
  • Line-by-line blame showing who changed what and when
  • Author color coding for easy identification
  • Bug risk indicators (⚠) for suspicious commits
  • Timeline navigation through commit history
  • Commit details with stats and messages

BUG RISK INDICATORS:
  ⚠ High Risk     Large refactors, late-night commits, WIP
  ⚠ Medium Risk   Significant changes, quick fixes
  No indicator    Low risk

Press ? to close this help.
`

	return helpStyle.Render(help)
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

// truncate shortens a string to max length
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
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
