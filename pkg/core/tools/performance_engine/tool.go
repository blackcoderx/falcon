package performance_engine

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/blackcoderx/falcon/pkg/core/tools/shared"
	"github.com/blackcoderx/falcon/pkg/core/tools/spec_ingester"
)

// PerformanceEngineTool provides multi-mode load testing and metrics tracking.
type PerformanceEngineTool struct {
	falconDir    string
	httpTool     *shared.HTTPTool
	reportWriter *shared.ReportWriter
}

// NewPerformanceEngineTool creates a new performance engine tool.
func NewPerformanceEngineTool(falconDir string, httpTool *shared.HTTPTool, reportWriter *shared.ReportWriter) *PerformanceEngineTool {
	return &PerformanceEngineTool{
		falconDir:    falconDir,
		httpTool:     httpTool,
		reportWriter: reportWriter,
	}
}

// PerformanceParams defines parameters for performance testing.
type PerformanceParams struct {
	Mode        string   `json:"mode"`                  // load, stress, spike, soak
	BaseURL     string   `json:"base_url"`              // Base URL of the API
	Endpoints   []string `json:"endpoints,omitempty"`    // Specific endpoints to test
	Concurrency int      `json:"concurrency,omitempty"`  // Number of concurrent virtual users (default: 10)
	Duration    int      `json:"duration_sec,omitempty"` // Duration of test in seconds (default: 30)
	RPS         int      `json:"rps,omitempty"`          // Target requests per second (optional)
	ReportName  string   `json:"report_name,omitempty"`  // e.g. "performance_report_dummyjson_products"
}

// Name returns the tool name.
func (t *PerformanceEngineTool) Name() string {
	return "run_performance"
}

// Description returns the tool description.
func (t *PerformanceEngineTool) Description() string {
	return "Execute multi-mode performance tests (load, stress, spike, soak) and track high-resolution latency metrics"
}

// Parameters returns the tool parameter description.
func (t *PerformanceEngineTool) Parameters() string {
	return `{
  "mode": "load|stress|spike|soak",
  "base_url": "http://localhost:3000",
  "endpoints": ["GET /api/users"],
  "concurrency": 10,
  "duration_sec": 30,
  "rps": 50,
  "report_name": "performance_report_<api>_<resource>"
}`
}

// Execute performs the performance test.
func (t *PerformanceEngineTool) Execute(args string) (string, error) {
	var params PerformanceParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	if params.BaseURL == "" {
		return "", fmt.Errorf("base_url is required")
	}

	// Default mode is load
	if params.Mode == "" {
		params.Mode = "load"
	}

	// Get endpoints
	endpoints, err := t.getEndpoints(params.Endpoints)
	if err != nil {
		return "", fmt.Errorf("failed to get endpoints: %w", err)
	}

	runner := NewLoadTestRunner(t.httpTool, params)

	startTime := time.Now()
	metrics := runner.Run(endpoints)
	duration := time.Since(startTime)

	reportContent := formatPerformanceReport(params, metrics, startTime, duration)
	reportPath, err := t.reportWriter.Write(params.ReportName, "performance_report", reportContent)
	if err != nil {
		return metrics.FormatSummary(params.Mode) + fmt.Sprintf("\n\nWarning: failed to save report: %v", err), nil
	}

	return metrics.FormatSummary(params.Mode) + fmt.Sprintf("\n\nReport saved to: %s", reportPath), nil
}

// formatPerformanceReport builds the Markdown content for a performance report.
func formatPerformanceReport(params PerformanceParams, metrics ExecutionMetrics, startTime time.Time, duration time.Duration) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "# Performance Test Report\n\n")
	fmt.Fprintf(&sb, "**Date:** %s\n\n", startTime.Format(time.RFC1123))
	fmt.Fprintf(&sb, "**Target:** %s\n\n", params.BaseURL)
	fmt.Fprintf(&sb, "**Mode:** %s\n\n", params.Mode)

	fmt.Fprintf(&sb, "## Configuration\n\n")
	fmt.Fprintf(&sb, "| Parameter | Value |\n|-----------|-------|\n")
	fmt.Fprintf(&sb, "| Concurrency | %d virtual users |\n", params.Concurrency)
	fmt.Fprintf(&sb, "| Duration | %ds |\n", params.Duration)
	if params.RPS > 0 {
		fmt.Fprintf(&sb, "| Target RPS | %d |\n", params.RPS)
	}
	if len(params.Endpoints) > 0 {
		fmt.Fprintf(&sb, "| Endpoints | %s |\n", strings.Join(params.Endpoints, ", "))
	}
	fmt.Fprintf(&sb, "\n")

	fmt.Fprintf(&sb, "## Results\n\n")
	fmt.Fprintf(&sb, "| Metric | Value |\n|--------|-------|\n")
	fmt.Fprintf(&sb, "| Total Requests | %d |\n", metrics.Total)
	fmt.Fprintf(&sb, "| Successful | %d |\n", metrics.Success)
	fmt.Fprintf(&sb, "| Failed | %d |\n", metrics.Fail)
	fmt.Fprintf(&sb, "| Success Rate | %.2f%% |\n", metrics.SuccessRate)
	fmt.Fprintf(&sb, "| Test Duration | %v |\n\n", duration)

	fmt.Fprintf(&sb, "## Latency\n\n")
	fmt.Fprintf(&sb, "| Percentile | Latency |\n|------------|--------|\n")
	fmt.Fprintf(&sb, "| Avg | %v |\n", metrics.AvgLatency)
	fmt.Fprintf(&sb, "| Min | %v |\n", metrics.Min)
	fmt.Fprintf(&sb, "| p50 | %v |\n", metrics.P50)
	fmt.Fprintf(&sb, "| p95 | %v |\n", metrics.P95)
	fmt.Fprintf(&sb, "| p99 | %v |\n", metrics.P99)
	fmt.Fprintf(&sb, "| Max | %v |\n", metrics.Max)

	return sb.String()
}

func (t *PerformanceEngineTool) getEndpoints(specified []string) (map[string]shared.EndpointAnalysis, error) {
	if len(specified) > 0 {
		endpoints := make(map[string]shared.EndpointAnalysis)
		for _, ep := range specified {
			endpoints[ep] = shared.EndpointAnalysis{Summary: "User specified"}
		}
		return endpoints, nil
	}

	builder := spec_ingester.NewGraphBuilder(t.falconDir)
	graph, err := builder.LoadGraph()
	if err != nil {
		return nil, err
	}
	if graph == nil || len(graph.Endpoints) == 0 {
		return nil, fmt.Errorf("no endpoints in Knowledge Graph - run ingest_spec first")
	}

	return graph.Endpoints, nil
}
