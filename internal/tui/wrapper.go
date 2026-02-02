package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// StandaloneWrapper wraps a model to handle BackMsg/Quit
// This allows models designed for nested use (returning BackMsg) to work standalone (Quitting on BackMsg)
type StandaloneWrapper struct {
	model tea.Model
}

func Wrap(m tea.Model) StandaloneWrapper {
	return StandaloneWrapper{model: m}
}

func (m StandaloneWrapper) Init() tea.Cmd {
	return m.model.Init()
}

func (m StandaloneWrapper) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Intercept BackMsg variants and Quit
	switch msg.(type) {
	case BackMsg, DevServerBackMsg, VenvBackMsg, BoilerplateBackMsg, BonusBackMsg:
		return m, tea.Quit
	}

	newModel, cmd := m.model.Update(msg)
	m.model = newModel
	return m, cmd
}

func (m StandaloneWrapper) View() string {
	return m.model.View()
}
