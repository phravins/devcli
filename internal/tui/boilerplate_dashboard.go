package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/phravins/devcli/internal/boilerplate"
)

// Boilerplate States
const (
	StateBPMenu = iota
	StateBPSnippets
	StateBPLanguage // Select language for snippet
	StateBPTemplates
	StateBPArchList
	StateBPSuccess
	StateBPSaveTemplate  // Input for saving custom template
	StateBPSelectProject // Select folder to save as template
	StateBPLoadTemplate  // Input for naming new project from template
	StateBPInputPath     // Input for destination path
	StateBPShowResult    // Show what was generated
	StateBPHelp          // Help screen
)

type BoilerplateDashboardModel struct {
	menuList     list.Model
	snippetList  list.Model
	languageList list.Model // Select language
	templateList list.Model
	archList     list.Model
	input        textinput.Model // For naming custom templates
	loadInput    textinput.Model // For naming new project
	pathInput    textinput.Model // For selecting destination
	projectList  list.Model      // For selecting source project
	viewport     viewport.Model  // For showing results
	helpView     viewport.Model  // For help content

	state         int
	width, height int
	manager       *boilerplate.Manager
	tplManager    *boilerplate.TemplateManager

	selectedTemplate string
	selectedProject  string
	selectedItem     string // Generic selection (Snippet Name or Arch Name)
	targetLang       string // For snippet language

	statusMsg   string
	err         error
	fullContent string // For streaming effect
	streamIndex int    // Current position in stream
}

type bpTickMsg time.Time

func bpTickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*20, func(t time.Time) tea.Msg {
		return bpTickMsg(t)
	})
}

func NewBoilerplateDashboardModel(workspace string) BoilerplateDashboardModel {
	mgr := boilerplate.NewManager(workspace)
	home, _ := os.UserHomeDir()
	tplMgr := boilerplate.NewTemplateManager(home)

	// 1. Main Menu
	menuItems := []list.Item{
		item{title: "Code Snippet Presets", desc: "Ready-to-drop blocks (CRUD, Auth, DB Connection)"},
		item{title: "Custom Boilerplate Templates", desc: "Save/Load your own folder structures"},
		item{title: "Architecture Generator", desc: "Build structures like MVC, Clean Arch, etc."},
	}
	menu := list.New(menuItems, list.NewDefaultDelegate(), 0, 0)
	menu.Title = "Boilerplate Generator"
	menu.SetShowHelp(false)

	// 2. Snippets List
	var snipItems []list.Item
	for key, s := range boilerplate.Snippets {
		snipItems = append(snipItems, item{id: key, title: s.Name, desc: s.Description})
	}
	// Sort for stability
	sort.Slice(snipItems, func(i, j int) bool { return snipItems[i].(item).title < snipItems[j].(item).title })

	snipList := list.New(snipItems, list.NewDefaultDelegate(), 0, 0)
	snipList.Title = "Select Snippet"
	snipList.SetShowHelp(false)

	// 2.5 Language List (Empty initially)
	langList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	langList.Title = "Select Language"
	langList.SetShowHelp(false)

	// 3. Architecture List
	var archItems []list.Item
	for key, a := range boilerplate.Architectures {
		archItems = append(archItems, item{id: key, title: a.Name, desc: a.Description})
	}
	archList := list.New(archItems, list.NewDefaultDelegate(), 0, 0)
	archList.Title = "Select Architecture"
	archList.SetShowHelp(false)

	// 4. Template List (Dynamic, but init empty)
	tplList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	tplList.Title = "Custom Templates"
	tplList.SetShowHelp(false)

	ti := textinput.New()
	ti.Placeholder = "Template Name"
	ti.CharLimit = 50
	ti.Width = 40

	li := textinput.New()
	li.Placeholder = "New Project Name"
	li.CharLimit = 50
	li.Width = 40

	pi := textinput.New()
	pi.Placeholder = "Destination Path (e.g. . or ./myfolder)"
	pi.SetValue("./") // Default to current dir (more user friendly)
	pi.CharLimit = 100
	pi.Width = 50

	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("62"))

	// Project List for selection
	pl := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	pl.Title = "Select Project to Save"
	pl.SetShowHelp(false)

	// Help View
	hv := viewport.New(80, 20)
	// Content is set dynamically using RenderHelp

	return BoilerplateDashboardModel{
		menuList:     menu,
		snippetList:  snipList,
		languageList: langList,
		archList:     archList,
		templateList: tplList,
		input:        ti,
		loadInput:    li,
		pathInput:    pi,
		projectList:  pl,
		viewport:     vp,
		helpView:     hv,
		state:        StateBPMenu,
		manager:      mgr,
		tplManager:   tplMgr,
	}
}

func (m BoilerplateDashboardModel) Init() tea.Cmd {
	return nil
}

func (m BoilerplateDashboardModel) Update(msg tea.Msg) (BoilerplateDashboardModel, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global Help Toggle (unless input focused)
		if msg.String() == "?" && m.state != StateBPSaveTemplate && m.state != StateBPLoadTemplate && m.state != StateBPInputPath && m.state != StateBPHelp {
			m.state = StateBPHelp
			m.helpView.SetContent(RenderHelp(BoilerplateHelp, m.width-4, m.height))
			return m, nil
		}

		switch m.state {
		case StateBPHelp:
			if msg.String() == "esc" || msg.String() == "q" || msg.String() == "?" || msg.String() == "enter" {
				m.state = StateBPMenu
				return m, nil
			}
			var cmd tea.Cmd
			m.helpView, cmd = m.helpView.Update(msg)
			return m, cmd

		case StateBPMenu:
			switch msg.String() {
			case "enter":
				i, ok := m.menuList.SelectedItem().(item)
				if ok {
					switch i.title {
					case "Code Snippet Presets":
						m.state = StateBPSnippets
					case "Custom Boilerplate Templates":
						// Refresh templates before showing
						m.refreshTemplates()
						m.state = StateBPTemplates
					case "Architecture Generator":
						m.state = StateBPArchList
					}
					return m, nil
				}
			case "esc":
				// User wants to go back to Project Tools
				// We don't handle that here effectively, purely internal state.
				// The parent handles "esc" if we don't catch it?
				// We'll return a special Loop/Signal if needed,
				// but essentially we are a sub-view.
				// If we are at root Top Level (Menu), allow parent to take back control?
				// Returning the model as-is lets parent see we didn't handle it?
				// Actually standard pattern: return BackMsg
				return m, func() tea.Msg { return BoilerplateBackMsg{} }

			}
			m.menuList, cmd = m.menuList.Update(msg)
			return m, cmd

		case StateBPSnippets:
			switch msg.String() {
			case "enter":
				i, ok := m.snippetList.SelectedItem().(item)
				if ok {
					// 1. Store the ID (key) of the snippet
					m.selectedItem = i.id

					// 2. Fetch the snippet to get available languages
					snip, exists := boilerplate.Snippets[i.id]
					if !exists {
						m.err = fmt.Errorf("Snippet logic error: ID %s not found", i.id)
						return m, nil
					}

					// 3. Populate Language List
					var langItems []list.Item
					for lang := range snip.Content {
						langItems = append(langItems, item{title: lang, desc: "Generate in " + lang})
					}
					// Sort languages
					sort.Slice(langItems, func(a, b int) bool {
						return langItems[a].(item).title < langItems[b].(item).title
					})
					m.languageList.SetItems(langItems)

					// 4. Transition to Language Select
					m.state = StateBPLanguage
					return m, nil
				}
			case "esc":
				m.state = StateBPMenu
				return m, nil
			}
			m.snippetList, cmd = m.snippetList.Update(msg)
			return m, cmd

		case StateBPLanguage:
			switch msg.String() {
			case "enter":
				i, ok := m.languageList.SelectedItem().(item)
				if ok {
					m.targetLang = i.title
					m.state = StateBPInputPath
					m.pathInput.SetValue("./") // Default to explicit relative CWD
					m.pathInput.Focus()
					return m, nil
				}
			case "esc":
				m.state = StateBPSnippets
				return m, nil
			}
			m.languageList, cmd = m.languageList.Update(msg)
			return m, cmd

		case StateBPArchList:
			switch msg.String() {
			case "enter":
				i, ok := m.archList.SelectedItem().(item)
				if ok {
					// Go to Path Input
					m.selectedItem = i.id
					m.state = StateBPInputPath
					m.pathInput.SetValue("./")
					m.pathInput.Focus()
					return m, nil
				}
			case "esc":
				m.state = StateBPMenu
				return m, nil
			}
			m.archList, cmd = m.archList.Update(msg)
			return m, cmd

		case StateBPTemplates:
			switch msg.String() {
			case "enter":
				i, ok := m.templateList.SelectedItem().(item)
				if ok {
					if i.title == "+ Save New Template" { // Changed text slightly to match logic
						// 1. Refresh project list
						m.refreshProjectList()
						m.state = StateBPSelectProject
						return m, nil
					} else {
						// Load this template -> Ask for new project name
						m.selectedTemplate = i.title // Store state
						m.state = StateBPLoadTemplate
						m.loadInput.SetValue("")
						m.loadInput.Focus()
						return m, nil
					}
				}
			case "d":
				if m.templateList.FilterState() != list.Filtering {
					i, ok := m.templateList.SelectedItem().(item)
					if ok && i.title != "+ Save New Template" {
						err := m.tplManager.DeleteTemplate(i.title)
						if err != nil {
							m.err = err
						}
						// Refresh regardless to show updated list
						m.refreshTemplates()
						return m, nil
					}
				}
			case "esc":
				m.state = StateBPMenu
				return m, nil
			}
			m.templateList, cmd = m.templateList.Update(msg)
			return m, cmd

		case StateBPSelectProject:
			switch msg.String() {
			case "enter":
				i, ok := m.projectList.SelectedItem().(item)
				if ok {
					m.selectedProject = i.title
					m.state = StateBPSaveTemplate
					m.input.SetValue("")
					m.input.Focus()
					return m, nil
				}
			case "esc":
				m.state = StateBPTemplates
				return m, nil
			}
			m.projectList, cmd = m.projectList.Update(msg)
			return m, cmd

		case StateBPLoadTemplate:
			switch msg.String() {
			case "enter":
				name := m.loadInput.Value()
				if name != "" {
					// Create destination path
					destPath := filepath.Join(m.manager.Workspace, name)
					if _, err := os.Stat(destPath); err == nil {
						m.err = fmt.Errorf("directory '%s' already exists", name)
						return m, nil
					}

					// Load
					err := m.tplManager.LoadTemplate(m.selectedTemplate, destPath)
					if err != nil {
						m.err = err
					} else {
						m.statusMsg = fmt.Sprintf("Created project '%s' from template '%s'!", name, m.selectedTemplate)
						m.state = StateBPSuccess
						return m, tea.Tick(2*time.Second, func(_ time.Time) tea.Msg { return BoilerplateBackMsg{} })

					}
				}
			case "esc":
				m.state = StateBPTemplates
				return m, nil
			}
			m.loadInput, cmd = m.loadInput.Update(msg)
			return m, cmd

		case StateBPInputPath:
			switch msg.String() {
			case "enter":
				pathVal := m.pathInput.Value()
				// Create dir if not exists
				os.MkdirAll(pathVal, 0755)

				// Determine what we were doing based on selected item lookup
				// Hacky? logic check
				if _, ok := boilerplate.Snippets[m.selectedItem]; ok {
					// Generating Snippet
					path, err := m.manager.GenerateSnippet(m.selectedItem, m.targetLang, pathVal)
					if err != nil {
						m.err = err
					} else {
						// Read content to show
						content, _ := os.ReadFile(path)
						m.fullContent = string(content)
						m.streamIndex = 0
						m.viewport.SetContent("")
						absPath, _ := filepath.Abs(path)
						m.statusMsg = fmt.Sprintf("Generated: %s", absPath)
						m.state = StateBPShowResult
						return m, bpTickCmd()
					}
				} else if _, ok := boilerplate.Architectures[m.selectedItem]; ok {
					// Generating Architecture
					paths, err := m.manager.GenerateArchitecture(m.selectedItem, pathVal)
					if err != nil {
						m.err = err
					} else {
						// Build tree string
						var sb strings.Builder
						sb.WriteString(fmt.Sprintf("Generated %d files in %s:\n\n", len(paths), pathVal))
						for _, p := range paths {
							rel, _ := filepath.Rel(pathVal, p)
							sb.WriteString(fmt.Sprintf(" %s\n", rel))
						}
						m.fullContent = sb.String()
						m.streamIndex = 0
						m.viewport.SetContent("")
						absPath, _ := filepath.Abs(pathVal)
						m.statusMsg = fmt.Sprintf("Generated Architecture in: %s", absPath)
						m.state = StateBPShowResult
						return m, bpTickCmd()
					}
				} else {
					// Only Snippet and Arch use this flow currently
					m.state = StateBPMenu
				}
				return m, nil

			case "esc":
				m.state = StateBPMenu
				return m, nil
			}
			m.pathInput, cmd = m.pathInput.Update(msg)
			return m, cmd

		case StateBPShowResult:
			switch msg.String() {
			case "enter", "esc":
				m.state = StateBPMenu
				return m, nil
			}
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd

		case StateBPSaveTemplate:
			switch msg.String() {
			case "enter":
				name := m.input.Value()
				if name != "" {
					srcPath := filepath.Join(m.manager.Workspace, m.selectedProject)
					err := m.tplManager.SaveAsTemplate(name, srcPath)
					if err != nil {
						m.err = err
					} else {
						m.statusMsg = fmt.Sprintf("Saved template '%s'!", name)
						m.state = StateBPSuccess
						return m, tea.Tick(2*time.Second, func(_ time.Time) tea.Msg { return BoilerplateBackMsg{} })

					}
				}
			case "esc":
				m.state = StateBPTemplates
				return m, nil
			}
			m.input, cmd = m.input.Update(msg)
			return m, cmd

		case StateBPSuccess:
			if msg.String() == "enter" || msg.String() == "esc" {
				m.state = StateBPMenu
				return m, nil
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizeLists(msg.Width, msg.Height)
	case BoilerplateBackMsg:
		// If we get a Tick backmsg, we reset
		if m.state == StateBPSuccess {
			m.state = StateBPMenu
		}
		return m, nil

	case bpTickMsg:
		if m.state == StateBPShowResult && m.streamIndex < len(m.fullContent) {
			// Speed: Add 2 chars per tick (smoother animation)
			chunkSize := 2
			end := m.streamIndex + chunkSize
			if end > len(m.fullContent) {
				end = len(m.fullContent)
			}
			m.streamIndex = end
			m.viewport.SetContent(m.fullContent[:m.streamIndex])
			m.viewport.GotoBottom()
			return m, bpTickCmd()
		}

	case tea.MouseMsg:
		var cmd tea.Cmd

		// Manual Handling for reliability
		if msg.Type == tea.MouseWheelUp {
			switch m.state {
			case StateBPMenu:
				m.menuList.CursorUp()
			case StateBPSnippets:
				m.snippetList.CursorUp()
			case StateBPLanguage:
				m.languageList.CursorUp()
			case StateBPArchList:
				m.archList.CursorUp()
			case StateBPTemplates:
				m.templateList.CursorUp()
			case StateBPSelectProject:
				m.projectList.CursorUp()
			case StateBPShowResult:
				m.viewport.LineUp(3)
			case StateBPHelp:
				m.helpView.LineUp(3)
			}
			// Fallthrough to regular update just in case, but usually return
			return m, nil
		}
		if msg.Type == tea.MouseWheelDown {
			switch m.state {
			case StateBPMenu:
				m.menuList.CursorDown()
			case StateBPSnippets:
				m.snippetList.CursorDown()
			case StateBPLanguage:
				m.languageList.CursorDown()
			case StateBPArchList:
				m.archList.CursorDown()
			case StateBPTemplates:
				m.templateList.CursorDown()
			case StateBPSelectProject:
				m.projectList.CursorDown()
			case StateBPShowResult:
				m.viewport.LineDown(3)
			case StateBPHelp:
				m.helpView.LineDown(3)
			}
			return m, nil
		}

		switch m.state {
		case StateBPMenu:
			m.menuList, cmd = m.menuList.Update(msg)
		case StateBPSnippets:
			m.snippetList, cmd = m.snippetList.Update(msg)
		case StateBPLanguage:
			m.languageList, cmd = m.languageList.Update(msg)
		case StateBPArchList:
			m.archList, cmd = m.archList.Update(msg)
		case StateBPTemplates:
			m.templateList, cmd = m.templateList.Update(msg)
		case StateBPSelectProject:
			m.projectList, cmd = m.projectList.Update(msg)
		case StateBPShowResult:
			m.viewport, cmd = m.viewport.Update(msg)
		case StateBPHelp:
			m.helpView, cmd = m.helpView.Update(msg)
		}
		return m, cmd
	}

	return m, nil
}

func (m *BoilerplateDashboardModel) refreshTemplates() {
	tpls, _ := m.tplManager.ListTemplates()
	items := []list.Item{
		item{title: "+ Save New Template", desc: "Save a project from workspace as template"},
	}
	for _, t := range tpls {
		items = append(items, item{title: t.Name, desc: "Created: " + t.CreatedAt.Format("2006-01-02")})
	}
	m.templateList.SetItems(items)
}

func (m *BoilerplateDashboardModel) resizeLists(w, h int) {
	// Simple resize logic
	m.menuList.SetSize(w, h-4)
	m.snippetList.SetSize(w, h-4)
	m.languageList.SetSize(w, h-4)
	m.templateList.SetSize(w, h-4)
	m.archList.SetSize(w, h-4)
	m.projectList.SetSize(w, h-4)
	m.viewport.Width = w - 4
	m.viewport.Height = h - 6

	m.helpView.Width = w
	m.helpView.Height = h
	if m.state == StateBPHelp {
		m.helpView.SetContent(RenderHelp(BoilerplateHelp, m.width-2, m.height)) // Use closer width
	}
}

func (m *BoilerplateDashboardModel) refreshProjectList() {
	entries, err := os.ReadDir(m.manager.Workspace)
	if err != nil {
		return
	}
	var items []list.Item
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			items = append(items, item{title: e.Name(), desc: "Project Folder"})
		}
	}
	m.projectList.SetItems(items)
}

func (m BoilerplateDashboardModel) View() string {
	if m.state == StateBPSuccess {
		return successBoxStyle.Render(m.statusMsg + "\n\n(Press Enter)")
	}
	if m.err != nil {
		return errorBoxStyle.Render(fmt.Sprintf("Error: %v\n\n(Press Esc)", m.err))
	}

	switch m.state {
	case StateBPMenu:
		header := lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(
			titleStyle.Render("Boilerplate Generator"),
		)
		return docStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
			header,
			"\n",
			m.menuList.View(),
		))
	case StateBPSnippets:
		return docStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
			titleStyle.Render("Generate Code Snippet"),
			m.snippetList.View(),
		))
	case StateBPLanguage:
		return docStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
			titleStyle.Render("Select Language"),
			m.languageList.View(),
		))
	case StateBPArchList:
		return docStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
			titleStyle.Render("Generate Architecture"),
			m.archList.View(),
		))
	case StateBPTemplates:
		return docStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
			titleStyle.Render("Custom Templates")+subtleStyle.Render("  (d: Delete, Enter: Select)"),
			m.templateList.View(),
		))
	case StateBPSelectProject:
		return docStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
			titleStyle.Render("Select Project to Save"),
			m.projectList.View(),
		))
	case StateBPSaveTemplate:
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			lipgloss.JoinVertical(lipgloss.Center,
				titleStyle.Render("Save New Template"),
				focusedInputBoxStyle.Render(m.input.View()),
				subtleStyle.Render("(Enter name for your template)"),
			),
		)
	case StateBPLoadTemplate:
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			lipgloss.JoinVertical(lipgloss.Center,
				titleStyle.Render("Create Project from Template"),
				focusedInputBoxStyle.Render(m.loadInput.View()),
				subtleStyle.Render("(Enter name for new project)"),
			),
		)
	case StateBPInputPath:
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			lipgloss.JoinVertical(lipgloss.Center,
				titleStyle.Render("Select Destination"),
				focusedInputBoxStyle.Render(m.pathInput.View()),
				subtleStyle.Render("(Enter path to generate in)"),
			),
		)
	case StateBPShowResult:
		// Full screen result viewer
		return docStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
			titleStyle.Render(m.statusMsg),
			m.viewport.View(),
			subtleStyle.Render("\n(Press Enter to finish)"),
		))
	case StateBPHelp:
		helpWithBorder := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#0F9E99")).
			Render(m.helpView.View())

		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, helpWithBorder)
	}
	return "Unknown State"
}
