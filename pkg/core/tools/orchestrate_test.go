package tools

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRunTestsTool(t *testing.T) {
	// Setup dependencies
	rm := NewResponseManager()
	vs := NewVariableStore("")
	ht := NewHTTPTool(rm, vs)
	at := NewAssertTool(rm)

	tool := NewRunTestsTool(ht, at, vs)

	scenarios := []TestScenario{
		{
			ID:       "test-001",
			Name:     "Test 1",
			Category: "happy_path",
			Method:   "GET",
			URL:      "/health",
			Expected: TestExpectation{
				StatusCode: 200,
			},
		},
	}

	params := RunTestsParams{
		Scenarios: scenarios,
		BaseURL:   "http://localhost:invalid",
	}

	jsonArgs, _ := json.Marshal(params)
	result, err := tool.Execute(string(jsonArgs))
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !strings.Contains(result, "Test Results") {
		t.Errorf("Expected result to contain 'Test Results', got: %s", result)
	}

	if !strings.Contains(result, "✗") { // Should fail because of invalid URL
		t.Errorf("Expected result to contain failure icon '✗', got: %s", result)
	}
}
