package shared

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// TestExecutor provides reusable test scenario execution with standard assertions.
// Tools with custom execution logic can use ValidateExpectations directly
// without going through RunScenario/RunScenarios.
type TestExecutor struct {
	HTTPTool *HTTPTool
}

// NewTestExecutor creates a new TestExecutor backed by the given HTTPTool.
func NewTestExecutor(httpTool *HTTPTool) *TestExecutor {
	return &TestExecutor{HTTPTool: httpTool}
}

// RunScenario executes a single TestScenario against baseURL and returns a TestResult.
// The baseURL is prepended to the scenario's URL path.
func (e *TestExecutor) RunScenario(scenario TestScenario, baseURL string) TestResult {
	startTime := time.Now()

	// Build full URL
	url := baseURL + scenario.URL
	if baseURL != "" && !strings.HasPrefix(scenario.URL, "/") {
		url = baseURL + "/" + scenario.URL
	}

	req := HTTPRequest{
		Method:  scenario.Method,
		URL:     url,
		Headers: scenario.Headers,
		Body:    scenario.Body,
	}

	resp, err := e.HTTPTool.Run(req)
	durationMs := time.Since(startTime).Milliseconds()

	return e.buildResultWithDuration(scenario, resp, err, durationMs)
}

// RunScenarios executes multiple scenarios with configurable concurrency.
// Results are returned in the same order as the input scenarios.
// If concurrency <= 0, defaults to 5.
func (e *TestExecutor) RunScenarios(scenarios []TestScenario, baseURL string, concurrency int) []TestResult {
	if concurrency <= 0 {
		concurrency = 5
	}

	results := make([]TestResult, len(scenarios))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, concurrency)

	for i, scenario := range scenarios {
		wg.Add(1)
		go func(idx int, s TestScenario) {
			defer wg.Done()
			semaphore <- struct{}{}
			results[idx] = e.RunScenario(s, baseURL)
			<-semaphore
		}(i, scenario)
	}

	wg.Wait()
	return results
}

// ValidateExpectations checks an HTTPResponse against a TestExpectation.
// Returns a list of validation error strings. Empty list means all passed.
// This is exported so tools with custom execution can reuse assertion logic.
func ValidateExpectations(expected TestExpectation, resp *HTTPResponse, durationMs int64) []string {
	var errors []string

	// Status code exact match
	if expected.StatusCode != 0 && expected.StatusCode != resp.StatusCode {
		errors = append(errors, fmt.Sprintf("Status code mismatch: expected %d, got %d", expected.StatusCode, resp.StatusCode))
	}

	// Status code range
	if expected.StatusCodeRange != nil {
		if resp.StatusCode < expected.StatusCodeRange.Min || resp.StatusCode > expected.StatusCodeRange.Max {
			errors = append(errors, fmt.Sprintf("Status code %d out of range [%d-%d]", resp.StatusCode, expected.StatusCodeRange.Min, expected.StatusCodeRange.Max))
		}
	}

	// Body contains
	for _, needle := range expected.BodyContains {
		if !strings.Contains(resp.Body, needle) {
			errors = append(errors, fmt.Sprintf("Body missing expected string: '%s'", needle))
		}
	}

	// Body not contains
	for _, needle := range expected.BodyNotContains {
		if strings.Contains(resp.Body, needle) {
			errors = append(errors, fmt.Sprintf("Body contains forbidden string: '%s'", needle))
		}
	}

	// Header contains
	for key, expectedValue := range expected.HeaderContains {
		actualValue, ok := resp.Headers[key]
		if !ok {
			errors = append(errors, fmt.Sprintf("Header '%s' not found", key))
		} else if !strings.Contains(actualValue, expectedValue) {
			errors = append(errors, fmt.Sprintf("Header '%s': expected to contain '%s', got '%s'", key, expectedValue, actualValue))
		}
	}

	// Max duration
	if expected.MaxDurationMs > 0 && durationMs > int64(expected.MaxDurationMs) {
		errors = append(errors, fmt.Sprintf("Response time %dms exceeded max %dms", durationMs, expected.MaxDurationMs))
	}

	return errors
}

// buildResult constructs a TestResult from a scenario, response, and optional HTTP error.
// Uses the response's own duration if available.
func (e *TestExecutor) buildResult(scenario TestScenario, resp *HTTPResponse, httpErr error) TestResult {
	var durationMs int64
	if resp != nil && resp.Duration > 0 {
		durationMs = resp.Duration.Milliseconds()
	}
	return e.buildResultWithDuration(scenario, resp, httpErr, durationMs)
}

// buildResultWithDuration constructs a TestResult with an explicit duration.
func (e *TestExecutor) buildResultWithDuration(scenario TestScenario, resp *HTTPResponse, httpErr error, durationMs int64) TestResult {
	result := TestResult{
		ScenarioID:     scenario.ID,
		ScenarioName:   scenario.Name,
		Category:       scenario.Category,
		Severity:       scenario.Severity,
		OWASPRef:       scenario.OWASPRef,
		Timestamp:      time.Now().Format(time.RFC3339),
		ExpectedStatus: scenario.Expected.StatusCode,
		DurationMs:     durationMs,
		Passed:         true,
	}

	if httpErr != nil {
		result.Passed = false
		result.Error = fmt.Sprintf("Request failed: %v", httpErr)
		return result
	}

	if resp != nil {
		result.ActualStatus = resp.StatusCode
		result.ResponseBody = resp.Body
	}

	errors := ValidateExpectations(scenario.Expected, resp, durationMs)
	if len(errors) > 0 {
		result.Passed = false
		result.Error = strings.Join(errors, "; ")
		result.Logs = errors
	}

	return result
}
