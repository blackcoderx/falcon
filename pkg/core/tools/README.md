# pkg/core/tools

This package contains all 28+ tool implementations for ZAP. Each tool provides a specific capability that the agent can invoke during API testing and debugging.

## Package Overview

```
pkg/core/tools/
├── http.go          # HTTP request tool with variable substitution
├── file.go          # read_file, list_files tools
├── write.go         # write_file with human-in-the-loop confirmation
├── search.go        # search_code (ripgrep with native fallback)
├── persistence.go   # save_request, load_request, environments
├── assert.go        # Response validation (status, headers, body, timing)
├── extract.go       # Value extraction (JSON path, headers, cookies, regex)
├── variables.go     # Session/global variable management
├── timing.go        # wait, retry tools
├── schema.go        # JSON Schema validation
├── suite.go         # Test suite execution
├── diff.go          # Response comparison for regression testing
├── perf.go          # Performance/load testing
├── webhook.go       # Webhook listener (temporary HTTP server)
├── memory.go        # Agent memory operations
├── manager.go       # ResponseManager for sharing HTTP responses
├── confirm.go       # ConfirmationManager for file write approval
├── pathutil.go      # Path utilities (security bounds checking)
└── auth/            # Authentication tools subpackage
    ├── bearer.go    # Bearer token auth
    ├── basic.go     # HTTP Basic auth
    ├── oauth2.go    # OAuth2 flows
    └── helper.go    # JWT parsing, auth helpers
```

## Tool Interface

Every tool must implement:

```go
type Tool interface {
    Name() string                           // Unique identifier
    Description() string                    // Human-readable description (for LLM)
    Parameters() string                     // JSON Schema of parameters
    Execute(args string) (string, error)    // Main execution
}
```

## Tool Categories

### HTTP & Persistence

| Tool | File | Description |
|------|------|-------------|
| `http_request` | `http.go` | Make HTTP requests with variable substitution, status meanings, error hints |
| `save_request` | `persistence.go` | Save request to YAML with `{{VAR}}` placeholders |
| `load_request` | `persistence.go` | Load saved request with environment substitution |
| `list_requests` | `persistence.go` | List all saved requests |
| `set_environment` | `persistence.go` | Switch active environment |
| `list_environments` | `persistence.go` | List available environments |

### Codebase Analysis

| Tool | File | Description |
|------|------|-------------|
| `read_file` | `file.go` | Read file contents (100KB limit) |
| `list_files` | `file.go` | List files with glob patterns |
| `search_code` | `search.go` | Search patterns (ripgrep + native fallback) |
| `write_file` | `write.go` | Write files with human-in-the-loop confirmation |

### Testing & Validation

| Tool | File | Description |
|------|------|-------------|
| `assert_response` | `assert.go` | Validate status, headers, body, JSON path, timing |
| `extract_value` | `extract.go` | Extract from JSON path, headers, cookies, regex |
| `validate_json_schema` | `schema.go` | JSON Schema validation (draft-07, 2020-12) |
| `test_suite` | `suite.go` | Multi-test execution with assertions |
| `compare_responses` | `diff.go` | Regression testing with baseline comparison |

### Variables & Timing

| Tool | File | Description |
|------|------|-------------|
| `variable` | `variables.go` | Session/global variables with persistence |
| `wait` | `timing.go` | Add delays for async operations |
| `retry` | `timing.go` | Retry with exponential backoff |

### Performance & Webhooks

| Tool | File | Description |
|------|------|-------------|
| `performance_test` | `perf.go` | Load testing with p50/p95/p99 metrics |
| `webhook_listener` | `webhook.go` | Temporary HTTP server for callbacks |

### Authentication

See [auth/README.md](auth/README.md) for details.

| Tool | File | Description |
|------|------|-------------|
| `auth_bearer` | `auth/bearer.go` | Create Bearer token headers |
| `auth_basic` | `auth/basic.go` | Create HTTP Basic auth headers |
| `auth_oauth2` | `auth/oauth2.go` | OAuth2 flows (client_credentials, password) |
| `auth_helper` | `auth/helper.go` | Parse JWT tokens, decode auth headers |

## Creating a New Tool

### Step 1: Create the File

Create a new file in `pkg/core/tools/`:

```go
// pkg/core/tools/mytool.go
package tools

import (
    "encoding/json"
    "github.com/blackcoderx/zap/pkg/core"
)

type MyTool struct {
    // Add any dependencies here
}

func NewMyTool() *MyTool {
    return &MyTool{}
}
```

### Step 2: Implement the Interface

```go
func (t *MyTool) Name() string {
    return "my_tool"
}

func (t *MyTool) Description() string {
    return "Does something useful for API testing"
}

func (t *MyTool) Parameters() string {
    return `{
        "type": "object",
        "properties": {
            "input": {
                "type": "string",
                "description": "The input to process"
            },
            "option": {
                "type": "boolean",
                "description": "Optional flag",
                "default": false
            }
        },
        "required": ["input"]
    }`
}

func (t *MyTool) Execute(args string) (string, error) {
    // Parse arguments
    var params struct {
        Input  string `json:"input"`
        Option bool   `json:"option"`
    }
    if err := json.Unmarshal([]byte(args), &params); err != nil {
        return "", fmt.Errorf("invalid arguments: %w", err)
    }

    // Do the work
    result := processInput(params.Input, params.Option)

    // Return result (string)
    return result, nil
}
```

### Step 3: Register the Tool

In `pkg/tui/init.go`:

```go
agent.RegisterTool(tools.NewMyTool())
```

### Step 4: Update System Prompt (if needed)

If the tool requires special instructions, update `pkg/core/prompt.go`.

## Tool Patterns

### Accessing Shared State

Use `ResponseManager` to share HTTP responses between tools:

```go
type MyTool struct {
    responseManager *ResponseManager
}

func NewMyTool(rm *ResponseManager) *MyTool {
    return &MyTool{responseManager: rm}
}

func (t *MyTool) Execute(args string) (string, error) {
    // Get the last HTTP response
    resp := t.responseManager.GetLastResponse()
    if resp == nil {
        return "No response available", nil
    }
    // Use resp.StatusCode, resp.Body, resp.Headers, etc.
}
```

### Human-in-the-Loop Confirmation

For tools that modify files, implement `ConfirmableTool`:

```go
type MyWriteTool struct {
    confirmManager *ConfirmationManager
}

func (t *MyWriteTool) SetConfirmationManager(cm *ConfirmationManager) {
    t.confirmManager = cm
}

func (t *MyWriteTool) Execute(args string) (string, error) {
    // Generate the changes
    oldContent := readFile(path)
    newContent := generateNewContent(oldContent)

    // Request confirmation
    approved, err := t.confirmManager.RequestConfirmation(
        path, oldContent, newContent,
    )
    if err != nil {
        return "", err
    }
    if !approved {
        return "Change rejected by user", nil
    }

    // Apply the change
    return writeFile(path, newContent)
}
```

### Variable Substitution

Use the `SubstituteVariables` function:

```go
func (t *MyTool) Execute(args string) (string, error) {
    var params struct {
        URL string `json:"url"`
    }
    json.Unmarshal([]byte(args), &params)

    // Substitute {{VAR}} placeholders
    url := storage.SubstituteVariables(params.URL, t.variables)
    // url now has variables replaced
}
```

### Error Handling

Return helpful error messages:

```go
func (t *MyTool) Execute(args string) (string, error) {
    // Return errors that help the LLM understand what went wrong
    if err := validate(args); err != nil {
        return "", fmt.Errorf("invalid arguments: %w (expected: {\"url\": \"...\"})", err)
    }

    result, err := doWork(args)
    if err != nil {
        // Include context for debugging
        return "", fmt.Errorf("failed to process: %w (input was: %s)", err, args)
    }

    return result, nil
}
```

## Key Files Explained

### http.go

The HTTP tool is the most used tool. Key features:

- Variable substitution (`{{VAR}}` in URL, headers, body)
- Status code meanings (human-readable explanations)
- Error hints (framework-specific debugging tips)
- Response timing and size display

### search.go

Code search with two backends:

1. **ripgrep** (preferred) - Fast, respects .gitignore
2. **Native Go** (fallback) - Works without external dependencies

```go
// Tries ripgrep first
results, err := searchWithRipgrep(pattern, path)
if err != nil {
    // Falls back to native Go implementation
    results, err = searchNative(pattern, path)
}
```

### confirm.go

The `ConfirmationManager` coordinates file write approval:

1. Tool calls `RequestConfirmation(path, old, new)`
2. Manager emits `confirmation_required` event
3. TUI displays diff and waits for Y/N
4. User decision sent back via channel
5. Tool proceeds or aborts

### manager.go

The `ResponseManager` stores the last HTTP response so other tools (assert, extract, schema) can access it without re-making the request.

```go
type ResponseManager struct {
    lastResponse *HTTPResponse
    mu           sync.RWMutex
}
```

## Testing Tools

Each tool should have tests in `*_test.go`:

```go
func TestMyTool_Execute(t *testing.T) {
    tool := NewMyTool()

    result, err := tool.Execute(`{"input": "test"}`)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    if !strings.Contains(result, "expected") {
        t.Errorf("unexpected result: %s", result)
    }
}
```

Run tests:

```bash
go test ./pkg/core/tools/...
```

## Security Considerations

1. **Path bounds checking** - `pathutil.go` prevents directory traversal
2. **File size limits** - `read_file` has 100KB limit
3. **Human approval** - `write_file` requires confirmation
4. **Variable scoping** - Environment variables isolated per environment
5. **No credential logging** - Sensitive values masked in output
