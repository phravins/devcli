package boilerplate

import (
	"fmt"
	"os"
	"path/filepath"
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

func (m *Manager) GenerateSnippet(name, language string, destDir string) (string, error) {
	snippet, ok := Snippets[name]
	if !ok {
		return "", fmt.Errorf("snippet '%s' not found", name)
	}

	content, ok := snippet.Content[language]
	if !ok {

		return "", fmt.Errorf("language '%s' not supported for snippet '%s'", language, name)
	}

	fileName := snippet.DefaultFile
	if fileName == "" {
		fileName = "snippet"
	} else {

		ext := filepath.Ext(fileName)
		if ext != "" {
			fileName = fileName[:len(fileName)-len(ext)]
		}
	}

	fileName += getExt(language)

	fullPath := filepath.Join(destDir, fileName)
	return fullPath, os.WriteFile(fullPath, []byte(content), 0644)
}

func getExt(lang string) string {
	switch lang {
	case "Go":
		return ".go"
	case "Python":
		return ".py"
	case "JavaScript":
		return ".js"
	case "TypeScript":
		return ".ts"
	case "Java":
		return ".java"
	case "C":
		return ".c"
	case "C++":
		return ".cpp"
	case "Rust":
		return ".rs"
	case "HTML":
		return ".html"
	case "CSS":
		return ".css"
	case "React":
		return ".jsx"
	case "Node.js":
		return ".js"
	case "HTML + Tailwind":
		return ".html"
	case "React + Tailwind":
		return ".jsx"
	default:
		return ".txt"
	}
}
