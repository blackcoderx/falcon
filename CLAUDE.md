# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ZAP is an AI-powered API debugging assistant that runs in the terminal. It combines API testing with codebase awareness - when an API returns an error, ZAP can search your code to find the cause and suggest fixes. Uses local LLMs (Ollama) or cloud providers.

## Build & Run Commands

```bash
# Build the application
go build -o zap.exe ./cmd/zap

# Run the application
./zap.exe

# Run with custom config
./zap --config path/to/config.json
```

## Architecture

### Package Structure

- **cmd/zap/** - Application entry point using Cobra CLI framework
- **pkg/core/** - Agent logic, event system, and initialization
- **pkg/core/tools/** - Agent tools (HTTP, file, search)
- **pkg/llm/** - LLM client implementations (Ollama)
- **pkg/tui/** - Minimal terminal UI using Bubble Tea

### Core Components

**Agent (pkg/core/agent.go)**: Implements ReAct (Reason+Act) loop with event system:
- `ProcessMessage(input)` - Blocking, returns final answer
- `ProcessMessageWithEvents(input, callback)` - Emits events for real-time UI updates
- Events: `thinking`, `tool_call`, `observation`, `answer`, `error`
- Max 5 iterations to prevent infinite loops

**Tool Interface**:
```go
type Tool interface {
    Name() string
    Description() string
    Parameters() string
    Execute(args string) (string, error)
}
```

**TUI (pkg/tui/app.go)**: Minimal Claude Code-style interface:
- `bubbles/viewport` - Scrollable log area (pgup/pgdown, mouse wheel)
- `bubbles/textinput` - Single-line input with `> ` prompt
- `bubbles/spinner` - Loading indicator
- `glamour` - Markdown rendering for responses
- Streaming display (text appears as it arrives)
- Status line showing current state (thinking/streaming/executing tool)
- Input history navigation (↑/↓ arrows)

**Styling (pkg/tui/styles.go)**: Minimal 7-color palette with log prefixes:
- `> ` user input
- `  thinking ` agent reasoning
- `  tool ` tool calls
- `  result ` observations
- `  error ` errors
- `───` conversation separator

**Keyboard Shortcuts**:
- `enter` - Send message
- `↑` / `↓` - Navigate input history
- `pgup` / `pgdown` - Scroll viewport
- `ctrl+l` - Clear screen
- `ctrl+u` - Clear input line
- `ctrl+c` / `esc` - Quit

### Configuration

On first run, creates `.zap/` folder containing:
- `config.json` - Ollama URL, model settings
- `history.jsonl` - Conversation log
- `memory.json` - Agent memory

Environment: `OLLAMA_API_KEY` loaded from `.env` file.

### Adding New Tools

1. Create a new file in `pkg/core/tools/`
2. Implement the `core.Tool` interface
3. Register in `pkg/tui/app.go` via `agent.RegisterTool()`

### Message Flow

```
User Input → TUI captures Enter
           → runAgentAsync() starts goroutine
           → Agent.ProcessMessageWithEvents() runs
           → Callback sends AgentEvent via program.Send()
           → TUI Update() receives agentEventMsg
           → Appends to logs[], updates viewport
           → agentDoneMsg signals completion
```

## Key Files

| File | Purpose |
|------|---------|
| `pkg/core/agent.go` | ReAct loop + event system + codebase-aware system prompt |
| `pkg/tui/app.go` | Minimal TUI with viewport, textinput, spinner, status line, history |
| `pkg/tui/styles.go` | 7-color palette, log prefixes, keyboard shortcut styles |
| `pkg/llm/ollama.go` | Ollama Cloud client with Bearer auth + streaming |
| `pkg/core/tools/http.go` | HTTP request tool |
| `pkg/core/tools/file.go` | `read_file` and `list_files` tools |
| `pkg/core/tools/search.go` | `search_code` tool (ripgrep with native fallback) |

## Available Tools

| Tool | Description |
|------|-------------|
| `http_request` | Make HTTP requests (GET/POST/PUT/DELETE) |
| `read_file` | Read file contents (with 100KB limit) |
| `list_files` | List files with glob patterns (`**/*.go`) |
| `search_code` | Search for patterns in codebase (uses ripgrep if available) |
