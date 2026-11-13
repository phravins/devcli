package web

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode"
)

//go:embed static/index.html static/app.js static/style.css
var staticAssets embed.FS

const (
	sessionCookieName = "devcli_session"
	// execTimeout bounds every compile/run/shell invocation so a runaway
	// program cannot outlive the HTTP request that started it.
	execTimeout = 25 * time.Second
)

var (
	serverMu      sync.Mutex
	serverStarted bool
	serverPort    string

	sessionMu    sync.Mutex
	sessionToken string

	logMu   sync.Mutex
	logChan chan string

	currentDir string
	activeCmd  *exec.Cmd
	activeMu   sync.Mutex
)

// ensureSessionToken lazily creates the per-process secret that authorizes
// the browser tab DevCLI opens. Nothing is served without it.
func ensureSessionToken() (string, error) {
	sessionMu.Lock()
	defer sessionMu.Unlock()
	if sessionToken == "" {
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			return "", fmt.Errorf("failed to generate session token: %w", err)
		}
		sessionToken = hex.EncodeToString(b)
	}
	return sessionToken, nil
}

// LaunchURL returns the URL the TUI should open in the browser. It carries a
// one-time token that the server exchanges for an HttpOnly session cookie.
func LaunchURL(port string) (string, error) {
	token, err := ensureSessionToken()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("http://127.0.0.1:%s/?token=%s", port, token), nil
}

func sendLog(msg string) {
	logMu.Lock()
	ch := logChan
	logMu.Unlock()
	if ch == nil {
		return
	}
	// Never block a request handler on the TUI: if nobody is listening (the
	// user left the web-compiler screen) or the buffer is full, drop it.
	select {
	case ch <- msg:
	default:
	}
}

func setupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", handleRoot)
	mux.HandleFunc("/logs", handleLogs)
	mux.HandleFunc("/cancel", handleCancel)
	mux.HandleFunc("/save", handleSave)
	mux.HandleFunc("/run", handleRun)
	mux.HandleFunc("/terminal", handleTerminal)
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := strings.TrimPrefix(r.URL.Path, "/")
	if name == "" {
		name = "index.html"
	}

	contentType := ""
	switch name {
	case "index.html":
		contentType = "text/html; charset=utf-8"
	case "style.css":
		contentType = "text/css; charset=utf-8"
	case "app.js":
		contentType = "application/javascript; charset=utf-8"
	default:
		http.NotFound(w, r)
		return
	}

	data, err := staticAssets.ReadFile("static/" + name)
	if err != nil {
		http.Error(w, "Error loading page", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "no-store")
	w.Write(data)
}

func handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading body", http.StatusInternalServerError)
		return
	}

	sanitizedBody := strings.ReplaceAll(string(body), "\r", "\\r")
	sanitizedBody = strings.ReplaceAll(sanitizedBody, "\n", "\\n")

	// Do not print to stdout: the TUI owns the terminal (alt screen) and raw
	// writes corrupt it. The TUI shows these via the log channel.
	sendLog(fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), sanitizedBody))
	w.WriteHeader(http.StatusOK)
}

func handleCancel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
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

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var payload struct {
		Filename string `json:"filename"`
		Content  string `json:"content"`
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
			sendLog("Failed to create directory: " + err.Error())
			http.Error(w, "Failed to create directory", http.StatusInternalServerError)
			return
		}
	}

	if err := writeFileNoFollow(filename, []byte(payload.Content)); err != nil {
		sendLog("Error saving file: " + err.Error())
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	sendLog("File saved successfully: " + filename)
	w.WriteHeader(http.StatusOK)
}

// writeFileNoFollow refuses to write through a symlink, so a link planted in
// the working directory cannot redirect a save outside of it.
func writeFileNoFollow(name string, data []byte) error {
	if fi, err := os.Lstat(name); err == nil && fi.Mode()&os.ModeSymlink != 0 {
		return errors.New("refusing to write through symlink")
	}
	return os.WriteFile(name, data, 0644)
}

func handleRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}

	language := "python"
	code := string(bodyBytes)

	var reqPayload struct {
		Language string `json:"language"`
		Code     string `json:"code"`
	}
	if err := json.Unmarshal(bodyBytes, &reqPayload); err == nil && reqPayload.Code != "" {
		code = reqPayload.Code
		if reqPayload.Language != "" {
			language = strings.ToLower(reqPayload.Language)
		}
	}

	output, err := runCode(r.Context(), language, code)

	if err != nil {
		sendLog(fmt.Sprintf("[%s Execution Error]: %v", titleCase(language), err))
	} else {
		sendLog(fmt.Sprintf("[%s Execution Success]", titleCase(language)))
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

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}

	command := string(body)
	output, err := runShell(r.Context(), command)

	response := map[string]string{
		"output": output,
	}
	if err != nil {
		response["error"] = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func setSecurityHeaders(w http.ResponseWriter) {
	h := w.Header()
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("X-Frame-Options", "DENY")
	h.Set("Content-Security-Policy",
		"default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; "+
			"img-src 'self' data:; connect-src 'self'; font-src 'self'; "+
			"object-src 'none'; base-uri 'none'; form-action 'self'; frame-ancestors 'none'")
	h.Set("Referrer-Policy", "no-referrer")
	h.Set("Cross-Origin-Opener-Policy", "same-origin")
	h.Set("Cross-Origin-Resource-Policy", "same-origin")
}

// newSecureHandler wraps the routes with the checks that keep the local code
// execution endpoints from being reachable by anything but the browser tab
// DevCLI itself opened:
//   - Host must be 127.0.0.1/localhost on our port (blocks DNS rebinding)
//   - Origin, when sent, must be exactly our origin, port included (blocks
//     other localhost apps and cross-site pages)
//   - every request must carry the session cookie, obtained only by visiting
//     the one-time LaunchURL (blocks other local users/processes)
func newSecureHandler(port, token string, next http.Handler) http.Handler {
	allowed := map[string]bool{
		"127.0.0.1:" + port: true,
		"localhost:" + port: true,
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setSecurityHeaders(w)

		if !allowed[strings.ToLower(r.Host)] {
			http.Error(w, "Forbidden: invalid Host header", http.StatusForbidden)
			return
		}

		if origin := r.Header.Get("Origin"); origin != "" {
			u, err := url.Parse(origin)
			if err != nil || !allowed[strings.ToLower(u.Host)] {
				http.Error(w, "Forbidden: Invalid Origin header", http.StatusForbidden)
				return
			}
		} else if r.Method != http.MethodGet && r.Method != http.MethodHead &&
			r.Header.Get("Sec-Fetch-Site") == "cross-site" {
			http.Error(w, "Forbidden: cross-site request", http.StatusForbidden)
			return
		}

		if token == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if r.Method == http.MethodGet && r.URL.Path == "/" && r.URL.Query().Has("token") {
			if subtle.ConstantTimeCompare([]byte(r.URL.Query().Get("token")), []byte(token)) != 1 {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			http.SetCookie(w, &http.Cookie{
				Name:     sessionCookieName,
				Value:    token,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
			})
			// Drop the token from the address bar and history.
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		c, err := r.Cookie(sessionCookieName)
		if err != nil || subtle.ConstantTimeCompare([]byte(c.Value), []byte(token)) != 1 {
			http.Error(w, "Unauthorized: open the Web Compiler from DevCLI", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func newServer(port, token string) *http.Server {
	mux := http.NewServeMux()
	setupRoutes(mux)
	return &http.Server{
		Addr:              "127.0.0.1:" + port,
		Handler:           newSecureHandler(port, token, mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      execTimeout + 5*time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 16,
	}
}

func StartServer(port string, logs chan string) error {
	logMu.Lock()
	logChan = logs
	logMu.Unlock()

	serverMu.Lock()
	if serverStarted {
		running := serverPort
		serverMu.Unlock()
		if running == port {
			return nil
		}
		return fmt.Errorf("server already running on port %s", running)
	}

	token, err := ensureSessionToken()
	if err != nil {
		serverMu.Unlock()
		return err
	}
	serverStarted = true
	serverPort = port
	serverMu.Unlock()

	sendLog("[TUI] Web compiler listening on http://127.0.0.1:" + port)

	err = newServer(port, token).ListenAndServe()
	serverMu.Lock()
	serverStarted = false
	serverMu.Unlock()
	return err
}

func runCode(parent context.Context, lang, code string) (string, error) {
	ctx, cancel := context.WithTimeout(parent, execTimeout)
	defer cancel()

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
		cmd = exec.CommandContext(ctx, cmdName, tmpfile.Name())

	case "go", "golang":
		cmdName := "go"
		if path, err := exec.LookPath("go"); err == nil {
			cmdName = path
		} else if _, err := os.Stat("/usr/local/go/bin/go"); err == nil {
			cmdName = "/usr/local/go/bin/go"
		}
		cmd = exec.CommandContext(ctx, cmdName, "run", tmpfile.Name())

	case "rust":
		binPath := tmpfile.Name() + ".bin"
		defer os.Remove(binPath)
		compileCmd := exec.CommandContext(ctx, "rustc", tmpfile.Name(), "-o", binPath)
		if out, err := compileCmd.CombinedOutput(); err != nil {
			return string(out), fmt.Errorf("rust compilation error: %w", err)
		}
		cmd = exec.CommandContext(ctx, binPath)

	case "cpp", "c++", "c":
		binPath := tmpfile.Name() + ".bin"
		defer os.Remove(binPath)
		compiler := "g++"
		if _, err := exec.LookPath("g++"); err != nil {
			compiler = "gcc"
		}
		compileCmd := exec.CommandContext(ctx, compiler, tmpfile.Name(), "-o", binPath)
		if out, err := compileCmd.CombinedOutput(); err != nil {
			return string(out), fmt.Errorf("C/C++ compilation error: %w", err)
		}
		cmd = exec.CommandContext(ctx, binPath)

	default:
		cmdName := "python"
		if _, err := exec.LookPath("python"); err != nil {
			cmdName = "python3"
		}
		cmd = exec.CommandContext(ctx, cmdName, "-u", tmpfile.Name())
	}

	cmd.Env = os.Environ()

	activeMu.Lock()
	activeCmd = cmd
	activeMu.Unlock()

	output, err := cmd.CombinedOutput()

	activeMu.Lock()
	activeCmd = nil
	activeMu.Unlock()

	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		err = fmt.Errorf("execution timed out after %s", execTimeout)
	}

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
	"ls":      true,
	"pwd":     true,
	"cat":     true,
	"echo":    true,
	"go":      true,
	"python":  true,
	"python3": true,
	"node":    true,
	"git":     true,
	"grep":    true,
	"head":    true,
	"tail":    true,
	"npm":     true,
}

func runShell(parent context.Context, command string) (string, error) {
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

	ctx, cancel := context.WithTimeout(parent, execTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, baseCmd, args[1:]...)

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
