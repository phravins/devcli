package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func generateCommandsHelp() string {
	sectionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#0F9E99")).Bold(true).Underline(true)
	cmdStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF79C6")).Bold(true)

	var cmds strings.Builder
	cmds.WriteString("\n")

	addCmd := func(name, desc string) {
		cmds.WriteString(fmt.Sprintf("  %s %s\n", cmdStyle.Render(fmt.Sprintf("%-18s", name)), descStyle.Render(desc)))
	}
	addKey := func(key, desc string) {
		cmds.WriteString(fmt.Sprintf("  %s %s\n", keyStyle.Render(fmt.Sprintf("%-18s", key)), descStyle.Render(desc)))
	}

	// 1. CLI Commands
	cmds.WriteString(sectionStyle.Render("CORE CLI:") + "\n")
	addCmd("devcli dev", "Project Creation & Tools")
	addCmd("devcli file", "File Manager")
	addCmd("devcli ai", "AI Chat")
	addCmd("devcli editor", "TUI Editor")
	cmds.WriteString("\n")

	// 2. Global Navigation
	cmds.WriteString(sectionStyle.Render("NAVIGATION:") + "\n")
	addKey("↑ / ↓", "Move Up / Down")
	addKey("Enter", "Select / Confirm")
	addKey("Esc / q", "Go Back / Exit")
	cmds.WriteString("\n")

	// 3. Project Tools
	cmds.WriteString(sectionStyle.Render("PROJECT TOOLS:") + "\n")
	addKey("b", "Backup Project (List)")
	addKey("d", "Delete History (History)")
	cmds.WriteString("\n")

	// 4. Dev Server
	cmds.WriteString(sectionStyle.Render("DEV SERVER:") + "\n")
	addKey("s", "Start/Stop Server")
	addKey("f", "Toggle Filters")
	addKey("b", "Toggle Server Source (Fullstack)")
	addKey("/", "Search Logs")
	addKey("a", "Toggle Auto-scroll")
	addKey("c", "Clear Logs")
	addKey("?", "Help & Documentation")
	cmds.WriteString("\n")

	// 5. Venv Wizard
	cmds.WriteString(sectionStyle.Render("VENV WIZARD:") + "\n")
	addKey("n", "New Environment")
	addKey("s", "Scan System")
	addKey("y", "Sync Packages")
	addKey("c", "Clone Environment")
	addKey("d", "Delete Environment")
	cmds.WriteString("\n")

	// 6. File Manager
	cmds.WriteString(sectionStyle.Render("FILE MANAGER:") + "\n")
	addKey("Tab", "Toggle Global Search")
	addKey("Alt+M", "Move/Rename File")
	addKey("Alt+C", "Copy File")
	addKey("Alt+E", "Edit File")
	cmds.WriteString("\n")

	// 7. AI Chat
	cmds.WriteString(sectionStyle.Render("AI CHAT:") + "\n")
	addKey("Enter", "Send Message")
	addKey("Esc", "Exit Chat")
	cmds.WriteString("\n")

	// 8. Editor Shortcuts
	cmds.WriteString(sectionStyle.Render("EDITOR (Multi-Lang):") + "\n")
	addKey("Ctrl+R", "Run Code")
	addKey("Ctrl+S", "Save File")
	addKey("Ctrl+N", "New File")
	addKey("Ctrl+P", "Command Prompt")
	addKey("Ctrl+H", "Toggle Help")
	addKey("Ctrl+C", "Exit Editor")

	cmds.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("  Press Esc to go back"))

	return cmds.String()
}
