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
# DevCLI - Unified Developer Workspace

DevCLI is a terminal-based power tool designed to consolidate your entire development workflow into a single, keyboard-driven interface. It replaces scattered scripts and context switching with a unified dashboard for project management, coding, and AI assistance.

> **Philosophy**: "Stay in the flow." DevCLI brings your tools to you, right in your terminal.

---

## 🚀 Key Features

### 1. Project Management
*   **Project Dashboard**: Get a bird's-eye view of all your projects (status, tech stack, last modified).
*   **One-Click Scaffolding**: Create production-ready projects in Go, Python, Node.js, React, and more.
*   **Task Runner**: Auto-detects ` + "`package.json`" + `, ` + "`Makefile`" + `, ` + "`go.mod`" + `, etc., and lets you run build/test commands instantly.
*   **Smart File Creator**: Generate ` + "`.gitignore`" + `, ` + "`Dockerfile`" + `, ` + "`README.md`" + `, or CI/CD configs in seconds.

### 2. Development Environment
*   **Dev Server Launcher**: Automatically detects your web framework (Next.js, Flask, Django) and launches the dev server with live log streaming.
*   **Virtual Environment Wizard**: Centralized management for Python ` + "`venvs`" + ` and Node ` + "`node_modules`" + `. scan, sync, and clean up to save disk space.
*   **Built-in Editor**: A lightweight, nano-like editor with syntax highlighting for quick edits without leaving DevCLI.
*   **File Manager**: A fully functional file explorer with fuzzy search and file operations.

### 3. AI & Analysis
*   **AI Assistant**: Chat with LLMs (Ollama, OpenAI, Claude, Gemini) directly in your terminal. Context-aware code generation and debugging.
*   **Code Time Machine**: A visual interface for Git history. Step through commits, see blame annotations, and get AI-powered bug risk analysis.
*   **Snippet Library**: Save and organize your favorite code blocks for instant reuse.

---

## ⚙️ Configuration

DevCLI stores its configuration in ` + "`~/.devcli/config.yaml`" + ` (or ` + "`%USERPROFILE%\\.devcli\\config.yaml`" + ` on Windows).

### AI Providers
You can configure multiple AI backends. Go to **Settings** in the main menu or edit the config file directly.

` + "```yaml" + `
ai:
  provider: "ollama" # or "openai", "anthropic", "gemini"
  model: "llama3"    # Model name
  api_key: ""        # Required for cloud providers
  base_url: ""       # Optional custom endpoint
` + "```" + `

### Customizing Styles
DevCLI supports themes. Currently, it defaults to an adaptive theme based on your terminal's background color.

---

## ⌨️ Global Shortcuts

| Key | Action |
| :--- | :--- |
| **Ctrl+C** | Quit Application |
| **Esc / Q** | Go Back / Close View |
| **Arrow Keys** | Navigate Menus & Lists |
| **Enter** | Select / Confirm |
| **?** | Show Help (Context Sensitive) |
| **Ctrl+L** | Clear Screen / Redraw |

---

## ❓ FAQ & Troubleshooting

**Q: DevCLI doesn't detect my project type.**
A: Ensure your project has standard marker files like ` + "`package.json`" + `, ` + "`go.mod`" + `, ` + "`requirements.txt`" + `, or ` + "`pom.xml`" + `.

**Q: The AI Assistant isn't working.**
A: 
1. Check your internet connection.
2. If using **Ollama**, ensure the service is running (` + "`ollama serve`" + `).
3. If using **OpenAI/Claude**, verify your API key in **Settings**.

**Q: How do I update DevCLI?**
A: Go to **Bonus Features** -> **Check for Updates**. DevCLI can self-update by pulling the latest code and rebuilding itself.

---

## 🤝 Contributing

DevCLI is open source! We welcome contributions.
*   **Repo**: https://github.com/phravins/devcli
*   **Issues**: Report bugs or request features on GitHub issues.

*Built with ❤️ using Go, Bubble Tea, and Lip Gloss.*
`
}

func getDocsContentSpanish() string {
	return `
# DevCLI - Espacio de Trabajo Unificado para Desarrolladores

DevCLI es una herramienta de terminal diseñada para consolidar todo tu flujo de trabajo de desarrollo en una única interfaz controlada por teclado. Reemplaza scripts dispersos y cambios de contexto con un panel unificado para la gestión de proyectos, programación y asistencia de IA.

> **Filosofía**: "Mantente en el flujo." DevCLI trae tus herramientas hacia ti, directamente en tu terminal.

---

## 🚀 Características Principales

### 1. Gestión de Proyectos
*   **Panel de Proyectos**: Obtén una vista general de todos tus proyectos (estado, pila tecnológica, última modificación).
*   **Andamiaje con Un Clic**: Crea proyectos listos para producción en Go, Python, Node.js, React y más.
*   **Ejecutor de Tareas**: Detecta automáticamente ` + "`package.json`" + `, ` + "`Makefile`" + `, ` + "`go.mod`" + `, etc., y te permite ejecutar comandos de construcción/prueba al instante.
*   **Creador Inteligente de Archivos**: Genera ` + "`.gitignore`" + `, ` + "`Dockerfile`" + `, ` + "`README.md`" + `, o configuraciones de CI/CD en segundos.

### 2. Entorno de Desarrollo
*   **Lanzador de Servidor de Desarrollo**: Detecta automáticamente tu marco web (Next.js, Flask, Django) y lanza el servidor de desarrollo con transmisión de registros en vivo.
*   **Asistente de Entorno Virtual**: Gestión centralizada para ` + "`venvs`" + ` de Python y ` + "`node_modules`" + ` de Node. Escanea, sincroniza y limpia para ahorrar espacio en disco.
*   **Editor Integrado**: Un editor ligero similar a nano con resaltado de sintaxis para ediciones rápidas sin salir de DevCLI.
*   **Gestor de Archivos**: Un explorador de archivos completamente funcional con búsqueda difusa y operaciones de archivos.

### 3. IA y Análisis
*   **Asistente de IA**: Chatea con LLMs (Ollama, OpenAI, Claude, Gemini) directamente en tu terminal. Generación de código y depuración conscientes del contexto.
*   **Máquina del Tiempo de Código**: Una interfaz visual para el historial de Git. Recorre commits, ve anotaciones de culpabilidad y obtén análisis de riesgo de errores impulsados por IA.
*   **Biblioteca de Fragmentos**: Guarda y organiza tus bloques de código favoritos para su reutilización instantánea.

---

## ⚙️ Configuración

DevCLI almacena su configuración en ` + "`~/.devcli/config.yaml`" + ` (o ` + "`%USERPROFILE%\\.devcli\\config.yaml`" + ` en Windows).

### Proveedores de IA
Puedes configurar múltiples backends de IA. Ve a **Configuración** en el menú principal o edita el archivo de configuración directamente.

` + "```yaml" + `
ai:
  provider: "ollama" # o "openai", "anthropic", "gemini"
  model: "llama3"    # Nombre del modelo
  api_key: ""        # Requerido para proveedores en la nube
  base_url: ""       # Endpoint personalizado opcional
` + "```" + `

---

## ⌨️ Atajos Globales

| Tecla | Acción |
| :--- | :--- |
| **Ctrl+C** | Salir de la Aplicación |
| **Esc / Q** | Volver / Cerrar Vista |
| **Flechas** | Navegar Menús y Listas |
| **Enter** | Seleccionar / Confirmar |
| **?** | Mostrar Ayuda |

---

## 🤝 Contribuyendo

¡DevCLI es de código abierto! Agradecemos las contribuciones.
*   **Repo**: https://github.com/phravins/devcli

*Creado con ❤️ usando Go, Bubble Tea y Lip Gloss.*
`
}

func getDocsContentFrench() string {
	return `
# DevCLI - Espace de Travail Unifié pour Développeurs

DevCLI est un outil puissant basé sur le terminal conçu pour consolider l'ensemble de votre flux de travail de développement dans une interface unique pilotée par clavier. Il remplace les scripts dispersés et les changements de contexte par un tableau de bord unifié pour la gestion de projets, le codage et l'assistance IA.

> **Philosophie**: "Restez dans le flux." DevCLI apporte vos outils à vous, directement dans votre terminal.

---

## 🚀 Fonctionnalités Clés

### 1. Gestion de Projets
*   **Tableau de Bord de Projets**: Obtenez une vue d'ensemble de tous vos projets (statut, pile technologique, dernière modification).
*   **Échafaudage en Un Clic**: Créez des projets prêts pour la production en Go, Python, Node.js, React, et plus encore.
*   **Exécuteur de Tâches**: Détecte automatiquement ` + "`package.json`" + `, ` + "`Makefile`" + `, ` + "`go.mod`" + `, etc., et vous permet d'exécuter des commandes de construction/test instantanément.
*   **Créateur de Fichiers Intelligent**: Générez ` + "`.gitignore`" + `, ` + "`Dockerfile`" + `, ` + "`README.md`" + `, ou des configurations CI/CD en quelques secondes.

### 2. Environnement de Développement
*   **Lanceur de Serveur de Dév**: Détecte automatiquement votre framework web (Next.js, Flask, Django) et lance le serveur de développement avec diffusion de logs en direct.
*   **Assistant d'Environnement Virtuel**: Gestion centralisée pour ` + "`venvs`" + ` Python et ` + "`node_modules`" + ` Node. Scannez, synchronisez et nettoyez pour économiser de l'espace disque.
*   **Éditeur Intégré**: Un éditeur léger de type nano avec coloration syntaxique pour des modifications rapides sans quitter DevCLI.
*   **Gestionnaire de Fichiers**: Un explorateur de fichiers entièrement fonctionnel avec recherche floue et opérations sur les fichiers.

### 3. IA & Analyse
*   **Assistant IA**: Discutez avec des LLM (Ollama, OpenAI, Claude, Gemini) directement dans votre terminal. Génération de code et débogage conscients du contexte.
*   **Machine à Remonter le Temps du Code**: Une interface visuelle pour l'historique Git. Parcourez les commits, voyez les annotations de blâme et obtenez une analyse des risques de bugs alimentée par l'IA.

---

## ⚙️ Configuration

DevCLI stocke sa configuration dans ` + "`~/.devcli/config.yaml`" + ` (ou ` + "`%USERPROFILE%\\.devcli\\config.yaml`" + ` sous Windows).

### Fournisseurs d'IA
Vous pouvez configurer plusieurs backends d'IA. Allez dans **Paramètres** dans le menu principal ou modifiez le fichier de configuration directement.

` + "```yaml" + `
ai:
  provider: "ollama" # ou "openai", "anthropic", "gemini"
  model: "llama3"    # Nom du modèle
  api_key: ""        # Requis pour les fournisseurs cloud
` + "```" + `

---

## ⌨️ Raccourcis Globaux

| Touche | Action |
| :--- | :--- |
| **Ctrl+C** | Quitter l'Application |
| **Esc / Q** | Retour / Fermer la Vue |
| **Flèches** | Naviguer dans les Menus & Listes |
| **Entrée** | Sélectionner / Confirmer |
| **?** | Afficher l'Aide |

---

## 🤝 Contribuer

DevCLI est open source ! Nous accueillons les contributions.
*   **Repo**: https://github.com/phravins/devcli

*Construit avec ❤️ en utilisant Go, Bubble Tea, et Lip Gloss.*
`
}

func getDocsContentGerman() string {
	return `
# DevCLI - Einheitlicher Entwickler-Arbeitsbereich

DevCLI ist ein terminalbasiertes Power-Tool, das entwickelt wurde, um Ihren gesamten Entwicklungsworkflow in einer einzigen, tastaturgesteuerten Oberfläche zu konsolidieren. Es ersetzt verstreute Skripte und Kontextwechsel durch ein einheitliches Dashboard für Projektmanagement, Codierung und KI-Unterstützung.

> **Philosophie**: "Im Fluss bleiben." DevCLI bringt Ihre Werkzeuge zu Ihnen, direkt in Ihr Terminal.

---

## 🚀 Hauptmerkmale

### 1. Projektmanagement
*   **Projekt-Dashboard**: Erhalten Sie einen Überblick über alle Ihre Projekte (Status, Tech-Stack, letzte Änderung).
*   **Ein-Klick-Gerüstbau**: Erstellen Sie produktionsreife Projekte in Go, Python, Node.js, React und mehr.
*   **Task Runner**: Erkennt automatisch ` + "`package.json`" + `, ` + "`Makefile`" + `, ` + "`go.mod`" + ` usw. und ermöglicht Ihnen das sofortige Ausführen von Build-/Testbefehlen.
*   **Intelligenter Dateiersteller**: Generieren Sie ` + "`.gitignore`" + `, ` + "`Dockerfile`" + `, ` + "`README.md`" + ` oder CI/CD-Konfigurationen in Sekunden.

### 2. Entwicklungsumgebung
*   **Dev-Server-Launcher**: Erkennt automatisch Ihr Web-Framework (Next.js, Flask, Django) und startet den Entwicklungsserver mit Live-Log-Streaming.
*   **Assistent für virtuelle Umgebungen**: Zentrale Verwaltung für Python ` + "`venvs`" + ` und Node ` + "`node_modules`" + `. Scannen, synchronisieren und bereinigen, um Speicherplatz zu sparen.
*   **Integrierter Editor**: Ein leichter, nano-ähnlicher Editor mit Syntaxhervorhebung für schnelle Bearbeitungen, ohne DevCLI zu verlassen.
*   **Dateimanager**: Ein voll funktionsfähiger Datei-Explorer mit unscharfer Suche und Dateioperationen.

### 3. KI & Analyse
*   **KI-Assistent**: Chatten Sie mit LLMs (Ollama, OpenAI, Claude, Gemini) direkt in Ihrem Terminal. Kontextbezogene Codegenerierung und Debugging.
*   **Code-Zeitmaschine**: Eine visuelle Oberfläche für den Git-Verlauf. Gehen Sie Commits durch, sehen Sie Schuldzuweisungen und erhalten Sie KI-gestützte Fehler-Risikoanalysen.

---

## ⚙️ Konfiguration

DevCLI speichert seine Konfiguration in ` + "`~/.devcli/config.yaml`" + ` (oder ` + "`%USERPROFILE%\\.devcli\\config.yaml`" + ` unter Windows).

### KI-Anbieter
Sie können mehrere KI-Backends konfigurieren. Gehen Sie im Hauptmenü zu **Einstellungen** oder bearbeiten Sie die Konfigurationsdatei direkt.

` + "```yaml" + `
ai:
  provider: "ollama" # oder "openai", "anthropic", "gemini"
  model: "llama3"    # Modellname
  api_key: ""        # Erforderlich für Cloud-Anbieter
` + "```" + `

---

## ⌨️ Globale Tastenkombinationen

| Taste | Aktion |
| :--- | :--- |
| **Ctrl+C** | Anwendung beenden |
| **Esc / Q** | Zurück / Ansicht schließen |
| **Pfeiltasten** | Menüs & Listen navigieren |
| **Enter** | Auswählen / Bestätigen |
| **?** | Hilfe anzeigen |

---

## 🤝 Mitwirken

DevCLI ist Open Source! Wir freuen uns über Beiträge.
*   **Repo**: https://github.com/phravins/devcli

*Erstellt mit ❤️ unter Verwendung von Go, Bubble Tea und Lip Gloss.*
`
}

func getDocsContentChinese() string {
	return `
# DevCLI - 统一开发者工作区

DevCLI 是一个基于终端的强大工具，旨在将您的整个开发工作流程整合到一个单一的、键盘驱动的界面中。它用一个统一的项目管理、编码和 AI 辅助仪表板取代了分散的脚本和上下文切换。

> **理念**: "保持流畅。" DevCLI 将您的工具带给您，就在您的终端里。

---

## 🚀 主要功能

### 1. 项目管理
*   **项目仪表板**: 鸟瞰您的所有项目（状态、技术栈、最后修改）。
*   **一键脚手架**: 创建 Go, Python, Node.js, React 等生产就绪的项目。
*   **任务运行器**: 自动检测 ` + "`package.json`" + `, ` + "`Makefile`" + `, ` + "`go.mod`" + ` 等，并允许您立即运行构建/测试命令。
*   **智能文件创建器**: 在几秒钟内生成 ` + "`.gitignore`" + `, ` + "`Dockerfile`" + `, ` + "`README.md`" + ` 或 CI/CD 配置。

### 2. 开发环境
*   **开发服务器启动器**: 自动检测您的 Web 框架（Next.js, Flask, Django）并启动带有实时日志流的开发服务器。
*   **虚拟环境向导**: Python ` + "`venvs`" + ` 和 Node ` + "`node_modules`" + ` 的集中管理。扫描、同步和清理以节省磁盘空间。
*   **内置编辑器**: 一个轻量级的、类似 nano 的编辑器，带有语法高亮显示，无需离开 DevCLI 即可快速编辑。
*   **文件管理器**: 一个功能齐全的文件资源管理器，带有模糊搜索和文件操作功能。

### 3. AI & 分析
*   **AI 助手**: 直接在您的终端中与 LLM（Ollama, OpenAI, Claude, Gemini）聊天。上下文感知的代码生成和调试。
*   **代码时光机**: Git 历史记录的可视化界面。逐步浏览提交，查看责任注释，并获得 AI 驱动的错误风险分析。

---

## ⚙️ 配置

DevCLI 将其配置存储在 ` + "`~/.devcli/config.yaml`" + `（或 Windows 上的 ` + "`%USERPROFILE%\\.devcli\\config.yaml`" + `）。

### AI 提供商
您可以配置多个 AI 后端。转到主菜单中的 **设置** 或直接编辑配置文件。

` + "```yaml" + `
ai:
  provider: "ollama" # 或 "openai", "anthropic", "gemini"
  model: "llama3"    # 模型名称
  api_key: ""        # 云提供商所需
` + "```" + `

---

## ⌨️ 全局快捷键

| 按键 | 动作 |
| :--- | :--- |
| **Ctrl+C** | 退出应用程序 |
| **Esc / Q** | 返回 / 关闭视图 |
| **方向键** | 导航菜单和列表 |
| **Enter** | 选择 / 确认 |
| **?** | 显示帮助 |

---

## 🤝 贡献

DevCLI 是开源的！我们欢迎贡献。
*   **仓库**: https://github.com/phravins/devcli

*使用 Go, Bubble Tea, 和 Lip Gloss 用 ❤️ 构建。*
`
}

func getDocsContentJapanese() string {
	return `
# DevCLI - 統合開発者ワークスペース

DevCLI は、開発ワークフロー全体を単一のキーボード駆動インターフェースに統合するために設計された、ターミナルベースのパワーツールです。分散したスクリプトやコンテキストの切り替えを、プロジェクト管理、コーディング、AI 支援のための統一ダッシュボードに置き換えます。

> **哲学**: "フローを維持する。" DevCLI は、ツールをあなたのターミナルに直接提供します。

---

## 🚀 主な機能

### 1. プロジェクト管理
*   **プロジェクトダッシュボード**: すべてのプロジェクト（ステータス、技術スタック、最終変更）を俯瞰できます。
*   **ワンクリック・スキャフォールディング**: Go, Python, Node.js, React などで本番環境対応のプロジェクトを作成します。
*   **タスクランナー**: ` + "`package.json`" + `, ` + "`Makefile`" + `, ` + "`go.mod`" + ` などを自動検出し、ビルド/テストコマンドを即座に実行できるようにします。
*   **スマートファイルクリエイター**: ` + "`.gitignore`" + `, ` + "`Dockerfile`" + `, ` + "`README.md`" + `, または CI/CD 設定を数秒で生成します。

### 2. 開発環境
*   **開発サーバーランチャー**: Web フレームワーク（Next.js, Flask, Django）を自動検出し、ライブログストリーミング付きで開発サーバーを起動します。
*   **仮想環境ウィザード**: Python ` + "`venvs`" + ` と Node ` + "`node_modules`" + ` の集中管理。スキャン、同期、クリーンアップしてディスク容量を節約します。
*   **組み込みエディタ**: DevCLI を離れることなく素早く編集できる、シンタックスハイライト付きの軽量な nano 風エディタ。
*   **ファイルマネージャー**: あいまい検索とファイル操作を備えた完全に機能するファイルエクスプローラー。

### 3. AI & 分析
*   **AI アシスタント**: ターミナルで直接 LLM（Ollama, OpenAI, Claude, Gemini）とチャットできます。コンテキスト認識型のコード生成とデバッグ。
*   **コードタイムマシン**: Git 履歴のビジュアルインターフェース。コミットをステップスルーし、blame 注釈を確認し、AI 駆動のバグリスク分析を取得します。

---

## ⚙️ 設定

DevCLI は設定を ` + "`~/.devcli/config.yaml`" + `（または Windows の ` + "`%USERPROFILE%\\.devcli\\config.yaml`" + `）に保存します。

### AI プロバイダー
複数の AI バックエンドを設定できます。メインメニューの **Settings** に移動するか、設定ファイルを直接編集してください。

` + "```yaml" + `
ai:
  provider: "ollama" # または "openai", "anthropic", "gemini"
  model: "llama3"    # モデル名
  api_key: ""        # クラウドプロバイダーに必要
` + "```" + `

---

## ⌨️ グローバルショートカット

| キー | 動作 |
| :--- | :--- |
| **Ctrl+C** | アプリケーションを終了 |
| **Esc / Q** | 戻る / ビューを閉じる |
| **矢印キー** | メニューとリストをナビゲート |
| **Enter** | 選択 / 確認 |
| **?** | ヘルプを表示 |

---

## 🤝 貢献

DevCLI はオープンソースです！貢献を歓迎します。
*   **リポジトリ**: https://github.com/phravins/devcli

*Go, Bubble Tea, と Lip Gloss を使用して ❤️ で構築されました。*
`
}
