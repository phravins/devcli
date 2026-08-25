package auth

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var AuthCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage DevCLI security, user credentials, and authentication",
	Long: `DevCLI security suite for setting up master credentials, checking security status,
authenticating, and managing account credentials.`,
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Perform post-installation user account setup (Username & Password)",
	Run: func(cmd *cobra.Command, args []string) {
		if IsSetup() {
			fmt.Println("🔒 Account setup already exists for DevCLI.")
			fmt.Print("Do you want to re-configure your account? [y/N]: ")
			reader := bufio.NewReader(os.Stdin)
			ans, _ := reader.ReadString('\n')
			if strings.ToLower(strings.TrimSpace(ans)) != "y" {
				fmt.Println("Setup cancelled.")
				return
			}
		}

		fmt.Println("==================================================")
		fmt.Println("       DevCLI Production Security Account Setup   ")
		fmt.Println("==================================================")

		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter Username: ")
		username, _ := reader.ReadString('\n')
		username = strings.TrimSpace(username)
		if username == "" {
			fmt.Println("❌ Error: Username cannot be empty.")
			return
		}

		fmt.Print("Enter Strong Password (min 8 chars, A-Z, a-z, 0-9): ")
		bytePassword, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			fmt.Printf("❌ Error reading password: %v\n", err)
			return
		}
		password := string(bytePassword)

		if err := ValidatePasswordStrength(password); err != nil {
			fmt.Printf("❌ Weak Password: %v\n", err)
			return
		}

		fmt.Print("Confirm Password: ")
		byteConfirm, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil || string(byteConfirm) != password {
			fmt.Println("❌ Passwords do not match!")
			return
		}

		if err := SetupUser(username, password); err != nil {
			fmt.Printf("❌ Failed to setup account: %v\n", err)
			return
		}

		path, _ := GetAuthFilePath()
		fmt.Println("==================================================")
		fmt.Println("✅ Account created successfully!")
		fmt.Printf("   User       : %s\n", username)
		fmt.Printf("   Credentials: %s (Permissions: 0600)\n", path)
		fmt.Println("==================================================")
	},
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate DevCLI session with your password",
	Run: func(cmd *cobra.Command, args []string) {
		if !IsSetup() {
			fmt.Println("⚠️  No account setup found. Please run 'devcli auth setup' first.")
			return
		}

		data, err := GetAuthData()
		if err != nil || data == nil {
			fmt.Println("❌ Error loading account credentials.")
			return
		}

		fmt.Printf("Authenticating user '%s'...\n", data.Username)
		fmt.Print("Enter Password: ")
		bytePassword, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			return
		}

		valid, err := VerifyPassword(string(bytePassword))
		if err != nil || !valid {
			fmt.Println("❌ Authentication failed: Invalid password.")
			os.Exit(1)
		}

		fmt.Println("✅ Authentication successful! Session unlocked.")
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check DevCLI security status and storage permissions",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("==================================================")
		fmt.Println("           DevCLI Security & Auth Status          ")
		fmt.Println("==================================================")

		path, err := GetAuthFilePath()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		setup := IsSetup()
		if setup {
			data, _ := GetAuthData()
			fmt.Println(" Account Status    : ACTIVE 🔒")
			fmt.Printf(" Username          : %s\n", data.Username)
			fmt.Printf(" Created At        : %s\n", data.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf(" Last Login        : %s\n", data.LastLogin.Format("2006-01-02 15:04:05"))
			fmt.Printf(" Security Standard : %s (bcrypt cost %d)\n", data.SecurityVersion, BcryptCost)
		} else {
			fmt.Println(" Account Status    : UNCONFIGURED ⚠️")
			fmt.Println(" Action Required   : Run 'devcli auth setup' to create master credentials")
		}

		fmt.Println("\n Storage Integrity:")
		fmt.Printf(" Vault Path        : %s\n", path)
		if info, err := os.Stat(path); err == nil {
			fmt.Printf(" File Permissions  : %04o (Owner read/write only)\n", info.Mode().Perm())
		} else {
			fmt.Println(" Vault Status      : File not created yet")
		}

		dir, _ := GetAuthDir()
		if dirInfo, err := os.Stat(dir); err == nil {
			fmt.Printf(" Dir Permissions   : %04o (%s)\n", dirInfo.Mode().Perm(), dir)
		}
		fmt.Println("==================================================")
	},
}

var changePasswordCmd = &cobra.Command{
	Use:   "change-password",
	Short: "Update your account password",
	Run: func(cmd *cobra.Command, args []string) {
		if !IsSetup() {
			fmt.Println("⚠️ No account setup found. Please run 'devcli auth setup' first.")
			return
		}

		fmt.Print("Enter Current Password: ")
		byteCurrent, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			return
		}

		fmt.Print("Enter New Strong Password: ")
		byteNew, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			return
		}

		fmt.Print("Confirm New Password: ")
		byteConfirm, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil || string(byteConfirm) != string(byteNew) {
			fmt.Println("❌ Passwords do not match!")
			return
		}

		if err := ChangePassword(string(byteCurrent), string(byteNew)); err != nil {
			fmt.Printf("❌ Failed to change password: %v\n", err)
			return
		}

		fmt.Println("✅ Password changed successfully!")
	},
}

func init() {
	AuthCmd.AddCommand(setupCmd)
	AuthCmd.AddCommand(loginCmd)
	AuthCmd.AddCommand(statusCmd)
	AuthCmd.AddCommand(changePasswordCmd)
}

func SetupCLI() error {
	fmt.Println("==================================================")
	fmt.Println("       DevCLI Production Security Account Setup   ")
	fmt.Println("==================================================")

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter Username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	fmt.Print("Enter Strong Password (min 8 chars, A-Z, a-z, 0-9): ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	password := string(bytePassword)

	if err := ValidatePasswordStrength(password); err != nil {
		return fmt.Errorf("weak password: %w", err)
	}

	fmt.Print("Confirm Password: ")
	byteConfirm, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil || string(byteConfirm) != password {
		return fmt.Errorf("passwords do not match")
	}

	if err := SetupUser(username, password); err != nil {
		return fmt.Errorf("failed to setup account: %w", err)
	}

	path, _ := GetAuthFilePath()
	fmt.Println("==================================================")
	fmt.Println("✅ Account created successfully!")
	fmt.Printf("   User       : %s\n", username)
	fmt.Printf("   Credentials: %s (Permissions: 0600)\n", path)
	fmt.Println("==================================================")
	return nil
}

func RequireCLILogin() error {
	data, err := GetAuthData()
	if err != nil || data == nil {
		return fmt.Errorf("failed to load account credentials")
	}

	fmt.Printf("🔒 DevCLI Locked :: User '%s'\n", data.Username)
	fmt.Print("Enter Password: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return fmt.Errorf("error reading password: %w", err)
	}

	valid, err := VerifyPassword(string(bytePassword))
	if err != nil || !valid {
		return fmt.Errorf("authentication failed: invalid password")
	}

	fmt.Println("✅ Authentication successful!")
	return nil
}
