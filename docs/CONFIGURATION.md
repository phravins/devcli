# DevCLI v1.1.0 Configuration Guide

DevCLI automatically persists system preferences and API keys in `~/.devcli.yaml` (or `%USERPROFILE%\.devcli.yaml` on Windows).

---

## Configuration Options

| Key | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `ai_backend` | string | `ollama` | Selected AI backend (`ollama`, `gemini`, `openai`, `claude`, `huggingface`) |
| `ai_model` | string | `mistral` | Default AI model name |
| `ai_api_key` | string | `""` | Global / OpenAI API key |
| `gemini_api_key` | string | `""` | Google Gemini API key |
| `anthropic_api_key` | string | `""` | Anthropic Claude API key |
| `hf_access_token` | string | `""` | HuggingFace user access token |
| `ai_base_url` | string | `http://localhost:11434` | Custom endpoint URL for local Ollama or OpenAI-compatible server |
| `editor_theme` | string | `default` | Code editor theme selection |

---

## Example Config File (`~/.devcli.yaml`)

```yaml
ai_backend: gemini
ai_model: gemini-1.5-flash
gemini_api_key: AIzaSyYourKeyHere...
editor_theme: dracula
user_name: Developer
```

---

## Configuring Keys via TUI

You can update configuration keys interactively inside DevCLI:
1. Launch `devcli`.
2. Select **`⚙️ Settings / Configuration`** or **`🔄 Auto-Update` -> `Update AI Keys`**.
3. Input your API key and press **[Enter]** to save instantly.
