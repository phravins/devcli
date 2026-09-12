package boilerplate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewTemplateManager(t *testing.T) {
	baseDir := t.TempDir()
	tm := NewTemplateManager(baseDir)
	expected := filepath.Join(baseDir, ".devcli", "custom_templates")
	if tm.TemplatesDir != expected {
		t.Errorf("expected TemplatesDir %q, got %q", expected, tm.TemplatesDir)
	}
}

func TestLoadTemplate_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	tm := NewTemplateManager(tempDir)

	destDir := filepath.Join(tempDir, "dest")
	err := tm.LoadTemplate("nonexistent-template", destDir)
	if err == nil {
		t.Fatal("expected error when loading nonexistent template, got nil")
	}

	expectedSubstring := "template 'nonexistent-template' does not exist"
	if !strings.Contains(err.Error(), expectedSubstring) {
		t.Errorf("expected error containing %q, got %q", expectedSubstring, err.Error())
	}
}

func TestLoadTemplate_SuccessAndIgnoredDirs(t *testing.T) {
	baseDir := t.TempDir()
	tm := NewTemplateManager(baseDir)

	templateName := "test-template"
	templateDir := filepath.Join(tm.TemplatesDir, templateName)

	// Create source template structure
	if err := os.MkdirAll(filepath.Join(templateDir, "subdir"), 0755); err != nil {
		t.Fatalf("failed to create template dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(templateDir, ".git"), 0755); err != nil {
		t.Fatalf("failed to create .git dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(templateDir, "node_modules"), 0755); err != nil {
		t.Fatalf("failed to create node_modules dir: %v", err)
	}

	// Create regular files
	file1Path := filepath.Join(templateDir, "file1.txt")
	if err := os.WriteFile(file1Path, []byte("hello world"), 0644); err != nil {
		t.Fatalf("failed to write file1: %v", err)
	}

	file2Path := filepath.Join(templateDir, "subdir", "file2.txt")
	if err := os.WriteFile(file2Path, []byte("nested content"), 0644); err != nil {
		t.Fatalf("failed to write file2: %v", err)
	}

	// Create files inside ignored dirs
	if err := os.WriteFile(filepath.Join(templateDir, ".git", "config"), []byte("git config"), 0644); err != nil {
		t.Fatalf("failed to write git config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(templateDir, "node_modules", "package.json"), []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to write package.json: %v", err)
	}

	destDir := filepath.Join(baseDir, "dest")
	err := tm.LoadTemplate(templateName, destDir)
	if err != nil {
		t.Fatalf("unexpected error loading template: %v", err)
	}

	// Verify regular files were restored
	content1, err := os.ReadFile(filepath.Join(destDir, "file1.txt"))
	if err != nil {
		t.Errorf("failed to read restored file1.txt: %v", err)
	} else if string(content1) != "hello world" {
		t.Errorf("expected content 'hello world', got %q", string(content1))
	}

	content2, err := os.ReadFile(filepath.Join(destDir, "subdir", "file2.txt"))
	if err != nil {
		t.Errorf("failed to read restored file2.txt: %v", err)
	} else if string(content2) != "nested content" {
		t.Errorf("expected content 'nested content', got %q", string(content2))
	}

	// Verify ignored directories were skipped
	if _, err := os.Stat(filepath.Join(destDir, ".git")); !os.IsNotExist(err) {
		t.Errorf(".git directory should have been skipped during load")
	}
	if _, err := os.Stat(filepath.Join(destDir, "node_modules")); !os.IsNotExist(err) {
		t.Errorf("node_modules directory should have been skipped during load")
	}
}

func TestSaveAsTemplate(t *testing.T) {
	tempDir := t.TempDir()
	tm := NewTemplateManager(tempDir)

	sourceDir := filepath.Join(tempDir, "source")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "app.go"), []byte("package main"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	// Save as template
	err := tm.SaveAsTemplate("my-app", sourceDir)
	if err != nil {
		t.Fatalf("unexpected error saving template: %v", err)
	}

	// Verify template created
	savedFile := filepath.Join(tm.TemplatesDir, "my-app", "app.go")
	if content, err := os.ReadFile(savedFile); err != nil || string(content) != "package main" {
		t.Errorf("expected file in template with 'package main', got err: %v, content: %q", err, string(content))
	}

	// Saving again with same name should return error
	err = tm.SaveAsTemplate("my-app", sourceDir)
	if err == nil {
		t.Fatal("expected error when saving template with existing name, got nil")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected error containing 'already exists', got %q", err.Error())
	}
}

func TestListTemplates(t *testing.T) {
	tempDir := t.TempDir()
	tm := NewTemplateManager(tempDir)

	// Case 1: TemplatesDir does not exist yet
	templates, err := tm.ListTemplates()
	if err != nil {
		t.Fatalf("unexpected error listing non-existent templates dir: %v", err)
	}
	if len(templates) != 0 {
		t.Errorf("expected 0 templates, got %d", len(templates))
	}

	// Case 2: TemplatesDir exists with directories and a file
	if err := os.MkdirAll(tm.TemplatesDir, 0755); err != nil {
		t.Fatalf("failed to create templates dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tm.TemplatesDir, "tpl1"), 0755); err != nil {
		t.Fatalf("failed to create tpl1: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tm.TemplatesDir, "tpl2"), 0755); err != nil {
		t.Fatalf("failed to create tpl2: %v", err)
	}
	// Add a regular file in TemplatesDir (should be ignored by ListTemplates)
	if err := os.WriteFile(filepath.Join(tm.TemplatesDir, "ignore.txt"), []byte("ignore"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	templates, err = tm.ListTemplates()
	if err != nil {
		t.Fatalf("unexpected error listing templates: %v", err)
	}
	if len(templates) != 2 {
		t.Fatalf("expected 2 templates, got %d", len(templates))
	}

	names := map[string]bool{templates[0].Name: true, templates[1].Name: true}
	if !names["tpl1"] || !names["tpl2"] {
		t.Errorf("expected tpl1 and tpl2 in list, got %v", templates)
	}
}

func TestDeleteTemplate(t *testing.T) {
	tempDir := t.TempDir()
	tm := NewTemplateManager(tempDir)

	// Case 1: Delete non-existent template
	err := tm.DeleteTemplate("nonexistent")
	if err == nil {
		t.Fatal("expected error deleting nonexistent template, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected error containing 'not found', got %q", err.Error())
	}

	// Case 2: Delete existing template
	tplDir := filepath.Join(tm.TemplatesDir, "to-delete")
	if err := os.MkdirAll(tplDir, 0755); err != nil {
		t.Fatalf("failed to create template dir: %v", err)
	}

	err = tm.DeleteTemplate("to-delete")
	if err != nil {
		t.Fatalf("unexpected error deleting template: %v", err)
	}

	if _, err := os.Stat(tplDir); !os.IsNotExist(err) {
		t.Errorf("expected template directory to be deleted")
	}
}
