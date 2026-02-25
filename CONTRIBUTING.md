# Contributing to Falcon

Thank you for your interest in contributing to Falcon! This document provides guidelines and information for contributors.

## Getting Started

### Prerequisites

- Go 1.25.3 or higher
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
   go build -o zap.exe ./cmd/zap
   ```
4. Run tests:
   ```bash
   go test ./...
   ```

## Project Structure

```
falcon/
├── cmd/zap/           # Application entry point
├── pkg/
│   ├── core/          # Agent logic and ReAct loop
│   │   └── tools/     # 4-tier tool system
│   │       ├── shared/      # Foundation & HTTP
│   │       ├── debugging/   # Code investigation
│   │       ├── persistence/ # State & Environments
│   │       ├── agent/       # Agent lifecycle
│   │       ├── spec_ingester/ # API Intelligence
│   │       └── (modules)/   # Security, Performance, QA, etc.
│   ├── llm/           # LLM client implementations
│   ├── storage/       # Low-level I/O
│   ├── web/           # Embedded web dashboard
│   └── tui/           # Terminal UI
├── .zap/              # Runtime config & memory (created on first run)
├── CLAUDE.md          # Development guidelines for AI assistants
└── README.md          # User documentation
```

See the README files in each package for detailed documentation:

- [pkg/core/README.md](pkg/core/README.md) — Agent and ReAct loop
- [pkg/core/tools/README.md](pkg/core/tools/README.md) — Tool implementation guide
- [pkg/llm/README.md](pkg/llm/README.md) — LLM providers
- [pkg/storage/README.md](pkg/storage/README.md) — Persistence layer
- [pkg/tui/README.md](pkg/tui/README.md) — Terminal UI

## How to Contribute

### Reporting Bugs

1. Check if the issue already exists in [GitHub Issues](https://github.com/blackcoderx/falcon/issues)
2. Create a new issue with:
   - Clear title describing the bug
   - Steps to reproduce
   - Expected vs actual behavior
   - Environment details (OS, Go version, Ollama version)

### Suggesting Features

1. Check existing issues and discussions
2. Create a new issue with:
   - Clear description of the feature
   - Use case / motivation
   - Proposed implementation (optional)

### Pull Requests

1. Create a feature branch:
   ```bash
   git checkout -b feature/my-feature
   ```

2. Make your changes following the [code style guidelines](#code-style)

3. Add tests for new functionality

4. Run tests:
   ```bash
   go test ./...
   ```

5. Commit with clear messages:
   ```bash
   git commit -m "feat: add new tool for X"
   ```

6. Push and create a PR:
   ```bash
   git push origin feature/my-feature
   ```

## Code Style

### Go Conventions

- Follow standard Go formatting (`go fmt`)
- Use meaningful variable names
- Add comments for exported functions
- Keep functions focused and small

### Commit Messages

Use conventional commits:

- `feat:` — New feature
- `fix:` — Bug fix
- `docs:` — Documentation changes
- `test:` — Test changes
- `refactor:` — Code refactoring
- `chore:` — Build/tooling changes

Examples:
```
feat: add OAuth2 PKCE flow support
fix: handle empty response body in HTTP tool
docs: update tool creation guide
test: add tests for variable substitution
```

## Common Tasks

### Adding a New Tool

1. **Choose a tier**: Place your tool in the appropriate folder under `pkg/core/tools/`:
   - `shared/` — Foundation tools used by other tools
   - `debugging/` — Tools for codebase analysis and fixing
   - `persistence/` — Tools for state and environment management
   - `agent/` — Agent-internal management tools
   - `(module)/` — Specific autonomous modules (e.g., `security_scanner/`)

2. **Implement the tool**:
   ```go
   package mytier

   type MyTool struct{}

   func NewMyTool() *MyTool { return &MyTool{} }

   func (t *MyTool) Name() string        { return "my_tool" }
   func (t *MyTool) Description() string { return "Does something useful" }
   func (t *MyTool) Parameters() string  { return `{"type":"object","properties":{...}}` }
   func (t *MyTool) Execute(args string) (string, error) { ... }
   ```

3. **Register in `pkg/core/tools/registry.go`**: Add your tool to the relevant `register*()` method so it's wired into the agent.

4. **Add tests**: Create `tool_test.go` in the same directory.

5. **Update the system prompt**: If the tool adds a new capability category, update `pkg/core/prompt/`.

### Adding a New LLM Provider

1. Create `pkg/llm/newprovider.go` implementing the `LLMClient` interface
2. Update the config struct in `pkg/core/init.go`
3. Add a provider option to the setup wizard in `pkg/core/init.go`
4. Update `pkg/tui/init.go` to instantiate the new client
5. Add documentation to `pkg/llm/README.md`

### Adding a New Keyboard Shortcut

1. Add the key binding in `pkg/tui/keys.go`
2. Add the handler in `pkg/tui/update.go`
3. Update the footer help text in `pkg/tui/view.go`

## Testing

### Running Tests

```bash
# All tests
go test ./...

# Specific package
go test ./pkg/core/...

# With coverage
go test -cover ./...

# Verbose output
go test -v ./pkg/core/tools/...
```

### Writing Tests

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

## Questions?

- Open a [GitHub Discussion](https://github.com/blackcoderx/falcon/discussions)
- Check existing issues and PRs

Thank you for contributing to Falcon!
