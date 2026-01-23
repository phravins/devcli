package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Help content for all features (without emojis)
const (
	ProjectToolsHelp = `
          PROJECT TOOLS - Help & Usage Guide                  


OVERVIEW
Project Tools helps you create, manage, and maintain software projects
with pre-configured templates and automated setup workflows.

KEYBOARD SHORTCUTS
Key         Description
         
?           Show this help
Esc         Go back / Return to previous menu
Up/Down     Navigate through lists
Enter       Select / Confirm action
b           Backup selected project (in project list)
d           Delete history entry (in history view)

HOW TO USE

1. PROJECT CREATION
   • Select "Project Creation & Management" from main menu
   • Choose "+ New Project" to start wizard
   • Pick a template (Go, Python, Web, Full-Stack, etc.)
   • Enter project name (auto-suggested based on template)
   • Specify parent directory path
   • Wait for automated setup and dependency installation

2. PROJECT TEMPLATES
   Available templates include:
   • Go Web Server - Basic HTTP server with routing
   • Python Flask - Web framework with virtual environment
   • Full-Stack - React frontend + Python backend
   • Node.js Express - JavaScript web server
   • And more...

3. PROJECT BACKUP
   • Navigate to project list
   • Select the project you want to backup
   • Press 'b' key
   • Enter destination path
   • Project will be copied with all files

4. PROJECT HISTORY
   • View all previously created projects
   • See creation timestamps and paths
   • Delete old entries with 'd' key
   • Auto-cleanup for entries older than 30 days

PROJECT STRUCTURE
Each project is created with:
• Template-specific file structure
• Dependency configuration files
• README with usage instructions
• Git initialization (if applicable)
• Virtual environment (for Python projects)

Press Esc to close this help`

	VenvWizardHelp = `
      VIRTUAL ENVIRONMENT WIZARD - Help & Usage Guide         


OVERVIEW
The Virtual Environment Wizard is your central dashboard for managing
Python virtual environments and Node.js modules across all your projects.
It prevents dependency conflicts by isolating project requirements.

KEYBOARD SHORTCUTS
Key         Description
         
?           Show this help
Esc/q       Go back to main menu
Up/Down     Navigate environment list
Enter       Open action menu for selected environment
n           Create new virtual environment
s           Scan system for environments
r           Refresh environment list
y           Sync packages (in action menu)
c           Clone environment (in action menu)
d           Delete environment (in action menu)

HOW TO USE

1. CREATE NEW ENVIRONMENT
   • Press 'n' from main list
   • Enter project directory path
   • Wizard creates .venv folder automatically
   • Python packages are installed
   • Environment is ready to use

2. SCAN FOR ENVIRONMENTS
   • Press 's' to scan any directory
   • Wizard recursively searches for:
     - Python venv folders (.venv, venv, env)
     - Node.js node_modules
     - Anaconda environments
   • Hidden folders are included
   • Results show size and location

3. SYNC PACKAGES
   • Select an environment and press Enter
   • Choose 'y' for Sync
   • Specify path for requirements.txt
   • Wizard runs 'pip freeze' automatically
   • File is saved with all dependencies

4. CLONE ENVIRONMENT
   • Select source environment and press Enter
   • Choose 'c' for Clone
   • Enter destination directory path
   • New .venv is created with identical packages
   • Perfect for replicating setups

5. DELETE ENVIRONMENT
   • Select environment and press Enter
   • Choose 'd' for Delete
   • Confirm deletion (cannot be undone)
   • Frees up disk space

WHAT IS A VIRTUAL ENVIRONMENT?
Virtual environments create isolated Python installations per project.
Example: Project A needs Django 3.0, Project B needs Django 4.0
Without venvs: Conflict! Only one version can be installed globally
With venvs: Each project has its own isolated Django version

ENVIRONMENT TYPES
Python venv - Standard Python virtual environment
Node modules - JavaScript package dependencies
Anaconda - Conda environment

Press Esc to close this help`

	DevServerHelp = `
          DEV SERVER - Help & Usage Guide                     


OVERVIEW
Dev Server automatically detects and launches development servers for
your projects with live log viewing and filtering capabilities.

KEYBOARD SHORTCUTS
Key         Description
         
?           Show this help
Esc/q       Go back to main menu
s           Start/Stop server
f           Toggle log filters
b           Toggle backend/frontend (Full-stack projects)
/           Search logs
a           Toggle auto-scroll
c           Clear logs
Up/Down     Scroll through logs

HOW TO USE

1. AUTO-DETECTION
   • Dev Server scans your project for:
     - package.json (Node.js/React)
     - go.mod (Go projects)
     - requirements.txt (Python/Flask)
     - Detects full-stack setups automatically

2. START SERVER
   • Press 's' to start detected server
   • Logs appear in real-time
   • Color-coded by severity:
     - Green: Success messages
     - Yellow: Warnings
     - Red: Errors
     - Blue: Info

3. LOG FILTERING
   • Press 'f' to toggle filters
   • Filter by log level (ERROR, WARN, INFO)
   • Only matching logs are displayed
   • Great for debugging specific issues

4. SEARCH LOGS
   • Press '/' to activate search
   • Type search query
   • Matching lines are highlighted
   • Navigate with arrow keys

5. FULL-STACK PROJECTS
   • Press 'b' to switch between:
     - Frontend server (React/Vue)
     - Backend server (Python/Go)
   • Each runs in separate process
   • Independent log streams

SUPPORTED FRAMEWORKS
• Node.js (npm start, npm run dev)
• React (npm start, vite)
• Python Flask (flask run, python app.py)
• Go (go run main.go)
• Express.js (node server.js)

Press Esc to close this help`

	FileManagerHelp = `
# File Manager - Help Guide

## Overview
A powerful keyboard-driven file explorer with fuzzy search across all drives, mouse support, and essential file operations.

## Keyboard Shortcuts
| Key | Description |
| :--- | :--- |
| **?** | Show this help |
| **Esc** | Go back / Clear search / Quit |
| **Up/Down** | Navigate file list |
| **Enter** | Open directory / Select file |
| **Tab** | Toggle global/local search mode |
| **Alt+M** | Move/Rename selected file |
| **Alt+C** | Copy selected file |
| **Alt+E** | Edit selected file |
| **Backspace** | Go up one directory (when search empty) |
| **Ctrl+L** | Customizable path search |

## Mouse Support
- **Click** on files/folders to select
- **Scroll Wheel** to navigate
- **Double-click** to open

## How to Use

### 1. Navigation
- Use arrow keys or mouse to browse.
- Press **Enter** on folders to open them.
- Press **Esc** to go back to the parent folder.

### 2. Global vs Local Search
- **Tab** toggles between modes.
- **Global Search**: Searches ALL indexed drives instantly.
- **Local Search**: Searches only the current directory.

### 3. File Operations
- **Alt+M**: Move or rename files across drives.
- **Alt+C**: Copy files to a new destination.
- **Alt+E**: Open text files in the built-in editor.

### 4. Drive Switching
- Available drives are shown in the footer.
- Navigate to the drive root (e.g., C:\, D:\) to switch.

---
*Press **Esc** to close this guide*`

	EditorHelp = `
# DevCLI Editor - Complete Guide

## Overview
A powerful, lightweight terminal-based IDE integrated into DevCLI. Featuring syntax highlighting, multi-language execution, and automatic compiler detection across all drives.

## Keyboard Shortcuts

### 1. Language Selection Menu
- **Arrow Keys / Mouse**: Navigate language list
- **Enter**: Select and open Editor
- **?**: Open this Help Guide
- **Esc / q**: Back to main dashboard

### 2. Code Editor Workspace
- **Arrow Keys / Mouse**: Move cursor / Scroll viewport
- **Ctrl + R**: **RUN** current code (Auto-detects language)
- **Ctrl + S**: **SAVE** current file (Prompts for path)
- **Ctrl + N**: **NEW FILE** (Clear current buffer)
- **Ctrl + O**: **FOCUS** Output Terminal
- **Ctrl + E**: **FOCUS** Code Editor
- **Ctrl + M**: **MAXIMIZE / MINIMIZE** Output area
- **Ctrl + P**: **SHELL** Prompt (Run system commands)
- **? / Ctrl + H**: **TOGGLE** this Help Guide
- **Esc**: **BACK** to Language Selection menu
- **Ctrl + C**: **EXIT** Editor immediately

## Compiler & Runtime Guide

DevCLI tries to find these automatically if they are in your PATH:

- **Python**: Requires Python 3.x. Ensure "Add to PATH" is checked.
- **Java**: Requires JDK 11+. Needs "javac" and "java".
- **C / C++**: Requires GCC or Clang (e.g., MinGW-w64 on Windows).
- **Rust**: Requires Rust toolchain (rustc, cargo).
- **Zig**: Requires Zig compiler from ziglang.org.
- **C#**: Requires .NET SDK 6.0+.
- **Web**: Automatically launches a local dev server.

---
*Press **Esc** or **Ctrl+H** to close this guide*`

	AIchatHelp = `
# AI Chat - Help Guide

## Overview
Chat with AI models from multiple providers including **Ollama** (local), **OpenAI**, **Google Gemini**, **Anthropic Claude**, and more.

## Keyboard Shortcuts
| Key | Description |
| :--- | :--- |
| **?** | Show this help |
| **Enter** | Send message |
| **Up/Down** | Scroll chat history |
| **Mouse Wheel** | Scroll history |
| **Esc / Ctrl+C** | Exit chat |

## How to Use

### 1. Sending Messages
- Type your question in the input box and press **Enter**.
- AI responses include **Markdown** rendering and **Code Syntax Highlighting**.

### 2. Provider & Model Setup
- To change settings, **Exit (Esc)** and go to the **Settings** menu.
- **Backends**: ollama, gemini, openai, claude, mistral, groq, etc.
- **Example Models**:
  - *Ollama*: llama3, mistral
  - *OpenAI*: gpt-4, gpt-3.5-turbo
  - *Gemini*: gemini-1.5-flash
  - *Claude*: claude-3-sonnet

### 3. Local AI (Ollama)
- **Free and Private**: No API key needed.
- Install from [ollama.ai](https://ollama.ai) and run **ollama pull llama3**.

---
*Press **Esc** to close this guide*`

	SettingsHelp = `
# Settings - Help & Usage Guide

## Overview
Configure AI providers, API keys, models, and other DevCLI settings. All settings are saved to **config.yaml** in your home directory.

## Keyboard Shortcuts
| Key | Description |
| :--- | :--- |
| **?** | Show this help |
| **Esc** | Cancel and return |
| **Tab/Up/Down** | Navigate between fields |
| **Enter** | Save settings (on last field) |

## How to Use

### 1. AI Backend
- Select your AI provider (Ollama, Gemini, OpenAI, Claude, etc.)
- **Ollama** is free and local (no API key needed)
- Others require API keys for cloud access

### 2. AI Model
- Specify the model name for your backend
- **Examples**:
  - *Ollama*: llama3, codellama, mistral
  - *OpenAI*: gpt-4, gpt-3.5-turbo
  - *Gemini*: gemini-1.5-flash, gemini-pro
  - *Claude*: claude-3-opus, claude-3-sonnet

### 3. API Key
- Required for cloud providers
- Get keys from:
// [OpenAI](https://platform.openai.com), 
// [Gemini](https://makersuite.google.com), 
// [Claude](https://console.anthropic.com)
- Paste key in the field (masked for security)

### 4. Base URL (Optional)
- For custom API endpoints (e.g., LM Studio: http://localhost:1234/v1)
- Leave empty for default provider endpoints

## Configuration File
Settings are stored at:
- **Windows**: C:\Users\<user>\.devcli\config.yaml
- **Linux/Mac**: ~/.devcli/config.yaml

---
*Press **Esc** to close this guide*`

	BoilerplateHelp = `
        BOILERPLATE GENERATOR - Help & Usage Guide           


OVERVIEW
Generate code snippets, templates, and complete project architectures for multiple programming languages and frameworks.

KEYBOARD SHORTCUTS
Key         Description
         
?           Show this help
Esc         Go back to previous menu
Up/Down     Navigate lists
Enter       Select / Generate
Tab         Switch between categories

HOW TO USE

1. CODE SNIPPETS
   • Pre-built code patterns for common tasks.
   • Available snippets: CRUD API endpoints, Authentication systems, Database connections, Error handling patterns, Utility functions.
   • Copy directly to your project.

2. PROJECT TEMPLATES
   • Complete application structures including MVC architecture, REST API structure, Microservices setup, and Full-stack templates.
   • All files and folders created automatically.

3. LANGUAGE SUPPORT
   • Go (Web servers, APIs, CLI tools)
   • Python (Flask, Django, FastAPI)
   • JavaScript (Express, React, Vue)
   • Java (Spring Boot)
   • Rust (Actix, Rocket)
   • C/C++ (Project templates)
   • Zig (Basic structures)

4. SAVE CUSTOM TEMPLATES
   • Save your own project as a template to reuse across multiple projects. Keep consistent architecture and share with team members.

FEATURES
• Instant code generation
• Runnable examples
• Best practices built-in
• Documentation included
• Dependencies configured

TIPS
• Preview generated code before saving. Snippets are ready to run. Templates include all config files. Can combine multiple snippets. Custom templates save time.

Press Esc to close this help`

	BonusFeaturesHelp = `
          BONUS FEATURES - Help & Usage Guide                 


OVERVIEW
Additional productivity tools for project analysis, task automation, file generation, code snippet management, and AI assistance.

AVAILABLE FEATURES

1. PROJECT DASHBOARD
   • Overview of all projects in workspace.
   • Features: Technology stack detection, Project status (Active/Broken/Archived), Last modified dates, Size calculations.
   • Sortable by name, date, or status.

2. TASK RUNNER
   • Auto-detect build/test/lint tasks. One-click task execution.
   • Support for npm/yarn scripts, Go commands (build, test, fmt), Python tools (pytest, black), Rust cargo commands, and Makefile targets.
   • Live output streaming and ability to cancel running tasks.

3. SMART FILE CREATOR
   • Generate common configuration files: .env, .gitignore, Dockerfile, docker-compose.yml, .editorconfig, GitHub Actions CI/CD, Makefile.
   • Customizable for your language with preview before saving.
   • Production-ready templates.

4. SNIPPET LIBRARY
   • Personal vault of reusable code.
   • Save frequently used code, organize by category and language.
   • Search and filter snippets, quick copy to clipboard.
   • Includes default snippets.

5. AI ASSISTANT
   • Your intelligent coding companion.
   • Features: Code generation, Algorithm explanations, Bug fix suggestions, Documentation creation, Best practices advice.
   • Multi-turn conversations.

KEYBOARD SHORTCUTS
?           Show this help
Esc         Return to main menu
Up/Down     Navigate
Enter       Select feature

HOW TO ACCESS
1. Open Project Tools from main menu.
2. Select "Bonus Features".
3. Choose desired tool.

TIPS
• Use Project Dashboard for quick overview.
• Task Runner saves typing commands.
• Smart File Creator ensures consistency.
• Snippet Library speeds up development.
• AI Assistant helps when stuck.

Press Esc to close this help`

	TaskRunnerHelp = `
           TASK RUNNER - Help & Usage Guide                 


OVERVIEW
The Task Runner automatically detects build, test, lint, and run scripts
in your project and allows you to execute them with a single click.

KEYBOARD SHORTCUTS
Key         Description
         
?           Show this help
Esc/q       Go back / Quick exit
R           Refresh task list
Enter       Run selected task
Ctrl+C      Cancel currently running task
Up/Down     Navigate tasks / Scroll output
Mouse Wheel Scroll through tasks and output

HOW TO USE

1. TASK DETECTION
   • Task Runner scans your project root for:
     - package.json (npm scripts)
     - go.mod (go build/test/fmt)
     - requirements.txt / .py files (pytest, black, flake8)
     - pom.xml / build.gradle (Maven/Gradle)
     - Cargo.toml (Rust build/test)
     - CMakeLists.txt / .c / .cpp (C/C++ compilation)
     - Makefile (Make targets)

2. RUNNING A TASK
   • Select a task from the list and press Enter.
   • The interface switches to "Running" state.
   • Real-time output is streamed to the viewport.

3. MANAGING OUTPUT
   • Use Arrow Keys or Mouse Wheel to scroll through logs.
   • Tasks automatically scroll to the bottom as they run.
   • Once complete, review the logs and press Esc to return.

4. CANCELLING TASKS
   • If a task is taking too long or hung, press Ctrl+C.
   • Task Runner will attempt to terminate the process.

SUPPORTED LANGUAGES
• Go, Python, Node.js, Java, Rust, C/C++, Makefile targets.

Press Esc to close this help`

	AutoUpdateHelp = `
         AUTO-UPDATE CENTER - Help & Usage Guide             


OVERVIEW
Check and update programming language versions, DevCLI itself,
and manage API keys for AI providers.

KEYBOARD SHORTCUTS
Key         Description
         
?           Show this help
Esc         Return to main menu
Up/Down     Navigate options
Enter       Select option
y/n         Confirm/Cancel updates

FEATURES

1. CHECK LANGUAGE VERSIONS
   • Displays installed versions of:
     - Go
     - Python
     - Node.js
     - Java
     - Rust
     - Zig
     - C/C++ (gcc/g++)
   • Shows installation paths
   • Helps identify outdated tools

2. UPDATE AI KEYS
   • Manage API keys for:
     - Google Gemini
     - OpenAI
     - Anthropic Claude
     - HuggingFace
     - Ollama (base URL)
   • Secure key entry
   • Keys saved to config
   • Immediate effect

3. CHECK DEVCLI UPDATES
   • Checks for new versions via Git
   • AI-generated release notes
   • Shows new features and fixes
   • One-click update process
   • Automatic rebuild

HOW TO USE

LANGUAGE VERSION CHECK:
1. Select "Check Language Versions"
2. View installed versions and paths
3. Manually update tools if needed
4. Press Esc to return

UPDATE AI KEYS:
1. Select "Update AI Keys"
2. Choose provider
3. Enter API key
4. Key is saved automatically
5. Restart chat to use new key

UPDATE DEVCLI:
1. Select "Check DevCLI Updates"
2. Wait for Git fetch
3. Review AI-generated summary
4. Press 'y' to install
5. DevCLI rebuilds automatically
6. Restart to use new version

REQUIREMENTS
• Git (for DevCLI updates)
• Internet connection
• Configured Git remote

TIPS
• Check language versions after fresh install
• Update API keys when they expire
• DevCLI updates preserve your config
• Backup important work before updating
• Keep tools up-to-date for security

Press Esc to close this help`

	SmartFileHelp = `
          SMART FILE CREATOR - Help & Usage Guide             
                                                                
OVERVIEW
Generate common configuration files and project scaffolding or creating custom 
files using AI.
                                                                
KEYBOARD SHORTCUTS
Key         Description
                                                                
?           Show this help
Esc         Go back / Return to previous menu
Up/Down     Navigate templates
Enter       Select template / Confirm action
                                                                
HOW TO USE
                                                                
1. SELECT TEMPLATE
   • Browse through the list of available templates
   • templates include .gitignore, Dockerfile, Makefile, etc.
   • select "Custom File (AI)" for AI generation
                                                                
2. CONFIGURE
   • For static templates (e.g., .gitignore), choose the destination folder.
   • For dynamic templates, enter the filename or specific options.
   • For AI files, describe what you want the file to contain.
                                                                
3. PREVIEW & SAVE
   • Review the generated content in the preview window.
   • Press Enter to save the file to your workspace.
   • Press Esc to go back and edit parameters.

Press Esc to close this help`

	SnippetLibraryHelp = `
           SNIPPET LIBRARY - Help & Usage Guide               
                                                                
OVERVIEW
Manage your personal collection of code snippets. Save, search, 
and retrieve useful code blocks across languages.
                                                                
KEYBOARD SHORTCUTS
Key         Description
                                                                
?           Show this help
Esc         Go back to main menu
Up/Down     Navigate snippet list
Enter       View details of selected snippet
A           Add a new snippet
/           Search snippets
R           Refresh list
D           Delete snippet (in View mode)
                                                                
HOW TO USE
                                                                
1. ADD SNIPPET
   • Press 'A' to open the creation form.
   • Enter Title, Description, and Language.
   • Paste or type your code.
                                                                
2. VIEW & COPY
   • Select a snippet and press Enter.
   • View the code with syntax highlighting.
   • Use Mouse Wheel or Up/Down to scroll.
                                                                
3. SEARCH
   • Press '/' to focus the search bar.
   • Type keywords to filter snippets by title or content.
   • Press Esc to clear search.

Press Esc to close this help`

	AIAssistantHelp = `
             AI ASSISTANT - Help & Usage Guide                
                                                                
OVERVIEW
Your intelligent coding companion. Generate code, ask questions, 
debug issues, and get architectural advice directly in the terminal.
                                                                
KEYBOARD SHORTCUTS
Key         Description
                                                                
?           Show this help
Esc         Go back / Clear input
Ctrl+D      Send prompt to AI
Tab         Switch Agent (CodeGen -> Architect -> Debugger)
Shift+Tab   Switch Agent backwards
Up/Down     Scroll output
                                                                
AGENTS
                                                                
1. CODE GEN AGENT
   • Optimized for writing code snippets and functions.
   • usage: "Write a Go function to reverse a string"
                                                                
2. ARCHITECT AGENT
   • Focuses on high-level design and structure.
   • usage: "How should I structure a microservices app?"
                                                                
3. DEBUGGER AGENT
   • Specializes in finding and fixing errors.
   • usage: Paste an error message or buggy code.
                                                                
HOW TO USE
1. Type your prompt in the input box.
2. Press Ctrl+D to send.
3. Wait for the AI streaming response.
4. Scroll through the output to read the response.
                                                        
Press Esc to close this help`
)

func RenderHelp(content string, width, height int) string {
	return lipgloss.NewStyle().
		Padding(1, 2).
		Width(width - 4).
		Render(content)
}
