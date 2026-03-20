package debugging

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/blackcoderx/falcon/pkg/llm"
)

// ProposeFixTool generates and applies code fixes
type ProposeFixTool struct {
	llmClient llm.LLMClient
	workDir   string
}

// NewProposeFixTool creates a new propose_fix tool
func NewProposeFixTool(llmClient llm.LLMClient, workDir string) *ProposeFixTool {
	return &ProposeFixTool{llmClient: llmClient, workDir: workDir}
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
	return "Generates a code fix for a vulnerability or failure. Reads the full file content automatically — no need to call read_file first. Returns a unified diff, a complete patched_content field ready for write_file, and an explanation of the changes."
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

	// Always try to read the full file content for accurate diff generation
	if params.File != "" && t.workDir != "" {
		if fullContent, err := os.ReadFile(filepath.Join(t.workDir, params.File)); err == nil {
			params.CurrentCode = string(fullContent)
		}
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
  "patched_content": "The complete fixed file content after applying the changes",
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
