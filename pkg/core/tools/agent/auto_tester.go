package agent

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/blackcoderx/zap/pkg/core/tools/debugging"
	"github.com/blackcoderx/zap/pkg/core/tools/shared"
)

// AutoTestTool orchestrates the full autonomous testing workflow
type AutoTestTool struct {
	analyzeTool        *debugging.AnalyzeEndpointTool
	generateTool       *debugging.GenerateTestsTool
	orchestrateTool    *RunTestsTool
	analyzeFailureTool *debugging.AnalyzeFailureTool
}

func NewAutoTestTool(a *debugging.AnalyzeEndpointTool, g *debugging.GenerateTestsTool, o *RunTestsTool, f *debugging.AnalyzeFailureTool) *AutoTestTool {
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

	analyzeParams := debugging.AnalyzeEndpointParams{
		EndpointDescription: params.Endpoint,
		Method:              method,
		URL:                 url,
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

	// 2. Generate
	genParams := debugging.GenerateTestsParams{
		Analysis: analysis,
		Count:    20,
	}
	genJSON, err := json.Marshal(genParams)
	if err != nil {
		return "", fmt.Errorf("failed to marshal generate params: %w", err)
	}
	genResult, err := t.generateTool.Execute(string(genJSON))
	if err != nil {
		return "", fmt.Errorf("generation failed: %w", err)
	}

	var scenarios []shared.TestScenario
	if err := json.Unmarshal([]byte(genResult), &scenarios); err != nil {
		return "", fmt.Errorf("failed to parse generated scenarios: %w", err)
	}

	// 3. Run
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

		failParams := debugging.AnalyzeFailureParams{
			TestResult:       res,
			ResponseBody:     res.ResponseBody,
			ExpectedBehavior: fmt.Sprintf("Status %d", res.ExpectedStatus),
		}
		var failAnalysis string
		failJSON, err := json.Marshal(failParams)
		if err != nil {
			failAnalysis = fmt.Sprintf("Failed to marshal failure params: %v", err)
		} else {
			failAnalysis, err = t.analyzeFailureTool.Execute(string(failJSON))
			if err != nil {
				failAnalysis = fmt.Sprintf("Failed to analyze failure: %v", err)
			}
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
