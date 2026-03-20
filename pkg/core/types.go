// Package core provides the central agent logic, tool management, and ReAct loop
// implementation for the Falcon API debugging assistant.
package core

// Tool represents an agent capability that can be executed.
// Each tool has a name, description, parameters schema, and execution logic.
// Tools are registered with the Agent and can be invoked during the ReAct loop.
type Tool interface {
	// Name returns the unique identifier for this tool.
	Name() string
	// Description returns a human-readable description of what this tool does.
	Description() string
	// Parameters returns a description of the JSON parameters this tool accepts.
	Parameters() string
	// Execute runs the tool with the given JSON arguments and returns the result.
	Execute(args string) (string, error)
}

// AgentEvent represents a state change during agent processing.
// Events are emitted via callbacks to enable real-time UI updates.
type AgentEvent struct {
	// Type indicates the event type: "thinking", "tool_call", "observation",
	// "answer", "error", "streaming", "confirmation_required"
	Type string
	// Content holds the main event payload (varies by type)
	Content string
	// ToolArgs contains tool arguments (present only for "tool_call" events)
	ToolArgs string
	// FileConfirmation contains file write info (present only for "confirmation_required" events)
	FileConfirmation *FileConfirmation
}

// FileConfirmation contains information for file write confirmation prompts.
// This enables human-in-the-loop approval before any file modifications.
type FileConfirmation struct {
	// FilePath is the path to the file being modified
	FilePath string
	// IsNewFile is true if creating a new file, false if modifying existing
	IsNewFile bool
	// Diff is the unified diff showing the proposed changes
	Diff string
}

// EventCallback is the function signature for agent event handlers.
// Callbacks receive events as the agent progresses through the ReAct loop.
type EventCallback func(AgentEvent)

// ConfirmableTool is a tool that requires user confirmation before executing.
// Tools implementing this interface can emit confirmation requests back to the TUI,
// enabling human-in-the-loop approval for potentially destructive operations.
type ConfirmableTool interface {
	Tool
	// SetEventCallback sets the callback function for emitting events
	SetEventCallback(callback EventCallback)
}

