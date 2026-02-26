package core

import (
	"github.com/blackcoderx/falcon/pkg/core/prompt"
	"github.com/blackcoderx/falcon/pkg/core/tools/shared"
)

// buildSystemPrompt constructs the complete system prompt for the LLM using the modular prompt system.
// This replaces the old monolithic prompt.go approach with a clean, context-efficient builder pattern.
func (a *Agent) buildSystemPrompt() string {
	// Get current .falcon folder state
	manifestSummary := shared.GetManifestSummary(FalconFolderName)

	// Get memory preview if available
	memoryPreview := ""
	if a.memoryStore != nil {
		memoryPreview = a.memoryStore.GetCompactSummary()
	}

	// Convert tools to prompt.Tool interface (Go doesn't allow direct map type conversion)
	promptTools := make(map[string]prompt.Tool)
	a.toolsMu.RLock()
	for name, tool := range a.tools {
		promptTools[name] = tool
	}
	a.toolsMu.RUnlock()

	// Build prompt using modular system
	builder := prompt.NewBuilder().
		WithZapFolder(FalconFolderName).
		WithFramework(a.framework).
		WithManifestSummary(manifestSummary).
		WithMemoryPreview(memoryPreview).
		WithTools(promptTools)

	return builder.Build()
}

// GetPromptTokenEstimate returns an estimate of how many tokens the system prompt uses.
// Useful for monitoring context window consumption.
func (a *Agent) GetPromptTokenEstimate() int {
	manifestSummary := shared.GetManifestSummary(FalconFolderName)
	memoryPreview := ""
	if a.memoryStore != nil {
		memoryPreview = a.memoryStore.GetCompactSummary()
	}

	// Convert tools to prompt.Tool interface
	promptTools := make(map[string]prompt.Tool)
	a.toolsMu.RLock()
	for name, tool := range a.tools {
		promptTools[name] = tool
	}
	a.toolsMu.RUnlock()

	builder := prompt.NewBuilder().
		WithZapFolder(FalconFolderName).
		WithFramework(a.framework).
		WithManifestSummary(manifestSummary).
		WithMemoryPreview(memoryPreview).
		WithTools(promptTools)

	return builder.GetTokenEstimate()
}
