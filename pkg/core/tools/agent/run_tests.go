package agent

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/blackcoderx/falcon/pkg/core/tools/shared"
)

// RunTestsTool executes multiple test scenarios
type RunTestsTool struct {
	falconDir    string
	reportWriter *shared.ReportWriter
	testExecutor *shared.TestExecutor
}

// NewRunTestsTool creates a new run_tests tool
func NewRunTestsTool(falconDir string, testExecutor *shared.TestExecutor, reportWriter *shared.ReportWriter) *RunTestsTool {
	return &RunTestsTool{
		falconDir:    falconDir,
		reportWriter: reportWriter,
		testExecutor: testExecutor,
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
	ReportName  string                `json:"report_name,omitempty"` // e.g. "test_report_users_api"
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

	results := t.testExecutor.RunScenarios(scenariosToRun, params.BaseURL, concurrency)

	// Summarize
	passed := 0
	failed := 0
	var sb strings.Builder
	fmt.Fprintf(&sb, "Test Results (%d tests):\n\n", len(results))

	for _, res := range results {
		statusIcon := "✓"
		if !res.Passed {
			statusIcon = "✗"
			failed++
		} else {
			passed++
		}
		fmt.Fprintf(&sb, "%s [%s] %s (%dms)\n", statusIcon, res.Category, res.ScenarioName, res.DurationMs)
		if !res.Passed {
			fmt.Fprintf(&sb, "  Error: %s\n", res.Error)
			if res.ExpectedStatus != 0 {
				fmt.Fprintf(&sb, "  Status: Expected %d, Got %d\n", res.ExpectedStatus, res.ActualStatus)
			}
		}
	}

	fmt.Fprintf(&sb, "\nSummary: %d Passed, %d Failed\n", passed, failed)
	summary := sb.String()

	reportContent := formatTestReport(results, passed, failed)
	reportPath, err := t.reportWriter.Write(params.ReportName, "test_report", reportContent)
	if err != nil {
		return summary + fmt.Sprintf("\n\nWarning: failed to save report: %v", err), nil
	}

	return summary + fmt.Sprintf("\n\nReport saved to: %s", reportPath), nil
}

// formatTestReport builds the Markdown content for a test report.
func formatTestReport(results []shared.TestResult, passed, failed int) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "# Test Results Report\n\n")
	fmt.Fprintf(&sb, "**Generated:** %s\n\n", time.Now().Format(time.RFC1123))
	fmt.Fprintf(&sb, "## Summary\n\n")
	fmt.Fprintf(&sb, "| Metric | Value |\n|--------|-------|\n")
	fmt.Fprintf(&sb, "| Total  | %d |\n", len(results))
	fmt.Fprintf(&sb, "| Passed | %d |\n", passed)
	fmt.Fprintf(&sb, "| Failed | %d |\n\n", failed)

	fmt.Fprintf(&sb, "## Results\n\n")
	for _, res := range results {
		status := "PASS"
		if !res.Passed {
			status = "FAIL"
		}
		fmt.Fprintf(&sb, "### [%s] %s — %s\n\n", res.ScenarioID, res.ScenarioName, status)
		fmt.Fprintf(&sb, "- **Category:** %s\n", res.Category)
		fmt.Fprintf(&sb, "- **Duration:** %dms\n", res.DurationMs)
		fmt.Fprintf(&sb, "- **Status Code:** %d\n", res.ActualStatus)
		if res.Error != "" {
			fmt.Fprintf(&sb, "- **Error:** %s\n", res.Error)
		}
		fmt.Fprintf(&sb, "\n")
	}

	return sb.String()
}
