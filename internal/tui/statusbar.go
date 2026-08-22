package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/phravins/devcli/internal/config"
)

var (
	statusCacheMu   sync.Mutex
	lastStatusCheck time.Time
	cachedBranch    string = "No Git"
	cachedVenv      string = "venv:none"
	cachedAIBackend string = "Ollama"
)

func getCachedStatusInfo() (string, string, string) {
	statusCacheMu.Lock()
	defer statusCacheMu.Unlock()

	// Refresh metrics at most once every 5 seconds to prevent process spawning lag in View()
	if time.Since(lastStatusCheck) > 5*time.Second || lastStatusCheck.IsZero() {
		lastStatusCheck = time.Now()

		// 1. Git Branch
		branch := "No Git"
		if out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output(); err == nil {
			b := strings.TrimSpace(string(out))
			if b != "" {
				branch = "git:" + b
			}
		}
		cachedBranch = branch

		// 2. Virtual Environment
		venvStr := "venv:none"
		if v := os.Getenv("VIRTUAL_ENV"); v != "" {
			venvStr = "venv:" + filepath.Base(v)
		} else if _, err := os.Stat(".venv"); err == nil {
			venvStr = "venv:.venv"
		}
		cachedVenv = venvStr

		// 3. Configured AI Provider
		cfg, _ := config.LoadConfig()
		aiBackend := "Ollama"
		if cfg != nil && cfg.AIBackend != "" {
			aiBackend = strings.Title(cfg.AIBackend)
		}
		cachedAIBackend = aiBackend
	}

	return cachedBranch, cachedVenv, cachedAIBackend
}

// RenderStatusBar generates a modern responsive live status bar for DevCLI
func RenderStatusBar(width int) string {
	if width < 20 {
		width = 80
	}

	branch, venvStr, aiBackend := getCachedStatusInfo()

	// Memory Usage (Allocated MB - fast in-memory calculation)
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memUsed := float64(m.Alloc) / 1024 / 1024
	memStr := fmt.Sprintf("%.1fMB", memUsed)

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
		brandStyle.Render("⚡ DevCLI v1.1"),
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
