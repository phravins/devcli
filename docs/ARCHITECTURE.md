# Architecture & Dependencies

## Project Structure

DevCLI is modular, separating UI from business logic:

*   `cmd/` - Command-line entry points
*   `internal/` - Private application code
    *   `ai/` - AI provider implementations
    *   `tui/` - Terminal UI components (Bubble Tea)
    *   `project/` - Project management logic
    *   `...` (other feature modules)

## Dependencies

### Core
*   **Bubble Tea** (v1.3.4): TUI framework
*   **Bubbles** (v0.21.0): UI components
*   **Lipgloss** (v1.1.1): Styling
*   **Glamour** (v0.10.0): Markdown rendering
*   **Cobra** (v1.8.0): CLI framework
*   **Viper** (v1.21.0): Configuration

### Utilities
*   **Chroma**: Syntax highlighting
*   **Fuzzy**: Fuzzy string matching
*   **YAML v2**: YAML parsing

## Dependency Management

Dependencies are managed via Go modules.

*   View dependency tree: `go mod graph`
*   Update dependencies: `go get -u ./...` && `go mod tidy`
