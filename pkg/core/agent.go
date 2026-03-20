// Package core provides the central agent logic, tool management, and ReAct loop
// implementation for the Falcon API debugging assistant.
package core

import (
	"fmt"
	"sync"

	"github.com/blackcoderx/falcon/pkg/llm"
)

// Agent represents the Falcon AI agent that processes user messages,
// executes tools, and provides API debugging assistance.
type Agent struct {
	llmClient    llm.LLMClient
	clientMu     sync.RWMutex // Protects access to llmClient
	tools        map[string]Tool
	toolsMu      sync.RWMutex // Protects access to tools map
	history      []llm.Message
	historyMu    sync.RWMutex // Protects access to history slice
	lastResponse interface{}  // Store last tool response for chaining

	// History management
	maxHistory int // maximum number of messages to keep in history (0 = unlimited)

	// User's API framework (gin, fastapi, express, etc.)
	framework string

	// Persistent memory across sessions
	memoryStore *MemoryStore
}

// Default limits for history management.
const (
	DefaultMaxHistory = 100 // Default max messages to keep in history
)

// NewAgent creates a new Falcon agent with the given LLM client.
func NewAgent(llmClient llm.LLMClient) *Agent {
	return &Agent{
		llmClient:    llmClient,
		tools:        make(map[string]Tool),
		history:      []llm.Message{},
		lastResponse: nil,
		maxHistory:   DefaultMaxHistory,
	}
}

// RegisterTool adds a tool to the agent's arsenal.
// This method is thread-safe.
func (a *Agent) RegisterTool(tool Tool) {
	a.toolsMu.Lock()
	defer a.toolsMu.Unlock()
	a.tools[tool.Name()] = tool
}

// ExecuteTool executes a tool by name (used by retry tool).
// This method is thread-safe for looking up the tool.
func (a *Agent) ExecuteTool(toolName string, args string) (string, error) {
	a.toolsMu.RLock()
	tool, ok := a.tools[toolName]
	a.toolsMu.RUnlock()
	if !ok {
		return "", fmt.Errorf("tool '%s' not found", toolName)
	}
	return tool.Execute(args)
}

// SetLastResponse stores the last response from a tool for chaining.
func (a *Agent) SetLastResponse(response interface{}) {
	a.lastResponse = response
}

// LLMClient returns the agent's LLM client.
func (a *Agent) LLMClient() llm.LLMClient {
	a.clientMu.RLock()
	defer a.clientMu.RUnlock()
	return a.llmClient
}

// SwapLLMClient replaces the agent's LLM client at runtime.
func (a *Agent) SwapLLMClient(client llm.LLMClient) {
	a.clientMu.Lock()
	defer a.clientMu.Unlock()
	a.llmClient = client
}

// SetFramework sets the user's API framework for context-aware assistance.
// Supported frameworks include: gin, echo, chi, fiber, fastapi, flask, django,
// express, nestjs, hono, spring, laravel, rails, actix, axum, other.
func (a *Agent) SetFramework(framework string) {
	a.framework = framework
}

// GetFramework returns the configured API framework.
func (a *Agent) GetFramework() string {
	return a.framework
}

// SetMemoryStore sets the persistent memory store for the agent.
func (a *Agent) SetMemoryStore(store *MemoryStore) {
	a.memoryStore = store
}

// GetHistory returns a copy of the agent's conversation history.
// This method is thread-safe.
func (a *Agent) GetHistory() []llm.Message {
	a.historyMu.RLock()
	defer a.historyMu.RUnlock()
	cp := make([]llm.Message, len(a.history))
	copy(cp, a.history)
	return cp
}

// SetMaxHistory sets the maximum number of messages to keep in history.
// Set to 0 for unlimited history (not recommended for long sessions).
func (a *Agent) SetMaxHistory(max int) {
	a.maxHistory = max
}

// AppendHistory adds a message to the history and truncates if necessary.
// When maxHistory is reached, older messages are removed to make room.
// The truncation keeps the most recent messages while preserving context.
// This method is thread-safe.
func (a *Agent) AppendHistory(msg llm.Message) {
	a.historyMu.Lock()
	defer a.historyMu.Unlock()
	a.history = append(a.history, msg)
	a.truncateHistory()
}

// AppendHistoryPair adds an assistant message and observation to history atomically.
// This ensures tool call and observation stay together during truncation.
// This method is thread-safe.
func (a *Agent) AppendHistoryPair(assistantMsg, observationMsg llm.Message) {
	a.historyMu.Lock()
	defer a.historyMu.Unlock()
	a.history = append(a.history, assistantMsg, observationMsg)
	a.truncateHistory()
}

// truncateHistory removes old messages if history exceeds maxHistory.
// Keeps the most recent messages. If maxHistory is 0, no truncation occurs.
// Caller must hold historyMu lock.
func (a *Agent) truncateHistory() {
	if a.maxHistory <= 0 {
		return // Unlimited history
	}

	if len(a.history) > a.maxHistory {
		// Calculate how many messages to remove
		// Keep at least 2 messages for context (a user message and a response)
		excess := len(a.history) - a.maxHistory
		if excess > 0 {
			// Remove from the beginning (oldest messages)
			a.history = a.history[excess:]
		}
	}
}
