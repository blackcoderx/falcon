package agent

import (
	"encoding/json"
	"fmt"

	"github.com/blackcoderx/zap/pkg/core/tools/shared"
)

// RunSingleTestTool executes a single test scenario
type RunSingleTestTool struct {
	httpTool   *shared.HTTPTool
	assertTool *shared.AssertTool
	varStore   *shared.VariableStore
}

func NewRunSingleTestTool(httpTool *shared.HTTPTool, assertTool *shared.AssertTool, varStore *shared.VariableStore) *RunSingleTestTool {
	return &RunSingleTestTool{
		httpTool:   httpTool,
		assertTool: assertTool,
		varStore:   varStore,
	}
}

type RunSingleTestParams struct {
	TestID   string               `json:"test_id"`
	Scenario *shared.TestScenario `json:"scenario,omitempty"` // Option to pass full scenario
	BaseURL  string               `json:"base_url"`
}

func (t *RunSingleTestTool) Name() string {
	return "run_single_test"
}

func (t *RunSingleTestTool) Description() string {
	return "Execute one specific test scenario by name or ID, useful for re-running a particular test after applying a fix to verify it works"
}

func (t *RunSingleTestTool) Parameters() string {
	return `{
  "test_id": "sec-001",
  "base_url": "http://localhost:8080",
  "scenario": { ... } // Optional check to provide full scenario definition if not found in memory
}`
}

func (t *RunSingleTestTool) Execute(args string) (string, error) {
	var params RunSingleTestParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("parse error: %w", err)
	}

	if params.Scenario == nil {
		return "", fmt.Errorf("scenario definition is required for run_single_test currently (persistence not implemented)")
	}

	runner := NewRunTestsTool(t.httpTool, t.assertTool, t.varStore)
	result := runner.runSingleScenario(*params.Scenario, params.BaseURL)

	resJSON, _ := json.MarshalIndent(result, "", "  ")
	return string(resJSON), nil
}
