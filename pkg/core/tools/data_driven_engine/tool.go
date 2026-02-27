package data_driven_engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/blackcoderx/falcon/pkg/core/tools/shared"
)

// DataDrivenEngineTool executes test scenarios using data from external sources.
type DataDrivenEngineTool struct {
	falconDir string
	httpTool  *shared.HTTPTool
}

// NewDataDrivenEngineTool creates a new data-driven engine tool.
func NewDataDrivenEngineTool(falconDir string, httpTool *shared.HTTPTool) *DataDrivenEngineTool {
	return &DataDrivenEngineTool{
		falconDir: falconDir,
		httpTool:  httpTool,
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

	// 2. Process rows
	tempEngine := &TemplateEngine{}
	var results []shared.TestResult
	passed := 0

	for i, row := range rows {
		// Populate scenario with row data
		populated := tempEngine.Populate(params.Scenario, row)
		populated.ID = fmt.Sprintf("%s_row_%d", params.Scenario.ID, i)

		// Execute (Simulation/Simplified)
		req := shared.HTTPRequest{
			Method: populated.Method,
			URL:    populated.URL,
			Body:   populated.Body,
		}

		resp, err := t.httpTool.Run(req)

		res := shared.TestResult{
			ScenarioID:   populated.ID,
			ScenarioName: fmt.Sprintf("%s (Row %d)", params.Scenario.Name, i),
			Passed:       err == nil && resp.StatusCode < 400,
		}
		if res.Passed {
			passed++
		}

		results = append(results, res)
	}

	result := DataDrivenResult{
		TotalRows:  len(rows),
		PassedRows: passed,
		FailedRows: len(rows) - passed,
		Results:    results,
	}
	result.Summary = t.formatSummary(result)

	reportPath, err := generateDataDrivenReport(t.falconDir, params.ReportName, result)
	if err != nil {
		return result.Summary + fmt.Sprintf("\n\nWarning: failed to save report: %v", err), nil
	}

	return result.Summary + fmt.Sprintf("\n\nReport saved to: %s", reportPath), nil
}

// generateDataDrivenReport writes data-driven test results to a Markdown file in .falcon/reports/.
func generateDataDrivenReport(falconDir, reportName string, result DataDrivenResult) (string, error) {
	reportsDir := filepath.Join(falconDir, "reports")
	if err := os.MkdirAll(reportsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create reports directory: %w", err)
	}

	name := reportName
	if name == "" {
		name = fmt.Sprintf("data_driven_report_%s", time.Now().Format("20060102_150405"))
	}
	name = strings.ReplaceAll(name, " ", "_")
	if !strings.HasSuffix(name, ".md") {
		name += ".md"
	}
	reportPath := filepath.Join(reportsDir, name)

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

	if err := os.WriteFile(reportPath, []byte(sb.String()), 0644); err != nil {
		return "", fmt.Errorf("failed to write report: %w", err)
	}

	if err := shared.ValidateReport(reportPath); err != nil {
		return "", err
	}

	return reportPath, nil
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
