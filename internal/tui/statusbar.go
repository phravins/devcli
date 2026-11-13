package tui

import (
	"context"
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

const statusRefreshInterval = 5 * time.Second

// statusInfo is everything shown in the global status bar.
type statusInfo struct {
	branch    string
	venv      string
	memory    string
	aiBackend string
}

var (
	statusMu         sync.Mutex
	statusCached     = statusInfo{branch: "No Git", venv: "venv:none", memory: "-", aiBackend: "Ollama"}
	statusLastUpdate time.Time
	statusRefreshing bool
	statusConfigRead bool

	statusRenderWidth int
	statusRenderInfo  statusInfo
	statusRendered    string
)

var (
	statusBrandStyle  = statusSegment("#BD93F9")
	statusBranchStyle = statusSegment("#FF79C6")
	statusVenvStyle   = statusSegment("#50FA7B")
	statusMemStyle    = statusSegment("#8BE9FD")
	statusAIStyle     = statusSegment("#F1FA8C")
	statusBgColor     = lipgloss.Color("#44475a")
)

func statusSegment(bg string) lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color(bg)).
		Foreground(lipgloss.Color("#282a36")).
		Padding(0, 1)
}

// collectStatusInfo does the slow work (spawning git, stat, a
// stop-the-world ReadMemStats). It must never run on the render path.
func collectStatusInfo() statusInfo {
	info := statusInfo{branch: "No Git", venv: "venv:none"}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if out, err := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD").Output(); err == nil {
		if b := strings.TrimSpace(string(out)); b != "" {
			info.branch = "git:" + b
		}
	}

	if v := os.Getenv("VIRTUAL_ENV"); v != "" {
		info.venv = "venv:" + filepath.Base(v)
	} else if _, err := os.Stat(".venv"); err == nil {
		info.venv = "venv:.venv"
	}

	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	info.memory = fmt.Sprintf("%.1fMB", float64(ms.Alloc)/1024/1024)

	return info
}

func currentAIBackend() string {
	// Read the config file from disk once; afterwards use viper's in-memory
	// values, which Settings updates via config.Set, so a backend change is
	// reflected immediately without re-reading the file every few seconds.
	if !statusConfigRead {
		statusConfigRead = true
		_, _ = config.LoadConfig()
	}
	b := config.GetString("ai_backend")
	if b == "" {
		return "Ollama"
	}
	return strings.ToUpper(b[:1]) + b[1:]
}

// getStatusInfo returns the latest cached info immediately and, when it is
// stale, refreshes it in the background so View() never blocks.
func getStatusInfo() statusInfo {
	statusMu.Lock()
	defer statusMu.Unlock()

	if !statusRefreshing && time.Since(statusLastUpdate) > statusRefreshInterval {
		statusRefreshing = true
		go func() {
			info := collectStatusInfo()
			statusMu.Lock()
			statusCached.branch = info.branch
			statusCached.venv = info.venv
			statusCached.memory = info.memory
			statusLastUpdate = time.Now()
			statusRefreshing = false
			statusMu.Unlock()
		}()
	}

	info := statusCached
	info.aiBackend = currentAIBackend()
	return info
}

func RenderStatusBar(width int) string {
	if width < 20 {
		width = 80
	}

	info := getStatusInfo()

	statusMu.Lock()
	defer statusMu.Unlock()
	if statusRendered != "" && width == statusRenderWidth && info == statusRenderInfo {
		return statusRendered
	}

	bar := lipgloss.JoinHorizontal(
		lipgloss.Top,
		statusBrandStyle.Render("⚡ DevCLI "+config.Version),
		statusBranchStyle.Render(" "+info.branch),
		statusVenvStyle.Render(" "+info.venv),
		statusMemStyle.Render(" RAM:"+info.memory),
		statusAIStyle.Render(" AI:"+info.aiBackend),
	)

	// Truncate before padding: applying Width to an over-long bar would wrap
	// it onto a second line and push every view down by one row.
	bar = lipgloss.NewStyle().MaxWidth(width).Render(bar)
	statusRendered = lipgloss.NewStyle().
		Background(statusBgColor).
		Width(width).
		Render(bar)
	statusRenderWidth = width
	statusRenderInfo = info
	return statusRendered
}
