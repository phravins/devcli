package project

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/phravins/devcli/internal/templates"
)

type ProjectConfig struct {
	Name		string
	Path		string
	Stack		string
	InitGit		bool
	AddReadme	bool
}

func Generate(cfg ProjectConfig) (string, error) {

	var selectedTpl templates.Template
	found := false
	for _, t := range templates.Registry {
		if t.Name == cfg.Stack {
			selectedTpl = t
			found = true
			break
		}
	}

	if !found {

		if cfg.Stack == "Go" {
			selectedTpl = templates.Registry[0]
			found = true
		}
		if cfg.Stack == "Python" {
			selectedTpl = templates.Registry[1]
			found = true
		}
		if cfg.Stack == "Node" {
			selectedTpl = templates.Registry[2]
			found = true
		}
	}

	targetDir := cfg.Path
	if targetDir == "" {
		targetDir = cfg.Name
	}
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	if found {
		for filename, content := range selectedTpl.Files {

			tmpl, err := template.New(filename).Parse(content)
			if err != nil {
				return "", err
			}
			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, cfg); err != nil {
				return "", err
			}

			fullPath := filepath.Join(targetDir, filename)

			if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
				return "", err
			}
			if err := os.WriteFile(fullPath, buf.Bytes(), 0644); err != nil {
				return "", err
			}
		}
	}

	if cfg.InitGit {
		initGit(targetDir)
	}

	if cfg.AddReadme && found {
		createReadme(targetDir, cfg, selectedTpl)
	}

	if found && selectedTpl.InstallCmd != "" {
		return selectedTpl.InstallCmd, nil
	}

	return "", nil
}

func initGit(dir string) {
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		fmt.Printf("Warning: git init failed: %v\n", err)
	}
}

func createReadme(dir string, cfg ProjectConfig, tpl templates.Template) {
	content := fmt.Sprintf(`# %s

![Build Status](https://img.shields.io/badge/build-passing-brightgreen)
![Version](https://img.shields.io/badge/version-0.1.0-blue)
![Stack](https://img.shields.io/badge/stack-%s-orange)

> %s

## Getting Started

### Prerequisites
- %s environment

### Installation
Run the following command to install dependencies:

`, cfg.Name, tpl.Stack, tpl.Description, tpl.Stack)

	content += fmt.Sprintf("```bash\n%s\n```\n\n", tpl.InstallCmd)
	content += "### Usage\nTo start the project:\n\n"
	content += fmt.Sprintf("```bash\n%s\n```\n", tpl.RunCmd)

	os.WriteFile(filepath.Join(dir, "README.md"), []byte(content), 0644)
}
