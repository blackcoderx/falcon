# Contributing to Falcon

Thank you for your interest in contributing to Falcon! This document provides guidelines and information for contributors.

## Getting Started

### Prerequisites

- Go 1.21 or higher
- [Ollama](https://ollama.ai/) (optional, for testing with a local LLM)
- Git

### Setup

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR-USERNAME/falcon.git
   cd falcon
   ```
3. Build the project:
   ```bash
   go build -o falcon.exe ./cmd/falcon   # Windows
   go build -o falcon ./cmd/falcon       # Unix
   ```
4. Run tests:
   ```bash
   go test ./...
   ```

---

## Project Structure

```
falcon/
├── cmd/falcon/        # Application entry point (Cobra CLI)
│   ├── main.go        # Root command, flags, CLI mode
│   └── update.go      # Self-update logic
├── pkg/
│   ├── core/          # Agent logic, ReAct loop, tool interfaces
│   │   ├── tools/     # 28+ tools in tiered packages
│   │   │   ├── shared/               # Foundation: HTTP, auth, assertions
│   │   │   ├── debugging/            # Code analysis and fixes
│   │   │   ├── persistence/          # Requests, environments, variables
│   │   │   ├── agent/                # Memory, test execution, automation
│   │   │   ├── spec_ingester/        # OpenAPI/Postman spec parsing
│   │   │   ├── functional_test_generator/
│   │   │   ├── security_scanner/
│   │   │   ├── performance_engine/
│   │   │   ├── smoke_runner/
│   │   │   ├── idempotency_verifier/
│   │   │   ├── data_driven_engine/
│   │   │   ├── regression_watchdog/
│   │   │   └── integration_orchestrator/
│   │   └── prompt/    # System prompt builder
│   ├── llm/           # Pluggable LLM provider system
│   │   ├── ollama/    # Ollama client + self-registration
│   │   ├── gemini/    # Gemini client + self-registration
│   │   └── openrouter/  # OpenRouter client + self-registration
│   ├── storage/       # Low-level YAML/env file I/O
│   └── tui/           # Terminal UI (Bubble Tea)
├── .falcon/           # Runtime config & memory (created on first run)
├── CLAUDE.md          # Development guidelines for AI assistants
└── README.md          # User documentation
```

See the README files in each package for detailed documentation:

- [pkg/core/README.md](pkg/core/README.md) — Agent and ReAct loop
- [pkg/core/tools/README.md](pkg/core/tools/README.md) — Tool implementation guide
- [pkg/llm/README.md](pkg/llm/README.md) — LLM providers
- [pkg/storage/README.md](pkg/storage/README.md) — Persistence layer
- [pkg/tui/README.md](pkg/tui/README.md) — Terminal UI

---

## How to Contribute

### Reporting Bugs

1. Check if the issue already exists in [GitHub Issues](https://github.com/blackcoderx/falcon/issues)
2. Create a new issue with:
   - Clear title describing the bug
   - Steps to reproduce
   - Expected vs actual behavior
   - Environment details (OS, Go version, provider/model)

### Suggesting Features

1. Check existing issues and discussions
2. Create a new issue with:
   - Clear description of the feature
   - Use case / motivation
   - Proposed implementation (optional)

### Pull Requests

1. Create a feature branch:
   ```bash
   git checkout -b feat/my-feature
   ```

2. Make your changes following the [code style guidelines](#code-style)

3. Add tests for new functionality

4. Run tests and verify the build:
   ```bash
   go test ./...
   go build -o falcon.exe ./cmd/falcon
   ```

5. Commit with a clear message (see [commit messages](#commit-messages)):
   ```bash
   git commit -m "feat: add new tool for X"
   ```

6. Push and create a PR:
   ```bash
   git push origin feat/my-feature
   ```

---

## Common Tasks

### Adding a New Tool

Tools are the primary way to extend Falcon's capabilities. Each tool is a Go struct that implements the `core.Tool` interface.

#### 1. Choose a location

Place your tool in the appropriate package under `pkg/core/tools/`:

| Package | Purpose |
|---------|---------|
| `shared/` | Foundation tools used by other tools (HTTP, auth) |
| `debugging/` | Tools for codebase analysis and fixing |
| `persistence/` | Tools for state and environment management |
| `agent/` | Agent-internal management tools |
| `<module>/` | Autonomous domain modules (e.g., `security_scanner/`) |

#### 2. Implement the `Tool` interface

```go
package mytier

type MyTool struct{}

func NewMyTool() *MyTool { return &MyTool{} }

func (t *MyTool) Name() string        { return "my_tool" }
func (t *MyTool) Description() string { return "Does something useful" }
func (t *MyTool) Parameters() string {
    return `{
        "type": "object",
        "properties": {
            "input": {"type": "string", "description": "The input value"}
        },
        "required": ["input"]
    }`
}

func (t *MyTool) Execute(args string) (string, error) {
    var params struct {
        Input string `json:"input"`
    }
    if err := json.Unmarshal([]byte(args), &params); err != nil {
        return "", fmt.Errorf("invalid args: %w", err)
    }
    // ... implementation
    return result, nil
}
```

#### 3. Register in `pkg/core/tools/registry.go`

Add your tool to the relevant `register*()` method:

```go
func (r *Registry) registerSharedTools() {
    // existing tools...
    r.agent.RegisterTool(NewMyTool())
}
```

#### 4. Add tests

Create `my_tool_test.go` in the same directory:

```go
func TestMyTool_Execute(t *testing.T) {
    tool := NewMyTool()

    tests := []struct {
        name    string
        args    string
        want    string
        wantErr bool
    }{
        {
            name: "valid input",
            args: `{"input": "test"}`,
            want: "expected output",
        },
        {
            name:    "invalid JSON",
            args:    "invalid",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := tool.Execute(tt.args)
            if (err != nil) != tt.wantErr {
                t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
            }
            if got != tt.want {
                t.Errorf("Execute() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

#### 5. Update the system prompt (if needed)

If the tool introduces a new capability category, update `pkg/core/prompt/` to inform the LLM about it.

#### For tools requiring human approval (file writes)

Implement `core.ConfirmableTool` to hook into the confirmation workflow:

```go
type MyWriteTool struct {
    eventCallback core.EventCallback
}

func (t *MyWriteTool) SetEventCallback(cb core.EventCallback) {
    t.eventCallback = cb
}
```

---

### Adding a New LLM Provider

Providers use a self-registration pattern — each provider registers itself via `init()`.

#### 1. Create the client

`pkg/llm/<name>/<name>.go` — implement `llm.LLMClient`:

```go
type MyClient struct {
    model  string
    apiKey string
}

func (c *MyClient) Chat(messages []llm.Message) (string, error) { ... }
func (c *MyClient) ChatStream(messages []llm.Message, cb llm.StreamCallback) (string, error) { ... }
func (c *MyClient) CheckConnection() error { ... }
func (c *MyClient) GetModel() string { return c.model }
```

#### 2. Create the provider

`pkg/llm/<name>/<name>_provider.go` — implement `llm.Provider` and self-register:

```go
type MyProvider struct{}

func (p *MyProvider) ID() string          { return "myprovider" }
func (p *MyProvider) DisplayName() string { return "My Provider" }
func (p *MyProvider) DefaultModel() string { return "my-default-model" }

func (p *MyProvider) SetupFields() []llm.SetupField {
    return []llm.SetupField{
        {
            Key:         "api_key",
            Type:        llm.FieldTypeInput,
            Title:       "API Key",
            Secret:      true,
            EnvFallback: "MYPROVIDER_API_KEY",
        },
    }
}

func (p *MyProvider) BuildClient(values map[string]string, model string) (llm.LLMClient, error) {
    return &MyClient{model: model, apiKey: values["api_key"]}, nil
}

func init() {
    llm.Register(&MyProvider{})
}
```

#### 3. Activate the provider

Add a blank import to `pkg/llm/register_providers.go`:

```go
import (
    _ "github.com/blackcoderx/falcon/pkg/llm/myprovider"
)
```

#### 4. Document it

Add a section to `pkg/llm/README.md` describing setup fields and any special requirements.

---

### Adding a New Keyboard Shortcut

1. Define the key in `pkg/tui/keys.go`
2. Handle the keypress in `pkg/tui/update.go`
3. Update the footer help text in `pkg/tui/view.go`

---

## Code Style

### Go Conventions

- Follow standard Go formatting (`go fmt`)
- Use meaningful variable names
- Add comments for exported functions and types
- Keep functions focused and small
- Thread-safety: use mutex guards for shared state (see `Agent`, `ResponseManager`, `VariableStore`)

### Commit Messages

Use conventional commits:

| Prefix | When to use |
|--------|-------------|
| `feat:` | New feature |
| `fix:` | Bug fix |
| `docs:` | Documentation changes |
| `test:` | Test changes |
| `refactor:` | Code refactoring |
| `chore:` | Build/tooling changes |

Examples:
```
feat: add OAuth2 PKCE flow support
fix: handle empty response body in HTTP tool
docs: update tool creation guide
test: add tests for variable substitution
refactor: extract HTTP retry logic into shared helper
```

### Security Guidelines

- Never store secrets (API keys, tokens) in tool outputs or memory entries — check for the `Secret` field pattern used in existing tools
- Path safety: block `../` and absolute paths in tools that read/write files (see `debugging/write_file.go`)
- Tool limits: respect per-tool and total call limits via `agent.ExecuteTool()` — do not bypass them
- File writes always go through the confirmation workflow for user approval

---

## Testing

### Running Tests

```bash
# All tests
go test ./...

# Specific package
go test ./pkg/core/...

# Single test by name
go test -v -run TestMyTool ./pkg/core/tools/...

# With coverage
go test -cover ./...
```

### Writing Tests

Follow the table-driven test pattern used throughout the codebase:

```go
func TestMyTool_Execute(t *testing.T) {
    tool := NewMyTool()

    tests := []struct {
        name    string
        args    string
        want    string
        wantErr bool
    }{
        {name: "valid input", args: `{"input": "test"}`, want: "expected output"},
        {name: "invalid JSON", args: "invalid", wantErr: true},
        {name: "missing required field", args: `{}`, wantErr: true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := tool.Execute(tt.args)
            if (err != nil) != tt.wantErr {
                t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
            }
            if !tt.wantErr && got != tt.want {
                t.Errorf("Execute() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

---

## Questions?

- Open a [GitHub Discussion](https://github.com/blackcoderx/falcon/discussions)
- Check existing issues and PRs at [GitHub Issues](https://github.com/blackcoderx/falcon/issues)

Thank you for contributing to Falcon!
