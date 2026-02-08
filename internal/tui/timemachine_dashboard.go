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
	helpViewport   viewport.Model
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

	colors := generateAuthorColors(timeline.GetAuthors())
	blameVp := viewport.New(80, 20)
	detailVp := viewport.New(80, 15)
	helpVp := viewport.New(80, 30)
	helpVp.MouseWheelEnabled = true

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
				m.helpViewport.SetContent(m.getHelpContent())
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

	// Build the content stack with specific alignment
	header := m.renderHeader()
	timeline := m.renderTimeline()

	// Main content is now just the blame view
	mainContent := m.renderMainContent()

	// Create bottom row with Footer (left) and Detail Box (right)
	footer := m.renderFooter()
	detailBox := m.renderDetailBox()

	// Calculate spacer width to push detail box to the right
	// Footer width is approximate, but we can rely on Flex layout or safe assumptions
	// Better approach: JoinHorizontal with alignment

	// Since lipgloss doesn't have "Flex", we calculate space manually or use Place
	// A simpler way:
	// [ Footer .................... DetailBox ]

	footerWidth := lipgloss.Width(footer)
	detailWidth := lipgloss.Width(detailBox)
	spacerWidth := m.width - footerWidth - detailWidth - 4 // -4 for margins
	if spacerWidth < 0 {
		spacerWidth = 0
	}
	spacer := strings.Repeat(" ", spacerWidth)

	bottomRow := lipgloss.JoinHorizontal(lipgloss.Bottom, footer, spacer, detailBox)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		timeline,
		"", // Spacer
		mainContent,
		bottomRow,
	)

	// Anchoring to Top with a 3-line margin is even safer for Windows command prompts
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Top,
		lipgloss.NewStyle().PaddingTop(3).Render(content))
}
func (m *TimeMachineModel) setupViewports() {
	// Fixed height for top UI elements (Header + Timeline + Padding)
	topUIHeight := 6

	// Fixed height for bottom Detail Box
	detailHeight := 10 // Small fixed size for details

	// Blame View takes the rest
	// Total height - TopUI - DetailHeight - Margins
	blameHeight := m.height - topUIHeight - detailHeight - 4
	if blameHeight < 10 {
		blameHeight = 10
	}

	// Width overhead: 4 for borders/padding, plus 6 for safety (total 10 instead of 6)
	availableWidth := m.width - 10
	if availableWidth < 60 {
		availableWidth = 60
	}

	// Detail box width: Fixed or Percentage (e.g., 40% of width or min 60 chars)
	detailWidth := 60
	if detailWidth > availableWidth {
		detailWidth = availableWidth
	}

	m.blameViewport = viewport.New(availableWidth, blameHeight)
	m.detailViewport = viewport.New(detailWidth, detailHeight)
}

// resizeViewports adjusts viewport sizes
func (m *TimeMachineModel) resizeViewports() {
	// Re-run setup logic with new dimensions
	m.setupViewports()
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

// renderTimeline creates the timeline visualization
func (m *TimeMachineModel) renderTimeline() string {
	if len(m.timeline.Commits) == 0 {
		return ""
	}

	progress := m.timeline.GetProgress()
	current := m.timeline.GetCurrentCommit()

	// Timeline bar
	// Timeline bar - account for global padding (4) and labels
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

// renderBlameView creates the blame/code view with rock-solid column alignment
func (m *TimeMachineModel) renderBlameView() string {
	var lines []string

	suspiciousCommits := make(map[string]bool)
	for _, suspect := range m.bugSuspects {
		suspiciousCommits[suspect.Commit.Hash] = true
	}

	// Column Styles
	numStyle := lipgloss.NewStyle().Width(5).Align(lipgloss.Right).Foreground(lipgloss.Color("#666666"))
	sepStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#444444"))
	riskStyle := lipgloss.NewStyle().Width(3).Foreground(lipgloss.Color("#FF4444"))
	authorStyle := lipgloss.NewStyle().Width(15)
	dateStyle := lipgloss.NewStyle().Width(13).Foreground(lipgloss.Color("#888888"))
	codeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#E0E0E0"))

	// Overhead: 5(num) + 3(sep) + 3(risk) + 15(author) + 1(space) + 13(date) + 3(sep) = 43
	overhead := 43
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

		// Author
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
		fullLine := lipgloss.JoinHorizontal(lipgloss.Bottom, lNum, sep, risk, author, " ", date, sep, code)
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

// renderMainContent returns JUST the blame view box
func (m *TimeMachineModel) renderMainContent() string {
	// Box style with MaxWidth/MaxHeight for strict containment
	blameBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#4ECDC4")).
		Padding(0, 1).
		Width(m.blameViewport.Width).
		Height(m.blameViewport.Height).
		Render(m.blameViewport.View())

	return blameBox
}

// renderDetailBox returns the standalone detail box
func (m *TimeMachineModel) renderDetailBox() string {
	detailBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF6B6B")).
		Padding(0, 1). // Reduced padding to save space
		Width(m.detailViewport.Width).
		Height(m.detailViewport.Height).
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

// getHelpContent returns the help text content
func (m *TimeMachineModel) getHelpContent() string {
	return `
===== CODE TIME MACHINE - HELP =====

WHAT IS IT:
  Code Time Machine lets you travel through your code's history to see
  how it evolved over time. You can view line-by-line changes, identify
  who made each change, and spot potentially risky commits.

HOW TO USE:

  1. TRACKING HISTORY BOX (Top):
     - Shows line-by-line code with author and date information
     - Each line is color-coded by author for easy identification
     - Lines with '!' indicator show commits with higher bug risk
     - Use arrow keys to navigate through different commits

  2. COMMIT DETAILS BOX (Bottom):
     - Shows detailed information about the current commit
     - Displays: commit hash, author, date, message, and file stats
     - Risk analysis for potentially problematic commits

  3. TIMELINE (Middle):
     - Visual progress bar showing your position in commit history
     - Displays current commit number and date

KEYBOARD SHORTCUTS:

  Navigation:
    Left Arrow / H    Go to previous commit (newer)
    Right Arrow / L   Go to next commit (older)
    Home              Jump to newest commit (current)
    End               Jump to oldest commit (initial)

  Scrolling (in this help screen):
    Up / Down         Scroll help content
    Mouse Wheel       Scroll with mouse

  Actions:
    ?                 Toggle this help screen
    Q / Esc           Return to bonus features menu
    Ctrl+C            Exit the application

BUG RISK INDICATORS:
  !  High Risk   - Large refactors, late-night commits, WIP messages
  !  Medium Risk - Significant changes, quick fixes
     No mark     - Low risk, normal commits

TIPS:
  - Use the timeline to quickly see how far back in history you are
  - Different author names appear in different colors
  - The tracking history box shows the code as it was at that commit
  - Press ? again to close this help and return to the main view
  - Scroll up/down with arrow keys or mouse wheel in this help screen

========================================
`
}

// renderHelp shows the help screen
func (m *TimeMachineModel) renderHelp() string {
	helpBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#4ECDC4")).
		Padding(1)

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Padding(0, 1).
		Render("↑/↓ Scroll │ ? Close Help │ Q/Esc Back")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		helpBox.Render(m.helpViewport.View()),
		footer,
	)
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

// truncate shortens a string to max length using literal dots
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max < 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
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
