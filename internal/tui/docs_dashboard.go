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
	LangEnglish = "en"
	LangSpanish = "es"
	LangHindi   = "hi"
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
				m.language = LangHindi
			case LangHindi:
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
	case LangHindi:
		return "Hindi"
	}
	return "English"
}

func getDocsContent(lang string) string {
	switch lang {
	case LangSpanish:
		return getDocsContentSpanish()
	case LangHindi:
		return getDocsContentHindi()
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

func getDocsContentHindi() string {
	return `
# DevCLI - यूनिफाइड डेवलपर वर्कस्पेस

DevCLI एक टर्मिनल-आधारित पावर टूल है जिसे आपके पूरे डेवलपमेंट वर्कफ़्लो को एक सिंगल, कीबोर्ड-संचालित इंटरफ़ेस में समेकित करने के लिए डिज़ाइन किया गया है।

> **दर्शन**: "प्रवाह में रहें।" DevCLI आपके टूल को सीधे आपके टर्मिनल में लाता है।

---

## 🚀 मुख्य विशेषताएं

### 1. प्रोजेक्ट प्रबंधन
*   **प्रोजेक्ट डैशबोर्ड**: अपने सभी प्रोजेक्ट्स (स्थिति, तकनीक स्टैक, अंतिम संशोधन) का विहंगम दृश्य प्राप्त करें।
*   **वन-क्लिक स्कैफोल्डिंग**: Go, Python, Node.js, React, आदि में प्रोडक्शन-रेडी प्रोजेक्ट्स बनाएं।
*   **टास्क रनर**: स्वचालित रूप से ` + "`package.json`" + `, ` + "`Makefile`" + `, ` + "`go.mod`" + `, आदि का पता लगाता है, और आपको बिल्ड/टेस्ट कमांड तुरंत चलाने देता है।
*   **स्मार्ट फाइल क्रिएटर**: सेकंड में ` + "`.gitignore`" + `, ` + "`Dockerfile`" + `, ` + "`README.md`" + ` उत्पन्न करें।

### 2. डेवलपमेंट एनवायरनमेंट
*   **डेव सर्वर लॉन्चर**: स्वचालित रूप से आपके वेब फ्रेमवर्क (Next.js, Flask, Django) का पता लगाता है और लाइव लॉग स्ट्रीमिंग के साथ डेव सर्वर लॉन्च करता है।
*   **वर्चुअल एनवायरनमेंट विजार्ड**: Python ` + "`venvs`" + ` और Node ` + "`node_modules`" + ` के लिए केंद्रीकृत प्रबंधन। डिस्क स्थान बचाने के लिए स्कैन, सिंक और क्लीन अप करें।
*   **बिल्ट-इन एडिटर**: DevCLI को छोड़े बिना त्वरित संपादन के लिए सिंटैक्स हाइलाइटिंग के साथ एक हल्का एडिटर।
*   **फाइल मैनेजर**: फजी सर्च और फाइल ऑपरेशंस के साथ एक पूरी तरह से कार्यात्मक फाइल एक्सप्लोरर।

### 3. एआई और विश्लेषण
*   **एआई असिस्टेंट**: सीधे अपने टर्मिनल में एलएलएम (Ollama, OpenAI, Claude, Gemini) के साथ चैट करें। संदर्भ-जागरूक कोड जनरेशन और डिबगिंग।
*   **कोड टाइम मशीन**: Git इतिहास के लिए एक दृश्य इंटरफ़ेस। कमिट्स के माध्यम से कदम बढ़ाएं और एआई-संचालित बग जोखिम विश्लेषण प्राप्त करें।

---

## ⚙️ कॉन्फ़िगरेशन

DevCLI अपना कॉन्फ़िगरेशन ` + "`~/.devcli/config.yaml`" + ` में स्टोर करता है।

### एआई प्रदाता
आप कई एआई बैकएंड कॉन्फ़िगर कर सकते हैं। मुख्य मेनू में **सेटिंग्स** पर जाएं।

` + "```yaml" + `
ai:
  provider: "ollama" # या "openai", "anthropic", "gemini"
  model: "llama3"    # मॉडल का नाम
  api_key: ""        # क्लाउड प्रदाताओं के लिए आवश्यक
` + "```" + `

---

## ⌨️ ग्लोबल शॉर्टकट्स

| कुंजी | कार्रवाई |
| :--- | :--- |
| **Ctrl+C** | एप्लिकेशन छोड़ें |
| **Esc / Q** | वापस जाएं / दृश्य बंद करें |
| **Arrow Keys** | मेनू और सूचियों नेविगेट करें |
| **Enter** | चुनें / पुष्टि करें |
| **?** | मदद दिखाएं |

---

## 🤝 योगदान

DevCLI ओपन सोर्स है! हम योगदान का स्वागत करते हैं।
*   **रिपो**: https://github.com/phravins/devcli

*Go, Bubble Tea, और Lip Gloss के साथ ❤️ से बनाया गया।*
`
}
