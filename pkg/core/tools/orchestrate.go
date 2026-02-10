package tools

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

// RunTestsTool executes multiple test scenarios
type RunTestsTool struct {
	httpTool   *HTTPTool
	assertTool *AssertTool
	varStore   *VariableStore
}

// NewRunTestsTool creates a new run_tests tool
func NewRunTestsTool(httpTool *HTTPTool, assertTool *AssertTool, varStore *VariableStore) *RunTestsTool {
	return &RunTestsTool{
		httpTool:   httpTool,
		assertTool: assertTool,
		varStore:   varStore,
	}
}

// RunTestsParams defines input for run_tests
type RunTestsParams struct {
	Scenarios   []TestScenario `json:"scenarios"`
	BaseURL     string         `json:"base_url"`
	Category    string         `json:"category,omitempty"`
	Categories  []string       `json:"categories,omitempty"`
	Concurrency int            `json:"concurrency,omitempty"`
	TimeoutMs   int            `json:"timeout_ms,omitempty"`
}

func (t *RunTestsTool) Name() string {
	return "run_tests"
}

func (t *RunTestsTool) Description() string {
	return "Execute multiple test scenarios in parallel with configurable concurrency, collect all results, and stream real-time progress updates to the TUI. Supports category filtering."
}

func (t *RunTestsTool) Parameters() string {
	return `{
  "scenarios": [ /* array of TestScenario */ ],
  "base_url": "http://localhost:8080",
  "category": "security",
  "categories": ["security", "validation"],
  "concurrency": 5,
  "timeout_ms": 30000
}`
}

func (t *RunTestsTool) Execute(args string) (string, error) {
	var params RunTestsParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	if params.BaseURL == "" {
		return "", fmt.Errorf("base_url is required")
	}

	// Filter scenarios
	var scenariosToRun []TestScenario
	for _, s := range params.Scenarios {
		if params.Category != "" && s.Category != params.Category {
			continue
		}
		if len(params.Categories) > 0 {
			match := false
			for _, cat := range params.Categories {
				if s.Category == cat {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}
		scenariosToRun = append(scenariosToRun, s)
	}

	if len(scenariosToRun) == 0 {
		return "No scenarios matched the criteria.", nil
	}

	concurrency := params.Concurrency
	if concurrency <= 0 {
		concurrency = 5
	}

	results := make([]TestResult, len(scenariosToRun))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, concurrency)

	// TUI Update: "Running %d tests..."
	fmt.Printf("Running %d tests with concurrency %d...\n", len(scenariosToRun), concurrency)

	for i, scenario := range scenariosToRun {
		wg.Add(1)
		go func(idx int, s TestScenario) {
			defer wg.Done()
			semaphore <- struct{}{} // Acquire
			results[idx] = t.runSingleScenario(s, params.BaseURL)
			<-semaphore // Release
			// Optional: TUI update for progress
		}(i, scenario)
	}

	wg.Wait()

	// Summarize results
	passed := 0
	failed := 0
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Test Results (%d tests):\n\n", len(results)))

	for _, res := range results {
		statusIcon := "✓"
		if !res.Passed {
			statusIcon = "✗"
			failed++
		} else {
			passed++
		}
		sb.WriteString(fmt.Sprintf("%s [%s] %s (%dms)\n", statusIcon, res.Category, res.ScenarioName, res.DurationMs))
		if !res.Passed {
			sb.WriteString(fmt.Sprintf("  Error: %s\n", res.Error))
			if res.ExpectedStatus != 0 {
				sb.WriteString(fmt.Sprintf("  Status: Expected %d, Got %d\n", res.ExpectedStatus, res.ActualStatus))
			}
		}
	}

	sb.WriteString(fmt.Sprintf("\nSummary: %d Passed, %d Failed\n", passed, failed))

	// Check if this tool is expected to return raw JSON or human text.
	// The prompt implies returning "[]TestResult with full summary" but description says "collect all results".
	// Usually tools return text.
	// We can return the summary text AND maybe full JSON if needed?
	// For now, text summary with detailed failures is good.
	// To enable programmatic access, we can dump JSON if requested.
	// But let's stick to text for TUI compatibility for now.

	return sb.String(), nil
}

func (t *RunTestsTool) runSingleScenario(s TestScenario, baseURL string) TestResult {
	startTime := time.Now()
	res := TestResult{
		ScenarioID:   s.ID,
		ScenarioName: s.Name,
		Category:     s.Category,
		Severity:     s.Severity,
		OWASPRef:     s.OWASPRef,
		Timestamp:    time.Now().Format(time.RFC3339),
		Passed:       true,
	}

	// Prepare request
	url := baseURL + s.URL
	if !strings.HasPrefix(s.URL, "/") {
		url = baseURL + "/" + s.URL
	}

	req := HTTPRequest{
		Method:  s.Method,
		URL:     url,
		Headers: s.Headers,
		Body:    s.Body,
	}

	// Run request
	httpResp, err := t.httpTool.Run(req)
	duration := time.Since(startTime)
	res.DurationMs = duration.Milliseconds()

	if err != nil {
		res.Passed = false
		res.Error = fmt.Sprintf("Request failed: %v", err)
		return res
	}

	res.ActualStatus = httpResp.StatusCode
	res.ResponseBody = httpResp.Body
	res.ExpectedStatus = s.Expected.StatusCode

	// Assertions
	// 1. Status Code
	if s.Expected.StatusCode != 0 && s.Expected.StatusCode != httpResp.StatusCode {
		res.Passed = false
		res.Error = fmt.Sprintf("Status code mismatch: expected %d, got %d", s.Expected.StatusCode, httpResp.StatusCode)
	}

	// 2. Status Code Range
	if s.Expected.StatusCodeRange != nil {
		if httpResp.StatusCode < s.Expected.StatusCodeRange.Min || httpResp.StatusCode > s.Expected.StatusCodeRange.Max {
			res.Passed = false
			res.Error = fmt.Sprintf("Status code %d out of range [%d-%d]", httpResp.StatusCode, s.Expected.StatusCodeRange.Min, s.Expected.StatusCodeRange.Max)
		}
	}

	// 3. Body Contains
	for _, contains := range s.Expected.BodyContains {
		if !strings.Contains(httpResp.Body, contains) {
			res.Passed = false
			res.Error = fmt.Sprintf("Body missing expected string: '%s'", contains)
		}
	}

	// 4. Body Not Contains
	for _, notContains := range s.Expected.BodyNotContains {
		if strings.Contains(httpResp.Body, notContains) {
			res.Passed = false
			res.Error = fmt.Sprintf("Body contains forbidden string: '%s'", notContains)
		}
	}

	// 5. Max Duration
	if s.Expected.MaxDurationMs > 0 && res.DurationMs > int64(s.Expected.MaxDurationMs) {
		res.Passed = false
		res.Error = fmt.Sprintf("Response time %dms exceeded max %dms", res.DurationMs, s.Expected.MaxDurationMs)
	}

	// TODO: AssertTool could be used for more complex assertions if we matched the params type
	// For now, this inline logic covers the basics defined in TestScenario

	return res
}

// RunSingleTestTool executes a single test scenario
type RunSingleTestTool struct {
	httpTool   *HTTPTool
	assertTool *AssertTool
	varStore   *VariableStore
}

func NewRunSingleTestTool(httpTool *HTTPTool, assertTool *AssertTool, varStore *VariableStore) *RunSingleTestTool {
	return &RunSingleTestTool{
		httpTool:   httpTool,
		assertTool: assertTool,
		varStore:   varStore,
	}
}

type RunSingleTestParams struct {
	TestID   string        `json:"test_id"`
	Scenario *TestScenario `json:"scenario,omitempty"` // Option to pass full scenario
	BaseURL  string        `json:"base_url"`
}

func (t *RunSingleTestTool) Name() string {
	return "run_single_test"
}

func (t *RunSingleTestTool) Description() string {
	return "Execute one specific test scenario by name or ID, useful for re-running a particular test after applying a fix to verify it works"
}

func (t *RunSingleTestTool) Parameters() string {
	return `{
  "test_id": "sec-001",
  "base_url": "http://localhost:8080",
  "scenario": { ... } // Optional check to provide full scenario definition if not found in memory
}`
}

func (t *RunSingleTestTool) Execute(args string) (string, error) {
	// Re-uses logic from RunTestsTool, effectively just running one.
	// But in a real system, we'd need to look up the test ID from somewhere.
	// Since we don't have a persistent test execution DB yet, we might rely on the user passing the scenario again
	// or assume the agent context has it.
	// For this sprint, let's assume the user passes the scenario OR we just implement it to accept the scenario.

	var params RunSingleTestParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("parse error: %w", err)
	}

	if params.Scenario == nil {
		return "", fmt.Errorf("scenario definition is required for run_single_test currently (persistence not implemented)")
	}

	runner := NewRunTestsTool(t.httpTool, t.assertTool, t.varStore)
	result := runner.runSingleScenario(*params.Scenario, params.BaseURL)

	resJSON, _ := json.MarshalIndent(result, "", "  ")
	return string(resJSON), nil
}

// AutoTestTool orchestrates the full autonomous testing workflow
type AutoTestTool struct {
	analyzeTool        *AnalyzeEndpointTool
	generateTool       *GenerateTestsTool
	orchestrateTool    *RunTestsTool
	analyzeFailureTool *AnalyzeFailureTool
}

func NewAutoTestTool(a *AnalyzeEndpointTool, g *GenerateTestsTool, o *RunTestsTool, f *AnalyzeFailureTool) *AutoTestTool {
	return &AutoTestTool{
		analyzeTool:        a,
		generateTool:       g,
		orchestrateTool:    o,
		analyzeFailureTool: f,
	}
}

type AutoTestParams struct {
	Endpoint string `json:"endpoint"` // "POST /api/checkout"
	BaseURL  string `json:"base_url"`
	Context  string `json:"context,omitempty"`
}

func (t *AutoTestTool) Name() string {
	return "auto_test"
}

func (t *AutoTestTool) Description() string {
	return "Full autonomous workflow: analyze endpoint -> generate comprehensive test scenarios -> execute all tests -> analyze failures. The main orchestrator tool."
}

func (t *AutoTestTool) Parameters() string {
	return `{ "endpoint": "POST /api/checkout", "base_url": "http://localhost:8080", "context": "optional" }`
}

func (t *AutoTestTool) Execute(args string) (string, error) {
	var params AutoTestParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("parse error: %w", err)
	}

	// 1. Analyze
	parts := strings.SplitN(params.Endpoint, " ", 2)
	method := "GET"
	url := params.Endpoint
	if len(parts) == 2 {
		method = parts[0]
		url = parts[1]
	}

	analyzeParams := AnalyzeEndpointParams{
		EndpointDescription: params.Endpoint,
		Method:              method,
		URL:                 url,
		Context:             params.Context,
	}
	analyzeJSON, _ := json.Marshal(analyzeParams)
	analysisResult, err := t.analyzeTool.Execute(string(analyzeJSON))
	if err != nil {
		return "", fmt.Errorf("analysis failed: %w", err)
	}

	var analysis EndpointAnalysis
	if err := json.Unmarshal([]byte(analysisResult), &analysis); err != nil {
		return "", fmt.Errorf("failed to parse analysis result: %w", err)
	}

	// 2. Generate
	genParams := GenerateTestsParams{
		Analysis: analysis,
		Count:    20,
	}
	genJSON, _ := json.Marshal(genParams)
	genResult, err := t.generateTool.Execute(string(genJSON))
	if err != nil {
		return "", fmt.Errorf("generation failed: %w", err)
	}

	var scenarios []TestScenario
	if err := json.Unmarshal([]byte(genResult), &scenarios); err != nil {
		return "", fmt.Errorf("failed to parse generated scenarios: %w", err)
	}

	// 3. Run
	results := make([]TestResult, len(scenarios))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 5)

	for i, s := range scenarios {
		wg.Add(1)
		go func(idx int, scenario TestScenario) {
			defer wg.Done()
			semaphore <- struct{}{}
			results[idx] = t.orchestrateTool.runSingleScenario(scenario, params.BaseURL)
			<-semaphore
		}(i, s)
	}
	wg.Wait()

	// 4. Analyze Failures
	var failureReports []string
	passCount := 0
	failCount := 0

	for _, res := range results {
		if res.Passed {
			passCount++
			continue
		}
		failCount++

		failParams := AnalyzeFailureParams{
			TestResult:       res,
			ResponseBody:     res.ResponseBody,
			ExpectedBehavior: fmt.Sprintf("Status %d", res.ExpectedStatus),
		}
		failJSON, _ := json.Marshal(failParams)
		failAnalysis, err := t.analyzeFailureTool.Execute(string(failJSON))
		if err != nil {
			failAnalysis = fmt.Sprintf("Failed to analyze failure: %v", err)
		}
		failureReports = append(failureReports, fmt.Sprintf("## Failure: %s\n%s\n", res.ScenarioName, failAnalysis))
	}

	// 5. Report
	var report strings.Builder
	report.WriteString(fmt.Sprintf("# Auto-Test Report for %s\n\n", params.Endpoint))
	report.WriteString(fmt.Sprintf("Summary: %d Passed, %d Failed\n", passCount, failCount))
	report.WriteString("## Analysis\n" + analysis.Summary + "\n\n")

	if len(failureReports) > 0 {
		report.WriteString("## Failure Analysis\n")
		for _, f := range failureReports {
			report.WriteString(f + "\n")
		}
	} else {
		report.WriteString("## No Failures Detected! 🎉\n")
	}

	return report.String(), nil
}
