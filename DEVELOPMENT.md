# Development Guide

## Current Status

✅ **Phase 1 Foundation Complete**

We have successfully scaffolded the ZAP project with:
- Modern Go project structure
- Cobra CLI framework
- Beautiful TUI with Charm ecosystem
- `.zap` folder auto-initialization
- Working build system

## Running ZAP

### Build
```bash
go build -o zap.exe ./cmd/zap
```

### Run
```bash
./zap
```

On first run, ZAP will create `.zap/` folder with:
- `config.json` - Configuration
- `history.jsonl` - Conversation log
- `memory.json` - Agent memory

## Project Structure

```
zap/
├── cmd/zap/
│   └── main.go              # Entry point, Cobra setup
├── pkg/
│   ├── core/
│   │   ├── init.go          # .zap initialization
│   │   ├── agent.go         # [TODO] ReAct loop
│   │   ├── context.go       # [TODO] Context manager
│   │   └── tools/
│   │       ├── http.go      # [TODO] HTTP client tool
│   │       ├── filesystem.go # [TODO] File reading
│   │       ├── search.go    # [TODO] Code search
│   │       └── env.go       # [TODO] Environment secrets
│   ├── llm/
│   │   └── ollama.go        # [TODO] Ollama client
│   └── tui/
│       ├── app.go           # ✅ Bubble Tea app
│       ├── styles.go        # [TODO] Centralized styles
│       └── components/      # [TODO] Reusable components
├── .zap/                    # Created at runtime
├── .gitignore
├── go.mod
├── go.sum
├── README.md
├── progress.md              # AI agent handoff doc
└── project.md               # Full architecture plan
```

## Next Steps

### 1. Enhance TUI with Huh Forms
Replace basic string input with proper forms from `huh`.

**File:** `pkg/tui/app.go`

```go
import "github.com/charmbracelet/huh"

// Add proper input forms for:
// - API endpoint
// - HTTP method selection
// - Request body
```

### 2. Implement Ollama Client
Create raw HTTP client to connect to Ollama API.

**File:** `pkg/llm/ollama.go`

```go
package llm

type OllamaClient struct {
    baseURL string
    model   string
}

func (c *OllamaClient) Chat(messages []Message) (string, error) {
    // POST to /api/chat
}
```

### 3. Implement HTTP Client Tool
First tool: execute HTTP requests.

**File:** `pkg/core/tools/http.go`

```go
type HTTPTool struct{}

func (t *HTTPTool) Execute(method, url string, body interface{}) (*Response, error) {
    // Perform HTTP request
    // Return structured response
}
```

### 4. Build ReAct Loop
Agent orchestration logic.

**File:** `pkg/core/agent.go`

```go
type Agent struct {
    llm   *llm.OllamaClient
    tools map[string]Tool
}

func (a *Agent) Run(userInput string) error {
    // Reason -> Act -> Observe loop
}
```

## Development Workflow

1. **Make changes** to code
2. **Build**: `go build -o zap.exe ./cmd/zap`
3. **Test**: `./zap`
4. **Iterate**: Fix issues, repeat

## Dependencies

All Charm libraries are installed:
- ✅ `github.com/charmbracelet/bubbletea`
- ✅ `github.com/charmbracelet/lipgloss`
- ✅ `github.com/charmbracelet/huh`
- ✅ `github.com/charmbracelet/bubbles`
- ✅ `github.com/charmbracelet/glamour`
- ✅ `github.com/spf13/cobra`
- ✅ `github.com/spf13/viper`
- ✅ `github.com/joho/godotenv`

## Tips

1. **Color Palette** (defined in `pkg/tui/app.go`):
   - Primary: `#FF6B9D` (pink)
   - Secondary: `#C792EA` (purple)
   - Accent: `#89DDFF` (blue)
   - Background: `#1E1E2E` (dark)

2. **Cobra Commands**: To add subcommands, create new files in `cmd/zap/`

3. **Testing**: Run `go test ./...` to run all tests

## Debugging

- Check `.zap/config.json` for configuration issues
- View `.zap/history.jsonl` for conversation logs
- Use `fmt.Println()` for quick debugging

## Resources

- [Charm Docs](https://charm.sh/)
- [Cobra Docs](https://cobra.dev/)
- [Viper Docs](https://github.com/spf13/viper)
- [Ollama API](https://github.com/ollama/ollama/blob/main/docs/api.md)
