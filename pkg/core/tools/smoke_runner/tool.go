package smoke_runner

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/blackcoderx/falcon/pkg/core/tools/shared"
	"github.com/blackcoderx/falcon/pkg/core/tools/spec_ingester"
)

// SmokeRunnerTool performs fast health checks and core functionality verification.
type SmokeRunnerTool struct {
	falconDir string
	httpTool *shared.HTTPTool
}

// NewSmokeRunnerTool creates a new smoke runner tool.
func NewSmokeRunnerTool(falconDir string, httpTool *shared.HTTPTool) *SmokeRunnerTool {
	return &SmokeRunnerTool{
		falconDir: falconDir,
		httpTool: httpTool,
	}
}

// SmokeParams defines parameters for smoke testing.
type SmokeParams struct {
	BaseURL   string   `json:"base_url"`             // Base URL of the API
	Endpoints []string `json:"endpoints,omitempty"`  // Specific critical endpoints to check
	Timeout   int      `json:"timeout_ms,omitempty"` // Timeout per request
	Detailed  bool     `json:"detailed,omitempty"`   // Whether to provide detailed health diagnostics
}

// SmokeResult represents the outcome of a smoke test.
type SmokeResult struct {
	Status   string        `json:"status"` // pass, fail, partial
	Duration string        `json:"duration"`
	Checks   []HealthCheck `json:"checks"`
	Summary  string        `json:"summary"`
}

// HealthCheck represents a single health or reachability check.
type HealthCheck struct {
	Name     string `json:"name"`
	Endpoint string `json:"endpoint"`
	Status   string `json:"status"` // ok, error
	Latency  string `json:"latency"`
	Message  string `json:"message,omitempty"`
}

func (t *SmokeRunnerTool) Name() string {
	return "run_smoke"
}

func (t *SmokeRunnerTool) Description() string {
	return "Perform a fast smoke test to verify API reachability, health endpoints, and core functionality"
}

func (t *SmokeRunnerTool) Parameters() string {
	return `{
  "base_url": "http://localhost:3000",
  "endpoints": ["GET /health", "GET /api/v1/status"],
  "detailed": true
}`
}

func (t *SmokeRunnerTool) Execute(args string) (string, error) {
	var params SmokeParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	if params.BaseURL == "" {
		return "", fmt.Errorf("base_url is required")
	}

	startTime := time.Now()

	// 1. Identify critical endpoints
	endpoints, err := t.getCriticalEndpoints(params.Endpoints)
	if err != nil {
		return "", err
	}

	// 2. Run reachability and health checks
	checks := t.runHealthChecks(params.BaseURL, endpoints)

	// 3. Determine overall status
	status := "pass"
	failedCount := 0
	for _, c := range checks {
		if c.Status != "ok" {
			failedCount++
		}
	}
	if failedCount > 0 {
		if failedCount == len(checks) {
			status = "fail"
		} else {
			status = "partial"
		}
	}

	duration := time.Since(startTime).String()

	result := SmokeResult{
		Status:   status,
		Duration: duration,
		Checks:   checks,
	}
	result.Summary = t.formatSummary(result)

	_ = result
	return result.Summary, nil
}

func (t *SmokeRunnerTool) getCriticalEndpoints(specified []string) ([]string, error) {
	if len(specified) > 0 {
		return specified, nil
	}

	// Try to find common health endpoints in the Knowledge Graph
	builder := spec_ingester.NewGraphBuilder(t.falconDir)
	graph, err := builder.LoadGraph()
	if err != nil {
		// If no graph, just return standard defaults
		return []string{"GET /health", "GET /status", "GET /ping"}, nil
	}

	var critical []string
	healthKeywords := []string{"health", "status", "ping", "version"}

	for epKey := range graph.Endpoints {
		_ = epKey // Suppress unused for now
		_ = healthKeywords
	}

	// For now return a few common ones + first 2 from graph if empty
	critical = append(critical, "GET /health")
	count := 0
	for epKey := range graph.Endpoints {
		if count >= 2 {
			break
		}
		critical = append(critical, epKey)
		count++
	}

	return critical, nil
}

func (t *SmokeRunnerTool) formatSummary(r SmokeResult) string {
	icon := "✅"
	switch r.Status {
case "fail":
		icon = "❌"
	case "partial":
		icon = "⚠️"
	}

	summary := fmt.Sprintf("%s Smoke Test: %s\n", icon, r.Status)
	summary += fmt.Sprintf("Duration: %s\n\n", r.Duration)
	summary += "Checks:\n"
	for _, c := range r.Checks {
		cIcon := "✓"
		if c.Status != "ok" {
			cIcon = "✗"
		}
		summary += fmt.Sprintf("  %s %s (%s) - %s\n", cIcon, c.Name, c.Latency, c.Status)
		if c.Message != "" {
			summary += fmt.Sprintf("    Message: %s\n", c.Message)
		}
	}
	return summary
}
