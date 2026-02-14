package persistence

import (
	"strings"

	"github.com/blackcoderx/zap/pkg/storage"
)

// ListRequestsTool lists all saved requests
type ListRequestsTool struct {
	manager *PersistenceManager
}

func NewListRequestsTool(manager *PersistenceManager) *ListRequestsTool {
	return &ListRequestsTool{manager: manager}
}

func (t *ListRequestsTool) Name() string { return "list_requests" }

func (t *ListRequestsTool) Description() string {
	return "List all saved API requests in the .zap/requests directory."
}

func (t *ListRequestsTool) Parameters() string {
	return `{}`
}

func (t *ListRequestsTool) Execute(args string) (string, error) {
	requests, err := storage.ListRequests(t.manager.GetBaseDir())
	if err != nil {
		return "", err
	}

	if len(requests) == 0 {
		return "No saved requests found. Use save_request to save a request.", nil
	}

	var sb strings.Builder
	sb.WriteString("Saved requests:\n")
	for _, req := range requests {
		sb.WriteString("  - " + req + "\n")
	}

	return sb.String(), nil
}
