package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/phravins/devcli/internal/venv"
)

type VenvDashboardModel struct {
	list    list.Model
	spinner spinner.Model
	manager *venv.Manager

	// State
	state         int
	width, height int

	message     string
	err         error // Global error state
	input       textinput.Model
	selectedEnv venv.Environment

	// Logging
	logView    viewport.Model
	logBuf     *strings.Builder
	targetPath string
	countdown  int            // Countdown timer for success screen
	helpView   viewport.Model // New
}

const (
	StateVenvList          = iota
	StateVenvActionMenu    // Activate, Sync, Delete, Clone
	StateVenvCloneInput    // Enter destination path
	StateVenvCreateInput   // Enter new venv path
	StateVenvSyncInput     // Enter path for requirements.txt (NEW)
	StateVenvScanInput     // Enter path to scan
	StateVenvDeleteConfirm // Confirm deletion
	StateVenvProcessing
	StateVenvCreating // Active logging state
	StateVenvSuccess  // Final success screen
	StateVenvHelp     // Educational screen
)

func NewVenvDashboardModel() VenvDashboardModel {
	mgr := venv.NewManager("")

	// Initial List - Delegate handles styling
	items := loadVenvs(mgr)
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = venvSelectedStyle
	delegate.Styles.SelectedDesc = venvSelectedStyle.Copy().Foreground(colorGray)

	l := list.New(items, delegate, 0, 0)
	l.Title = "Virtual Environment Wizard (v2.0 Recursive)"
	l.SetShowTitle(false) // We render our own fancy title

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	ti := textinput.New()
	ti.Placeholder = "New Project Path (e.g. C:\\MyProject)"
	ti.Width = 50

	vp := viewport.New(60, 10)
	vp.Style = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)

	// Help Viewport
	hv := viewport.New(0, 0)
	hv.Style = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#0F9E99")).Padding(1, 2)
	hv.SetContent(VenvWizardHelp)

	// Check requirements immediately
	var initErr error
	if err := mgr.CheckPrerequisites(); err != nil {
		initErr = err
	}

	return VenvDashboardModel{
		list:     l,
		spinner:  s,
		manager:  mgr,
		input:    ti,
		state:    StateVenvList,
		err:      initErr,
		logView:  vp,
		logBuf:   &strings.Builder{},
		helpView: hv,
	}
}

func loadVenvs(mgr *venv.Manager) []list.Item {
	envs, err := mgr.List()
	if err != nil {
		return []list.Item{item{title: "Error", desc: err.Error()}}
	}

	var items []list.Item
	for _, e := range envs {
		icon := ""
		switch e.Type {
		case venv.TypeNodeModules:
			icon = ""
		case venv.TypeAnaconda:
			icon = ""
		}

		// Check if this is a cloned environment
		envName := e.Name
		parentDir := filepath.Base(filepath.Dir(e.Path))
		// If parent directory contains "copy" or "clone", mark it as cloned
		if strings.Contains(strings.ToLower(parentDir), "copy") ||
			strings.Contains(strings.ToLower(parentDir), "clone") {
			envName = e.Name + " (cloned)"
		}

		title := fmt.Sprintf("%s %s", icon, envName)
		desc := fmt.Sprintf("%s | %s | %s", e.Type, e.Size, e.Path)
		items = append(items, item{title: title, desc: desc})
	}

	if len(items) == 0 {
		items = append(items, item{title: "No environments found", desc: "Press 'n' to create one properly!"})
	}

	return items
}

func (m VenvDashboardModel) Init() tea.Cmd {
	return m.spinner.Tick
}

type venvMsg struct {
	err error
	msg string
}

// Feature-specific Back Messages are now defined in root.go

func (m VenvDashboardModel) Update(msg tea.Msg) (VenvDashboardModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.state == StateVenvProcessing {
			return m, nil
		}

		if m.state == StateVenvSuccess {
			// Require explicit Enter/Esc to dismiss, not just any key
			if msg.String() == "enter" || msg.String() == "esc" {
				m.state = StateVenvList
				// Auto-Switch to the new environment's parent directory
				// m.targetPath is the full path to the environment.
				// If we created C:\Proj\venv, we want to look at C:\Proj
				if m.targetPath != "" {
					// For venv creation, targetPath IS the venv folder.
					// Workspace should be the parent to see it in the list (usually).
					// OR if targetPath is C:\Proj and we made C:\Proj\venv (standard CreateVenv behavior creates dirs if needed).
					// Wait, CreateVenv(target) makes target. So target is the venv.
					m.manager.Workspace = filepath.Dir(m.targetPath)
				}
				m.list.SetItems(loadVenvs(m.manager))
			}
			return m, nil
		}

		if m.state == StateVenvHelp {
			switch msg.String() {
			case "esc", "enter", "?":
				m.state = StateVenvList
				return m, nil
			default:
				var cmd tea.Cmd
				m.helpView, cmd = m.helpView.Update(msg)
				return m, cmd
			}
		}

		// Error Dismissal
		if m.err != nil {
			if msg.String() == "esc" || msg.String() == "enter" {
				m.err = nil
			}
			return m, nil
		}

		if m.state == StateVenvList {
			switch msg.String() {
			case "q", "esc":
				// SIGNAL PARENT TO GO BACK
				return m, func() tea.Msg { return VenvBackMsg{} }
			case "?":
				m.state = StateVenvHelp
				m.message = "" // Clear message
				m.helpView.GotoTop()
				return m, nil
			case "n":
				m.state = StateVenvCreateInput
				m.input.Placeholder = "New Environment Path (e.g. ./my-venv)"
				m.input.SetValue("") // Clear previous
				m.input.Focus()
				m.message = "" // Clear message
				return m, nil
			case "s": // Scan (was 'o')
				m.state = StateVenvScanInput
				m.input.Placeholder = "Scan Folder Path"
				m.input.SetValue("")
				m.input.Focus()
				m.message = "" // Clear message
				return m, nil
			case "o": // Keep 'o' as alias or remove? Removing for strictness.
				// m.state = StateVenvScanInput ...
				// User asked for 's', let's stick to 's'.
				return m, nil
			case "r":
				m.list.SetItems(loadVenvs(m.manager))
				m.message = "" // Clear message on refresh
				return m, nil
			case "enter":
				i, ok := m.list.SelectedItem().(item)
				if ok && i.title != "No environments found" && i.title != "Error" {
					m.state = StateVenvActionMenu
					m.message = "" // Clear message when entering action menu
					parts := strings.Split(i.desc, " | ")
					if len(parts) >= 3 {
						m.selectedEnv = venv.Environment{Name: i.title, Path: parts[2]}
					}
					return m, nil // CRITICAL: Return here to prevent list update
				}
			}
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

		if m.state == StateVenvActionMenu {
			switch msg.String() {
			case "esc":
				m.state = StateVenvList
			case "y": // Sync (was 's')
				m.state = StateVenvSyncInput
				m.input.Placeholder = "Path for requirements.txt"
				// Default to parent/requirements.txt
				defaultPath := filepath.Join(filepath.Dir(m.selectedEnv.Path), "requirements.txt")
				m.input.SetValue(defaultPath)
				m.input.Focus()
				return m, nil
			case "d": // Delete - Show confirmation first
				m.state = StateVenvDeleteConfirm
				return m, nil
			case "c": // Clone
				m.state = StateVenvCloneInput
				m.input.Placeholder = "Destination directory (e.g., D:\\MyNewProject)"
				// Start empty so user can type any path they want
				m.input.SetValue("")
				m.input.Focus()
				return m, nil
			}
		}

		if m.state == StateVenvCloneInput {
			switch msg.String() {
			case "esc":
				m.state = StateVenvActionMenu
			case "enter":
				target := m.input.Value()
				if target != "" {
					if abs, err := filepath.Abs(target); err == nil {
						target = abs
					}
					// Append .venv to destination (consistent with create)
					venvPath := filepath.Join(target, ".venv")
					m.targetPath = venvPath // Store for navigation
					m.state = StateVenvProcessing
					m.message = "Cloning environment..."

					// Store source name for success message
					sourceName := filepath.Base(filepath.Dir(m.selectedEnv.Path))

					return m, func() tea.Msg {
						err := m.manager.Clone(m.selectedEnv.Path, venvPath)
						if err != nil {
							return venvMsg{err: err, msg: ""}
						}
						return venvMsg{err: nil, msg: fmt.Sprintf("Successfully cloned '%s' to %s (cloned)", sourceName, venvPath)}
					}
				}
			}
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}

		if m.state == StateVenvDeleteConfirm {
			switch msg.String() {
			case "esc", "n":
				m.state = StateVenvActionMenu
				return m, nil
			case "enter", "y":
				// Confirmed - proceed with deletion
				m.state = StateVenvProcessing
				m.message = "Deleting environment..."
				return m, func() tea.Msg {
					err := m.manager.Delete(m.selectedEnv.Path)
					if err != nil {
						return venvMsg{err: err, msg: ""}
					}
					return venvMsg{err: nil, msg: fmt.Sprintf("Successfully deleted %s", m.selectedEnv.Name)}
				}
			}
			return m, nil
		}

		if m.state == StateVenvCreateInput {
			switch msg.String() {
			case "esc":
				m.state = StateVenvList
			case "enter":
				target := m.input.Value()
				if target != "" {
					// Resolve absolute path for clarity
					if abs, err := filepath.Abs(target); err == nil {
						target = abs
					}

					// Append .venv to create venv inside a .venv subfolder
					venvPath := filepath.Join(target, ".venv")

					m.targetPath = venvPath
					m.state = StateVenvCreating
					m.logBuf.Reset()
					m.logBuf.WriteString("Initializing Virtual Environment Wizard...\n")
					m.logBuf.WriteString(fmt.Sprintf("Project Directory: %s\n", target))
					m.logBuf.WriteString(fmt.Sprintf("Venv Location: %s\n", venvPath))
					m.logView.SetContent(m.logBuf.String())
					return m, checkPythonCmd(m.manager)
				}
			}
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}

		// Allow exiting the log view (Creating) manually ONLY with Esc
		if m.state == StateVenvCreating {
			if msg.String() == "esc" {
				m.state = StateVenvList
				// Refresh list if we are leaving
				m.list.SetItems(loadVenvs(m.manager))
				return m, nil
			}
		}

		if m.state == StateVenvScanInput {
			switch msg.String() {
			case "esc":
				m.state = StateVenvList
			case "enter":
				target := m.input.Value()
				if target != "" {
					m.manager.Workspace = target // Update workspace
					m.state = StateVenvList
					// m.list.Title = "Scanning: " + target
					m.list.SetItems(loadVenvs(m.manager))
					return m, nil
				}
			}
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}

		if m.state == StateVenvSyncInput {
			switch msg.String() {
			case "esc":
				m.state = StateVenvActionMenu
			case "enter":
				target := m.input.Value()
				if target != "" {
					m.state = StateVenvProcessing
					m.message = "Generating requirements.txt..."
					return m, func() tea.Msg {
						err := m.manager.Sync(m.selectedEnv.Path, target)
						if err != nil {
							return venvMsg{err: err, msg: ""}
						}
						return venvMsg{err: nil, msg: fmt.Sprintf("Successfully saved requirements.txt to %s", target)}
					}
				}
			}
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}

	case venvMsg:
		m.state = StateVenvList
		if msg.err != nil {
			m.err = msg.err // Set global error
			m.message = ""  // Clear any previous success message
		} else {
			// Show success message to user
			if msg.msg != "" {
				m.message = msg.msg
			}
			// Auto-Navigation for Clone/Sync success
			if m.targetPath != "" {
				// For Clone: targetPath is the new venv. Switch to parent to see it.
				// For Sync: targetPath is the requirements.txt file. Switch to its folder.
				if strings.HasSuffix(m.targetPath, "requirements.txt") {
					m.manager.Workspace = filepath.Dir(m.targetPath)
				} else {
					m.manager.Workspace = filepath.Dir(m.targetPath)
				}
			}
		}
		m.list.SetItems(loadVenvs(m.manager))

	case pythonFoundMsg:
		m.logBuf.WriteString(" Python found in system PATH.\n")
		m.logBuf.WriteString(fmt.Sprintf("User specified path: %s\n", m.targetPath))
		m.logBuf.WriteString(fmt.Sprintf(" Creating venv at %s...\n", m.targetPath))
		m.logBuf.WriteString("Installing virtual environment packages...\n")
		m.logView.SetContent(m.logBuf.String())
		m.logView.GotoBottom()
		return m, createVenvStepCmd(m.manager, m.targetPath)

	case venvCreatedMsg:
		if msg.err != nil {
			// Do NOT jump back to list. Show error in log.
			m.logBuf.WriteString(fmt.Sprintf(" Creation Failed: %v\n", msg.err))
			m.logBuf.WriteString("\n(Press Esc to return)")
			m.logView.SetContent(m.logBuf.String())
			m.logView.GotoBottom()
			return m, nil
		}
		m.logBuf.WriteString(" Virtual environment files generated successfully.\n")
		m.logBuf.WriteString(" Packages installed.\n")
		m.logBuf.WriteString("Verifying environment integrity...\n")
		m.logView.SetContent(m.logBuf.String())
		m.logView.GotoBottom()
		return m, verifyVenvCmd(m.manager, m.targetPath)

	case venvVerifiedMsg:
		if msg.err != nil {
			m.logBuf.WriteString(fmt.Sprintf(" Verification Failed: %v\n", msg.err))
			m.logBuf.WriteString("\n(Press Esc to return)")
			m.logView.SetContent(m.logBuf.String())
			m.logView.GotoBottom()
			// Stop here. No auto-tick.
			return m, nil
		} else {
			m.logBuf.WriteString(" Environment verified! Ready to use.\n")
			m.logBuf.WriteString(fmt.Sprintf(" Successfully created at: %s\n", m.targetPath))
			m.logBuf.WriteString("Success screen in: 4...\n")
			m.logView.SetContent(m.logBuf.String())
			m.logView.GotoBottom()
			// Start countdown from 4
			m.countdown = 4
			return m, tea.Tick(1*time.Second, func(_ time.Time) tea.Msg { return venvCountdownMsg{count: 3} })
		}

	case venvCountdownMsg:
		if msg.count > 0 {
			// Update the countdown display
			m.countdown = msg.count
			// Update last line with new countdown
			lines := strings.Split(m.logBuf.String(), "\n")
			if len(lines) > 0 {
				lines[len(lines)-2] = fmt.Sprintf("Success screen in: %d...\n", msg.count)
			}
			m.logBuf.Reset()
			m.logBuf.WriteString(strings.Join(lines, "\n"))
			m.logView.SetContent(m.logBuf.String())
			m.logView.GotoBottom()
			// Continue countdown
			return m, tea.Tick(1*time.Second, func(_ time.Time) tea.Msg { return venvCountdownMsg{count: msg.count - 1} })
		} else {
			// Countdown finished, show success screen
			return m, func() tea.Msg { return venvSuccessMsg{} }
		}

	case venvSuccessMsg:
		m.state = StateVenvSuccess
		return m, nil

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		// Adjust list height to account for Header (3 lines) + Spacing (1) + Help (2) = ~6 lines
		m.list.SetSize(msg.Width-h, msg.Height-v-6)
		m.width = msg.Width
		m.height = msg.Height
		m.logView.Width = msg.Width - 4
		m.logView.Height = msg.Height - 10

		// Resize Help View
		m.helpView.Width = msg.Width - h - 4
		m.helpView.Height = msg.Height - v - 4

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.MouseMsg:
		if m.state == StateVenvHelp {
			m.helpView, cmd = m.helpView.Update(msg)
			return m, cmd
		}
		if m.state == StateVenvCreating {
			m.logView, cmd = m.logView.Update(msg)
			return m, cmd
		}
		// List handles mouse events if delegate supports it (bubbles/list/defaultitem doesn't always, but list.Model does for scrolling)
		if m.state == StateVenvList {
			switch msg.Type {
			case tea.MouseWheelUp:
				m.list.CursorUp()
			case tea.MouseWheelDown:
				m.list.CursorDown()
			}
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m VenvDashboardModel) View() string {
	h, v := docStyle.GetFrameSize()

	if m.err != nil {
		// Error Overlay
		content := fmt.Sprintf("Error Detected\n\n%v\n\n(Press Esc/Enter to Dismiss)", m.err)
		return docStyle.Render(
			lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
				errorBoxStyle.Render(content),
			),
		)
	}

	if m.state == StateVenvHelp {
		return docStyle.Render(
			lipgloss.Place(m.width-h, m.height-v, lipgloss.Center, lipgloss.Center, m.helpView.View()),
		)
	}

	if m.state == StateVenvCreating {
		header := venvTitleStyle.Render(fmt.Sprintf("%s Creating Environment...", m.spinner.View()))
		return docStyle.Render(lipgloss.JoinVertical(lipgloss.Left, header, m.logView.View()))
	}

	if m.state == StateVenvProcessing {
		// Simple centered spinner
		content := fmt.Sprintf("%s %s", m.spinner.View(), m.message)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}

	if m.state == StateVenvSuccess {
		title := lipgloss.NewStyle().Foreground(colorGreen).Bold(true).Render(" SUCCESS ")
		msg := fmt.Sprintf("Virtual Environment Created at:\n%s\n\n(Press Enter to Continue)", m.targetPath)

		content := lipgloss.JoinVertical(lipgloss.Center, title, "\n", msg)
		return docStyle.Render(
			lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
				successBoxStyle.Render(content),
			),
		)
	}

	// Centered Inputs
	if m.state == StateVenvCloneInput || m.state == StateVenvCreateInput || m.state == StateVenvScanInput || m.state == StateVenvSyncInput {
		var title, inputView, footer string

		switch m.state {
		case StateVenvCloneInput:
			title = "Clone Environment"
			inputView = m.input.View()
			footer = "(Enter to Clone, Esc to Back)"
		case StateVenvCreateInput:
			title = "Create New Python Venv"
			inputView = m.input.View()
			footer = "(Enter to Create, Esc to Back)"
		case StateVenvScanInput:
			title = "Scan Directory"
			inputView = m.input.View()
			footer = "(Enter to Scan, Esc to Back)"
		case StateVenvSyncInput:
			title = "Sync Packages"
			inputView = m.input.View()
			footer = "(Enter path for requirements.txt, Esc to Back)"
		}

		content := lipgloss.JoinVertical(lipgloss.Center,
			venvTitleStyle.Render(title),
			"\n",
			focusedInputBoxStyle.Render(inputView),
			"\n",
			subtleStyle.Render(footer),
		)

		return docStyle.Render(
			lipgloss.Place(m.width-h, m.height-v, lipgloss.Center, lipgloss.Center, content),
		)
	}

	if m.state == StateVenvDeleteConfirm {
		// Delete Confirmation Dialog
		title := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true).Render("Confirm Deletion")
		envName := venvSelectedStyle.Render(m.selectedEnv.Name)
		envPath := subtleStyle.Render(m.selectedEnv.Path)

		warning := lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Render(
			"This action cannot be undone!",
		)

		content := lipgloss.JoinVertical(lipgloss.Center,
			title,
			"\n",
			"You are about to delete:",
			envName,
			envPath,
			"\n",
			warning,
			"\n",
			subtleStyle.Render("[Y/Enter] Confirm Delete • [N/Esc] Cancel"),
		)

		return docStyle.Render(
			lipgloss.Place(m.width-h, m.height-v, lipgloss.Center, lipgloss.Center,
				errorBoxStyle.Render(content),
			),
		)
	}

	if m.state == StateVenvActionMenu {
		// Action Menu as a centered card
		title := venvTitleStyle.Render("Manage Environment")
		env := venvSelectedStyle.Render(m.selectedEnv.Name)

		menu := lipgloss.JoinVertical(lipgloss.Left,
			"",
			"[y] Sync Packages",
			"    Generate requirements.txt",
			"",
			"[c] Clone Environment",
			"    Duplicate to another project",
			"",
			"[d] Delete Environment",
			"    Remove from disk",
			"",
			"[Esc] Back",
		)

		content := lipgloss.JoinVertical(lipgloss.Center,
			title,
			env,
			"\n",
			venvCardStyle.Render(menu),
		)
		return docStyle.Render(
			lipgloss.Place(m.width-h, m.height-v, lipgloss.Center, lipgloss.Center, content),
		)
	}

	// Main Dashboard View
	header := lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(
		titleStyle.Render("Virtual Environment Wizard"),
	)

	// Show success message if present
	var successMsg string
	if m.message != "" {
		successMsg = lipgloss.NewStyle().
			Foreground(colorGreen).
			Bold(true).
			Render(" " + m.message)
	}

	help := subtleStyle.Render("\n [?] Help • [n] New Env • [s] Scan System • [r] Refresh • [q] Quit")

	// Build view with optional success message
	var content string
	if successMsg != "" {
		content = lipgloss.JoinVertical(lipgloss.Left,
			header,
			"\n",
			successMsg+"\n",
			m.list.View(),
			help,
		)
	} else {
		content = lipgloss.JoinVertical(lipgloss.Left,
			header,
			"\n",
			m.list.View(),
			help,
		)
	}

	return docStyle.Render(content)
}

// Messages & Commands logic
type pythonFoundMsg struct{}
type venvCreatedMsg struct{ err error }
type venvVerifiedMsg struct{ err error }
type venvCountdownMsg struct{ count int }
type venvSuccessMsg struct{}

func checkPythonCmd(mgr *venv.Manager) tea.Cmd {
	return func() tea.Msg {
		if err := mgr.CheckPrerequisites(); err != nil {
			return venvCreatedMsg{err: err} // Reuse err msg
		}
		// Artificial delay for UX
		time.Sleep(500 * time.Millisecond)
		return pythonFoundMsg{}
	}
}

func createVenvStepCmd(mgr *venv.Manager, path string) tea.Cmd {
	return func() tea.Msg {
		err := mgr.CreateVenv(path)
		return venvCreatedMsg{err: err}
	}
}

func verifyVenvCmd(mgr *venv.Manager, path string) tea.Cmd {
	return func() tea.Msg {
		// Artificial delay
		time.Sleep(500 * time.Millisecond)
		err := mgr.Verify(path)
		return venvVerifiedMsg{err: err}
	}
}
