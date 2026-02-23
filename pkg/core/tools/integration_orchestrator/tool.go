package integration_orchestrator

import (
	"encoding/json"
	"fmt"

	"github.com/blackcoderx/falcon/pkg/core/tools/shared"
)

// IntegrationOrchestratorTool handles multi-step API workflows and integration tests.
type IntegrationOrchestratorTool struct {
	zapDir   string
	httpTool *shared.HTTPTool
}

// NewIntegrationOrchestratorTool creates a new integration orchestrator tool.
func NewIntegrationOrchestratorTool(zapDir string, httpTool *shared.HTTPTool) *IntegrationOrchestratorTool {
	return &IntegrationOrchestratorTool{
		zapDir:   zapDir,
		httpTool: httpTool,
	}
}

// OrchestrateParams defines the parameters for a workflow orchestration.
type OrchestrateParams struct {
	Workflow      []WorkflowStep `json:"workflow"`           // List of steps to execute
	BaseURL       string         `json:"base_url,omitempty"` // Global base URL
	StopOnFailure bool           `json:"stop_on_failure"`    // Whether to halt if a step fails
}

// WorkflowStep represents a single action in the integration test.
type WorkflowStep struct {
	ID          string                 `json:"id"`
	Description string                 `json:"description,omitempty"`
	Action      string                 `json:"action"`           // http, wait, extract, assert
	Params      map[string]interface{} `json:"params,omitempty"` // Action-specific parameters
}

// OrchestrateResult represents the outcome of the entire workflow.
type OrchestrateResult struct {
	TotalSteps  int          `json:"total_steps"`
	Completed   int          `json:"completed_steps"`
	Failed      int          `json:"failed_steps"`
	StepResults []StepResult `json:"step_results"`
	Summary     string       `json:"summary"`
}

// StepResult details the outcome of an individual step.
type StepResult struct {
	StepID      string `json:"step_id"`
	Description string `json:"description"`
	Status      string `json:"status"` // pass, fail, skipped
	Message     string `json:"message,omitempty"`
}

func (t *IntegrationOrchestratorTool) Name() string {
	return "orchestrate_integration"
}

func (t *IntegrationOrchestratorTool) Description() string {
	return "Execute a multi-step API integration workflow (e.g., Create -> Login -> Order -> Delete) with state sharing between steps"
}

func (t *IntegrationOrchestratorTool) Parameters() string {
	return `{
  "workflow": [
    {"id": "create", "action": "POST /users", "params": {"name": "Test User"}},
    {"id": "verify", "action": "GET /users/{{id}}"}
  ],
  "base_url": "http://localhost:3000"
}`
}

func (t *IntegrationOrchestratorTool) Execute(args string) (string, error) {
	var params OrchestrateParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	if len(params.Workflow) == 0 {
		return "", fmt.Errorf("workflow must contain at least one step")
	}

	orchestrator := &WorkflowManager{
		httpTool: t.httpTool,
		env:      NewEnvironment(params.BaseURL),
	}

	result := orchestrator.Run(params.Workflow, params.StopOnFailure)
	result.Summary = t.formatSummary(result)

	return result.Summary, nil
}

func (t *IntegrationOrchestratorTool) formatSummary(r OrchestrateResult) string {
	summary := "⛓️ Integration Workflow Orchestration\n\n"
	summary += fmt.Sprintf("Total Steps:     %d\n", r.TotalSteps)
	summary += fmt.Sprintf("Completed:       %d\n", r.Completed)
	summary += fmt.Sprintf("Failed:          %d\n\n", r.Failed)

	for _, step := range r.StepResults {
		icon := "✓"
		switch step.Status {
case "fail":
			icon = "❌"
		case "skipped":
			icon = "⏭️"
		}
		summary += fmt.Sprintf("  %s [%s] %s\n", icon, step.StepID, step.Description)
		if step.Message != "" && step.Status == "fail" {
			summary += fmt.Sprintf("    Status: %s\n", step.Message)
		}
	}

	if r.Failed == 0 {
		summary += "\n✨ All integration steps completed successfully."
	}

	return summary
}
