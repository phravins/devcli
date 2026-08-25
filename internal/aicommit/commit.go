package aicommit

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/phravins/devcli/internal/ai"
	"github.com/phravins/devcli/internal/ai/providers"
	"github.com/phravins/devcli/internal/config"
	"github.com/spf13/cobra"
)

var CommitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Auto-generate conventional Git commit messages using AI",
	Long:  `Inspects your Git diff, generates a conventional commit message via AI, and commits changes upon approval.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔍 Analyzing git changes...")

		// Check if inside a git repository
		if err := exec.Command("git", "rev-parse", "--is-inside-work-tree").Run(); err != nil {
			fmt.Println("❌ Error: Not inside a Git repository.")
			return
		}

		// Get staged diff
		diffBytes, err := exec.Command("git", "diff", "--cached").Output()
		if err != nil {
			fmt.Printf("❌ Error running git diff: %v\n", err)
			return
		}
		diffStr := strings.TrimSpace(string(diffBytes))

		// If no staged diff, fallback to unstaged diff
		if diffStr == "" {
			diffBytes, _ = exec.Command("git", "diff").Output()
			diffStr = strings.TrimSpace(string(diffBytes))
		}

		if diffStr == "" {
			fmt.Println("ℹ️ No git changes detected. Stage or modify files first!")
			return
		}

		// Limit diff size to prevent token overflow
		if len(diffStr) > 4000 {
			diffStr = diffStr[:4000] + "\n... [diff truncated]"
		}

		// Load provider
		cfg, _ := config.LoadConfig()
		provider, err := providers.GetProvider(cfg)
		if err != nil || provider == nil {
			provider = &providers.OpenAIProvider{}
			provider.Configure(cfg)
		}

		fmt.Println("🤖 Generating AI commit message...")

		prompt := fmt.Sprintf("Generate a single-line conventional git commit message (e.g. 'feat: ...' or 'fix: ...') for the following diff. Output ONLY the commit title line and nothing else.\n\nDiff:\n%s", diffStr)

		msgs := []ai.Message{{Role: "user", Content: prompt}}
		resp, err := provider.Send(msgs)
		if err != nil {
			fmt.Printf("❌ AI generation error: %v\n", err)
			return
		}

		commitMsg := strings.TrimSpace(resp)
		// Clean up quotes or markdown if model added them
		commitMsg = strings.Trim(commitMsg, "`\"'")

		fmt.Println("\n=======================================================")
		fmt.Printf("✨ Suggested Commit Message:\n   %s\n", commitMsg)
		fmt.Println("=======================================================")

		fmt.Print("\nDo you want to apply this commit? [Y/n/e (edit)]: ")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		if input == "e" {
			fmt.Print("Enter custom commit message: ")
			customMsg, _ := reader.ReadString('\n')
			customMsg = strings.TrimSpace(customMsg)
			if customMsg != "" {
				commitMsg = customMsg
			}
			input = "y"
		}

		if input == "" || input == "y" {
			// Stage all changes if nothing staged
			if len(diffBytes) == 0 {
				exec.Command("git", "add", ".").Run()
			}
			cCmd := exec.Command("git", "commit", "-m", commitMsg)
			if out, err := cCmd.CombinedOutput(); err != nil {
				fmt.Printf("❌ Git commit failed: %s\n", string(out))
			} else {
				fmt.Println("✅ Successfully committed changes!")
				fmt.Println(string(out))
			}
		} else {
			fmt.Println("Cancelled commit.")
		}
	},
}
