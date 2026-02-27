# pkg/core

The `core` package contains Falcon's agent logic, the ReAct loop, and the tool management system. This is the brain of Falcon.

## Package Overview

```
pkg/core/
├── types.go               # Core interfaces (Tool, AgentEvent, ConfirmableTool)
├── agent.go               # Agent struct, tool registration, call limit enforcement
├── react.go               # ReAct loop: ProcessMessage, ProcessMessageWithEvents
├── init.go                # .falcon folder setup, setup wizard, config migration
├── memory.go              # Persistent MemoryStore across sessions
├── analysis.go            # Stack trace parsing, error context extraction
├── manifest.go            # Tool manifest metadata
├── secrets.go             # Secrets detection and handling
├── prompt_integration.go  # Helpers for injecting tool descriptions into system prompt
├── react_test.go          # Unit tests for the ReAct loop
├── prompt/                # Modular system prompt builder (20 sections)
└── tools/                 # All 40+ tool implementations (see tools/README.md)
```

## Core Interfaces

### Tool

Every Falcon tool must implement this interface:

```go
type Tool interface {
    Name() string                        // Unique identifier, e.g. "http_request"
    Description() string                 // Human-readable description for the LLM
    Parameters() string                  // JSON Schema of accepted parameters
    Execute(args string) (string, error) // Main execution; args is a JSON string
}
```

### ConfirmableTool

Tools that write files must also implement this:

```go
type ConfirmableTool interface {
    Tool
    SetEventCallback(callback EventCallback)
}
```

When the agent calls a `ConfirmableTool`, it emits a `confirmation_required` event to the TUI before writing anything. The user must approve (Y) or reject (N) the change.

### AgentEvent

Events emitted by the ReAct loop to drive the TUI in real time:

```go
type AgentEvent struct {
    Type             string            // Event type (see below)
    Content          string            // Main payload
    ToolArgs         string            // Tool arguments (tool_call events)
    ToolUsage        *ToolUsageEvent   // Stats (tool_usage events)
    FileConfirmation *FileConfirmation // File write details (confirmation_required events)
}
```

**Event Types:**

| Type | Description |
|------|-------------|
| `thinking` | Agent is reasoning (not displayed) |
| `tool_call` | Agent is invoking a tool |
| `observation` | Tool returned a result |
| `answer` | Final answer from the agent |
| `error` | An error occurred |
| `streaming` | Partial LLM response chunk (real-time display) |
| `tool_usage` | Per-tool call counter update |
| `confirmation_required` | File write awaiting user approval |

## Agent Structure

```go
type Agent struct {
    llmClient    llm.LLMClient     // LLM provider (Ollama, Gemini)
    tools        map[string]Tool   // Registered tools by name
    history      []llm.Message     // Conversation history
    toolCounts   map[string]int    // Per-tool call counters
    toolLimits   map[string]int    // Per-tool max limits
    totalCalls   int               // Total tool calls this session
    totalLimit   int               // Safety cap on total calls (default 200)
    defaultLimit int               // Fallback limit for unlisted tools (default 50)
    framework    string            // API framework (gin, fastapi, express, etc.)
    memoryStore  *MemoryStore      // Persistent memory
}
```

### Creating an Agent

```go
agent := core.NewAgent(llmClient)
agent.SetFramework("gin")
agent.SetToolLimit("http_request", 25)
agent.RegisterTool(tools.NewHTTPTool())
```

### Processing Messages

**Blocking (simple use):**

```go
response, err := agent.ProcessMessage("GET http://localhost:8000/users")
```

**With Events (for TUI):**

```go
err := agent.ProcessMessageWithEvents(ctx, input, func(event core.AgentEvent) {
    switch event.Type {
    case "streaming":
        // Update UI with partial response
    case "tool_call":
        // Show which tool is being executed
    case "answer":
        // Display final answer
    }
})
```

## ReAct Loop

The ReAct (Reason + Act) loop in `react.go`:

```
1. Build system prompt (20 sections)
2. Send conversation to LLM via ChatStream
3. Parse response for tool calls
4. If no tool call → emit "answer" event and return
5. Check per-tool and total call limits
6. Execute tool → emit "tool_call" + "observation" events
7. Append observation to conversation history
8. GOTO 2
```

### LLM Response Format

The LLM produces structured output that the parser understands:

```
Thought: <reasoning>
ACTION: tool_name({"arg": "value"})
```

or for a final response:

```
Final Answer: <response>
```

The parser (`parseResponse`) handles common LLM formatting variations — missing `ACTION:` prefix, raw `tool_name(...)` calls, and case differences.

### Limit Enforcement

Before executing any tool:

```go
if agent.toolCounts[toolName] >= agent.toolLimits[toolName] {
    return "Tool limit reached for " + toolName
}
if agent.totalCalls >= agent.totalLimit {
    return "Total tool call limit reached"
}
```

## System Prompt

The prompt in `pkg/core/prompt/` is built from 20 modular sections:

1. Identity & response format
2. Scope (what Falcon does and does not do)
3. Guardrails & safety
4. Behavioral rules
5. Autonomous workflow
6. `.falcon` folder awareness
7. Secrets handling
8. Tool usage rules
9. Memory operations
10. Tools catalog
11. Framework-specific hints (gin, fastapi, express, etc.)
12. Natural language parsing
13. Error diagnosis
14. Common errors
15. Persistence (save/load patterns)
16. Testing patterns
17. Request chaining
18. Authentication flows
19. Test suites
20. Output format

## Memory System

`MemoryStore` in `memory.go` persists facts across sessions in `.falcon/memory.json`:

```go
store := core.NewMemoryStore(".falcon/memory.json")
store.AddFact("The users endpoint is at /api/v2/users")
facts := store.GetFacts()
```

## Error Analysis

`analysis.go` provides utilities for parsing error responses:

```go
// Parse stack trace from an error response body
files := core.ParseStackTrace(errorBody)
// Returns: []FileLocation{{File: "api.py", Line: 42}, ...}

// Extract structured error context from a JSON response
ctx := core.ExtractErrorContext(jsonResponse)
// Returns: ErrorContext{Message: "...", Type: "...", Fields: [...]}
```

## .falcon Folder

The `.falcon` directory is created on first run by `InitializeZapFolder()`:

```
.falcon/
├── config.yaml         # LLM provider, model, framework, tool limits
├── memory.json         # Persistent agent memory (versioned)
├── manifest.json       # Workspace manifest (counts for requests, environments, etc.)
├── falcon.md           # Falcon knowledge base template
├── requests/           # Saved API requests (YAML)
├── environments/       # Environment variable files (dev.yaml, prod.yaml, staging.yaml)
├── baselines/          # Reference snapshots for regression testing
└── flows/              # Saved multi-step API flows
└── reports/
```

`config.yaml` is YAML (not JSON). Example:

```yaml
provider: ollama
ollama:
  mode: local
  url: http://localhost:11434
  api_key: ""
default_model: llama3
framework: gin
theme: dark
tool_limits:
  default_limit: 50
  total_limit: 200
  per_tool:
    http_request: 25
    variable: 100
web_ui:
  enabled: true
  port: 0
```

Config migration (`migrateLegacyConfig`) automatically promotes legacy top-level `ollama_url` / `ollama_api_key` fields into the `ollama` sub-object.

## Adding New Functionality

### New Event Type

1. Add constant in `types.go`
2. Emit from `react.go` at the appropriate point
3. Handle in `pkg/tui/update.go`

### New System Prompt Section

1. Add a new file in `pkg/core/prompt/`
2. Register it in `prompt/builder.go`
3. Test with different LLM providers

### New Config Option

1. Add field to the config struct in `init.go`
2. Update the setup wizard if it should be user-configurable
3. Update `config.yaml` documentation

## Testing

```bash
go test ./pkg/core/...
```

`react_test.go` covers:

- Tool call parsing from LLM output
- Limit enforcement (per-tool and total)
- Event emission order
- History management
