package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func RunDevServer(path string, opts ...tea.ProgramOption) error {
	if path == "" {
		path, _ = os.Getwd()
	}
	allOpts := append([]tea.ProgramOption{tea.WithAltScreen()}, opts...)
	p := tea.NewProgram(Wrap(NewDevServerDashboardModel(path)), allOpts...)
	_, err := p.Run()
	if err != nil {
		fmt.Printf("Error running dev server dashboard: %v\n", err)
	}
	return err
}

func RunFileManager(path string, opts ...tea.ProgramOption) error {
	if path == "" {
		path, _ = os.Getwd()
	}
	allOpts := append([]tea.ProgramOption{tea.WithAltScreen()}, opts...)
	p := tea.NewProgram(Wrap(NewFileManagerModel(path)), allOpts...)
	_, err := p.Run()
	if err != nil {
		fmt.Printf("Error running file manager: %v\n", err)
	}
	return err
}
