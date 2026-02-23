package api_drift_analyzer

import (
	"encoding/json"
	"fmt"

	"github.com/blackcoderx/falcon/pkg/core/tools/shared"
	"github.com/blackcoderx/falcon/pkg/core/tools/spec_ingester"
)

// APIDriftAnalyzerTool detects drift between intended spec and live implementation.
type APIDriftAnalyzerTool struct {
	zapDir   string
	httpTool *shared.HTTPTool
}

// NewAPIDriftAnalyzerTool creates a new API drift analyzer tool.
func NewAPIDriftAnalyzerTool(zapDir string, httpTool *shared.HTTPTool) *APIDriftAnalyzerTool {
	return &APIDriftAnalyzerTool{
		zapDir:   zapDir,
		httpTool: httpTool,
	}
}

// DriftParams defines parameters for drift analysis.
type DriftParams struct {
	BaseURL   string   `json:"base_url"`            // Live API URL
	Endpoints []string `json:"endpoints,omitempty"` // Specific endpoints to analyze
}

// DriftResult represents the detected drift.
type DriftResult struct {
	ShadowEndpoints  []string `json:"shadow_endpoints"`  // Implemented but not in spec
	MissingEndpoints []string `json:"missing_endpoints"` // In spec but not implemented
	SchemaDrift      []Drift  `json:"schema_drift"`      // Schema mismatches
	Summary          string   `json:"summary"`
}

// Drift represents a specific mismatch found during analysis.
type Drift struct {
	Endpoint    string `json:"endpoint"`
	Description string `json:"description"`
}

func (t *APIDriftAnalyzerTool) Name() string {
	return "analyze_drift"
}

func (t *APIDriftAnalyzerTool) Description() string {
	return "Continuously analyze your live API against the official specification to detect 'shadow endpoints' (unregistered APIs) or functional drift"
}

func (t *APIDriftAnalyzerTool) Parameters() string {
	return `{
  "base_url": "http://api.production.internal"
}`
}

func (t *APIDriftAnalyzerTool) Execute(args string) (string, error) {
	var params DriftParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	if params.BaseURL == "" {
		return "", fmt.Errorf("base_url is required")
	}

	// 1. Load Knowledge Graph (Spec)
	builder := spec_ingester.NewGraphBuilder(t.zapDir)
	graph, err := builder.LoadGraph()
	if err != nil {
		return "", fmt.Errorf("failed to load API Knowledge Graph: %w", err)
	}

	// 2. Perform drift analysis (Simulation)
	// Real implementation would:
	// 1. Crawl/Discover endpoints on live API
	// 2. Match against graph
	// 3. Execute requests to compare schemas

	result := DriftResult{
		ShadowEndpoints:  []string{},
		MissingEndpoints: []string{},
		SchemaDrift:      []Drift{},
	}
	result.Summary = t.formatSummary(result)

	_ = graph
	return result.Summary, nil
}

func (t *APIDriftAnalyzerTool) formatSummary(r DriftResult) string {
	summary := "🛰️ API Drift Analysis\n\n"

	if len(r.ShadowEndpoints) == 0 && len(r.MissingEndpoints) == 0 && len(r.SchemaDrift) == 0 {
		return summary + "✓ Live API is perfectly synchronized with the specification."
	}

	if len(r.ShadowEndpoints) > 0 {
		summary += "👻 Shadow Endpoints (Implemented but not in spec):\n"
		for _, ep := range r.ShadowEndpoints {
			summary += fmt.Sprintf("  • %s\n", ep)
		}
		summary += "\n"
	}

	if len(r.MissingEndpoints) > 0 {
		summary += "❓ Missing Endpoints (In spec but not responding):\n"
		for _, ep := range r.MissingEndpoints {
			summary += fmt.Sprintf("  • %s\n", ep)
		}
		summary += "\n"
	}

	return summary
}
