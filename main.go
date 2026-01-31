package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

			// Destination Directory
			var binDir string
			var destPath string
			if runtime.GOOS == "windows" {
				binDir = filepath.Join(home, ".devcli", "bin")
				destPath = filepath.Join(binDir, "devcli.exe")
			} else {
				// Unix: standard is often ~/.local/bin or just ~/go/bin, but let's stick to .devcli/bin for consistency
				// or use /usr/local/bin if root? Let's use user-local to avoid permission issues.
				binDir = filepath.Join(home, ".devcli", "bin")
				destPath = filepath.Join(binDir, "devcli")
			}

			if err := os.MkdirAll(binDir, 0755); err != nil {
				fmt.Printf("Error creating bin directory: %v\n", err)
				return
			}

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

			// Update PATH
			if runtime.GOOS == "windows" {
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
			} else {
				// Unix/Linux/Mac
				// Check current Shell
				shell := os.Getenv("SHELL")
				rcFile := ""
				if strings.Contains(shell, "zsh") {
					rcFile = filepath.Join(home, ".zshrc")
				} else if strings.Contains(shell, "bash") {
					rcFile = filepath.Join(home, ".bashrc")
				}

				// Check if already in PATH (rough check)
				pathEnv := os.Getenv("PATH")
				if !strings.Contains(pathEnv, binDir) {
					if rcFile != "" {
						fmt.Printf("Detecting shell: %s. Attempting to update %s...\n", shell, rcFile)
						// Append export line
						exportLine := fmt.Sprintf("\nexport PATH=$PATH:%s\n", binDir)

						// Check if file already has it
						content, _ := os.ReadFile(rcFile)
						if strings.Contains(string(content), binDir) {
							fmt.Println("PATH already seems to be configured in RC file.")
						} else {
							f, err := os.OpenFile(rcFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
							if err != nil {
								fmt.Printf("Error opening rc file: %v\n", err)
							} else {
								defer f.Close()
								if _, err = f.WriteString(exportLine); err != nil {
									fmt.Printf("Error writing to rc file: %v\n", err)
								} else {
									fmt.Println("Successfully added to install path to shell configuration.")
									fmt.Printf("Run 'source %s' or restart terminal to apply changes.\n", rcFile)
								}
							}
						}
					} else {
						fmt.Printf("Could not detect shell configuration file (.bashrc/.zshrc).\n")
						fmt.Printf("Please manually add the following to your PATH:\n%s\n", binDir)
					}
				} else {
					fmt.Println("DevCLI is already in your PATH.")
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
