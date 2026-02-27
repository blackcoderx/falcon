package prompt

import (
	"fmt"
	"strings"
)

// BuildToolsSection generates a context-efficient tool reference.
// Tools are grouped by the 8 API testing types plus support categories.
func BuildToolsSection(tools map[string]Tool) string {
	var sb strings.Builder

	sb.WriteString("# AVAILABLE TOOLS\n\n")
	sb.WriteString("**Call Format**: ACTION: tool_name({\"param\": \"value\"})\n\n")

	domains := map[string][]Tool{
		"Core":          {},
		"Persistence":   {},
		"Spec":          {},
		"Unit":          {},
		"Integration":   {},
		"Smoke":         {},
		"Functional":    {},
		"Contract":      {},
		"Performance":   {},
		"Security":      {},
		"Debugging":     {},
		"Orchestration": {},
	}

	for _, tool := range tools {
		name := tool.Name()
		switch name {
		case "http_request", "variable", "auth", "wait", "retry":
			domains["Core"] = append(domains["Core"], tool)

		case "request", "environment", "falcon_write", "falcon_read", "memory", "session_log":
			domains["Persistence"] = append(domains["Persistence"], tool)

		case "ingest_spec":
			domains["Spec"] = append(domains["Spec"], tool)

		case "assert_response", "extract_value", "validate_json_schema":
			domains["Unit"] = append(domains["Unit"], tool)

		case "orchestrate_integration":
			domains["Integration"] = append(domains["Integration"], tool)

		case "run_smoke":
			domains["Smoke"] = append(domains["Smoke"], tool)

		case "generate_functional_tests", "run_data_driven":
			domains["Functional"] = append(domains["Functional"], tool)

		case "verify_idempotency", "compare_responses", "check_regression":
			domains["Contract"] = append(domains["Contract"], tool)

		case "run_performance", "webhook_listener":
			domains["Performance"] = append(domains["Performance"], tool)

		case "scan_security":
			domains["Security"] = append(domains["Security"], tool)

		case "find_handler", "analyze_endpoint", "analyze_failure", "propose_fix",
			"create_test_file", "read_file", "search_code", "write_file", "list_files":
			domains["Debugging"] = append(domains["Debugging"], tool)

		case "auto_test", "run_tests", "test_suite":
			domains["Orchestration"] = append(domains["Orchestration"], tool)
		}
	}

	order := []string{
		"Core", "Persistence", "Spec",
		"Unit", "Integration", "Smoke", "Functional", "Contract",
		"Performance", "Security",
		"Debugging", "Orchestration",
	}

	for _, domain := range order {
		toolList := domains[domain]
		if len(toolList) == 0 {
			continue
		}

		sb.WriteString(fmt.Sprintf("## %s\n\n", domain))
		for _, tool := range toolList {
			sb.WriteString(fmt.Sprintf("**%s**\n", tool.Name()))
			sb.WriteString(fmt.Sprintf("%s\n", tool.Description()))
			sb.WriteString(fmt.Sprintf("Params: %s\n\n", tool.Parameters()))
		}
	}

	return sb.String()
}

// CompactToolReference provides a quick lookup table for the agent (28 tools).
const CompactToolReference = `# QUICK TOOL REFERENCE

## By Intent
| Intent | Tool | Key Params |
|--------|------|------------|
| Make API call | http_request | method, url, headers?, body? |
| Set/get variable | variable | action="set\|get", name, value, scope |
| Authenticate | auth | action="bearer\|basic\|oauth2\|parse_jwt", token/credentials |
| Delay | wait | seconds |
| Retry a tool | retry | tool, args, max_attempts |
| Save/load/list requests | request | action="save\|load\|list", name?, method?, url? |
| Manage environments | environment | action="set\|list", name?, variables? |
| Write to .falcon/ | falcon_write | path, content, format="yaml\|json\|markdown" |
| Read from .falcon/ | falcon_read | path, format="raw\|yaml\|json" |
| Session audit | session_log | action="start\|end\|list\|read", summary? |
| Save/recall API knowledge | memory | action="save\|recall\|forget\|list\|update_knowledge" |
| Parse OpenAPI/Postman spec | ingest_spec | file_path, format |
| Assert HTTP response | assert_response | status_code?, body_contains?, json_path? |
| Extract value from response | extract_value | json_path/header/cookie/regex, save_as |
| Validate JSON schema | validate_json_schema | schema |
| Compare two responses | compare_responses | response_a, response_b |
| Check regression baseline | check_regression | baseline, current |
| Verify idempotency | verify_idempotency | endpoint, method |
| Generate functional tests | generate_functional_tests | strategy (happy/negative/boundary/all) |
| Run test scenarios | run_tests | scenarios, base_url, scenario? (optional single) |
| Data-driven test | run_data_driven | endpoint, data_file |
| Auto full test flow | auto_test | endpoint, base_url |
| Smoke test | run_smoke | - |
| Integration workflow | orchestrate_integration | workflow |
| Test suite | test_suite | name, tests |
| Load/stress test | run_performance | mode (load/stress/spike/soak), duration_seconds, users |
| Webhook capture | webhook_listener | port?, timeout? |
| Security scan | scan_security | type (owasp/fuzz/auth/all) |
| Find handler in code | find_handler | endpoint, method |
| Analyze endpoint code | analyze_endpoint | endpoint |
| Diagnose test failure | analyze_failure | test_results |
| Propose code fix | propose_fix | file, vulnerability_description |
| Create test file | create_test_file | file, framework |
| Search codebase | search_code | pattern, file_pattern? |
| Read source file | read_file | path, start_line?, end_line? |
| Write source file | write_file | path, content |
| List source files | list_files | path?, pattern? |

## By Domain
**Core**: http_request, variable, auth, wait, retry
**Persistence**: request, environment, falcon_write, falcon_read, memory, session_log
**Spec**: ingest_spec
**Unit/Functional Testing**: assert_response, extract_value, validate_json_schema, generate_functional_tests, run_tests, run_data_driven
**Contract Testing**: compare_responses, check_regression, verify_idempotency
**Integration/E2E**: orchestrate_integration, auto_test, test_suite
**Smoke**: run_smoke
**Performance**: run_performance, webhook_listener
**Security**: scan_security
**Debugging**: find_handler, analyze_endpoint, analyze_failure, propose_fix, create_test_file, read_file, search_code, write_file, list_files

`
