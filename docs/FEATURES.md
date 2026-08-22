# DevCLI v1.1.0 Feature Overview

DevCLI provides a comprehensive suite of terminal development workspace tools:

---

## 1. Dual-Pane TUI Workspace & Live Status Bar
* **Framed Dual-Pane Layout**: Navigation menu on the left pane and dynamic live workspace stats & feature details card on the right pane.
* **Persistent Live Status Bar**: Displays active Git branch (`git:main`), memory usage in MB, active Python `venv`, and connected AI provider status.
* **Vibrant Lip Gloss Color Themes**: High-contrast Dracula, Neon, and Nord styling with status badges and glowing active cursor pills.

---

## 2. 🐳 Docker & Container Dashboard
Native Docker container management inside the terminal:
* **Container Inspection**: View all running and stopped Docker containers with image tags, status badges, and IDs.
* **State Controls**: Start (`s`), stop (`s`), and restart (`r`) containers directly from the list.
* **Live Container Logs**: Stream container stdout/stderr logs into a scrollable TUI viewport (`l` or `Enter`).

---

## 3. 🌐 API & HTTP Client Playground (Postman Alternative in TUI)
Interactive API client built into DevCLI:
* **HTTP Method Selector**: Send `GET`, `POST`, `PUT`, `DELETE`, and `PATCH` requests.
* **JSON Payload Editor**: Input custom headers and JSON request bodies.
* **Status Badges & Latency Counter**: Real-time response status badges (e.g. `200 OK` in emerald green, `404` in red) and latency timer in milliseconds.
* **Formatted JSON Viewport**: Pretty-printed JSON response formatting with scrollable viewport.

---

## 4. 💻 Multi-Language Web IDE & Compiler Server
Embedded HTTP server running at `http://127.0.0.1:8080`:
* **Multi-Language Runner**: Supports online code execution for **Python**, **JavaScript (Node.js)**, **Go**, **Rust**, and **C / C++**.
* **Browser IDE**: Glassmorphic UI with integrated browser terminal output, file saving, user auth, and mock Google Drive sync.

---

## 5. 🤖 AI Assistant & Conventional Git Commit Auto-Generator
* **Multi-Provider AI Chat**: Connects to Ollama, Google Gemini, OpenAI, Anthropic Claude, or HuggingFace.
* **`devcli ai commit` Subcommand**: Analyzes your `git diff`, queries AI to auto-generate conventional commit messages (`feat: ...`, `fix: ...`), and commits changes upon confirmation.
* **Timeout-Protected Execution**: Fast 8-second timeout preventing UI stalls if AI providers are slow or offline.

---

## 6. 📂 Project Manager & Scaffolder
Scaffold full project structures with instant Git initialization:
* **Supported Stacks**: Go Fiber API, Python FastAPI/Flask, Node.js Express/React, and custom templates.
* **Automated Setup**: Generates structured README, `.gitignore`, dependency installers (`npm install`, `pip install`), and git repository setup.

---

## 7. 🗂️ Terminal File Manager & Built-in Code Editor
* **File Manager**: Keyboard-driven file browser with tree navigation, search, read/write, rename, and delete actions.
* **Built-in Editor**: Terminal code editor with syntax highlighting, line numbers, and file saving.

---

## 8. 🐍 Virtual Environment Wizard
Centralized Python `venv` and `node_modules` manager:
* Recursive scanning for virtual environments in your workspace.
* Verification, cloning, package freezing (`pip freeze`), and deletion.

---

## 9. 🕰️ Code Time Machine
Git-powered code evolution tracker:
* Interactive line-by-line Git blame examiner.
* Risk analyzer highlighting late-night commits and large refactor risks.

---

## 10. 🔄 Auto-Update Center
* System language version scanner for Go, Python, Node, Java, Rust, Zig, and C/C++.
* Self-updating mechanism to pull latest commits and rebuild the DevCLI executable.
