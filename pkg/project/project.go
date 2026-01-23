package project

import (
	"fmt"
	"os"
	"path/filepath"
)

// CreateStructure creates a new project folder with subdirectories and basic files
func CreateStructure(name string) error {
	if err := os.MkdirAll(name, 0755); err != nil {
		return fmt.Errorf("failed to create project dir: %w", err)
	}

	dirs := []string{"src", "tests", "docs", "config"}
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(name, d), 0755); err != nil {
			return fmt.Errorf("failed to create %s: %w", d, err)
		}
	}

	// README.md
	readme := fmt.Sprintf("# %s\n\nProject description here.\n\n## Getting Started\n\n```bash\ncd %s\nnpm install\nnpm run dev\n```", name, name)
	if err := os.WriteFile(filepath.Join(name, "README.md"), []byte(readme), 0644); err != nil {
		return err
	}

	// package.json stub
	pkg := fmt.Sprintf(`{
  "name": "%s",
  "version": "1.0.0",
  "description": "",
  "main": "src/index.js",
  "scripts": {
    "dev": "node src/index.js"
  },
  "keywords": [],
  "author": "Phravins",
  "license": "Apache License 2.0"
}`, name)
	return os.WriteFile(filepath.Join(name, "package.json"), []byte(pkg), 0644)
}
