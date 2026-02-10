package tools

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/blackcoderx/zap/pkg/llm"
)

// CreateTestFileTool generates framework-specific test files
type CreateTestFileTool struct {
	llmClient llm.LLMClient
}

func NewCreateTestFileTool(llmClient llm.LLMClient) *CreateTestFileTool {
	return &CreateTestFileTool{llmClient: llmClient}
}

type CreateTestFileParams struct {
	HandlerFile   string `json:"handler_file"`
	Vulnerability string `json:"vulnerability"`
	FixApplied    bool   `json:"fix_applied"`
	Framework     string `json:"framework"`
}

func (t *CreateTestFileTool) Name() string {
	return "create_test_file"
}

func (t *CreateTestFileTool) Description() string {
	return "Generate framework-appropriate test file with test cases that verify the fix works and prevent future regression of the vulnerability."
}

func (t *CreateTestFileTool) Parameters() string {
	return `{
  "handler_file": "handlers/checkout.go",
  "vulnerability": "SQL injection",
  "fix_applied": true,
  "framework": "gin"
}`
}

func (t *CreateTestFileTool) Execute(args string) (string, error) {
	var params CreateTestFileParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	prompt := fmt.Sprintf(`Generate a unit/integration test file for the following fix using the specified framework.

Handler File: %s
Vulnerability Fixed: %s
Framework: %s

Requirements:
1. Use framework-specific testing conventions.
2. Ensure the test specifically checks for the vulnerability prevention.
3. Include setup and teardown if necessary (e.g., mock DB).

Return ONLY a valid JSON object matching this structure:
{
  "filename": "e.g., checkout_test.go",
  "content": "Full source code of the test file"
}`, params.HandlerFile, params.Vulnerability, params.Framework)

	messages := []llm.Message{
		{Role: "system", Content: "You are an expert QA engineer and developer. Output ONLY valid JSON."},
		{Role: "user", Content: prompt},
	}

	response, err := t.llmClient.Chat(messages)
	if err != nil {
		return "", fmt.Errorf("LLM test generation failed: %w", err)
	}

	// Clean up markdown code blocks
	response = strings.TrimSpace(response)
	if strings.HasPrefix(response, "```") {
		lines := strings.Split(response, "\n")
		if len(lines) >= 2 {
			response = strings.Join(lines[1:len(lines)-1], "\n")
		}
	}
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")

	return strings.TrimSpace(response), nil
}
