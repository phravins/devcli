package tui

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/phravins/devcli/internal/devserver"
)

func TestNewDevServerDashboardModel(t *testing.T) {
	testPath := "/tmp/test-project"
	model := NewDevServerDashboardModel(testPath)

	if model.projectPath != testPath {
		t.Errorf("expected projectPath %q, got %q", testPath, model.projectPath)
	}

	if model.state != StateDevServerPathInput {
		t.Errorf("expected state %d (StateDevServerPathInput), got %d", StateDevServerPathInput, model.state)
	}

	if model.filterMode != "all" {
		t.Errorf("expected filterMode %q, got %q", "all", model.filterMode)
	}

	if model.serverFilter != "all" {
		t.Errorf("expected serverFilter %q, got %q", "all", model.serverFilter)
	}

	if !model.autoScroll {
		t.Errorf("expected autoScroll to be true")
	}

	if model.showHelp {
		t.Errorf("expected showHelp to be false")
	}

	if model.pathInput.Value() != testPath {
		t.Errorf("expected pathInput value %q, got %q", testPath, model.pathInput.Value())
	}

	if len(model.logs) != 0 {
		t.Errorf("expected initial logs slice to be empty, got length %d", len(model.logs))
	}
}

func TestDevServerDashboardModel_Init(t *testing.T) {
	model := NewDevServerDashboardModel("/tmp/test-project")
	cmd := model.Init()
	if cmd == nil {
		t.Errorf("expected Init() to return a non-nil tea.Cmd")
	}
}

func TestDevServerDashboardModel_Update_PathInput(t *testing.T) {
	t.Run("Enter with path value", func(t *testing.T) {
		model := NewDevServerDashboardModel("/tmp/test-project")
		updatedModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

		m := updatedModel.(DevServerDashboardModel)
		if m.state != StateDevServerDetecting {
			t.Errorf("expected state StateDevServerDetecting (%d), got %d", StateDevServerDetecting, m.state)
		}
		if cmd == nil {
			t.Errorf("expected non-nil tea.Cmd on path input enter")
		}
	})

	t.Run("Esc key returns back message", func(t *testing.T) {
		model := NewDevServerDashboardModel("/tmp/test-project")
		_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})

		if cmd == nil {
			t.Fatalf("expected non-nil tea.Cmd for esc key")
		}

		msg := cmd()
		if _, ok := msg.(DevServerBackMsg); !ok {
			t.Errorf("expected DevServerBackMsg, got %T", msg)
		}
	})

	t.Run("Ctrl+C or q quits", func(t *testing.T) {
		model := NewDevServerDashboardModel("/tmp/test-project")
		_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

		if cmd == nil {
			t.Fatalf("expected non-nil tea.Cmd for q key")
		}

		msg := cmd()
		if msg != tea.Quit() {
			t.Errorf("expected tea.Quit msg, got %v", msg)
		}
	})
}

func TestDevServerDashboardModel_Update_DetectDone(t *testing.T) {
	model := NewDevServerDashboardModel("/tmp/test-project")
	model.state = StateDevServerDetecting

	t.Run("Detection success", func(t *testing.T) {
		info := devserver.ProjectInfo{Type: devserver.TypeGo}
		updatedModel, _ := model.Update(detectDoneMsg{info: info, err: nil})

		m := updatedModel.(DevServerDashboardModel)
		if m.state != StateDevServerReady {
			t.Errorf("expected state StateDevServerReady (%d), got %d", StateDevServerReady, m.state)
		}
		if m.projectInfo.Type != devserver.TypeGo {
			t.Errorf("expected project type Go, got %s", m.projectInfo.Type)
		}
		if m.err != nil {
			t.Errorf("expected nil error, got %v", m.err)
		}
	})

	t.Run("Detection error", func(t *testing.T) {
		detectErr := fmt.Errorf("unable to detect project type")
		updatedModel, _ := model.Update(detectDoneMsg{info: devserver.ProjectInfo{}, err: detectErr})

		m := updatedModel.(DevServerDashboardModel)
		if m.err != detectErr {
			t.Errorf("expected error %v, got %v", detectErr, m.err)
		}
	})
}

func TestDevServerDashboardModel_Update_LogReceived(t *testing.T) {
	model := NewDevServerDashboardModel("/tmp/test-project")

	logMsg := logReceivedMsg{
		log: devserver.LogLine{
			ServerName: "backend",
			Line:       "WARNING: High memory usage",
			IsError:    false,
		},
	}

	updatedModel, _ := model.Update(logMsg)
	m := updatedModel.(DevServerDashboardModel)

	if len(m.logs) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(m.logs))
	}

	entry := m.logs[0]
	if entry.serverName != "backend" {
		t.Errorf("expected serverName backend, got %s", entry.serverName)
	}
	if entry.line != "WARNING: High memory usage" {
		t.Errorf("expected line matching, got %s", entry.line)
	}
	if !entry.isWarning {
		t.Errorf("expected isWarning to be true due to line containing 'warn'")
	}
}

func TestDevServerDashboardModel_ExecutePendingAction(t *testing.T) {
	tests := []struct {
		name           string
		action         string
		initialState   DevServerDashboardModel
		expectedState  int
		checkModelFunc func(t *testing.T, m DevServerDashboardModel)
	}{
		{
			name:   "filter action cycles filterMode",
			action: "filter",
			initialState: DevServerDashboardModel{
				pendingAction: "filter",
				filterMode:    "all",
			},
			expectedState: StateDevServerRunning,
			checkModelFunc: func(t *testing.T, m DevServerDashboardModel) {
				if m.filterMode != "errors" {
					t.Errorf("expected filterMode 'errors', got %q", m.filterMode)
				}
			},
		},
		{
			name:   "source action cycles serverFilter for fullstack",
			action: "source",
			initialState: DevServerDashboardModel{
				pendingAction: "source",
				serverFilter:  "all",
				projectInfo:   devserver.ProjectInfo{Type: devserver.TypeFullstack},
			},
			expectedState: StateDevServerRunning,
			checkModelFunc: func(t *testing.T, m DevServerDashboardModel) {
				if m.serverFilter != "backend" {
					t.Errorf("expected serverFilter 'backend', got %q", m.serverFilter)
				}
			},
		},
		{
			name:   "clear action empties logs",
			action: "clear",
			initialState: DevServerDashboardModel{
				pendingAction: "clear",
				logs:          []logEntry{{line: "log 1"}},
			},
			expectedState: StateDevServerRunning,
			checkModelFunc: func(t *testing.T, m DevServerDashboardModel) {
				if len(m.logs) != 0 {
					t.Errorf("expected empty logs, got %d entries", len(m.logs))
				}
			},
		},
		{
			name:   "autoscroll action toggles autoScroll",
			action: "autoscroll",
			initialState: DevServerDashboardModel{
				pendingAction: "autoscroll",
				autoScroll:    true,
			},
			expectedState: StateDevServerRunning,
			checkModelFunc: func(t *testing.T, m DevServerDashboardModel) {
				if m.autoScroll != false {
					t.Errorf("expected autoScroll to be false")
				}
			},
		},
		{
			name:   "help action sets showHelp to true",
			action: "help",
			initialState: DevServerDashboardModel{
				pendingAction: "help",
				showHelp:      false,
			},
			expectedState: StateDevServerRunning,
			checkModelFunc: func(t *testing.T, m DevServerDashboardModel) {
				if !m.showHelp {
					t.Errorf("expected showHelp to be true")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, _ := tt.initialState.executePendingAction()
			if m.state != tt.expectedState {
				t.Errorf("expected state %d, got %d", tt.expectedState, m.state)
			}
			if m.pendingAction != "" {
				t.Errorf("expected pendingAction to be cleared, got %q", m.pendingAction)
			}
			if tt.checkModelFunc != nil {
				tt.checkModelFunc(t, m)
			}
		})
	}
}

func TestDevServerDashboardModel_View(t *testing.T) {
	model := NewDevServerDashboardModel("/tmp/test-project")
	model.width = 100
	model.height = 30

	states := []struct {
		name      string
		configure func(m *DevServerDashboardModel)
		contains  string
	}{
		{
			name:      "Path Input State",
			configure: func(m *DevServerDashboardModel) { m.state = StateDevServerPathInput },
			contains:  "Auto-Detect Framework",
		},
		{
			name:      "Detecting State",
			configure: func(m *DevServerDashboardModel) { m.state = StateDevServerDetecting },
			contains:  "Scanning project folder...",
		},
		{
			name: "Ready State",
			configure: func(m *DevServerDashboardModel) {
				m.state = StateDevServerReady
				m.projectInfo = devserver.ProjectInfo{Type: devserver.TypeGo}
			},
			contains: "Detected:",
		},
		{
			name: "Running State",
			configure: func(m *DevServerDashboardModel) {
				m.state = StateDevServerRunning
				m.projectInfo = devserver.ProjectInfo{Type: devserver.TypeGo}
			},
			contains: "Dev Server - Go",
		},
		{
			name: "Confirmation State",
			configure: func(m *DevServerDashboardModel) {
				m.state = StateDevServerConfirmation
				m.confirmationMessage = "Are you sure?"
			},
			contains: "Confirmation Required",
		},
		{
			name: "Help Overlay",
			configure: func(m *DevServerDashboardModel) {
				m.showHelp = true
			},
			contains: "Dev Server",
		},
	}

	for _, st := range states {
		t.Run(st.name, func(t *testing.T) {
			m := model
			st.configure(&m)
			viewOutput := m.View()
			if viewOutput == "" {
				t.Errorf("View() output should not be empty")
			}
		})
	}
}
