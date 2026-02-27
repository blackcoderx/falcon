package agent

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/blackcoderx/falcon/pkg/core/tools/debugging"
	"github.com/blackcoderx/falcon/pkg/core/tools/shared"
	"github.com/blackcoderx/falcon/pkg/llm"
)

// AutoTestTool orchestrates the full autonomous testing workflow:
// analyze endpoint → generate test scenarios via LLM → run all tests → diagnose failures.
type AutoTestTool struct {
	llmClient          llm.LLMClient
	analyzeTool        *debugging.AnalyzeEndpointTool
	orchestrateTool    *RunTestsTool
	analyzeFailureTool *debugging.AnalyzeFailureTool
}

func NewAutoTestTool(
	llmClient llm.LLMClient,
	a *debugging.AnalyzeEndpointTool,
	o *RunTestsTool,
	f *debugging.AnalyzeFailureTool,
) *AutoTestTool {
	return &AutoTestTool{
		llmClient:          llmClient,
		analyzeTool:        a,
		orchestrateTool:    o,
		analyzeFailureTool: f,
	}
}

type AutoTestParams struct {
	Endpoint string `json:"endpoint"` // e.g. "POST /api/checkout"
	BaseURL  string `json:"base_url"`
	Context  string `json:"context,omitempty"`
}

func (t *AutoTestTool) Name() string {
	return "auto_test"
}

func (t *AutoTestTool) Description() string {
	return "Full autonomous workflow: analyze endpoint → generate comprehensive test scenarios via LLM → execute all tests → diagnose failures. The main end-to-end orchestrator."
}

func (t *AutoTestTool) Parameters() string {
	return `{ "endpoint": "POST /api/checkout", "base_url": "http://localhost:8080", "context": "optional extra context" }`
}

func (t *AutoTestTool) Execute(args string) (string, error) {
	var params AutoTestParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("parse error: %w", err)
	}

	if params.BaseURL == "" {
		return "", fmt.Errorf("base_url is required")
	}

	// 1. Analyze the endpoint
	parts := strings.SplitN(params.Endpoint, " ", 2)
	method := "GET"
	path := params.Endpoint
	if len(parts) == 2 {
		method = parts[0]
		path = parts[1]
	}

	analyzeParams := debugging.AnalyzeEndpointParams{
		EndpointDescription: params.Endpoint,
		Method:              method,
		URL:                 path,
		Context:             params.Context,
	}
	analyzeJSON, err := json.Marshal(analyzeParams)
	if err != nil {
		return "", fmt.Errorf("failed to marshal analyze params: %w", err)
	}
	analysisResult, err := t.analyzeTool.Execute(string(analyzeJSON))
	if err != nil {
		return "", fmt.Errorf("analysis failed: %w", err)
	}

	var analysis shared.EndpointAnalysis
	if err := json.Unmarshal([]byte(analysisResult), &analysis); err != nil {
		return "", fmt.Errorf("failed to parse analysis result: %w", err)
	}

	// 2. Generate test scenarios via LLM
	scenarios, err := t.generateScenarios(params.Endpoint, params.BaseURL, analysis, params.Context)
	if err != nil {
		return "", fmt.Errorf("scenario generation failed: %w", err)
	}

	if len(scenarios) == 0 {
		return "No test scenarios could be generated for this endpoint.", nil
	}

	// 3. Run all scenarios in parallel
	results := make([]shared.TestResult, len(scenarios))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 5)

	for i, s := range scenarios {
		wg.Add(1)
		go func(idx int, scenario shared.TestScenario) {
			defer wg.Done()
			semaphore <- struct{}{}
			results[idx] = t.orchestrateTool.runSingleScenario(scenario, params.BaseURL)
			<-semaphore
		}(i, s)
	}
	wg.Wait()

	// 4. Diagnose failures
	var failureReports []string
	passCount := 0
	failCount := 0

	for _, res := range results {
		if res.Passed {
			passCount++
			continue
		}
		failCount++

		failParams := debugging.AnalyzeFailureParams{
			TestResult:       res,
			ResponseBody:     res.ResponseBody,
			ExpectedBehavior: fmt.Sprintf("Status %d", res.ExpectedStatus),
		}
		failJSON, marshalErr := json.Marshal(failParams)
		var failAnalysis string
		if marshalErr != nil {
			failAnalysis = fmt.Sprintf("Could not marshal failure params: %v", marshalErr)
		} else {
			failAnalysis, err = t.analyzeFailureTool.Execute(string(failJSON))
			if err != nil {
				failAnalysis = fmt.Sprintf("Failure analysis error: %v", err)
			}
		}
		failureReports = append(failureReports, fmt.Sprintf("## Failure: %s\n%s\n", res.ScenarioName, failAnalysis))
	}

	// 5. Build report
	var report strings.Builder
	fmt.Fprintf(&report, "# Auto-Test Report: %s\n\n", params.Endpoint)
	fmt.Fprintf(&report, "**Summary:** %d Passed, %d Failed out of %d total\n\n", passCount, failCount, len(scenarios))
	fmt.Fprintf(&report, "## Endpoint Analysis\n%s\n\n", analysis.Summary)

	if len(failureReports) > 0 {
		fmt.Fprintf(&report, "## Failure Analysis\n")
		for _, f := range failureReports {
			report.WriteString(f + "\n")
		}
	} else {
		fmt.Fprintf(&report, "## All Tests Passed\n\nNo failures detected across %d scenarios.\n", len(scenarios))
	}

	return report.String(), nil
}

// generateScenarios uses the LLM to create test scenarios based on endpoint analysis.
func (t *AutoTestTool) generateScenarios(endpoint, baseURL string, analysis shared.EndpointAnalysis, ctx string) ([]shared.TestScenario, error) {
	prompt := fmt.Sprintf(`You are an API testing expert. Generate test scenarios for this endpoint.

Endpoint: %s
Base URL: %s
Analysis: %s
Context: %s

Return a JSON array of TestScenario objects covering:
- Happy path (valid inputs, expected 2xx)
- Missing required fields (expect 4xx)
- Invalid data types (expect 4xx)
- Boundary values (empty strings, max lengths)
- Auth: test with no auth token (expect 401) and wrong token (expect 401/403)

Each scenario must have: id, name, category, method, url (path only), headers (object), body (object or null), expected.status_code.
Return ONLY the JSON array, no markdown, no explanation.`, endpoint, baseURL, analysis.Summary, ctx)

	resp, err := t.llmClient.Chat([]llm.Message{
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	// Strip markdown code fences if present
	cleaned := strings.TrimSpace(resp)
	if strings.HasPrefix(cleaned, "```") {
		lines := strings.Split(cleaned, "\n")
		if len(lines) > 2 {
			cleaned = strings.Join(lines[1:len(lines)-1], "\n")
		}
	}

	var scenarios []shared.TestScenario
	if err := json.Unmarshal([]byte(cleaned), &scenarios); err != nil {
		return nil, fmt.Errorf("failed to parse LLM scenarios: %w (raw: %s)", err, cleaned[:min(200, len(cleaned))])
	}

	return scenarios, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
