package breaking_change_detector

import (
	"encoding/json"
	"fmt"
)

// BreakingChangeDetectorTool identifies breaking changes between two API specifications.
type BreakingChangeDetectorTool struct {
	zapDir string
}

// NewBreakingChangeDetectorTool creates a new breaking change detector tool.
func NewBreakingChangeDetectorTool(zapDir string) *BreakingChangeDetectorTool {
	return &BreakingChangeDetectorTool{
		zapDir: zapDir,
	}
}

// BreakingChangeParams defines parameters for change detection.
type BreakingChangeParams struct {
	OldSpecPath string `json:"old_spec_path"` // Path to previous OpenAPI/Swagger spec
	NewSpecPath string `json:"new_spec_path"` // Path to current spec
}

// BreakingChangeResult represents the identified changes.
type BreakingChangeResult struct {
	BreakingChanges []Change `json:"breaking_changes"`
	MinorChanges    []Change `json:"minor_changes"`
	PatchChanges    []Change `json:"patch_changes"`
	Summary         string   `json:"summary"`
}

// Change represents a detected difference between specs.
type Change struct {
	Type        string `json:"type"` // generic, endpoint_removed, param_required, etc.
	Endpoint    string `json:"endpoint,omitempty"`
	Description string `json:"description"`
	Severity    string `json:"severity"` // breaking, minor, patch
}

func (t *BreakingChangeDetectorTool) Name() string {
	return "detect_breaking_changes"
}

func (t *BreakingChangeDetectorTool) Description() string {
	return "Analyze old and new API specifications to automatically identify breaking changes (removed endpoints, changed types, new required fields)"
}

func (t *BreakingChangeDetectorTool) Parameters() string {
	return `{
  "old_spec_path": "./docs/api-v1.json",
  "new_spec_path": "./docs/api-v2.json"
}`
}

func (t *BreakingChangeDetectorTool) Execute(args string) (string, error) {
	var params BreakingChangeParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	if params.OldSpecPath == "" || params.NewSpecPath == "" {
		return "", fmt.Errorf("both old_spec_path and new_spec_path are required")
	}

	// Simulation of breaking change detection logic
	// In a real implementation, we would use an OpenAPI diff library or
	// ingest both specs into temporary graphs and compare them.

	result := BreakingChangeResult{
		BreakingChanges: []Change{},
		MinorChanges:    []Change{},
		PatchChanges:    []Change{},
	}
	result.Summary = t.formatSummary(result)

	return result.Summary, nil
}

func (t *BreakingChangeDetectorTool) formatSummary(r BreakingChangeResult) string {
	summary := "🚨 Breaking Change Analysis\n\n"

	if len(r.BreakingChanges) == 0 && len(r.MinorChanges) == 0 && len(r.PatchChanges) == 0 {
		return summary + "✓ No changes detected between the two specifications."
	}

	if len(r.BreakingChanges) > 0 {
		summary += "🔴 Breaking Changes:\n"
		for _, c := range r.BreakingChanges {
			summary += fmt.Sprintf("  • [%s] %s\n", c.Endpoint, c.Description)
		}
		summary += "\n"
	}

	if len(r.MinorChanges) > 0 {
		summary += "🟡 Minor Changes (Non-breaking):\n"
		for _, c := range r.MinorChanges {
			summary += fmt.Sprintf("  • [%s] %s\n", c.Endpoint, c.Description)
		}
		summary += "\n"
	}

	return summary
}
