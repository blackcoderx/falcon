# ZAP Development Guide

## Project Status: SPRINT 2 COMPLETE - ERROR-CODE PIPELINE

The killer demo is ready: API errors trigger codebase search and diagnosis with file:line references. Ready for Sprint 3 (Persistence & Storage).

### Current Structure
```
zap/
├── cmd/zap/main.go           # Entry point with Cobra/Viper/Env
├── pkg/
│   ├── core/
│   │   ├── init.go           # .zap folder initialization
│   │   ├── agent.go          # ReAct Agent + Event System + Error Diagnosis
│   │   ├── analysis.go       # Stack trace parsing, error extraction
│   │   └── tools/
│   │       ├── http.go       # HTTP Tool + status code helpers
│   │       ├── file.go       # read_file, list_files tools
│   │       └── search.go     # search_code tool (ripgrep/native)
│   ├── llm/
│   │   └── ollama.go         # Ollama Cloud client (Bearer auth)
│   └── tui/
│       ├── app.go            # Minimal TUI (viewport, textinput, spinner)
│       └── styles.go         # Minimal styling (7 colors, log prefixes)
```

## Working with the Agent

### Tool Interface
Every new capability must implement the `Tool` interface in `pkg/core/agent.go`:
```go
type Tool interface {
    Name() string
    Description() string
    Parameters() string
    Execute(args string) (string, error)
}
```

### Agent Event System (New)
The agent now supports real-time event emission:
```go
type AgentEvent struct {
    Type    string // "thinking", "tool_call", "observation", "answer", "error"
    Content string
}

type EventCallback func(AgentEvent)

// Use this for real-time UI updates
agent.ProcessMessageWithEvents(input, callback)

// Or use the original blocking version
agent.ProcessMessage(input)
```

### Logging
- Use `fmt.Fprintf(os.Stderr, ...)` for debug info
- stdout belongs to the TUI - never print there directly

## Getting Started

### Requirements
- Go 1.23+
- Ollama Cloud API Key (for `ollama.com`)

### Configuration
Create a `.env` file in the root:
```env
OLLAMA_API_KEY=your_key_here
```

Ensure `.zap/config.json` uses a cloud model:
```json
{
  "ollama_url": "https://ollama.com",
  "default_model": "gpt-oss:20b-cloud"
}
```

### Build & Run
```bash
go build -o zap.exe ./cmd/zap
./zap.exe
```

## TUI Architecture

### Components Used
- `bubbles/viewport` - Scrollable log area
- `bubbles/textinput` - Single-line input with `> ` prompt
- `bubbles/spinner` - Loading indicator
- `glamour` - Markdown rendering for responses
- `lipgloss` - Minimal styling

### Styling
Minimal 7-color palette:
- `#6c6c6c` - Dim (thinking, observations, help)
- `#e0e0e0` - Text (user input, responses)
- `#7aa2f7` - Accent (prompt, title, shortcuts)
- `#f7768e` - Error
- `#9ece6a` - Tool calls
- `#545454` - Muted (separators)
- `#73daca` - Success (future use)

Log prefixes:
- `> ` - User input
- `  thinking ` - Agent reasoning
- `  tool ` - Tool being called
- `  result ` - Tool observation
- `  error ` - Errors
- `───` - Conversation separator

### Keyboard Shortcuts
- `enter` - Send message
- `↑` / `↓` - Navigate input history
- `pgup` / `pgdown` - Scroll viewport
- `ctrl+l` - Clear screen
- `ctrl+u` - Clear input
- `ctrl+c` / `esc` - Quit

### Message Flow
```
User Input
    ↓
TUI captures Enter key
    ↓
runAgentAsync() starts goroutine
    ↓
Agent.ProcessMessageWithEvents() runs
    ↓
Callback sends AgentEvent via program.Send()
    ↓
TUI Update() receives agentEventMsg
    ↓
Appends to logs[], updates viewport
    ↓
agentDoneMsg signals completion
```

## What's Still Needed

### Sprint 1 - Codebase Tools - COMPLETE
- ✓ `read_file`, `list_files`, `search_code` tools
- ✓ Codebase-aware system prompt

### Sprint 2 - Error-Code Pipeline - COMPLETE
- ✓ Enhanced error diagnosis prompt with workflow
- ✓ HTTP status code meanings and hints
- ✓ Stack trace parsing (Python, Go, JS)
- ✓ Error context extraction from JSON
- ✓ Natural language → HTTP request conversion

### Sprint 3 Goals (Persistence & Storage)
1. YAML request schema definition
2. Save/load requests to YAML files
3. Request history in session
4. Environment variable substitution

## Running on Other Projects

Run ZAP from the target project directory:
```bash
cd /path/to/your/project
/c/Users/user/zap/zap.exe
```

The tools use the current working directory as the project root for security bounds.

## Debugging

- If agent returns empty responses, check stderr logs
- Model name must match Ollama Cloud exactly (use `:cloud` suffix)
- Use `ctrl+c` or `esc` to quit cleanly
- Mouse wheel scrolls the viewport
