package tui

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

type DocsModel struct {
	viewport viewport.Model
	width    int
	height   int
	ready    bool
	renderer *glamour.TermRenderer
	language string // Current language code
}

const (
	LangEnglish  = "en"
	LangSpanish  = "es"
	LangFrench   = "fr"
	LangGerman   = "de"
	LangChinese  = "zh"
	LangJapanese = "jp"

	Backtick = "`"
)

func NewDocsModel() DocsModel {
	// Initialize glamour renderer
	r, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(100), // Default, will be updated on resize
	)

	return DocsModel{
		width:    100, // Default, will be resized
		height:   30,
		renderer: r,
		language: LangEnglish,
	}
}

func (m DocsModel) Init() tea.Cmd {
	return nil
}

func (m DocsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, func() tea.Msg { return BackMsg{} }
		case "ctrl+c":
			return m, tea.Quit
		case "l", "L":
			// Cycle languages
			switch m.language {
			case LangEnglish:
				m.language = LangSpanish
			case LangSpanish:
				m.language = LangFrench
			case LangFrench:
				m.language = LangGerman
			case LangGerman:
				m.language = LangChinese
			case LangChinese:
				m.language = LangJapanese
			case LangJapanese:
				m.language = LangEnglish
			default:
				m.language = LangEnglish
			}
			m.resizeViewport() // Re-render with new language
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizeViewport()
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m DocsModel) View() string {
	if !m.ready {
		return "Loading Docs..."
	}

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#0F9E99")).
		Padding(1, 0).
		Align(lipgloss.Center).
		Width(m.width).
		Render("DevCLI Documentation")

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Padding(1, 0).
		Align(lipgloss.Center).
		Width(m.width).
		Render("Esc/Q: Back • ↑/↓: Scroll • L: Switch Language (" + m.getLangName() + ")")

	return lipgloss.JoinVertical(
		lipgloss.Center,
		header,
		m.viewport.View(),
		footer,
	)
}

func (m *DocsModel) resizeViewport() {
	headerHeight := 4 // Header + Padding
	footerHeight := 3 // Footer + Padding
	verticalMarginHeight := headerHeight + footerHeight

	// Re-create renderer with new width
	contentWidth := m.width - 4 // Padding
	if contentWidth < 20 {
		contentWidth = 20
	}

	m.renderer, _ = glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(contentWidth),
	)

	renderedContent, err := m.renderer.Render(getDocsContent(m.language))
	if err != nil {
		renderedContent = "Error rendering documentation: " + err.Error()
	}

	if !m.ready {
		// First time initialization
		m.viewport = viewport.New(m.width, m.height-verticalMarginHeight)
		m.viewport.SetContent(renderedContent)
		m.ready = true
	} else {
		m.viewport.Width = m.width
		m.viewport.Height = m.height - verticalMarginHeight
		m.viewport.SetContent(renderedContent) // Re-render content for wrapping
	}
}

func (m DocsModel) getLangName() string {
	switch m.language {
	case LangEnglish:
		return "English"
	case LangSpanish:
		return "Español"
	case LangFrench:
		return "Français"
	case LangGerman:
		return "Deutsch"
	case LangChinese:
		return "中文"
	case LangJapanese:
		return "日本語"
	}
	return "English"
}

func getDocsContent(lang string) string {
	switch lang {
	case LangSpanish:
		return getDocsContentSpanish()
	case LangFrench:
		return getDocsContentFrench()
	case LangGerman:
		return getDocsContentGerman()
	case LangChinese:
		return getDocsContentChinese()
	case LangJapanese:
		return getDocsContentJapanese()
	default:
		return getDocsContentEnglish()
	}
}

func getDocsContentEnglish() string {
	return `
# DevCLI - The Ultimate Developer Survival Kit

DevCLI is not just a tool; it's a complete rethink of the developer experience. It unifies project management, coding, debugging, and AI assistance into a single, cohesive terminal interface. Say goodbye to fragmented workflows and context switching.

> **Philosophy**: "Stay in the flow." Everything you need is one keystroke away.

---

## 🚀 Complete Feature Breakdown

### 1. Project Management & Scaffolding
*   **Unified Dashboard (` + Backtick + `dev` + Backtick + `)**: 
    *   Auto-detects project types (Go, Python, React, Rust, Node, etc.).
    *   Displays git status, active ports, tech stack, and last modified times.
    *   Quick actions: Open, Run, Edit, Delete.
*   **Instant Scaffolding (` + Backtick + `dev new` + Backtick + `)**: 
    *   **Templates**: Robust, production-ready templates for:
        *   **Go**: CHI, Gin, Fiber, Cobra CLI.
        *   **Python**: Flask, FastAPI, Django.
        *   **Frontend**: React (Vite), Vue, Svelte.
        *   **Rust**: Actix-web, Axum.
    *   **Smart Init**: Generates ` + Backtick + `.gitignore` + Backtick + `, ` + Backtick + `Dockerfile` + Backtick + `, ` + Backtick + `Makefile` + Backtick + `, and CI/CD pipelines automatically.

### 2. Development Power Tools
*   **Task Runner (` + Backtick + `dev run` + Backtick + `)**: 
    *   Intelligently parses ` + Backtick + `package.json` + Backtick + `, ` + Backtick + `Makefile` + Backtick + `, ` + Backtick + `Justfile` + Backtick + `, ` + Backtick + `go.mod` + Backtick + `.
    *   Runs scripts (build, test, lint, deploy) without needing to remember specific tool syntax.
*   **Dev Server Launcher (` + Backtick + `dev server` + Backtick + `)**: 
    *   Auto-detects framework (Next.js, Flask, Laravel, etc.) and start command.
    *   Manages ports automatically to avoid conflicts.
    *   Streams live logs directly to the dashboard.
*   **Virtual Environment Manager (` + Backtick + `dev venv` + Backtick + `)**: 
    *   Centralized view of all Python ` + Backtick + `venvs` + Backtick + ` and Node ` + Backtick + `node_modules` + Backtick + ` across your system.
    *   One-click activate/deactivate.
    *   Deep Clean: Find and delete abandoned environments to reclaim GBs of disk space.

### 3. Integrated Editing & File Management
*   **Nano-Style Editor**: 
    *   Fast, lightweight editor embedded in DevCLI.
    *   Syntax highlighting for 50+ languages.
    *   Go-to-line, search/replace, and undo/redo support.
*   **Smart File Manager (` + Backtick + `dev file` + Backtick + `)**: 
    *   Fuzzy search for files and directories.
    *   CRUD operations (Create, Read, Update, Delete).
    *   Bulk actions and file preview.

### 4. AI & Intelligence (The Brain)
*   **AI Chat Assistant (` + Backtick + `dev chat` + Backtick + `)**: 
    *   **Multi-Model Support**: Switch instantly between local (Ollama/Llama3) and cloud (OpenAI GPT-4, Claude 3.5, Gemini 1.5).
    *   **Context Awareness**: 
        *   Reference files with ` + Backtick + `@filename` + Backtick + `.
        *   Reference directories with ` + Backtick + `@dirname` + Backtick + `.
        *   Paste clipboard content/images seamlessly.
    *   **Modes**: Coding, Debugging, Creative Writing, General Chat.
*   **Code Time Machine (` + Backtick + `dev time` + Backtick + `)**: 
    *   Visual Git history explorer.
    *   Step through commits to see code evolution.
    *   **AI Analysis**: "Explain this commit" and "Find potential bugs in this diff".
    *   Blame view to see who wrote what and when.

---

## ⚙️ Advanced Configuration

Edit ` + Backtick + `~/.devcli/config.yaml` + Backtick + ` to customize your experience.

### AI Configuration
` + Backtick + "```yaml" + Backtick + `
ai:
  default_provider: "ollama"
  providers:
    ollama:
      model: "llama3"
      base_url: "http://localhost:11434"
    openai:
      api_key: "sk-..."
      model: "gpt-4-turbo"
    anthropic:
      api_key: "sk-ant-..."
      model: "claude-3-opus"
` + Backtick + "```" + Backtick + `

### UI Customization
` + Backtick + "```yaml" + Backtick + `
ui:
  theme: "dracula" # options: dracula, nord, monokai, system
  editor:
    line_numbers: true
    mouse_support: true
` + Backtick + "```" + Backtick + `

---

## ⌨️ Master Keybindings

| Context | Key | Action |
| :--- | :--- | :--- |
| **Global** | ` + Backtick + `Ctrl+C` + Backtick + ` | Quit Application |
| | ` + Backtick + `Esc` + Backtick + ` | Go Back / Close View |
| | ` + Backtick + `?` + Backtick + ` | Toggle Help Overlay |
| **Navigation** | ` + Backtick + `↑/↓` + Backtick + ` | Navigate Lists |
| | ` + Backtick + `Enter` + Backtick + ` | Select / Execute |
| | ` + Backtick + `Tab` + Backtick + ` | Switch Focus |
| **Editor** | ` + Backtick + `Ctrl+S` + Backtick + ` | Save File |
| | ` + Backtick + `Ctrl+F` + Backtick + ` | Find Text |
| **Docs** | ` + Backtick + `L` + Backtick + ` | Cycle Language |

---

## ❓ Troubleshooting

**Q: "Command not found: ollama"**
A: Ensure you have installed Ollama from [ollama.com](https://ollama.com) and the service is running (` + Backtick + `ollama serve` + Backtick + `).

**Q: DevCLI is slow to start.**
A: If you have thousands of projects, try configuring ` + Backtick + `scan_depth: 2` + Backtick + ` in your config.yaml to limit directory traversal depth.

**Q: Git integration is failing.**
A: DevCLI requires ` + Backtick + `git` + Backtick + ` to be installed and available in your system PATH.

---

## 🤝 Community & Support

*   **GitHub**: [github.com/phravins/devcli](https://github.com/phravins/devcli)
*   **Issues**: Submit bug reports and feature requests on GitHub.
*   **Contribution**: PRs are welcome! See ` + Backtick + `CONTRIBUTING.md` + Backtick + `.

*Empowering developers to build faster, cleaner, and smarter.*
`
}

func getDocsContentSpanish() string {
	return `
# DevCLI - El Kit de Supervivencia Definitivo para Desarrolladores

DevCLI no es solo una herramienta; es un replanteamiento completo de la experiencia del desarrollador. Unifica la gestión de proyectos, la codificación, la depuración y la asistencia de IA en una única interfaz de terminal cohesiva. Dile adiós a los flujos de trabajo fragmentados y los cambios de contexto.

> **Filosofía**: "Mantente en el flujo." Todo lo que necesitas está a una pulsación de tecla.

---

## 🚀 Desglose Completo de Características

### 1. Gestión de Proyectos y Andamiaje
*   **Panel Unificado (` + Backtick + `dev` + Backtick + `)**: 
    *   Detecta automáticamente tipos de proyectos (Go, Python, React, Rust, Node, etc.).
    *   Muestra el estado de git, puertos activos, pila tecnológica y tiempos de última modificación.
    *   Acciones rápidas: Abrir, Ejecutar, Editar, Eliminar.
*   **Andamiaje Instantáneo (` + Backtick + `dev new` + Backtick + `)**: 
    *   **Plantillas**: Plantillas robustas y listas para producción para:
        *   **Go**: CHI, Gin, Fiber, Cobra CLI.
        *   **Python**: Flask, FastAPI, Django.
        *   **Frontend**: React (Vite), Vue, Svelte.
        *   **Rust**: Actix-web, Axum.
    *   **Inicio Inteligente**: Genera ` + Backtick + `.gitignore` + Backtick + `, ` + Backtick + `Dockerfile` + Backtick + `, ` + Backtick + `Makefile` + Backtick + ` y pipelines de CI/CD automáticamente.

### 2. Herramientas de Desarrollo Potentes
*   **Ejecutor de Tareas (` + Backtick + `dev run` + Backtick + `)**: 
    *   Analiza inteligentemente ` + Backtick + `package.json` + Backtick + `, ` + Backtick + `Makefile` + Backtick + `, ` + Backtick + `Justfile` + Backtick + `, ` + Backtick + `go.mod` + Backtick + `.
    *   Ejecuta scripts (build, test, lint, deploy) sin necesidad de recordar la sintaxis específica de la herramienta.
*   **Lanzador de Servidor de Desarrollo (` + Backtick + `dev server` + Backtick + `)**: 
    *   Detecta automáticamente el framework (Next.js, Flask, Laravel, etc.) y el comando de inicio.
    *   Gestiona puertos automáticamente para evitar conflictos.
    *   Transmite registros en vivo directamente al panel.
*   **Gestor de Entornos Virtuales (` + Backtick + `dev venv` + Backtick + `)**: 
    *   Vista centralizada de todos los ` + Backtick + `venvs` + Backtick + ` de Python y ` + Backtick + `node_modules` + Backtick + ` de Node en tu sistema.
    *   Activar/desactivar con un clic.
    *   Limpieza Profunda: Encuentra y elimina entornos abandonados para recuperar GBs de espacio en disco.

### 3. Edición Integrada y Gestión de Archivos
*   **Editor Estilo Nano**: 
    *   Editor rápido y ligero integrado en DevCLI.
    *   Resaltado de sintaxis para más de 50 lenguajes.
    *   Ir a línea, buscar/reemplazar y deshacer/rehacer.
*   **Gestor de Archivos Inteligente (` + Backtick + `dev file` + Backtick + `)**: 
    *   Búsqueda difusa para archivos y directorios.
    *   Operaciones CRUD (Crear, Leer, Actualizar, Eliminar).
    *   Acciones masivas y vista previa de archivos.

### 4. IA e Inteligencia (El Cerebro)
*   **Asistente de Chat IA (` + Backtick + `dev chat` + Backtick + `)**: 
    *   **Soporte Multi-Modelo**: Cambia instantáneamente entre local (Ollama/Llama3) y nube (OpenAI GPT-4, Claude 3.5, Gemini 1.5).
    *   **Conciencia de Contexto**: 
        *   Referencia archivos con ` + Backtick + `@filename` + Backtick + `.
        *   Referencia directorios con ` + Backtick + `@dirname` + Backtick + `.
        *   Pega contenido/imágenes del portapapeles sin problemas.
    *   **Modos**: Codificación, Depuración, Escritura Creativa, Chat General.
*   **Máquina del Tiempo de Código (` + Backtick + `dev time` + Backtick + `)**: 
    *   Explorador visual del historial de Git.
    *   Recorre commits para ver la evolución del código.
    *   **Análisis de IA**: "Explica este commit" y "Encuentra posibles errores en esta diferencia".
    *   Vista de culpa para ver quién escribió qué y cuándo.

---

## ⚙️ Configuración Avanzada

Edita ` + Backtick + `~/.devcli/config.yaml` + Backtick + ` para personalizar tu experiencia.

### Configuración de IA
` + Backtick + "```yaml" + Backtick + `
ai:
  default_provider: "ollama"
  providers:
    ollama:
      model: "llama3"
      base_url: "http://localhost:11434"
    openai:
      api_key: "sk-..."
      model: "gpt-4-turbo"
    anthropic:
      api_key: "sk-ant-..."
      model: "claude-3-opus"
` + Backtick + "```" + Backtick + `

### Personalización de UI
` + Backtick + "```yaml" + Backtick + `
ui:
  theme: "dracula" # opciones: dracula, nord, monokai, system
  editor:
    line_numbers: true
    mouse_support: true
` + Backtick + "```" + Backtick + `

---

## ⌨️ Atajos Maestros

| Contexto | Tecla | Acción |
| :--- | :--- | :--- |
| **Global** | ` + Backtick + `Ctrl+C` + Backtick + ` | Salir de la Aplicación |
| | ` + Backtick + `Esc` + Backtick + ` | Volver / Cerrar Vista |
| | ` + Backtick + `?` + Backtick + ` | Alternar Superposición de Ayuda |
| **Navigation** | ` + Backtick + `↑/↓` + Backtick + ` | Navegar Listas |
| | ` + Backtick + `Enter` + Backtick + ` | Seleccionar / Ejecutar |
| | ` + Backtick + `Tab` + Backtick + ` | Cambiar Foco |
| **Editor** | ` + Backtick + `Ctrl+S` + Backtick + ` | Guardar Archivo |
| | ` + Backtick + `Ctrl+F` + Backtick + ` | Buscar Texto |
| **Docs** | ` + Backtick + `L` + Backtick + ` | Cambiar Idioma |

---

## ❓ Solución de Problemas

**P: "Comando no encontrado: ollama"**
R: Asegúrate de haber instalado Ollama desde [ollama.com](https://ollama.com) y que el servicio esté en ejecución (` + Backtick + `ollama serve` + Backtick + `).

**P: DevCLI tarda en iniciar.**
R: Si tienes miles de proyectos, intenta configurar ` + Backtick + `scan_depth: 2` + Backtick + ` en tu config.yaml para limitar la profundidad de recorrido del directorio.

**P: La integración de Git está fallando.**
R: DevCLI requiere que ` + Backtick + `git` + Backtick + ` esté instalado y disponible en el PATH de tu sistema.

---

## 🤝 Comunidad y Soporte

*   **GitHub**: [github.com/phravins/devcli](https://github.com/phravins/devcli)
*   **Problemas**: Envía informes de errores y solicitudes de funciones en GitHub.
*   **Contribución**: ¡Los PR son bienvenidos! Consulta ` + Backtick + `CONTRIBUTING.md` + Backtick + `.

*Empoderando a los desarrolladores para construir más rápido, más limpio y más inteligente.*
`
}

func getDocsContentFrench() string {
	return `
# DevCLI - Le Kit de Survie Ultime pour Développeurs

DevCLI n'est pas seulement un outil ; c'est une refonte complète de l'expérience développeur. Il unifie la gestion de projets, le codage, le débogage et l'assistance IA dans une interface de terminal unique et cohérente. Dites adieu aux flux de travail fragmentés et aux changements de contexte.

> **Philosophie**: "Restez dans le flux." Tout ce dont vous avez besoin est à une touche de distance.

---

## 🚀 Répartition Complète des Fonctionnalités

### 1. Gestion de Projets & Échafaudage
*   **Tableau de Bord Unifié (` + Backtick + `dev` + Backtick + `)**: 
    *   Détecte automatiquement les types de projets (Go, Python, React, Rust, Node, etc.).
    *   Affiche l'état git, les ports actifs, la pile technologique et les heures de dernière modification.
    *   Actions rapides : Ouvrir, Exécuter, Éditer, Supprimer.
*   **Échafaudage Instantané (` + Backtick + `dev new` + Backtick + `)**: 
    *   **Modèles**: Des modèles robustes et prêts pour la production pour :
        *   **Go**: CHI, Gin, Fiber, Cobra CLI.
        *   **Python**: Flask, FastAPI, Django.
        *   **Frontend**: React (Vite), Vue, Svelte.
        *   **Rust**: Actix-web, Axum.
    *   **Initialisation Intelligente**: Génère ` + Backtick + `.gitignore` + Backtick + `, ` + Backtick + `Dockerfile` + Backtick + `, ` + Backtick + `Makefile` + Backtick + ` et pipelines CI/CD automatiquement.

### 2. Outils de Développement Puissants
*   **Exécuteur de Tâches (` + Backtick + `dev run` + Backtick + `)**: 
    *   Analyse intelligemment ` + Backtick + `package.json` + Backtick + `, ` + Backtick + `Makefile` + Backtick + `, ` + Backtick + `Justfile` + Backtick + `, ` + Backtick + `go.mod` + Backtick + `.
    *   Exécute des scripts (build, test, lint, deploy) sans avoir besoin de se souvenir de la syntaxe spécifique de l'outil.
*   **Lanceur de Serveur de Dév (` + Backtick + `dev server` + Backtick + `)**: 
    *   Détecte automatiquement le framework (Next.js, Flask, Laravel, etc.) et la commande de démarrage.
    *   Gère les ports automatiquement pour éviter les conflits.
    *   Diffuse les logs en direct directement sur le tableau de bord.
*   **Gestionnaire d'Environnements Virtuels (` + Backtick + `dev venv` + Backtick + `)**: 
    *   Vue centralisée de tous les ` + Backtick + `venvs` + Backtick + ` Python et ` + Backtick + `node_modules` + Backtick + ` Node sur votre système.
    *   Activer/Désactiver en un clic.
    *   Nettoyage en Profondeur : Trouve et supprime les environnements abandonnés pour récupérer des Go d'espace disque.

### 3. Édition Intégrée & Gestion de Fichiers
*   **Éditeur Style Nano**: 
    *   Éditeur rapide et léger intégré dans DevCLI.
    *   Coloration syntaxique pour plus de 50 langages.
    *   Aller à la ligne, rechercher/remplacer et annuler/rétablir.
*   **Gestionnaire de Fichiers Intelligent (` + Backtick + `dev file` + Backtick + `)**: 
    *   Recherche floue pour fichiers et répertoires.
    *   Opérations CRUD (Créer, Lire, Mettre à jour, Supprimer).
    *   Actions en masse et aperçu de fichiers.

### 4. IA & Intelligence (Le Cerveau)
*   **Assistant Chat IA (` + Backtick + `dev chat` + Backtick + `)**: 
    *   **Support Multi-Modèles**: Basculez instantanément entre local (Ollama/Llama3) et cloud (OpenAI GPT-4, Claude 3.5, Gemini 1.5).
    *   **Conscience du Contexte**: 
        *   Référencez des fichiers avec ` + Backtick + `@filename` + Backtick + `.
        *   Référencez des répertoires avec ` + Backtick + `@dirname` + Backtick + `.
        *   Collez du contenu/images du presse-papiers sans problème.
    *   **Modes**: Codage, Débogage, Écriture Créative, Chat Général.
*   **Machine à Remonter le Temps du Code (` + Backtick + `dev time` + Backtick + `)**: 
    *   Explorateur visuel de l'historique Git.
    *   Parcourez les commits pour voir l'évolution du code.
    *   **Analyse IA**: "Expliquez ce commit" et "Trouvez des bugs potentiels dans ce diff".
    *   Vue 'Blame' pour voir qui a écrit quoi et quand.

---

## ⚙️ Configuration Avancée

Éditez ` + Backtick + `~/.devcli/config.yaml` + Backtick + ` pour personnaliser votre expérience.

### Configuration IA
` + Backtick + "```yaml" + Backtick + `
ai:
  default_provider: "ollama"
  providers:
    ollama:
      model: "llama3"
      base_url: "http://localhost:11434"
    openai:
      api_key: "sk-..."
      model: "gpt-4-turbo"
    anthropic:
      api_key: "sk-ant-..."
      model: "claude-3-opus"
` + Backtick + "```" + Backtick + `

### Personnalisation UI
` + Backtick + "```yaml" + Backtick + `
ui:
  theme: "dracula" # options: dracula, nord, monokai, system
  editor:
    line_numbers: true
    mouse_support: true
` + Backtick + "```" + Backtick + `

---

## ⌨️ Raccourcis Maîtres

| Contexte | Touche | Action |
| :--- | :--- | :--- |
| **Global** | ` + Backtick + `Ctrl+C` + Backtick + ` | Quitter l'Application |
| | ` + Backtick + `Esc` + Backtick + ` | Retour / Fermer la Vue |
| | ` + Backtick + `?` + Backtick + ` | Basculer l'Aide |
| **Navigation** | ` + Backtick + `↑/↓` + Backtick + ` | Naviguer dans les Listes |
| | ` + Backtick + `Entrée` + Backtick + ` | Sélectionner / Exécuter |
| | ` + Backtick + `Tab` + Backtick + ` | Changer le Focus |
| **Éditeur** | ` + Backtick + `Ctrl+S` + Backtick + ` | Enregistrer le Fichier |
| | ` + Backtick + `Ctrl+F` + Backtick + ` | Rechercher du Texte |
| **Docs** | ` + Backtick + `L` + Backtick + ` | Changer de Langue |

---

## ❓ Dépannage

**Q: "Commande non trouvée : ollama"**
R: Assurez-vous d'avoir installé Ollama depuis [ollama.com](https://ollama.com) et que le service est en cours d'exécution (` + Backtick + `ollama serve` + Backtick + `).

**Q: DevCLI est lent au démarrage.**
R: Si vous avez des milliers de projets, essayez de configurer ` + Backtick + `scan_depth: 2` + Backtick + ` dans votre config.yaml pour limiter la profondeur de parcours des répertoires.

**Q: L'intégration Git échoue.**
R: DevCLI nécessite que ` + Backtick + `git` + Backtick + ` soit installé et disponible dans le PATH de votre système.

---

## 🤝 Communauté & Support

*   **GitHub**: [github.com/phravins/devcli](https://github.com/phravins/devcli)
*   **Problèmes**: Soumettez des rapports de bugs et des demandes de fonctionnalités sur GitHub.
*   **Contribution**: Les PR sont les bienvenus ! Voir ` + Backtick + `CONTRIBUTING.md` + Backtick + `.

*Permettre aux développeurs de construire plus vite, plus proprement et plus intelligemment.*
`
}

func getDocsContentGerman() string {
	return `
# DevCLI - Das ultimative Entwickler-Survival-Kit

DevCLI ist nicht nur ein Werkzeug; es ist ein komplettes Überdenken der Entwicklererfahrung. Es vereint Projektmanagement, Codierung, Debugging und KI-Unterstützung in einer einzigen, kohärenten Terminaloberfläche. Verabschieden Sie sich von fragmentierten Workflows und Kontextwechseln.

> **Philosophie**: "Im Fluss bleiben." Alles, was Sie brauchen, ist nur einen Tastendruck entfernt.

---

## 🚀 Vollständige Funktionsübersicht

### 1. Projektmanagement & Gerüstbau
*   **Einheitliches Dashboard (` + Backtick + `dev` + Backtick + `)**: 
    *   Erkennt automatisch Projekttypen (Go, Python, React, Rust, Node, usw.).
    *   Zeigt Git-Status, aktive Ports, Tech-Stack und letzte Änderungszeiten an.
    *   Schnellaktionen: Öffnen, Ausführen, Bearbeiten, Löschen.
*   **Sofortiges Gerüstbau (` + Backtick + `dev new` + Backtick + `)**: 
    *   **Vorlagen**: Robuste, produktionsreife Vorlagen für:
        *   **Go**: CHI, Gin, Fiber, Cobra CLI.
        *   **Python**: Flask, FastAPI, Django.
        *   **Frontend**: React (Vite), Vue, Svelte.
        *   **Rust**: Actix-web, Axum.
    *   **Smart Init**: Generiert automatisch ` + Backtick + `.gitignore` + Backtick + `, ` + Backtick + `Dockerfile` + Backtick + `, ` + Backtick + `Makefile` + Backtick + ` und CI/CD-Pipelines.

### 2. Entwickler-Power-Tools
*   **Task Runner (` + Backtick + `dev run` + Backtick + `)**: 
    *   Analysiert intelligent ` + Backtick + `package.json` + Backtick + `, ` + Backtick + `Makefile` + Backtick + `, ` + Backtick + `Justfile` + Backtick + `, ` + Backtick + `go.mod` + Backtick + `.
    *   Führt Skripte (Build, Test, Lint, Deploy) aus, ohne sich an spezifische Tool-Syntax erinnern zu müssen.
*   **Dev-Server-Launcher (` + Backtick + `dev server` + Backtick + `)**: 
    *   Erkennt automatisch Framework (Next.js, Flask, Laravel, usw.) und Startbefehl.
    *   Verwaltet Ports automatisch, um Konflikte zu vermeiden.
    *   Streamt Live-Logs direkt auf das Dashboard.
*   **Manager für virtuelle Umgebungen (` + Backtick + `dev venv` + Backtick + `)**: 
    *   Zentrale Ansicht aller Python ` + Backtick + `venvs` + Backtick + ` und Node ` + Backtick + `node_modules` + Backtick + ` auf Ihrem System.
    *   Aktivieren/Deaktivieren mit einem Klick.
    *   Tiefenreinigung: Finden und löschen Sie verwaiste Umgebungen, um GBs an Speicherplatz zurückzugewinnen.

### 3. Integrierte Bearbeitung & Dateimanagement
*   **Editor im Nano-Stil**: 
    *   Schneller, leichter Editor, eingebettet in DevCLI.
    *   Syntaxhervorhebung für 50+ Sprachen.
    *   Gehe zu Zeile, Suchen/Ersetzen und Rückgängig/Wiederherstellen.
*   **Intelligenter Dateimanager (` + Backtick + `dev file` + Backtick + `)**: 
    *   Unscharfe Suche nach Dateien und Verzeichnissen.
    *   CRUD-Operationen (Erstellen, Lesen, Aktualisieren, Löschen).
    *   Massenaktionen und Dateivorschau.

### 4. KI & Intelligenz (Das Gehirn)
*   **KI-Chat-Assistent (` + Backtick + `dev chat` + Backtick + `)**: 
    *   **Multi-Modell-Unterstützung**: Wechseln Sie sofort zwischen lokal (Ollama/Llama3) und Cloud (OpenAI GPT-4, Claude 3.5, Gemini 1.5).
    *   **Kontextbewusstsein**: 
        *   Referenzieren Sie Dateien mit ` + Backtick + `@filename` + Backtick + `.
        *   Referenzieren Sie Verzeichnisse mit ` + Backtick + `@dirname` + Backtick + `.
        *   Fügen Sie Inhalte/Bilder aus der Zwischenablage nahtlos ein.
    *   **Modi**: Codierung, Debugging, Kreatives Schreiben, Allgemeiner Chat.
*   **Code-Zeitmaschine (` + Backtick + `dev time` + Backtick + `)**: 
    *   Visueller Git-Verlauf-Explorer.
    *   Gehen Sie durch Commits, um die Code-Evolution zu sehen.
    *   **KI-Analyse**: "Erkläre diesen Commit" und "Finde potenzielle Fehler in diesem Diff".
    *   Blame-Ansicht, um zu sehen, wer was wann geschrieben hat.

---

## ⚙️ Erweiterte Konfiguration

Bearbeiten Sie ` + Backtick + `~/.devcli/config.yaml` + Backtick + ` um Ihre Erfahrung anzupassen.

### KI-Konfiguration
` + Backtick + "```yaml" + Backtick + `
ai:
  default_provider: "ollama"
  providers:
    ollama:
      model: "llama3"
      base_url: "http://localhost:11434"
    openai:
      api_key: "sk-..."
      model: "gpt-4-turbo"
    anthropic:
      api_key: "sk-ant-..."
      model: "claude-3-opus"
` + Backtick + "```" + Backtick + `

### UI-Anpassung
` + Backtick + "```yaml" + Backtick + `
ui:
  theme: "dracula" # optionen: dracula, nord, monokai, system
  editor:
    line_numbers: true
    mouse_support: true
` + Backtick + "```" + Backtick + `

---

## ⌨️ Haupttastenkombinationen

| Kontext | Taste | Aktion |
| :--- | :--- | :--- |
| **Global** | ` + Backtick + `Ctrl+C` + Backtick + ` | Anwendung beenden |
| | ` + Backtick + `Esc` + Backtick + ` | Zurück / Ansicht schließen |
| | ` + Backtick + `?` + Backtick + ` | Hilfe-Overlay umschalten |
| **Navigation** | ` + Backtick + `↑/↓` + Backtick + ` | Listen navigieren |
| | ` + Backtick + `Enter` + Backtick + ` | Auswählen / Ausführen |
| | ` + Backtick + `Tab` + Backtick + ` | Fokus wechseln |
| **Editor** | ` + Backtick + `Ctrl+S` + Backtick + ` | Datei speichern |
| | ` + Backtick + `Ctrl+F` + Backtick + ` | Text suchen |
| **Docs** | ` + Backtick + `L` + Backtick + ` | Sprache wechseln |

---

## ❓ Fehlerbehebung

**F: "Befehl nicht gefunden: ollama"**
A: Stellen Sie sicher, dass Sie Ollama von [ollama.com](https://ollama.com) installiert haben und der Dienst läuft (` + Backtick + `ollama serve` + Backtick + `).

**F: DevCLI startet langsam.**
A: Wenn Sie Tausende von Projekten haben, versuchen Sie, ` + Backtick + `scan_depth: 2` + Backtick + ` in Ihrer config.yaml zu konfigurieren, um die Verzeichnistiefe zu begrenzen.

**F: Git-Integration schlägt fehl.**
A: DevCLI erfordert, dass ` + Backtick + `git` + Backtick + ` installiert und in Ihrem System-PATH verfügbar ist.

---

## 🤝 Community & Support

*   **GitHub**: [github.com/phravins/devcli](https://github.com/phravins/devcli)
*   **Probleme**: Senden Sie Fehlerberichte und Funktionsanfragen auf GitHub.
*   **Beitrag**: PRs sind willkommen! Siehe ` + Backtick + `CONTRIBUTING.md` + Backtick + `.

*Entwickler befähigen, schneller, sauberer und intelligenter zu bauen.*
`
}

func getDocsContentChinese() string {
	return `
# DevCLI - 终极开发者生存工具包

DevCLI 不仅仅是一个工具；它是对开发者体验的彻底重新思考。它将项目管理、编码、调试和 AI 辅助统一到一个单一、连贯的终端界面中。告别碎片化的工作流程和上下文切换。

> **理念**: "保持流畅。" 您需要的一切都在一键之遥。

---

## 🚀 完整功能细分

### 1. 项目管理与脚手架
*   **统一仪表板 (` + Backtick + `dev` + Backtick + `)**:
    *   自动检测项目类型（Go, Python, React, Rust, Node 等）。
    *   显示 git 状态、活动端口、技术栈和最后修改时间。
    *   快速操作：打开、运行、编辑、删除。
*   **即时脚手架 (` + Backtick + `dev new` + Backtick + `)**:
    *   **模板**: 强大、生产就绪的模板，适用于：
        *   **Go**: CHI, Gin, Fiber, Cobra CLI.
        *   **Python**: Flask, FastAPI, Django.
        *   **Frontend**: React (Vite), Vue, Svelte.
        *   **Rust**: Actix-web, Axum.
    *   **智能初始化**: 自动生成 ` + Backtick + `.gitignore` + Backtick + `, ` + Backtick + `Dockerfile` + Backtick + `, ` + Backtick + `Makefile` + Backtick + ` 和 CI/CD 管道。

### 2. 开发强大工具
*   **任务运行器 (` + Backtick + `dev run` + Backtick + `)**:
    *   智能解析 ` + Backtick + `package.json` + Backtick + `, ` + Backtick + `Makefile` + Backtick + `, ` + Backtick + `Justfile` + Backtick + `, ` + Backtick + `go.mod` + Backtick + `。
    *   运行脚本（构建、测试、lint、部署），无需记住特定工具的语法。
*   **开发服务器启动器 (` + Backtick + `dev server` + Backtick + `)**:
    *   自动检测框架（Next.js, Flask, Laravel 等）和启动命令。
    *   自动管理端口以避免冲突。
    *   将实时日志直接流式传输到仪表板。
*   **虚拟环境管理器 (` + Backtick + `dev venv` + Backtick + `)**:
    *   集中查看系统中所有的 Python ` + Backtick + `venvs` + Backtick + ` 和 Node ` + Backtick + `node_modules` + Backtick + `。
    *   一键激活/停用。
    *   深度清理：查找并删除废弃的环境以回收 GB 级的磁盘空间。

### 3. 集成编辑与文件管理
*   **Nano 风格编辑器**:
    *   嵌入在 DevCLI 中的快速、轻量级编辑器。
    *   支持 50+ 种语言的语法高亮显示。
    *   转到行、搜索/替换和撤消/重做。
*   **智能文件管理器 (` + Backtick + `dev file` + Backtick + `)**:
    *   模糊搜索文件和目录。
    *   CRUD 操作（创建、读取、更新、删除）。
    *   批量操作和文件预览。

### 4. AI 与智能 (大脑)
*   **AI 聊天助手 (` + Backtick + `dev chat` + Backtick + `)**:
    *   **多模型支持**: 在本地 (Ollama/Llama3) 和云 (OpenAI GPT-4, Claude 3.5, Gemini 1.5) 之间即时切换。
    *   **上下文感知**:
        *   使用 ` + Backtick + `@filename` + Backtick + ` 引用文件。
        *   使用 ` + Backtick + `@dirname` + Backtick + ` 引用目录。
        *   无缝粘贴剪贴板内容/图像。
    *   **模式**: 编码、调试、创意写作、一般聊天。
*   **代码时光机 (` + Backtick + `dev time` + Backtick + `)**:
    *   可视化 Git 历史探索器。
    *   逐步浏览提交以查看代码演变。
    *   **AI 分析**: "解释此提交" 和 "在此差异中查找潜在的错误"。
    *   责备视图，查看谁在何时编写了什么。

---

## ⚙️ 高级配置

编辑 ` + Backtick + `~/.devcli/config.yaml` + Backtick + ` 以自定义您的体验。

### AI 配置
` + Backtick + "```yaml" + Backtick + `
ai:
  default_provider: "ollama"
  providers:
    ollama:
      model: "llama3"
      base_url: "http://localhost:11434"
    openai:
      api_key: "sk-..."
      model: "gpt-4-turbo"
    anthropic:
      api_key: "sk-ant-..."
      model: "claude-3-opus"
` + Backtick + "```" + Backtick + `

### UI 自定义
` + Backtick + "```yaml" + Backtick + `
ui:
  theme: "dracula" # 选项: dracula, nord, monokai, system
  editor:
    line_numbers: true
    mouse_support: true
` + Backtick + "```" + Backtick + `

---

## ⌨️主要快捷键

| 上下文 | 按键 | 动作 |
| :--- | :--- | :--- |
| **全局** | ` + Backtick + `Ctrl+C` + Backtick + ` | 退出应用程序 |
| | ` + Backtick + `Esc` + Backtick + ` | 返回 / 关闭视图 |
| | ` + Backtick + `?` + Backtick + ` | 切换帮助覆盖层 |
| **导航** | ` + Backtick + `↑/↓` + Backtick + ` | 导航列表 |
| | ` + Backtick + `Enter` + Backtick + ` | 选择 / 执行 |
| | ` + Backtick + `Tab` + Backtick + ` | 切换焦点 |
| **编辑器** | ` + Backtick + `Ctrl+S` + Backtick + ` | 保存文件 |
| | ` + Backtick + `Ctrl+F` + Backtick + ` | 查找文本 |
| **文档** | ` + Backtick + `L` + Backtick + ` | 切换语言 |

---

## ❓ 故障排除

**问: "找不到命令: ollama"**
答: 确保您已从 [ollama.com](https://ollama.com) 安装 Ollama 并且服务正在运行 (` + Backtick + `ollama serve` + Backtick + `)。

**问: DevCLI 启动缓慢。**
答: 如果您有数千个项目，请尝试在 config.yaml 中配置 ` + Backtick + `scan_depth: 2` + Backtick + ` 以限制目录遍历深度。

**问: Git 集成失败。**
答: DevCLI 需要 ` + Backtick + `git` + Backtick + ` 已安装并在您的系统中可用 PATH。

---

## 🤝 社区与支持

*   **GitHub**: [github.com/phravins/devcli](https://github.com/phravins/devcli)
*   **问题**: 在 GitHub 上提交错误报告和功能请求。
*   **贡献**: 欢迎 PR！ 参见 ` + Backtick + `CONTRIBUTING.md` + Backtick + `.

*赋能开发者更快，更干净，更智能地构建。*
`
}

func getDocsContentJapanese() string {
	return `
# DevCLI - 究極の開発者サバイバルキット

DevCLI は単なるツールではありません。開発者体験の完全な再考です。プロジェクト管理、コーディング、デバッグ、AI 支援を単一の統合されたターミナルインターフェースに統合します。断片化されたワークフローとコンテキストの切り替えに別れを告げましょう。

> **哲学**: "フローを維持する。" 必要なものはすべて、キーを押すだけですぐに利用できます。

---

## 🚀 完全な機能内訳

### 1. プロジェクト管理とスキャフォールディング
*   **統合ダッシュボード (` + "`" + `dev` + "`" + `)**: 
    *   プロジェクトタイプ（Go, Python, React, Rust, Node など）を自動検出します。
    *   Git ステータス、アクティブポート、技術スタック、最終更新時間を表示します。
    *   クイックアクション：開く、実行、編集、削除。
*   **インスタント・スキャフォールディング (` + "`" + `dev new` + "`" + `)**: 
    *   **テンプレート**: 以下のための堅牢で本番環境対応のテンプレート：
        *   **Go**: CHI, Gin, Fiber, Cobra CLI.
        *   **Python**: Flask, FastAPI, Django.
        *   **Frontend**: React (Vite), Vue, Svelte.
        *   **Rust**: Actix-web, Axum.
    *   **スマート初期化**: ` + "`" + `.gitignore` + "`" + `, ` + "`" + `Dockerfile` + "`" + `, ` + "`" + `Makefile` + "`" + `, CI/CD パイプラインを自動生成します。

### 2. 開発パワーツール
*   **タスクランナー (` + "`" + `dev run` + "`" + `)**: 
    *   ` + "`" + `package.json` + "`" + `, ` + "`" + `Makefile` + "`" + `, ` + "`" + `Justfile` + "`" + `, ` + "`" + `go.mod` + "`" + ` をインテリジェントに解析します。
    *   特定のツール構文を覚える必要なく、スクリプト（ビルド、テスト、リント、デプロイ）を実行します。
*   **開発サーバーランチャー (` + "`" + `dev server` + "`" + `)**: 
    *   フレームワーク（Next.js, Flask, Laravel など）と開始コマンドを自動検出します。
    *   競合を避けるためにポートを自動管理します。
    *   ライブログをダッシュボードに直接ストリーミングします。
*   **仮想環境マネージャー (` + "`" + `dev venv` + "`" + `)**: 
    *   システム上のすべての Python ` + "`" + `venvs` + "`" + ` と Node ` + "`" + `node_modules` + "`" + ` の一元管理ビュー。
    *   ワンクリックで有効化/無効化。
    *   ディープクリーン：放棄された環境を見つけて削除し、GB 単位のディスク容量を再利用します。

### 3. 統合編集とファイル管理
*   **Nano スタイルエディタ**: 
    *   DevCLI に組み込まれた高速で軽量なエディタ。
    *   50以上の言語のシンタックスハイライト。
    *   行への移動、検索/置換、元に戻す/やり直し。
*   **スマートファイルマネージャー (` + "`" + `dev file` + "`" + `)**: 
    *   ファイルとディレクトリのあいまい検索。
    *   CRUD 操作（作成、読み取り、更新、削除）。
    *   一括アクションとファイルプレビュー。

### 4. AI とインテリジェンス (頭脳)
*   **AI チャットアシスタント (` + "`" + `dev chat` + "`" + `)**: 
    *   **マルチモデルサポート**: ローカル (Ollama/Llama3) とクラウド (OpenAI GPT-4, Claude 3.5, Gemini 1.5) を即座に切り替えます。
    *   **コンテキスト認識**: 
        *   ` + "`" + `@filename` + "`" + ` でファイルを参照。
        *   ` + "`" + `@dirname` + "`" + ` でディレクトリを参照。
        *   クリップボードのコンテンツ/画像をシームレスに貼り付け。
    *   **モード**: コーディング、デバッグ、クリエイティブライティング、一般チャット。
*   **コードタイムマシン (` + "`" + `dev time` + "`" + `)**: 
    *   ビジュアル Git 履歴エクスプローラー。
    *   コミットをステップスルーしてコードの進化を確認。
    *   **AI 分析**: "このコミットを説明" および "この差分で潜在的なバグを見つける"。
    *   誰がいつ何を書いたかを確認するための Blame ビュー。

---

## ⚙️ 高度な設定

` + "`" + `~/.devcli/config.yaml` + "`" + ` を編集して、エクスペリエンスをカスタマイズします。

### AI 設定
` + "`" + "```yaml" + "`" + `
ai:
  default_provider: "ollama"
  providers:
    ollama:
      model: "llama3"
      base_url: "http://localhost:11434"
    openai:
      api_key: "sk-..."
      model: "gpt-4-turbo"
    anthropic:
      api_key: "sk-ant-..."
      model: "claude-3-opus"
` + "`" + "```" + "`" + `

### UI カスタマイズ
` + "`" + "```yaml" + "`" + `
ui:
  theme: "dracula" # オプション: dracula, nord, monokai, system
  editor:
    line_numbers: true
    mouse_support: true
` + "`" + "```" + "`" + `

---

## ⌨️ マスターキーバインディング

| コンテキスト | キー | アクション |
| :--- | :--- | :--- |
| **グローバル** | ` + "`" + `Ctrl+C` + "`" + ` | アプリケーションを終了 |
| | ` + "`" + `Esc` + "`" + ` | 戻る / ビューを閉じる |
| | ` + "`" + `?` + "`" + ` | ヘルプオーバーレイを切り替え |
| **ナビゲーション** | ` + "`" + `↑/↓` + "`" + ` | リストをナビゲート |
| | ` + "`" + `Enter` + "`" + ` | 選択 / 実行 |
| | ` + "`" + `Tab` + "`" + ` | フォーカスを切り替え |
| **エディタ** | ` + "`" + `Ctrl+S` + "`" + ` | ファイルを保存 |
| | ` + "`" + `Ctrl+F` + "`" + ` | テキストを検索 |
| **ドキュメント** | ` + "`" + `L` + "`" + ` | 言語を切り替え |

---

## ❓ トラブルシューティング

**Q: "コマンドが見つかりません: ollama"**
A: [ollama.com](https://ollama.com) から Ollama をインストールし、サービスが実行されていること (` + "`" + `ollama serve` + "`" + `) を確認してください。

**Q: DevCLI の起動が遅い。**
A: 数千のプロジェクトがある場合は、config.yaml で ` + "`" + `scan_depth: 2` + "`" + ` を設定して、ディレクトリトラバーサルの深さを制限してみてください。

**Q: Git 統合が失敗する。**
A: DevCLI を使用するには、` + "`" + `git` + "`" + ` がインストールされ、システム PATH で利用可能である必要があります。

---

## 🤝 コミュニティとサポート

*   **GitHub**: [github.com/phravins/devcli](https://github.com/phravins/devcli)
*   **問題**: GitHub でバグレポートや機能リクエストを送信してください。
*   **貢献**: PR 歓迎！ ` + "`" + `CONTRIBUTING.md` + "`" + ` を参照してください。

*開発者がより速く、よりクリーンに、よりスマートに構築できるようにします。*
`
}
