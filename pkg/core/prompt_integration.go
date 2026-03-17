package core

import (
	"github.com/blackcoderx/falcon/pkg/core/prompt"
	"github.com/blackcoderx/falcon/pkg/core/tools/shared"
)

// preparePromptBuilder constructs a fully configured prompt.Builder
// with the current agent state. Used by both buildSystemPrompt and
// GetPromptTokenEstimate to avoid duplicating the setup logic.
func (a *Agent) preparePromptBuilder() *prompt.Builder {
	manifestSummary := shared.GetManifestSummary(FalconFolderName)

	memoryPreview := ""
	if a.memoryStore != nil {
		memoryPreview = a.memoryStore.GetCompactSummary()
	}

	promptTools := make(map[string]prompt.Tool)
	a.toolsMu.RLock()
	for name, tool := range a.tools {
		promptTools[name] = tool
	}
	a.toolsMu.RUnlock()

	return prompt.NewBuilder().
		WithZapFolder(FalconFolderName).
		WithFramework(a.framework).
		WithManifestSummary(manifestSummary).
		WithMemoryPreview(memoryPreview).
		WithTools(promptTools)
}

// buildSystemPrompt constructs the complete system prompt for the LLM.
func (a *Agent) buildSystemPrompt() string {
	return a.preparePromptBuilder().Build()
}

// GetPromptTokenEstimate returns an estimate of how many tokens the system prompt uses.
func (a *Agent) GetPromptTokenEstimate() int {
	return a.preparePromptBuilder().GetTokenEstimate()
}
