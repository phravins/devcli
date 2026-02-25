package tui

import (
	"fmt"
	"hash/fnv"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"github.com/phravins/devcli/internal/timemachine"
	"golang.org/x/term"
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

	// Detect real terminal window size at startup.
	// golang.org/x/term reads the actual console window dimensions,
	// not just the buffer width, which is what bubbletea sometimes reports on Windows CMD.
	initW, initH := 80, 40
	if w, h, err2 := term.GetSize(int(os.Stdout.Fd())); err2 == nil && w > 0 {
		initW, initH = w, h
	}

	helpVp := viewport.New(initW, initH)
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
		blameViewport:  viewport.New(initW, initH),
		detailViewport: viewport.New(initW, 5),
		helpViewport:   helpVp,
		width:          initW,
		height:         initH,
		ready:          true,
	}

	model.setupViewports()
	model.updateViewports()
	return model, nil
}

// Init initializes the model and immediately requests the current terminal size.
func (m *TimeMachineModel) Init() tea.Cmd {
	return tea.WindowSize()
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
				m.helpViewport.Width = m.width - 8
				m.helpViewport.Height = m.height - 6
				m.helpViewport.GotoTop()
			}
			return m, nil

		case "left", "h":
			if err := m.timeline.Previous(); err == nil {
				m.updateViewports()
			}
			return m, nil

		case "right", "l":
			if err := m.timeline.Next(); err == nil {
				m.updateViewports()
			}
			return m, nil

		case "home":
			if err := m.timeline.MoveToIndex(0); err == nil {
				m.updateViewports()
			}
			return m, nil

		case "end":
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
		if m.showHelp {
			m.helpViewport.Width = m.width - 8
			m.helpViewport.Height = m.height - 6
		}
		return m, nil
	}

	// Update viewports
	var cmd tea.Cmd
	if m.showHelp {
		m.helpViewport, cmd = m.helpViewport.Update(msg)
	} else {
		m.blameViewport, cmd = m.blameViewport.Update(msg)
	}
	return m, cmd
}

// View renders the UI using a flat layout that fills the full terminal width.
// No lipgloss.Place, no outer box — each section is rendered at terminal width.
func (m *TimeMachineModel) View() string {
	if !m.ready {
		return "Initializing Code Time Machine..."
	}

	if m.showHelp {
		return m.renderHelp()
	}

	w := m.width
	if w <= 0 {
		w = 80
	}

	// Header bar (2 lines)
	header := m.renderHeader(w)

	// Teal divider
	divider := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4ECDC4")).
		Render(strings.Repeat("─", w))

	// Timeline bar (1 line)
	timeline := m.renderTimeline(w)

	// Dark divider
	divider2 := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#333333")).
		Render(strings.Repeat("─", w))

	// Blame viewport (fills remaining height)
	blameView := m.blameViewport.View()

	// Commit info bar (1 line)
	commitBar := m.renderCommitDetailsCompact(w)

	// Dark divider
	divider3 := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#333333")).
		Render(strings.Repeat("─", w))

	// Footer (1 line)
	footer := m.renderFooter(w)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		divider,
		timeline,
		divider2,
		blameView,
		divider3,
		commitBar,
		footer,
	)
}

func (m *TimeMachineModel) setupViewports() {
	// Flat layout line counts:
	//   header:    2
	//   divider:   1
	//   timeline:  1
	//   divider:   1
	//   blameVP:   variable
	//   divider:   1
	//   commitBar: 1
	//   footer:    1
	//   TOTAL fixed = 9
	fixedLines := 9

	blameHeight := m.height - fixedLines
	if blameHeight < 6 {
		blameHeight = 6
	}

	viewportWidth := m.width
	if viewportWidth < 40 {
		viewportWidth = 40
	}

	m.blameViewport = viewport.New(viewportWidth, blameHeight)
	m.detailViewport = viewport.New(viewportWidth, 5)
}

func (m *TimeMachineModel) resizeViewports() {
	m.setupViewports()
}

func (m *TimeMachineModel) updateViewports() {
	m.blameViewport.SetContent(m.renderBlameView())
	m.detailViewport.SetContent(m.renderCommitDetails())
}

// renderHeader renders the 2-line header bar at full terminal width.
func (m *TimeMachineModel) renderHeader(w int) string {
	bgStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#1A1A2E")).
		Width(w)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FF6B6B")).
		Background(lipgloss.Color("#1A1A2E")).
		Padding(0, 1)

	fileStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6BCF7F")).
		Background(lipgloss.Color("#1A1A2E")).
		Padding(0, 1)

	title := bgStyle.Render(titleStyle.Render("Code Time Machine"))

	maxFileWidth := w - 4
	if maxFileWidth < 10 {
		maxFileWidth = 10
	}
	fPath := m.timeline.FilePath
	if runewidth.StringWidth(fPath) > maxFileWidth {
		fPath = truncate(fPath, maxFileWidth)
	}
	fileLine := bgStyle.Render(fileStyle.Render(fPath))

	return lipgloss.JoinVertical(lipgloss.Left, title, fileLine)
}

// renderTimeline renders a 1-line timeline progress bar at full terminal width.
func (m *TimeMachineModel) renderTimeline(w int) string {
	if len(m.timeline.Commits) == 0 {
		return strings.Repeat(" ", w)
	}

	progress := m.timeline.GetProgress()
	current := m.timeline.GetCurrentCommit()

	position := fmt.Sprintf("Commit %d/%d", m.timeline.CurrentIndex+1, len(m.timeline.Commits))
	dateStr := ""
	if current != nil {
		dateStr = current.Date.Format("Jan 02, 2006")
	}

	prefix := " ● "
	suffix := " ● "
	rightText := position + "  " + dateStr
	barWidth := w - len(prefix) - len(suffix) - len(rightText)
	if barWidth < 5 {
		barWidth = 5
	}

	filledWidth := int(float64(barWidth) * progress)
	if filledWidth > barWidth {
		filledWidth = barWidth
	}

	filled := strings.Repeat("█", filledWidth)
	empty := strings.Repeat("░", barWidth-filledWidth)

	bar := lipgloss.NewStyle().Foreground(lipgloss.Color("#4ECDC4")).Render(filled) +
		lipgloss.NewStyle().Foreground(lipgloss.Color("#333333")).Render(empty)

	prefixS := lipgloss.NewStyle().Foreground(lipgloss.Color("#4ECDC4")).Bold(true).Render(prefix)
	suffixS := lipgloss.NewStyle().Foreground(lipgloss.Color("#4ECDC4")).Bold(true).Render(suffix)
	rightS := lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")).Render(rightText)

	return prefixS + bar + suffixS + rightS
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

	// Fixed columns: linenum(5) + " | "(3) + risk(3) + author(15) + " "(1) + date(13) + " | "(3) + trailing " | "(3) = 46
	overhead := 46
	availableCodeWidth := m.blameViewport.Width - overhead
	if availableCodeWidth < 10 {
		availableCodeWidth = 10
	}

	for _, line := range m.timeline.BlameData {
		lNum := fmt.Sprintf("%5d", line.LineNumber)
		lNum = numStyle.Render(lNum)

		sep := sepStyle.Render(" │ ")

		rStr := "   "
		if suspiciousCommits[line.CommitHash] {
			rStr = "!  "
		}
		risk := riskStyle.Render(rStr)

		aName := truncate(line.Author, 15)
		aName = runewidth.FillRight(aName, 15)
		author := authorStyle.Foreground(m.authorColors[line.Author]).Render(aName)

		dStr := line.Timestamp.Format("Jan 02 15:04")
		dStr = runewidth.FillRight(dStr, 13)
		date := dateStyle.Render(dStr)

		cStr := strings.ReplaceAll(line.Content, "\t", "  ")
		if runewidth.StringWidth(cStr) > availableCodeWidth {
			cStr = truncate(cStr, availableCodeWidth)
		}
		cStr = runewidth.FillRight(cStr, availableCodeWidth)
		code := codeStyle.Render(cStr)

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

	hashStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFD700")).
		Bold(true)

	authorStyle := lipgloss.NewStyle().Foreground(m.authorColors[current.Author])

	details = append(details, hashStyle.Render("Commit: ")+current.ShortHash)

	authorDateLine := "Author: " + authorStyle.Render(current.Author) +
		"    Date: " + current.Date.Format("Mon Jan 02, 2006 15:04")
	details = append(details, authorDateLine)
	details = append(details, "")

	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true)

	details = append(details, messageStyle.Render("Message:"))
	for _, line := range strings.Split(current.Message, "\n") {
		details = append(details, "  "+line)
	}

	statsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#4ECDC4"))
	statsLine := fmt.Sprintf("Files changed: %d    +%d -%d lines",
		len(current.FilesChanged), current.LinesAdded, current.LinesRemoved)
	details = append(details, statsStyle.Render(statsLine))

	for _, suspect := range m.bugSuspects {
		if suspect.Commit.Hash == current.Hash {
			riskStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(timemachine.GetRiskColor(suspect.Risk))).
				Bold(true)

			riskLevel := timemachine.GetRiskLevel(suspect.Risk)
			details = append(details, riskStyle.Render(fmt.Sprintf("[RISK] %s (%.0f%%)", riskLevel, suspect.Risk*100)))
			details = append(details, "Reason: "+suspect.Reason)
		}
	}

	return strings.Join(details, "\n")
}

// renderCommitDetailsCompact creates a compact one-line commit summary at full terminal width.
func (m *TimeMachineModel) renderCommitDetailsCompact(w int) string {
	current := m.timeline.GetCurrentCommit()
	if current == nil {
		return strings.Repeat(" ", w)
	}

	hashStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700")).Bold(true)
	authorStyle := lipgloss.NewStyle().Foreground(m.authorColors[current.Author])
	msgStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))
	statsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#4ECDC4"))

	msgLines := strings.Split(current.Message, "\n")
	firstLineMsg := msgLines[0]
	// hash ~8 + author ~15 + date ~22 + stats ~10 + separators ~6 = ~61 reserved
	maxMsg := w - 61
	if maxMsg < 10 {
		maxMsg = 10
	}
	if runewidth.StringWidth(firstLineMsg) > maxMsg {
		firstLineMsg = truncate(firstLineMsg, maxMsg)
	}

	statsStr := fmt.Sprintf("+%d -%d", current.LinesAdded, current.LinesRemoved)
	dateStr := current.Date.Format("Jan 02, 2006 15:04")

	return hashStyle.Render("● "+current.ShortHash) + "  " +
		authorStyle.Render(current.Author) + "  " +
		msgStyle.Render(firstLineMsg) + "  " +
		statsStyle.Render(statsStr) + "  " +
		lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(dateStr)
}

// renderMainContent returns the tracking history box (blame view)
func (m *TimeMachineModel) renderMainContent() string {
	blameBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#4ECDC4")).
		Padding(0, 1).
		Width(m.blameViewport.Width + 4).
		Height(m.blameViewport.Height + 2).
		Render(m.blameViewport.View())

	return blameBox
}

// renderDetailBox returns the commit details box
func (m *TimeMachineModel) renderDetailBox() string {
	detailBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF6B6B")).
		Padding(0, 0).
		Width(m.detailViewport.Width + 2).
		Render(m.detailViewport.View())

	return detailBox
}

// renderFooter creates the footer with shortcuts at full terminal width.
func (m *TimeMachineModel) renderFooter(w int) string {
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#555555")).
		Background(lipgloss.Color("#0D0D1A")).
		Width(w)

	shortcuts := " ←/→ Navigate │ Home/End Jump │ ? Help │ Q/Esc Back │ Ctrl+C Quit"
	return footerStyle.Render(shortcuts)
}

// renderHelp shows the help screen
func (m *TimeMachineModel) renderHelp() string {
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.helpViewport.View())
}

// generateAuthorColors creates consistent colors for authors
func generateAuthorColors(authors []string) map[string]lipgloss.Color {
	colors := map[string]lipgloss.Color{}

	palette := []string{
		"#FF6B6B", "#4ECDC4", "#45B7D1", "#FFA07A",
		"#98D8C8", "#F7DC6F", "#BB8FCE", "#85C1E2",
		"#F8B4D1", "#52D3AA", "#FDA7DF", "#87CEEB",
	}

	for i, author := range authors {
		if i < len(palette) {
			colors[author] = lipgloss.Color(palette[i])
		} else {
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

	r := (hash & 0xFF0000) >> 16
	g := (hash & 0x00FF00) >> 8
	b := hash & 0x0000FF

	r = (r % 156) + 100
	g = (g % 156) + 100
	b = (b % 156) + 100

	return lipgloss.Color(fmt.Sprintf("#%02X%02X%02X", r, g, b))
}

// truncate shortens a string to max valid width using go-runewidth
func truncate(s string, max int) string {
	return runewidth.Truncate(s, max, "...")
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
