# ⚡ ZAP

> AI-powered API testing that understands your codebase

**ZAP** is the terminal-based AI assistant that doesn't just test your APIs—it debugs them. When an endpoint returns an error, ZAP searches your actual code to find the cause and suggests fixes. Uses local LLMs (Ollama) or cloud providers (OpenAI/Anthropic).

## 🚀 Quick Start

### Prerequisites

- Go 1.25.3 or higher
- [Ollama](https://ollama.ai/) (optional, for local AI)

### Build and Run

```bash
git clone https://github.com/blackcoderx/zap.git
cd zap
go build -o zap.exe ./cmd/zap
./zap
```

**First run:**
1. Creates `.zap` folder with config, history, and memory
2. Prompts you to select your API framework (gin, fastapi, express, etc.)
3. Launches interactive TUI

**Try it:**
```bash
# In the TUI, type:
> GET http://localhost:8000/api/users

# ZAP will make the request, show the response, and if there's an error,
# it can search your code to find the cause
```

## ✨ Features

### 🧠 **Codebase-Aware Debugging**
- Analyzes error responses and searches your code for causes
- Parses stack traces (Python/Go/JavaScript)
- Framework-specific hints for common errors
- Suggests fixes with code examples

### 🛠️ **28+ Tools Across API Testing**

**Core API Tools:**
- HTTP requests with status explanations and error hints
- Save/load requests as YAML with `{{VAR}}` substitution
- Environment switching (dev/prod/staging)

**Testing & Validation:**
- Response assertions (status, headers, body, JSON path, timing)
- JSON Schema validation (draft-07, draft-2020-12)
- Test suite execution with pass/fail reporting
- Regression testing with baseline comparison

**Request Chaining:**
- Extract values from responses (JSON path, headers, cookies, regex)
- Variable management (session/global with persistence)
- Retry with exponential backoff

**Authentication:**
- OAuth2 (client_credentials, password flows)
- JWT/Bearer tokens with automatic header creation
- HTTP Basic auth
- JWT token parsing (decode claims, expiration)

**Performance:**
- Load testing with concurrent users
- Latency metrics (p50/p95/p99)
- Webhook listener (temporary HTTP server for callbacks)
- Wait/delay tools for async operations

**Codebase Analysis:**
- Read files (with 100KB security limit)
- Write/modify files with human-in-the-loop confirmation
- Search code patterns (ripgrep with fallback)
- List files with glob patterns

### 💅 **Beautiful Terminal Interface**
- Minimal Claude Code-style UI with Charm stack
- Streaming responses (text appears as it arrives)
- Markdown rendering with syntax highlighting
- Input history navigation (Shift+↑/↓)
- Copy responses to clipboard (Ctrl+Y)

### 🔒 **Secure & Local-First**
- API keys stored in `.env`, never in plain text
- Persistent conversation history and agent memory
- Everything works offline except LLM calls
- Human-in-the-loop approval for file modifications

## ⚙️ Configuration

### Framework Setup

**First-time setup** (interactive wizard):
```bash
./zap
# Select your API framework:
# 1. gin    2. echo    3. chi    4. fiber
# 5. fastapi    6. flask    7. django
# 8. express    9. nestjs    10. hono
# ...
```

**With CLI flag** (skip wizard):
```bash
./zap --framework gin
./zap -f fastapi
```

**Update existing config:**
```bash
./zap --framework express
# Updated framework to: express
```

### Configuration Files

ZAP creates a `.zap` folder containing:

**`config.json`** - Main settings:
```json
{
  "ollama_url": "http://localhost:11434",
  "default_model": "llama3",
  "framework": "gin",
  "tool_limits": {
    "default_limit": 50,
    "total_limit": 200,
    "per_tool": {
      "http_request": 25,
      "read_file": 50,
      "search_code": 30
    }
  }
}
```

**Tool Limits** prevent runaway execution:
- `default_limit` - Fallback for tools without specific limits (50)
- `total_limit` - Safety cap on total calls per session (200)
- `per_tool` - Per-tool overrides by name

**`.env`** - API keys (optional):
```env
OPENAI_API_KEY=your_key_here
ANTHROPIC_API_KEY=your_key_here
```

**`requests/`** - Saved API requests:
```yaml
# .zap/requests/get-users.yaml
name: Get Users
method: GET
url: "{{BASE_URL}}/api/users"
headers:
  Authorization: "Bearer {{API_TOKEN}}"
```

**`environments/`** - Environment variables:
```yaml
# .zap/environments/dev.yaml
BASE_URL: http://localhost:3000
API_TOKEN: dev-token-123
```

## 📖 Usage

### Interactive Mode (Default)

```bash
./zap
```

Launch the TUI and interact with natural language:
- `> GET http://localhost:8000/api/users` - Make requests
- `> save this request as get-users` - Save for reuse
- `> switch to prod environment` - Change environments
- `> search for the /users endpoint` - Find code

**Keyboard Shortcuts:**
- `Enter` - Send message
- `Shift+↑/↓` - Navigate input history
- `PgUp/PgDown` - Scroll output
- `Ctrl+L` - Clear screen
- `Ctrl+Y` - Copy last response
- `Ctrl+C` or `Esc` - Quit

### CLI Mode (Automation)

```bash
# Execute saved request with environment
./zap --request get-users --env prod
./zap -r get-users -e dev

# Combine with framework setup
./zap --framework gin --request health-check
```

Perfect for CI/CD pipelines and automation scripts.

## 🎨 Design Philosophy

1. **Context is King** - The agent sees your actual code, not guesses
2. **Human in the Loop** - File modifications require approval (shows diff)
3. **Fail Loudly** - Errors are visible and helpful with debugging hints
4. **Beautiful UX** - Production-quality interface, not a prototype

## 🛠️ Tech Stack

- **Language:** Go 1.25.3
- **CLI Framework:** Cobra + Viper
- **TUI Stack:** Bubble Tea, Lip Gloss, Huh, Bubbles, Glamour
- **LLM Interface:** Raw HTTP Client (no LangChain)
- **Local LLM:** Ollama (with cloud provider fallback)
- **Search:** ripgrep (automatic fallback to native Go implementation)

## 📝 License

MIT

## 🤝 Contributing

Contributions welcome! This is an actively developed project.

## 🙏 Acknowledgments

Built with the amazing [Charm](https://charm.sh/) ecosystem.