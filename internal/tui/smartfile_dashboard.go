package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/phravins/devcli/internal/ai"
	"github.com/phravins/devcli/internal/ai/providers"
	"github.com/phravins/devcli/internal/config"
	"github.com/phravins/devcli/internal/smartfile"
)

type SmartFileModel struct {
	workspace     string
	templateList  list.Model
	input         textinput.Model
	languageInput textinput.Model
	aiPromptInput textinput.Model
	preview       viewport.Model
	spinner       spinner.Model
	provider      ai.Provider
	state         int // 0: select template, 1: customize, 2: preview, 3: save
	selectedTpl   *smartfile.FileTemplate
	options       map[string]string
	savePath      string
	customPrompt  string
	result        string
	width         int
	height        int
	err           error

	// Help
	helpView viewport.Model
}

const (
	sfStateSelectTemplate = iota
	sfStateCustomize
	sfStateAIPrompt
	sfStateFilename
	sfStatePreview
	sfStateSave
	sfStateGenerating
	sfStateSuccess
	sfStateHelp
)

func NewSmartFileModel(workspace string) SmartFileModel {
	items := make([]list.Item, len(smartfile.Templates))
	for i, tpl := range smartfile.Templates {
		items[i] = item{
			title: tpl.Name,
			desc:  tpl.Description,
		}
	}

	lst := list.New(items, list.NewDefaultDelegate(), 60, 14)
	lst.Title = "Select File Template"
	lst.SetShowHelp(false)

	// Input for customization
	ti := textinput.New()
	ti.Placeholder = "Enter value"
	ti.Width = 45

	langInput := textinput.New()
	langInput.Placeholder = "e.g., Go, Python, Node.js"
	langInput.Width = 30 // 30 is fine

	aiInput := textinput.New()
	aiInput.Placeholder = "What should this file do? (e.g., 'API client for Stripe')"
	aiInput.Width = 45

	// Preview viewport
	vp := viewport.New(80, 20)

	// Spinner for generation process
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	// Help Viewport
	hv := viewport.New(80, 20)

	cfg, _ := config.LoadConfig()
	p, _ := providers.GetProvider(cfg)

	return SmartFileModel{
		workspace:     workspace,
		templateList:  lst,
		input:         ti,
		languageInput: langInput,
		aiPromptInput: aiInput,
		preview:       vp,
		spinner:       s,
		provider:      p,
		state:         sfStateSelectTemplate,
		helpView:      hv,
		options:       make(map[string]string),
	}
}

func (m SmartFileModel) Init() tea.Cmd {
	return nil
}

func (m SmartFileModel) Update(msg tea.Msg) (SmartFileModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case sfStateSelectTemplate:
			switch msg.String() {
			case "esc", "q":
				return m, func() tea.Msg { return SubFeatureBackMsg{} }
			case "?":
				m.state = sfStateHelp
				m.helpView.SetContent(RenderHelp(SmartFileHelp, m.width, m.height))
				return m, nil

			case "enter":
				idx := m.templateList.Index()
				if idx >= 0 && idx < len(smartfile.Templates) {
					m.selectedTpl = &smartfile.Templates[idx]

					if m.selectedTpl.Name == "Custom File (AI)" {
						m.state = sfStateFilename
						m.input.Placeholder = "Enter filename (e.g., helper.py)"
						m.input.SetValue("")
						m.input.Focus()
						return m, textinput.Blink
					}

					// Some templates need language selection
					needsLang := m.selectedTpl.Name == ".gitignore" ||
						m.selectedTpl.Name == "Dockerfile" ||
						m.selectedTpl.Name == "Makefile" ||
						m.selectedTpl.Name == "GitHub Actions"

					if needsLang {
						m.state = sfStateCustomize
						m.languageInput.Focus()
						return m, textinput.Blink
					} else {
						m.state = sfStateGenerating
						return m, tea.Batch(m.spinner.Tick, m.generateStandardPreviewCmd())
					}
				}
			}
			m.templateList, cmd = m.templateList.Update(msg)
			return m, cmd

		case sfStateFilename:
			switch msg.String() {
			case "esc":
				m.resetState()
				return m, nil
			case "enter":
				if m.input.Value() != "" {
					m.savePath = filepath.Join(m.workspace, m.input.Value())
					m.state = sfStateAIPrompt
					m.aiPromptInput.Focus()
					return m, textinput.Blink
				}
			}
			m.input, cmd = m.input.Update(msg)
			return m, cmd

		case sfStateAIPrompt:
			switch msg.String() {
			case "esc":
				m.state = sfStateFilename
				m.aiPromptInput.Blur()
				return m, nil
			case "enter":
				if m.aiPromptInput.Value() != "" {
					m.customPrompt = m.aiPromptInput.Value()
					m.state = sfStateGenerating
					m.aiPromptInput.Blur()
					return m, tea.Batch(m.spinner.Tick, m.generateAIFileCmd())
				}
			}
			m.aiPromptInput, cmd = m.aiPromptInput.Update(msg)
			return m, cmd

		case sfStateCustomize:
			switch msg.String() {
			case "esc":
				m.resetState()
				return m, nil
			case "enter":
				m.options["language"] = m.languageInput.Value()
				if m.selectedTpl.Name == ".env" {
					m.options["app_name"] = filepath.Base(m.workspace)
				}
				m.state = sfStateGenerating
				m.languageInput.Blur()
				return m, tea.Batch(m.spinner.Tick, m.generateStandardPreviewCmd())
			}
			m.languageInput, cmd = m.languageInput.Update(msg)
			return m, cmd

		case sfStatePreview:
			switch msg.String() {
			case "esc":
				if m.selectedTpl.Name == "Custom File (AI)" {
					m.state = sfStateAIPrompt
					return m, nil
				}
				m.state = sfStateCustomize
				return m, nil
			case "enter":
				m.state = sfStateSave

				// Logic: If Custom File, ask verify full path.
				// If Static Template, ask for Directory only.
				isCustom := m.selectedTpl.Name == "Custom File (AI)"

				if isCustom {
					m.input.Placeholder = "Enter full save path..."
					if m.savePath == "" {
						m.savePath = filepath.Join(m.workspace, m.input.Value())
					}
					m.input.SetValue(m.savePath)
				} else {
					m.input.Placeholder = "Enter destination folder..."
					// Default to workspace root
					m.input.SetValue(m.workspace)
				}

				m.input.Focus()
				return m, textinput.Blink
			}
			m.preview, cmd = m.preview.Update(msg)
			return m, cmd

		case sfStateSave:
			switch msg.String() {
			case "esc":
				m.state = sfStatePreview
				m.input.Blur()
				return m, nil
			case "enter":
				inputVal := m.input.Value()
				isCustom := m.selectedTpl.Name == "Custom File (AI)"

				if isCustom {
					if inputVal == "" {
						// Fallback if empty, though unlikely
						inputVal = filepath.Join(m.workspace, "custom_file")
					}
					m.savePath = inputVal
				} else {
					// Static template: Input is directory
					if inputVal == "" {
						inputVal = m.workspace
					}
					m.savePath = filepath.Join(inputVal, m.selectedTpl.Name)
				}

				m.state = sfStateGenerating
				m.input.Blur()
				return m, tea.Batch(m.spinner.Tick, m.saveFileCmd())
			}
			m.input, cmd = m.input.Update(msg)
			return m, cmd

		case sfStateGenerating:
			// Just handle spinner updates (handled below)
			return m, nil

		case sfStateSuccess:
			switch msg.String() {
			case "esc", "enter", "q":
				m.resetState()
				return m, nil
			}

		case sfStateHelp:
			if msg.String() == "esc" || msg.String() == "enter" || msg.String() == "q" || msg.String() == "?" {
				m.state = sfStateSelectTemplate
				return m, nil
			}
			var cmd tea.Cmd
			m.helpView, cmd = m.helpView.Update(msg)
			return m, cmd
		}

	case tea.MouseMsg:
		var cmd tea.Cmd
		switch m.state {
		case sfStateSelectTemplate:
			if msg.Type == tea.MouseWheelUp {
				m.templateList.CursorUp()
				return m, nil
			}
			if msg.Type == tea.MouseWheelDown {
				m.templateList.CursorDown()
				return m, nil
			}
			m.templateList, cmd = m.templateList.Update(msg)
			return m, cmd
		case sfStatePreview:
			if msg.Type == tea.MouseWheelUp {
				m.preview.LineUp(3)
				return m, nil
			}
			if msg.Type == tea.MouseWheelDown {
				m.preview.LineDown(3)
				return m, nil
			}
			m.preview, cmd = m.preview.Update(msg)
			return m, cmd
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.templateList.SetSize(msg.Width-4, msg.Height-10)
		m.preview.Width = msg.Width - 6
		m.preview.Height = msg.Height - 12
		m.helpView.Width = msg.Width
		m.helpView.Height = msg.Height

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case sfPreviewReadyMsg:
		m.generatePreview()
		m.state = sfStatePreview
		return m, nil

	case sfAIGeneratedMsg:
		m.result = string(msg)
		m.selectedTpl.Generator = func(_ map[string]string) string { return m.result }
		m.generatePreview()
		m.state = sfStatePreview
		return m, nil

	case sfSaveResult:
		m.err = msg.err
		m.state = sfStateSuccess
		return m, nil
	}

	return m, nil
}

func (m *SmartFileModel) resetState() {
	m.state = sfStateSelectTemplate
	m.options = make(map[string]string)
	m.languageInput.SetValue("")
	m.aiPromptInput.SetValue("")
	m.input.SetValue("")
	m.savePath = ""
	m.customPrompt = ""
	m.result = ""
	m.err = nil
	m.languageInput.Blur()
	m.aiPromptInput.Blur()
	m.input.Blur()
}

type sfSaveResult struct{ err error }
type sfAIGeneratedMsg string
type sfPreviewReadyMsg struct{}

func (m SmartFileModel) generateStandardPreviewCmd() tea.Cmd {
	return func() tea.Msg {
		// Small delay to simulate "building" for non-AI templates
		time.Sleep(600 * time.Millisecond)
		return sfPreviewReadyMsg{}
	}
}

func (m SmartFileModel) generateAIFileCmd() tea.Cmd {
	return func() tea.Msg {
		if m.provider == nil {
			return sfAIGeneratedMsg("AI Provider not configured")
		}

		prompt := fmt.Sprintf("Generate the content for a file named '%s'. The file should: %s. Output ONLY the file content, no explanations.",
			filepath.Base(m.savePath), m.customPrompt)

		messages := []ai.Message{
			{Role: "system", Content: "You are a specialized file generator. Output only code/content."},
			{Role: "user", Content: prompt},
		}

		resp, err := m.provider.Send(messages)
		if err != nil {
			return sfAIGeneratedMsg("Error: " + err.Error())
		}
		return sfAIGeneratedMsg(resp)
	}
}

func (m SmartFileModel) saveFileCmd() tea.Cmd {
	return func() tea.Msg {
		// Artificial delay to show the "generation" process
		time.Sleep(800 * time.Millisecond)

		// Ensure directory exists
		dir := filepath.Dir(m.savePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return sfSaveResult{err: err}
		}

		// Save file
		content := m.selectedTpl.Generator(m.options)
		if err := os.WriteFile(m.savePath, []byte(content), 0644); err != nil {
			return sfSaveResult{err: err}
		}
		return sfSaveResult{err: nil}
	}
}

func (m *SmartFileModel) generatePreview() {
	content := m.selectedTpl.Generator(m.options)

	// Syntax highlighting using Glamour
	lang := m.selectedTpl.Extension
	if lang != "" && lang[0] == '.' {
		lang = lang[1:]
	}
	// Fallback/Special cases
	if m.selectedTpl.Name == "Dockerfile" {
		lang = "dockerfile"
	} else if m.selectedTpl.Name == "Makefile" {
		lang = "makefile"
	} else if lang == "yml" || lang == "yaml" {
		lang = "yaml"
	}

	md := fmt.Sprintf("```%s\n%s\n```", lang, content)
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(m.preview.Width-4),
	)
	out, err := renderer.Render(md)
	if err != nil {
		m.preview.SetContent(content)
	} else {
		m.preview.SetContent(out)
	}
	m.preview.GotoTop()
}

func (m SmartFileModel) View() string {
	switch m.state {
	case sfStateSelectTemplate:
		// Enhanced Menu Layout
		header := lipgloss.NewStyle().
			Foreground(colorCyan).
			Bold(true).
			Render("Smart File Creator")

		subtext := lipgloss.NewStyle().Foreground(colorGray).Render("Select a template to generate")

		innerContent := lipgloss.JoinVertical(lipgloss.Left,
			header,
			subtext,
			"\n",
			docStyle.Render(m.templateList.View()),
			"\n",
			subtleStyle.Render("↑/↓: Navigate • Enter: Select • Esc: Back"),
		)
		// Use a container with padding for the left-aligned view
		// We avoid lipgloss.Place here to ensuring strictly top-left positioning without centering logic interfering
		return lipgloss.NewStyle().Padding(1, 2).Render(innerContent)

	case sfStateCustomize:
		step := StepStyle.Render("Configuration")
		// Use simple bold purple instead of titleStyle to avoid double border
		title := lipgloss.NewStyle().Foreground(colorPurple).Bold(true).Render(fmt.Sprintf("Customize: %s", m.selectedTpl.Name))
		prompt := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render("Select programming language or specialized options:")

		content := lipgloss.JoinVertical(lipgloss.Center,
			step,
			title,
			"\n",
			prompt,
			"\n",
			focusedInputBoxStyle.Render(m.languageInput.View()),
			"\n",
			subtleStyle.Render("Enter: Continue • Esc: Back"),
		)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, WizardCardStyle.Render(content))

	case sfStateFilename:
		step := StepStyle.Render("Step 1 of 3")
		title := lipgloss.NewStyle().Foreground(colorPurple).Bold(true).Render("Set Filename")
		prompt := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render("Enter name for the new file:")

		content := lipgloss.JoinVertical(lipgloss.Center,
			step,
			title,
			"\n",
			prompt,
			"\n",
			focusedInputBoxStyle.Render(m.input.View()),
			"\n",
			subtleStyle.Render("e.g. config.yml, auth.go • Enter: Next • Esc: Back"),
		)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, WizardCardStyle.Render(content))

	case sfStateAIPrompt:
		step := StepStyle.Render("Step 2 of 3")
		title := lipgloss.NewStyle().Foreground(colorPurple).Bold(true).Render("Describe Content")
		prompt := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render("Instructions for the AI Generator:")

		content := lipgloss.JoinVertical(lipgloss.Center,
			step,
			title,
			"\n",
			prompt,
			"\n",
			focusedInputBoxStyle.Render(m.aiPromptInput.View()),
			"\n",
			subtleStyle.Render("e.g. 'A text parser in Python' • Enter: Generate"),
		)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, WizardCardStyle.Render(content))

	case sfStatePreview:
		// Premium Preview Window
		headerText := fmt.Sprintf(" PREVIEW: %s ", filepath.Base(m.savePath))
		if m.selectedTpl.Name != "Custom File (AI)" && m.savePath == "" {
			headerText = fmt.Sprintf(" PREVIEW: %s ", m.selectedTpl.Name)
		}

		header := PreviewHeaderStyle.Render(headerText)

		// Ensure viewport fits nicely with header
		vp := m.preview.View()

		content := lipgloss.JoinVertical(lipgloss.Left,
			header,
			vp,
			"\n",
			subtleStyle.Render("Enter: Save & Write • Esc: Edit • ↑/↓: Scroll Code"),
		)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)

	case sfStateSave:
		step := StepStyle.Render("Final Step")

		var titleStr, promptStr string
		if m.selectedTpl.Name == "Custom File (AI)" {
			titleStr = "Confirm Save Path"
			promptStr = "Verify absolute path before writing:"
		} else {
			titleStr = "Confirm Destination"
			promptStr = fmt.Sprintf("Verify destination folder (filename will be %s):", m.selectedTpl.Name)
		}

		title := lipgloss.NewStyle().Foreground(colorPurple).Bold(true).Render(titleStr)
		prompt := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render(promptStr)

		content := lipgloss.JoinVertical(lipgloss.Center,
			step,
			title,
			"\n",
			prompt,
			"\n",
			focusedInputBoxStyle.Render(m.input.View()),
			"\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(fmt.Sprintf("In Workspace: %s", m.workspace)),
			"\n",
			subtleStyle.Render("Enter: Write File • Esc: Cancel"),
		)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, WizardCardStyle.Render(content))

	case sfStateGenerating:
		status := "Building file..."
		if m.selectedTpl != nil && m.selectedTpl.Name == "Custom File (AI)" {
			status = "🤖 AI is generating content..."
		}
		if m.state == sfStateSave {
			status = "💾 Writing file to disk..."
		}

		msg := lipgloss.JoinVertical(lipgloss.Center,
			m.spinner.View(),
			"\n",
			lipgloss.NewStyle().Foreground(colorPurple).Bold(true).Render(status),
		)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, WizardCardStyle.Render(msg))

	case sfStateSuccess:
		if m.err != nil {
			content := lipgloss.JoinVertical(lipgloss.Center,
				lipgloss.NewStyle().Foreground(colorRed).Bold(true).Render("❌ Error Saving File"),
				"\n",
				m.err.Error(),
				"\n",
				subtleStyle.Render("Press Esc to go back"),
			)
			return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, WizardCardStyle.Render(content))
		}

		content := lipgloss.JoinVertical(lipgloss.Center,
			lipgloss.NewStyle().Foreground(colorGreen).Bold(true).Render("✅ File Created Successfully!"),
			"\n",
			lipgloss.NewStyle().Foreground(colorGray).Render(m.savePath),
			"\n",
			subtleStyle.Render("Press Key to Continue"),
		)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, WizardCardStyle.Render(content))
	case sfStateHelp:
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.helpView.View())
	}

	return ""
}
