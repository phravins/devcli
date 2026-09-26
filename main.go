package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/phravins/devcli/assets"
	"github.com/phravins/devcli/internal/ai"
	"github.com/phravins/devcli/internal/aicommit"
	"github.com/phravins/devcli/internal/fileops"
	"github.com/phravins/devcli/internal/project"
	"github.com/phravins/devcli/internal/tui"
	"github.com/phravins/devcli/internal/updater"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:		"devcli",
	Version:	"1.1.0",
	Short:		"A comprehensive CLI for developers",
	Long: `DevCLI is a powerful command-line interface that provides:
- Local development tools
- File operations
- AI chatbot integration
- Built-in Python IDE`,
}

var startCmd = &cobra.Command{
	Use:	"start [name] [stack]",
	Short:	"Initialize a new project",
	Args:	cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		stack := "Go"
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
}

var timemachineCmd = &cobra.Command{
	Use:	"timemachine [file]",
	Short:	"Code Time Machine - Track code evolution and find bugs",
	Long:	`Interactive Git blame and history viewer showing line-by-line changes, bug risks, and commit timeline.`,
	Args:	cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var filePath string
		if len(args) > 0 {
			filePath = args[0]
		}
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error getting current directory: %v\n", err)
			os.Exit(1)
		}
		if err := tui.RunTimeMachine(cwd, filePath); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	},
}

var installCmd = &cobra.Command{
	Use:	"install",
	Short:	"Install DevCLI globally to your system",
	Long:	`Copies the DevCLI binary to your home directory and adds it to your system PATH.`,
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

		var binDir string
		var destPath string
		if runtime.GOOS == "windows" {
			binDir = filepath.Join(home, ".devcli", "bin")
			destPath = filepath.Join(binDir, "devcli.exe")
		} else {
			binDir = filepath.Join(home, ".devcli", "bin")
			destPath = filepath.Join(binDir, "devcli")
		}

		if err := os.MkdirAll(binDir, 0755); err != nil {
			fmt.Printf("Error creating bin directory: %v\n", err)
			return
		}

		evalExe, err1 := filepath.EvalSymlinks(exePath)
		evalDest, err2 := filepath.EvalSymlinks(destPath)
		isSame := (err1 == nil && err2 == nil && evalExe == evalDest) || (exePath == destPath)

		if !isSame {
			data, err := os.ReadFile(exePath)
			if err != nil {
				fmt.Printf("Error reading current binary: %v\n", err)
				return
			}
			if err := os.WriteFile(destPath, data, 0755); err != nil {
				fmt.Printf("Error copying binary: %v\n", err)
				return
			}
		}

		fmt.Printf("Binary deployed to: %s\n", destPath)

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

			out, err := exec.Command("powershell", "-NonInteractive", "-NoProfile", "-Command", script).CombinedOutput()
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

			if runtime.GOOS == "linux" {
				fmt.Println("Configuring Linux Desktop integration...")

				logoBytes, err := assets.GetLogo()
				if err == nil && len(logoBytes) > 0 {
					iconDirs := []string{
						filepath.Join(home, ".local", "share", "icons", "hicolor", "512x512", "apps"),
						filepath.Join(home, ".local", "share", "pixmaps"),
					}
					for _, iconDir := range iconDirs {
						if err := os.MkdirAll(iconDir, 0755); err == nil {
							iconPath := filepath.Join(iconDir, "devcli.png")
							os.WriteFile(iconPath, logoBytes, 0644)
						}
					}
					fmt.Println("Installed DevCLI icon to ~/.local/share/icons/")
				}

				appsDir := filepath.Join(home, ".local", "share", "applications")
				if err := os.MkdirAll(appsDir, 0755); err == nil {
					desktopPath := filepath.Join(appsDir, "devcli.desktop")
					desktopContent := fmt.Sprintf(`[Desktop Entry]
Version=1.0
Type=Application
Name=DevCLI
GenericName=Developer CLI Workspace
Comment=Terminal-based developer workspace, IDE, and AI assistant
Exec=%s
Icon=devcli
Terminal=true
Categories=Development;IDE;ConsoleOnly;
Keywords=developer;cli;terminal;ide;ai;workspace;
StartupNotify=true
`, destPath)

					if err := os.WriteFile(desktopPath, []byte(desktopContent), 0644); err == nil {
						fmt.Println("Created Linux Desktop application entry: " + desktopPath)
					}
				}

				exec.Command("update-desktop-database", filepath.Join(home, ".local", "share", "applications")).Run()
				exec.Command("gtk-update-icon-cache", filepath.Join(home, ".local", "share", "icons", "hicolor")).Run()
			}

			shell := os.Getenv("SHELL")
			rcFile := ""

			if strings.Contains(shell, "zsh") {
				rcFile = filepath.Join(home, ".zshrc")
			} else if strings.Contains(shell, "bash") {
				rcFile = filepath.Join(home, ".bashrc")
			} else if strings.Contains(shell, "fish") {

				configDir, _ := os.UserConfigDir()
				rcFile = filepath.Join(configDir, "fish", "config.fish")
			}

			pathEnv := os.Getenv("PATH")
			if !strings.Contains(pathEnv, binDir) {
				if rcFile != "" {
					fmt.Printf("Detecting shell: %s. Attempting to update %s...\n", shell, rcFile)

					var exportLine string
					if strings.Contains(shell, "fish") {
						exportLine = fmt.Sprintf("\n# DevCLI\nset -gx PATH $PATH %s\n", binDir)
					} else {
						exportLine = fmt.Sprintf("\n# DevCLI\nexport PATH=\"$PATH:%s\"\n", binDir)
					}

					content, err := os.ReadFile(rcFile)

					if os.IsNotExist(err) {

						if strings.Contains(shell, "fish") {
							os.MkdirAll(filepath.Dir(rcFile), 0755)
						}
					}

					if err == nil && strings.Contains(string(content), binDir) {
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
								fmt.Println("Successfully added install path to shell configuration.")
								fmt.Printf("Run 'source %s' or restart terminal to apply changes.\n", rcFile)
							}
						}
					}
				} else {
					fmt.Printf("Could not detect shell configuration file (.bashrc/.zshrc/config.fish).\n")
					fmt.Printf("Please manually add the following to your PATH:\n%s\n", binDir)
				}
			} else {
				fmt.Println("DevCLI is already in your PATH.")
			}
		}
	},
}

var updateCmd = &cobra.Command{
	Use:	"update",
	Short:	"Update DevCLI to the latest version",
	Long:	`Checks for the latest version of DevCLI on GitHub and updates the binary if a new version is available.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔍 Checking for updates...")

		info, err := updater.CheckForUpdates()
		if err != nil {
			fmt.Printf("❌ Error checking for updates: %v\n", err)
			return
		}

		fmt.Printf("📦 Current version: %s\n", info.CurrentVersion)
		fmt.Printf("📦 Latest version:  %s\n", info.LatestVersion)

		if !info.IsUpdateAvailable {
			fmt.Println("✅ You are already running the latest version!")
			return
		}

		fmt.Printf("\n🎉 New version available: %s\n", info.LatestVersion)
		if info.ReleaseNotes != "" {
			fmt.Printf("\n📝 Release Notes:\n%s\n", info.ReleaseNotes)
		}

		fmt.Println("\n⬇️  Downloading and installing update...")
		if err := updater.PerformUpdate(); err != nil {
			fmt.Printf("❌ Update failed: %v\n", err)
			fmt.Println("\n💡 You can try updating manually by downloading from:")
			fmt.Printf("   %s\n", info.ReleaseURL)
			return
		}

		fmt.Println("✅ Update successful!")
		fmt.Println("🔄 Please restart DevCLI to use the new version.")
	},
}

func init() {

	fileops.FileCmd.Run = func(cmd *cobra.Command, args []string) {
		tui.RunFileManager("")
	}
	rootCmd.AddCommand(fileops.FileCmd)
	rootCmd.AddCommand(ai.AICmd)
	rootCmd.AddCommand(tui.EditorCmd)
	ai.AICmd.AddCommand(tui.ChatCmd)
	ai.AICmd.AddCommand(aicommit.CommitCmd)

	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(timemachineCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(updateCmd)
}

func main() {

	if len(os.Args) > 1 {
		if err := rootCmd.Execute(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return
	}

	tui.RunRoot()
}
