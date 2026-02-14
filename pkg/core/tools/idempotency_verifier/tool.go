package idempotency_verifier

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/blackcoderx/zap/pkg/core/tools/shared"
	"github.com/blackcoderx/zap/pkg/core/tools/spec_ingester"
)

// IdempotencyVerifierTool checks if API endpoints are correctly idempotent.
type IdempotencyVerifierTool struct {
	zapDir   string
	httpTool *shared.HTTPTool
}

// NewIdempotencyVerifierTool creates a new idempotency verifier tool.
func NewIdempotencyVerifierTool(zapDir string, httpTool *shared.HTTPTool) *IdempotencyVerifierTool {
	return &IdempotencyVerifierTool{
		zapDir:   zapDir,
		httpTool: httpTool,
	}
}

// IdempotencyParams defines parameters for idempotency verification.
type IdempotencyParams struct {
	BaseURL     string   `json:"base_url"`               // Base URL of the API
	Endpoints   []string `json:"endpoints,omitempty"`    // Specific endpoints to verify
	RepeatCount int      `json:"repeat_count,omitempty"` // How many times to repeat the request (default: 2)
	IncludeGET  bool     `json:"include_get,omitempty"`  // Whether to verify GET/HEAD (usually idempotent by default)
}

// IdempotencyResult represents the outcome of the verification.
type IdempotencyResult struct {
	TotalVerified   int         `json:"total_verified"`
	IdempotentCount int         `json:"idempotent_count"`
	Violations      []Violation `json:"violations"`
	Summary         string      `json:"summary"`
}

// Violation represents a case where idempotency was broken.
type Violation struct {
	Endpoint    string `json:"endpoint"`
	Description string `json:"description"`
	Diff        string `json:"diff,omitempty"`
}

func (t *IdempotencyVerifierTool) Name() string {
	return "verify_idempotency"
}

func (t *IdempotencyVerifierTool) Description() string {
	return "Verify that API endpoints are idempotent by repeating requests and detecting side effects, state changes, or duplicate record creation"
}

func (t *IdempotencyVerifierTool) Parameters() string {
	return `{
  "base_url": "http://localhost:3000",
  "endpoints": ["POST /api/orders"],
  "repeat_count": 3
}`
}

func (t *IdempotencyVerifierTool) Execute(args string) (string, error) {
	var params IdempotencyParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	if params.BaseURL == "" {
		return "", fmt.Errorf("base_url is required")
	}

	if params.RepeatCount <= 0 {
		params.RepeatCount = 2
	}

	// 1. Get endpoints to verify
	endpoints, err := t.getEndpoints(params.Endpoints, params.IncludeGET)
	if err != nil {
		return "", err
	}

	// 2. Run repeat engine
	engine := &RepeatEngine{
		httpTool:    t.httpTool,
		baseURL:     params.BaseURL,
		repeatCount: params.RepeatCount,
	}

	result := engine.Verify(endpoints)
	result.Summary = t.formatSummary(result)

	return result.Summary, nil
}

func (t *IdempotencyVerifierTool) getEndpoints(specified []string, includeGET bool) (map[string]shared.EndpointAnalysis, error) {
	if len(specified) > 0 {
		endpoints := make(map[string]shared.EndpointAnalysis)
		for _, ep := range specified {
			endpoints[ep] = shared.EndpointAnalysis{}
		}
		return endpoints, nil
	}

	builder := spec_ingester.NewGraphBuilder(t.zapDir)
	graph, err := builder.LoadGraph()
	if err != nil {
		return nil, fmt.Errorf("failed to load API Knowledge Graph: %w", err)
	}

	filtered := make(map[string]shared.EndpointAnalysis)
	for epKey, analysis := range graph.Endpoints {
		method := "GET"
		if parts := strings.SplitN(epKey, " ", 2); len(parts) == 2 {
			method = parts[0]
		}

		// By default check non-safe/non-idempotent methods: POST, (PUT/PATCH/DELETE often idempotent but worth checking)
		if includeGET || (method != "GET" && method != "HEAD" && method != "OPTIONS") {
			filtered[epKey] = analysis
		}
	}
	return filtered, nil
}

func (t *IdempotencyVerifierTool) formatSummary(r IdempotencyResult) string {
	summary := "♻️ Idempotency Verification Results\n\n"
	summary += fmt.Sprintf("Total Verified: %d\n", r.TotalVerified)
	summary += fmt.Sprintf("Idempotent:     %d\n", r.IdempotentCount)
	summary += fmt.Sprintf("Violations:     %d\n\n", len(r.Violations))

	if len(r.Violations) > 0 {
		summary += "Vulnerabilities/Violations Found:\n"
		for _, v := range r.Violations {
			summary += fmt.Sprintf("  ❌ %s: %s\n", v.Endpoint, v.Description)
		}
	} else if r.TotalVerified > 0 {
		summary += "✓ All checked endpoints correctly implemented idempotency.\n"
	}

	return summary
}
