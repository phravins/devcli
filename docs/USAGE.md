# DevCLI v1.1.0 Usage & Keyboard Shortcuts

## Launching DevCLI

Launch the main interactive dual-pane TUI dashboard:

```bash
devcli
```

Or run direct subcommands:

```bash
devcli ai commit             # Auto-generate AI Git commit message from staged changes
devcli start [name] [stack]  # Scaffold a new project (Go, Python, Node, etc.)
devcli editor [file]         # Open file in terminal code editor
devcli file                  # Open file manager
devcli timemachine [file]    # Launch Git Code Time Machine
devcli install               # Configure Linux desktop launcher entry & icon
devcli update                # Self-update DevCLI binary
```

---

## Global Navigation Shortcuts

* `↑ / ↓` or `k / j`: Navigate lists and menus
* `Enter`: Open selected feature or confirm action
* `Esc` or `q`: Return to previous menu or exit
* `?`: View commands and hotkey help
* `Ctrl+C`: Force quit

---

## Feature-Specific Shortcuts

### 🐳 Docker Dashboard
* `s`: Start or Stop selected container
* `r`: Restart selected container
* `l` or `Enter`: Open live container logs viewer
* `Esc` or `q`: Return to container list

### 🌐 API & HTTP Client Playground
* `↑ / ↓`: Select HTTP Method (`GET`, `POST`, `PUT`, `DELETE`, `PATCH`)
* `Enter`: Submit URL / Request Body
* `r`: Resend current API request
* `Esc` or `q`: Clear response or return to method selection

### 📂 Project Scaffolder
* `B`: Backup project
* `D`: Delete project history record
* `?`: Help

### 🐍 Virtual Environment Wizard
* `N`: Create new Python `venv`
* `S`: Scan workspace for environments
* `Y`: Sync `requirements.txt` (`pip freeze`)
* `C`: Clone environment
* `D`: Delete environment

### 🗂️ File Manager
* `C`: Copy file/directory
* `M`: Move / Rename file
* `D`: Delete file
* `E`: Edit file in terminal editor
* `N`: Create new file
* `H`: Toggle hidden files (`.env`, `.git`)

### ✏️ Built-in Code Editor
* `Ctrl+S`: Save file
* `Ctrl+R`: Run code
* `Ctrl+N`: New file
* `Ctrl+C`: Exit editor
