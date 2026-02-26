package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/blackcoderx/falcon/pkg/core/tools/shared"
)

// ExportResultsTool handles exporting data to various formats
type ExportResultsTool struct {
	falconDir string
}

func NewExportResultsTool(falconDir string) *ExportResultsTool {
	return &ExportResultsTool{falconDir: falconDir}
}

type ExportResultsParams struct {
	Results    []shared.TestResult `json:"results"`
	Format     string              `json:"format"` // json, markdown
	OutputPath string              `json:"output_path,omitempty"`
}

func (t *ExportResultsTool) Name() string {
	return "export_results"
}

func (t *ExportResultsTool) Description() string {
	return "Export test results in multiple formats (JSON, Markdown)."
}

func (t *ExportResultsTool) Parameters() string {
	return `{
  "results": [ /* TestResult array */ ],
  "format": "json|markdown",
  "output_path": ".falcon/reports/report.json"
}`
}

func (t *ExportResultsTool) Execute(args string) (string, error) {
	var params ExportResultsParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	path := params.OutputPath
	if path == "" {
		reportsDir := filepath.Join(t.falconDir, "reports")
		_ = os.MkdirAll(reportsDir, 0755)
		filename := fmt.Sprintf("report-%d.%s", time.Now().Unix(), params.Format)
		path = filepath.Join(reportsDir, filename)
	}

	var data []byte
	var err error

	switch strings.ToLower(params.Format) {
	case "json":
		data, err = json.MarshalIndent(params.Results, "", "  ")
	case "markdown":
		// Simple markdown export
		var sb strings.Builder
		sb.WriteString("# API Test Results\n\n")
		for _, res := range params.Results {
			status := "PASS"
			if !res.Passed {
				status = "FAIL"
			}
			sb.WriteString(fmt.Sprintf("## [%s] %s - %s\n", res.ScenarioID, res.ScenarioName, status))
			sb.WriteString(fmt.Sprintf("- **Duration**: %dms\n", res.DurationMs))
			if res.Error != "" {
				sb.WriteString(fmt.Sprintf("- **Error**: %s\n", res.Error))
			}
			sb.WriteString("\n")
		}
		data = []byte(sb.String())
	default:
		return "", fmt.Errorf("unsupported format: %s", params.Format)
	}

	if err != nil {
		return "", fmt.Errorf("failed to encode data: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fmt.Sprintf("Successfully exported to %s", path), nil
}
