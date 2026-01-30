package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/phravins/devcli/internal/ai"
	"github.com/phravins/devcli/internal/devserver"
	"github.com/phravins/devcli/internal/fileops"
	"github.com/phravins/devcli/internal/project"
	"github.com/phravins/devcli/internal/tui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "devcli",
	Version: "1.0.0",
	Short:   "A comprehensive CLI for developers",
	Long: `DevCLI is a powerful command-line interface that provides:
- Local development tools
- File operations
- AI chatbot integration
- Built-in Python IDE`,
}

func init() {
	// Add all subcommands
	rootCmd.AddCommand(fileops.FileCmd)
	rootCmd.AddCommand(ai.AICmd)
	rootCmd.AddCommand(tui.EditorCmd)
	ai.AICmd.AddCommand(tui.ChatCmd)
	rootCmd.AddCommand(&cobra.Command{
		Use:   "start [name] [stack]",
		Short: "Initialize a new project",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			stack := "Go" // Default
			if len(args) > 1 {
				stack = args[1]
			}

			mgr := project.NewManager("")
			fmt.Printf("Creating %s project '%s'...\n", stack, name)
			if _, _, err := mgr.CreateProject(name, "Go Fiber API", ""); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Project created successfully in ./%s\n", name)
			}
		},
	})
	rootCmd.AddCommand(&cobra.Command{
		Use:   "dev",
		Short: "Auto-detect and run dev server",
		Run: func(cmd *cobra.Command, args []string) {
			cwd, _ := os.Getwd()
			info := devserver.Detect(cwd)
			if info.Type == devserver.TypeUnknown {
				fmt.Println(" Could not detect project type (Node, Python, Go).")
				return
			}
			if err := devserver.Run(info); err != nil {
				fmt.Printf("Error running server: %v\n", err)
			}
		},
	})
	rootCmd.AddCommand(&cobra.Command{
		Use:   "install",
		Short: "Install DevCLI globally to your system",
		Long:  `Copies the DevCLI binary to your home directory and adds it to your system PATH.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Starting DevCLI installation...")

			exePath, err := os.Executable()
			if err != nil {
				fmt.Printf("Error finding executable: %v\n", err)
				return
			}

			home, err := os.UserHomeDir()
			if err != nil {
				fmt.Printf("Error finding home directory: %v\n", err)
				return
			}

			binDir := filepath.Join(home, ".devcli", "bin")
			if err := os.MkdirAll(binDir, 0755); err != nil {
				fmt.Printf("Error creating bin directory: %v\n", err)
				return
			}

			destPath := filepath.Join(binDir, "devcli.exe")
			// Read binary
			data, err := os.ReadFile(exePath)
			if err != nil {
				fmt.Printf("Error reading current binary: %v\n", err)
				return
			}

			// Write to destination
			if err := os.WriteFile(destPath, data, 0755); err != nil {
				fmt.Printf("Error copying binary: %v\n", err)
				return
			}

			fmt.Printf("Binary deployed to: %s\n", destPath)
			script := fmt.Sprintf(`
				$binPath = "%s"
				$currentPath = [System.Environment]::GetEnvironmentVariable("Path", "User")
				if ($currentPath -notlike "*$binPath*") {
					[System.Environment]::SetEnvironmentVariable("Path", $currentPath + ";" + $binPath, "User")
					Write-Output "ADDED"
				} else {
					Write-Output "EXISTS"
				}
			`, binDir)

			out, err := exec.Command("powershell", "-Command", script).CombinedOutput()
			if err != nil {
				fmt.Printf("Warning: Automated PATH update failed: %v\n", err)
				fmt.Printf("Please add this folder to your PATH manually: %s\n", binDir)
			} else {
				res := strings.TrimSpace(string(out))
				if res == "ADDED" {
					fmt.Println("Successfully added DevCLI to your User PATH!")
					fmt.Println("Installation complete. PLEASE RESTART YOUR TERMINAL to use 'devcli' from anywhere.")
				} else {
					fmt.Println("DevCLI already exists in your PATH.")
					fmt.Println("Installation complete.")
				}
			}
		},
	})

}
func main() {
	// If args were passed (CLI mode), just run once
	if len(os.Args) > 1 {
		if err := rootCmd.Execute(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return
	}
	// Default TUI mode (Unified Root)
	tui.RunRoot()
}
