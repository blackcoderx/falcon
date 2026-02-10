package tools

import (
	"strings"
	"testing"
)

func TestGenerateTestsTool(t *testing.T) {
	mockResponse := `[
        {
            "id": "sec-001",
            "name": "SQL Injection",
            "category": "security",
            "severity": "critical",
            "description": "SQLi test",
            "method": "POST",
            "url": "/api/test",
            "expected": {
                "status_code": 400
            }
        }
    ]`

	client := &MockLLMClient{Response: mockResponse}
	tool := NewGenerateTestsTool(client)

	params := `{
        "analysis": {
            "summary": "Test Endpoint",
            "parameters": [],
            "auth_type": "None",
            "responses": [],
            "security_risks": []
        },
        "count": 5
    }`

	result, err := tool.Execute(params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !strings.Contains(result, "sec-001") {
		t.Errorf("Expected result to contain 'sec-001', got: %s", result)
	}
}
