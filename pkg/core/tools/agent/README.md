# Agent Core & Orchestration (`pkg/core/tools/agent`)

This directory contains the high-level tools that govern Falcon's "brain" and autonomous testing workflows.

## Primary Tools (5)

### `memory`
Recall and save project-specific knowledge (base URLs, auth patterns, API schemas) across sessions.
- `action: recall` — Load facts from `.falcon/memory.json` and `falcon.md`
- `action: save` — Save a single fact to `.falcon/memory.json`
- `action: update_knowledge` — Update the `.falcon/falcon.md` knowledge base (auto-validated)
- `action: forget` — Remove a fact
- `action: list` — View all facts

**Mandatory on session start:**
```
ACTION: memory({"action":"recall"})
```

### `run_tests`
Execute test scenarios from the spec (merged: previously ran_tests + run_single_test).
- Run all scenarios: `run_tests({"scenarios": [...], "base_url": "..."})`
- Run single scenario: `run_tests({"scenarios": [...], "base_url": "...", "scenario": "test_name"})`

Writes validated report to `.falcon/reports/unit_report_<timestamp>.md` (or other type prefix).

### `auto_test`
The autonomous testing engine. Orchestrates: analyze spec → generate test scenarios → run tests → fix failures.
- Input: endpoint URL
- Output: Test scenarios, report, proposed fixes

### `test_suite`
Bundle multiple flows into a named, reusable test suite.
- `test_suite({"name": "user_auth_flow", "tests": [...]})` → saves to `.falcon/flows/`

### Orchestration & Sessions

**Session management** is part of the mandatory workflow:
- `session_log({"action": "start"})` at conversation start
- `session_log({"action": "end", "summary": "..."})` before final answer

Session records persist to `.falcon/sessions/` for audit trail and review.

## Reports

All test runners automatically write validated reports to `.falcon/reports/`:
- Unit tests → `unit_report_<timestamp>.md`
- Functional tests → `functional_report_<timestamp>.md`
- Integration tests → `integration_report_<timestamp>.md`
- Security scans → `security_report_<timestamp>.md`
- Performance tests → `performance_report_<timestamp>.md`

Validators (`ValidateReportContent`, `ValidateFalconMD`) ensure reports have meaningful content before returning success.

## Usage

These tools are invoked by Falcon during autonomous cycles, or by the user when requesting test execution, knowledge updates, or multi-step workflows.

## Example Prompts

Trigger these tools by asking:
- "Run a full autonomous test on the user service."
- "Remember that the base URL for the staging environment is `https://api.staging.example.com`."
- "Recall what we know about the auth endpoints."
- "Update the knowledge base with what we learned about the payment API."
- "Run the functional tests for the users endpoint."
- "Create a test suite that chains login → create user → delete user."
