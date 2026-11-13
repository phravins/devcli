package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// The dashboard (plus the global status bar) must fill the terminal exactly:
// overflowing pushes the logo off-screen, and wrapping garbles it.
func TestDashboardFitsTerminal(t *testing.T) {
	for _, sz := range [][2]int{{36, 24}, {60, 14}, {80, 20}, {80, 24}, {120, 40}} {
		var m tea.Model = NewRootModel()
		m, _ = m.Update(tea.WindowSizeMsg{Width: sz[0], Height: sz[1]})
		out := m.View()
		if got := len(strings.Split(out, "\n")); got != sz[1] {
			t.Errorf("%dx%d: rendered %d lines, want %d", sz[0], sz[1], got, sz[1])
		}
		if got := lipgloss.Width(out); got > sz[0] {
			t.Errorf("%dx%d: rendered width %d exceeds terminal", sz[0], sz[1], got)
		}
	}
}

func TestDashboardLogoIntact(t *testing.T) {
	out := renderDashboardHeader(100, 40)
	for _, line := range strings.Split(dashLogo, "\n") {
		if !strings.Contains(out, strings.TrimRight(line, " ")) {
			t.Fatalf("logo line missing or wrapped: %q", line)
		}
	}
	if strings.Contains(out, `\\`) {
		t.Fatal("logo contains doubled backslashes")
	}
}

func BenchmarkRootView(b *testing.B) {
	var m tea.Model = NewRootModel()
	m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = m.View()
	}
}
