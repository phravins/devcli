package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func RunDevServer(path string) {
	if path == "" {
		path, _ = os.Getwd()
	}
	p := tea.NewProgram(Wrap(NewDevServerDashboardModel(path)), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running dev server dashboard: %v\n", err)
		os.Exit(1)
	}
}

func RunFileManager(path string) {
	if path == "" {
		path, _ = os.Getwd()
	}
	p := tea.NewProgram(Wrap(NewFileManagerModel(path)), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running file manager: %v\n", err)
		os.Exit(1)
	}
}
