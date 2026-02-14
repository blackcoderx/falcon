# Falcon — Tool Restructuring Plan (Final)

> Mapping the current 36 flat tools → a modular 17-module architecture.

---

## 🤖 AI Agent Implementation Prompt

**You are an AI agent tasked with restructuring the Falcon (ZAP) tool architecture according to this plan. Follow these instructions precisely:**

### Your Mission
Transform the current flat 36-tool structure in `pkg/core/tools/` into a modular architecture with:
- **16 kept agent-callable tools** organized into `debugging/`, `persistence/`, and `agent/` folders
- **13 shared internal helpers** in a flat `shared/` folder
- **17 new module tools** (6 enhanced from existing, 11 built from scratch)

### Critical Rules
1. **DO NOT delete any existing tool code** — every tool either moves to a kept folder, moves to `shared/`, or gets absorbed into a module
2. **Preserve all dependencies** — tools depend on `ResponseManager`, `VariableStore`, `ConfirmationManager`, `LLMClient`, `PersistenceTool`, `workDir`, `zapDir`. Wire these correctly.
3. **`pkg/storage/` is untouched** — it's an external dependency. Don't modify it.
4. **Each module's `tool.go` implements `core.Tool`** — (Name/Description/Parameters/Execute methods)
5. **Shared helpers are NOT tools** — they export Go functions/types, not `core.Tool` implementations
6. **Test after each phase** — run `go build ./cmd/zap` to verify no broken imports

### Implementation Steps

#### Phase 1: Foundation (Do this first)
1. **Create the new folder structure** (empty folders for now):
   ```
   pkg/core/tools/
   ├── shared/
   ├── debugging/
   ├── persistence/
   ├── agent/
   ├── spec_ingester/
   ├── functional_test_generator/
   ├── ... (all 17 module folders)
   ```

2. **Move infrastructure to `shared/`**:
   - `manager.go` → `shared/response_manager.go`
   - `confirm.go` → `shared/confirmation.go`
   - `pathutil.go` → `shared/pathutil.go`
   - `test_types.go` → `shared/types.go`

3. **Move internalized tools to `shared/`** (see Section 2 "REFACTOR" table):
   - `http.go` → `shared/http.go`
   - `auth/*.go` → `shared/auth.go` (combine all 4 auth files)
   - `assert.go` → `shared/assertions.go`
   - `extract.go` → `shared/extraction.go`
   - `schema.go` → `shared/schema_validator.go`
   - `diff.go` → `shared/diff.go`
   - `timing.go` → `shared/timing.go`
   - `webhook.go` → `shared/webhook.go`
   - `suite.go` → `shared/suite.go`

4. **Move kept tools to their folders** (see Section 2 "KEEP" table):
   - `debugging/`: analyze_failure.go, propose_fix.go, create_test_file.go, find_handler.go, read_file.go, write_file.go, list_files.go (from file.go), search_code.go
   - `persistence/`: save_request.go, load_request.go, list_requests.go, set_environment.go, list_environments.go (all from persistence.go), variables.go
   - `agent/`: memory.go, export_results.go (from report.go)

5. **Update all import paths** in moved files:
   - `import "github.com/blackcoderx/zap/pkg/core/tools"` → `import "github.com/blackcoderx/zap/pkg/core/tools/shared"`
   - Fix cross-references between files

6. **Create `registry.go`** at `pkg/core/tools/registry.go`:
   - Function: `RegisterAllTools(agent *core.Agent, llmClient llm.LLMClient, workDir, zapDir string, ...)`
   - Create shared dependencies (ResponseManager, VariableStore, etc.)
   - Register all kept tools from debugging/, persistence/, agent/
   - Register module tools (initially just placeholders)

7. **Update `pkg/tui/init.go`**:
   - Replace all individual `agent.RegisterTool(...)` calls with one call: `tools.RegisterAllTools(agent, llmClient, workDir, zapDir, ...)`

8. **Validation checkpoint**:
   - Run `go build ./cmd/zap`
   - All kept tools should still work
   - No broken imports

#### Phase 2: Build Core Modules (Priority order from Section 9)
For each module (start with `spec_ingester/`, then `functional_test_generator/`, etc.):

1. **Create `tool.go`** implementing `core.Tool`:
   ```go
   type SpecIngesterTool struct {
       // dependencies
   }
   func (t *SpecIngesterTool) Name() string { return "ingest_spec" }
   func (t *SpecIngesterTool) Description() string { ... }
   func (t *SpecIngesterTool) Parameters() string { ... }
   func (t *SpecIngesterTool) Execute(args string) (string, error) { ... }
   ```

2. **Create helper files** in the same folder (e.g., `parser.go`, `graph_builder.go`)

3. **Import shared helpers** as needed:
   ```go
   import "github.com/blackcoderx/zap/pkg/core/tools/shared"
   ```

4. **Register in `registry.go`**:
   ```go
   agent.RegisterTool(spec_ingester.NewSpecIngesterTool(...))
   ```

5. **Test the module**:
   - Build: `go build ./cmd/zap`
   - Run Falcon and try calling the new tool

#### Phase 3: Absorb Existing Tools into Modules
For modules that replace existing tools (e.g., `functional_test_generator` replaces `generate_tests`):

1. **Copy logic** from the old tool file (e.g., `generate.go`) into the module's helper files
2. **Adapt to use shared helpers** instead of direct tool calls
3. **DO NOT delete the old file yet** — keep it until the module is fully tested
4. **Once verified**, delete the old file and remove its registration

### Validation Checklist
After each phase, verify:
- [ ] `go build ./cmd/zap` succeeds
- [ ] `go test ./pkg/core/tools/...` passes
- [ ] All kept tools are still registered and callable
- [ ] No import cycles
- [ ] `pkg/storage/` is unchanged

### When You're Done
1. Update `pkg/core/tools/doc.md` to reflect the new structure
2. Update `README.md` tool list
3. Run full test suite: `go test ./...`
4. Create a summary of what was moved/created/deleted

### Questions to Ask the User
- **Before starting Phase 2**: "Which module should I build first? (Recommend: spec_ingester)"
- **If you find ambiguity**: "Tool X is used by both module A and B. Should it go to shared/ or stay in one module?"
- **After each module**: "Module Y is complete. Test it, or should I continue to the next one?"

---

## 1. Complete Inventory of Current Tools

Verified from **actual code** in `pkg/core/tools/`, `pkg/core/tools/auth/`, and **registration** in `pkg/tui/init.go`:

| # | Tool Name | Source File | Dependencies |
|---|-----------|-------------|--------------|
| 1 | `http_request` | `http.go` | ResponseManager, VariableStore |
| 2 | `read_file` | `file.go` | workDir |
| 3 | `write_file` | `write.go` | workDir, ConfirmationManager |
| 4 | `list_files` | `file.go` | workDir |
| 5 | `search_code` | `search.go` | workDir |
| 6 | `save_request` | `persistence.go` | PersistenceTool → pkg/storage |
| 7 | `load_request` | `persistence.go` | PersistenceTool → pkg/storage |
| 8 | `list_requests` | `persistence.go` | PersistenceTool → pkg/storage |
| 9 | `list_environments` | `persistence.go` | PersistenceTool → pkg/storage |
| 10 | `set_environment` | `persistence.go` | PersistenceTool → pkg/storage |
| 11 | `assert_response` | `assert.go` | ResponseManager |
| 12 | `extract_value` | `extract.go` | ResponseManager, VariableStore |
| 13 | `variable` | `variables.go` | VariableStore (disk-backed) |
| 14 | `wait` | `timing.go` | *(none)* |
| 15 | `retry` | `timing.go` | Agent (for re-executing tools) |
| 16 | `validate_json_schema` | `schema.go` | ResponseManager |
| 17 | `auth_bearer` | `auth/bearer.go` | VariableStore |
| 18 | `auth_basic` | `auth/basic.go` | VariableStore |
| 19 | `auth_oauth2` | `auth/oauth2.go` | VariableStore |
| 20 | `auth_helper` | `auth/helper.go` | ResponseManager, VariableStore |
| 21 | `test_suite` | `suite.go` | HTTPTool, AssertTool, ExtractTool, ResponseManager, VariableStore, zapDir |
| 22 | `compare_responses` | `diff.go` | ResponseManager, zapDir |
| 23 | `performance_test` | `perf.go` | HTTPTool, VariableStore |
| 24 | `webhook_listener` | `webhook.go` | VariableStore |
| 25 | `analyze_endpoint` | `analyze.go` | LLMClient |
| 26 | `analyze_failure` | `analyze.go` | LLMClient |
| 27 | `generate_tests` | `generate.go` | LLMClient |
| 28 | `run_tests` | `orchestrate.go` | HTTPTool, AssertTool, VariableStore |
| 29 | `run_single_test` | `orchestrate.go` | HTTPTool, AssertTool, VariableStore |
| 30 | `auto_test` | `orchestrate.go` | AnalyzeEndpointTool, GenerateTestsTool, RunTestsTool, AnalyzeFailureTool |
| 31 | `find_handler` | `handler.go` | workDir |
| 32 | `propose_fix` | `fix.go` | LLMClient |
| 33 | `create_test_file` | `test_gen.go` | LLMClient |
| 34 | `security_report` | `report.go` | zapDir |
| 35 | `export_results` | `report.go` | zapDir |
| 36 | `memory` | `memory.go` | MemoryStore |

### Internal Infrastructure (not tools — support code)

| File | Type | Purpose |
|------|------|---------|
| `manager.go` | `ResponseManager` | Thread-safe shared HTTP response state between tools |
| `confirm.go` | `ConfirmationManager` | Human-in-the-loop approval channel for file writes |
| `pathutil.go` | `ValidatePathWithinWorkDir()` | Path traversal prevention (security) |
| `test_types.go` | Structs | `TestScenario`, `TestResult`, `TestExpectation`, `StatusCodeRange` |

### External Dependency: `pkg/storage/`

| File | Purpose |
|------|---------|
| `schema.go` | `Request`, `Environment`, `Collection` structs |
| `yaml.go` | YAML save/load, request listing, directory helpers |
| `env.go` | `{{VAR}}` substitution engine, environment management |

> **`pkg/storage/` needs NO changes.** It's a clean YAML I/O + substitution layer that persistence tools import.

---

## 2. Tool Classification

### 🟢 KEEP as Agent-Callable Tools

These are **Falcon's differentiators** — the debug+fix loop and workspace management tools the agent calls directly. They stay as registered tools the LLM can invoke.

| # | Tool | Why Keep |
|---|------|----------|
| 1 | `analyze_failure` | Core debug — explains *why* a test failed with OWASP/CWE references |
| 2 | `propose_fix` | Generates unified diffs to fix code issues |
| 3 | `create_test_file` | Creates regression tests to ensure bugs stay fixed |
| 4 | `find_handler` | Locates endpoint handler source code (15+ framework support) |
| 5 | `read_file` | Context gathering for debugging |
| 6 | `write_file` | Applying fixes (with human-in-the-loop diff confirmation) |
| 7 | `list_files` | Codebase navigation with glob patterns |
| 8 | `search_code` | Pattern search (ripgrep with native fallback) |
| 9 | `save_request` | Persist requests as reusable YAML with `{{VAR}}` placeholders |
| 10 | `load_request` | Replay saved requests with environment variable substitution |
| 11 | `list_requests` | Browse saved requests |
| 12 | `set_environment` | Switch variable contexts (dev/staging/prod) |
| 13 | `list_environments` | Show available environments |
| 14 | `variable` | Session/global variable management (disk-backed) |
| 15 | `memory` | Agent memory across sessions |
| 16 | `export_results` | Export findings to JSON/Markdown for CI/CD |

### 🟡 REFACTOR → Become Internal Helpers (No Longer Agent-Callable)

These currently exist as standalone tools the LLM calls individually. In the new architecture, they become **internal helper functions** inside `shared/` that the high-level module tools use internally. The LLM no longer sees or calls them directly.

| Tool | New Location | Why Internalize |
|------|-------------|-----------------|
| `http_request` | `shared/http.go` | Every module needs HTTP — it's infrastructure, not a standalone action |
| `auth_bearer` | `shared/auth.go` | Auth is injected into requests by modules, not called standalone |
| `auth_basic` | `shared/auth.go` | Same |
| `auth_oauth2` | `shared/auth.go` | Same |
| `auth_helper` | `shared/auth.go` | Same |
| `assert_response` | `shared/assertions.go` | Modules run assertions internally, no need for LLM to call this |
| `extract_value` | `shared/extraction.go` | Modules extract values internally during workflows |
| `validate_json_schema` | `shared/schema_validator.go` | Used by modules for response validation |
| `compare_responses` | `shared/diff.go` | Absorbed into regression-watchdog module |
| `test_suite` | `shared/suite.go` | Used internally by functional-test-gen, smoke-runner, etc. |
| `wait` / `retry` | `shared/timing.go` | Timing is infrastructure used by orchestration modules |
| `webhook_listener` | `shared/webhook.go` | Used by integration-orchestrator and resilience-simulator |
| `analyze_endpoint` | Absorbed into `spec_ingester/` | Endpoint analysis becomes part of spec ingestion |
| `generate_tests` | Replaced by `functional_test_generator/` | Upgraded to spec-driven generation with strategies |
| `run_tests` | Internal to module orchestration | Modules run their own tests |
| `run_single_test` | Internal to module orchestration | Same |
| `auto_test` | Replaced by module-level orchestration | The whole module system replaces this |
| `performance_test` | Replaced by `performance_engine/` | Upgraded with stress/spike/soak modes |
| `security_report` | Replaced by `security_scanner/` | Upgraded to full OWASP scanner with reportin |

### 🔴 DISCARD — Nothing

After review, **no tools are deleted**. Every tool either stays as a kept agent-callable tool or becomes a shared internal helper.

---

## 3. New Modules (from `refined.md`)

Each module is **one agent-callable tool** backed by a folder with a main `tool.go` and helper files.

| # | Module | Agent Tool Name | Origin |
|---|--------|-----------------|--------|
| 1 | Spec Ingester | `ingest_spec` | 🆕 NEW — parses OpenAPI/Swagger, builds API Knowledge Graph |
| 2 | Functional Test Generator | `generate_functional_tests` | Replaces `generate_tests` — spec-driven with happy/negative/boundary strategies |
| 3 | Unit Test Scaffolder | `scaffold_unit_tests` | 🆕 NEW — scans codebase, generates unit test skeletons with mocks |
| 4 | Integration Test Orchestrator | `orchestrate_integration` | Replaces `auto_test` — multi-step workflow execution |
| 5 | Performance Engine | `run_performance` | Replaces `performance_test` — adds stress/spike/soak modes |
| 6 | Security Scanner | `scan_security` | Replaces `security_report` — full OWASP-mapped scanner |
| 7 | Contract Guardian | `check_contracts` | 🆕 NEW — provider/consumer contract testing |
| 8 | Regression Watchdog | `check_regression` | Absorbs `compare_responses` — baseline snapshots + diff engine |
| 9 | Smoke Test Runner | `run_smoke` | 🆕 NEW — fast post-deployment health checks |
| 10 | Resilience Simulator | `simulate_resilience` | 🆕 NEW — chaos engineering, dependency failure |
| 11 | Compatibility Checker | `check_compatibility` | 🆕 NEW — cross-version/platform testing |
| 12 | Compliance Auditor | `audit_compliance` | 🆕 NEW — PII/GDPR/HIPAA/PCI-DSS checks |
| 13 | Documentation Validator | `validate_docs` | 🆕 NEW — execute doc examples, find ghost endpoints |
| 14 | Exploratory Test Assistant | `explore_api` | 🆕 NEW — interactive "what-if" with session recording |
| 15 | Idempotency Verifier | `verify_idempotency` | 🆕 NEW — duplicate requests, race condition detection |
| 16 | Data-Driven Test Engine | `run_data_driven` | 🆕 NEW — template-based bulk testing with CSV/faker data |
| 17 | Version Test Manager | `manage_versions` | 🆕 NEW — per-version test suites, deprecation checks |

---

## 4. New Folder Structure

```
pkg/core/tools/
│
├── shared/                             # Internal helpers (NOT agent-callable)
│   ├── http.go                         # HTTPTool → internal HTTP client
│   ├── auth.go                         # Bearer + Basic + OAuth2 + Helper combined
│   ├── assertions.go                   # assert_response logic
│   ├── extraction.go                   # extract_value logic
│   ├── schema_validator.go             # JSON Schema validation
│   ├── diff.go                         # Response comparison engine
│   ├── timing.go                       # Wait + retry with backoff
│   ├── webhook.go                      # Temporary webhook listener
│   ├── suite.go                        # Test suite runner logic
│   ├── response_manager.go             # ← manager.go (shared HTTP response state)
│   ├── confirmation.go                 # ← confirm.go (human-in-the-loop)
│   ├── pathutil.go                     # ← pathutil.go (path traversal prevention)
│   └── types.go                        # ← test_types.go (TestScenario, TestResult, etc.)
│
├── debugging/                          # 🟢 KEPT — Code Fixing & Debug tools
│   ├── analyze_failure.go              # Tool: analyze_failure
│   ├── propose_fix.go                  # Tool: propose_fix
│   ├── create_test_file.go             # Tool: create_test_file
│   ├── find_handler.go                 # Tool: find_handler
│   ├── read_file.go                    # Tool: read_file
│   ├── write_file.go                   # Tool: write_file
│   ├── list_files.go                   # Tool: list_files
│   └── search_code.go                  # Tool: search_code
│
├── persistence/                        # 🟢 KEPT — Request & Environment management
│   ├── save_request.go                 # Tool: save_request
│   ├── load_request.go                 # Tool: load_request
│   ├── list_requests.go                # Tool: list_requests
│   ├── set_environment.go              # Tool: set_environment
│   ├── list_environments.go            # Tool: list_environments
│   └── variables.go                    # Tool: variable
│
├── agent/                              # 🟢 KEPT — Agent internals
│   ├── memory.go                       # Tool: memory
│   └── export_results.go              # Tool: export_results
│
├── spec_ingester/                      # Module 1 — Spec Ingester
│   ├── tool.go                         # Main: ingest_spec (Name, Desc, Execute)
│   ├── parser.go                       # OpenAPI / Swagger / Postman parsing
│   └── graph_builder.go               # API Knowledge Graph construction
│
├── functional_test_generator/          # Module 2 — Functional Test Generator
│   ├── tool.go                         # Main: generate_functional_tests
│   ├── generator.go                    # Test case creation logic
│   ├── strategies.go                   # Happy path / negative / boundary strategies
│   └── templates.go                    # Code templates for generated tests
│
├── unit_test_scaffolder/               # Module 3 — Unit Test Scaffolder
│   ├── tool.go                         # Main: scaffold_unit_tests
│   ├── scanner.go                      # Codebase analysis (controllers, services)
│   └── mock_generator.go              # Auto-mock generation
│
├── integration_orchestrator/           # Module 4 — Integration Test Orchestrator
│   ├── tool.go                         # Main: orchestrate_integration
│   ├── workflow.go                     # Multi-step workflow execution
│   └── environment.go                 # Test environment setup/teardown
│
├── performance_engine/                 # Module 5 — Performance Engine
│   ├── tool.go                         # Main: run_performance
│   ├── load_test.go                    # Load / Stress / Spike / Soak modes
│   └── metrics.go                     # p50/p95/p99, throughput, SLA comparison
│
├── security_scanner/                   # Module 6 — Security Scanner
│   ├── tool.go                         # Main: scan_security
│   ├── owasp_checks.go                # OWASP Top 10 mapped checks
│   ├── fuzzer.go                       # Input fuzzing engine
│   └── auth_audit.go                  # Auth/authz probing
│
├── contract_guardian/                  # Module 7 — Contract Guardian
│   ├── tool.go                         # Main: check_contracts
│   ├── registry.go                    # Contract registry management
│   └── checker.go                     # Breaking change detection
│
├── regression_watchdog/                # Module 8 — Regression Watchdog
│   ├── tool.go                         # Main: check_regression
│   ├── baseline.go                    # Baseline snapshot management
│   └── diff_engine.go                 # Response diffing (uses shared/diff.go)
│
├── smoke_runner/                       # Module 9 — Smoke Test Runner
│   ├── tool.go                         # Main: run_smoke
│   └── health_checks.go              # Fast health checks
│
├── resilience_simulator/               # Module 10 — Resilience Simulator
│   ├── tool.go                         # Main: simulate_resilience
│   ├── chaos.go                       # Dependency failure simulation
│   └── scorecard.go                   # Resilience scoring
│
├── compatibility_checker/              # Module 11 — Compatibility Checker
│   ├── tool.go                         # Main: check_compatibility
│   └── version_matrix.go             # Cross-version testing
│
├── compliance_auditor/                 # Module 12 — Compliance Auditor
│   ├── tool.go                         # Main: audit_compliance
│   ├── pii_scanner.go                 # PII detection
│   └── report_generator.go           # Compliance reports
│
├── doc_validator/                      # Module 13 — Documentation Validator
│   ├── tool.go                         # Main: validate_docs
│   ├── example_runner.go              # Execute documented examples
│   └── coverage.go                    # Doc completeness scoring
│
├── exploratory_assistant/              # Module 14 — Exploratory Test Assistant
│   ├── tool.go                         # Main: explore_api
│   └── suggestion_engine.go          # "What if" scenario suggestions
│
├── idempotency_verifier/               # Module 15 — Idempotency Verifier
│   ├── tool.go                         # Main: verify_idempotency
│   └── repeat_engine.go              # Duplicate request testing
│
├── data_driven_engine/                 # Module 16 — Data-Driven Test Engine
│   ├── tool.go                         # Main: run_data_driven
│   ├── template_engine.go            # Template variable filling
│   └── data_loader.go                # CSV/JSON data loading + faker
│
├── version_manager/                    # Module 17 — Version Test Manager
│   ├── tool.go                         # Main: manage_versions
│   └── version_router.go             # Multi-version test routing
│
├── registry.go                         # Central tool registration (replaces scattered init.go logic)
└── doc.md                              # Updated documentation
```

---

## 5. What the Agent Sees

### Before (Flat — 36 individual tools)
The LLM sees 36 tools and must figure out which combination to chain. Cognitive burden is high.

### After (Modular — ~33 tools)
The LLM sees:
- **16 kept tools** (debugging, persistence, agent) — the same tools they know
- **17 module tools** (one per module) — high-level actions
- `shared/` helpers are **invisible** to the LLM — they are internal implementation details

### Example Flow
```
User: "Test all endpoints in my API"

Agent calls: ingest_spec(spec_path="openapi.yaml")
  → Internally uses: shared/http, shared/auth
  → Returns: API Knowledge Graph (endpoints, methods, params, auth)

Agent calls: generate_functional_tests(strategies=["happy","negative","boundary"])
  → Internally uses: shared/http, shared/assertions, shared/extraction
  → Returns: Test results with pass/fail per endpoint

Agent calls: analyze_failure(test_result=..., response_body="...")  ← KEPT tool
  → Returns: Root cause with OWASP/CWE references

Agent calls: find_handler(endpoint="/api/users", method="POST")    ← KEPT tool
  → Returns: handlers/users.go:42

Agent calls: propose_fix(file="handlers/users.go", issue="missing input validation")  ← KEPT tool
  → Returns: Unified diff

Agent calls: write_file(path="handlers/users.go", content="...")   ← KEPT tool (with confirmation)
  → User approves → File updated
```

---

## 6. `pkg/storage/` — No Changes Needed

The `pkg/storage/` package stays as-is:
- `schema.go` — `Request`, `Environment`, `Collection` structs
- `yaml.go` — YAML save/load, `GetRequestsDir()`, `GetEnvironmentsDir()`
- `env.go` — `{{VAR}}` substitution engine, `ApplyEnvironment()`, `SubstituteVariables()`

The persistence tools in `persistence/` continue to import it. No coupling changes needed.

---

## 7. Dependency Wiring

The existing tools have concrete dependencies that must be preserved when restructuring. Here's how the shared dependencies get wired:

| Shared Dependency | What It Is | Who Creates It | Who Needs It |
|---|---|---|---|
| `ResponseManager` | Thread-safe last-HTTP-response store | `init.go` (one instance) | shared/http, shared/assertions, shared/extraction, shared/schema_validator, shared/diff |
| `VariableStore` | Session/global variable store (disk-backed) | `init.go` (one instance) | shared/http, shared/auth, shared/extraction, persistence/variables, performance_engine |
| `ConfirmationManager` | Human-in-the-loop approval channel | `init.go` (one instance) | debugging/write_file only |
| `LLMClient` | AI model access (Ollama/Gemini) | `init.go` via LLM provider | debugging/analyze_failure, debugging/propose_fix, debugging/create_test_file, spec_ingester, functional_test_generator |
| `PersistenceTool` | Request/env base directory manager | `init.go` (one instance) | persistence/* tools (imports pkg/storage) |
| `workDir` | Current working directory string | `init.go` | debugging/read_file, write_file, list_files, search_code, find_handler |
| `zapDir` | `.zap/` config directory string | `init.go` | persistence/*, agent/export_results, shared/suite, shared/diff |

### Registration Approach

Instead of the current `init.go` registering 36 individual tools, `registry.go` will:
1. Create all shared dependencies (ResponseManager, VariableStore, etc.)
2. Initialize shared helpers with those dependencies
3. Register kept tools (debugging, persistence, agent folders)
4. Register module tools (each module's `tool.go` receives the shared helpers it needs)

---

## 8. Summary Statistics

| Category | Count |
|----------|-------|
| **Current Tools (verified from init.go)** | 36 |
| **Kept as Agent-Callable** | 16 |
| **Internalized to shared/** | 19 (includes infrastructure) |
| **Discarded** | 0 |
| **New Module Tools** | 17 |
| **Total Agent Tools After** | ~33 (16 kept + 17 modules) |
| **Shared Helper Files** | 13 |
| **New Modules to Build** | 11 (6 modules absorb existing tools) |

---

## 9. Priority Order for Implementation

### Phase 1 — Foundation
1. **`shared/`** — Move all shared helpers first. Every module depends on these.
2. **`debugging/`** — Relocate kept debug/fix tools.
3. **`persistence/`** — Relocate kept persistence tools.
4. **`agent/`** — Relocate memory + export.
5. **`registry.go`** — New central registration replacing init.go wiring.
6. **`spec_ingester/`** — The new foundation. Everything reads from the API Knowledge Graph.

### Phase 2 — Core Testing (Existing tools, enhanced)
7. **`functional_test_generator/`** — Spec-driven test generation with strategies.
8. **`integration_orchestrator/`** — Multi-step workflow orchestration.
9. **`regression_watchdog/`** — Baseline snapshots + diff engine.
10. **`security_scanner/`** — Full OWASP-mapped scanner.
11. **`performance_engine/`** — Stress/spike/soak modes.

### Phase 3 — New Capabilities
12. **`smoke_runner/`** — Post-deployment health checks.
13. **`unit_test_scaffolder/`** — Codebase scanning + mock generation.
14. **`idempotency_verifier/`** — Duplicate request + race condition testing.
15. **`data_driven_engine/`** — Template-based bulk testing.
16. **`exploratory_assistant/`** — Interactive what-if mode.

### Phase 4 — Advanced Modules
17. **`contract_guardian/`** — Consumer/provider contract testing.
18. **`compliance_auditor/`** — Regulatory compliance checks.
19. **`doc_validator/`** — Documentation accuracy verification.
20. **`compatibility_checker/`** — Cross-version/platform testing.
21. **`version_manager/`** — Multi-version test management.

---

## 10. Key Design Decisions

1. **Each module has ONE main `tool.go`** — this implements `core.Tool` (Name/Description/Parameters/Execute). The main file orchestrates the helpers in its folder.

2. **`shared/` helpers are NOT agent-callable** — they don't implement `core.Tool`. They export Go functions/types that modules import and call internally.

3. **Kept tools stay agent-callable** — the 16 debugging, persistence, and agent tools remain registered with the agent and are directly invocable by the LLM.

4. **The 2+ rule for shared:** if a helper is used by only one module, it stays in that module's folder. If 2+ modules need it, it goes to `shared/`.

5. **`registry.go` replaces init.go tool wiring** — a single function `RegisterAllTools(agent, llmClient, workDir, zapDir, ...)` that creates dependencies and registers everything.

6. **`pkg/storage/` stays untouched** — it's a clean external package. The persistence tools continue to import it directly.

7. **Go package naming** — each folder under `tools/` uses underscores (e.g., `functional_test_generator`) to match Go conventions. The package name within each folder is the folder name.
