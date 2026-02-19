# ZAP (Rebranding to Falcon Soon) 

> AI-powered API testing that understands your codebase

**ZAP** is a terminal-based AI assistant that doesn't just test your APIs—it debugs them. When an endpoint returns an error, ZAP searches your actual code to find the cause and suggests fixes. Works with local LLMs (Ollama) or cloud providers (Gemini).

![A picture of the TUI of ZAP](zap-interface.png)

## Installation

### Manual Installation

Download the latest pre-built binary for your operating system from [Releases](https://github.com/blackcoderx/zap/releases).

**Windows:**
1. Download `zap_Windows_x86_64.zip`.
2. Extract the archive.
3. Add the extracted folder to your system `PATH`.

**macOS/Linux:**
1. Download the `tar.gz` archive for your architecture.
2. Extract the archive: `tar -xzf zap_...tar.gz`
3. Move the binary to a location in your `PATH` (e.g., `/usr/local/bin`).

### From Source

```bash
go install github.com/blackcoderx/zap/cmd/zap@latest
```

## Updating ZAP

ZAP includes a self-update command to easily upgrade to the latest version:

```bash
zap update
```

This will check for the latest release on GitHub and update your binary in place (requires write permissions to the binary location).

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Features](#features)
- [Architecture](#architecture)
- [Configuration](#configuration)
- [Usage](#usage)
- [Available Tools](#available-tools)
- [Contributing](#contributing)
- [License](#license)

## Quick Start

### Prerequisites

- Go 1.25.3 or higher
- [Ollama](https://ollama.ai/) for local AI (or Gemini API key for cloud)

### Build and Run

```bash
git clone https://github.com/blackcoderx/zap.git
cd zap
go build -o zap.exe ./cmd/zap
./zap
```

### First Run

1. ZAP creates a `.zap/` folder with config, history, and memory
2. Select your LLM provider (Ollama local, Ollama cloud, or Gemini)
3. Choose your API framework (gin, fastapi, express, etc.)
4. The interactive TUI launches

### Try It

```bash
# In the TUI, type natural language commands:
> GET http://localhost:8000/api/users

# ZAP makes the request, shows the response, and if there's an error,
# searches your code to find and explain the cause
```

## Features

### Codebase-Aware Debugging

ZAP doesn't just show you errors—it explains them:

- **Stack trace parsing** - Extracts file:line from Python, Go, and JavaScript tracebacks
- **Autonomous Testing** - One-click `auto_test` workflow: Analyze -> Generate -> Execute -> Diagnose
- **Intelligent Fixes** - Not just suggestions, but full unified diffs with `propose_fix`
- **Regression Testing** - Automatically generate test files to ensure bugs stay fixed
- **Framework patterns** - Detects endpoint handlers using framework-specific idioms (15+ frameworks)

### 33+ Advanced Tools for API Testing

ZAP follows a tiered architecture, organizing 33+ tools into logical domains:

| Tier | Category | Purpose | Included Tools |
|------|----------|---------|----------------|
| **1** | **Foundation** | Core HTTP, Auth, and Assertion primitives | `http_request`, `assert_response`, `validate_json_schema`, `extract_value`, `auth_bearer`, `auth_basic`, `auth_oauth2` |
| **2** | **Codebase** | Deep codebase analysis and local fixing | `read_file`, `write_file`, `search_code`, `find_handler`, `propose_fix`, `analyze_failure` |
| **3** | **Persistence** | State management and environments | `variable`, `save_request`, `load_request`, `set_environment`, `list_environments` |
| **4** | **Modules** | Autonomous testing and API Intelligence | `ingest_spec`, `generate_functional_tests`, `scan_security`, `run_performance`, `run_smoke` |

#### Detailed Tool Tiering:

- **Tier 1: Foundation (`shared/`)**
  - **HTTP & Auth**: `http_request`, `auth_bearer`, `auth_basic`, `auth_oauth2`, `auth_helper`.
  - **Validation**: `assert_response`, `validate_json_schema`, `compare_responses`, `extract_value`.
  - **Flow control**: `wait`, `retry`, `test_suite`.

- **Tier 2: Codebase & Persistence**
  - **Debugging (`debugging/`)**: `read_file`, `write_file`, `list_files`, `search_code`, `find_handler`, `analyze_endpoint`, `analyze_failure`, `propose_fix`, `create_test_file`.
  - **Persistence (`persistence/`)**: `variable`, `save_request`, `load_request`, `list_requests`, `set_environment`, `list_environments`.
  - **Agent Lifecycle (`agent/`)**: `memory`, `export_results`, `run_tests`, `run_single_test`, `auto_test`.

- **Tier 3: API Intelligence (`spec_ingester/`)**
  - **Ingestion**: `ingest_spec` — Transforms OpenAPI/Swagger into a Knowledge Graph.

- **Tier 4: Autonomous Modules**
  - **High-level QA**: `generate_functional_tests`, `run_smoke`, `verify_idempotency`, `run_data_driven`, `verify_schema_conformance`.
  - **Security & Performance**: `scan_security`, `run_performance`.
  - **Operations**: `orchestrate_integration`, `check_regression`, `map_dependencies`, `scaffold_unit_tests`.

### Beautiful Terminal Interface

Built with the [Charm](https://charm.sh/) ecosystem:

- **Streaming responses** - Text appears as the LLM generates it
- **Markdown rendering** - Responses are beautifully formatted with syntax highlighting
- **Input history** - Navigate with Shift+Up/Down
- **Clipboard support** - Copy responses with Ctrl+Y
- **Status line** - See what ZAP is doing (thinking, executing tool, streaming)

### Human-in-the-Loop Safety

When ZAP wants to modify a file:

1. Shows a colored diff of the proposed changes
2. Waits for your approval (Y/N)
3. Only writes the file if you confirm

No surprises, no unauthorized changes.

## Architecture

```
zap/
├── cmd/zap/              # Application entry point
├── pkg/
│   ├── core/             # Agent logic, ReAct loop
│   │   └── tools/        # 33 modular tools (4-tier system)
│   │       ├── shared/   # Tier 1: Foundation (HTTP, Auth, Assert)
│   │       ├── debugging/# Tier 2: Codebase intelligence
│   │       ├── persistence/# Tier 2: Session & Environment state
│   │       ├── agent/    # Tier 2: Agent lifecycle & memory
│   │       ├── spec_ingester/ # Tier 3: API Intelligence
│   │       └── ... (autonomous modules: security, performance, etc.)
│   ├── llm/              # LLM providers
│   ├── storage/          # Low-level I/O
│   └── tui/              # Terminal UI
├── .zap/                 # User config & memory (see folder structure below)
└── go.mod
```

### Core Components

| Component | Location | Purpose |
|-----------|----------|---------|
| **Agent** | `pkg/core/agent.go` | Tool registration, call counting, limit enforcement |
| **ReAct Loop** | `pkg/core/react.go` | Reason-Act-Observe loop for tool execution |
| **System Prompt** | `pkg/core/prompt.go` | 20-section LLM instructions |
| **Tools** | `pkg/core/tools/` | 28+ tool implementations |
| **LLM Clients** | `pkg/llm/` | Ollama and Gemini implementations |
| **TUI** | `pkg/tui/` | Bubble Tea-based terminal interface |
| **Storage** | `pkg/storage/` | YAML I/O, variable substitution |

### Message Flow

```
User Input → TUI (keys.go)
           → runAgentAsync() goroutine
           → Agent.ProcessMessageWithEvents()
           → LLM generates response
           → Parse for tool calls
           → Execute tool → Observe result → Loop or Final Answer
           → Events emitted to TUI
           → View rendered
```

See [pkg/core/README.md](pkg/core/README.md) for detailed architecture documentation.

## Configuration

### Setup Wizard

On first run, ZAP walks you through configuration:

```bash
./zap

# Step 1: Select LLM provider
# 1. Ollama (local)
# 2. Ollama (cloud)
# 3. Gemini

# Step 2: Select your API framework
# gin, echo, chi, fiber, fastapi, flask, django, express, nestjs, hono, spring, laravel, rails, actix, axum, other
```

### CLI Flags

```bash
# Skip wizard with flags
./zap --framework gin

# Execute saved request
./zap --request get-users --env prod
./zap -r get-users -e dev

# Show help
./zap --help
```

### .zap Folder Structure

The `.zap` directory serves as the brain, memory, and output center for the agent.

```
.zap/
├── baselines/          # "The Standard of Truth" - Reference snapshots
├── snapshots/          # "The Current Reality" - Knowledge Graph (api-graph.json)
├── requests/           # "Saved Actions" - Library of reusable requests
├── runs/               # "The History Book" - Immutable execution records
├── exports/            # "The Filing Cabinet" - Human-readable reports
├── logs/               # "The Diary" - Internal tool logs
├── state/              # "The Brain" - Agent memory and context
└── config/             # "The Settings" - Tool configuration (config.json)
```

**`.env`** - API keys (optional, at project root):

```env
OLLAMA_API_KEY=your_key_here
GEMINI_API_KEY=your_key_here
```

**`.zap/requests/`** - Saved requests with variable substitution:

```yaml
# .zap/requests/get-users.yaml
name: Get Users
method: GET
url: "{{BASE_URL}}/api/users"
headers:
  Authorization: "Bearer {{API_TOKEN}}"
```

**`.zap/environments/`** - Environment variables:

```yaml
# .zap/environments/dev.yaml
BASE_URL: http://localhost:3000
API_TOKEN: dev-token-123
```

### Tool Limits

Prevent runaway execution with per-tool and global limits:

| Setting | Default | Description |
|---------|---------|-------------|
| `default_limit` | 50 | Fallback for tools without specific limits |
| `total_limit` | 200 | Safety cap on total calls per session |
| `per_tool` | varies | Per-tool overrides by name |

## Usage

### Interactive Mode

```bash
./zap
```

#### Natural Language Commands

```bash
# Make HTTP requests
> GET http://localhost:8000/api/users
> POST /api/users with {"name": "John"}

# Save and load requests
> save this request as get-users
> load the get-users request
> list my saved requests

# Environment management
> switch to prod environment
> show available environments

# Code analysis
> search for the /users endpoint
> read the file api/handlers.go
> find where UserService is defined

# Testing
> validate the response matches this schema: {...}
> run a load test with 10 concurrent users
> compare this response to the baseline
```

#### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Enter` | Send message |
| `Shift+↑/↓` | Navigate input history |
| `PgUp/PgDown` | Scroll output |
| `Ctrl+L` | Clear screen |
| `Ctrl+U` | Clear input line |
| `Ctrl+Y` | Copy last response |
| `Esc` | Stop agent (running) / Quit (idle) |
| `Ctrl+C` | Quit |

#### File Write Confirmation

When ZAP wants to modify a file:

| Key | Action |
|-----|--------|
| `Y` | Approve change |
| `N` | Reject change |
| `PgUp/PgDown` | Scroll diff |
| `Esc` | Reject and continue |

### CLI Mode (Automation)

Perfect for CI/CD pipelines:

```bash
# Execute saved request with environment
./zap --request get-users --env prod

# Combine with framework setup
./zap --framework gin --request health-check
```

## Available Tools

### AI Analysis & Generation

| Tool | Description |
|------|-------------|
| `analyze_endpoint` | Deep analysis of endpoint structure, parameters, and security risks |
| `analyze_failure` | Expert assessment of why a test failed with remediation steps |
| `generate_tests` | AI-driven generation of comprehensive, categorized test scenarios |

### Orchestration & Automation

| Tool | Description |
|------|-------------|
| `run_tests` | Parallel execution of multiple scenarios with filtering |
| `run_single_test` | Trigger a specific scenario (useful for fix verification) |
| `auto_test` | Autonomous workflow: analyze → generate → execute → assess |

### Core API Tools

| Tool | Description |
|------|-------------|
| `http_request` | Make HTTP requests with status code meanings and error hints |
| `save_request` | Save API request to YAML with `{{VAR}}` placeholders |
| `load_request` | Load saved request with environment variable substitution |
| `list_requests` | List all saved requests in `.zap/requests/` |
| `set_environment` | Set active environment (dev, prod, staging) |
| `list_environments` | List available environments |

### Testing & Validation

| Tool | Description |
|------|-------------|
| `assert_response` | Validate status codes, headers, body, JSON path, timing |
| `extract_value` | Extract values using JSON path, headers, cookies, regex |
| `validate_json_schema` | Validate against JSON Schema (draft-07, draft-2020-12) |
| `test_suite` | Run organized test suites with assertions |
| `compare_responses` | Regression testing with baseline comparison |

### Variables & Timing

| Tool | Description |
|------|-------------|
| `variable` | Manage session/global variables with disk persistence |
| `wait` | Add delays for async operations |
| `retry` | Retry with configurable attempts and exponential backoff |

### Authentication

| Tool | Description |
|------|-------------|
| `auth_bearer` | Create Bearer token headers (JWT, API tokens) |
| `auth_basic` | Create HTTP Basic authentication headers |
| `auth_oauth2` | OAuth2 flows (client_credentials, password) |
| `auth_helper` | Parse JWT tokens, decode Basic auth |

### Performance & Webhooks

| Tool | Description |
|------|-------------|
| `performance_test` | Load test with concurrent users, p50/p95/p99 latency |
| `webhook_listener` | Temporary HTTP server to capture callbacks |

### Codebase Analysis

| Tool | Description |
|------|-------------|
| `read_file` | Read file contents (100KB security limit) |
| `write_file` | Write files with human-in-the-loop confirmation |
| `list_files` | List files with glob patterns (`**/*.go`) |
| `search_code` | Search patterns with ripgrep (native fallback) |
| `find_handler` | Framework-aware discovery of endpoint source code |
| `propose_fix` | Generate unified diffs to fix identified vulnerabilities |
| `create_test_file` | Generate regression tests to ensure bugs stay fixed |

### Reporting & Scoring

| Tool | Description |
|------|-------------|
| `security_report` | Comprehensive security assessment and action plan |
| `export_results` | Export results to JSON (CI/CD) or Markdown (Reporting) |

## Contributing

Contributions are welcome! See the package-level documentation for understanding the codebase:

- [pkg/core/README.md](pkg/core/README.md) - Agent and ReAct loop
- [pkg/core/tools/README.md](pkg/core/tools/README.md) - Tool implementation guide
- [pkg/llm/README.md](pkg/llm/README.md) - Adding new LLM providers
- [pkg/storage/README.md](pkg/storage/README.md) - Persistence layer
- [pkg/tui/README.md](pkg/tui/README.md) - Terminal UI

### Adding a New Tool

1. Create a new file in `pkg/core/tools/`
2. Implement the `core.Tool` interface:

```go
type Tool interface {
    Name() string
    Description() string
    Parameters() string  // JSON Schema
    Execute(args string) (string, error)
}
```

3. Register in `pkg/tui/init.go` via `agent.RegisterTool()`

### Development Guidelines

See [CLAUDE.md](CLAUDE.md) for detailed development guidelines.

## Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.25.3 |
| CLI Framework | Cobra + Viper |
| TUI | Bubble Tea, Lip Gloss, Bubbles, Glamour, Huh |
| LLM Providers | Ollama, Google Gemini |
| Search | ripgrep (with native Go fallback) |
| Data | YAML for requests/environments |
| Validation | gojsonschema |

## License

MIT

## Acknowledgments

Built with the amazing [Charm](https://charm.sh/) ecosystem.
