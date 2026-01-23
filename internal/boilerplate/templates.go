package boilerplate

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// TemplateManager handles saving/restoring custom user templates
type TemplateManager struct {
	TemplatesDir string
}

func NewTemplateManager(baseDir string) *TemplateManager {
	return &TemplateManager{
		TemplatesDir: filepath.Join(baseDir, ".devcli", "custom_templates"),
	}
}

// CustomTemplate represents a saved user template
type CustomTemplate struct {
	Name      string
	Path      string
	CreatedAt time.Time
}

// SaveAsTemplate saves the sourceDir as a new template
func (tm *TemplateManager) SaveAsTemplate(name, sourceDir string) error {
	destDir := filepath.Join(tm.TemplatesDir, name)
	if _, err := os.Stat(destDir); err == nil {
		return fmt.Errorf("template '%s' already exists", name)
	}

	return copyDir(sourceDir, destDir)
}

// ListTemplates returns all custom saved templates
func (tm *TemplateManager) ListTemplates() ([]CustomTemplate, error) {
	if _, err := os.Stat(tm.TemplatesDir); os.IsNotExist(err) {
		return []CustomTemplate{}, nil // No templates yet
	}

	entries, err := os.ReadDir(tm.TemplatesDir)
	if err != nil {
		return nil, err
	}

	var templates []CustomTemplate
	for _, e := range entries {
		if e.IsDir() {
			info, _ := e.Info()
			templates = append(templates, CustomTemplate{
				Name:      e.Name(),
				Path:      filepath.Join(tm.TemplatesDir, e.Name()),
				CreatedAt: info.ModTime(),
			})
		}
	}
	return templates, nil
}

// LoadTemplate restores a template to the destDir
func (tm *TemplateManager) LoadTemplate(templateName, destDir string) error {
	srcPath := filepath.Join(tm.TemplatesDir, templateName)
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return fmt.Errorf("template '%s' does not exist", templateName)
	}
	return copyDir(srcPath, destDir)
}

// DeleteTemplate permanently removes a custom template
func (tm *TemplateManager) DeleteTemplate(templateName string) error {
	path := filepath.Join(tm.TemplatesDir, templateName)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("template '%s' not found", templateName)
	}
	return os.RemoveAll(path)
}

// Helper to copy simple directories recursively
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(src, path)
		destPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			// Skip things like .git, node_modules, etc if we want to be smart?
			// For now, raw copy.
			if info.Name() == ".git" || info.Name() == "node_modules" || info.Name() == "venv" || info.Name() == ".venv" || info.Name() == "__pycache__" || info.Name() == "bin" || info.Name() == "obj" || info.Name() == ".idea" || info.Name() == ".vscode" {
				return filepath.SkipDir
			}
			return os.MkdirAll(destPath, info.Mode())
		}

		sourceFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer sourceFile.Close()

		destFile, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer destFile.Close()

		_, err = io.Copy(destFile, sourceFile)
		return err
	})
}
