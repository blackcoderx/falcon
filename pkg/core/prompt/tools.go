package prompt

import (
	"fmt"
	"strings"
)

// BuildToolsSection generates a context-efficient tool reference.
// Instead of verbose descriptions, it uses a compact tabular format grouped by domain.
func BuildToolsSection(tools map[string]Tool) string {
	var sb strings.Builder

	sb.WriteString("# AVAILABLE TOOLS\n\n")
	sb.WriteString("**Call Format**: ACTION: tool_name({\"param\": \"value\"})\n\n")

	// Group tools by domain
	domains := map[string][]Tool{
		"Foundation":    {},
		"Discovery":     {},
		"Test Gen":      {},
		"Execution":     {},
		"Validation":    {},
		"Security":      {},
		"Performance":   {},
		"Debugging":     {},
		"Orchestration": {},
		"Reporting":     {},
	}

	// Categorize tools
	for _, tool := range tools {
		name := tool.Name()
		switch name {
		case "http_request", "variable", "save_request", "load_request", "list_requests", "set_environment", "list_environments":
			domains["Foundation"] = append(domains["Foundation"], tool)

		case "ingest_spec", "map_dependencies", "analyze_endpoint":
			domains["Discovery"] = append(domains["Discovery"], tool)

		case "generate_functional_tests", "generate_tests":
			domains["Test Gen"] = append(domains["Test Gen"], tool)

		case "auto_test", "run_tests", "run_single_test", "run_smoke", "run_data_driven":
			domains["Execution"] = append(domains["Execution"], tool)

		case "assert_response", "extract_value", "verify_schema_conformance", "check_regression", "verify_idempotency", "compare_responses", "validate_json_schema":
			domains["Validation"] = append(domains["Validation"], tool)

		case "scan_security", "auth_bearer", "auth_basic", "auth_oauth2", "auth_helper":
			domains["Security"] = append(domains["Security"], tool)

		case "run_performance", "performance_test", "wait", "retry", "webhook_listener":
			domains["Performance"] = append(domains["Performance"], tool)

		case "find_handler", "analyze_failure", "propose_fix", "create_test_file", "read_file", "search_code", "write_file", "list_files":
			domains["Debugging"] = append(domains["Debugging"], tool)

		case "orchestrate_integration", "test_suite", "scaffold_unit_tests":
			domains["Orchestration"] = append(domains["Orchestration"], tool)

		case "security_report", "export_results", "memory":
			domains["Reporting"] = append(domains["Reporting"], tool)
		}
	}

	// Render each domain
	order := []string{"Foundation", "Discovery", "Test Gen", "Execution", "Validation", "Security", "Performance", "Debugging", "Orchestration", "Reporting"}

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

// CompactToolReference provides a quick lookup table (ultra-compact for context efficiency).
const CompactToolReference = `# QUICK TOOL REFERENCE

## By Intent
| Intent | Tool | Key Params |
|--------|------|------------|
| Make API call | http_request | method, url, headers?, body? |
| Save request | save_request | name, method, url, headers?, body? |
| Load request | load_request | name |
| List saved requests | list_requests | - |
| Set variable | variable | action="set", name, value, scope |
| Get variable | variable | action="get", name |
| Set environment | set_environment | name, variables |
| List environments | list_environments | - |
| Assert response | assert_response | status_code?, body_contains?, json_path? |
| Extract value | extract_value | json_path/header/cookie/regex, save_as |
| Validate schema | validate_json_schema | schema |
| Compare responses | compare_responses | response_a, response_b |
| Parse spec | ingest_spec | file_path, format (openapi/postman) |
| Map resources | map_dependencies | - |
| Gen tests | generate_functional_tests | strategy (happy/negative/boundary/all) |
| Run tests | run_tests | scenarios, concurrency? |
| Run single test | run_single_test | scenario |
| Auto test | auto_test | endpoint, method |
| Smoke test | run_smoke | - |
| Data-driven test | run_data_driven | endpoint, data_file |
| Security scan | scan_security | type (owasp/fuzz/auth/all) |
| Auth (bearer) | auth_bearer | token |
| Auth (basic) | auth_basic | username, password |
| Auth (OAuth2) | auth_oauth2 | token_url, client_id, client_secret |
| Load test | run_performance | mode (load/stress/spike/soak), duration_seconds, users |
| Regression check | check_regression | baseline, current |
| Idempotency check | verify_idempotency | endpoint, method |
| Schema conformance | verify_schema_conformance | endpoint, method |
| Find handler | find_handler | endpoint, method |
| Analyze failure | analyze_failure | test_results |
| Propose fix | propose_fix | file, vulnerability_description |
| Search code | search_code | pattern, file_pattern? |
| Read file | read_file | path, start_line?, end_line? |
| Write file | write_file | path, content |
| List files | list_files | path?, pattern? |
| Export results | export_results | format (json/markdown), output_path? |
| Save to memory | memory | action="save", key, value |
| Recall memory | memory | action="recall" |
| Wait/delay | wait | seconds |
| Retry tool | retry | tool, args, max_attempts, retry_delay_ms |

## By Domain
**Foundation**: http_request, variable, save/load/list_requests, set/list_environments
**Discovery**: ingest_spec, map_dependencies, analyze_endpoint
**Testing**: generate_functional_tests, generate_tests, run_tests, run_single_test, auto_test, run_smoke, run_data_driven
**Validation**: assert_response, extract_value, validate_json_schema, compare_responses, verify_schema_conformance, check_regression, verify_idempotency
**Security**: scan_security, auth_bearer, auth_basic, auth_oauth2, auth_helper
**Debug**: find_handler, analyze_failure, propose_fix, read_file, search_code, write_file, list_files, create_test_file
**Performance**: run_performance, performance_test, wait, retry, webhook_listener
**Orchestration**: orchestrate_integration, test_suite, scaffold_unit_tests
**Reports**: export_results, memory

`
