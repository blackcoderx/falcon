package tools

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/blackcoderx/zap/pkg/llm"
)

// ProposeFixTool generates and applies code fixes
type ProposeFixTool struct {
	llmClient llm.LLMClient
}

// NewProposeFixTool creates a new propose_fix tool
func NewProposeFixTool(llmClient llm.LLMClient) *ProposeFixTool {
	return &ProposeFixTool{llmClient: llmClient}
}

type ProposeFixParams struct {
	File          string `json:"file"`
	Vulnerability string `json:"vulnerability"`
	CurrentCode   string `json:"current_code"`
	FailedTest    string `json:"failed_test,omitempty"`
}

func (t *ProposeFixTool) Name() string {
	return "propose_fix"
}

func (t *ProposeFixTool) Description() string {
	return "Generate a unified diff showing proposed code changes to fix vulnerability, with explanation of changes and required imports."
}

func (t *ProposeFixTool) Parameters() string {
	return `{
  "file": "handlers/checkout.go",
  "vulnerability": "SQL injection in query parameter",
  "current_code": "...",
  "failed_test": "..."
}`
}

func (t *ProposeFixTool) Execute(args string) (string, error) {
	var params ProposeFixParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	prompt := fmt.Sprintf(`Generate a security fix for the following code vulnerability.

File: %s
Vulnerability: %s
Current Code:
%s

Failed Test Info:
%s

Return ONLY a valid JSON object matching this structure:
{
  "explanation": "Brief explanation of the fix",
  "diff": "Unified diff format of the changes",
  "risk_assessment": "Low|Medium|High risk analysis",
  "required_imports": ["list of new imports if any"]
}`, params.File, params.Vulnerability, params.CurrentCode, params.FailedTest)

	messages := []llm.Message{
		{Role: "system", Content: "You are an expert security engineer and polyglot developer. Output ONLY valid JSON."},
		{Role: "user", Content: prompt},
	}

	response, err := t.llmClient.Chat(messages)
	if err != nil {
		return "", fmt.Errorf("LLM fix proposal failed: %w", err)
	}

	// Clean up markdown code blocks if present
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
