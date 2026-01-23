package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/phravins/devcli/internal/projectdash"
)

type ProjectDashModel struct {
	workspace string
	projects  []projectdash.ProjectInfo
	list      list.Model
	loading   bool
	err       error
	width     int
	height    int
	sortBy    string // "name", "date", "status"
}

type projectsLoadedMsg struct {
	projects []projectdash.ProjectInfo
	err      error
}

func NewProjectDashModel(workspace string) ProjectDashModel {
	return ProjectDashModel{
		workspace: workspace,
		loading:   true,
		sortBy:    "date",
		list:      list.New([]list.Item{}, list.NewDefaultDelegate(), 60, 14),
	}
}

func (m ProjectDashModel) Init() tea.Cmd {
	return func() tea.Msg {
		projects, err := projectdash.ScanWorkspace(m.workspace)
		return projectsLoadedMsg{projects: projects, err: err}
	}
}

func (m ProjectDashModel) Update(msg tea.Msg) (ProjectDashModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case projectsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.projects = msg.projects
		m.sortProjects()
		m.updateList()
		return m, nil

	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}

		switch msg.String() {
		case "esc", "q":
			return m, func() tea.Msg { return BackMsg{} }
		case "r":
			// Refresh
			m.loading = true
			return m, m.Init()
		case "s":
			// Cycle sort
			switch m.sortBy {
			case "name":
				m.sortBy = "date"
			case "date":
				m.sortBy = "status"
			case "status":
				m.sortBy = "name"
			}
			m.sortProjects()
			m.updateList()
			return m, nil
		}

		m.list, cmd = m.list.Update(msg)
		return m, cmd

	case tea.MouseMsg:
		switch msg.Type {
		case tea.MouseWheelUp:
			m.list.CursorUp()
		case tea.MouseWheelDown:
			m.list.CursorDown()
		}
		m.list, cmd = m.list.Update(msg)
		return m, cmd

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width-4, msg.Height-10)
	}

	return m, nil
}

func (m *ProjectDashModel) sortProjects() {
	switch m.sortBy {
	case "name":
		sort.Slice(m.projects, func(i, j int) bool {
			return m.projects[i].Name < m.projects[j].Name
		})
	case "date":
		sort.Slice(m.projects, func(i, j int) bool {
			return m.projects[i].LastModified.After(m.projects[j].LastModified)
		})
	case "status":
		sort.Slice(m.projects, func(i, j int) bool {
			return m.projects[i].Status < m.projects[j].Status
		})
	}
}

func (m *ProjectDashModel) updateList() {
	items := make([]list.Item, len(m.projects))
	for i, proj := range m.projects {
		techStr := strings.Join(proj.TechStack, ", ")
		if len(techStr) > 40 {
			techStr = techStr[:37] + "..."
		}

		desc := fmt.Sprintf("%s %s | %s | %s | %s",
			proj.Status.Icon(),
			proj.Status.String(),
			techStr,
			proj.LastModified.Format("2006-01-02"),
			projectdash.FormatSize(proj.Size),
		)

		items[i] = item{
			title: proj.Name,
			desc:  desc,
		}
	}
	m.list.SetItems(items)
}

func (m ProjectDashModel) View() string {
	if m.loading {
		return loadingStyle.Render("Scanning projects...")
	}

	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error: %v\n\nPress Esc to go back", m.err))
	}

	if len(m.projects) == 0 {
		empty := lipgloss.NewStyle().
			Align(lipgloss.Center).
			Foreground(lipgloss.Color("240")).
			Render("No projects found in workspace\n\nPress R to refresh • Esc to go back")
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, empty)
	}

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#0F9E99")).
		Render("Project Dashboard")

	sortInfo := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(fmt.Sprintf("Sorted by: %s (Press S to change)", m.sortBy))

	footer := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(m.width).
		Render(subtleStyle.Render("↑/↓: Navigate • R: Refresh • S: Sort • Esc: Back"))

	content := lipgloss.JoinVertical(lipgloss.Left,
		"\n",
		header,
		sortInfo,
		"\n",
		m.list.View(),
		"\n",
		footer,
	)

	return docStyle.Render(content)
}
