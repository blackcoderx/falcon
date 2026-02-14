package performance_engine

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/blackcoderx/zap/pkg/core/tools/shared"
	"github.com/blackcoderx/zap/pkg/core/tools/spec_ingester"
)

// PerformanceEngineTool provides multi-mode load testing and metrics tracking.
type PerformanceEngineTool struct {
	zapDir   string
	httpTool *shared.HTTPTool
}

// NewPerformanceEngineTool creates a new performance engine tool.
func NewPerformanceEngineTool(zapDir string, httpTool *shared.HTTPTool) *PerformanceEngineTool {
	return &PerformanceEngineTool{
		zapDir:   zapDir,
		httpTool: httpTool,
	}
}

// PerformanceParams defines parameters for performance testing.
type PerformanceParams struct {
	Mode        string   `json:"mode"`                   // load, stress, spike, soak
	BaseURL     string   `json:"base_url"`               // Base URL of the API
	Endpoints   []string `json:"endpoints,omitempty"`    // Specific endpoints to test
	Concurrency int      `json:"concurrency,omitempty"`  // Number of concurrent virtual users (default: 10)
	Duration    int      `json:"duration_sec,omitempty"` // Duration of test in seconds (default: 30)
	RPS         int      `json:"rps,omitempty"`          // Target requests per second (optional)
}

// PerformanceResult holds the aggregated metrics of a test run.
type PerformanceResult struct {
	Mode          string           `json:"mode"`
	TotalRequests int              `json:"total_requests"`
	SuccessRate   float64          `json:"success_rate"`
	Metrics       ExecutionMetrics `json:"metrics"`
	StartTime     time.Time        `json:"start_time"`
	Duration      time.Duration    `json:"duration"`
	Summary       string           `json:"summary"`
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
  "rps": 50
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

	// Initialize test runner
	runner := NewLoadTestRunner(t.httpTool, params)

	// Execute test
	startTime := time.Now()
	metrics := runner.Run(endpoints)
	duration := time.Since(startTime)

	// Build result
	result := PerformanceResult{
		Mode:          params.Mode,
		TotalRequests: metrics.Total,
		SuccessRate:   metrics.SuccessRate,
		Metrics:       metrics,
		StartTime:     startTime,
		Duration:      duration,
		Summary:       metrics.FormatSummary(params.Mode),
	}
	_ = result // Suppress unused write to field info lint
	return result.Summary, nil
}

func (t *PerformanceEngineTool) getEndpoints(specified []string) (map[string]shared.EndpointAnalysis, error) {
	if len(specified) > 0 {
		endpoints := make(map[string]shared.EndpointAnalysis)
		for _, ep := range specified {
			endpoints[ep] = shared.EndpointAnalysis{Summary: "User specified"}
		}
		return endpoints, nil
	}

	builder := spec_ingester.NewGraphBuilder(t.zapDir)
	graph, err := builder.LoadGraph()
	if err != nil {
		return nil, err
	}
	if graph == nil || len(graph.Endpoints) == 0 {
		return nil, fmt.Errorf("no endpoints in Knowledge Graph - run ingest_spec first")
	}

	return graph.Endpoints, nil
}
