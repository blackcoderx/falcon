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
├── prompt/                # Modular system prompt builder (workflow + tools sections)
└── tools/                 # 28-tool implementations organized by testing type (see tools/README.md)
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

The prompt in `pkg/core/prompt/` is built from modular sections defined in `workflow.go` and `tools.go`:

**workflow.go:**
- **Mandatory Session Start**: memory(recall) + session_log(start) at every conversation beginning
- **Mandatory Session End**: session_log(end, summary) before final answer
- **Which Testing Type?**: Decision table for 8 API testing types (Unit, Integration, Smoke, Functional, Contract, Performance, Security, E2E)
- **The Five Phases**: Orient, Hypothesize, Act, Interpret, Persist
- **Tool Disambiguation**: Clarifies merged tools (auth, request, environment) and new tools (falcon_write, falcon_read, session_log)
- **.falcon File Naming Convention**: Flat structure with type-prefixed filenames
- **Tool Selection**: Cost hierarchy and when to use which tool
- **Confidence Calibration**: When to stop investigating vs. admit uncertainty
- **Reports**: Validation rules for all report types
- **Persistence Rules**: What to save and where

**tools.go:**
- **Available Tools**: All 28 tools organized by domain (Core, Persistence, Spec, Unit, Integration, Smoke, Functional, Contract, Performance, Security, Debugging, Orchestration)
- **Compact Tool Reference**: Quick lookup table with tool intent, params, and domains

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

## .falcon Folder Structure

The `.falcon` directory is created on first run by `InitializeDefaultFolder()`. It uses a **flat structure** — no subdirectories in `reports/` or `flows/`. Filenames carry context via type prefix:

```
.falcon/
├── config.yaml              # LLM provider, model, framework, tool limits
├── manifest.json            # Workspace manifest (counts)
├── memory.json              # Persistent agent memory
├── falcon.md                # API knowledge base (validated on write)
├── spec.yaml                # Ingested API spec (YAML, single file)
├── variables.json           # Global variables
├── sessions/
│   ├── session_2026-02-27T14:32:01Z.json
│   └── session_2026-02-27T15:45:22Z.json
├── environments/
│   ├── dev.yaml
│   ├── prod.yaml
│   └── staging.yaml
├── requests/
│   ├── create_user.yaml
│   └── get_users.yaml
├── baselines/
│   ├── baseline_users_api.json
│   └── baseline_products_api.json
├── flows/
│   ├── unit_get_users.yaml
│   ├── integration_login_create_delete.yaml
│   ├── smoke_all_endpoints.yaml
│   └── security_auth_bypass.yaml
└── reports/
    ├── performance_report_dummyjson_products_20260227.md
    ├── security_report_products_api_20260227.md
    ├── functional_report_users_api_20260227.md
    └── unit_report_get_users_20260227.md
```

### File Naming Conventions

**Reports**: `<type>_report_<api-name>_<timestamp>.md`
- Examples: `performance_report_dummyjson_products.md`, `security_report_auth_api.md`

**Flows**: `<type>_<description>.yaml`
- Examples: `unit_get_users.yaml`, `integration_login_flow.yaml`, `security_auth_bypass.yaml`

**Spec**: `spec.yaml` (single file, YAML format for human readability)

### Validators

Two new validators ensure report and knowledge base quality:

1. **ValidateReportContent** (`shared/report_validator.go`):
   - Checks file exists and is > 64 bytes
   - Verifies Markdown heading present (`# ` or `## `)
   - Confirms result indicators (table `|`, code block ` ``` `, or keywords like `PASS`, `FAIL`, `✓`, `✗`)
   - Rejects unresolved placeholders (`{{`, `TODO`, `[placeholder]`)
   - Called after every report write; returns error if validation fails so agent can retry

2. **ValidateFalconMD** (`shared/report_validator.go`):
   - Checks file exists and is > 200 bytes
   - Verifies required sections: `# Base URLs`, `# Known Endpoints`
   - Confirms each section has non-empty content
   - Called after `memory(action=update_knowledge)`; returns validation warning if it fails

### config.yaml Format

Example:

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
    auth: 50
    request: 50
    memory: 100
    run_performance: 10
web_ui:
  enabled: true
  port: 0
```

Config migration (`migrateLegacyConfig`) automatically promotes legacy top-level `ollama_url` / `ollama_api_key` fields into the `ollama` sub-object.

### Session Logging

Each session writes a JSON record to `.falcon/sessions/session_<timestamp>.json`:

```json
{
  "session_id": "2026-02-27T14:32:01Z",
  "start_time": "2026-02-27T14:32:01Z",
  "end_time": "2026-02-27T14:45:22Z",
  "summary": "Tested POST /users — found 422 on missing email field, fixed validation in user_handler.go"
}
```

Sessions are an audit trail: user can call `session_log(action=list)` to see recent tests or `session_log(action=read, session="...")` to review a specific session.

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
