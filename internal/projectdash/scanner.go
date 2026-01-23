package projectdash

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ProjectStatus int

const (
	StatusActive ProjectStatus = iota
	StatusBroken
	StatusArchived
)

func (s ProjectStatus) String() string {
	switch s {
	case StatusActive:
		return "Active"
	case StatusBroken:
		return "Broken"
	case StatusArchived:
		return "Archived"
	default:
		return "Unknown"
	}
}

func (s ProjectStatus) Icon() string {
	switch s {
	case StatusActive:
		return ""
	case StatusBroken:
		return ""
	case StatusArchived:
		return ""
	default:
		return ""
	}
}

type ProjectInfo struct {
	Name         string
	Path         string
	TechStack    []string
	Status       ProjectStatus
	LastModified time.Time
	HasErrors    bool
	Size         int64 // in bytes
}

// DetectTechStack analyzes a project directory and identifies technologies used
func DetectTechStack(projectPath string) []string {
	var stack []string
	seen := make(map[string]bool)

	// Helper to add unique items
	add := func(tech string) {
		if !seen[tech] {
			stack = append(stack, tech)
			seen[tech] = true
		}
	}

	// Check for project markers
	markers := map[string][]string{
		"go.mod":             {"Go"},
		"package.json":       {"Node.js"},
		"requirements.txt":   {"Python"},
		"Pipfile":            {"Python", "Pipenv"},
		"poetry.lock":        {"Python", "Poetry"},
		"Cargo.toml":         {"Rust"},
		"pom.xml":            {"Java", "Maven"},
		"build.gradle":       {"Java", "Gradle"},
		"Gemfile":            {"Ruby"},
		"composer.json":      {"PHP"},
		"Dockerfile":         {"Docker"},
		"docker-compose.yml": {"Docker Compose"},
		".git":               {"Git"},
	}

	for marker, techs := range markers {
		if _, err := os.Stat(filepath.Join(projectPath, marker)); err == nil {
			for _, tech := range techs {
				add(tech)
			}
		}
	}

	// Check package.json for frameworks
	pkgPath := filepath.Join(projectPath, "package.json")
	if data, err := os.ReadFile(pkgPath); err == nil {
		content := string(data)
		frameworks := map[string]string{
			"react":      "React",
			"vue":        "Vue",
			"angular":    "Angular",
			"next":       "Next.js",
			"express":    "Express",
			"fastify":    "Fastify",
			"typescript": "TypeScript",
			"vite":       "Vite",
		}
		for pkg, name := range frameworks {
			if strings.Contains(content, `"`+pkg+`"`) {
				add(name)
			}
		}
	}

	// Check requirements.txt for frameworks
	reqPath := filepath.Join(projectPath, "requirements.txt")
	if data, err := os.ReadFile(reqPath); err == nil {
		content := strings.ToLower(string(data))
		frameworks := map[string]string{
			"django":     "Django",
			"flask":      "Flask",
			"fastapi":    "FastAPI",
			"pytest":     "Pytest",
			"numpy":      "NumPy",
			"pandas":     "Pandas",
			"pytorch":    "PyTorch",
			"tensorflow": "TensorFlow",
		}
		for pkg, name := range frameworks {
			if strings.Contains(content, pkg) {
				add(name)
			}
		}
	}

	// Check for Kotlin
	if hasExtension(projectPath, ".kt") {
		add("Kotlin")
	}

	// Check for C/C++
	if hasExtension(projectPath, ".c") {
		add("C")
	}
	if hasExtension(projectPath, ".cpp") || hasExtension(projectPath, ".cxx") {
		add("C++")
	}

	// Check for C#
	if hasExtension(projectPath, ".cs") {
		add("C#")
	}

	// Check for Zig
	if hasExtension(projectPath, ".zig") {
		add("Zig")
	}

	return stack
}

// hasExtension checks if directory contains files with given extension
func hasExtension(dir, ext string) bool {
	found := false
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || found {
			return filepath.SkipDir
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ext) {
			found = true
			return filepath.SkipDir
		}
		return nil
	})
	return found
}

// AnalyzeProjectStatus determines if a project is Active, Broken, or Archived
func AnalyzeProjectStatus(projectPath string) ProjectStatus {
	info, err := os.Stat(projectPath)
	if err != nil {
		return StatusBroken
	}

	// Check if archived (not modified in 90+ days)
	if time.Since(info.ModTime()) > 90*24*time.Hour {
		return StatusArchived
	}

	// Check for common error indicators
	errorMarkers := []string{
		"node_modules/.package-lock.json", // Broken npm install
		"__pycache__",
	}

	for _, marker := range errorMarkers {
		if _, err := os.Stat(filepath.Join(projectPath, marker)); err == nil {
			// Just presence doesn't mean broken
		}
	}

	return StatusActive
}

// GetProjectLastModified returns the most recent modification time
func GetProjectLastModified(projectPath string) time.Time {
	var latest time.Time

	// Check common files that indicate recent activity
	activityFiles := []string{
		"package.json",
		"go.mod",
		"requirements.txt",
		"main.go",
		"main.py",
		"index.js",
		"README.md",
	}

	for _, file := range activityFiles {
		path := filepath.Join(projectPath, file)
		if info, err := os.Stat(path); err == nil {
			if info.ModTime().After(latest) {
				latest = info.ModTime()
			}
		}
	}

	// Fallback to directory modification time
	if latest.IsZero() {
		if info, err := os.Stat(projectPath); err == nil {
			latest = info.ModTime()
		}
	}

	return latest
}

// GetProjectSize calculates the total size of a project directory
func GetProjectSize(projectPath string) int64 {
	var size int64
	filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		// Skip certain directories to speed up calculation
		if info.IsDir() {
			name := info.Name()
			if name == "node_modules" || name == ".git" || name == "__pycache__" || name == "venv" || name == ".venv" {
				return filepath.SkipDir
			}
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}

// ScanWorkspace scans a workspace directory for projects
func ScanWorkspace(workspacePath string) ([]ProjectInfo, error) {
	var projects []ProjectInfo

	entries, err := os.ReadDir(workspacePath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		projectPath := filepath.Join(workspacePath, entry.Name())

		// Check if it's actually a project
		if !isProject(projectPath) {
			continue
		}

		info := ProjectInfo{
			Name:         entry.Name(),
			Path:         projectPath,
			TechStack:    DetectTechStack(projectPath),
			Status:       AnalyzeProjectStatus(projectPath),
			LastModified: GetProjectLastModified(projectPath),
			Size:         GetProjectSize(projectPath),
		}

		projects = append(projects, info)
	}

	return projects, nil
}

// isProject checks if a directory is a project
func isProject(dir string) bool {
	markers := []string{
		"go.mod",
		"package.json",
		"requirements.txt",
		".git",
		"main.py",
		"main.go",
		"index.js",
		"Cargo.toml",
		"pom.xml",
	}

	for _, marker := range markers {
		if _, err := os.Stat(filepath.Join(dir, marker)); err == nil {
			return true
		}
	}
	return false
}

// FormatSize formats bytes into human-readable string
func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
