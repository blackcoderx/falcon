package tools

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/blackcoderx/zap/pkg/llm"
)

// AnalyzeEndpointTool uses LLM to analyze API endpoint structure
type AnalyzeEndpointTool struct {
	llmClient llm.LLMClient
}

// NewAnalyzeEndpointTool creates a new analyze_endpoint tool
func NewAnalyzeEndpointTool(llmClient llm.LLMClient) *AnalyzeEndpointTool {
	return &AnalyzeEndpointTool{
		llmClient: llmClient,
	}
}

// AnalyzeEndpointParams defines input for analyze_endpoint
type AnalyzeEndpointParams struct {
	EndpointDescription string `json:"endpoint_description"`
	Method              string `json:"method"`
	URL                 string `json:"url"`
	SampleRequest       string `json:"sample_request,omitempty"`
	Context             string `json:"context,omitempty"`
}

// EndpointAnalysis represents the structured output of analysis
type EndpointAnalysis struct {
	Summary    string          `json:"summary"`
	Parameters []Parameter     `json:"parameters"`
	AuthType   string          `json:"auth_type"`
	Responses  []Response      `json:"responses"`
	Security   []SecurityRisks `json:"security_risks"`
}

type Parameter struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

type Response struct {
	StatusCode  int    `json:"status_code"`
	Description string `json:"description"`
}

type SecurityRisks struct {
	Risk        string `json:"risk"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
}

func (t *AnalyzeEndpointTool) Name() string {
	return "analyze_endpoint"
}

func (t *AnalyzeEndpointTool) Description() string {
	return "Use LLM to analyze API endpoint structure, parameters, auth requirements, expected responses, and security considerations from URL and optional sample request"
}

func (t *AnalyzeEndpointTool) Parameters() string {
	return `{
  "endpoint_description": "POST /api/checkout - creates order",
  "method": "POST",
  "url": "/api/checkout",
  "sample_request": "{\"item_id\": 123, \"quantity\": 2}",
  "context": "optional code snippet or OpenAPI spec"
}`
}

func (t *AnalyzeEndpointTool) Execute(args string) (string, error) {
	var params AnalyzeEndpointParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	prompt := fmt.Sprintf(`Analyze the following API endpoint and provide a structured JSON response.

Endpoint: %s %s
Description: %s
Sample Request: %s
Context: %s

Return ONLY a valid JSON object matching this structure:
{
  "summary": "Brief description of what the endpoint does",
  "parameters": [
    {"name": "param_name", "type": "string|int|bool", "required": true, "description": "what it is"}
  ],
  "auth_type": "Bearer Token|API Key|None",
  "responses": [
    {"status_code": 200, "description": "Success response"}
  ],
  "security_risks": [
    {"risk": "SQL Injection", "severity": "High", "description": "Potential in id param"}
  ]
}`, params.Method, params.URL, params.EndpointDescription, params.SampleRequest, params.Context)

	messages := []llm.Message{
		{Role: "system", Content: "You are an expert API security analyst. Output ONLY valid JSON."},
		{Role: "user", Content: prompt},
	}

	response, err := t.llmClient.Chat(messages)
	if err != nil {
		return "", fmt.Errorf("LLM analysis failed: %w", err)
	}

	// Clean up markdown code blocks if present
	runes := []rune(response)
	if len(runes) > 0 && runes[0] == '`' {
		response = strings.TrimPrefix(response, "```json")
		response = strings.TrimPrefix(response, "```")
		response = strings.TrimSuffix(response, "```")
	}

	return strings.TrimSpace(response), nil
}

// AnalyzeFailureTool uses LLM to explain test failures
type AnalyzeFailureTool struct {
	llmClient llm.LLMClient
}

// NewAnalyzeFailureTool creates a new analyze_failure tool
func NewAnalyzeFailureTool(llmClient llm.LLMClient) *AnalyzeFailureTool {
	return &AnalyzeFailureTool{
		llmClient: llmClient,
	}
}

// AnalyzeFailureParams defines input for analyze_failure
type AnalyzeFailureParams struct {
	TestResult       TestResult `json:"test_result"`
	ResponseBody     string     `json:"response_body"`
	ExpectedBehavior string     `json:"expected_behavior"`
}

func (t *AnalyzeFailureTool) Name() string {
	return "analyze_failure"
}

func (t *AnalyzeFailureTool) Description() string {
	return "Use LLM to explain why a test failed, assess vulnerability severity, identify OWASP/CWE category, estimate impact, and provide detailed fix suggestions with code examples"
}

func (t *AnalyzeFailureTool) Parameters() string {
	return `{
  "test_result": { ... },
  "response_body": "actual response from server",
  "expected_behavior": "expected 400 Bad Request"
}`
}

func (t *AnalyzeFailureTool) Execute(args string) (string, error) {
	var params AnalyzeFailureParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	// Marshal TestResult to pretty JSON for the prompt
	resultJSON, _ := json.MarshalIndent(params.TestResult, "", "  ")

	prompt := fmt.Sprintf(`Analyze this API test failure and provide a structured assessment.

Test Result:
%s

Actual Response Body:
%s

Expected Behavior:
%s

Return ONLY a valid JSON object matching this structure:
{
  "explanation": "Why it failed",
  "severity": "Critical|High|Medium|Low",
  "owasp_category": "A01:2021 Broken Access Control",
  "cwe_id": "CWE-89",
  "impact": "Description of potential impact",
  "remediation": "Step-by-step fix",
  "code_suggestion": "Example code fix"
}`, string(resultJSON), params.ResponseBody, params.ExpectedBehavior)

	messages := []llm.Message{
		{Role: "system", Content: "You are an expert API security auditor. Output ONLY valid JSON."},
		{Role: "user", Content: prompt},
	}

	response, err := t.llmClient.Chat(messages)
	if err != nil {
		return "", fmt.Errorf("LLM failure analysis failed: %w", err)
	}

	// Clean up markdown code blocks if present
	runes := []rune(response)
	if len(runes) > 0 && runes[0] == '`' {
		response = strings.TrimPrefix(response, "```json")
		response = strings.TrimPrefix(response, "```")
		response = strings.TrimSuffix(response, "```")
	}

	return strings.TrimSpace(response), nil
}
