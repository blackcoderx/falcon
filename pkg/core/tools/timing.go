package tools

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// WaitTool provides delay functionality
type WaitTool struct{}

// NewWaitTool creates a new wait tool
func NewWaitTool() *WaitTool {
	return &WaitTool{}
}

// WaitParams defines wait parameters
type WaitParams struct {
	DurationMs int    `json:"duration_ms"`
	Reason     string `json:"reason,omitempty"`
}

// Name returns the tool name
func (t *WaitTool) Name() string {
	return "wait"
}

// Description returns the tool description
func (t *WaitTool) Description() string {
	return "Wait for a specified duration (useful for async operations, webhooks, rate limiting)"
}

// Parameters returns the tool parameter description
func (t *WaitTool) Parameters() string {
	return `{"duration_ms": 1000, "reason": "waiting for webhook processing"}`
}

// Execute waits for the specified duration
func (t *WaitTool) Execute(args string) (string, error) {
	var params WaitParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	if params.DurationMs <= 0 {
		return "", fmt.Errorf("duration_ms must be positive")
	}

	if params.DurationMs > 60000 {
		return "", fmt.Errorf("duration_ms cannot exceed 60000 (60 seconds)")
	}

	duration := time.Duration(params.DurationMs) * time.Millisecond
	time.Sleep(duration)

	message := fmt.Sprintf("Waited %dms", params.DurationMs)
	if params.Reason != "" {
		message += fmt.Sprintf(" (%s)", params.Reason)
	}

	return message, nil
}

// RetryTool provides retry logic with exponential backoff
type RetryTool struct {
	agent ToolExecutor // Interface to execute other tools
}

// ToolExecutor interface for executing tools (to avoid circular dependency)
type ToolExecutor interface {
	ExecuteTool(toolName string, args string) (string, error)
}

// NewRetryTool creates a new retry tool
func NewRetryTool(executor ToolExecutor) *RetryTool {
	return &RetryTool{agent: executor}
}

// RetryParams defines retry parameters
type RetryParams struct {
	Tool          string `json:"tool"`
	Args          string `json:"args"`
	MaxAttempts   int    `json:"max_attempts"`
	RetryDelayMs  int    `json:"retry_delay_ms"`
	Backoff       string `json:"backoff,omitempty"` // "linear" or "exponential"
	RetryOnStatus []int  `json:"retry_on_status,omitempty"`
}

// Name returns the tool name
func (t *RetryTool) Name() string {
	return "retry"
}

// Description returns the tool description
func (t *RetryTool) Description() string {
	return "Retry a tool execution with configurable attempts, delay, and exponential backoff"
}

// Parameters returns the tool parameter description
func (t *RetryTool) Parameters() string {
	return `{
  "tool": "http_request",
  "args": "{...}",
  "max_attempts": 3,
  "retry_delay_ms": 500,
  "backoff": "exponential",
  "retry_on_status": [500, 502, 503]
}`
}

// Execute retries a tool execution
func (t *RetryTool) Execute(args string) (string, error) {
	var params RetryParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	if params.Tool == "" {
		return "", fmt.Errorf("'tool' parameter is required")
	}

	if params.MaxAttempts <= 0 {
		params.MaxAttempts = 3
	}

	if params.MaxAttempts > 10 {
		return "", fmt.Errorf("max_attempts cannot exceed 10")
	}

	if params.RetryDelayMs <= 0 {
		params.RetryDelayMs = 500
	}

	if params.Backoff == "" {
		params.Backoff = "linear"
	}

	var lastError error
	var result string
	var attemptLogs []string

	for attempt := 1; attempt <= params.MaxAttempts; attempt++ {
		// Execute the tool
		if t.agent == nil {
			return "", fmt.Errorf("retry tool not properly initialized (no executor)")
		}

		result, lastError = t.agent.ExecuteTool(params.Tool, params.Args)

		if lastError == nil {
			// Success!
			attemptLogs = append(attemptLogs, fmt.Sprintf("Attempt %d: Success", attempt))
			break
		}

		// Check if we should retry based on HTTP status codes
		if len(params.RetryOnStatus) > 0 {
			shouldRetry := false
			errorMsg := lastError.Error()
			for _, statusCode := range params.RetryOnStatus {
				if strings.Contains(errorMsg, fmt.Sprintf("%d", statusCode)) {
					shouldRetry = true
					break
				}
			}
			if !shouldRetry {
				attemptLogs = append(attemptLogs, fmt.Sprintf("Attempt %d: Failed (non-retryable error)", attempt))
				break
			}
		}

		attemptLogs = append(attemptLogs, fmt.Sprintf("Attempt %d: Failed - %v", attempt, lastError))

		// Don't sleep after the last attempt
		if attempt < params.MaxAttempts {
			delay := t.calculateDelay(params.RetryDelayMs, attempt, params.Backoff)
			attemptLogs = append(attemptLogs, fmt.Sprintf("  Waiting %dms before retry...", delay))
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
	}

	// Build result message
	var sb strings.Builder
	sb.WriteString("Retry execution log:\n\n")
	for _, log := range attemptLogs {
		sb.WriteString(log + "\n")
	}

	if lastError != nil {
		sb.WriteString(fmt.Sprintf("\nFinal result: FAILED after %d attempts\nLast error: %v", params.MaxAttempts, lastError))
		return sb.String(), lastError
	}

	sb.WriteString(fmt.Sprintf("\nFinal result: SUCCESS\n\n%s", result))
	return sb.String(), nil
}

// calculateDelay computes retry delay based on backoff strategy
func (t *RetryTool) calculateDelay(baseDelay, attempt int, backoff string) int {
	switch backoff {
	case "exponential":
		// Exponential: delay * 2^(attempt-1)
		multiplier := 1 << (attempt - 1) // 2^(attempt-1)
		return baseDelay * multiplier
	case "linear":
		fallthrough
	default:
		return baseDelay
	}
}
