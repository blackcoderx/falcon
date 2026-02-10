package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SecurityReportTool generates comprehensive analysis and action plans
type SecurityReportTool struct {
	zapDir string
}

func NewSecurityReportTool(zapDir string) *SecurityReportTool {
	return &SecurityReportTool{zapDir: zapDir}
}

type SecurityReportParams struct {
	Results  []TestResult `json:"results"`
	Endpoint string       `json:"endpoint"`
}

func (t *SecurityReportTool) Name() string {
	return "security_report"
}

func (t *SecurityReportTool) Description() string {
	return "Create comprehensive security report with vulnerabilities found, severity scores, OWASP mappings, and prioritized action plan."
}

func (t *SecurityReportTool) Parameters() string {
	return `{
  "results": [ /* array of TestResult */ ],
  "endpoint": "POST /api/checkout"
}`
}

func (t *SecurityReportTool) Execute(args string) (string, error) {
	var params SecurityReportParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	// 1. Aggregate stats
	total := len(params.Results)
	passed := 0
	failed := 0
	severityCounts := make(map[string]int)

	for _, res := range params.Results {
		if res.Passed {
			passed++
		} else {
			failed++
			severity := strings.ToLower(res.Severity)
			if severity == "" {
				severity = "medium" // Default
			}
			severityCounts[severity]++
		}
	}

	// 2. Build markdown report
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Security Report: %s\n", params.Endpoint))
	sb.WriteString(fmt.Sprintf("Date: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	sb.WriteString("## Summary\n")
	sb.WriteString(fmt.Sprintf("- **Total Tests**: %d\n", total))
	sb.WriteString(fmt.Sprintf("- **Passed**: %d\n", passed))
	sb.WriteString(fmt.Sprintf("- **Failed**: %d\n", failed))
	sb.WriteString("\n")

	sb.WriteString("## Severity Breakdown\n")
	sb.WriteString(fmt.Sprintf("- 🔴 **Critical**: %d\n", severityCounts["critical"]))
	sb.WriteString(fmt.Sprintf("- 🟠 **High**: %d\n", severityCounts["high"]))
	sb.WriteString(fmt.Sprintf("- 🟡 **Medium**: %d\n", severityCounts["medium"]))
	sb.WriteString(fmt.Sprintf("- 🔵 **Low**: %d\n", severityCounts["low"]))
	sb.WriteString("\n")

	sb.WriteString("## Vulnerabilities Found\n")
	if failed == 0 {
		sb.WriteString("No vulnerabilities identified in this session. 🎉\n")
	} else {
		for _, res := range params.Results {
			if !res.Passed {
				sb.WriteString(fmt.Sprintf("### [%s] %s\n", res.ScenarioID, res.ScenarioName))
				sb.WriteString(fmt.Sprintf("- **Severity**: %s\n", res.Severity))
				if res.OWASPRef != "" {
					sb.WriteString(fmt.Sprintf("- **OWASP**: %s\n", res.OWASPRef))
				}
				sb.WriteString(fmt.Sprintf("- **Error**: %s\n", res.Error))
				sb.WriteString("\n")
			}
		}
	}

	return sb.String(), nil
}

// ExportResultsTool handles exporting data to various formats
type ExportResultsTool struct {
	zapDir string
}

func NewExportResultsTool(zapDir string) *ExportResultsTool {
	return &ExportResultsTool{zapDir: zapDir}
}

type ExportResultsParams struct {
	Results    []TestResult `json:"results"`
	Format     string       `json:"format"` // json, markdown
	OutputPath string       `json:"output_path,omitempty"`
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
  "output_path": ".zap/reports/report.json"
}`
}

func (t *ExportResultsTool) Execute(args string) (string, error) {
	var params ExportResultsParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	path := params.OutputPath
	if path == "" {
		reportsDir := filepath.Join(t.zapDir, "reports")
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
