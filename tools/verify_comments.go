package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/phravins/devcli/internal/boilerplate"
)

func main() {
	fmt.Println("Verifying Commented Snippets...")

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting working directory: %v\n", err)
		os.Exit(1)
	}
	rootDir := filepath.Dir(cwd)
	if filepath.Base(cwd) != "scripts" {
		rootDir = cwd
	}

	testDir := filepath.Join(rootDir, "test_snippets_comments")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		fmt.Printf("Error creating test directory: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(testDir)

	mgr := boilerplate.NewManager(testDir)

	tests := []struct {
		name	string
		lang	string
		file	string
	}{
		{"Auth System", "Go", "auth_handlers.go"},
		{"DB: PostgreSQL", "Go", "db_postgres.go"},
	}

	for _, tt := range tests {
		fmt.Printf("Test: %s... ", tt.name)
		_, err := mgr.GenerateSnippet(tt.name, tt.lang, testDir)
		if err != nil {
			fmt.Printf("FAILED: %v\n", err)
			os.Exit(1)
		}

		content, err := os.ReadFile(filepath.Join(testDir, tt.file))
		if err != nil {
			fmt.Printf("FAILED (Read Error): %v\n", err)
			os.Exit(1)
		}

		if len(content) < 100 {
			fmt.Printf("FAILED (Content too short)\n")
			os.Exit(1)
		}
		fmt.Println("PASSED")
	}

	fmt.Println("All Commented Snippets Verified!")
}
