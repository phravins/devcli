package utils

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// OpenBrowser opens the specified URL in the default browser in a cross-platform way.
func OpenBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		// rundll32 is generally reliable for file protocols, 'start' via cmd is also common but rundll32 is used in the existing code.
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return cmd.Start()
}
func OpenFile(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", path)
	case "windows":
		// 'explorer' is the standard way to open files/folders in Windows
		cmd = exec.Command("explorer", path)
	case "darwin":
		cmd = exec.Command("open", path)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return cmd.Start()
}

func GetShellCommand(command string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		// Try PowerShell first, then Cmd
		if _, err := exec.LookPath("powershell"); err == nil {
			return exec.Command("powershell", "-Command", command)
		}
		return exec.Command("cmd", "/C", command)
	}

	// Unix-like (Linux, macOS)
	// Use $SHELL if set, otherwise fallback to sh
	shell := os.Getenv("SHELL")
	if shell != "" {
		return exec.Command(shell, "-c", command)
	}
	// Fallback
	if _, err := exec.LookPath("bash"); err == nil {
		return exec.Command("bash", "-c", command)
	}
	return exec.Command("sh", "-c", command)
}
