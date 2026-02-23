package data_driven_engine

import (
	"encoding/json"
	"fmt"

	"github.com/blackcoderx/falcon/pkg/core/tools/shared"
)

// DataDrivenEngineTool executes test scenarios using data from external sources.
type DataDrivenEngineTool struct {
	httpTool *shared.HTTPTool
}

// NewDataDrivenEngineTool creates a new data-driven engine tool.
func NewDataDrivenEngineTool(httpTool *shared.HTTPTool) *DataDrivenEngineTool {
	return &DataDrivenEngineTool{
		httpTool: httpTool,
	}
}

// DataDrivenParams defines parameters for data-driven testing.
type DataDrivenParams struct {
	Scenario   shared.TestScenario `json:"scenario"`           // Base scenario template
	DataSource string              `json:"data_source"`        // Path to CSV/JSON file or 'fake'
	Variables  []string            `json:"variables"`          // Variable names to map
	MaxRows    int                 `json:"max_rows,omitempty"` // Limit number of rows to process
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

	_ = result
	return result.Summary, nil
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
