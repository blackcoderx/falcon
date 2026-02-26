package prompt

import (
	"strings"
)

// Tool is a minimal interface for tools needed by the prompt builder.
// This avoids circular imports with pkg/core.
type Tool interface {
	Name() string
	Description() string
	Parameters() string
}

// Builder constructs the complete system prompt from modular components.
type Builder struct {
	falconFolder    string
	framework       string
	manifestSummary string
	memoryPreview   string
	tools           map[string]Tool
	useCompactTools bool // If true, use compact reference instead of full descriptions
}

// NewBuilder creates a new prompt builder with configuration.
func NewBuilder() *Builder {
	return &Builder{
		useCompactTools: true, // Default to compact for context efficiency
	}
}

// WithZapFolder sets the workspace path.
func (b *Builder) WithZapFolder(path string) *Builder {
	b.falconFolder = path
	return b
}

// WithFramework sets the user's API framework.
func (b *Builder) WithFramework(framework string) *Builder {
	b.framework = framework
	return b
}

// WithManifestSummary sets the current .falcon folder state.
func (b *Builder) WithManifestSummary(summary string) *Builder {
	b.manifestSummary = summary
	return b
}

// WithMemoryPreview sets the agent's long-term memory context.
func (b *Builder) WithMemoryPreview(preview string) *Builder {
	b.memoryPreview = preview
	return b
}

// WithTools sets the available tools.
func (b *Builder) WithTools(tools map[string]Tool) *Builder {
	b.tools = tools
	return b
}

// UseFullToolDescriptions switches to verbose tool descriptions (more context usage).
func (b *Builder) UseFullToolDescriptions() *Builder {
	b.useCompactTools = false
	return b
}

// Build constructs the final system prompt.
// The order is critical - most important sections first.
func (b *Builder) Build() string {
	var sb strings.Builder

	// 1. Identity - WHO is the agent
	sb.WriteString(Identity)
	sb.WriteString("\n")

	// 2. Guardrails - HARD BOUNDARIES (impenetrable)
	sb.WriteString(Guardrails)
	sb.WriteString("\n")

	// 3. Workflow - HOW to operate
	sb.WriteString(Workflow)
	sb.WriteString("\n")

	// 4. Context - Current session state
	sb.WriteString(BuildContextSection(b.falconFolder, b.framework, b.manifestSummary, b.memoryPreview))
	sb.WriteString("\n")

	// 5. Tools - WHAT capabilities are available
	if b.useCompactTools {
		// Use ultra-compact reference table
		sb.WriteString(CompactToolReference)
	} else {
		// Use full tool descriptions (verbose)
		if b.tools != nil {
			sb.WriteString(BuildToolsSection(b.tools))
		}
	}
	sb.WriteString("\n")

	// 6. Output Format - HOW to respond (always last)
	sb.WriteString(OutputFormat)

	return sb.String()
}

// GetTokenEstimate provides a rough estimate of token usage.
// Useful for monitoring context window consumption.
func (b *Builder) GetTokenEstimate() int {
	// Rough estimate: 1 token ≈ 4 characters
	prompt := b.Build()
	return len(prompt) / 4
}
