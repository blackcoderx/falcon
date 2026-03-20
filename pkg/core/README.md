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
| `confirmation_required` | File write awaiting user approval |

---

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
}

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
- Event emission order
- History management
