package venv

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type EnvironmentType string

const (
	TypePythonVenv	EnvironmentType	= "Python venv"
	TypeAnaconda	EnvironmentType	= "Conda"
	TypeNodeModules	EnvironmentType	= "Node Modules"
	TypeUnknown	EnvironmentType	= "Unknown"
)

type Environment struct {
	Name	string
	Path	string
	Type	EnvironmentType
	Size	string
}

type Manager struct {
	Workspace	string
	PythonPath	string
}

func NewManager(workspace string) *Manager {
	if workspace == "" {
		workspace, _ = os.Getwd()
	}
	return &Manager{Workspace: filepath.Clean(workspace)}
}

func (m *Manager) CheckPrerequisites() error {
	candidates := []string{"python", "python3", "py"}

	for _, c := range candidates {
		path, err := exec.LookPath(c)
		if err == nil {
			m.PythonPath = path
			return nil
		}
	}

	return fmt.Errorf("python is not installed or not in PATH (tried python, python3, py)")
}

func (m *Manager) List() ([]Environment, error) {
	var envs []Environment

	home, _ := os.UserHomeDir()
	globalPaths := []string{
		filepath.Join(home, ".virtualenvs"),
		filepath.Join(home, "anaconda3", "envs"),
		filepath.Join(home, "miniconda3", "envs"),
		`C:\ProgramData\Anaconda3\envs`,
	}

	for _, gPath := range globalPaths {
		if _, err := os.Stat(gPath); err == nil {
			entries, _ := os.ReadDir(gPath)
			for _, e := range entries {
				if e.IsDir() {
					fullPath := filepath.Join(gPath, e.Name())
					if t := detectType(fullPath); t != TypeUnknown {
						envs = append(envs, Environment{
							Name:	fmt.Sprintf("Global: %s", e.Name()),
							Path:	fullPath,
							Type:	t,
							Size:	getSize(fullPath),
						})
					}
				}
			}
		}
	}

	workspace := filepath.Clean(m.Workspace)
	baseDepth := strings.Count(workspace, string(os.PathSeparator))

	shouldSkip := func(path string) bool {
		base := filepath.Base(path)
		if strings.HasPrefix(base, ".") {
			return base != ".venv"
		}
		if base == "__pycache__" || base == "vendor" {
			return true
		}
		return false
	}

	filepath.Walk(workspace, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			return nil
		}

		currentDepth := strings.Count(path, string(os.PathSeparator))
		if currentDepth-baseDepth > 3 {
			return filepath.SkipDir
		}

		if shouldSkip(path) && path != workspace {
			if detectType(path) == TypeUnknown {
				return filepath.SkipDir
			}
		}
		if t := detectType(path); t != TypeUnknown {
			name := path
			if rel, err := filepath.Rel(workspace, path); err == nil {
				name = rel
			}
			if name == "." {
				name = filepath.Base(path)
			}

			envs = append(envs, Environment{
				Name:	name,
				Path:	path,
				Type:	t,
				Size:	getSize(path),
			})
			return filepath.SkipDir
		}

		return nil
	})

	return envs, nil
}

func (m *Manager) CreateVenv(projectPath string) error {
	if m.PythonPath == "" {
		if err := m.CheckPrerequisites(); err != nil {
			return err
		}
	}

	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	parentDir := filepath.Dir(absPath)

	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create parent dir: %w", err)
	}

	cmd := exec.Command(m.PythonPath, "-m", "venv", absPath)

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("venv creation failed: %s: %w", string(out), err)
	}

	return nil
}

func (m *Manager) Verify(path string) error {
	if t := detectType(path); t != TypePythonVenv {
		return fmt.Errorf("verification failed: venv folder not detected or invalid")
	}
	return nil
}

func (m *Manager) Delete(path string) error {
	return os.RemoveAll(path)
}

func (m *Manager) Clone(srcPath, destPath string) error {

	t := detectType(srcPath)
	if t != TypePythonVenv {
		return fmt.Errorf("only python venv cloning is supported currently")
	}

	if err := m.CreateVenv(destPath); err != nil {
		return err
	}

	pip, err := findPip(srcPath)
	if err != nil {
		return fmt.Errorf("could not find pip in source: %w", err)
	}

	out, err := exec.Command(pip, "freeze").Output()
	if err != nil {
		return fmt.Errorf("failed to get packages from source: %w", err)
	}

	destPip, err := findPip(destPath)
	if err != nil {
		return fmt.Errorf("could not find pip in destination: %w", err)
	}

	reqFile := filepath.Join(destPath, "requirements.temp.txt")
	os.WriteFile(reqFile, out, 0644)
	defer os.Remove(reqFile)

	install := exec.Command(destPip, "install", "-r", reqFile)
	if out, err := install.CombinedOutput(); err != nil {
		return fmt.Errorf("cloning install failed: %s: %w", string(out), err)
	}
	return nil
}

func (m *Manager) Sync(venvPath string, destPath string) error {

	pip, err := findPip(venvPath)
	if err != nil {
		return err
	}

	out, err := exec.Command(pip, "freeze").Output()
	if err != nil {
		return fmt.Errorf("freeze failed: %w", err)
	}

	return os.WriteFile(destPath, out, 0644)
}

func findPip(venvPath string) (string, error) {
	candidates := []string{
		filepath.Join(venvPath, "Scripts", "pip.exe"),
		filepath.Join(venvPath, "Scripts", "pip"),
		filepath.Join(venvPath, "bin", "pip"),
		filepath.Join(venvPath, "bin", "pip3"),
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, nil
		}
	}
	return "", fmt.Errorf("pip executable not found in %s", venvPath)
}

func detectType(path string) EnvironmentType {

	if _, err := os.Stat(filepath.Join(path, "pyvenv.cfg")); err == nil {
		return TypePythonVenv
	}

	if _, err := os.Stat(filepath.Join(path, "Scripts", "activate")); err == nil {
		return TypePythonVenv
	}

	if _, err := os.Stat(filepath.Join(path, "bin", "activate")); err == nil {
		return TypePythonVenv
	}

	if filepath.Base(path) == "node_modules" {
		return TypeNodeModules
	}

	return TypeUnknown
}

func getSize(path string) string {
	var size int64
	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	mb := float64(size) / 1024 / 1024
	if mb < 1024 {
		return fmt.Sprintf("%.1f MB", mb)
	}
	return fmt.Sprintf("%.1f GB", mb/1024)
}
