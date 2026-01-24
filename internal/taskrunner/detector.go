package taskrunner

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type TaskType string

const (
	TaskBuild  TaskType = "build"
	TaskTest   TaskType = "test"
	TaskFormat TaskType = "format"
	TaskLint   TaskType = "lint"
	TaskRun    TaskType = "run"
	TaskClean  TaskType = "clean"
)

type Task struct {
	Name        string
	Type        TaskType
	Command     string
	Description string
	Icon        string
}

// DetectTasks scans a project directory and detects available tasks
func DetectTasks(projectPath string) []Task {
	var tasks []Task

	// Check for Node.js/npm tasks
	if pkgJSON := filepath.Join(projectPath, "package.json"); fileExists(pkgJSON) {
		tasks = append(tasks, detectNpmTasks(pkgJSON)...)
	}

	// Check for Python tasks
	if fileExists(filepath.Join(projectPath, "requirements.txt")) ||
		fileExists(filepath.Join(projectPath, "setup.py")) ||
		fileExists(filepath.Join(projectPath, "pyproject.toml")) ||
		hasFilesWithExtension(projectPath, ".py") {
		tasks = append(tasks, detectPythonTasks(projectPath)...)
	}

	// Check for Go tasks
	if fileExists(filepath.Join(projectPath, "go.mod")) {
		tasks = append(tasks, detectGoTasks(projectPath)...)
	}

	// Check for Makefile tasks
	if fileExists(filepath.Join(projectPath, "Makefile")) {
		tasks = append(tasks, detectMakefileTasks(filepath.Join(projectPath, "Makefile"))...)
	}

	// Check for Java tasks (Maven/Gradle/Plain)
	if fileExists(filepath.Join(projectPath, "pom.xml")) ||
		fileExists(filepath.Join(projectPath, "build.gradle")) ||
		fileExists(filepath.Join(projectPath, "build.gradle.kts")) ||
		hasFilesWithExtension(projectPath, ".java") {
		tasks = append(tasks, detectJavaTasks(projectPath)...)
	}

	// Check for C/C++ tasks
	if fileExists(filepath.Join(projectPath, "CMakeLists.txt")) ||
		hasFilesWithExtension(projectPath, ".c") ||
		hasFilesWithExtension(projectPath, ".cpp") ||
		hasFilesWithExtension(projectPath, ".cc") {
		tasks = append(tasks, detectCTasks(projectPath)...)
	}

	// Check for Rust tasks
	if fileExists(filepath.Join(projectPath, "Cargo.toml")) {
		tasks = append(tasks, detectRustTasks()...)
	}

	return tasks
}

func detectNpmTasks(pkgJSONPath string) []Task {
	var tasks []Task

	data, err := os.ReadFile(pkgJSONPath)
	if err != nil {
		return tasks
	}

	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}

	if err := json.Unmarshal(data, &pkg); err != nil {
		return tasks
	}

	// Map script names to task types
	typeMap := map[string]TaskType{
		"build":  TaskBuild,
		"test":   TaskTest,
		"lint":   TaskLint,
		"format": TaskFormat,
		"dev":    TaskRun,
		"start":  TaskRun,
		"clean":  TaskClean,
	}

	iconMap := map[TaskType]string{
		TaskBuild:  "",
		TaskTest:   "",
		TaskLint:   "",
		TaskFormat: "",
		TaskRun:    "",
		TaskClean:  "",
	}

	for scriptName, scriptCmd := range pkg.Scripts {
		taskType := TaskRun // default
		for key, t := range typeMap {
			if strings.Contains(scriptName, key) {
				taskType = t
				break
			}
		}

		tasks = append(tasks, Task{
			Name:        fmt.Sprintf("npm run %s", scriptName),
			Type:        taskType,
			Command:     fmt.Sprintf("npm run %s", scriptName),
			Description: scriptCmd,
			Icon:        iconMap[taskType],
		})
	}

	return tasks
}

func detectPythonTasks(projectPath string) []Task {
	tasks := []Task{
		{
			Name:        "Run Tests (pytest)",
			Type:        TaskTest,
			Command:     "pytest",
			Description: "Run Python tests with pytest",
			Icon:        "",
		},
		{
			Name:        "Format Code (black)",
			Type:        TaskFormat,
			Command:     "black .",
			Description: "Format Python code with Black",
			Icon:        "",
		},
		{
			Name:        "Lint Code (flake8)",
			Type:        TaskLint,
			Command:     "flake8 .",
			Description: "Lint Python code with flake8",
			Icon:        "",
		},
		{
			Name:        "Type Check (mypy)",
			Type:        TaskLint,
			Command:     "mypy .",
			Description: "Run static type checker",
			Icon:        "",
		},
	}

	if fileExists(filepath.Join(projectPath, "requirements.txt")) {
		tasks = append(tasks, Task{
			Name:        "Install Dependencies",
			Type:        TaskRun,
			Command:     "pip install -r requirements.txt",
			Description: "Install project requirements",
			Icon:        "",
		})
	}

	// Search for main-like files in root and src/
	possibleMains := []string{"main.py", "app.py", "src/main.py", "src/app.py"}
	for _, pm := range possibleMains {
		if fileExists(filepath.Join(projectPath, pm)) {
			icon := ""
			tasks = append(tasks, Task{
				Name:        fmt.Sprintf("Run %s", pm),
				Type:        TaskRun,
				Command:     fmt.Sprintf("python %s", pm),
				Description: fmt.Sprintf("Execute %s", pm),
				Icon:        icon,
			})
		}
	}

	return tasks
}

func detectJavaTasks(projectPath string) []Task {
	var tasks []Task

	if fileExists(filepath.Join(projectPath, "pom.xml")) {
		tasks = append(tasks, []Task{
			{
				Name:        "Maven: Build (install)",
				Type:        TaskBuild,
				Command:     "mvn install",
				Description: "Build and install Maven project",
				Icon:        "",
			},
			{
				Name:        "Maven: Test",
				Type:        TaskTest,
				Command:     "mvn test",
				Description: "Run Maven tests",
				Icon:        "",
			},
			{
				Name:        "Maven: Clean",
				Type:        TaskClean,
				Command:     "mvn clean",
				Description: "Clean Maven project",
				Icon:        "",
			},
		}...)
	}

	if fileExists(filepath.Join(projectPath, "build.gradle")) || fileExists(filepath.Join(projectPath, "build.gradle.kts")) {
		gradleCmd := "gradle"
		if fileExists(filepath.Join(projectPath, "gradlew")) {
			gradleCmd = "./gradlew"
		} else if fileExists(filepath.Join(projectPath, "gradlew.bat")) {
			gradleCmd = "gradlew.bat"
		}

		tasks = append(tasks, []Task{
			{
				Name:        "Gradle: Build (assemble)",
				Type:        TaskBuild,
				Command:     gradleCmd + " assemble",
				Description: "Build Gradle project",
				Icon:        "",
			},
			{
				Name:        "Gradle: Test",
				Type:        TaskTest,
				Command:     gradleCmd + " test",
				Description: "Run Gradle tests",
				Icon:        "",
			},
			{
				Name:        "Gradle: Clean",
				Type:        TaskClean,
				Command:     gradleCmd + " clean",
				Description: "Clean Gradle project",
				Icon:        "",
			},
		}...)
	}

	// Also check one level deeper for common package structures
	filepath.WalkDir(projectPath, func(path string, d os.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(projectPath, path)
		if len(strings.Split(rel, string(os.PathSeparator))) > 3 {
			return filepath.SkipDir
		}
		mainPath := filepath.Join(path, "Main.java")
		if fileExists(mainPath) {
			relMain, _ := filepath.Rel(projectPath, mainPath)
			tasks = append(tasks, Task{
				Name:        fmt.Sprintf("Run %s", relMain),
				Type:        TaskRun,
				Command:     fmt.Sprintf("java %s", relMain),
				Description: fmt.Sprintf("Compile and run %s", relMain),
				Icon:        "",
			})
		}
		return nil
	})

	return tasks
}

func detectGoTasks(projectPath string) []Task {
	tasks := []Task{
		{
			Name:        "Build Go Project",
			Type:        TaskBuild,
			Command:     "go build ./...",
			Description: "Build Go project",
			Icon:        "",
		},
		{
			Name:        "Run Tests",
			Type:        TaskTest,
			Command:     "go test ./...",
			Description: "Run all Go tests",
			Icon:        "",
		},
		{
			Name:        "Format Code (gofmt)",
			Type:        TaskFormat,
			Command:     "gofmt -w .",
			Description: "Format Go code",
			Icon:        "",
		},
		{
			Name:        "Run Go Vet",
			Type:        TaskLint,
			Command:     "go vet ./...",
			Description: "Check Go code for issues",
			Icon:        "",
		},
	}

	// Check for main.go
	if fileExists(filepath.Join(projectPath, "main.go")) {
		tasks = append(tasks, Task{
			Name:        "Run main.go",
			Type:        TaskRun,
			Command:     "go run main.go",
			Description: "Execute main.go",
			Icon:        "",
		})
	}

	return tasks
}

func detectMakefileTasks(makefilePath string) []Task {
	var tasks []Task

	data, err := os.ReadFile(makefilePath)
	if err != nil {
		return tasks
	}

	// Simple makefile parsing - look for targets
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasSuffix(line, ":") && !strings.HasPrefix(line, "#") {
			target := strings.TrimSuffix(line, ":")
			target = strings.TrimSpace(target)
			if target != "" && !strings.Contains(target, " ") {
				tasks = append(tasks, Task{
					Name:        fmt.Sprintf("make %s", target),
					Type:        TaskRun,
					Command:     fmt.Sprintf("make %s", target),
					Description: fmt.Sprintf("Run make target: %s", target),
					Icon:        "",
				})
			}
		}
	}

	return tasks
}

func detectRustTasks() []Task {
	return []Task{
		{
			Name:        "Build Rust Project",
			Type:        TaskBuild,
			Command:     "cargo build",
			Description: "Build Rust project",
			Icon:        "",
		},
		{
			Name:        "Build Release",
			Type:        TaskBuild,
			Command:     "cargo build --release",
			Description: "Build optimized release",
			Icon:        "",
		},
		{
			Name:        "Run Tests",
			Type:        TaskTest,
			Command:     "cargo test",
			Description: "Run Rust tests",
			Icon:        "",
		},
		{
			Name:        "Format Code",
			Type:        TaskFormat,
			Command:     "cargo fmt",
			Description: "Format Rust code",
			Icon:        "",
		},
		{
			Name:        "Lint Code (clippy)",
			Type:        TaskLint,
			Command:     "cargo clippy",
			Description: "Lint Rust code",
			Icon:        "",
		},
		{
			Name:        "Run Project",
			Type:        TaskRun,
			Command:     "cargo run",
			Description: "Run Rust project",
			Icon:        "",
		},
	}
}

func detectCTasks(projectPath string) []Task {
	var tasks []Task

	// CMake support
	if fileExists(filepath.Join(projectPath, "CMakeLists.txt")) {
		tasks = append(tasks, []Task{
			{
				Name:        "CMake: Configure",
				Type:        TaskBuild,
				Command:     "cmake -B build",
				Description: "Configure C/C++ project with CMake",
				Icon:        "",
			},
			{
				Name:        "CMake: Build",
				Type:        TaskBuild,
				Command:     "cmake --build build",
				Description: "Build C/C++ project with CMake",
				Icon:        "",
			},
		}...)
	}

	// Search for .c and .cpp files in root and src/
	searchDirs := []string{projectPath, filepath.Join(projectPath, "src")}
	for _, dir := range searchDirs {
		if !fileExists(dir) {
			continue
		}
		files, _ := os.ReadDir(dir)
		for _, f := range files {
			if f.IsDir() {
				continue
			}
			name := f.Name()
			fullPath := filepath.Join(dir, name)
			relPath, _ := filepath.Rel(projectPath, fullPath)

			if strings.HasSuffix(name, ".c") {
				baseName := strings.TrimSuffix(name, ".c")
				tasks = append(tasks, Task{
					Name:        fmt.Sprintf("Compile & Run %s", relPath),
					Type:        TaskRun,
					Command:     fmt.Sprintf("gcc %s -o %s && ./%s", relPath, baseName, baseName),
					Description: "Compile and execute single C file",
					Icon:        "",
				})
			} else if strings.HasSuffix(name, ".cpp") || strings.HasSuffix(name, ".cc") {
				ext := ".cpp"
				if strings.HasSuffix(name, ".cc") {
					ext = ".cc"
				}
				baseName := strings.TrimSuffix(name, ext)
				tasks = append(tasks, Task{
					Name:        fmt.Sprintf("Compile & Run %s", relPath),
					Type:        TaskRun,
					Command:     fmt.Sprintf("g++ %s -o %s && ./%s", relPath, baseName, baseName),
					Description: "Compile and execute single C++ file",
					Icon:        "",
				})
			}
		}
	}

	return tasks
}

// ExecuteTask runs a task in the specified directory
func ExecuteTask(ctx context.Context, task Task, workDir string, outputChan chan<- string) error {
	defer close(outputChan)

	// Parse command
	parts := strings.Fields(task.Command)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	cmd.Dir = workDir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		return err
	}

	// Stream output
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case outputChan <- scanner.Text():
		}
	}

	return cmd.Wait()
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func hasFilesWithExtension(dir, ext string) bool {
	ext = strings.ToLower(ext)
	found := false
	filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Limit depth to 3 levels (root + 2) to avoid scanning whole system or deep node_modules
		rel, _ := filepath.Rel(dir, path)
		depth := len(strings.Split(rel, string(os.PathSeparator)))
		if d.IsDir() {
			if depth > 2 || d.Name() == "node_modules" || d.Name() == ".git" || d.Name() == ".venv" || d.Name() == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(strings.ToLower(d.Name()), ext) {
			found = true
			return fmt.Errorf("found") // shortcut to stop walking
		}
		return nil
	})
	return found
}
