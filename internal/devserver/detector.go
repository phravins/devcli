package devserver

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

type ProjectType string

const (
	TypeNode	ProjectType	= "Node.js"
	TypeExpress	ProjectType	= "Express"
	TypeNextJS	ProjectType	= "Next.js"
	TypeNestJS	ProjectType	= "Nest.js"
	TypeAngular	ProjectType	= "Angular"
	TypeVue		ProjectType	= "Vue.js"
	TypePython	ProjectType	= "Python"
	TypeGo		ProjectType	= "Go"
	TypeDjango	ProjectType	= "Django"
	TypeFastAPI	ProjectType	= "FastAPI"
	TypeFlask	ProjectType	= "Flask"
	TypeReact	ProjectType	= "React"
	TypeVite	ProjectType	= "Vite"
	TypeWebpack	ProjectType	= "Webpack"
	TypeSpring	ProjectType	= "Spring Boot"
	TypeFullstack	ProjectType	= "Fullstack"
	TypeUnknown	ProjectType	= "Unknown"
)

type ServerConfig struct {
	Name	string
	Type	ProjectType
	Cmd	string
	Args	[]string
	Dir	string
}

type ProjectInfo struct {
	Type	ProjectType
	Servers	[]ServerConfig
}

func Detect(path string) ProjectInfo {
	if path == "" {
		path, _ = os.Getwd()
	}

	var servers []ServerConfig
	detectedType := TypeUnknown

	if exists(filepath.Join(path, "manage.py")) {
		servers = append(servers, ServerConfig{
			Name:	"Django Server",
			Type:	TypeDjango,
			Cmd:	"python",
			Args:	[]string{"manage.py", "runserver"},
			Dir:	path,
		})
		detectedType = TypeDjango
	}

	if isFastAPI(path) {
		entrypoint := "main.py"
		if !exists(filepath.Join(path, "main.py")) && exists(filepath.Join(path, "app.py")) {
			entrypoint = "app.py"
		}
		appName := strings.TrimSuffix(entrypoint, ".py") + ":app"
		servers = append(servers, ServerConfig{
			Name:	"FastAPI Server",
			Type:	TypeFastAPI,
			Cmd:	"uvicorn",
			Args:	[]string{appName, "--reload"},
			Dir:	path,
		})
		detectedType = TypeFastAPI
	}

	if exists(filepath.Join(path, "pom.xml")) {
		servers = append(servers, ServerConfig{
			Name:	"Spring Boot",
			Type:	TypeSpring,
			Cmd:	"mvn",
			Args:	[]string{"spring-boot:run"},
			Dir:	path,
		})
		detectedType = TypeSpring
	}

	if exists(filepath.Join(path, "next.config.js")) || exists(filepath.Join(path, "next.config.mjs")) || exists(filepath.Join(path, "pages")) || exists(filepath.Join(path, "app")) {
		servers = append(servers, ServerConfig{
			Name:	"Next.js Dev Server",
			Type:	TypeNextJS,
			Cmd:	"npm",
			Args:	[]string{"run", "dev"},
			Dir:	path,
		})
		detectedType = TypeNextJS
	}

	if exists(filepath.Join(path, "nest-cli.json")) {
		servers = append(servers, ServerConfig{
			Name:	"Nest.js Dev Server",
			Type:	TypeNestJS,
			Cmd:	"npm",
			Args:	[]string{"run", "start:dev"},
			Dir:	path,
		})
		detectedType = TypeNestJS
	}

	if exists(filepath.Join(path, "angular.json")) {
		servers = append(servers, ServerConfig{
			Name:	"Angular Dev Server",
			Type:	TypeAngular,
			Cmd:	"npm",
			Args:	[]string{"start"},
			Dir:	path,
		})
		detectedType = TypeAngular
	}

	var pkgJsonContent []byte
	var pkgJsonRead bool
	var pkgJsonExists bool

	getPkgJson := func() ([]byte, bool) {
		if !pkgJsonRead {
			pkgJsonRead = true
			pkgPath := filepath.Join(path, "package.json")
			if exists(pkgPath) {
				pkgJsonExists = true
				pkgJsonContent, _ = os.ReadFile(pkgPath)
			}
		}
		return pkgJsonContent, pkgJsonExists
	}

	if isVue(path, getPkgJson) && len(servers) == 0 {
		servers = append(servers, ServerConfig{
			Name:	"Vue Dev Server",
			Type:	TypeVue,
			Cmd:	"npm",
			Args:	[]string{"run", "dev"},
			Dir:	path,
		})
		detectedType = TypeVue
	}

	if exists(filepath.Join(path, "vite.config.js")) || exists(filepath.Join(path, "vite.config.ts")) {
		if len(servers) == 0 {
			servers = append(servers, ServerConfig{
				Name:	"Vite Dev Server",
				Type:	TypeVite,
				Cmd:	"npm",
				Args:	[]string{"run", "dev"},
				Dir:	path,
			})
			detectedType = TypeVite
		}
	}

	if exists(filepath.Join(path, "webpack.config.js")) && len(servers) == 0 {
		servers = append(servers, ServerConfig{
			Name:	"Webpack Dev Server",
			Type:	TypeWebpack,
			Cmd:	"npm",
			Args:	[]string{"run", "dev"},
			Dir:	path,
		})
		detectedType = TypeWebpack
	}

	if isReact(getPkgJson) && len(servers) == 0 {
		servers = append(servers, ServerConfig{
			Name:	"React Dev Server",
			Type:	TypeReact,
			Cmd:	"npm",
			Args:	[]string{"start"},
			Dir:	path,
		})
		detectedType = TypeReact
	}

	if isExpress(getPkgJson) && len(servers) == 0 {
		servers = append(servers, ServerConfig{
			Name:	"Express Server",
			Type:	TypeExpress,
			Cmd:	"npm",
			Args:	[]string{"start"},
			Dir:	path,
		})
		detectedType = TypeExpress
	}

	_, pkgExists := getPkgJson()
	if pkgExists && len(servers) == 0 {
		servers = append(servers, ServerConfig{
			Name:	"Node.js Server",
			Type:	TypeNode,
			Cmd:	"npm",
			Args:	[]string{"start"},
			Dir:	path,
		})
		detectedType = TypeNode
	}

	if isFlask(path) && len(servers) == 0 {
		servers = append(servers, ServerConfig{
			Name:	"Flask Server",
			Type:	TypeFlask,
			Cmd:	"flask",
			Args:	[]string{"run", "--debug"},
			Dir:	path,
		})
		detectedType = TypeFlask
	}

	if (exists(filepath.Join(path, "requirements.txt")) || exists(filepath.Join(path, "main.py")) || exists(filepath.Join(path, "app.py"))) && len(servers) == 0 {
		cmd := "python"
		args := []string{}
		if exists(filepath.Join(path, "main.py")) {
			args = append(args, "main.py")
		} else if exists(filepath.Join(path, "app.py")) {
			args = append(args, "app.py")
		}
		servers = append(servers, ServerConfig{
			Name:	"Python Server",
			Type:	TypePython,
			Cmd:	cmd,
			Args:	args,
			Dir:	path,
		})
		detectedType = TypePython
	}

	if exists(filepath.Join(path, "go.mod")) && len(servers) == 0 {
		servers = append(servers, ServerConfig{
			Name:	"Go Server",
			Type:	TypeGo,
			Cmd:	"go",
			Args:	[]string{"run", "."},
			Dir:	path,
		})
		detectedType = TypeGo
	}

	fullstackPatterns := []struct {
		backend		string
		frontend	string
	}{
		{"backend", "frontend"},
		{"server", "client"},
		{"api", "web"},
		{"api", "client"},
		{"backend", "client"},
		{"server", "frontend"},
	}

	for _, pattern := range fullstackPatterns {
		backendPath := filepath.Join(path, pattern.backend)
		frontendPath := filepath.Join(path, pattern.frontend)

		if exists(backendPath) && exists(frontendPath) {
			backendInfo := Detect(backendPath)
			frontendInfo := Detect(frontendPath)

			if len(backendInfo.Servers) > 0 && len(frontendInfo.Servers) > 0 {
				servers = []ServerConfig{}
				for _, srv := range backendInfo.Servers {
					srv.Name = "Backend"
					servers = append(servers, srv)
				}
				for _, srv := range frontendInfo.Servers {
					srv.Name = "Frontend"
					servers = append(servers, srv)
				}
				detectedType = TypeFullstack
				break
			}
		}
	}

	if len(servers) > 1 && detectedType != TypeFullstack {
		detectedType = TypeFullstack
	}

	return ProjectInfo{
		Type:		detectedType,
		Servers:	servers,
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func isFastAPI(path string) bool {
	files := []string{"main.py", "app.py"}
	for _, file := range files {
		fullPath := filepath.Join(path, file)
		if exists(fullPath) {
			if containsImport(fullPath, "fastapi") {
				return true
			}
		}
	}
	return false
}

func isReact(getPkgJson func() ([]byte, bool)) bool {
	if content, exists := getPkgJson(); exists {
		return strings.Contains(string(content), "\"react\"")
	}
	return false
}

func isVue(path string, getPkgJson func() ([]byte, bool)) bool {
	if content, exists := getPkgJson(); exists {
		if strings.Contains(string(content), "\"vue\"") {
			return true
		}
	}
	return exists(filepath.Join(path, "vue.config.js"))
}

func isExpress(getPkgJson func() ([]byte, bool)) bool {
	if content, exists := getPkgJson(); exists {
		return strings.Contains(string(content), "\"express\"")
	}
	return false
}

func isFlask(path string) bool {
	files := []string{"app.py", "main.py"}
	for _, file := range files {
		fullPath := filepath.Join(path, file)
		if exists(fullPath) {
			if containsImport(fullPath, "flask") {
				return true
			}
		}
	}
	return false
}

func containsImport(filePath, importName string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "import ") || strings.HasPrefix(line, "from ") {
			if strings.Contains(line, importName) {
				return true
			}
		}
	}
	return false
}
