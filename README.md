# DevCLI - Developer Command Line Interface

DevCLI is a terminal-based development workspace that consolidates essential
developer tools into a single unified interface. It manages projects, files,
virtual environments, and provides AI-powered assistance without requiring
you to leave the command line.

The application is built using Go and the Bubble Tea framework, providing a
fast and responsive terminal user interface that works across all major
operating systems.

## What is DevCLI?

DevCLI serves as a central hub for common development tasks. Instead of grouping scattered scripts or remembering complex CLIs, DevCLI provides a unified, interactive workspace containing a suite of powerful internal features:

*   **Project Manager**: Scaffolding, templates, and history tracking.
*   **Task Runner**: One-click execution of build, test, and lint commands for any language (Go, Python, Node, Rust, C++).
*   **Virtual Environment Wizard**: Centralized management of Python venvs and Node modules.
*   **Dev Server**: Auto-detecting live reload servers for web development.
*   **Smart File Creator**: Instant generation of Dockerfiles, .env, Makefiles, and CI/CD configs.
*   **Boilerplate Generator**: Instant code snippets and architectural patterns.
*   **Snippet Library**: Your personal vault for reusable code blocks.
*   **AI Assistant**: Built-in chat for coding help, debugging, and explanations.
*   **File Manager & Editor**: Keyboard-driven filesystem navigation and quick editing.
*   **Auto-Update System**: Keeps your languages and tools current.

The tool is particularly useful for developers who:
  - Work with multiple programming languages and frameworks
  - Manage several projects simultaneously
  - Need quick access to project scaffolding and templates
  - Want to maintain consistent development environments
  - Prefer keyboard navigation over graphical tools

## System Requirements

**Operating System:**
*   **Windows:** Windows 10, Windows 11 (or newer)
*   **Linux:** Ubuntu, Debian, Fedora, Arch Linux (or any modern distribution)
*   **macOS:** macOS 11 (Big Sur) or newer

**Hardware:**
*   **Processor:** Dual-core processor (Intel/AMD x64 or ARM64)
*   **RAM:** 4GB minimum (8GB+ recommended for development workloads)
*   **Storage:** 200MB free disk space for application components

**Software:**
*   Go 1.21 or higher (for building from source)
*   Python 3.8 or newer (required for virtual environment features)
*   Node.js (optional, needed for JavaScript project templates)
*   Terminal with Unicode support (for proper rendering)

## Dependencies

DevCLI is built on top of several robust Go packages that provide its
functionality. All dependencies are automatically managed through Go modules.

CORE DEPENDENCIES

  - Bubble Tea (v1.3.4) - Terminal User Interface framework
    Provides the foundation for the interactive TUI experience
    
  - Bubbles (v0.21.0) - TUI components library
    Pre-built UI components (lists, text inputs, viewports, etc.)
    
  - Lipgloss (v1.1.1) - Terminal styling library
    Handles colors, borders, and layout styling
    
  - Glamour (v0.10.0) - Markdown renderer for terminals
    Renders markdown content in chat and help screens
    
  - Cobra (v1.8.0) - CLI framework
    Command-line interface structure and flags
    
  - Viper (v1.21.0) - Configuration management
    Handles config.yaml reading and writing

ADDITIONAL PACKAGES

  - Chroma (v2.20.0) - Syntax highlighting
    Provides code syntax highlighting in the editor
    
  - Fuzzy (v0.1.1) - Fuzzy string matching
    Powers the file manager's fuzzy search feature
    
  - YAML v2 (v2.4.0) - YAML parsing
    Configuration file handling

INSTALLATION NOTE

When you install DevCLI using 'go install' or build from source, Go will
automatically download and install all required dependencies. No manual
package installation is needed.

To see the complete dependency tree:

```bash
go mod graph
```

To update dependencies to latest versions:

```bash
go get -u ./...
go mod tidy
```


Installation
------------

METHOD 1: Automated Installation (Windows - Recommended)

For a complete "no-hassle" installation that sets up Go and DevCLI automatically:

1.  Download the [setup_devcli.bat](setup_devcli.bat) script.
2.  Right-click the file and select **"Run as administrator"**.
3.  The script will:
    -   Check if Go is installed (and automatically download/install it if missing).
    -   Install the latest version of DevCLI.
    -   Configure your system PATH.

Once finished, simply restart your terminal and run:

```bash
devcli
```

METHOD 2: Single Command Installation (If Go is already installed)

Install DevCLI directly using the `go install` command:

```bash
go install github.com/phravins/devcli@latest
```

This will download, build, and install the DevCLI binary to your `$GOPATH/bin`
directory. Ensure that `$GOPATH/bin` is in your system PATH to access the
`devcli` command from anywhere.

To verify the installation:

```bash
devcli --version
```

METHOD 3: Building from Source

Clone the repository and build manually:

```bash
git clone https://github.com/phravins/devcli.git
cd devcli
go build -o devcli.exe .
```

After building, move the executable to a directory in your PATH to access
it from anywhere in your terminal.


Core Features
-------------

PROJECT CREATION AND MANAGEMENT

DevCLI provides project scaffolding tools that generate complete project
structures from templates. This eliminates the need to manually create
directory structures and configuration files for new projects.

Key capabilities:
  - Create projects from predefined templates for Go, Python, Node.js,
    React, and other popular frameworks
  - Smart project naming with automatic incrementing for duplicate names
  - Customizable project location with path validation
  - Project history tracking with automatic cleanup of old entries
  - Backup functionality to safely archive existing projects

The project creator includes installation automation. After generating
project files, it automatically runs the appropriate package manager
commands (npm install, pip install, etc.) and displays real-time output.


VIRTUAL ENVIRONMENT WIZARD

A comprehensive tool for managing Python virtual environments and Node.js
node_modules across your entire workspace. It provides centralized control
over dependency management.

Key capabilities:
  - Recursive scanning that finds all virtual environments and node_modules
    directories within your workspace, regardless of nesting depth
  - Requirements synchronization that generates requirements.txt files from
    installed packages in Python environments
  - Environment cloning that replicates entire virtual environment setups,
    including all dependencies and their versions
  - Cleanup tools to identify and remove unused or stale environments
  - Package inspection showing all installed packages with version numbers
  - Dependency conflict detection across multiple environments

The wizard can scan thousands of nested directories efficiently and provides
visual feedback during long-running operations.


DEV SERVER LAUNCHER

An intelligent development server manager that automatically detects the
correct server command for your project type and captures all output in
a searchable log viewer.

Key capabilities:
  - Automatic detection of project framework (detects package.json scripts,
    go.mod files, Python web frameworks, etc.)
  - Live log streaming with colored output preservation
  - Log filtering by log level (info, warn, error) or custom patterns
  - Full-text search across server logs
  - Auto-scroll toggle for following new log entries
  - Server source switching for full-stack projects (frontend/backend)
  - Clean server shutdown handling

The dev server feature eliminates the need to remember project-specific
commands like "npm run dev", "python manage.py runserver", or "go run main.go".


BOILERPLATE CODE GENERATOR

A code snippet and template generator that provides instant access to common
code patterns and full project architectures.

Key capabilities:
  - Ready-to-use code snippets for CRUD APIs, authentication systems,
    database connections, and common algorithms
  - Multi-language support (Go, Python, JavaScript, Java, Rust, C, C++, Zig)
  - Project architecture generation for complete application structures
  - Template saving that lets you convert existing projects into reusable
    templates for future use
  - Syntax highlighting in snippet preview
  - Direct file creation from snippets

The generator includes production-ready code that follows best practices
for each language and framework.


BONUS FEATURES

Additional productivity tools accessible through a dedicated menu:

Project Dashboard:
  - Scans workspace and displays all projects with metadata
  - Detects technology stack automatically (identifies languages, frameworks,
    and tools used in each project)
  - Shows project status (Active, Broken, Archived based on modification time)
  - Displays last modified date and project size
  - Sortable by name, date, or status

Task Runner:
  - Auto-detects available tasks from package.json, Makefile, Cargo.toml,
    and other project files
  - One-click execution of build, test, lint, and format commands
  - Live output streaming during task execution
  - Task categorization and organization
  - Support for npm/yarn, Go tools, Python tools, Rust cargo, and Make

Smart File Creator:
  - Generates common configuration files with best practices built-in
  - Supported files: .env, .gitignore, Dockerfile, docker-compose.yml,
    .editorconfig, Makefile, GitHub Actions workflows
  - Language-specific customization (generates appropriate .gitignore
    patterns for your language)
  - File preview before saving
  - Production-ready templates for Docker multi-stage builds

Snippet Library:
  - Personal code snippet storage with search and categorization
  - Metadata tracking (language, category, tags, creation date)
  - Search functionality across all snippets
  - JSON-based storage at ~/.devcli/snippets.json
  - Includes default snippets for common patterns

AI Assistant:
  - Natural language interface for code generation and explanations
  - Integrates with existing AI provider configuration
  - Multi-turn conversations support
  - Context-aware code suggestions


AI INTEGRATION

DevCLI includes a built-in AI chat interface that supports multiple backends.
This allows you to get coding assistance, explain complex code, or generate
boilerplate without leaving your terminal.

Key capabilities:
  - Provider support for Ollama (local), Hugging Face, OpenAI, Anthropic,
    and Google Gemini
  - Configurable system prompts for customizing AI behavior
  - Model selection with support for different model sizes and capabilities
  - Chat history with scroll-back
  - Streaming responses for immediate feedback
  - Local AI execution tools for privacy-conscious developers

The AI integration is designed to work offline when using local models,
making it suitable for air-gapped or restricted network environments.


FILE MANAGER

A keyboard-driven file explorer designed for developers who prefer not to
leave the terminal. It provides full file system navigation and manipulation.

Key capabilities:
  - Tree-style directory navigation with arrow keys
  - Fuzzy search for locating files quickly
  - Standard operations: copy, move, rename, delete, create
  - File editing integration with the built-in editor
  - Multi-drive support for Windows systems
  - Hidden file toggle
  - Permissions and size display

The file manager emphasizes speed and keyboard efficiency, allowing
experienced users to perform file operations faster than with a mouse.


BUILT-IN CODE EDITOR

A lightweight terminal-based code editor optimized for quick edits and
Python script testing. While not intended to replace full IDEs, it serves
well for config file editing and simple scripts.

Key capabilities:
  - Syntax highlighting for Python code
  - Direct code execution (run Python scripts with Ctrl+R)
  - Multi-language support (Java, C++, C, Rust, Zig, C#, JavaScript, Go)
  - Integrated terminal for running system commands
  - File save functionality
  - Line numbers and cursor position display

The editor runs Python code in the same environment as DevCLI, making
it useful for testing snippets or running utility scripts.


AUTO-UPDATE SYSTEM

A centralized update management interface for keeping development tools
current.

Key capabilities:
  - Programming language version checking (Go, Python, Node.js, Java,
    Rust, Zig, C, C++)
  - Installation path detection and display
  - DevCLI self-update with Git integration
  - AI provider API key management and updates
  - Status display for all installed tools

This feature helps maintain an up-to-date development environment without
manually checking each tool's version.


Usage
-----

INTERACTIVE MODE

Launch the main dashboard:

```bash
devcli
```

Use arrow keys to navigate menus and Enter to select. Press Esc or Q to
go back through menus or exit the application.

DIRECT SUBCOMMANDS

Access specific features directly:

```bash
devcli dev          # Open project management tools
devcli file         # Launch file manager
devcli ai           # Start AI chat session
devcli editor FILE  # Open file in built-in editor
```

Direct subcommands are useful for scripting or when you know exactly which
tool you need.


Configuration
-------------

DevCLI stores configuration in ~/.devcli/config.yaml (or equivalent on Windows).

Important configuration options:

  provider        AI backend selection
  api_key         Credentials for cloud AI services
  model          Default model for AI chat
  workspace      Default directory for new projects
  ollama_url     Local Ollama server address

The configuration file is created on first run with sensible defaults.


Keyboard Shortcuts
------------------

GLOBAL NAVIGATION

  Arrow Up/Down     Navigate through lists and menus
  Enter            Confirm selection or execute action
  Esc or Q         Return to previous screen or exit
  Ctrl+C           Force quit application

FEATURE-SPECIFIC SHORTCUTS

Project Tools:
  B               Create backup of selected project
  D               Delete history entry
  ?               Display help information

Virtual Environment Wizard:
  N               Create new virtual environment
  S               Scan workspace for environments
  Y               Sync packages to requirements.txt
  C               Clone selected environment
  D               Delete environment

Dev Server:
  S               Start/stop server
  F               Toggle log filters
  Slash (/)       Search logs
  A               Toggle auto-scroll
  C               Clear log buffer
  ?               Show help

File Manager:
  C               Copy file or directory
  M               Move/rename
  D               Delete
  E               Edit with built-in editor
  N               Create new file
  H               Toggle hidden files

Editor:
  Ctrl+R          Run code
  Ctrl+S          Save file
  Ctrl+N          New file
  Ctrl+H          Toggle help
  Ctrl+C          Exit editor

Each feature displays its available shortcuts in the footer area of the
interface.


Architecture
------------

DevCLI is structured as a modular application with clear separation between
UI and business logic:

  cmd/            Command-line entry points
  internal/
    ai/           AI provider implementations
    boilerplate/  Code template system
    config/       Configuration management
    devserver/    Server launch and log parsing
    fileops/      File system operations
    history/      Project history tracking
    project/      Project creation and management
    projectdash/  Project analysis tools
    smartfile/    Configuration file generation
    snippets/     Code snippet storage
    taskrunner/   Build tool integration
    templates/    Project scaffolding templates
    tui/          Terminal UI components
    venv/         Virtual environment management
    web/          Web server utilities

The internal package structure prevents external dependencies on implementation
details, making the codebase more maintainable.


Contributing
------------

Contributions are welcome. Please ensure code follows Go conventions and
includes appropriate tests. Use gofmt for code formatting.

License
-------

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

Support
-------

For issues, questions, or feature requests, please use the GitHub issue
tracker at https://github.com/phravins/devcli/issues
