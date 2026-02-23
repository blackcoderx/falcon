package debugging

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/blackcoderx/falcon/pkg/core/tools/shared"
	"github.com/blackcoderx/falcon/pkg/llm"
)

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

// TestResult is now imported from shared.TestResult

// AnalyzeFailureParams defines input for analyze_failure
type AnalyzeFailureParams struct {
	TestResult       shared.TestResult `json:"test_result"`
	ResponseBody     string            `json:"response_body"`
	ExpectedBehavior string            `json:"expected_behavior"`
}

func (t *AnalyzeFailureTool) Name() string {
	return "analyze_failure"
}

func (t *AnalyzeFailureTool) Description() string {
	return "Use LLM to explain why a test failed, assess vulnerability severity, identify OWASP/CWE category, estimate impact, and provide detailed fix suggestions with code examples"
}

func (t *AnalyzeFailureTool) Parameters() string {
	return `{
  "test_result": { "passed": false, "failures": ["Expected 200, got 500"] },
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
