package agent

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/blackcoderx/falcon/pkg/core/tools/shared"
)

// RunTestsTool executes multiple test scenarios
type RunTestsTool struct {
	httpTool   *shared.HTTPTool
	assertTool *shared.AssertTool
	varStore   *shared.VariableStore
}

// NewRunTestsTool creates a new run_tests tool
func NewRunTestsTool(httpTool *shared.HTTPTool, assertTool *shared.AssertTool, varStore *shared.VariableStore) *RunTestsTool {
	return &RunTestsTool{
		httpTool:   httpTool,
		assertTool: assertTool,
		varStore:   varStore,
	}
}

// RunTestsParams defines input for run_tests
type RunTestsParams struct {
	Scenarios   []shared.TestScenario `json:"scenarios"`
	BaseURL     string                `json:"base_url"`
	Category    string                `json:"category,omitempty"`
	Categories  []string              `json:"categories,omitempty"`
	Concurrency int                   `json:"concurrency,omitempty"`
	TimeoutMs   int                   `json:"timeout_ms,omitempty"`
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
	var scenariosToRun []shared.TestScenario
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

	results := make([]shared.TestResult, len(scenariosToRun))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, concurrency)

	// In the real app, we might use a callback for TUI updates.
	// For CLI simple output:
	// fmt.Printf("Running %d tests with concurrency %d...\n", len(scenariosToRun), concurrency)

	for i, scenario := range scenariosToRun {
		wg.Add(1)
		go func(idx int, s shared.TestScenario) {
			defer wg.Done()
			semaphore <- struct{}{} // Acquire
			results[idx] = t.runSingleScenario(s, params.BaseURL)
			<-semaphore // Release
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

	return sb.String(), nil
}

func (t *RunTestsTool) runSingleScenario(s shared.TestScenario, baseURL string) shared.TestResult {
	startTime := time.Now()
	res := shared.TestResult{
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

	req := shared.HTTPRequest{
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

	return res
}
