# CLAUDE.md

This file provides guidance to AI coding assistants (Claude, Gemini, etc.) when working with code in this repository.

## Project Overview

**Falcon** (formerly ZAP) is an AI-powered API debugging assistant that runs in the terminal. It combines API testing with deep codebase awareness — when an API returns an error, Falcon searches your actual source code to find the cause and suggests fixes. Supports local LLMs (Ollama) and cloud providers (Google Gemini).

The binary is named `zap` (module path: `github.com/blackcoderx/zap`). The TUI displays **Falcon** branding (ASCII art splash screen).

---

## Build & Run Commands

```bash
# Build
go build -o zap.exe ./cmd/zap

# Run (interactive TUI)
./zap.exe

# Run with a saved request (CLI mode)
./zap.exe --request get-users --env prod

# Run with framework pre-selected (skip wizard step)
./zap.exe --framework gin

# Skip automatic API spec indexing on first run
./zap.exe --no-index

# Print version info
./zap.exe version

# Self-update binary
./zap.exe update

# Run tests
go test ./...

# Lint
go vet ./...
```

---

## Architecture

### Package Structure

```
cmd/zap/
  main.go          — Cobra root command, CLI flags, runCLI() for non-interactive mode
  update.go        — Self-update subcommand (go-github-selfupdate)

pkg/core/
  agent.go         — Agent struct: tool map, per-tool/total call limits, history mgmt
  react.go         — ReAct loop: ProcessMessage() and ProcessMessageWithEvents()
  init.go          — .zap folder creation, first-run setup wizard, config migration
  memory.go        — MemoryStore: persistent agent memory (.zap/memory.json)
  analysis.go      — Stack trace parsing, error context extraction
  types.go         — Core interfaces: Tool, AgentEvent, ConfirmableTool, ToolUsageStats
  prompt_integration.go — Helpers for injecting tool descriptions into system prompt
  prompt/          — 20-section system prompt builder

pkg/core/tools/
  registry.go      — Central Registry: RegisterAllTools() loads all 19 tool packages
  shared/          — Tier 1: http_request, assert_response, extract_value,
                     validate_json_schema, compare_responses, auth_bearer, auth_basic,
                     auth_oauth2, auth_helper, wait, retry, test_suite,
                     webhook_listener, performance_test
  debugging/       — Tier 2: read_file, write_file, list_files, search_code,
                     find_handler, analyze_endpoint, analyze_failure,
                     generate_tests, propose_fix, create_test_file
  persistence/     — Tier 2: variable, save_request, load_request, list_requests,
                     set_environment, list_environments
  agent/           — Tier 2: memory, export_results, run_tests, run_single_test, auto_test
  spec_ingester/         — Tier 3: ingest_spec (OpenAPI/Swagger/Postman → api-graph.json)
  functional_test_generator/ — generate_functional_tests
  security_scanner/      — security_scanner
  performance_engine/    — performance_engine (burst, ramp, soak modes)
  smoke_runner/          — smoke_runner
  idempotency_verifier/  — idempotency_verifier
  data_driven_engine/    — data_driven_engine
  schema_conformance/    — schema_conformance
  breaking_change_detector/ — breaking_change_detector
  dependency_mapper/     — dependency_mapper
  documentation_validator/  — documentation_validator
  api_drift_analyzer/    — api_drift_analyzer
  integration_orchestrator/ — integration_orchestrator
  regression_watchdog/   — regression_watchdog
  unit_test_scaffolder/  — unit_test_scaffolder

pkg/llm/
  client.go        — LLMClient interface (Chat, ChatStream)
  ollama.go        — Ollama client: local (http://localhost:11434) and cloud (https://ollama.com)
  gemini.go        — Google Gemini client (google.golang.org/genai)

pkg/storage/
  yaml.go          — YAML read/write for requests & environments
  env.go           — .env loading, {{VAR}} substitution
  schema.go        — JSON Schema helpers

pkg/tui/
  app.go           — tui.Run() entry point
  init.go          — InitialModel(): LLM client, agent, tools, ConfirmationManager, MemoryStore
  model.go         — Model struct (all UI state)
  update.go        — Bubble Tea Update(): AgentEvent → UI state transitions
  view.go          — Bubble Tea View(): renders logs, input, status, confirmation mode
  keys.go          — Key bindings, input history navigation
  styles.go        — Lip Gloss color palette (AccentColor, InputAreaBg, etc.)
  highlight.go     — Syntax highlighting helpers
```

### Core Interfaces

```go
// Tool — every agent capability implements this
type Tool interface {
    Name() string
    Description() string
    Parameters() string  // JSON Schema
    Execute(args string) (string, error)
}

// ConfirmableTool — tools requiring human approval before side effects
type ConfirmableTool interface {
    Tool
    SetEventCallback(callback EventCallback)
}

// EventCallback — emitted by the ReAct loop to update the TUI in real time
type EventCallback func(AgentEvent)

// AgentEvent types: "thinking", "tool_call", "observation", "answer",
//                   "error", "streaming", "tool_usage", "confirmation_required"
```

### ReAct Loop (`pkg/core/react.go`)

- `ProcessMessage(input)` — blocking, returns final answer string
- `ProcessMessageWithEvents(ctx, input, callback)` — streaming + events for TUI

The LLM is expected to produce:
```
Thought: <reasoning>
ACTION: tool_name({"arg": "value"})
```
or:
```
Final Answer: <response>
```

The parser (`parseResponse`) handles variations (missing `ACTION:` prefix, raw `tool_name(...)` calls, case differences).

### Tool Registration (`pkg/core/tools/registry.go`)

All tools are registered via a central `Registry` struct, never directly in the TUI. The `RegisterAllTools()` method calls:

```
initServices()                      — shared ResponseManager, VariableStore, PersistenceManager, HTTPTool
registerSharedTools()               — pkg/core/tools/shared/
registerDebuggingTools()            — pkg/core/tools/debugging/
registerPersistenceTools()          — pkg/core/tools/persistence/
registerAgentTools()                — pkg/core/tools/agent/
registerSpecIngesterTools()         — pkg/core/tools/spec_ingester/
registerFunctionalTestGeneratorTools()
registerSecurityScannerTools()
registerPerformanceEngineTools()
registerModuleTools()               — smoke, idempotency, data-driven, schema, breaking-change,
                                      dependency, doc-validator, api-drift, unit-test-scaffolder
registerWorkflowTools()             — integration_orchestrator, regression_watchdog
```

**When adding a new tool**: implement `core.Tool`, place it in the appropriate package, and add registration to the relevant `register*()` method in `registry.go`.

### Message Flow

```
User types → keys.go captures Enter
           → runAgentAsync() goroutine starts
           → Agent.ProcessMessageWithEvents(ctx, input, callback)
               → LLM ChatStream() → callback("streaming", chunk) → TUI renders live
               → parseResponse() extracts toolName + toolArgs
               → executeTool():
                   callback("tool_call", toolName) → TUI shows status
                   tool.Execute(args) called
                   callback("observation", result)
                   callback("tool_usage", stats) → counters updated
               → loop until Final Answer
           → callback("answer", finalAnswer)
           → program.Send(agentDoneMsg{})
```

---

## Configuration

### .zap Folder (created on first run)

```
.zap/
├── config.json         # LLM provider, model, framework, tool limits
├── memory.json         # Persistent MemoryStore (versioned)
├── history.jsonl       # Conversation history
├── manifest.json       # Workspace manifest (created by shared.CreateManifest)
├── requests/           # Saved API requests (YAML)
├── environments/       # Environment files (dev.yaml, prod.yaml, ...)
├── baselines/          # Reference snapshots for regression testing
├── snapshots/          # api-graph.json (from spec_ingester)
├── runs/               # Immutable test run records
├── exports/            # Markdown/JSON reports
├── logs/               # Internal tool logs
└── state/              # Agent state files
```

### config.json Schema

```json
{
  "provider": "ollama",
  "ollama": {
    "mode": "local",
    "url": "http://localhost:11434",
    "api_key": ""
  },
  "gemini": {
    "api_key": ""
  },
  "default_model": "llama3",
  "framework": "gin",
  "theme": "dark",
  "tool_limits": {
    "default_limit": 50,
    "total_limit": 200,
    "per_tool": {
      "http_request": 25,
      "performance_test": 5,
      "auto_test": 5,
      "read_file": 50,
      "search_code": 30,
      "variable": 100
    }
  }
}
```

**Config migration**: `init.go:migrateLegacyConfig()` automatically moves legacy top-level `ollama_url`/`ollama_api_key` fields into the new `ollama` sub-object.

### Supported Frameworks

| Language | Frameworks |
|----------|-----------|
| Go | gin, echo, chi, fiber |
| Python | fastapi, flask, django |
| Node.js | express, nestjs, hono |
| Java | spring |
| PHP | laravel |
| Ruby | rails |
| Rust | actix, axum |
| Other | other |

### Tool Limits

Default per-tool limits (defined in `pkg/core/init.go:DefaultToolLimits`):

| Risk Level | Tools | Default Limit |
|-----------|-------|---------------|
| High (external I/O) | `http_request` (25), `performance_test` (5), `webhook_listener` (10), `auth_oauth2` (10) | low |
| Medium (filesystem) | `read_file` (50), `list_files` (50), `search_code` (30), `save_request` (20) | medium |
| Low (in-memory) | `variable` (100), `assert_response` (100), `extract_value` (100) | high |
| AI/Orchestration | `auto_test` (5), `analyze_endpoint` (15), `generate_tests` (10), `run_tests` (10) | low-medium |

---

## Human-in-the-Loop (File Write Confirmation)

`write_file` implements `ConfirmableTool`. When the agent wants to write a file:

1. Tool calls `SetEventCallback` to get the TUI callback
2. Emits `AgentEvent{Type: "confirmation_required", FileConfirmation: &FileConfirmation{...}}`
3. TUI enters `confirmationMode`, shows colored diff, waits for `Y`/`N`
4. `ConfirmationManager.Approve()`/`.Reject()` unblocks the tool goroutine
5. Tool either writes the file or returns "rejected by user"

---

## Key Files Reference

| File | Purpose |
|------|---------|
| `cmd/zap/main.go` | CLI entry: Cobra commands, flags (`--framework`, `--request`, `--env`, `--no-index`) |
| `cmd/zap/update.go` | `zap update` self-update via `rhysd/go-github-selfupdate` |
| `pkg/core/agent.go` | Tool registry, per-tool/total call limits enforcement, history management |
| `pkg/core/react.go` | ReAct loop, `parseResponse()`, `executeTool()`, streaming |
| `pkg/core/init.go` | First-run wizard (Huh), `.zap` folder creation, config migration |
| `pkg/core/memory.go` | `MemoryStore`: persistent memory in `memory.json` |
| `pkg/core/analysis.go` | `ParseStackTrace()`, `ExtractErrorContext()`, `FormatErrorContext()` |
| `pkg/core/types.go` | `Tool`, `AgentEvent`, `ConfirmableTool`, `ToolUsageStats` interfaces/types |
| `pkg/core/tools/registry.go` | Central tool registry — edit this to add new tools |
| `pkg/core/tools/shared/` | HTTP, assertions, auth, retry, wait, webhooks, performance |
| `pkg/core/tools/debugging/` | File I/O, code search, handler finder, LLM analysis, fix proposals |
| `pkg/core/tools/persistence/` | Variable store, request save/load, environment management |
| `pkg/core/tools/agent/` | Memory, export, test runner, auto_test orchestrator |
| `pkg/core/tools/spec_ingester/` | OpenAPI/Swagger/Postman → Knowledge Graph |
| `pkg/llm/ollama.go` | Ollama client (local + cloud mode, streaming) |
| `pkg/llm/gemini.go` | Google Gemini client (`google.golang.org/genai`, streaming) |
| `pkg/tui/init.go` | `InitialModel()`, `registerTools()`, `newLLMClient()`, `configureToolLimits()` |
| `pkg/tui/styles.go` | Lip Gloss color palette, all `lipgloss.Style` definitions |
| `pkg/tui/update.go` | Bubble Tea `Update()` — handles all `tea.Msg` types |
| `pkg/tui/view.go` | Bubble Tea `View()` — renders full TUI layout |
| `pkg/storage/env.go` | `{{VAR}}` substitution from environment YAML files |

---

## Capabilities Summary

| Category | Tools |
|----------|-------|
| HTTP & Auth | `http_request`, `auth_bearer`, `auth_basic`, `auth_oauth2`, `auth_helper` |
| Validation | `assert_response`, `validate_json_schema`, `extract_value`, `compare_responses` |
| Flow Control | `wait`, `retry`, `test_suite`, `webhook_listener` |
| Codebase | `read_file`, `write_file`, `list_files`, `search_code`, `find_handler` |
| LLM Analysis | `analyze_endpoint`, `analyze_failure`, `generate_tests`, `propose_fix`, `create_test_file` |
| Persistence | `variable`, `save_request`, `load_request`, `list_requests`, `set_environment`, `list_environments` |
| Agent | `memory`, `export_results`, `run_tests`, `run_single_test`, `auto_test` |
| API Intelligence | `ingest_spec`, `generate_functional_tests`, `api_drift_analyzer`, `documentation_validator` |
| Security & Perf | `security_scanner`, `performance_engine`, `performance_test` |
| Quality | `smoke_runner`, `idempotency_verifier`, `data_driven_engine`, `schema_conformance`, `unit_test_scaffolder` |
| Observability | `breaking_change_detector`, `dependency_mapper`, `regression_watchdog`, `integration_orchestrator` |

---

## Coding Conventions

- All tool packages under `pkg/core/tools/<package>/` export a `New<ToolName>Tool(...)` constructor
- Tools must be registered in `pkg/core/tools/registry.go`, not in `pkg/tui/`
- Shared services (`ResponseManager`, `VariableStore`, `HTTPTool`) are initialized once in `registry.initServices()` and injected as dependencies
- The `shared.ConfirmationManager` is the single point of truth for file write approval — never block a goroutine directly
- Use `pkg/storage/env.go` for `{{VAR}}` substitution, not ad-hoc string replacement
- Config is read via `spf13/viper`; tool limits come from `core.DefaultToolLimits` (overridable via config)
- Error messages from tools should be plain strings (no `fmt.Fprintf(os.Stderr, ...)` in tool code)
