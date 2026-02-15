# ZAP Tools Overview & Gap Analysis

This document provides a detailed mapping of the current tool implementation in `pkg/core/tools` to the 17-module blueprint defined in `project/refined.md`.

## Summary of Alignment
- **Implemented Modules**: 12
- **Partially Implemented**: 1 (Exploratory Test Assistant)
- **Missing Modules**: 4 (Resilience Simulator, Compatibility Checker, Compliance Auditor, Version Test Manager)

---

## Tool Mapping & Detailed Descriptions

### Module 1: Spec Ingester 📋
**Tool**: `ingest_spec` (`pkg/core/tools/spec_ingester`)
- **Action**: `index`, `update`, `status`
- **Capability**: Ingests OpenAPI/Swagger/Postman specs into an internal API Knowledge Graph.
- **Blueprint Coverage**: Full.

### Module 2: Functional Test Generator 🧪
**Tool**: `generate_functional_tests` (`pkg/core/tools/functional_test_generator`)
- **Strategies**: `happy`, `negative`, `boundary`
- **Capability**: Auto-generates and optionally executes tests for every endpoint identified in the Knowledge Graph.
- **Blueprint Coverage**: Full.

### Module 3: Unit Test Scaffolder 🔬
**Tool**: `scaffold_unit_tests` (`pkg/core/tools/unit_test_scaffolder`)
- **Capability**: Scans codebase (controllers, services) and generates test skeletons with mocks for databases and external APIs.
- **Blueprint Coverage**: Full.

### Module 4: Integration Test Orchestrator 🔗
**Tool**: `orchestrate_integration` (`pkg/core/tools/integration_orchestrator`)
- **Note**: This is the tool referred to as "Ingestion Test Orchestrator" in the user request.
- **Capability**: Manages multi-step workflows (e.g., Create -> Login -> Order). Handles state sharing between steps and manages isolated test environments.
- **Blueprint Coverage**: Full.

### Module 5: Performance Engine ⚡
**Tool**: `run_performance` (`pkg/core/tools/performance_engine`)
- **Modes**: `load`, `stress`, `spike`, `soak`
- **Capability**: Ramps up concurrent users, measures latency p50/p95/p99, and flags SLA violations.
- **Blueprint Coverage**: Full.

### Module 6: Security Scanner 🔒
**Tool**: `scan_security` (`pkg/core/tools/security_scanner`)
- **Types**: `owasp`, `fuzz`, `auth`
- **Capability**: Performs OWASP Top 10 checks, fuzzes inputs for injections, and audits authentication/authorization logic.
- **Blueprint Coverage**: Full.

### Module 7: Contract Guardian 📜
**Tool**: `detect_breaking_changes` (`pkg/core/tools/breaking_change_detector`)
**Tool**: `verify_schema_conformance` (`pkg/core/tools/schema_conformance`)
- **Capability**: Detects differences between API versions (removed fields, type changes) and ensures live responses match their schema definitions.
- **Blueprint Coverage**: Full.

### Module 8: Regression Watchdog 🔄
**Tool**: `check_regression` (`pkg/core/tools/regression_watchdog`)
- **Capability**: Compares current behavior against a saved baseline snapshot. Detects drift in status codes, body structure, and timing.
- **Blueprint Coverage**: Full.

### Module 10: Resilience Simulator 💥
**STATUS: MISSING**
- **Required Capability**: Dependency failure simulation, network chaos, circuit breaker validation.
- **Work Left**: Implement chaos testing logic and mock failure injectors.

### Module 11: Compatibility Checker 🔄
**STATUS: MISSING**
- **Required Capability**: Test across multiple runtime versions and validate content negotiation (JSON/XML).
- **Work Left**: Implement cross-runtime environment management.

### Module 12: Compliance Auditor ✅
**STATUS: MISSING**
- **Required Capability**: PII scanning, GDPR/HIPAA checks, audit log validation.
- **Work Left**: Implement sensitive data detectors and compliance rule sets.

### Module 13: Documentation Validator 📖
**Tool**: `validate_docs` (`pkg/core/tools/documentation_validator`)
- **Capability**: Compares READMEs and wikis against the Knowledge Graph to identify ghost or undocumented endpoints.
- **Blueprint Coverage**: Full.

### Module 14: Exploratory Test Assistant 🔍
**STATUS: PARTIAL**
- **Current Tools**: `analyze_endpoint`, `analyze_failure` (in `pkg/core/tools/debugging`)
- **Required Capability**: Interactive assistant suggesting "what if" scenarios and converting manual steps to automation.
- **Work Left**: Formalize into a unified interactive assistant tool.

### Module 15: Idempotency Verifier 🔁
**Tool**: `verify_idempotency` (`pkg/core/tools/idempotency_verifier`)
- **Capability**: Repeats requests to endpoints that should be idempotent (POST orders, PUT updates) and verifies no unintended side effects occur.
- **Blueprint Coverage**: Full.

### Module 16: Data-Driven Test Engine 📊
**Tool**: `run_data_driven` (`pkg/core/tools/data_driven_engine`)
- **Capability**: Injects data from CSV/JSON into test templates; supports large-scale boundary and equivalence partition testing.
- **Blueprint Coverage**: Full.

### Module 17: Version Test Manager 🏷️
**STATUS: MISSING**
- **Required Capability**: Maintain suites per version, validate deprecation headers, and test version routing.
- **Work Left**: Implement version-aware test suite management.

---

## Auxiliary Tools
These tools support the primary modules but don't map 1:1 to a named module:
- **Dependency Mapper (`map_dependencies`)**: Builds resource relationship maps for the Knowledge Graph.
- **API Drift Analyzer (`analyze_drift`)**: Detects "shadow endpoints" during development.
- **Common Shared Tools (`shared/`)**: HTTP Client, Assertion logic, Template engine.
