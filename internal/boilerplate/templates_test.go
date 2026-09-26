package boilerplate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewTemplateManager(t *testing.T) {
	tmpDir := t.TempDir()
	tm := NewTemplateManager(tmpDir)

	expected := filepath.Join(tmpDir, ".devcli", "custom_templates")
	if tm.TemplatesDir != expected {
		t.Errorf("expected TemplatesDir to be %s, got %s", expected, tm.TemplatesDir)
	}
}

func TestSaveAsTemplate(t *testing.T) {
	tmpDir := t.TempDir()
	tm := NewTemplateManager(tmpDir)

	sourceDir := filepath.Join(tmpDir, "my_source_project")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}

	file1 := filepath.Join(sourceDir, "file1.txt")
	if err := os.WriteFile(file1, []byte("hello world"), 0644); err != nil {
		t.Fatalf("failed to write file1: %v", err)
	}

	subDir := filepath.Join(sourceDir, "src")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create subDir: %v", err)
	}
	file2 := filepath.Join(subDir, "main.go")
	if err := os.WriteFile(file2, []byte("package main"), 0644); err != nil {
		t.Fatalf("failed to write file2: %v", err)
	}

	gitDir := filepath.Join(sourceDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(gitDir, "config"), []byte("git config"), 0644); err != nil {
		t.Fatalf("failed to write git config: %v", err)
	}

	nodeModulesDir := filepath.Join(sourceDir, "node_modules")
	if err := os.MkdirAll(nodeModulesDir, 0755); err != nil {
		t.Fatalf("failed to create node_modules dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nodeModulesDir, "package.json"), []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to write package.json in node_modules: %v", err)
	}

	templateName := "my-template"
	err := tm.SaveAsTemplate(templateName, sourceDir)
	if err != nil {
		t.Fatalf("SaveAsTemplate failed: %v", err)
	}

	destDir := filepath.Join(tm.TemplatesDir, templateName)

	content1, err := os.ReadFile(filepath.Join(destDir, "file1.txt"))
	if err != nil || string(content1) != "hello world" {
		t.Errorf("file1.txt content mismatch or missing: %v", err)
	}

	content2, err := os.ReadFile(filepath.Join(destDir, "src", "main.go"))
	if err != nil || string(content2) != "package main" {
		t.Errorf("src/main.go content mismatch or missing: %v", err)
	}

	if _, err := os.Stat(filepath.Join(destDir, ".git")); !os.IsNotExist(err) {
		t.Errorf("expected .git directory to be skipped in template copy")
	}
	if _, err := os.Stat(filepath.Join(destDir, "node_modules")); !os.IsNotExist(err) {
		t.Errorf("expected node_modules directory to be skipped in template copy")
	}

	err = tm.SaveAsTemplate(templateName, sourceDir)
	if err == nil {
		t.Errorf("expected error when saving template that already exists, got nil")
	}

	nonExistentSource := filepath.Join(tmpDir, "non_existent")
	err = tm.SaveAsTemplate("another-template", nonExistentSource)
	if err == nil {
		t.Errorf("expected error when sourceDir does not exist, got nil")
	}
}

func TestListTemplates(t *testing.T) {
	tmpDir := t.TempDir()
	tm := NewTemplateManager(tmpDir)

	templates, err := tm.ListTemplates()
	if err != nil {
		t.Fatalf("ListTemplates failed on non-existent TemplatesDir: %v", err)
	}
	if len(templates) != 0 {
		t.Errorf("expected 0 templates, got %d", len(templates))
	}

	sourceDir := filepath.Join(tmpDir, "src_proj")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}
	if err := tm.SaveAsTemplate("tpl-1", sourceDir); err != nil {
		t.Fatalf("SaveAsTemplate tpl-1 failed: %v", err)
	}
	if err := tm.SaveAsTemplate("tpl-2", sourceDir); err != nil {
		t.Fatalf("SaveAsTemplate tpl-2 failed: %v", err)
	}

	templates, err = tm.ListTemplates()
	if err != nil {
		t.Fatalf("ListTemplates failed: %v", err)
	}
	if len(templates) != 2 {
		t.Errorf("expected 2 templates, got %d", len(templates))
	}

	found := make(map[string]bool)
	for _, tpl := range templates {
		found[tpl.Name] = true
		if tpl.Path != filepath.Join(tm.TemplatesDir, tpl.Name) {
			t.Errorf("template path mismatch: got %s, expected %s", tpl.Path, filepath.Join(tm.TemplatesDir, tpl.Name))
		}
		if tpl.CreatedAt.IsZero() {
			t.Errorf("expected non-zero CreatedAt for template %s", tpl.Name)
		}
	}

	if !found["tpl-1"] || !found["tpl-2"] {
		t.Errorf("expected templates tpl-1 and tpl-2 to be found, got: %v", found)
	}
}

func TestLoadTemplate(t *testing.T) {
	tmpDir := t.TempDir()
	tm := NewTemplateManager(tmpDir)

	sourceDir := filepath.Join(tmpDir, "src_proj")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "README.md"), []byte("# My Project"), 0644); err != nil {
		t.Fatalf("failed to write README: %v", err)
	}

	if err := tm.SaveAsTemplate("tpl-load", sourceDir); err != nil {
		t.Fatalf("SaveAsTemplate failed: %v", err)
	}

	destDir := filepath.Join(tmpDir, "restored_proj")
	if err := tm.LoadTemplate("tpl-load", destDir); err != nil {
		t.Fatalf("LoadTemplate failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(destDir, "README.md"))
	if err != nil || string(content) != "# My Project" {
		t.Errorf("restored file content mismatch or missing: %v", err)
	}

	err = tm.LoadTemplate("non-existent-tpl", destDir)
	if err == nil {
		t.Errorf("expected error loading non-existent template, got nil")
	}
}

func TestDeleteTemplate(t *testing.T) {
	tmpDir := t.TempDir()
	tm := NewTemplateManager(tmpDir)

	sourceDir := filepath.Join(tmpDir, "src_proj")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}
	if err := tm.SaveAsTemplate("tpl-del", sourceDir); err != nil {
		t.Fatalf("SaveAsTemplate failed: %v", err)
	}

	if err := tm.DeleteTemplate("tpl-del"); err != nil {
		t.Fatalf("DeleteTemplate failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tm.TemplatesDir, "tpl-del")); !os.IsNotExist(err) {
		t.Errorf("expected template directory to be deleted")
	}

	if err := tm.DeleteTemplate("tpl-del"); err == nil {
		t.Errorf("expected error deleting non-existent template, got nil")
	}
}
