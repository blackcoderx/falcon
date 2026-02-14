# Falcon — Tool Restructuring Sprints

> Step-by-step implementation of [refine-plan.md](file:///C:/Users/user/zap/project/refine-plan.md)

---

## Sprint 1: Shared Infrastructure

**Goal:** Move all internal infrastructure and internalized tools into `shared/`, creating the foundation everything else depends on.

### 1.1 — Create folder structure
- [x] Create all empty folders under `pkg/core/tools/`:
  - `shared/`, `debugging/`, `persistence/`, `agent/`
  - All 17 module folders: `spec_ingester/`, `functional_test_generator/`, `unit_test_scaffolder/`, `integration_orchestrator/`, `performance_engine/`, `security_scanner/`, `contract_guardian/`, `regression_watchdog/`, `smoke_runner/`, `resilience_simulator/`, `compatibility_checker/`, `compliance_auditor/`, `doc_validator/`, `exploratory_assistant/`, `idempotency_verifier/`, `data_driven_engine/`, `version_manager/`

### 1.2 — Move infrastructure files to `shared/`
- [x] `manager.go` → `shared/response_manager.go`
- [x] `confirm.go` → `shared/confirmation.go`
- [x] `pathutil.go` → `shared/pathutil.go`
- [x] `test_types.go` → `shared/types.go`
- [x] Update package declarations to `package shared`
- [x] Verify: `go build ./cmd/zap` (will break — expected, fix in 1.3)

### 1.3 — Move internalized tools to `shared/`
- [x] `http.go` → `shared/http.go` (strip `core.Tool` interface, keep `HTTPTool` struct + `Run()`)
- [x] `auth/bearer.go` + `auth/basic.go` + `auth/oauth2.go` + `auth/helper.go` → `shared/auth.go`
- [x] `assert.go` → `shared/assertions.go`
- [x] `extract.go` → `shared/extraction.go`
- [x] `schema.go` → `shared/schema_validator.go`
- [x] `diff.go` → `shared/diff.go`
- [x] `timing.go` → `shared/timing.go`l
- [x] `webhook.go` → `shared/webhook.go`
- [x] `suite.go` → `shared/suite.go`
- [x] Update all package declarations to `package shared`
- [x] Fix all internal cross-references within `shared/`

### 1.4 — Validation checkpoint
- [x] `go build ./pkg/core/tools/shared/...`
- [x] All exported types/functions accessible
- [x] No import cycles

**Deliverable:** A working `shared/` package that compiles independently.

---

## Sprint 2: Kept Tools Migration

**Goal:** Move the 16 kept agent-callable tools into `debugging/`, `persistence/`, and `agent/` folders.

### 2.1 — Move debugging tools
- [x] `analyze.go` (AnalyzeFailureTool only) → `debugging/analyze_failure.go`
- [x] `fix.go` → `debugging/propose_fix.go`
- [x] `test_gen.go` → `debugging/create_test_file.go`
- [x] `handler.go` → `debugging/find_handler.go`
- [x] `file.go` (ReadFileTool + ListFilesTool) → `debugging/read_file.go` + `debugging/list_files.go`
- [x] `write.go` → `debugging/write_file.go`
- [x] `search.go` → `debugging/search_code.go`
- [x] Update package to `package debugging`
- [x] Update imports to reference `shared/` for ResponseManager, VariableStore, etc.

### 2.2 — Move persistence tools
- [x] `persistence.go` → split into `persistence/save_request.go`, `persistence/load_request.go`, `persistence/list_requests.go`, `persistence/set_environment.go`, `persistence/list_environments.go`
- [x] `variables.go` → `persistence/variables.go`
- [x] Update package to `package persistence`
- [x] Keep `pkg/storage` imports as-is

### 2.3 — Move agent tools
- [x] `memory.go` → `agent/memory.go`
- [x] `report.go` (ExportResultsTool only) → `agent/export_results.go`
- [x] Update package to `package agent`

### 2.4 — Clean up old files
- [x] Remove all moved `.go` files from `pkg/core/tools/` root (keep only `registry.go` and `doc.md`)
- [x] Remove old `auth/` subdirectory (merged into `shared/auth.go`)

### 2.5 — Validation checkpoint
- [x] `go build ./pkg/core/tools/debugging/...`
- [x] `go build ./pkg/core/tools/persistence/...`
- [x] `go build ./pkg/core/tools/agent/...`
- [x] All 16 kept tools compile

**Deliverable:** All 16 kept tools compiled in their new packages.

---

## Sprint 3: Registry & Wiring — **COMPLETED** (Feb 14, 2026)

**Goal:** Create `registry.go` and update `init.go` so the app builds and runs end-to-end.

### 3.1 — Create `registry.go`
- [x] Create `pkg/core/tools/registry.go`
- [x] Function: `RegisterAllTools(agent, llmClient, workDir, zapDir, memStore, confirmManager)`
- [x] Inside: create shared dependencies (ResponseManager, VariableStore, HTTPTool, etc.)
- [x] Register all 16 kept tools from `debugging/`, `persistence/`, `agent/`
- [x] Integrate high-level orchestration tools (`run_tests`, `auto_test`)

### 3.2 — Update `pkg/tui/init.go`
- [x] Replace all individual `agent.RegisterTool(...)` calls with one call to `tools.RegisterAllTools(...)`
- [x] Pass through all necessary dependencies (ConfirmationManager, MemoryStore)

### 3.3 — Validation checkpoint
- [x] `go build ./cmd/zap` succeeds
- [x] `go build ./pkg/core/tools/...` succeeds (all sub-packages)
- [x] Verify full application build stability
- [x] Milestone: Application runs with modular toolset

**Deliverable:** App builds and runs with the new folder structure. All kept tools functional.


---

## Sprint 4: Spec Ingester Module — **COMPLETED** (Feb 14, 2026)

**Goal:** Build the foundation module. All subsequent modules will read from the API Knowledge Graph it produces.

### 4.1 — `spec_ingester/tool.go`
- [x] Implement `core.Tool` interface: `ingest_spec`
- [x] Accept params: `action` (index/status), `source` (path/URL)
- [x] Return: indexing summary or graph status

### 4.2 — `spec_ingester/parser.go`
- [x] OpenAPI 3.x parser (YAML/JSON) using `libopenapi`
- [x] Swagger 2.0 parser (OAS2) support
- [x] Postman Collection v2.1 parser using `go-postman-collection`
- [x] Auto-detect format from file contents

### 4.3 — `spec_ingester/graph_builder.go`
- [x] Build internal API Knowledge Graph struct:
  - Endpoints (path, method, params, response codes)
  - Persistence to `.zap/api_graph.json`
- [x] Enrichment from project context (framework info)

### 4.4 — Register & validate
- [x] Add to `registry.go`
- [x] Implement Automatic Initialization in `pkg/core/init.go`
- [x] Added `--no-index` opt-out flag in `cmd/zap/main.go`
- [x] Resolved circular dependencies by moving `secrets` and `manifest` to `shared/`
- [x] `go build ./cmd/zap` succeeds
- [x] Test: successful indexing of OpenAPI and Postman specs

**Deliverable:** `ingest_spec` tool registered and automatic first-run indexing flow functional.

---

## Sprint 5: Functional Test Generator Module — **COMPLETED** (Feb 14, 2026)

**Goal:** Replace `generate_tests` with a spec-driven test generator using happy/negative/boundary strategies.

### 5.1 — `functional_test_generator/tool.go`
- [x] Implement `core.Tool`: `generate_functional_tests`
- [x] Accept: Knowledge Graph data, strategy list, base_url
- [x] Orchestrate helpers to generate + execute tests

### 5.2 — `functional_test_generator/strategies.go`
- [x] Happy path strategy: valid data for every endpoint
- [x] Negative strategy: missing fields, wrong types, invalid values
- [x] Boundary strategy: min/max values, empty strings, large payloads

### 5.3 — `functional_test_generator/generator.go`
- [x] Generate `TestScenario` structs from the Knowledge Graph + selected strategies
- [x] Use `shared/http.go` to execute, `shared/assertions.go` to validate

### 5.4 — `functional_test_generator/templates.go`
- [x] Code templates for generated test output (if user wants exportable test files)

### 5.5 — Register & validate
- [x] Add to `registry.go`
- [x] `go build ./cmd/zap`
- [x] Test: ingest spec → generate tests → verify results

**Deliverable:** `generate_functional_tests` tool that creates and runs spec-based tests.

---

## Sprint 6: Integration Orchestrator + Regression Watchdog [COMPLETED]

**Goal:** Replace `auto_test`/`run_tests` and `compare_responses` with higher-level modules.

### 6.1 — `integration_orchestrator/`
- [x] `tool.go`: `orchestrate_integration` — multi-step workflow tool
- [x] `workflow.go`: sequential step execution (Create → Login → Order → Verify → Delete)
- [x] `environment.go`: test environment context management
- [x] Absorb `run_tests`, `run_single_test`, `auto_test` logic (Available via workflow steps)

### 6.2 — `regression_watchdog/`
- [x] `tool.go`: `check_regression` — baseline comparison tool
- [x] `baseline.go`: snapshot save/load to `.zap/baselines/`
- [x] `diff_engine.go`: compare current vs. baseline using `shared/diff.go`

### 6.3 — Register & validate
- [x] Add both to `registry.go`
- [x] `go build ./cmd/zap`
- [x] Test orchestration with a multi-step workflow
- [x] Test regression detection with a baseline

**Deliverable:** Two new module tools replacing 4 old tools with enhanced capabilities.

---

## Sprint 7: Security Scanner + Performance Engine [COMPLETED]

**Goal:** Upgrade security reporting to a full scanner and performance testing to multi-mode.

### 7.1 — `security_scanner/`
- [x] `tool.go`: `scan_security`
- [x] `owasp_checks.go`: OWASP Top 10 mapped checks
- [x] `fuzzer.go`: input fuzzing with injection payloads
- [x] `auth_audit.go`: auth/authz probing (expired tokens, cross-user, privilege escalation)
- [x] Absorb `security_report` logic into report generation

### 7.2 — `performance_engine/`
- [x] `tool.go`: `run_performance`
- [x] `load_runner.go`: 4 modes — load / stress / spike / soak
- [x] `metrics.go`: p50/p95/p99 calculation, SLA comparison, trend tracking
- [x] Absorb `performance_test` logic

### 7.3 — Register & validate
- [x] Add both to `registry.go`
- [x] `go build ./cmd/zap`
- [x] Test security scanner against a known-vulnerable endpoint
- [x] Test performance engine in each mode

**Deliverable:** Full OWASP security scanner + multi-mode performance engine.

---

## Sprint 8: New Capability Modules (Batch 1) [COMPLETED]

**Goal:** Build 4 new capability modules.

### 8.1 — `smoke_runner/`
- [x] `tool.go`: `run_smoke` — fast health check tool
- [x] `health_checks.go`: API reachability, auth system, DB connectivity, response format

### 8.2 — `unit_test_scaffolder/`
- [x] `tool.go`: `scaffold_unit_tests`
- [x] `scanner.go`: scan controllers/services/repos in codebase
- [x] `mock_generator.go`: auto-mock for DB calls, external APIs

### 8.3 — `idempotency_verifier/`
- [x] `tool.go`: `verify_idempotency`
- [x] `repeat_engine.go`: send duplicate requests, compare results, detect side effects

### 8.4 — `data_driven_engine/`
- [x] `tool.go`: `run_data_driven`
- [x] `template_engine.go`: parse templates with `{{placeholder}}` vars
- [x] `data_loader.go`: load CSV/JSON, generate fake data

### 8.5 — Register & validate
- [x] Add all 4 to `registry.go`
- [x] `go build ./cmd/zap`

**Deliverable:** 4 new module tools registered and callable.

---

## Sprint 9: New Capability Modules (Batch 2) [COMPLETED]

**Goal:** Build the remaining 5 advanced capability modules.

### 9.1 — `schema_conformance/`
- [x] `tool.go`: `verify_schema_conformance` — deep check against graph schemas

### 9.2 — `breaking_change_detector/`
- [x] `tool.go`: `detect_breaking_changes` — compare old vs new spec (OpenAPI diff)

### 9.3 — `dependency_mapper/`
- [x] `tool.go`: `map_dependencies` — map resource flow (e.g. POST creates ID for GET)

### 9.4 — `documentation_validator/`
- [x] `tool.go`: `validate_docs` — ensure README/Swagger matches actual graph

### 9.5 — `api_drift_analyzer/`
- [x] `tool.go`: `analyze_drift` — detect shadow/untracked endpoints

### 9.6 — Register & validate
- [x] Add all 5 to `registry.go`
- [x] `go build ./cmd/zap` — verify full system

**Deliverable:** Total 33 tools (24 base + 9 module) all operational.

---

## Sprint 10: Documentation & Final Validation [COMPLETED]

**Goal:** Update all docs, run full test suite, verify everything end-to-end.

### 10.1 — Update documentation
- [x] Rewrite `pkg/core/tools/doc.md` with new folder structure
- [x] Update `README.md` — tool list, architecture diagram, folder tree
- [x] Add doc.md files for each subfolder (debugging/, persistence/, agent/, shared/)

### 10.2 — Update / add tests
- [x] Update existing tests in `pkg/core/tools/` to reference new package paths
- [x] Add basic tests for each new module's `tool.go`

### 10.3 — Final validation
- [x] `go build ./cmd/zap`
- [x] `go test ./...` — all tests pass
- [x] `go vet ./...` — no warnings
- [x] Manual smoke test: run Falcon, call 3+ kept tools and 3+ module tools
- [x] Verify `pkg/storage/` is untouched (no changes in git diff)

**Deliverable:** Fully restructured, documented, and tested Falcon tool system.

---

## Summary

| Sprint | Focus | Tools Affected | Est. Effort |
|--------|-------|---------------|-------------|
| **1** | Shared infrastructure | 19 tools → `shared/` | **COMPLETED** |
| **2** | Kept tools migration | 16 tools → 3 folders | **COMPLETED** |
| **3** | Registry & wiring | App-level integration | **COMPLETED**|
| **4** | Spec Ingester | 1 new module | **COMPLETED** |
| **5** | Functional Test Generator | 1 new module (replaces `generate_tests`) | **COMPLETED** |
| **6** | Integration + Regression | 2 new modules (replace 4 old tools) | **COMPLETED** |
| **7** | Security + Performance | 2 new modules (replace 2 old tools) | **COMPLETED** |
| **8** | New capabilities (batch 1) | 4 new modules | **COMPLETED** |
| **9** | New capabilities (batch 2) | 5 new modules | **COMPLETED** |
| **10** | Docs & final validation | All | **COMPLETED** |
