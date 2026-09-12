package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewDevServerDashboardModelWithCustomPath(t *testing.T) {
	tempDir := t.TempDir()
	m := NewDevServerDashboardModel(tempDir)

	if m.projectPath != tempDir {
		t.Errorf("expected projectPath to be %s, got %s", tempDir, m.projectPath)
	}

	if m.state != StateDevServerPathInput {
		t.Errorf("expected state to be StateDevServerPathInput (%d), got %d", StateDevServerPathInput, m.state)
	}

	if m.pathInput.Value() != tempDir {
		t.Errorf("expected pathInput value to be %s, got %s", tempDir, m.pathInput.Value())
	}
}

func TestDevServerDashboardModelPathInputEnter(t *testing.T) {
	tempDir := t.TempDir()
	m := NewDevServerDashboardModel("")

	m.pathInput.SetValue(tempDir)

	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd := m.Update(enterMsg)

	devModel, ok := newModel.(DevServerDashboardModel)
	if !ok {
		t.Fatalf("expected model to be DevServerDashboardModel")
	}

	if devModel.projectPath != tempDir {
		t.Errorf("expected projectPath to be updated to %s, got %s", tempDir, devModel.projectPath)
	}

	if devModel.state != StateDevServerDetecting {
		t.Errorf("expected state to be StateDevServerDetecting (%d), got %d", StateDevServerDetecting, devModel.state)
	}

	if cmd == nil {
		t.Errorf("expected non-nil tea.Cmd for detection task")
	}
}

func TestRunDevServer(t *testing.T) {
	tempDir := t.TempDir()

	// Create dummy go.mod in tempDir
	goModPath := filepath.Join(tempDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte("module test"), 0644); err != nil {
		t.Fatalf("failed to create dummy go.mod: %v", err)
	}

	// \x03 is Ctrl+C which triggers tea.Quit in DevServerDashboardModel
	err := RunDevServer(tempDir, tea.WithoutRenderer(), tea.WithInput(strings.NewReader("\x03")))
	if err != nil {
		t.Errorf("expected RunDevServer with custom path to return nil error, got: %v", err)
	}

	errEmpty := RunDevServer("", tea.WithoutRenderer(), tea.WithInput(strings.NewReader("\x03")))
	if errEmpty != nil {
		t.Errorf("expected RunDevServer with empty path to return nil error, got: %v", errEmpty)
	}
}
