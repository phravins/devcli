package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/phravins/devcli/internal/config"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("LoadConfig err: %v\n", err)
		return
	}
	fmt.Printf("Loaded AIBackend: %q\n", cfg.AIBackend)
	config.Set("ai_backend", config.CleanString(cfg.AIBackend))
	config.Set("ai_model", config.CleanString(cfg.AIModel))
	config.Set("user_name", config.CleanString(cfg.UserName))
	if err := config.Write(); err != nil {
		fmt.Printf("Write err: %v\n", err)
	} else {
		fmt.Println("Successfully cleaned ~/.devcli.yaml configuration file!")
	}
	home, _ := os.UserHomeDir()
	content, _ := os.ReadFile(filepath.Join(home, ".devcli.yaml"))
	fmt.Printf("File content:\n%s\n", string(content))
}
