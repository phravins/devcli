<div align="center">
  <br>

  [![Version](https://img.shields.io/badge/Version-v1.1.0-50FA7B?style=for-the-badge&logo=go&logoColor=black)](https://github.com/phravins/devcli) [![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/) [![GitHub](https://img.shields.io/badge/GitHub-Repo-181717?style=for-the-badge&logo=github&logoColor=white)](https://github.com/phravins/devcli) [![Linux](https://img.shields.io/badge/Linux-FCC624?style=for-the-badge&logo=linux&logoColor=black)](https://kernel.org/) [![Windows](https://img.shields.io/badge/Windows-0078D6?style=for-the-badge&logo=windows&logoColor=white)](https://www.microsoft.com/windows/) [![macOS](https://img.shields.io/badge/macOS-000000?style=for-the-badge&logo=apple&logoColor=white)](https://apple.com/macos/)
</div>

# DevCLI v1.1.0 - Developer Command Line Interface

**DevCLI** is a terminal-based development workspace and suite of developer tools in a single unified interface. Built using **Go** and Charm's **Bubble Tea / Lip Gloss** framework, DevCLI provides a fast, responsive, dual-pane Terminal User Interface across Linux, macOS, and Windows.

---

## ⚡ Quick Start: One-Liner Linux Installation

Install DevCLI globally on any Linux desktop environment (GNOME, KDE, XFCE, Cinnamon, etc.) with a single command:

```bash
curl -fsSL https://raw.githubusercontent.com/phravins/devcli/main/install_linux.sh | bash
```

> **No Go Pre-installed?** No problem! The script automatically installs Go in user space (`~/.local/go`), sets up `devcli` in `~/.devcli/bin`, creates a Desktop Application Launcher (`.desktop` entry), and registers shell PATH exports automatically.

---

## 🔥 Key Features in v1.1.0

* **Dual-Pane TUI Interface**: Framed navigation menu alongside a live workspace info card displaying Git status, memory consumption, virtual environments, and AI provider status.
* **Live System & Git Status Bar**: Persistent header bar displaying active Git branch (`git:main`), memory usage in MB, active Python `venv`, and connected AI model.
* **🐳 Docker Container Dashboard**: Inspect running & stopped Docker containers, start/stop/restart containers (`s`/`r`), and stream live container logs in a scrollable viewport (`l`).
* **🌐 API & HTTP Client Playground**: Built-in TUI Postman alternative for testing `GET`, `POST`, `PUT`, `DELETE`, and `PATCH` endpoints with formatted JSON views and latency counters (ms).
* **🤖 AI Assistant & Conventional Git Commit Auto-Generator**: Connects to Ollama, Gemini, OpenAI, Claude, or HuggingFace. Run `devcli ai commit` to generate conventional git commit messages directly from your `git diff`.
* **💻 Multi-Language Web IDE & Compiler**: Embedded HTTP server (`http://127.0.0.1:8080`) supporting online compilation for **Python**, **JavaScript (Node.js)**, **Go**, **Rust**, and **C/C++**.
* **📂 Project Scaffolder & File Manager**: Instant scaffolding for Go, Python, Node, React, and FastAPI projects with Git initialization, alongside a built-in terminal file explorer.
* **🕰️ Code Time Machine**: Line-by-line Git blame examiner and history timeline viewer.

---

## 📚 Table of Contents

* [Installation Guide](docs/INSTALLATION.md)
* [Detailed Features Breakdown](docs/FEATURES.md)
* [Usage & Shortcuts Guide](docs/USAGE.md)
* [Configuration & AI Keys](docs/CONFIGURATION.md)
* [Architecture & Code Structure](docs/ARCHITECTURE.md)
* [Contributing](CONTRIBUTING.md)
* [License](LICENSE)

---

## 🛠️ Subcommands Cheat Sheet

| Command | Description |
| :--- | :--- |
| `devcli` | Launch the main dual-pane interactive TUI workspace |
| `devcli install` | Deploy binary globally and setup Linux Desktop entry |
| `devcli ai chat` | Launch terminal AI Chat assistant |
| `devcli ai commit` | Auto-generate AI git commit message from staged changes |
| `devcli start [name] [stack]` | Scaffold a new project (Go, Python, Node, etc.) |
| `devcli editor [file]` | Launch terminal code editor |
| `devcli file` | Open interactive file manager |
| `devcli timemachine [file]` | Open Code Time Machine Git history analyzer |
| `devcli update` | Check for updates and update DevCLI binary |

---

## 📄 License

DevCLI is licensed under the Apache License 2.0 - see [LICENSE](LICENSE) for details.
