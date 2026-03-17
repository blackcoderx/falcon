package data_driven_engine

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/blackcoderx/falcon/pkg/core/tools/shared"
)

// DataDrivenEngineTool executes test scenarios using data from external sources.
type DataDrivenEngineTool struct {
	falconDir    string
	httpTool     *shared.HTTPTool
	testExecutor *shared.TestExecutor
	reportWriter *shared.ReportWriter
}

// NewDataDrivenEngineTool creates a new data-driven engine tool.
func NewDataDrivenEngineTool(falconDir string, httpTool *shared.HTTPTool, testExecutor *shared.TestExecutor, reportWriter *shared.ReportWriter) *DataDrivenEngineTool {
	return &DataDrivenEngineTool{
		falconDir:    falconDir,
		httpTool:     httpTool,
		testExecutor: testExecutor,
		reportWriter: reportWriter,
	}
}

// DataDrivenParams defines parameters for data-driven testing.
type DataDrivenParams struct {
	Scenario   shared.TestScenario `json:"scenario"`              // Base scenario template
	DataSource string              `json:"data_source"`           // Path to CSV/JSON file or 'fake'
	Variables  []string            `json:"variables"`             // Variable names to map
	MaxRows    int                 `json:"max_rows,omitempty"`    // Limit number of rows to process
	ReportName string              `json:"report_name,omitempty"` // e.g. "data_driven_report_users"
}

// DataDrivenResult represents the outcome of the data-driven test run.
type DataDrivenResult struct {
	TotalRows  int                 `json:"total_rows"`
	PassedRows int                 `json:"passed_rows"`
	FailedRows int                 `json:"failed_rows"`
	Results    []shared.TestResult `json:"results"`
	Summary    string              `json:"summary"`
}

func (t *DataDrivenEngineTool) Name() string {
	return "run_data_driven"
}

func (t *DataDrivenEngineTool) Description() string {
	return "Execute test scenarios driven by external data sources (CSV/JSON) or automated data generators, mapping variables to request templates"
}

func (t *DataDrivenEngineTool) Parameters() string {
	return `{
  "scenario": {
    "method": "POST",
    "url": "http://localhost:3000/api/users",
    "body": {"name": "{{name}}", "email": "{{email}}"}
  },
  "data_source": "./data/users.csv",
  "variables": ["name", "email"]
}`
}

func (t *DataDrivenEngineTool) Execute(args string) (string, error) {
	var params DataDrivenParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	// 1. Load data
	loader := &DataLoader{Source: params.DataSource}
	rows, err := loader.Load(params.Variables, params.MaxRows)
	if err != nil {
		return "", fmt.Errorf("failed to load data: %w", err)
	}

	// 2. Process rows using TestExecutor
	tempEngine := &TemplateEngine{}
	var results []shared.TestResult
	passed := 0

	for i, row := range rows {
		populated := tempEngine.Populate(params.Scenario, row)
		populated.ID = fmt.Sprintf("%s_row_%d", params.Scenario.ID, i)

		// Use TestExecutor for scenario execution (empty baseURL since URLs are fully qualified)
		result := t.testExecutor.RunScenario(populated, "")
		result.ScenarioName = fmt.Sprintf("%s (Row %d)", params.Scenario.Name, i)

		if result.Passed {
			passed++
		}
		results = append(results, result)
	}

	result := DataDrivenResult{
		TotalRows:  len(rows),
		PassedRows: passed,
		FailedRows: len(rows) - passed,
		Results:    results,
	}
	result.Summary = t.formatSummary(result)

	reportContent := formatDataDrivenReport(result)
	reportPath, err := t.reportWriter.Write(params.ReportName, "data_driven_report", reportContent)
	if err != nil {
		return result.Summary + fmt.Sprintf("\n\nWarning: failed to save report: %v", err), nil
	}

	return result.Summary + fmt.Sprintf("\n\nReport saved to: %s", reportPath), nil
}

// formatDataDrivenReport builds the Markdown content for a data-driven report.
func formatDataDrivenReport(result DataDrivenResult) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "# Data-Driven Test Report\n\n")
	fmt.Fprintf(&sb, "**Generated:** %s\n\n", time.Now().Format(time.RFC1123))
	fmt.Fprintf(&sb, "## Summary\n\n")
	fmt.Fprintf(&sb, "| Metric | Value |\n|--------|-------|\n")
	fmt.Fprintf(&sb, "| Total Rows | %d |\n", result.TotalRows)
	fmt.Fprintf(&sb, "| Passed | %d |\n", result.PassedRows)
	fmt.Fprintf(&sb, "| Failed | %d |\n\n", result.FailedRows)

	if len(result.Results) > 0 {
		fmt.Fprintf(&sb, "## Row Results\n\n")
		for _, res := range result.Results {
			status := "PASS"
			if !res.Passed {
				status = "FAIL"
			}
			fmt.Fprintf(&sb, "### [%s] %s — %s\n\n", res.ScenarioID, res.ScenarioName, status)
			fmt.Fprintf(&sb, "- **Duration:** %dms\n", res.DurationMs)
			fmt.Fprintf(&sb, "- **Status Code:** %d\n", res.ActualStatus)
			if res.Error != "" {
				fmt.Fprintf(&sb, "- **Error:** %s\n", res.Error)
			}
			fmt.Fprintf(&sb, "\n")
		}
	}

	return sb.String()
}

func (t *DataDrivenEngineTool) formatSummary(r DataDrivenResult) string {
	summary := "📊 Data-Driven Test Run Complete\n\n"
	summary += fmt.Sprintf("Total Rows: %d\n", r.TotalRows)
	summary += fmt.Sprintf("Passed:     %d\n", r.PassedRows)
	summary += fmt.Sprintf("Failed:     %d\n\n", r.FailedRows)

	if r.FailedRows > 0 {
		summary += "Failed Rows Details:\n"
		count := 0
		for _, res := range r.Results {
			if !res.Passed {
				summary += fmt.Sprintf("  • %s failed.\n", res.ScenarioName)
				count++
				if count >= 5 {
					break
				}
			}
		}
	}

	return summary
}
