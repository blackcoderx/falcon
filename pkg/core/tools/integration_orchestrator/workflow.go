package integration_orchestrator

import (
	"fmt"
	"strings"

	"github.com/blackcoderx/zap/pkg/core/tools/shared"
)

// WorkflowManager orchestrates the execution of individual steps.
type WorkflowManager struct {
	httpTool *shared.HTTPTool
	env      *Environment
}

// Run executes the workflow steps in sequence.
func (m *WorkflowManager) Run(steps []WorkflowStep, stopOnFailure bool) OrchestrateResult {
	var results OrchestrateResult
	results.TotalSteps = len(steps)

	halted := false
	for _, step := range steps {
		if halted {
			results.StepResults = append(results.StepResults, StepResult{
				StepID:      step.ID,
				Description: step.Description,
				Status:      "skipped",
			})
			continue
		}

		res := m.executeStep(step)
		results.StepResults = append(results.StepResults, res)

		if res.Status == "pass" {
			results.Completed++
		} else {
			results.Failed++
			if stopOnFailure {
				halted = true
			}
		}
	}

	return results
}

func (m *WorkflowManager) executeStep(step WorkflowStep) StepResult {
	// 1. Resolve variables in step parameters
	// (Variable interpolation logic would go here)

	description := step.Description
	if description == "" {
		description = fmt.Sprintf("Execute %s", step.Action)
	}

	// 2. Dispatch action
	// Simplified: mainly focusing on HTTP for now as it's the core of integration
	if strings.Contains(step.Action, "/") {
		return m.executeHTTPRequest(step)
	}

	return StepResult{
		StepID:      step.ID,
		Description: description,
		Status:      "fail",
		Message:     "Unsupported action type",
	}
}

func (m *WorkflowManager) executeHTTPRequest(step WorkflowStep) StepResult {
	parts := strings.SplitN(step.Action, " ", 2)
	method := "GET"
	path := step.Action
	if len(parts) == 2 {
		method = parts[0]
		path = parts[1]
	}

	url := m.env.ResolveURL(path)

	req := shared.HTTPRequest{
		Method: method,
		URL:    url,
		// In a real implementation, we'd pull from step.Params
	}

	resp, err := m.httpTool.Run(req)
	if err != nil {
		return StepResult{
			StepID:      step.ID,
			Description: step.Description,
			Status:      "fail",
			Message:     err.Error(),
		}
	}

	if resp.StatusCode >= 400 {
		return StepResult{
			StepID:      step.ID,
			Description: step.Description,
			Status:      "fail",
			Message:     fmt.Sprintf("HTTP %d", resp.StatusCode),
		}
	}

	// 3. Post-execution: extract variables if defined in Params
	// (Logic for shared state management)

	return StepResult{
		StepID:      step.ID,
		Description: step.Description,
		Status:      "pass",
	}
}
