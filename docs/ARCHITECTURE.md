# DevCLI v1.1.0 Architecture & Package Map

DevCLI is built using Go with a modular, decoupled package architecture:

---

## Directory Structure & Package Map

```
devcli/
├── main.go                       # Main CLI entry point & Cobra subcommands
├── install_linux.sh               # Universal Linux single-command installer
├── install.sh / install.ps1      # Platform setup scripts
├── assets/                        # Embedded binary assets (logo png)
├── docs/                          # Comprehensive Markdown documentation
├── pkg/                           # Public utilities & shared helpers
│   ├── project/                   # Project models
│   └── utils/                     # Shell & OS execution helpers
└── internal/                      # Private internal packages
    ├── ai/                        # AI Provider interface & message models
    │   └── providers/             # Ollama, Gemini, OpenAI, Claude, HF implementations
    ├── aicommit/                  # AI conventional Git commit auto-generator
    ├── config/                    # Viper configuration management (`~/.devcli.yaml`)
    ├── devtools/                  # Docker container & HTTP API client engines
    ├── project/                   # Project generator, templates & scaffolding
    ├── timemachine/               # Git blame & line risk analyzer
    ├── tui/                       # Bubble Tea dual-pane TUI dashboards
    │   ├── statusbar.go           # Real-time system RAM, Git, Venv status bar
    │   ├── docker_dashboard.go    # Docker container list & logs viewport
    │   ├── api_client_dashboard.go# API & HTTP Client Playground
    │   └── ...                    # Other TUI sub-models
    ├── updater/                   # GitHub release self-updater
    ├── venv/                      # Python venv & node_modules manager
    └── web/                       # Embedded HTTP Multi-Language Compiler Server
        └── static/                # Browser web IDE assets
```

---

## Key Core Dependencies

* **Bubble Tea (`github.com/charmbracelet/bubbletea`)**: TUI Elm architecture event loop.
* **Lip Gloss (`github.com/charmbracelet/lipgloss`)**: Terminal layouts, borders, and color styling.
* **Bubbles (`github.com/charmbracelet/bubbles`)**: List, Textinput, Viewport, and Spinner components.
* **Glamour (`github.com/charmbracelet/glamour`)**: Terminal Markdown rendering.
* **Cobra (`github.com/spf13/cobra`)**: CLI command parsing.
* **Viper (`github.com/spf13/viper`)**: Configuration loading and persistence.
