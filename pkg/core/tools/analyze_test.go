package tools

import (
	"strings"
	"testing"
)

func TestAnalyzeEndpointTool(t *testing.T) {
	mockResponse := `{
        "summary": "Test Endpoint",
        "parameters": [{"name": "id", "type": "int", "required": true, "description": "ID"}],
        "auth_type": "None",
        "responses": [{"status_code": 200, "description": "OK"}],
        "security_risks": []
    }`

	client := &MockLLMClient{Response: mockResponse}
	tool := NewAnalyzeEndpointTool(client)

	params := `{
        "endpoint_description": "Test",
        "method": "GET",
        "url": "/test"
    }`

	result, err := tool.Execute(params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !strings.Contains(result, "Test Endpoint") {
		t.Errorf("Expected result to contain 'Test Endpoint', got: %s", result)
	}
}

func TestAnalyzeFailureTool(t *testing.T) {
	mockResponse := `{
        "explanation": "Test Failure",
        "severity": "High",
        "owasp_category": "A01",
        "cwe_id": "CWE-123",
        "impact": "Bad",
        "remediation": "Fix it",
        "code_suggestion": "code"
    }`

	client := &MockLLMClient{Response: mockResponse}
	tool := NewAnalyzeFailureTool(client)

	params := `{
        "test_result": {
            "scenario_id": "test-001",
            "scenario_name": "Test",
            "category": "security",
            "passed": false,
            "actual_status": 500
        },
        "response_body": "error",
        "expected_behavior": "success"
    }`

	result, err := tool.Execute(params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !strings.Contains(result, "Test Failure") {
		t.Errorf("Expected result to contain 'Test Failure', got: %s", result)
	}
}
