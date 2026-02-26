# pkg/storage

This package handles request and environment persistence for Falcon. It provides YAML-based file I/O and a `{{VAR}}` variable substitution engine used across all saved requests and environments.

## Package Overview

```
pkg/storage/
├── schema.go  # Data structures: Request, Environment, Collection
├── yaml.go    # YAML file read/write operations
└── env.go     # Variable substitution engine ({{VAR}} placeholders)
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

**Example YAML** (`.falcon/requests/get-users.yaml`):

```yaml
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

**Example YAML** (`.falcon/environments/dev.yaml`):

```yaml
# Development environment
BASE_URL: http://localhost:3000
API_KEY: your-dev-api-key
```

### Collection

A named group of related requests (reserved for future use):

```go
type Collection struct {
    Name        string    `yaml:"name"`
    Description string    `yaml:"description"`
    Requests    []Request `yaml:"requests"`
}
```

## YAML Operations (`yaml.go`)

### Requests

```go
// Save a request
err := storage.SaveRequest(".falcon/requests/get-users.yaml", request)

// Load a request
request, err := storage.LoadRequest(".falcon/requests/get-users.yaml")

// List all saved requests (returns names without extension)
names, err := storage.ListRequests(".falcon/requests/")
// Returns: []string{"get-users", "create-user", "health-check"}
```

### Environments

```go
// Save an environment
err := storage.SaveEnvironment(".falcon/environments/prod.yaml", env)

// Load an environment
env, err := storage.LoadEnvironment(".falcon/environments/prod.yaml")

// List all environments
names, err := storage.ListEnvironments(".falcon/environments/")
// Returns: []string{"dev", "staging", "prod"}
```

## Variable Substitution (`env.go`)

### Basic Substitution

Replace `{{VAR}}` placeholders with values from an environment:

```go
variables := map[string]string{
    "BASE_URL":  "http://localhost:3000",
    "API_TOKEN": "abc123",
}

result := storage.SubstituteVariables("{{BASE_URL}}/api/users", variables)
// Result: "http://localhost:3000/api/users"
```

### Full Request Substitution

Apply substitution to all fields of a request at once:

```go
substituted := storage.SubstituteRequest(request, variables)
// substituted.URL     → "http://localhost:3000/api/users"
// substituted.Headers → {"Authorization": "Bearer abc123"}
```

### Nested Variables

Variables can reference other variables and are resolved recursively:

```yaml
# Environment
BASE_URL: "http://localhost:3000"
API_URL:  "{{BASE_URL}}/api"
USERS_URL: "{{API_URL}}/users"
```

```go
result := storage.SubstituteVariables("{{USERS_URL}}", variables)
// Result: "http://localhost:3000/api/users"
```

### Undefined Variables

Undefined placeholders are left unchanged:

```go
result := storage.SubstituteVariables("{{UNDEFINED}}/path", variables)
// Result: "{{UNDEFINED}}/path"
```

## File Layout

Falcon's storage lives entirely inside `.falcon/`:

```
.falcon/
├── config.yaml              # Main configuration (YAML)
├── memory.json              # Agent memory (JSON)
├── manifest.json            # Workspace counts (JSON)
├── requests/                # Saved API requests
│   ├── get-users.yaml
│   ├── create-user.yaml
│   └── health-check.yaml
├── environments/            # Environment variable files
│   ├── dev.yaml
│   ├── staging.yaml
│   └── prod.yaml
├── baselines/               # Regression test snapshots
└── flows/                   # Multi-step API flows
```

## Usage Patterns

### Save → Switch Env → Load → Execute

```go
// 1. Save a request template
storage.SaveRequest(".falcon/requests/login.yaml", &storage.Request{
    Name:   "Login",
    Method: "POST",
    URL:    "{{BASE_URL}}/auth/login",
    Headers: map[string]string{"Content-Type": "application/json"},
    Body:   `{"username":"{{TEST_USER}}","password":"{{TEST_PASS}}"}`,
})

// 2. Load environment
env, _ := storage.LoadEnvironment(".falcon/environments/dev.yaml")

// 3. Load and substitute request
req, _ := storage.LoadRequest(".falcon/requests/login.yaml")
ready := storage.SubstituteRequest(req, env.Variables)

// 4. ready.URL, ready.Headers, ready.Body are all substituted
```

### Multi-Environment Request Templates

Define the request once, use different credentials per environment:

```yaml
# .falcon/environments/dev.yaml
BASE_URL: http://localhost:3000
TEST_USER: devuser
TEST_PASS: devpass

# .falcon/environments/staging.yaml
BASE_URL: https://staging.example.com
TEST_USER: stageuser
TEST_PASS: stagepass
```

## Error Handling

All functions return descriptive errors:

```go
req, err := storage.LoadRequest(".falcon/requests/missing.yaml")
// err: "request file not found: .falcon/requests/missing.yaml"

env, err := storage.LoadEnvironment(".falcon/environments/bad.yaml")
// err: "invalid YAML in .falcon/environments/bad.yaml: ..."
```

## Testing

```bash
go test ./pkg/storage/...
```

Example:

```go
func TestSubstituteVariables(t *testing.T) {
    vars := map[string]string{
        "BASE_URL": "http://localhost",
        "PORT":     "8080",
    }

    tests := []struct{ input, want string }{
        {"{{BASE_URL}}", "http://localhost"},
        {"{{BASE_URL}}:{{PORT}}", "http://localhost:8080"},
        {"{{UNDEFINED}}", "{{UNDEFINED}}"},
        {"no vars", "no vars"},
    }

    for _, tt := range tests {
        got := storage.SubstituteVariables(tt.input, vars)
        if got != tt.want {
            t.Errorf("got %q, want %q", got, tt.want)
        }
    }
}
```
