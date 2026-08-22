package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/phravins/devcli/internal/config"
)

// RenderStatusBar generates a modern responsive live status bar for DevCLI
func RenderStatusBar(width int) string {
	if width < 20 {
		width = 80
	}

	// 1. Git Branch
	branch := "No Git"
	if out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output(); err == nil {
		b := strings.TrimSpace(string(out))
		if b != "" {
			branch = "git:" + b
		}
	}

	// 2. Memory Usage (Allocated MB)
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memUsed := float64(m.Alloc) / 1024 / 1024
	memStr := fmt.Sprintf("%.1fMB", memUsed)

	// 3. Virtual Environment
	venvStr := "venv:none"
	if v := os.Getenv("VIRTUAL_ENV"); v != "" {
		venvStr = "venv:" + filepath.Base(v)
	} else if _, err := os.Stat(".venv"); err == nil {
		venvStr = "venv:.venv"
	}

	// 4. Configured AI Provider
	cfg, _ := config.LoadConfig()
	aiBackend := "Ollama"
	if cfg != nil && cfg.AIBackend != "" {
		aiBackend = strings.Title(cfg.AIBackend)
	}

	// Styles for Status Bar Pills
	brandStyle := lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color("#BD93F9")).
		Foreground(lipgloss.Color("#282a36")).
		Padding(0, 1)

	branchStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#FF79C6")).
		Foreground(lipgloss.Color("#282a36")).
		Bold(true).
		Padding(0, 1)

	venvStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#50FA7B")).
		Foreground(lipgloss.Color("#282a36")).
		Bold(true).
		Padding(0, 1)

	memStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#8BE9FD")).
		Foreground(lipgloss.Color("#282a36")).
		Bold(true).
		Padding(0, 1)

	aiStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#F1FA8C")).
		Foreground(lipgloss.Color("#282a36")).
		Bold(true).
		Padding(0, 1)

	barContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		brandStyle.Render("⚡ DevCLI v1.0"),
		branchStyle.Render(" "+branch),
		venvStyle.Render(" "+venvStr),
		memStyle.Render(" RAM:"+memStr),
		aiStyle.Render(" AI:"+aiBackend),
	)

	// Fill background to full width
	backgroundStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#44475a")).
		Width(width)

	return backgroundStyle.Render(barContent)
}
