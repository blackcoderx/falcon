package tools

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// AssertTool provides response validation capabilities
type AssertTool struct {
	responseManager *ResponseManager
}

// NewAssertTool creates a new assertion tool
func NewAssertTool(responseManager *ResponseManager) *AssertTool {
	return &AssertTool{
		responseManager: responseManager,
	}
}

// AssertParams defines validation criteria
type AssertParams struct {
	StatusCode          *int                `json:"status_code,omitempty"`
	StatusCodeNot       *int                `json:"status_code_not,omitempty"`
	Headers             map[string]string   `json:"headers,omitempty"`
	HeadersNotPresent   []string            `json:"headers_not_present,omitempty"`
	BodyContains        []string            `json:"body_contains,omitempty"`
	BodyNotContains     []string            `json:"body_not_contains,omitempty"`
	BodyEquals          interface{}         `json:"body_equals,omitempty"`
	BodyMatchesRegex    string              `json:"body_matches_regex,omitempty"`
	JSONPath            map[string]interface{} `json:"json_path,omitempty"` // path -> expected value
	ResponseTimeMaxMs   *int                `json:"response_time_max_ms,omitempty"`
	ContentType         string              `json:"content_type,omitempty"`
}

// AssertionResult represents the outcome of assertions
type AssertionResult struct {
	Passed       bool     `json:"passed"`
	TotalChecks  int      `json:"total_checks"`
	PassedChecks int      `json:"passed_checks"`
	FailedChecks int      `json:"failed_checks"`
	Failures     []string `json:"failures,omitempty"`
}

// Name returns the tool name
func (t *AssertTool) Name() string {
	return "assert_response"
}

// Description returns the tool description
func (t *AssertTool) Description() string {
	return "Validate the last HTTP response against expected criteria (status code, headers, body content, timing)"
}

// Parameters returns the tool parameter description
func (t *AssertTool) Parameters() string {
	return `{
  "status_code": 200,
  "headers": {"Content-Type": "application/json"},
  "body_contains": ["user_id", "email"],
  "body_not_contains": ["error"],
  "body_equals": {"status": "ok"},
  "json_path": {"$.data.id": 123, "$.status": "active"},
  "response_time_max_ms": 500
}`
}

// Execute performs assertions on the last HTTP response
func (t *AssertTool) Execute(args string) (string, error) {
	lastResponse := t.responseManager.GetHTTPResponse()
	if lastResponse == nil {
		return "", fmt.Errorf("no HTTP response available - make an http_request first")
	}

	var params AssertParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse assertion parameters: %w", err)
	}

	result := t.runAssertions(params, lastResponse)

	// Format result
	var sb strings.Builder
	if result.Passed {
		sb.WriteString(fmt.Sprintf("✓ All assertions passed (%d/%d checks)\n\n", result.PassedChecks, result.TotalChecks))
	} else {
		sb.WriteString(fmt.Sprintf("✗ Assertions failed (%d/%d checks passed)\n\n", result.PassedChecks, result.TotalChecks))
		sb.WriteString("Failures:\n")
		for i, failure := range result.Failures {
			sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, failure))
		}
	}

	return sb.String(), nil
}

// runAssertions executes all validation checks
func (t *AssertTool) runAssertions(params AssertParams, lastResponse *HTTPResponse) AssertionResult {
	result := AssertionResult{
		Passed:   true,
		Failures: []string{},
	}

	// Check status code
	if params.StatusCode != nil {
		result.TotalChecks++
		if lastResponse.StatusCode != *params.StatusCode {
			result.Failures = append(result.Failures,
				fmt.Sprintf("Expected status %d, got %d", *params.StatusCode, lastResponse.StatusCode))
			result.Passed = false
		} else {
			result.PassedChecks++
		}
	}

	// Check status code NOT equals
	if params.StatusCodeNot != nil {
		result.TotalChecks++
		if lastResponse.StatusCode == *params.StatusCodeNot {
			result.Failures = append(result.Failures,
				fmt.Sprintf("Status code should not be %d", *params.StatusCodeNot))
			result.Passed = false
		} else {
			result.PassedChecks++
		}
	}

	// Check headers
	for key, expectedValue := range params.Headers {
		result.TotalChecks++
		actualValue, ok := lastResponse.Headers[key]
		if !ok {
			result.Failures = append(result.Failures,
				fmt.Sprintf("Header '%s' not found", key))
			result.Passed = false
		} else if !strings.Contains(actualValue, expectedValue) {
			result.Failures = append(result.Failures,
				fmt.Sprintf("Header '%s': expected '%s', got '%s'", key, expectedValue, actualValue))
			result.Passed = false
		} else {
			result.PassedChecks++
		}
	}

	// Check headers NOT present
	for _, key := range params.HeadersNotPresent {
		result.TotalChecks++
		if _, ok := lastResponse.Headers[key]; ok {
			result.Failures = append(result.Failures,
				fmt.Sprintf("Header '%s' should not be present", key))
			result.Passed = false
		} else {
			result.PassedChecks++
		}
	}

	// Check body contains
	for _, needle := range params.BodyContains {
		result.TotalChecks++
		if !strings.Contains(lastResponse.Body, needle) {
			result.Failures = append(result.Failures,
				fmt.Sprintf("Body does not contain '%s'", needle))
			result.Passed = false
		} else {
			result.PassedChecks++
		}
	}

	// Check body NOT contains
	for _, needle := range params.BodyNotContains {
		result.TotalChecks++
		if strings.Contains(lastResponse.Body, needle) {
			result.Failures = append(result.Failures,
				fmt.Sprintf("Body should not contain '%s'", needle))
			result.Passed = false
		} else {
			result.PassedChecks++
		}
	}

	// Check body equals (JSON comparison)
	if params.BodyEquals != nil {
		result.TotalChecks++
		expectedJSON, _ := json.Marshal(params.BodyEquals)
		var actualData, expectedData interface{}

		if err := json.Unmarshal([]byte(lastResponse.Body), &actualData); err != nil {
			result.Failures = append(result.Failures,
				fmt.Sprintf("Response body is not valid JSON: %v", err))
			result.Passed = false
		} else if err := json.Unmarshal(expectedJSON, &expectedData); err != nil {
			result.Failures = append(result.Failures,
				fmt.Sprintf("Expected body is not valid JSON: %v", err))
			result.Passed = false
		} else if !deepEqual(actualData, expectedData) {
			result.Failures = append(result.Failures,
				fmt.Sprintf("Body mismatch:\nExpected: %s\nGot: %s", expectedJSON, lastResponse.Body))
			result.Passed = false
		} else {
			result.PassedChecks++
		}
	}

	// Check regex match
	if params.BodyMatchesRegex != "" {
		result.TotalChecks++
		matched, err := regexp.MatchString(params.BodyMatchesRegex, lastResponse.Body)
		if err != nil {
			result.Failures = append(result.Failures,
				fmt.Sprintf("Invalid regex pattern: %v", err))
			result.Passed = false
		} else if !matched {
			result.Failures = append(result.Failures,
				fmt.Sprintf("Body does not match regex: %s", params.BodyMatchesRegex))
			result.Passed = false
		} else {
			result.PassedChecks++
		}
	}

	// Check JSON path values
	if len(params.JSONPath) > 0 {
		var jsonData map[string]interface{}
		if err := json.Unmarshal([]byte(lastResponse.Body), &jsonData); err != nil {
			result.TotalChecks += len(params.JSONPath)
			result.Failures = append(result.Failures,
				fmt.Sprintf("Cannot parse response as JSON for JSONPath checks: %v", err))
			result.Passed = false
		} else {
			for path, expectedValue := range params.JSONPath {
				result.TotalChecks++
				actualValue, err := getJSONPath(jsonData, path)
				if err != nil {
					result.Failures = append(result.Failures,
						fmt.Sprintf("JSONPath '%s': %v", path, err))
					result.Passed = false
				} else if !deepEqual(actualValue, expectedValue) {
					result.Failures = append(result.Failures,
						fmt.Sprintf("JSONPath '%s': expected %v, got %v", path, expectedValue, actualValue))
					result.Passed = false
				} else {
					result.PassedChecks++
				}
			}
		}
	}

	// Check response time
	if params.ResponseTimeMaxMs != nil {
		result.TotalChecks++
		actualMs := lastResponse.Duration.Milliseconds()
		maxMs := int64(*params.ResponseTimeMaxMs)
		if actualMs > maxMs {
			result.Failures = append(result.Failures,
				fmt.Sprintf("Response time %dms exceeded maximum %dms", actualMs, maxMs))
			result.Passed = false
		} else {
			result.PassedChecks++
		}
	}

	// Check content type
	if params.ContentType != "" {
		result.TotalChecks++
		actualContentType, ok := lastResponse.Headers["Content-Type"]
		if !ok {
			result.Failures = append(result.Failures,
				"Content-Type header not found")
			result.Passed = false
		} else if !strings.Contains(actualContentType, params.ContentType) {
			result.Failures = append(result.Failures,
				fmt.Sprintf("Expected Content-Type '%s', got '%s'", params.ContentType, actualContentType))
			result.Passed = false
		} else {
			result.PassedChecks++
		}
	}

	result.FailedChecks = result.TotalChecks - result.PassedChecks
	return result
}

// deepEqual compares two interface{} values deeply
func deepEqual(a, b interface{}) bool {
	aJSON, _ := json.Marshal(a)
	bJSON, _ := json.Marshal(b)
	return string(aJSON) == string(bJSON)
}

// getJSONPath extracts a value from nested JSON using a simple path syntax
// Supports: $.field, $.nested.field, $.array[0]
func getJSONPath(data map[string]interface{}, path string) (interface{}, error) {
	// Remove leading $. if present
	path = strings.TrimPrefix(path, "$.")
	if path == "" || path == "$" {
		return data, nil
	}

	parts := strings.Split(path, ".")
	var current interface{} = data

	for _, part := range parts {
		// Handle array indexing: field[0]
		if strings.Contains(part, "[") {
			fieldName := part[:strings.Index(part, "[")]
			indexStr := part[strings.Index(part, "[")+1 : strings.Index(part, "]")]

			var index int
			if _, err := fmt.Sscanf(indexStr, "%d", &index); err != nil {
				return nil, fmt.Errorf("invalid array index: %s", indexStr)
			}

			if fieldName != "" {
				m, ok := current.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("expected object at '%s'", fieldName)
				}
				current = m[fieldName]
			}

			arr, ok := current.([]interface{})
			if !ok {
				return nil, fmt.Errorf("expected array at '%s'", part)
			}
			if index < 0 || index >= len(arr) {
				return nil, fmt.Errorf("array index %d out of bounds", index)
			}
			current = arr[index]
		} else {
			// Regular field access
			m, ok := current.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("expected object, got %T", current)
			}
			value, ok := m[part]
			if !ok {
				return nil, fmt.Errorf("field '%s' not found", part)
			}
			current = value
		}
	}

	return current, nil
}
