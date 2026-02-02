# pkg/storage

This package handles request persistence and environment management. It provides YAML-based storage for API requests and variable substitution for environments.

## Package Overview

```
pkg/storage/
├── schema.go    # Data structures: Request, Environment, Collection
├── yaml.go      # YAML file read/write operations
└── env.go       # Variable substitution engine ({{VAR}} placeholders)
```

## Data Structures

### Request

Represents a saved API request:

```go
type Request struct {
    Name        string            `yaml:"name"`
    Method      string            `yaml:"method"`
    URL         string            `yaml:"url"`
    Headers     map[string]string `yaml:"headers,omitempty"`
    Body        string            `yaml:"body,omitempty"`
    Description string            `yaml:"description,omitempty"`
}
```

**Example YAML:**

```yaml
# .zap/requests/get-users.yaml
name: Get Users
method: GET
url: "{{BASE_URL}}/api/users"
headers:
  Authorization: "Bearer {{API_TOKEN}}"
  Content-Type: application/json
description: Fetches all users from the API
```

### Environment

Represents a set of variables for a specific environment:

```go
type Environment struct {
    Name      string            `yaml:"name"`
    Variables map[string]string `yaml:"variables"`
}
```

**Example YAML:**

```yaml
# .zap/environments/dev.yaml
name: dev
variables:
  BASE_URL: http://localhost:3000
  API_TOKEN: dev-token-123
  DEBUG: "true"
```

### Collection

A group of related requests (for future use):

```go
type Collection struct {
    Name        string    `yaml:"name"`
    Description string    `yaml:"description"`
    Requests    []Request `yaml:"requests"`
}
```

## YAML Operations (yaml.go)

### Saving Requests

```go
request := &storage.Request{
    Name:   "Get Users",
    Method: "GET",
    URL:    "{{BASE_URL}}/api/users",
    Headers: map[string]string{
        "Authorization": "Bearer {{API_TOKEN}}",
    },
}

err := storage.SaveRequest(".zap/requests/get-users.yaml", request)
```

### Loading Requests

```go
request, err := storage.LoadRequest(".zap/requests/get-users.yaml")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Loaded: %s %s\n", request.Method, request.URL)
```

### Listing Requests

```go
requests, err := storage.ListRequests(".zap/requests/")
// Returns: []string{"get-users", "create-user", "delete-user"}
```

### Saving Environments

```go
env := &storage.Environment{
    Name: "prod",
    Variables: map[string]string{
        "BASE_URL":  "https://api.example.com",
        "API_TOKEN": "prod-token-xyz",
    },
}

err := storage.SaveEnvironment(".zap/environments/prod.yaml", env)
```

### Loading Environments

```go
env, err := storage.LoadEnvironment(".zap/environments/prod.yaml")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Loaded environment: %s\n", env.Name)
```

### Listing Environments

```go
envs, err := storage.ListEnvironments(".zap/environments/")
// Returns: []string{"dev", "prod", "staging"}
```

## Variable Substitution (env.go)

### Basic Substitution

Replace `{{VAR}}` placeholders with environment values:

```go
variables := map[string]string{
    "BASE_URL":  "http://localhost:3000",
    "API_TOKEN": "abc123",
}

url := "{{BASE_URL}}/api/users"
result := storage.SubstituteVariables(url, variables)
// Result: "http://localhost:3000/api/users"
```

### Substituting Requests

Apply substitution to all fields of a request:

```go
request := &storage.Request{
    URL: "{{BASE_URL}}/api/users",
    Headers: map[string]string{
        "Authorization": "Bearer {{API_TOKEN}}",
    },
}

substituted := storage.SubstituteRequest(request, variables)
// URL: "http://localhost:3000/api/users"
// Headers["Authorization"]: "Bearer abc123"
```

### Nested Variables

Variables can reference other variables (resolved recursively):

```yaml
# Environment
API_URL: "{{BASE_URL}}/api"
USERS_URL: "{{API_URL}}/users"
BASE_URL: "http://localhost:3000"
```

```go
result := storage.SubstituteVariables("{{USERS_URL}}", variables)
// Result: "http://localhost:3000/api/users"
```

### Undefined Variables

Undefined variables remain as placeholders:

```go
result := storage.SubstituteVariables("{{UNDEFINED}}/api", variables)
// Result: "{{UNDEFINED}}/api"
```

## File Structure

ZAP creates this structure in the `.zap/` folder:

```
.zap/
├── config.json              # Main configuration
├── history.jsonl            # Conversation history
├── memory.json              # Agent memory
├── requests/                # Saved API requests
│   ├── get-users.yaml
│   ├── create-user.yaml
│   └── health-check.yaml
└── environments/            # Environment configs
    ├── dev.yaml
    ├── staging.yaml
    └── prod.yaml
```

## Usage Patterns

### Save and Load Flow

```go
// Save a request
request := &storage.Request{
    Name:   "Get Users",
    Method: "GET",
    URL:    "{{BASE_URL}}/api/users",
}
storage.SaveRequest(".zap/requests/get-users.yaml", request)

// Load environment
env, _ := storage.LoadEnvironment(".zap/environments/dev.yaml")

// Load and substitute request
loaded, _ := storage.LoadRequest(".zap/requests/get-users.yaml")
substituted := storage.SubstituteRequest(loaded, env.Variables)

// Now substituted.URL is "http://localhost:3000/api/users"
```

### Environment Switching

```go
var currentEnv *storage.Environment

func switchEnvironment(envName string) error {
    path := filepath.Join(".zap/environments", envName+".yaml")
    env, err := storage.LoadEnvironment(path)
    if err != nil {
        return err
    }
    currentEnv = env
    return nil
}
```

### Request with Variables

Create a request template that works across environments:

```yaml
# .zap/requests/auth-login.yaml
name: Login
method: POST
url: "{{BASE_URL}}/auth/login"
headers:
  Content-Type: application/json
body: |
  {
    "username": "{{TEST_USER}}",
    "password": "{{TEST_PASSWORD}}"
  }
```

Then define different credentials per environment:

```yaml
# .zap/environments/dev.yaml
BASE_URL: http://localhost:3000
TEST_USER: testuser
TEST_PASSWORD: testpass123

# .zap/environments/staging.yaml
BASE_URL: https://staging.example.com
TEST_USER: stageuser
TEST_PASSWORD: stagepass456
```

## Error Handling

All functions return descriptive errors:

```go
func LoadRequest(path string) (*Request, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            return nil, fmt.Errorf("request file not found: %s", path)
        }
        return nil, fmt.Errorf("failed to read request file: %w", err)
    }

    var request Request
    if err := yaml.Unmarshal(data, &request); err != nil {
        return nil, fmt.Errorf("invalid YAML in %s: %w", path, err)
    }

    return &request, nil
}
```

## Integration with Tools

The persistence tools in `pkg/core/tools/persistence.go` use this package:

```go
type SaveRequestTool struct {
    zapDir string
}

func (t *SaveRequestTool) Execute(args string) (string, error) {
    var params struct {
        Name    string            `json:"name"`
        Method  string            `json:"method"`
        URL     string            `json:"url"`
        Headers map[string]string `json:"headers"`
        Body    string            `json:"body"`
    }
    json.Unmarshal([]byte(args), &params)

    request := &storage.Request{
        Name:    params.Name,
        Method:  params.Method,
        URL:     params.URL,
        Headers: params.Headers,
        Body:    params.Body,
    }

    path := filepath.Join(t.zapDir, "requests", params.Name+".yaml")
    if err := storage.SaveRequest(path, request); err != nil {
        return "", err
    }

    return fmt.Sprintf("Saved request to %s", path), nil
}
```

## Testing

```bash
go test ./pkg/storage/...
```

Test substitution:

```go
func TestSubstituteVariables(t *testing.T) {
    vars := map[string]string{
        "BASE_URL": "http://localhost",
        "PORT":     "8080",
    }

    tests := []struct {
        input    string
        expected string
    }{
        {"{{BASE_URL}}", "http://localhost"},
        {"{{BASE_URL}}:{{PORT}}", "http://localhost:8080"},
        {"{{UNDEFINED}}", "{{UNDEFINED}}"},
        {"no vars here", "no vars here"},
    }

    for _, tt := range tests {
        result := storage.SubstituteVariables(tt.input, vars)
        if result != tt.expected {
            t.Errorf("SubstituteVariables(%q) = %q, want %q",
                tt.input, result, tt.expected)
        }
    }
}
```
