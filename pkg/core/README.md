# pkg/core

The `core` package contains the agent logic, ReAct loop implementation, and tool management system. This is the brain of ZAP.

## Package Overview

```
pkg/core/
├── types.go       # Core interfaces (Tool, AgentEvent, FileConfirmation)
├── agent.go       # Agent struct, tool registration, call counting
├── react.go       # ReAct loop: ProcessMessage, ProcessMessageWithEvents
├── prompt.go      # System prompt construction (20 sections)
├── init.go        # Configuration loading, setup wizard, framework selection
├── memory.go      # Persistent memory store for facts across sessions
├── analysis.go    # Error context extraction, stack trace parsing
├── manifest.go    # Tool manifest metadata
├── secrets.go     # Secrets handling (API keys, credentials)
├── react_test.go  # Unit tests for ReAct loop
└── tools/         # Tool implementations (see tools/README.md)
```

## Core Interfaces

### Tool Interface

Every tool must implement this interface:

```go
type Tool interface {
    Name() string                           // Unique identifier (e.g., "http_request")
    Description() string                    // Human-readable description for LLM
    Parameters() string                     // JSON Schema of parameters
    Execute(args string) (string, error)    // Main execution, args is JSON string
}
```

### ConfirmableTool Interface

Tools that modify files should also implement:

```go
type ConfirmableTool interface {
    Tool
    SetConfirmationManager(cm *tools.ConfirmationManager)
}
```

### AgentEvent

Events emitted during agent execution:

```go
type AgentEvent struct {
    Type             string                // Event type (see below)
    Content          string                // Main payload
    ToolArgs         string                // Tool arguments (for tool_call events)
    ToolUsage        *ToolUsageEvent       // Stats (for tool_usage events)
    FileConfirmation *FileConfirmation     // File write details (for confirmation_required)
}
```

**Event Types:**

| Type | Description |
|------|-------------|
| `thinking` | Agent is reasoning (not shown to user) |
| `tool_call` | Agent is calling a tool |
| `observation` | Tool returned a result |
| `answer` | Final answer from agent |
| `error` | Error occurred |
| `streaming` | Partial response (real-time display) |
| `tool_usage` | Tool usage statistics update |
| `confirmation_required` | File write needs approval |

## Agent Structure

```go
type Agent struct {
    llmClient    llm.LLMClient           // LLM provider (Ollama, Gemini)
    tools        map[string]Tool         // Registered tools by name
    history      []llm.Message           // Conversation history
    toolCounts   map[string]int          // Per-tool call counters
    toolLimits   map[string]int          // Per-tool max limits
    totalCalls   int                     // Total tool calls this session
    totalLimit   int                     // Safety cap on total calls
    defaultLimit int                     // Default limit for unlisted tools
    framework    string                  // API framework (gin, fastapi, etc.)
    memoryStore  *MemoryStore            // Persistent memory
    stopRequested bool                   // User requested stop (Esc key)
}
```

## Key Functions

### Creating an Agent

```go
agent := core.NewAgent(llmClient)
agent.SetFramework("gin")
agent.SetToolLimit("http_request", 25)
agent.RegisterTool(tools.NewHTTPTool())
```

### Processing Messages

**Blocking (simple):**

```go
response, err := agent.ProcessMessage("GET http://localhost:8000/users")
```

**With Events (for TUI):**

```go
response, err := agent.ProcessMessageWithEvents(input, func(event core.AgentEvent) {
    switch event.Type {
    case "streaming":
        // Update UI with partial response
    case "tool_call":
        // Show tool being executed
    case "answer":
        // Display final answer
    }
})
```

### Tool Registration

```go
agent.RegisterTool(tools.NewHTTPTool())
agent.RegisterTool(tools.NewReadFileTool())
agent.RegisterTool(tools.NewSearchTool())
```

## ReAct Loop

The ReAct (Reason + Act) loop in `react.go`:

```
1. Build system prompt (20 sections)
2. Send conversation to LLM
3. Parse response for tool calls
4. If no tool call → Return final answer
5. Check tool limits
6. Execute tool
7. Add observation to history
8. GOTO 2
```

### Tool Call Parsing

The agent looks for tool calls in this format:

```
<tool_call>
{"name": "http_request", "arguments": {"method": "GET", "url": "..."}}
</tool_call>
```

### Limit Enforcement

Before executing a tool:

```go
if agent.toolCounts[toolName] >= agent.toolLimits[toolName] {
    return "Tool limit reached for " + toolName
}
if agent.totalCalls >= agent.totalLimit {
    return "Total tool call limit reached"
}
```

## System Prompt

The system prompt in `prompt.go` has 20 sections:

1. Identity & response format
2. Scope (what it does/doesn't)
3. Guardrails & safety
4. Behavioral rules
5. Autonomous workflow
6. .zap folder sync
7. Secrets handling
8. Tool usage rules
9. Memory operations
10. Tools catalog
11. Framework hints (language/framework-specific)
12. Natural language parsing
13. Error diagnosis
14. Common errors
15. Persistence (save/load)
16. Testing patterns
17. Request chaining
18. Authentication
19. Test suites
20. Output format

### Framework-Specific Hints

Based on `agent.framework`, the prompt includes relevant hints:

```go
// For "gin"
"Search for gin.Context, c.JSON, c.Bind..."

// For "fastapi"
"Search for @app.get, @app.post, Depends..."
```

## Memory System

The `MemoryStore` in `memory.go` persists facts across sessions:

```go
memoryStore := core.NewMemoryStore(".zap/memory.json")
memoryStore.AddFact("The users endpoint is at /api/v2/users")
facts := memoryStore.GetFacts()
```

## Error Analysis

The `analysis.go` file provides error parsing utilities:

```go
// Parse stack trace from error response
files := core.ParseStackTrace(errorBody)
// Returns: []FileLocation{{File: "api.py", Line: 42}, ...}

// Extract error context from JSON response
context := core.ExtractErrorContext(jsonResponse)
// Returns: ErrorContext{Message: "...", Type: "...", Fields: [...]}
```

## .zap Folder Structure

The `.zap` directory serves as the brain, memory, and output center for the agent.

```
.zap/
├── baselines/          # "The Standard of Truth"
├── snapshots/          # "The Current Reality"
├── requests/           # "Saved Actions"
├── runs/               # "The History Book"
├── exports/            # "The Filing Cabinet"
├── logs/               # "The Diary"
├── state/              # "The Brain"
└── config/             # "The Settings"
```

### Folder Breakdown

-   **`baselines/`**: Stores the "definition of normal." These are captured snapshots (functional, performance, schema) used by the regression watchdog to detect unintended changes.
-   **`snapshots/`**: Contains the **API Knowledge Graph**. Generated by the Spec Ingester, this represents ZAP's current understanding of the API structure.
-   **`requests/`**: A library of reusable, user-saved HTTP requests (YAML). Think of this as your project-specific Postman collection.
-   **`runs/`**: Immutable history of every test execution. Each run gets a timestamped folder containing results, logs, and artifacts.
-   **`exports/`**: Polished, human-readable reports (Markdown/PDF) generated for external consumption (e.g., "Weekly Security Scan").
-   **`logs/`**: Internal operational logs for ZAP itself. Useful for debugging why the *tool* failed (not why the *test* failed).
-   **`state/`**: The agent's long-term memory and context. Stores facts learned across sessions.
-   **`config/`**: Configuration files (`config.json`, `.env`) that control ZAP's behavior and environment settings.

## Configuration Loading

The `init.go` file handles:

1. **Config file loading** from `.zap/config.json`
2. **Setup wizard** for first-time configuration
3. **Framework selection** (interactive or via flag)
4. **Environment variable loading** from `.env`

```go
config, err := core.LoadConfig()
if config == nil {
    config = core.RunSetupWizard()
}
```

## Adding New Functionality

### Adding a New Event Type

1. Add constant in `types.go`
2. Emit from `react.go` at appropriate point
3. Handle in `pkg/tui/update.go`

### Adding New System Prompt Section

1. Edit `prompt.go`
2. Add section to `buildSystemPrompt()` function
3. Test with various LLM providers

### Adding New Configuration Option

1. Add field to config struct in `init.go`
2. Update setup wizard if needed
3. Update `.zap/config.json` schema

## Testing

Run tests:

```bash
go test ./pkg/core/...
```

The `react_test.go` file contains unit tests for:

- Tool call parsing
- Limit enforcement
- Event emission
- History management
