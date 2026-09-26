package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode"
)

var staticAssets embed.FS

var (
	serverStarted	bool
	serverPort	string
	currentDir	string
	activeCmd	*exec.Cmd
	activeMu	sync.Mutex
	logChan		chan string
)

func setupRoutes(mux *http.ServeMux) {

	fileServer := http.FileServer(http.FS(staticAssets))
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))

	mux.HandleFunc("/", handleRoot)


	mux.HandleFunc("/logs", handleLogs)
	mux.HandleFunc("/cancel", handleCancel)
	mux.HandleFunc("/save", handleSave)
	mux.HandleFunc("/run", handleRun)
	mux.HandleFunc("/terminal", handleTerminal)
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		data, err := staticAssets.ReadFile("static/index.html")
		if err != nil {
			http.Error(w, "Error loading page", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write(data)
		return
	}

	filePath := "static" + r.URL.Path
	data, err := staticAssets.ReadFile(filePath)
	if err == nil {
		contentType := "text/plain"
		switch {
		case strings.HasSuffix(filePath, ".css"):
			contentType = "text/css"
		case strings.HasSuffix(filePath, ".js"):
			contentType = "application/javascript"
		case strings.HasSuffix(filePath, ".png"):
			contentType = "image/png"
		}
		w.Header().Set("Content-Type", contentType)
		w.Write(data)
		return
	}

	http.NotFound(w, r)
}


func handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading body", http.StatusInternalServerError)
		return
	}

	sanitizedBody := strings.ReplaceAll(string(body), "\r", "\\r")
	sanitizedBody = strings.ReplaceAll(sanitizedBody, "\n", "\\n")

	msg := fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), sanitizedBody)
	fmt.Printf("\n[WEB-COMPILER LOG] %s\n", msg)
	if logChan != nil {
		logChan <- msg
	}
	w.WriteHeader(http.StatusOK)
}

func handleCancel(w http.ResponseWriter, r *http.Request) {
	activeMu.Lock()
	defer activeMu.Unlock()
	if activeCmd != nil && activeCmd.Process != nil {
		activeCmd.Process.Kill()
	}
	w.WriteHeader(http.StatusOK)
}

func handleSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}


	var payload struct {
		Filename	string	`json:"filename"`
		Content		string	`json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	filename := filepath.Clean(payload.Filename)
	if payload.Filename == "" || !filepath.IsLocal(filename) {
		http.Error(w, "Invalid filename or path traversal detected", http.StatusBadRequest)
		return
	}

	dir := filepath.Dir(filename)
	if dir != "." && dir != "/" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			msg := "Failed to create directory: " + err.Error()
			if logChan != nil {
				logChan <- msg
			}
			http.Error(w, "Failed to create directory", http.StatusInternalServerError)
			return
		}
	}

	err := os.WriteFile(filename, []byte(payload.Content), 0644)
	if err != nil {
		msg := "Error saving file: " + err.Error()
		if logChan != nil {
			logChan <- msg
		}
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	msg := "File saved successfully: " + filename
	if logChan != nil {
		logChan <- msg
	}
	w.WriteHeader(http.StatusOK)
}

func handleRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}


	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}

	language := "python"
	code := string(bodyBytes)

	var reqPayload struct {
		Language	string	`json:"language"`
		Code		string	`json:"code"`
	}
	if err := json.Unmarshal(bodyBytes, &reqPayload); err == nil && reqPayload.Code != "" {
		code = reqPayload.Code
		if reqPayload.Language != "" {
			language = strings.ToLower(reqPayload.Language)
		}
	}

	output, err := runCode(language, code)

	if logChan != nil {
		if err != nil {
			logChan <- fmt.Sprintf("[%s Execution Error]: %v", titleCase(language), err)
		} else {
			logChan <- fmt.Sprintf("[%s Execution Success]", titleCase(language))
		}
	}

	response := map[string]string{
		"output": output,
	}
	if err != nil {
		response["error"] = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleTerminal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}


	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}

	command := string(body)
	output, err := runShell(command)

	response := map[string]string{
		"output": output,
	}
	if err != nil {
		response["error"] = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func StartServer(port string, logs chan string) error {
	logChan = logs
	if serverStarted {
		if serverPort == port {
			return nil
		}
		return fmt.Errorf("server already running on port %s", serverPort)
	}

	mux := http.NewServeMux()
	setupRoutes(mux)

	serverStarted = true
	serverPort = port
	addr := "127.0.0.1:" + port
	fmt.Printf("Starting local premium compiler server at http://%s\n", addr)

	secureHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Content-Security-Policy", "default-src 'self' 'unsafe-inline' 'unsafe-eval'; img-src 'self' data:;")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		mux.ServeHTTP(w, r)
	})

	err := http.ListenAndServe(addr, secureHandler)
	if err != nil {
		serverStarted = false
	}
	return err
}

func runCode(lang, code string) (string, error) {
	lang = strings.ToLower(strings.TrimSpace(lang))

	var ext string
	switch lang {
	case "js", "javascript", "node":
		ext = ".js"
	case "go", "golang":
		ext = ".go"
	case "rust":
		ext = ".rs"
	case "cpp", "c++", "c":
		ext = ".cpp"
	default:
		ext = ".py"
	}

	tmpfile, err := os.CreateTemp("", "devcli-*"+ext)
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(code)); err != nil {
		return "", err
	}
	tmpfile.Close()

	var cmd *exec.Cmd
	switch lang {
	case "js", "javascript", "node":
		cmdName := "node"
		if _, err := exec.LookPath("node"); err != nil {
			return "", fmt.Errorf("node.js is not installed or not in PATH")
		}
		cmd = exec.Command(cmdName, tmpfile.Name())

	case "go", "golang":
		cmdName := "go"
		if path, err := exec.LookPath("go"); err == nil {
			cmdName = path
		} else if _, err := os.Stat("/usr/local/go/bin/go"); err == nil {
			cmdName = "/usr/local/go/bin/go"
		}
		cmd = exec.Command(cmdName, "run", tmpfile.Name())

	case "rust":
		binPath := tmpfile.Name() + ".bin"
		defer os.Remove(binPath)
		compileCmd := exec.Command("rustc", tmpfile.Name(), "-o", binPath)
		if out, err := compileCmd.CombinedOutput(); err != nil {
			return string(out), fmt.Errorf("rust compilation error: %w", err)
		}
		cmd = exec.Command(binPath)

	case "cpp", "c++", "c":
		binPath := tmpfile.Name() + ".bin"
		defer os.Remove(binPath)
		compiler := "g++"
		if _, err := exec.LookPath("g++"); err != nil {
			compiler = "gcc"
		}
		compileCmd := exec.Command(compiler, tmpfile.Name(), "-o", binPath)
		if out, err := compileCmd.CombinedOutput(); err != nil {
			return string(out), fmt.Errorf("C/C++ compilation error: %w", err)
		}
		cmd = exec.Command(binPath)

	default:
		cmdName := "python"
		if _, err := exec.LookPath("python"); err != nil {
			cmdName = "python3"
		}
		cmd = exec.Command(cmdName, "-u", tmpfile.Name())
	}

	cmd.Env = os.Environ()

	activeMu.Lock()
	activeCmd = cmd
	activeMu.Unlock()

	output, err := cmd.CombinedOutput()

	activeMu.Lock()
	activeCmd = nil
	activeMu.Unlock()

	outStr := string(output)
	if outStr == "" && err == nil {
		outStr = fmt.Sprintf("[No output]\n(Ran %s code)", lang)
	}

	return outStr, err
}

func parseCommand(command string) []string {
	var args []string
	var current strings.Builder
	inSingleQuote := false
	inDoubleQuote := false
	escaped := false

	for _, r := range command {
		if escaped {
			current.WriteRune(r)
			escaped = false
			continue
		}

		if r == '\\' && !inSingleQuote {
			escaped = true
			continue
		}

		if r == '\'' && !inDoubleQuote {
			inSingleQuote = !inSingleQuote
			continue
		}

		if r == '"' && !inSingleQuote {
			inDoubleQuote = !inDoubleQuote
			continue
		}

		if unicode.IsSpace(r) && !inSingleQuote && !inDoubleQuote {
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
			continue
		}

		current.WriteRune(r)
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}

var allowedCommands = map[string]bool{
	"ls":		true,
	"pwd":		true,
	"cat":		true,
	"echo":		true,
	"go":		true,
	"python":	true,
	"python3":	true,
	"node":		true,
	"git":		true,
	"grep":		true,
	"head":		true,
	"tail":		true,
	"npm":		true,
}

func runShell(command string) (string, error) {
	if currentDir == "" {
		currentDir, _ = os.Getwd()
	}

	command = strings.TrimSpace(command)
	if command == "" {
		return "", nil
	}

	if strings.HasPrefix(command, "cd ") || command == "cd" {
		path := ""
		if len(command) > 3 {
			path = strings.TrimSpace(command[3:])
		}

		if path == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("could not determine home directory")
			}
			path = homeDir
		}

		newDir := filepath.Join(currentDir, path)
		if filepath.IsAbs(path) {
			newDir = path
		}

		info, err := os.Stat(newDir)
		if err != nil {
			return "", fmt.Errorf("directory not found: %s", path)
		}
		if !info.IsDir() {
			return "", fmt.Errorf("not a directory: %s", path)
		}

		currentDir = newDir
		return fmt.Sprintf("Changed directory to %s", currentDir), nil
	}

	args := parseCommand(command)
	if len(args) == 0 {
		return "", nil
	}

	baseCmd := args[0]
	if !allowedCommands[baseCmd] {
		return "", fmt.Errorf("command not allowed for security reasons: %s. Allowed commands are: ls, pwd, cat, echo, go, python, node, git, grep, head, tail, npm", baseCmd)
	}

	var cmd *exec.Cmd
	if len(args) > 1 {
		cmd = exec.Command(baseCmd, args[1:]...)
	} else {
		cmd = exec.Command(baseCmd)
	}

	cmd.Dir = currentDir
	cmd.Env = os.Environ()

	activeMu.Lock()
	activeCmd = cmd
	activeMu.Unlock()

	output, err := cmd.CombinedOutput()

	activeMu.Lock()
	activeCmd = nil
	activeMu.Unlock()

	return string(output), err
}

func titleCase(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
