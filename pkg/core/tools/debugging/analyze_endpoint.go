package debugging

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/blackcoderx/falcon/pkg/llm"
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
