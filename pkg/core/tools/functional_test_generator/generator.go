package functional_test_generator

import (
	"fmt"
	"time"

	"github.com/blackcoderx/falcon/pkg/core/tools/shared"
)

// TestGenerator is responsible for executing generated test scenarios.
type TestGenerator struct {
	httpTool   *shared.HTTPTool
	assertTool *shared.AssertTool
}

// NewTestGenerator creates a new test generator.
func NewTestGenerator(httpTool *shared.HTTPTool, assertTool *shared.AssertTool) *TestGenerator {
	return &TestGenerator{
		httpTool:   httpTool,
		assertTool: assertTool,
	}
}

// ExecuteScenarios runs all test scenarios and returns the results.
func (g *TestGenerator) ExecuteScenarios(scenarios []shared.TestScenario) []shared.TestResult {
	results := make([]shared.TestResult, 0, len(scenarios))

	for _, scenario := range scenarios {
		result := g.executeScenario(scenario)
		results = append(results, result)
	}

	return results
}

// executeScenario runs a single test scenario and returns the result.
func (g *TestGenerator) executeScenario(scenario shared.TestScenario) shared.TestResult {
	startTime := time.Now()

	result := shared.TestResult{
		ScenarioID:     scenario.ID,
		ScenarioName:   scenario.Name,
		Category:       scenario.Category,
		Severity:       scenario.Severity,
		OWASPRef:       scenario.OWASPRef,
		Timestamp:      startTime.Format(time.RFC3339),
		Logs:           []string{},
		ExpectedStatus: scenario.Expected.StatusCode,
	}

	// Execute HTTP request
	httpReq := shared.HTTPRequest{
		Method:  scenario.Method,
		URL:     scenario.URL,
		Headers: scenario.Headers,
		Body:    scenario.Body,
	}

	resp, err := g.httpTool.Run(httpReq)
	if err != nil {
		result.Passed = false
		result.Error = fmt.Sprintf("HTTP request failed: %v", err)
		result.DurationMs = time.Since(startTime).Milliseconds()
		result.Logs = append(result.Logs, result.Error)
		return result
	}

	result.ActualStatus = resp.StatusCode
	result.ResponseBody = resp.Body
	result.DurationMs = resp.Duration.Milliseconds()

	// Validate response against expectations
	passed, validationErrors := g.validateResponse(scenario.Expected, resp)
	result.Passed = passed

	if !passed {
		result.Error = fmt.Sprintf("Validation failed: %s", joinErrors(validationErrors))
		result.Logs = validationErrors
	} else {
		result.Logs = append(result.Logs, "All validations passed")
	}

	return result
}

// validateResponse checks if the HTTP response meets the expectations.
func (g *TestGenerator) validateResponse(expected shared.TestExpectation, resp *shared.HTTPResponse) (bool, []string) {
	var errors []string

	// Check status code
	if expected.StatusCode > 0 {
		if resp.StatusCode != expected.StatusCode {
			errors = append(errors, fmt.Sprintf("Expected status %d, got %d", expected.StatusCode, resp.StatusCode))
		}
	}

	// Check status code range
	if expected.StatusCodeRange != nil {
		if resp.StatusCode < expected.StatusCodeRange.Min || resp.StatusCode > expected.StatusCodeRange.Max {
			errors = append(errors, fmt.Sprintf("Status code %d outside expected range %d-%d",
				resp.StatusCode, expected.StatusCodeRange.Min, expected.StatusCodeRange.Max))
		}
	}

	// Check body contains
	for _, needle := range expected.BodyContains {
		if !contains(resp.Body, needle) {
			errors = append(errors, fmt.Sprintf("Body does not contain '%s'", needle))
		}
	}

	// Check body does not contain
	for _, needle := range expected.BodyNotContains {
		if contains(resp.Body, needle) {
			errors = append(errors, fmt.Sprintf("Body should not contain '%s'", needle))
		}
	}

	// Check headers
	for key, expectedValue := range expected.HeaderContains {
		actualValue, ok := resp.Headers[key]
		if !ok {
			errors = append(errors, fmt.Sprintf("Header '%s' not found", key))
		} else if !contains(actualValue, expectedValue) {
			errors = append(errors, fmt.Sprintf("Header '%s': expected to contain '%s', got '%s'", key, expectedValue, actualValue))
		}
	}

	// Check max duration
	if expected.MaxDurationMs > 0 {
		if resp.Duration.Milliseconds() > int64(expected.MaxDurationMs) {
			errors = append(errors, fmt.Sprintf("Response time %dms exceeded maximum %dms",
				resp.Duration.Milliseconds(), expected.MaxDurationMs))
		}
	}

	return len(errors) == 0, errors
}

// contains checks if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || stringContains(s, substr))
}

// stringContains is a simple substring check.
func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// joinErrors combines multiple error messages into a single string.
func joinErrors(errors []string) string {
	if len(errors) == 0 {
		return ""
	}
	if len(errors) == 1 {
		return errors[0]
	}

	result := errors[0]
	for i := 1; i < len(errors); i++ {
		result += "; " + errors[i]
	}
	return result
}
