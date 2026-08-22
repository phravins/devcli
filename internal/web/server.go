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

	"github.com/phravins/devcli/pkg/utils"
)

//go:embed static/*
var staticAssets embed.FS

// Global state for the web server
var (
	serverStarted bool       // Tracks if the server is currently running
	serverPort    string     // The port number the server is listening on
	currentDir    string     // The working directory for terminal commands
	activeCmd     *exec.Cmd  // Currently running command (for cancellation)
	activeMu      sync.Mutex // Protects access to activeCmd from multiple threads
	logChan       chan string // Channel to send logs to the TUI
)

// StartServer launches the web-based Python compiler on the specified port
func StartServer(port string, logs chan string) error {
	logChan = logs
	if serverStarted {
		if serverPort == port {
			return nil // Server already running on the correct port, nothing to do
		}
		return fmt.Errorf("server already running on port %s", serverPort)
	}

	mux := http.NewServeMux()

	// Serve static files from embedded FS
	fileServer := http.FileServer(http.FS(staticAssets))
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))

	// Redirect root to index.html in static
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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

		// Serve other static files directly if they exist in static/
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
	})

	// Auth Routes
	mux.HandleFunc("/auth/login", handleLogin)
	mux.HandleFunc("/auth/register", handleRegister)
	mux.HandleFunc("/auth/logout", handleLogout)
	mux.HandleFunc("/auth/forgot-password", handleForgotPassword)
	mux.HandleFunc("/auth/verify-email", handleVerifyEmail)
	mux.HandleFunc("/auth/google", func(w http.ResponseWriter, r *http.Request) {
		// Mock Google OAuth redirect
		fmt.Fprintf(w, "Google OAuth redirect would happen here.")
	})

	// Storage Routes
	mux.HandleFunc("/drive/save", handleDriveSave)

	// Logging Sync Route
	mux.HandleFunc("/logs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading body", http.StatusInternalServerError)
			return
		}
		msg := fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), string(body))
		fmt.Printf("\n[WEB-COMPILER LOG] %s\n", msg)
		if logChan != nil {
			logChan <- msg
		}
		w.WriteHeader(http.StatusOK)
	})

	// Handle Ctrl+C from the web terminal
	mux.HandleFunc("/cancel", func(w http.ResponseWriter, r *http.Request) {
		activeMu.Lock()
		defer activeMu.Unlock()
		if activeCmd != nil && activeCmd.Process != nil {
			activeCmd.Process.Kill()
		}
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/save", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload struct {
			Filename string `json:"filename"`
			Content  string `json:"content"`
		}

		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		filename := payload.Filename
		if filename == "" {
			http.Error(w, "Filename required", http.StatusBadRequest)
			return
		}

		// Create parent directories if they don't exist
		dir := filepath.Dir(filename)
		if dir != "." && dir != "/" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				msg := "Failed to create directory: " + err.Error()
				if logChan != nil {
					logChan <- msg
				}
				http.Error(w, msg, http.StatusInternalServerError)
				return
			}
		}

		err := os.WriteFile(filename, []byte(payload.Content), 0644)
		if err != nil {
			msg := "Error saving file: " + err.Error()
			if logChan != nil {
				logChan <- msg
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		msg := "File saved successfully: " + filename
		if logChan != nil {
			logChan <- msg
		}
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/run", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading body", http.StatusBadRequest)
			return
		}

		output, err := runPython(string(body))

		if logChan != nil {
			if err != nil {
				logChan <- "Python Execution Error: " + err.Error()
			} else {
				logChan <- "Python Execution Success"
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
	})

	mux.HandleFunc("/terminal", func(w http.ResponseWriter, r *http.Request) {
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

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	serverStarted = true
	serverPort = port
	addr := "127.0.0.1:" + port
	fmt.Printf("Starting local premium compiler server at http://%s\n", addr)

	err := http.ListenAndServe(addr, mux)
	if err != nil {
		serverStarted = false
	}
	return err
}

// runPython executes Python code and returns the output
func runPython(code string) (string, error) {
	// Create a temporary Python file to hold the code
	tmpfile, err := os.CreateTemp("", "devcli-*.py")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(code)); err != nil {
		return "", err
	}
	if err := tmpfile.Close(); err != nil {
		return "", err
	}

	// Determine which Python command to use
	// Try "python" first (common on Windows), fallback to "python3" (common on Linux/Mac)
	cmdName := "python"
	if _, err := exec.LookPath("python"); err != nil {
		cmdName = "python3"
	}

	cmd := exec.Command(cmdName, "-u", tmpfile.Name()) // -u = unbuffered output
	cmd.Env = os.Environ()                             // Pass environment variables to the Python process

	// Register this command so it can be cancelled with Ctrl+C
	activeMu.Lock()
	activeCmd = cmd
	activeMu.Unlock()

	output, err := cmd.CombinedOutput()

	activeMu.Lock()
	activeCmd = nil
	activeMu.Unlock()

	// Provide helpful feedback if the code produced no output
	outStr := string(output)
	if outStr == "" && err == nil {
		outStr = fmt.Sprintf("[No output]\n(Ran: %s -u %s)", cmdName, tmpfile.Name())
	}

	return outStr, err
}

// runShell executes shell commands in the web terminal
func runShell(command string) (string, error) {
	if currentDir == "" {
		currentDir, _ = os.Getwd()
	}

	// Handle the 'cd' command specially to change directories
	if len(command) >= 3 && command[:3] == "cd " {
		path := strings.TrimSpace(command[3:])
		// Convert relative paths to absolute paths
		newDir := filepath.Join(currentDir, path)
		if filepath.IsAbs(path) {
			newDir = path
		}

		// Make sure the directory actually exists
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

	var cmd *exec.Cmd
	cmd = utils.GetShellCommand(command)

	cmd.Dir = currentDir
	cmd.Env = os.Environ() // Pass environment variables to the shell

	// Register this command so it can be cancelled with Ctrl+C
	activeMu.Lock()
	activeCmd = cmd
	activeMu.Unlock()

	output, err := cmd.CombinedOutput()

	activeMu.Lock()
	activeCmd = nil
	activeMu.Unlock()

	return string(output), err
}
