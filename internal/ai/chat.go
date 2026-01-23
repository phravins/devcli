package ai

import (
	"fmt"

	"github.com/spf13/cobra"
)

var AICmd = &cobra.Command{
	Use:   "ai",
	Short: "AI chatbot commands",
	Long:  "Local AI chatbot using Ollama, Llama.cpp, or GPT4All",
}

var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "List available AI models",
	Run: func(cmd *cobra.Command, args []string) {
		listModels()
	},
}

func init() {
	AICmd.AddCommand(modelsCmd)
}

func init() {
	AICmd.AddCommand(modelsCmd)
}
func listModels() {
	fmt.Println("Use 'devcli ai chat' or the dashboard to see/use models.")
}
