package debugging

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/blackcoderx/falcon/pkg/core/tools/shared"
	"github.com/blackcoderx/falcon/pkg/llm"
)

// GenerateTestsTool uses LLM to generate test scenarios
type GenerateTestsTool struct {
	llmClient llm.LLMClient
}

// NewGenerateTestsTool creates a new generate_tests tool
func NewGenerateTestsTool(llmClient llm.LLMClient) *GenerateTestsTool {
	return &GenerateTestsTool{
		llmClient: llmClient,
	}
}

// GenerateTestsParams defines input for generate_tests
type GenerateTestsParams struct {
	Analysis   shared.EndpointAnalysis `json:"analysis"`
	Categories []string                `json:"categories,omitempty"`
	Count      int                     `json:"count"`
}

func (t *GenerateTestsTool) Name() string {
	return "generate_tests"
}

func (t *GenerateTestsTool) Description() string {
	return "Use LLM and templates to generate 20-30 comprehensive test scenarios covering security (SQL injection, XSS), validation (missing fields, wrong types), edge cases (unicode, boundaries), and performance (rate limiting)"
}

func (t *GenerateTestsTool) Parameters() string {
	return `{
  "analysis": { /* output from analyze_endpoint */ },
  "categories": ["security", "validation", "edge_case"],
  "count": 25
}`
}

func (t *GenerateTestsTool) Execute(args string) (string, error) {
	var params GenerateTestsParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	analysisJSON, _ := json.MarshalIndent(params.Analysis, "", "  ")
	categories := "all (security, validation, happy_path, edge_case, performance)"
	if len(params.Categories) > 0 {
		categories = strings.Join(params.Categories, ", ")
	}

	count := params.Count
	if count == 0 {
		count = 20
	}

	prompt := fmt.Sprintf(`Generate %d API test scenarios for the following endpoint based on the analysis.

Endpoint Analysis:
%s

Categories to cover: %s

Requirements:
1. Generate DIVERSE scenarios:
   - Security: SQLi, XSS, Auth bypass, IDOR, etc. (High priority)
   - Validation: Missing fields, wrong types, empty values, nulls
   - Happy Path: Valid requests with different permutations
   - Edge Cases: Unicode, very long strings, boundary values
   - Performance: Rate limiting checks (e.g. rapid requests)
2. Use "id" format: "sec-001", "val-001", "hap-001", "edg-001", "perf-001"
3. Include specific "expected" assertions (status codes, body content constraints)
4. For security tests, include OWASP references (e.g. "A03:2021")

Return ONLY a valid JSON array of TestScenario objects matching this schema:
[
  {
    "id": "sec-001",
    "name": "SQL Injection in id",
    "category": "security",
    "severity": "critical",
    "description": "Attempt SQL injection payload in id parameter",
    "method": "POST",
    "url": "/api/checkout",
    "body": {"id": "1' OR '1'='1"},
    "expected": {
      "status_code": 400,
      "body_not_contains": ["SQL syntax", "error"]
    },
    "owasp_ref": "A03:2021"
  }
]`, count, string(analysisJSON), categories)

	messages := []llm.Message{
		{Role: "system", Content: "You are an expert software QA engineer and security researcher. Output ONLY valid JSON array."},
		{Role: "user", Content: prompt},
	}

	response, err := t.llmClient.Chat(messages)
	if err != nil {
		return "", fmt.Errorf("LLM test generation failed: %w", err)
	}

	// Clean up markdown code blocks if present
	response = strings.TrimSpace(response)
	if strings.HasPrefix(response, "```") {
		lines := strings.Split(response, "\n")
		if len(lines) >= 2 {
			// Remove first and last line (```json and ```)
			response = strings.Join(lines[1:len(lines)-1], "\n")
		}
	}
	// Another check for just leading ```json
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")

	// Validate that it parses as []TestScenario
	var scenarios []shared.TestScenario
	if err := json.Unmarshal([]byte(response), &scenarios); err != nil {
		return strings.TrimSpace(response), nil
	}

	return strings.TrimSpace(response), nil
}
