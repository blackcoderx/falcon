# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Falcon is a terminal-based AI agent for API developers. It combines API testing, debugging, and code analysis in an interactive TUI powered by a ReAct (Reason+Act) loop. Built in Go with the Charm ecosystem (Bubble Tea, Lip Gloss, Glamour).

## Build & Test Commands

```bash
# Build
go build -o falcon ./cmd/falcon          # Unix
go build -o falcon.exe ./cmd/falcon      # Windows

# Test
go test ./...                             # All tests
go test ./pkg/core/...                    # Core package only
go test -v -run TestName ./pkg/core/...   # Single test
go test -cover ./...                      # With coverage

# Release (multi-platform via GoReleaser)
goreleaser release
```

## Architecture

### Entry Point
`cmd/falcon/main.go` — Cobra CLI that loads `.env`, initializes `.falcon/` folder (runs setup wizard on first run), optionally starts the web dashboard, then launches the TUI or executes a saved request in CLI mode (`falcon --request get-users --env prod`).

### Core Packages

- **`pkg/core/`** — Agent struct (tool registry, history, call limits with mutex-guarded state), ReAct loop (`react.go`: think→act→observe cycle with streaming events and retry logic), memory store, system prompt builder, and the `Tool`/`ConfirmableTool` interfaces.

- **`pkg/core/tools/`** — 28+ tools organized in tiers, all registered via `registry.go:RegisterAllTools()`:
  - `shared/` — Foundation: HTTP requests, assertions, auth, variables, webhooks
  - `persistence/` — Saved requests, environments, variables
  - `debugging/` — File I/O, code search, handler discovery, LLM-powered analysis/fix proposals
  - `agent/` — Memory, test execution, autonomous testing
  - Specialized modules: `spec_ingester/`, `functional_test_generator/`, `security_scanner/`, `performance_engine/`, `smoke_runner/`, `regression_watchdog/`, `integration_orchestrator/`, `data_driven_engine/`, `idempotency_verifier/`

- **`pkg/llm/`** — Pluggable LLM providers via self-registering `Provider` interface. Clients implement `LLMClient` (Chat, ChatStream, CheckConnection). Current providers: Ollama, Gemini, OpenRouter. Adding a provider: create `<name>.go` (client) + `<name>_provider.go` (registration) + register in `register_providers.go`.

- **`pkg/tui/`** — Bubble Tea app (Elm architecture). `model.go` holds state, `update.go` handles events, `view.go` renders. Agent events stream in real-time. File writes require user confirmation with diff preview.

- **`pkg/storage/`** — Low-level YAML/env file I/O and JSON Schema helpers.

### Key Interfaces

```go
// pkg/core/types.go
type Tool interface {
    Name() string           // snake_case identifier
    Description() string    // For LLM system prompt
    Parameters() string     // JSON Schema string
    Execute(args string) (string, error)
}
```

### ReAct Loop Flow
User input → `Agent.ProcessMessageWithEvents()` → build system prompt with tool descriptions → LLM streaming call → parse response for `ACTION: tool_name({...})` or `Final Answer:` → execute tool → append observation → loop until final answer. Retries LLM calls up to 3× with exponential backoff.

### Runtime State (`.falcon/` folder)
`config.yaml`, `memory.json`, `spec.yaml`, `variables.json`, plus `requests/`, `environments/`, `flows/`, `reports/`, `baselines/`, `sessions/` directories. Naming convention: `<type>_report_<api>_<timestamp>.md` for reports, `<type>_<desc>.yaml` for flows.

## Configuration

Provider-agnostic YAML config (`.falcon/config.yaml`). Provider settings stored in generic `provider_config` map. Environment variable fallbacks: `OLLAMA_API_KEY`, `GEMINI_API_KEY`, `OPENROUTER_API_KEY`. Automatic migration from legacy config formats.

## Conventions

- Commit messages: `feat:`, `fix:`, `docs:`, `test:`, `refactor:`, `chore:`
- Thread safety: mutex-guarded shared state (Agent, ResponseManager, VariableStore)
- Tool safety: per-tool and total call limits, path validation (blocks `../` and absolute paths in write tools), confirmation workflow for file writes
- Standard Go formatting (`go fmt`)
