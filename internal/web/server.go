package web

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/phravins/devcli/pkg/utils"
)

const htmlContent = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>DevCLI Python Compiler</title>
    <link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;700&family=Inter:wght@400;600&display=swap" rel="stylesheet">
    <style>
        :root {
            --bg-primary: #EFE9E0; /* Soft Ivory */
            --bg-secondary: #FFFFFF;
            --accent: #0F9E99; /* Tropical Teal */
            --text-primary: #1E293B;
            --text-secondary: #475569;
            --border: #CBD5E1;
            --success: #059669;
            --error: #DC2626;
            --bg-terminal: #1E293B; /* Slate 800 */
            --text-terminal: #F8FAFC;
        }

        [data-theme="dark"] {
            --bg-primary: #0F172A; /* Slate 900 */
            --bg-secondary: #1E293B; /* Slate 800 */
            --text-primary: #F8FAFC;
            --text-secondary: #94A3B8;
            --border: #334155;
            --success: #4ADE80;
            --error: #F87171;
            --bg-terminal: #020617; /* Slate 950 */
            --text-terminal: #F8FAFC;
        }

        body {
            margin: 0;
            padding: 0;
            font-family: 'Inter', sans-serif;
            background-color: var(--bg-primary);
            color: var(--text-primary);
            height: 100vh;
            display: flex;
            flex-direction: column;
        }

        header {
            background-color: var(--bg-secondary);
            padding: 1rem 2rem;
            border-bottom: 1px solid var(--border);
            display: flex;
            justify-content: space-between;
            align-items: center;
        }

        h1 {
            font-size: 1.25rem;
            font-weight: 600;
            margin: 0;
            color: var(--accent);
        }

        .main-container {
            display: flex;
            flex-direction: column;
            flex: 1;
            overflow: hidden;
            padding: 1rem;
            gap: 1rem;
        }

        .pane {
            background-color: var(--bg-secondary);
            border: 1px solid var(--border);
            border-radius: 0.5rem;
            display: flex;
            flex-direction: column;
            overflow: hidden;
        }

        .top-pane {
            flex: 2;
        }

        .bottom-pane {
            flex: 1;
        }

        .panel-header {
            padding: 0.75rem 1rem;
            border-bottom: 1px solid var(--border);
            display: flex;
            gap: 1rem;
            align-items: center;
        }

        .tab-btn {
            background: none;
            border: none;
            color: var(--text-secondary);
            font-weight: 600;
            cursor: pointer;
            padding: 0.25rem 0.5rem;
            border-radius: 0.25rem;
        }

        .tab-btn.active {
            color: var(--accent);
            border-bottom: 2px solid var(--accent);
            background-color: transparent;
        }

        textarea {
            flex: 1;
            background-color: var(--bg-secondary);
            color: var(--text-primary);
            border: none;
            padding: 1rem;
            font-family: 'JetBrains Mono', monospace;
            font-size: 0.9rem;
            resize: none;
            outline: none;
            line-height: 1.5;
        }

        .terminal-container {
            flex: 1;
            display: flex;
            flex-direction: column;
            padding: 1rem;
            font-family: 'JetBrains Mono', monospace;
            font-size: 0.9rem;
            background-color: var(--bg-terminal);
            color: var(--text-terminal);
            overflow-y: auto;
        }

        .terminal-output {
            color: var(--text-terminal);
            white-space: pre-wrap;
            margin-bottom: 0.5rem;
        }

        .terminal-input-line {
            display: flex;
            gap: 0.5rem;
            color: #FFFFFF;
            font-weight: bold;
        }

        .terminal-input {
            flex: 1;
            background: none;
            border: none;
            color: #FFFFFF;
            font-family: inherit;
            font-size: inherit;
            outline: none;
            font-weight: bold;
        }

        .header-controls {
            display: flex;
            gap: 0.5rem;
            align-items: center;
        }

        .filename-input {
            background: var(--bg-primary);
            border: 1px solid var(--border);
            color: var(--text-primary);
            padding: 0.25rem 0.5rem;
            border-radius: 0.25rem;
            font-family: monospace;
        }

        .run-btn {
            background-color: var(--accent);
            color: white;
            border: none;
            padding: 0.5rem 1.5rem;
            border-radius: 0.375rem;
            font-weight: 600;
            cursor: pointer;
            transition: background-color 0.2s;
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }

        .run-btn:hover {
            background-color: #2563EB;
        }
        
        .save-btn {
            background-color: transparent;
            border: 1px solid var(--accent);
            color: var(--accent);
            padding: 0.5rem 1rem;
            border-radius: 0.375rem;
            cursor: pointer;
            font-weight: 600;
        }
        
        .save-btn:hover {
            background-color: rgba(59, 130, 246, 0.1);
        }

        .theme-btn {
            background: none;
            border: none;
            cursor: pointer;
            font-size: 1.25rem;
            padding: 0.5rem;
            border-radius: 50%;
            transition: background-color 0.2s;
        }

        .theme-btn:hover {
            background-color: var(--border);
        }
        .resizer {
            height: 8px;
            background-color: var(--bg-primary);
            cursor: row-resize;
            display: flex;
            justify-content: center;
            align-items: center;
            transition: background-color 0.2s;
        }

        .resizer:hover {
            background-color: var(--accent);
        }

        .resizer::after {
            content: "•••";
            color: var(--text-secondary);
            font-size: 10px;
            letter-spacing: 2px;
        }
    </style>
</head>
<body>
    <header>
        <h1>DevCLI Python Compiler</h1>
        <div class="header-controls">
             <button class="theme-btn" onclick="toggleTheme()" id="theme-icon">Theme</button>
             <button class="save-btn" onclick="saveCode()">Save</button>
             <button class="run-btn" onclick="runCode()">Run Code</button>
        </div>
    </header>
    
    <div class="main-container" id="main-split">
        <!-- Editor Section -->
        <div class="pane top-pane" id="top-pane" style="flex: 1; min-height: 100px;">
            <div class="panel-header">
                <input type="text" id="filename" class="filename-input" value="main.py" placeholder="/path/to/script.py">
            </div>
            <textarea id="code" spellcheck="false"># Install packages in the terminal below!
# Example: pip install numpy

import sys

def main():
    print("Python Version:", sys.version)
    print("Hello from DevCLI!")

if __name__ == "__main__":
    main()
</textarea>
        </div>

        <!-- Resizer -->
        <div class="resizer" id="dragMe"></div>

        <!-- Terminal / Output Section -->
        <div class="pane bottom-pane" id="bottom-pane" style="flex: 1; min-height: 100px;">
            <div class="panel-header">
                <button class="tab-btn active" onclick="switchTab('output')">Output</button>
                <button class="tab-btn" onclick="switchTab('terminal')">Terminal</button>
            </div>
            
            <div id="output-view" class="terminal-container">
                <div id="output-log" class="terminal-output">Ready...</div>
            </div>

            <div id="terminal-view" class="terminal-container" style="display: none;">
                <div id="terminal-log" class="terminal-output">Welcome to DevCLI Terminal. Using local shell.</div>
                <div class="terminal-input-line">
                    <span>$</span>
                    <input type="text" class="terminal-input" id="term-input" autocomplete="off">
                </div>
            </div>
        </div>
    </div>

    <script>
        // Resizer Logic
        const resizer = document.getElementById('dragMe');
        const topPane = document.getElementById('top-pane');
        const bottomPane = document.getElementById('bottom-pane');
        const container = document.getElementById('main-split');
        let isResizing = false;

        resizer.addEventListener('mousedown', function(e) {
            isResizing = true;
            document.body.style.cursor = 'row-resize';
            e.preventDefault();
        });

        document.addEventListener('mousemove', function(e) {
            if (!isResizing) return;
            
            const containerRect = container.getBoundingClientRect();
            const pointerRelativeY = e.clientY - containerRect.top;
            
            // Constrain
            const minSize = 100;
            const newTopHeight = Math.max(minSize, Math.min(pointerRelativeY, containerRect.height - minSize));
            
            const totalHeight = containerRect.height;
            const topFlex = newTopHeight / totalHeight;
            const bottomFlex = 1 - topFlex;

            topPane.style.flex = topFlex;
            bottomPane.style.flex = bottomFlex;
        });

        document.addEventListener('mouseup', function(e) {
            if (isResizing) {
                isResizing = false;
                document.body.style.cursor = '';
            }
        });

        function switchTab(tab) {
            document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
            event.target.classList.add('active');
            
            document.getElementById('output-view').style.display = tab === 'output' ? 'flex' : 'none';
            document.getElementById('terminal-view').style.display = tab === 'terminal' ? 'flex' : 'none';
            
            if (tab === 'terminal') {
                document.getElementById('term-input').focus();
            }
        }

        async function runCode() {
            switchTab('output');
            const code = document.getElementById('code').value;
            const log = document.getElementById('output-log');
            
            log.textContent = "Running...";
            
            try {
                const response = await fetch('/run', {
                    method: 'POST',
                    body: code
                });
                const result = await response.json();
                
                if (result.error) {
                    log.style.color = 'var(--error)';
                    log.textContent = result.output + "\nError: " + result.error;
                } else {
                    log.style.color = 'var(--success)';
                    log.textContent = result.output;
                }
            } catch (e) {
                log.style.color = 'var(--error)';
                log.textContent = "Error: " + e.message;
            }
        }
        
        async function saveCode() {
            const code = document.getElementById('code').value;
            const filename = document.getElementById('filename').value;
             
             if (!filename) {
                 alert("Filename required"); 
                 return;
             }
             
            try {
                const response = await fetch('/save', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ filename: filename, content: code })
                });
                
                if (response.ok) {
                    alert("Saved successfully to " + filename);
                } else {
                    const txt = await response.text();
                    alert("Error saving file: " + txt);
                }
            } catch (e) {
                alert("Network error: " + e.message);
            }
        }

        // Terminal Logic
        const termInput = document.getElementById('term-input');
        const termLog = document.getElementById('terminal-log');

        termInput.addEventListener('keydown', async (e) => {
            // Ctrl+C support
            if (e.ctrlKey && e.key === 'c') {
                e.preventDefault();
                termLog.textContent += "^C\n";
                try {
                    await fetch('/cancel', { method: 'POST' });
                } catch (e) {
                    console.error("Failed to cancel", e);
                }
                return;
            }

            if (e.key === 'Enter') {
                const cmd = termInput.value;
                if (!cmd.trim()) return;

                termLog.textContent += "\n$ " + cmd + "\n";
                termInput.value = '';
                termInput.disabled = true;

                try {
                    const response = await fetch('/terminal', {
                        method: 'POST',
                        body: cmd
                    });
                    const result = await response.json();
                    termLog.textContent += result.output + "\n";
                    
                    // Auto-scroll
                    const container = document.getElementById('terminal-view');
                    container.scrollTop = container.scrollHeight;
                } catch (e) {
                    termLog.textContent += "Error executing command.\n";
                }
                
                termInput.disabled = false;
                termInput.focus();
            }
        });

        // Editor Tab support
        document.getElementById('code').addEventListener('keydown', function(e) {
            if (e.key == 'Tab') {
                e.preventDefault();
                var start = this.selectionStart;
                var end = this.selectionEnd;
                this.value = this.value.substring(0, start) + "    " + this.value.substring(end);
                this.selectionStart = this.selectionEnd = start + 4;
            }
        });

        // Theme Logic
        function toggleTheme() {
            const body = document.body;
            const icon = document.getElementById('theme-icon');
            if (body.getAttribute('data-theme') === 'dark') {
                body.removeAttribute('data-theme');
                icon.textContent = 'Theme';
            } else {
                body.setAttribute('data-theme', 'dark');
                icon.textContent = 'Theme';
            }
        }
    </script>
</body>
</html>
`

// Global state for the web server
var (
	serverStarted bool       // Tracks if the server is currently running
	serverPort    string     // The port number the server is listening on
	currentDir    string     // The working directory for terminal commands
	activeCmd     *exec.Cmd  // Currently running command (for cancellation)
	activeMu      sync.Mutex // Protects access to activeCmd from multiple threads
)

// StartServer launches the web-based Python compiler on the specified port
func StartServer(port string) error {
	if serverStarted {
		if serverPort == port {
			return nil // Server already running on the correct port, nothing to do
		}
		// Attempting to start on a different port while server is running
		// is not supported in this implementation
		return fmt.Errorf("server already running on port %s", serverPort)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(htmlContent))
	})

	// Handle Ctrl+C from the web terminal
	mux.HandleFunc("/cancel", func(w http.ResponseWriter, r *http.Request) {
		activeMu.Lock()
		defer activeMu.Unlock()
		if activeCmd != nil && activeCmd.Process != nil {
			// Terminate the currently running process
			activeCmd.Process.Kill()
			// The process runner will clean up and return an error
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

		// Extract the filename from the request
		filename := payload.Filename
		if filename == "" {
			http.Error(w, "Filename required", http.StatusBadRequest)
			return
		}

		// Create parent directories if they don't exist
		dir := filepath.Dir(filename)
		if dir != "." && dir != "/" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				http.Error(w, "Failed to create directory: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}

		err := os.WriteFile(filename, []byte(payload.Content), 0644)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
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

		// Execute the Python code and capture output
		output, err := runPython(string(body))

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
		if err != nil {
			// Terminal commands show errors in the output itself, not as separate field
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	serverStarted = true
	serverPort = port

	// Use IP address instead of "localhost" to avoid DNS issues on Windows
	addr := "127.0.0.1:" + port
	fmt.Printf("Starting local compiler server at http://%s\n", addr)

	// Start the HTTP server (this call blocks until the server stops)
	// Note: If the port is already in use, ListenAndServe will fail immediately.
	// Checking the port beforehand would create a race condition, so we let
	// it fail naturally and handle the error.
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
