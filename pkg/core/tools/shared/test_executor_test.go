package shared

import (
	"fmt"
	"strings"
	"testing"
)

func TestRunScenario_StatusCodeMatch(t *testing.T) {
	executor := &TestExecutor{HTTPTool: nil}
	resp := &HTTPResponse{StatusCode: 200, Body: `{"ok":true}`}

	scenario := TestScenario{
		ID:   "t1",
		Name: "status check",
		Expected: TestExpectation{
			StatusCode: 200,
		},
	}

	result := executor.buildResult(scenario, resp, nil)
	if !result.Passed {
		t.Errorf("expected pass, got fail: %s", result.Error)
	}
}

func TestRunScenario_StatusCodeMismatch(t *testing.T) {
	executor := &TestExecutor{HTTPTool: nil}
	resp := &HTTPResponse{StatusCode: 404, Body: `not found`}

	scenario := TestScenario{
		ID:   "t2",
		Name: "wrong status",
		Expected: TestExpectation{
			StatusCode: 200,
		},
	}

	result := executor.buildResult(scenario, resp, nil)
	if result.Passed {
		t.Error("expected fail for status mismatch")
	}
	if !strings.Contains(result.Error, "expected 200") {
		t.Errorf("error should mention expected status: %s", result.Error)
	}
}

func TestRunScenario_StatusCodeRange(t *testing.T) {
	executor := &TestExecutor{HTTPTool: nil}
	resp := &HTTPResponse{StatusCode: 201, Body: ""}

	scenario := TestScenario{
		ID:   "t3",
		Name: "range check",
		Expected: TestExpectation{
			StatusCodeRange: &StatusCodeRange{Min: 200, Max: 299},
		},
	}

	result := executor.buildResult(scenario, resp, nil)
	if !result.Passed {
		t.Errorf("expected pass for 201 in [200-299], got: %s", result.Error)
	}

	resp.StatusCode = 400
	result = executor.buildResult(scenario, resp, nil)
	if result.Passed {
		t.Error("expected fail for 400 outside [200-299]")
	}
}

func TestRunScenario_BodyContains(t *testing.T) {
	executor := &TestExecutor{HTTPTool: nil}
	resp := &HTTPResponse{StatusCode: 200, Body: `{"name":"alice","role":"admin"}`}

	scenario := TestScenario{
		ID:   "t4",
		Name: "body check",
		Expected: TestExpectation{
			BodyContains:    []string{"alice"},
			BodyNotContains: []string{"bob"},
		},
	}

	result := executor.buildResult(scenario, resp, nil)
	if !result.Passed {
		t.Errorf("expected pass: %s", result.Error)
	}

	scenario.Expected.BodyContains = []string{"bob"}
	result = executor.buildResult(scenario, resp, nil)
	if result.Passed {
		t.Error("expected fail when body missing 'bob'")
	}
}

func TestRunScenario_MaxDuration(t *testing.T) {
	executor := &TestExecutor{HTTPTool: nil}
	resp := &HTTPResponse{StatusCode: 200, Body: "ok"}

	scenario := TestScenario{
		ID:   "t5",
		Name: "duration check",
		Expected: TestExpectation{
			MaxDurationMs: 100,
		},
	}

	// Duration within limit
	result := executor.buildResultWithDuration(scenario, resp, nil, 50)
	if !result.Passed {
		t.Errorf("expected pass for 50ms < 100ms limit: %s", result.Error)
	}

	// Duration exceeds limit
	result = executor.buildResultWithDuration(scenario, resp, nil, 200)
	if result.Passed {
		t.Error("expected fail for 200ms > 100ms limit")
	}
}

func TestRunScenario_HeaderContains(t *testing.T) {
	executor := &TestExecutor{HTTPTool: nil}
	resp := &HTTPResponse{
		StatusCode: 200,
		Body:       "ok",
		Headers:    map[string]string{"Content-Type": "application/json"},
	}

	scenario := TestScenario{
		ID:   "t6",
		Name: "header check",
		Expected: TestExpectation{
			HeaderContains: map[string]string{"Content-Type": "json"},
		},
	}

	result := executor.buildResult(scenario, resp, nil)
	if !result.Passed {
		t.Errorf("expected pass: %s", result.Error)
	}
}

func TestRunScenario_HTTPError(t *testing.T) {
	executor := &TestExecutor{HTTPTool: nil}

	scenario := TestScenario{
		ID:   "t7",
		Name: "http error",
	}

	result := executor.buildResult(scenario, nil, fmt.Errorf("connection refused"))
	if result.Passed {
		t.Error("expected fail on HTTP error")
	}
	if !strings.Contains(result.Error, "connection refused") {
		t.Errorf("error should contain cause: %s", result.Error)
	}
}
