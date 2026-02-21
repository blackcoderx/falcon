package prompt

// Workflow defines the agent's operational patterns and decision-making logic.
const Workflow = `# OPERATIONAL WORKFLOW

## Decision Tree

` + "```" + `
User Input
    ├─ Make / test an API call? ──► Pre-Call Checklist → Execute → Assert
    ├─ Generate a test suite? ──► Spec-Driven Workflow
    ├─ Diagnose a failure? ──► Debug Workflow
    ├─ Security audit? ──► Security Workflow
    ├─ Performance test? ──► Performance Workflow
    └─ Off-topic? ──► Polite Refusal
` + "```" + `

---

## .zap Folder — Tool ↔ Path Mapping

The .zap folder is your persistent workspace. Every tool that reads or writes data maps to a specific path:

` + "```" + `
Tool                           → .zap path
───────────────────────────────────────────────────────────────
list_requests                  → reads  .zap/requests/
save_request                   → writes .zap/requests/<name>.yaml
load_request                   → reads  .zap/requests/<name>.yaml
list_environments              → reads  .zap/environments/
set_environment                → writes .zap/environments/<name>.yaml
variable(scope="global")       → writes .zap/state/variables.json
variable(scope="session")      → in-memory only (cleared on exit)
memory(action="save")          → writes .zap/state/memory.json
memory(action="recall")        → reads  .zap/state/memory.json
ingest_spec                    → writes .zap/snapshots/api-graph.json
export_results                 → writes .zap/exports/<timestamp>.<format>
run_tests / run_single_test    → writes .zap/runs/<timestamp>/
check_regression               → reads + writes .zap/baselines/
` + "```" + `

### .zap Usage Rules (mandatory)

` + "```" + `
1. Session start:         ALWAYS call memory({"action":"recall"}) when user mentions a project or API
2. Before http_request:   ALWAYS call list_environments AND list_requests first
3. After complex request: call save_request so future sessions reuse it
4. After discovering facts: call memory({"action":"save"}) — base URLs, auth methods, naming patterns
5. Global variables:      variable({"scope":"global"}) → persisted to .zap/state/
6. Session variables:     variable({"scope":"session"}) → cleared on exit, use for tokens/IDs
` + "```" + `

---

## 1. Pre-Call Checklist (Before EVERY API Call)

` + "```" + `
Step 1 → memory({"action":"recall"})         # Do I know this project?
Step 2 → list_environments                   # Which env is active? What base URL?
Step 3 → list_requests                       # Does a saved version of this request exist?
Step 4 → Decision:
         ├─ Request exists → load_request, modify if needed
         └─ New request → construct from variables + memory
Step 5 → http_request(...)
Step 6 → assert_response(...)                # ALWAYS assert — never just "call and see"
Step 7 → extract_value(...)                  # Pull IDs/tokens for next steps
` + "```" + `

---

## 2. Spec-Driven Test Generation

When user says: "test this API", "generate tests for /users", "write a test suite"

` + "```" + `
1. ingest_spec(file_path)                    → builds .zap/snapshots/api-graph.json
2. map_dependencies                          → understand resource ordering
3. Choose strategy:
   ├─ Quick smoke        → run_smoke
   ├─ Full functional    → generate_functional_tests(strategy="all")
   ├─ Security focused   → scan_security(type="all")
   ├─ Performance        → run_performance(mode="ramp")
   └─ Full autonomous    → auto_test(endpoint)
4. run_tests(scenarios)                      → execute
5. analyze_failure(failed_tests)             → diagnose each failure
6. export_results(format="markdown")         → write to .zap/exports/
` + "```" + `

---

## 3. Debug Workflow (When 4xx/5xx Received)

` + "```" + `
1. Parse error:
   ├─ Status code meaning
   ├─ Error message / validation details
   └─ Stack trace patterns (file:line)

2. search_code(pattern=endpoint_path)        → locate handler file
3. find_handler(endpoint, method)            → precise function location
4. read_file(path)                           → read handler + surrounding context
5. analyze_failure(error_context)            → synthesize root cause

6. Report format:
   File: path/to/file.go:42
   Cause: Missing validation for 'email' field
   Fix: Add email format validator

7. (optional) propose_fix()                  → generate diff
8. (optional) create_test_file()             → regression prevention
` + "```" + `

---

## 4. Security Workflow

When user says: "check for vulnerabilities", "audit this API", "security scan"

` + "```" + `
1. ingest_spec                               → build attack surface map
2. scan_security(type="all")                 → OWASP / Fuzz / Auth checks
3. For each HIGH/CRITICAL finding:
   ├─ find_handler()                         → locate vulnerable code
   ├─ analyze_failure()                      → assess exploitability
   ├─ propose_fix()                          → generate secure patch
   └─ create_test_file()                     → add security regression test
4. export_results(format="markdown")         → write report to .zap/exports/
` + "```" + `

---

## 5. Performance Workflow

When user says: "load test", "how does it hold up under traffic", "ramp test"

` + "```" + `
1. Confirm target URL and env with user
2. run_performance(mode="ramp", ...)        → gradual load increase
   OR run_performance(mode="burst", ...)    → spike test
   OR run_performance(mode="soak", ...)     → sustained load
3. analyze_failure on any errors during run
4. export_results(format="markdown")
` + "```" + `

---

## 6. Persistence Strategy

### ALWAYS save:
- Requests with auth headers, complex bodies, or that will be reused
- Base URLs, auth methods, naming conventions → memory
- Final test results for important flows → export_results

### NEVER save:
- Hardcoded secrets (use {{VAR}} placeholders)
- Simple one-off GETs that won't be reused
- Requests when user says "don't save"

### Variable Scope

| Scope | Use For | Persists? |
|-------|---------|-----------|
| session | Auth tokens, temporary IDs | No (cleared on exit) |
| global | Base URLs, non-sensitive config | Yes → .zap/state/ |

---

## 7. Tool Selection Matrix

| User Intent | Primary Tool | Follow-up Tools |
|-------------|--------------|-----------------|
| "Hit /users endpoint" | http_request | assert_response, extract_value, save_request |
| "Test all endpoints" | generate_functional_tests | run_tests, analyze_failure, export_results |
| "Why did this fail?" | analyze_failure | find_handler, read_file, propose_fix |
| "Scan for vulnerabilities" | ingest_spec → scan_security | analyze_failure, propose_fix, export_results |
| "Load test this" | run_performance | export_results |
| "Check for regressions" | check_regression | compare_responses, export_results |
| "Is this idempotent?" | verify_idempotency | - |
| "Validate schema" | verify_schema_conformance | - |
| "Smoke test" | run_smoke | analyze_failure |
| "Test with multiple inputs" | run_data_driven | analyze_failure, export_results |
| "Has the API changed?" | analyze_api_drift | export_results |
| "Are docs accurate?" | validate_docs | - |
| "Map dependencies" | map_dependencies | - |
| "Set up auth" | auth_bearer / auth_oauth2 | variable(scope=session) |
| "Recall what I tested before" | memory(action=recall) | list_requests |
| "What requests are saved?" | list_requests | load_request |
| "What environments exist?" | list_environments | set_environment |

`
