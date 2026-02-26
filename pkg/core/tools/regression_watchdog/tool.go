package regression_watchdog

import (
	"encoding/json"
	"fmt"

	"github.com/blackcoderx/falcon/pkg/core/tools/shared"
)

// RegressionWatchdogTool compares current API behavior against a stored baseline.
type RegressionWatchdogTool struct {
	falconDir string
	httpTool  *shared.HTTPTool
}

// NewRegressionWatchdogTool creates a new regression watchdog tool.
func NewRegressionWatchdogTool(falconDir string, httpTool *shared.HTTPTool) *RegressionWatchdogTool {
	return &RegressionWatchdogTool{
		falconDir: falconDir,
		httpTool: httpTool,
	}
}

// RegressionParams defines parameters for checking regressions.
type RegressionParams struct {
	BaseURL      string   `json:"base_url"`                   // Current API URL
	BaselineName string   `json:"baseline_name"`              // Name of the snapshot to compare against
	Endpoints    []string `json:"endpoints,omitempty"`        // Specific endpoints to verify
	SaveBaseline bool     `json:"save_as_baseline,omitempty"` // Whether to update the baseline after check
}

// RegressionResult represents the outcome of the comparison.
type RegressionResult struct {
	BaselineDate string       `json:"baseline_date"`
	Regressions  []Regression `json:"regressions"`
	StableCount  int          `json:"stable_count"`
	Summary      string       `json:"summary"`
}

// Regression represents a detected behavioral change.
type Regression struct {
	Endpoint    string `json:"endpoint"`
	ChangeType  string `json:"change_type"` // status_code, body_diff, response_time
	Description string `json:"description"`
}

func (t *RegressionWatchdogTool) Name() string {
	return "check_regression"
}

func (t *RegressionWatchdogTool) Description() string {
	return "Verify current API behavior against a saved baseline snapshot (regression testing). Detects changes in status codes, JSON structure, and performance drift."
}

func (t *RegressionWatchdogTool) Parameters() string {
	return `{
  "base_url": "http://localhost:3000",
  "baseline_name": "stable_v1",
  "endpoints": ["GET /api/users"]
}`
}

func (t *RegressionWatchdogTool) Execute(args string) (string, error) {
	var params RegressionParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	if params.BaseURL == "" || params.BaselineName == "" {
		return "", fmt.Errorf("base_url and baseline_name are required")
	}

	store := NewBaselineStore(t.falconDir)
	baseline, err := store.Load(params.BaselineName)
	if err != nil {
		return "", fmt.Errorf("failed to load baseline: %w", err)
	}

	diffEngine := &DiffEngine{httpTool: t.httpTool, baseURL: params.BaseURL}
	result := diffEngine.Check(baseline, params.Endpoints)

	if params.SaveBaseline && len(result.Regressions) == 0 {
		// Logic to save current state as new baseline
	}

	result.Summary = t.formatSummary(result)

	return result.Summary, nil
}

func (t *RegressionWatchdogTool) formatSummary(r RegressionResult) string {
	summary := "🐕 Regression Watchdog Results\n\n"
	summary += fmt.Sprintf("Comparing against baseline from: %s\n", r.BaselineDate)
	summary += fmt.Sprintf("Stable Endpoints: %d\n", r.StableCount)
	summary += fmt.Sprintf("Regressions Found: %d\n\n", len(r.Regressions))

	if len(r.Regressions) > 0 {
		summary += "Detected Regressions:\n"
		for _, reg := range r.Regressions {
			summary += fmt.Sprintf("  ❌ %s: %s (%s)\n", reg.Endpoint, reg.Description, reg.ChangeType)
		}
	} else if r.StableCount > 0 {
		summary += "✓ No regressions detected. API behavior is consistent with the baseline."
	}

	return summary
}
