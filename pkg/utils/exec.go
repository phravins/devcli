package utils

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func OpenBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
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
		if _, err := exec.LookPath("powershell"); err == nil {
			return exec.Command("powershell", "-Command", command)
		}
		return exec.Command("cmd", "/C", command)
	}

	shell := os.Getenv("SHELL")
	if shell != "" {
		return exec.Command(shell, "-c", command)
	}
	if _, err := exec.LookPath("bash"); err == nil {
		return exec.Command("bash", "-c", command)
	}
	return exec.Command("sh", "-c", command)
}
