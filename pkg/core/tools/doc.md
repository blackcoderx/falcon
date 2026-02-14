# ZAP Toolset

ZAP (Zero-configuration API Pentester/Prober) features a modular, tier-based tool system.

## Tool Categories

### 🛠️ Tier 1: Foundation (`shared/`)
Base capabilities for API interaction and validation.
- **HTTP**: `http_request` — High-performance HTTP client.
- **Validation**: `assert_response`, `validate_json_schema`, `compare_responses`.
- **Extraction**: `extract_value`.
- **Utilities**: `wait`, `retry`.
- **Auth**: `auth_bearer`, `auth_basic`, `auth_oauth2`, `auth_helper`.
- **Orchestration Lite**: `test_suite`.

### 🔍 Tier 2: Codebase & Persistence
Tools for local environment interaction and state management.
- **Debugging (`debugging/`)**: `read_file`, `write_file`, `list_files`, `search_code`, `find_handler`, `analyze_endpoint`, `analyze_failure`, `propose_fix`, `create_test_file`.
- **Persistence (`persistence/`)**: `variable`, `save_request`, `load_request`, `list_requests`, `set_environment`, `list_environments`.
- **Agent Lifecycle (`agent/`)**: `memory`, `export_results`, `run_tests`, `run_single_test`, `auto_test`.

### 🏗️ Tier 3: API Intelligence (`spec_ingester/`)
Transformation of API specifications into the ZAP Knowledge Graph.
- **Ingestion**: `ingest_spec`.

### 🚀 Tier 4: Autonomous Modules
High-level capability modules for advanced testing scenarios.
- **Test Generation**: `generate_functional_tests`.
- **Security**: `scan_security`.
- **Performance**: `run_performance`.
- **Quality Assurance**: 
  - `run_smoke`: Fast health checks.
  - `verify_idempotency`: Detect side effects.
  - `run_data_driven`: Bulk variable testing.
  - `verify_schema_conformance`: Spec-to-implementation validation.
- **Maintenance & Drift**:
  - `detect_breaking_changes`: Compare API versions.
  - `analyze_drift`: Detect shadow endpoints.
  - `validate_docs`: Verify README vs Implementation.
- **Orchestration & DevOps**:
  - `orchestrate_integration`: Multi-step workflow runner.
  - `check_regression`: Baseline vs Current comparison.
  - `map_dependencies`: Resource relationship mapping.
  - `scaffold_unit_tests`: Auto-generation of mocks/tests.

## Architecture
The system uses a `Registry` (`registry.go`) to initialize all tools with shared dependencies like `HTTPTool`, `VariableStore`, and `LLMClient`. This ensures consistent behavior and global state management across all modular capabilities.
