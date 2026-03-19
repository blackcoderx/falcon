# pkg/core

The `core` package contains Falcon's agent logic, the ReAct loop, and the tool management system. This is the brain of Falcon.

## Package Overview

```
pkg/core/
├── types.go               # Core interfaces (Tool, AgentEvent, ConfirmableTool)
├── agent.go               # Agent struct, tool registration, call limit enforcement
├── react.go               # ReAct loop: ProcessMessage, ProcessMessageWithEvents
├── init.go                # .falcon folder setup, setup wizard, project config
├── globalconfig.go        # ~/.falcon global config management (providers, credentials)
├── memory.go              # Persistent MemoryStore across sessions
├── analysis.go            # Stack trace parsing, error context extraction
├── prompt_integration.go  # Helpers for injecting tool descriptions into system prompt
├── react_test.go          # Unit tests for the ReAct loop
├── prompt/                # Modular system prompt builder (workflow + tools sections)
└── tools/                 # 28+ tool implementations organized by testing type
```

## Core Interfaces

### Tool

Every Falcon tool must implement this interface:

```go
type Tool interface {
    Name() string                        // Unique identifier, e.g. "http"
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
| `thinking` | Agent is reasoning (not displayed directly) |
| `tool_call` | Agent is invoking a tool |
| `observation` | Tool returned a result |
| `answer` | Final answer from the agent |
| `error` | An error occurred |
| `streaming` | Partial LLM response chunk (real-time display) |
| `tool_usage` | Per-tool call counter update |
| `confirmation_required` | File write awaiting user approval |

---

## Agent Structure

```go
type Agent struct {
    llmClient    llm.LLMClient
    clientMu     sync.RWMutex  // Protects llmClient
    tools        map[string]Tool
    toolsMu      sync.RWMutex  // Protects tools map
    history      []llm.Message
    historyMu    sync.RWMutex  // Protects history slice
    lastResponse interface{}   // Last tool response for chaining

    // Per-tool call limiting
    toolLimits   map[string]int // max calls per tool per session
    toolCounts   map[string]int // current session call counts
    countersMu   sync.Mutex     // Protects toolCounts and totalCalls
    defaultLimit int            // fallback limit for unlisted tools (default 50)
    totalLimit   int            // safety cap on total calls (default 200)
    totalCalls   int            // current total calls in session

    maxHistory  int         // max messages to keep in history (default 100)
    framework   string      // gin, fastapi, express, etc.
    memoryStore *MemoryStore
}
```

**Constants:**

```go
const (
    DefaultToolCallLimit = 50   // Default max calls per tool per session
    DefaultTotalLimit    = 200  // Safety cap on total tool calls per session
    DefaultMaxHistory    = 100  // Default max messages to keep in history
)
```

### Creating an Agent

```go
agent := core.NewAgent(llmClient)
agent.SetFramework("gin")
agent.SetToolLimit("http", 25)
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

---

## ReAct Loop

The ReAct (Reason + Act) loop in `react.go`:

```
1. Add user message to history, reset tool counters
2. Build system prompt with tool descriptions
3. Call LLM via Chat/ChatStream (retry up to 3× with exponential backoff: 2s, 4s, 8s)
4. Parse response for tool call or Final Answer
5. If Final Answer → emit "answer" event and return
6. Check per-tool and total call limits
7. Execute tool → emit "tool_call" + "observation" events
8. Append observation to conversation history
9. GOTO 2
```

### LLM Response Format

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

---

## System Prompt

The prompt in `pkg/core/prompt/` is built from modular sections in `workflow.go` and `tools.go`:

**workflow.go sections:**
- **Mandatory Session Start** — `memory(recall)` + `session_log(start)` at every conversation start
- **Mandatory Session End** — `session_log(end, summary)` before every final answer
- **Testing Type Decision** — decision table for 8 API testing types (Unit, Integration, Smoke, Functional, Contract, Performance, Security, E2E)
- **The Five Phases** — Orient, Hypothesize, Act, Interpret, Persist
- **Tool Disambiguation** — clarifies merged tools (auth, request, environment) and write-only tools (falcon_write, session_log)
- **.falcon Naming Convention** — flat structure with type-prefixed filenames
- **Persistence Rules** — what to save and where

**tools.go sections:**
- **Available Tools** — all 28+ tools organized by domain
- **Compact Tool Reference** — quick lookup table with tool intent, params, and domains

---

## Configuration

Falcon splits configuration across two files:

### Global Config (`~/.falcon/config.yaml`)

Managed by `globalconfig.go`. Stores LLM provider credentials, shared across all projects.

```go
type GlobalConfig struct {
    DefaultProvider string                  `yaml:"default_provider"`
    Theme           string                  `yaml:"theme"`
    WebUI           WebUIConfig             `yaml:"web_ui"`
    Providers       map[string]ProviderEntry `yaml:"providers"`
}

type ProviderEntry struct {
    Model  string            `yaml:"model"`
    Config map[string]string `yaml:"config"`
}
```

**Example:**
```yaml
default_provider: gemini
theme: dark
web_ui:
  enabled: true
  port: 0
providers:
  ollama:
    model: llama3
    config:
      mode: local
      url: http://localhost:11434
  gemini:
    model: gemini-2.5-flash-lite
    config:
      api_key: your-key-here
  openrouter:
    model: google/gemini-2.5-flash-lite
    config:
      api_key: sk-or-...
```

**Key functions:**
- `LoadGlobalConfig()` — reads `~/.falcon/config.yaml`, auto-migrates legacy format
- `SaveGlobalConfig(cfg)` — writes with `0600` permissions
- `GetActiveProviderEntry(cfg)` — returns the active provider ID, model, and config values
- `SetProviderEntry(cfg, id, model, values)` — upserts one provider without touching others
- `RunGlobalConfigWizard()` — interactive Huh wizard for add/update/remove/set-default

### Project Config (`.falcon/config.yaml`)

Managed by `init.go`. Per-project overrides.

```go
type Config struct {
    Provider       string            `yaml:"provider"`
    ProviderConfig map[string]string `yaml:"provider_config,omitempty"`
    DefaultModel   string            `yaml:"default_model"`
    Theme          string            `yaml:"theme"`
    Framework      string            `yaml:"framework"`
    WebUI          WebUIConfig       `yaml:"web_ui"`
}
```

**Example:**
```yaml
framework: gin
web_ui:
  enabled: true
  port: 0
```

Config migration (`migrateToGlobalConfig`) automatically moves legacy per-provider credentials from `.falcon/config.yaml` into `~/.falcon/config.yaml` on first run.

---

## Memory System

`MemoryStore` in `memory.go` persists facts across sessions in `~/.falcon/memory.json`:

```go
type MemoryEntry struct {
    Key       string `json:"key"`
    Value     string `json:"value"`
    Category  string `json:"category"` // "preference", "endpoint", "error", "project", "general"
    Timestamp string `json:"timestamp"`
    Source    string `json:"source"`
}
```

```go
store := core.NewMemoryStore("~/.falcon/memory.json")
store.Save("auth-endpoint", "POST /api/auth/login", "endpoint")
entries := store.Recall("endpoint")
```

---

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

---

## .falcon Folder Structure

Created on first run by `InitializeFalconFolder()`. Uses a **flat structure** — no subdirectories in `reports/` or `flows/`. Filenames carry context via type prefix.

```
~/.falcon/                   # Global (credentials + memory)
├── config.yaml
└── memory.json

.falcon/                     # Per-project
├── config.yaml
├── falcon.md                # API knowledge base
├── spec.yaml                # Ingested API spec (YAML)
├── manifest.json            # Parsed endpoint graph
├── variables.json           # Global variables
├── sessions/
├── environments/
├── requests/
├── baselines/
├── flows/
└── reports/
```

**File naming conventions:**
- Reports: `<type>_report_<api-name>_<timestamp>.md`
- Flows: `<type>_<description>.yaml`

---

## Adding New Functionality

### New Event Type
1. Add a constant in `types.go`
2. Emit from `react.go` at the appropriate point
3. Handle in `pkg/tui/update.go`

### New System Prompt Section
1. Add a new file in `pkg/core/prompt/`
2. Register it in `prompt/builder.go`
3. Test with different LLM providers

### New LLM Provider
See `pkg/llm/README.md` — self-contained in `pkg/llm/`, no changes needed here.

### New Config Option
1. Add the field to `Config` in `init.go` (project-level) or `GlobalConfig` in `globalconfig.go` (global)
2. If user-configurable at setup time, add it to the wizard

---

## Testing

```bash
go test ./pkg/core/...
```

`react_test.go` covers:
- Tool call parsing from LLM output
- Limit enforcement (per-tool and total)
- Event emission order
- History management
