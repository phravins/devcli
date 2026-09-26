package project

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Manager struct {
	Workspace string
}

func NewManager(workspace string) *Manager {
	if workspace == "" {
		workspace, _ = os.Getwd()
	}
	return &Manager{Workspace: workspace}
}

func (m *Manager) CreateProject(name, stack, parentDir string) (string, string, error) {
	if parentDir == "" {
		parentDir = m.Workspace
	}

	parentDir = m.ExpandPath(parentDir)

	cfg := ProjectConfig{
		Name:		name,
		Path:		filepath.Join(parentDir, name),
		Stack:		stack,
		InitGit:	true,
		AddReadme:	true,
	}

	fmt.Printf("Generating project at: %s\n", cfg.Path)
	cmd, err := Generate(cfg)
	return cmd, cfg.Path, err
}

func (m *Manager) ValidateParentDir(path string) (string, error) {
	expanded := m.ExpandPath(path)
	info, err := os.Stat(expanded)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("directory does not exist: %s", expanded)
	}
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("path is not a directory: %s", expanded)
	}
	return expanded, nil
}

func (m *Manager) ExpandPath(path string) string {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			if path == "~" {
				return home
			}
			if strings.HasPrefix(path, "~/") || strings.HasPrefix(path, "~\\") {
				return filepath.Join(home, path[2:])
			}
		}
	}
	return os.ExpandEnv(path)
}

func (m *Manager) SuggestProjectName(templateName string) string {

	base := strings.ToLower(templateName)
	base = strings.ReplaceAll(base, " ", "-")
	base = strings.ReplaceAll(base, "api", "project")

	if strings.Contains(base, "go") {
		base = "go-project"
	}
	if strings.Contains(base, "python") || strings.Contains(base, "fastapi") {
		base = "fastapi-project"
	}
	if strings.Contains(base, "node") || strings.Contains(base, "react") {
		base = "node-project"
	}
	if strings.Contains(base, "java") {
		base = "java-project"
	}
	if strings.Contains(base, "kotlin") {
		base = "kotlin-project"
	}
	if strings.Contains(base, "dart") {
		base = "dart-project"
	}
	if strings.Contains(base, "c++") || strings.Contains(base, "cpp") {
		base = "cpp-project"
	}

	name := base
	counter := 1
	for {
		path := filepath.Join(m.Workspace, name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return name
		}
		name = fmt.Sprintf("%s-%02d", base, counter)
		counter++
	}
}

func (m *Manager) BackupProject(srcDir, destPath string) error {

	srcDir = m.ExpandPath(srcDir)
	destPath = m.ExpandPath(destPath)

	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return os.MkdirAll(destPath, info.Mode())
		}

		targetPath := filepath.Join(destPath, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		return copyFile(path, targetPath)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
