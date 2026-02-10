# Package tools

The `tools` package provides the agent capabilities for the ZAP API debugging assistant.
Each tool implements the `core.Tool` interface and handles a specific type of operation.

## Subpackages

| Package | Description |
|---------|-------------|
| [auth/](auth/doc.md) | Authentication tools (Bearer, Basic, OAuth2) |

## Tools by Category

### AI Analysis
| Tool | File | Description |
|------|------|-------------|
| `analyze_endpoint` | `analyze.go` | Deep analysis of endpoint structure and risks |
| `analyze_failure` | `analyze.go` | Expert assessment of test failures |

### Test Generation
| Tool | File | Description |
|------|------|-------------|
| `generate_tests` | `generate.go` | Create diverse, high-coverage scenarios |

### Orchestration
| Tool | File | Description |
|------|------|-------------|
| `run_tests` | `orchestrate.go` | Run multiple tests in parallel |
| `run_single_test` | `orchestrate.go` | Re-run a specific test scenario |
| `auto_test` | `orchestrate.go` | Full autonomous test-and-fix workflow |

### Codebase Intelligence
| Tool | File | Description |
|------|------|-------------|
| `read_file` | `file.go` | Read file contents |
| `write_file` | `write.go` | Write/update files |
| `list_files` | `file.go` | List directory contents |
| `search_code` | `search.go` | Search for patterns in code |
| `find_handler` | `handler.go` | Locate endpoint handlers in code |
| `propose_fix` | `fix.go` | Generate secure code changes (diff) |
| `create_test_file` | `test_gen.go` | Create regression tests for fixes |

### Reporting
| Tool | File | Description |
|------|------|-------------|
| `security_report` | `report.go` | Generate comprehensive analysis |
| `export_results` | `report.go` | Export findings to JSON/Markdown |

### Authentication
| Tool | File | Description |
|------|------|-------------|
| `auth_bearer` | `auth/bearer.go` | Bearer token creation |
| `auth_basic` | `auth/basic.go` | Basic auth creation |
| `auth_oauth2` | `auth/oauth2.go` | OAuth2 authentication flows |
| `auth_helper` | `auth/helper.go` | Token parsing and decoding |

### HTTP
| Tool | File | Description |
|------|------|-------------|
| `http_request` | `http.go` | Make HTTP requests |

### Testing & Validation
| Tool | File | Description |
|------|------|-------------|
| `assert_response` | `assert.go` | Validate HTTP responses |
| `extract_value` | `extract.go` | Extract values from responses |
| `validate_json_schema` | `schema.go` | JSON Schema validation |
| `compare_responses` | `diff.go` | Compare response differences |
| `test_suite` | `suite.go` | Run test suites |

### Performance & Timing
| Tool | File | Description |
|------|------|-------------|
| `performance_test` | `perf.go` | Load testing |
| `wait` | `timing.go` | Add delays |
| `retry` | `timing.go` | Retry with backoff |

### Variables & Persistence
| Tool | File | Description |
|------|------|-------------|
| `variable` | `variables.go` | Session/global variables |
| `save_request` | `persistence.go` | Save API requests |
| `load_request` | `persistence.go` | Load saved requests |
| `list_requests` | `persistence.go` | List saved requests |
| `set_environment` | `persistence.go` | Switch environments |

### Webhooks
| Tool | File | Description |
|------|------|-------------|
| `webhook_listener` | `webhook.go` | Start webhook server |

### Memory
| Tool | File | Description |
|------|------|-------------|
| `memory` | `memory.go` | Agent memory operations |
