// Package llm provides client implementations for Large Language Models.
// It defines a common interface (LLMClient) that all providers must implement,
// enabling easy switching between different LLM backends like Ollama and Gemini.
package llm

// LLMClient defines the interface that all LLM providers must implement.
// This allows the agent to work with any LLM backend without tight coupling.
type LLMClient interface {
	// Chat sends a non-streaming chat request and returns the complete response.
	Chat(messages []Message) (string, error)

	// ChatStream sends a streaming chat request and calls callback for each chunk.
	// Returns the complete response when streaming finishes.
	ChatStream(messages []Message, callback StreamCallback) (string, error)

	// CheckConnection verifies that the LLM service is accessible.
	CheckConnection() error

	// GetModel returns the name of the model being used.
	GetModel() string
}
