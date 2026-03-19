# Falcon Tools (`pkg/core/tools`)

This directory contains Falcon's tool set — modular, LLM-callable functions organized by domain and testing type. Tools are registered centrally via `registry.go` and exposed to the agent through the `core.Tool` interface.

## Architecture Overview

Three layers:

1. **The Registry** (`registry.go`) — instantiates all tools with their dependencies and registers them with the agent.
2. **Shared Libraries** (`shared/`) — HTTP client, auth, assertions, variable substitution, and `.falcon` artifact management used by most tools.
3. **Domain Modules** — standalone packages (`security_scanner/`, `performance_engine/`, etc.) that encapsulate specialized testing logic.

### Interaction Flow

```
Agent Core → Registry → Tool
                         ↓
                    Shared Libs
                         ↓
                   .falcon Artifacts
                         ↓
                     spec.yaml
```

---

## All Tools by Domain

### Core HTTP & Assertion Tools (`shared/`)

| Tool | Description |
|------|-------------|
| `http_request` | Make HTTP requests (GET/POST/PUT/DELETE) with headers, body, timeouts, and `{{VAR}}` substitution |
| `assert_response` | Validate status code, headers, body content, JSON paths, regex, response time |
| `extract_value` | Extract values from responses via JSONPath, headers, or cookies — store as session or global variables |
| `validate_json_schema` | Strict JSON Schema validation (draft-07, draft-2020-12) |
| `compare_responses` | Diff current vs. a previous response for regression detection |
| `auth` | Unified auth — Bearer tokens, Basic, OAuth2, API keys (replaces auth_bearer, auth_basic, auth_oauth2) |
| `wait` | Add delays for polling, backoff, or async operations |
| `retry` | Retry a failed tool call with exponential backoff |
| `webhook_listener` | Spawn a temporary HTTP server to capture incoming webhook callbacks |
| `test_suite` | Bundle multiple test flows into a named, reusable suite |

### Persistence Tools (`persistence/`, `shared/`, `agent/`)

| Tool | Description |
|------|-------------|
| `request` | Save/load/list/delete API requests as YAML in `.falcon/requests/` (replaces save_request, load_request, list_requests) |
| `environment` | Set/list environment variable files in `.falcon/environments/` (replaces set_environment, list_environments) |
| `variable` | Get/set variables scoped to the session or persisted to `variables.json` |
| `falcon_write` | Write validated YAML/JSON/Markdown files to `.falcon/` with path safety |
| `falcon_read` | Read artifacts from `.falcon/` (reports, flows, specs) |
| `memory` | Recall/save/update persistent knowledge across sessions (`~/.falcon/memory.json`) |
| `session_log` | Start/end session audit log, list/read past sessions in `.falcon/sessions/` |

### Debugging Tools (`debugging/`)

| Tool | Description |
|------|-------------|
| `find_handler` | Locates endpoint handlers via framework-specific patterns (Gin, Echo, FastAPI, Express); generic path fallback for other frameworks |
| `analyze_endpoint` | LLM-powered analysis of endpoint code structure, auth, and security risks |
| `analyze_failure` | LLM assessment of why a test failed, with root cause and remediation steps |
| `propose_fix` | Generate a unified diff patch to fix identified bugs or vulnerabilities |
| `read_file` | Read source file contents (100 KB limit) with line numbers |
| `list_files` | List source files filtered by extension (.go, .py, .js, .ts, .java, etc.) |
| `search_code` | Search codebase patterns with ripgrep (native Go fallback) |
| `write_file` | Write to source files — requires human-in-the-loop confirmation with diff preview |
| `create_test_file` | Auto-generate test cases for an endpoint |

### Testing Tools by Type

#### Smoke (`smoke_runner/`)
| Tool | Description |
|------|-------------|
| `run_smoke` | Hit all endpoints once to verify the API is up |

#### Functional (`functional_test_generator/`, `agent/`, `data_driven_engine/`)
| Tool | Description |
|------|-------------|
| `generate_functional_tests` | LLM-driven test generation with Happy/Negative/Boundary strategies |
| `run_tests` | Execute test scenarios in parallel; optional `scenario` param for a single test |
| `run_data_driven` | Bulk testing with CSV/JSON data sources via `{{VAR}}` template substitution |

#### Contract (`idempotency_verifier/`, `regression_watchdog/`)
| Tool | Description |
|------|-------------|
| `verify_idempotency` | Repeat requests and confirm they produce identical responses (no side effects) |
| `check_regression` | Compare current responses against baseline snapshots from `.falcon/baselines/` |

#### Performance (`performance_engine/`)
| Tool | Description |
|------|-------------|
| `run_performance` | *(Beta)* Load/stress/spike/soak test harness with p50/p95/p99 metrics — HTTP execution is currently mocked (see `gap.md`) |

#### Security (`security_scanner/`)
| Tool | Description |
|------|-------------|
| `scan_security` | OWASP vulnerability scanning, input fuzzing, auth bypass detection |

#### Integration (`integration_orchestrator/`)
| Tool | Description |
|------|-------------|
| `orchestrate_integration` | Chain multiple requests in a single transaction with resource linking and variable passing |

### Spec & Orchestration (`spec_ingester/`, `agent/`)

| Tool | Description |
|------|-------------|
| `ingest_spec` | Parse OpenAPI/Swagger or Postman specs into `.falcon/spec.yaml` and `.falcon/manifest.json` |
| `auto_test` | Autonomous loop: ingest spec → generate tests → run → analyze failures → fix |

---

## Module Organization

```
pkg/core/tools/
├── registry.go                    # Central registration of all tools
├── shared/                        # Foundation: HTTP, auth, assertions, artifacts
├── persistence/                   # Request, environment, variable storage
├── agent/                         # Memory, run_tests, auto_tester, session mgmt
├── debugging/                     # File I/O, code search, analysis, fix proposals
├── spec_ingester/                 # OpenAPI/Postman → spec.yaml + manifest.json
├── functional_test_generator/     # LLM-driven test scenario generation
├── data_driven_engine/            # CSV/JSON parameterized test execution
├── smoke_runner/                  # Minimal health check suite
├── idempotency_verifier/          # PUT/POST idempotency verification
├── regression_watchdog/           # Baseline snapshot comparisons
├── security_scanner/              # OWASP auditing + fuzzing
├── performance_engine/            # Load/stress/soak testing
└── integration_orchestrator/      # Multi-endpoint workflow execution
```

---

## .falcon Folder Structure (Flat)

All artifacts use a flat structure — no subdirectories in `reports/` or `flows/`. Filenames carry context via type prefix.

```
.falcon/
├── config.yaml
├── manifest.json
├── falcon.md                  # API knowledge base
├── spec.yaml                  # Ingested API spec
├── variables.json
├── sessions/
│   └── session_<timestamp>.json
├── environments/
│   ├── dev.yaml
│   └── prod.yaml
├── requests/
│   ├── get-users.yaml
│   └── create-user.yaml
├── baselines/
│   └── baseline_users_api.json
├── flows/
│   ├── unit_get_users.yaml
│   ├── integration_login_create_delete.yaml
│   └── security_auth_bypass.yaml
└── reports/
    ├── performance_report_users_api_20260227.md
    └── security_report_auth_api_20260227.md
```

**Naming conventions:**
- Reports: `<type>_report_<api-name>_<timestamp>.md`
- Flows: `<type>_<description>.yaml`

---

## Tool Safety & Validation

### Report Validation
After writing any report, `ValidateReportContent()` checks:
1. File exists and is > 64 bytes
2. Contains at least one Markdown heading (`# ` or `## `)
3. Contains at least one result indicator (table `|`, code block ` ``` `, or status keyword like `PASS`, `FAIL`, `✓`, `✗`)
4. Does not contain unresolved placeholders (`{{`, `TODO`, `[placeholder]`)

### falcon.md Validation
After `memory(action=update_knowledge)` writes to `falcon.md`, `ValidateFalconMD()` checks:
1. File exists and is > 200 bytes
2. Contains required sections: `# Base URLs`, `# Known Endpoints`
3. Each section has at least one non-blank line of content

### falcon_write Safety
The `falcon_write` tool enforces:
1. Blocks `../` (no directory traversal)
2. Blocks absolute paths
3. Blocks writes to protected files (`config.yaml`, `manifest.json`, `memory.json`)
4. Validates YAML/JSON syntax before writing

### write_file Confirmation
The `write_file` tool (source code writes) routes through `ConfirmationManager` — it emits a `confirmation_required` event with a unified diff before writing anything. The user must press Y to approve or N to reject.

---

## Extending Falcon

### Adding a New Tool

1. **Create a module**: `pkg/core/tools/<new_module>/tool.go`
2. **Implement the interface**:

```go
type MyTool struct{}

func NewMyTool() *MyTool { return &MyTool{} }

func (t *MyTool) Name() string        { return "my_tool" }
func (t *MyTool) Description() string { return "Does something useful" }
func (t *MyTool) Parameters() string {
    return `{
        "type": "object",
        "properties": {
            "input": {"type": "string", "description": "..."}
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
    // implementation
    return result, nil
}
```

3. **Register in `registry.go`**: add to the relevant `register*()` method
4. **Validate reports**: if your tool writes reports, call `ValidateReportContent()` after writing
5. **Update the system prompt**: if the tool adds a new capability, add it to `pkg/core/prompt/tools.go`

---

## Quick Reference: Which Tool for Which Task?

| Task | Tool |
|------|------|
| Make an HTTP request | `http_request` |
| Check response | `assert_response` |
| Extract a value for chaining | `extract_value` |
| Validate JSON Schema | `validate_json_schema` |
| Authenticate | `auth` |
| Check all endpoints are up | `run_smoke` |
| Generate test scenarios | `generate_functional_tests` |
| Run test scenarios | `run_tests` |
| Bulk test with data | `run_data_driven` |
| Load / stress test | `run_performance` *(mocked — see gap.md)* |
| Security audit | `scan_security` |
| Find endpoint handler in code | `find_handler` (specific patterns for Gin, Echo, FastAPI, Express; generic fallback for others) |
| Explain a test failure | `analyze_failure` |
| Generate a code fix | `propose_fix` |
| Read source file | `read_file` |
| Write source file | `write_file` |
| Search codebase | `search_code` |
| Save an API request | `request` (action=save) |
| Load a saved request | `request` (action=load) |
| Set environment variables | `environment` (action=set) |
| Set a session variable | `variable` |
| Save API knowledge | `memory` (action=save) |
| Recall API knowledge | `memory` (action=recall) |
| Write to .falcon/ | `falcon_write` |
| Read from .falcon/ | `falcon_read` |
| Log session start | `session_log` (action=start) |
| Chain multiple requests | `orchestrate_integration` |
| Detect regressions | `check_regression` |
| Ingest OpenAPI spec | `ingest_spec` |
| Full autonomous test run | `auto_test` |

---

## For More Details

- **Core Agent Logic**: See `pkg/core/README.md`
- **LLM Providers**: See `pkg/llm/README.md`
- **System Prompt**: `pkg/core/prompt/workflow.go` and `tools.go`
