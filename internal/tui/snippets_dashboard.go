package tui

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/phravins/devcli/internal/snippets"
)

type SnippetsModel struct {
	storage      *snippets.Storage
	snippetsList []snippets.Snippet
	list         list.Model
	viewport     viewport.Model
	titleInput   textinput.Model
	descInput    textinput.Model
	langInput    textinput.Model
	codeInput    textarea.Model // Multi-line
	searchInput  textinput.Model
	saveInput    textinput.Model
	state        int
	selectedSnip *snippets.Snippet
	width        int
	height       int
	err          error

	// Help
	helpView viewport.Model

	// Animation
	fullContent string
	streamIndex int
}

type snTickMsg time.Time

func snTickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*20, func(t time.Time) tea.Msg {
		return snTickMsg(t)
	})
}

const (
	snStateList = iota
	snStateView
	snStateAdd
	snStateEdit
	snStateSearch
	snStateSave
	snStateHelp
)

type snippetsLoadedMsg struct {
	snippets []snippets.Snippet
	err      error
}

func NewSnippetsModel() SnippetsModel {
	storage, _ := snippets.NewStorage()

	ti := textinput.New()
	ti.Placeholder = "Snippet title"
	ti.Width = 50

	di := textinput.New()
	di.Placeholder = "Description"
	di.Width = 50

	li := textinput.New()
	li.Placeholder = "Language (e.g., go, python, javascript)"
	li.Width = 30

	si := textinput.New()
	si.Placeholder = "Search snippets..."
	si.Width = 50

	savi := textinput.New()
	savi.Placeholder = "Path to save (e.g. ./snippet.go)"
	savi.Width = 50

	vp := viewport.New(80, 20)

	return SnippetsModel{
		storage:     storage,
		list:        list.New([]list.Item{}, list.NewDefaultDelegate(), 60, 14),
		viewport:    vp,
		titleInput:  ti,
		descInput:   di,
		langInput:   li,
		searchInput: si,
		saveInput:   savi,
		helpView:    viewport.New(80, 20),
		state:       snStateList,
	}
}

func (m SnippetsModel) Init() tea.Cmd {
	return func() tea.Msg {
		snips, err := m.storage.LoadAll()
		if err != nil {
			return snippetsLoadedMsg{err: err}
		}
		if len(snips) == 0 {
			snips = snippets.GetDefaultSnippets()
		}
		return snippetsLoadedMsg{snippets: snips}
	}
}

func (m SnippetsModel) Update(msg tea.Msg) (SnippetsModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case snippetsLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.snippetsList = msg.snippets
		m.updateList(msg.snippets)
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case snStateList:
			switch msg.String() {
			case "esc", "q":
				return m, func() tea.Msg { return SubFeatureBackMsg{} }
			case "?":
				m.state = snStateHelp
				// Use consistent margins (like list/viewport) for help
				m.helpView.SetContent(RenderHelp(SnippetLibraryHelp, m.width-4, m.height))
				return m, nil

			case "r":
				return m, m.Init()
			case "a":
				// Add new snippet
				m.state = snStateAdd
				m.titleInput.SetValue("")
				m.descInput.SetValue("")
				m.langInput.SetValue("")
				m.titleInput.Focus()
				return m, textinput.Blink
			case "/":
				// Search
				m.state = snStateSearch
				m.searchInput.Focus()
				return m, textinput.Blink
			case "enter":
				// View snippet
				idx := m.list.Index()
				if idx >= 0 && idx < len(m.snippetsList) {
					m.selectedSnip = &m.snippetsList[idx]
					// Start animation
					m.fullContent = m.selectedSnip.Code
					m.streamIndex = 0
					m.viewport.SetContent("")
					m.state = snStateView
					return m, snTickCmd()
				}
			}
			m.list, cmd = m.list.Update(msg)
			return m, cmd

		case snStateView:
			switch msg.String() {
			case "esc", "q":
				m.state = snStateList
				return m, nil
			case "s":
				// Go to save mode
				m.state = snStateSave
				m.saveInput.SetValue("./" + m.selectedSnip.ID + ".txt") // Default attempt
				// Try to guess extension
				if m.selectedSnip.Language == "go" {
					m.saveInput.SetValue("./snippet.go")
				} else if m.selectedSnip.Language == "python" {
					m.saveInput.SetValue("./snippet.py")
				} else if m.selectedSnip.Language == "javascript" {
					m.saveInput.SetValue("./snippet.js")
				} else {
					m.saveInput.SetValue("./snippet.txt")
				}
				m.saveInput.Focus()
				return m, textinput.Blink
			case "d":
				// Delete snippet
				if m.selectedSnip != nil {
					m.storage.Delete(m.selectedSnip.ID)
					m.state = snStateList
					return m, m.Init()
				}
			}
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd

		case snStateSave:
			switch msg.String() {
			case "esc":
				m.state = snStateView
				m.saveInput.Blur()
				return m, nil
			case "enter":
				path := m.saveInput.Value()
				if path != "" {
					err := os.WriteFile(path, []byte(m.selectedSnip.Code), 0644)
					if err != nil {
						m.err = err
					} else {
						// Success feedback by going back? or flash?
						// For now just go back.
						m.state = snStateView
					}
				}
				return m, nil
			}
			m.saveInput, cmd = m.saveInput.Update(msg)
			return m, cmd

		case snStateSearch:
			switch msg.String() {
			case "esc":
				m.state = snStateList
				m.searchInput.Blur()
				m.updateList(m.snippetsList)
				return m, nil
			case "enter":
				query := m.searchInput.Value()
				results, _ := m.storage.Search(query)
				m.updateList(results)
				m.state = snStateList
				m.searchInput.Blur()
				return m, nil
			}
			m.searchInput, cmd = m.searchInput.Update(msg)
			return m, cmd

		case snStateHelp:
			if msg.String() == "esc" || msg.String() == "enter" || msg.String() == "q" || msg.String() == "?" {
				m.state = snStateList
				return m, nil
			}
			var cmd tea.Cmd
			m.helpView, cmd = m.helpView.Update(msg)
			return m, cmd
		}

	case tea.MouseMsg:
		var cmd tea.Cmd
		switch m.state {
		case snStateList:
			if msg.Type == tea.MouseWheelUp {
				m.list.CursorUp()
				return m, nil
			}
			if msg.Type == tea.MouseWheelDown {
				m.list.CursorDown()
				return m, nil
			}
			m.list, cmd = m.list.Update(msg)
			return m, cmd

		case snStateView:
			if msg.Type == tea.MouseWheelUp {
				m.viewport.LineUp(3)
				return m, nil
			}
			if msg.Type == tea.MouseWheelDown {
				m.viewport.LineDown(3)
				return m, nil
			}
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}

	case snTickMsg:
		if m.state == snStateView && m.streamIndex < len(m.fullContent) {
			// Speed: Add 3 chars per tick
			chunkSize := 3
			end := m.streamIndex + chunkSize
			if end > len(m.fullContent) {
				end = len(m.fullContent)
			}
			m.streamIndex = end
			m.viewport.SetContent(m.fullContent[:m.streamIndex])

			// Auto scroll to bottom while typing?
			// Usually for code display, we might want to stay at top or follow cursor.
			// Let's just set content. Viewport stays at top by default unless we move it.
			// But if content grows, we might want to ensure it's visible?
			// Actually, typical "hacker" effect writes linearly.
			// Let's keep it simple.
			return m, snTickCmd()
		} else if m.state == snStateView {
			// Animation done, ensure full content is set cleanly
			m.viewport.SetContent(m.fullContent)
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width-4, msg.Height-10)
		m.viewport.Width = msg.Width - 12
		m.viewport.Height = msg.Height - 16
		m.helpView.Width = msg.Width - 4
		m.helpView.Height = msg.Height
	}

	return m, nil
}

func (m *SnippetsModel) updateList(snips []snippets.Snippet) {
	items := make([]list.Item, len(snips))
	for i, snip := range snips {
		desc := fmt.Sprintf("%s | %s", snip.Language, snip.Description)
		items[i] = item{
			title: snip.Title,
			desc:  desc,
		}
	}
	m.list.SetItems(items)
}

func (m SnippetsModel) View() string {
	switch m.state {
	case snStateList:
		if len(m.snippetsList) == 0 {
			empty := "No snippets found\n\n" +
				"Press A to add a snippet\nPress R to refresh\nPress Esc to go back"
			return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
				lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(empty))
		}

		header := titleStyle.Render("Snippet Library")
		count := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).
			Render(fmt.Sprintf("%d snippets", len(m.snippetsList)))
		footer := subtleStyle.Render("Enter: View • A: Add • /: Search • R: Refresh • Esc: Back")

		content := lipgloss.JoinVertical(lipgloss.Left,
			"\n",
			header,
			count,
			"\n",
			m.list.View(),
			"\n",
			footer,
		)
		return docStyle.Render(content)

	case snStateView:
		header := titleStyle.Render(fmt.Sprintf("%s", m.selectedSnip.Title))
		meta := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).
			Render(fmt.Sprintf("Language: %s | Category: %s",
				m.selectedSnip.Language,
				m.selectedSnip.Category))
		desc := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).
			Render(m.selectedSnip.Description)

		// Framed Viewport
		// Ensure viewport content has a border
		codeView := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2).
			Render(m.viewport.View())

		footer := subtleStyle.Render("S: Save • D: Delete • Esc: Back • ↑/↓: Scroll")

		content := lipgloss.JoinVertical(lipgloss.Left,
			header,
			meta,
			desc,
			"\n",
			codeView,
			"\n",
			footer,
		)
		return docStyle.Render(content)

	case snStateSave:
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			lipgloss.JoinVertical(lipgloss.Center,
				titleStyle.Render("Save Snippet"),
				focusedInputBoxStyle.Render(m.saveInput.View()),
				subtleStyle.Render("(Enter path to save file)"),
			),
		)

	case snStateSearch:
		header := titleStyle.Render("Search Snippets")
		content := lipgloss.JoinVertical(lipgloss.Left,
			"\n",
			header,
			"\n",
			m.searchInput.View(),
			"\n",
			subtleStyle.Render("Enter: Search • Esc: Cancel"),
		)
		return docStyle.Render(content)

	case snStateHelp:
		helpWithBorder := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#0F9E99")).
			Render(m.helpView.View())
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, helpWithBorder)
	}

	return ""
}
