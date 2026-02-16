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

	// Group tools by domain (extracted from module path or naming convention)
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
		switch {
		// Foundation
		case name == "http_request", name == "variable", name == "save_request", name == "load_request",
			name == "list_requests", name == "set_environment", name == "list_environments":
			domains["Foundation"] = append(domains["Foundation"], tool)

		// Discovery
		case name == "ingest_spec", name == "map_dependencies", name == "analyze_endpoint":
			domains["Discovery"] = append(domains["Discovery"], tool)

		// Test Generation
		case name == "generate_functional_tests", name == "generate_tests":
			domains["Test Gen"] = append(domains["Test Gen"], tool)

		// Execution
		case name == "auto_test", name == "run_tests", name == "run_single_test", name == "run_smoke", name == "run_data_driven":
			domains["Execution"] = append(domains["Execution"], tool)

		// Validation
		case name == "assert_response", name == "extract_value", name == "verify_schema_conformance",
			name == "check_regression", name == "verify_idempotency", name == "compare_responses",
			name == "validate_json_schema":
			domains["Validation"] = append(domains["Validation"], tool)

		// Security
		case name == "scan_security", name == "auth_bearer", name == "auth_basic", name == "auth_oauth2",
			name == "auth_helper":
			domains["Security"] = append(domains["Security"], tool)

		// Performance
		case name == "run_performance", name == "performance_test", name == "wait", name == "retry",
			name == "webhook_listener":
			domains["Performance"] = append(domains["Performance"], tool)

		// Debugging
		case name == "find_handler", name == "analyze_failure", name == "propose_fix", name == "create_test_file",
			name == "read_file", name == "search_code", name == "write_file":
			domains["Debugging"] = append(domains["Debugging"], tool)

		// Orchestration
		case name == "orchestrate_integration", name == "test_suite", name == "scaffold_unit_tests":
			domains["Orchestration"] = append(domains["Orchestration"], tool)

		// Reporting
		case name == "security_report", name == "export_results", name == "memory":
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
| Set variable | variable | action="set", name, value, scope |
| Get variable | variable | action="get", name |
| Parse spec | ingest_spec | file_path, format (openapi/postman) |
| Map resources | map_dependencies | - |
| Gen tests | generate_functional_tests | strategy (happy/negative/boundary/all) |
| Run tests | run_tests | scenarios, concurrency? |
| Auto test | auto_test | endpoint, method |
| Smoke test | run_smoke | - |
| Assert | assert_response | status_code?, body_contains?, json_path? |
| Extract | extract_value | json_path/header/cookie/regex, save_as |
| Security scan | scan_security | type (owasp/fuzz/auth/all) |
| Load test | run_performance | mode (load/stress/spike/soak), duration_seconds, users |
| Regression | check_regression | baseline, current |
| Idempotency | verify_idempotency | endpoint, method |
| Schema check | verify_schema_conformance | endpoint, method |
| Find handler | find_handler | endpoint, method |
| Analyze fail | analyze_failure | test_results |
| Propose fix | propose_fix | file, vulnerability_description |
| Search code | search_code | pattern, file_pattern? |
| Read file | read_file | path, start_line?, end_line? |
| Export | export_results | format (json/markdown), output_path? |
| Memory | memory | action (save/recall/list/delete), key?, value? |

## By Domain
**Foundation**: http_request, variable, save/load_request, set/list_environment
**Discovery**: ingest_spec, map_dependencies, analyze_endpoint
**Testing**: generate_functional_tests, run_tests, auto_test, run_smoke
**Validation**: assert_response, extract_value, verify_schema_conformance
**Security**: scan_security, auth_* tools
**Debug**: find_handler, analyze_failure, propose_fix, read/search/write_file
**Reports**: security_report, export_results

`
