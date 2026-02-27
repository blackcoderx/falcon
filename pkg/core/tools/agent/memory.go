package agent

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/blackcoderx/falcon/pkg/core"
	"github.com/blackcoderx/falcon/pkg/core/tools/shared"
)

// MemoryTool provides persistent memory operations for the agent.
type MemoryTool struct {
	store *core.MemoryStore
}

// NewMemoryTool creates a new memory tool.
func NewMemoryTool(store *core.MemoryStore) *MemoryTool {
	return &MemoryTool{store: store}
}

// MemoryParams defines memory tool operations.
type MemoryParams struct {
	Action   string `json:"action"`             // "save", "recall", "forget", "list", "update_knowledge"
	Key      string `json:"key,omitempty"`      // Key for save/forget
	Value    string `json:"value,omitempty"`    // Value for save
	Category string `json:"category,omitempty"` // Category for save/list: "preference", "endpoint", "error", "project", "general"
	Query    string `json:"query,omitempty"`    // Search query for recall
	Section  string `json:"section,omitempty"`  // Section name for update_knowledge
	Content  string `json:"content,omitempty"`  // Markdown content for update_knowledge
}

// Name returns the tool name.
func (t *MemoryTool) Name() string {
	return "memory"
}

// Description returns the tool description.
func (t *MemoryTool) Description() string {
	return "Manage persistent agent memory and API knowledge base. Actions: save (key/value facts), recall (search facts), forget (remove facts), list (all facts), update_knowledge (write API facts into a named section of falcon.md — use this whenever you discover an endpoint, auth method, data model, or error pattern)"
}

// Parameters returns the tool parameter description.
func (t *MemoryTool) Parameters() string {
	return `{
  "action": "save|recall|forget|list|update_knowledge",
  "key": "memory_key",
  "value": "memory_value",
  "category": "preference|endpoint|error|project|general",
  "query": "search_query",
  "section": "Base URLs|Authentication|Known Endpoints|Data Models|Known Errors|Project Notes",
  "content": "markdown content to write into the section"
}`
}

// Execute performs memory operations.
func (t *MemoryTool) Execute(args string) (string, error) {
	var params MemoryParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	switch params.Action {
	case "save":
		if params.Key == "" {
			return "", fmt.Errorf("'key' is required for save action")
		}
		if params.Value == "" {
			return "", fmt.Errorf("'value' is required for save action")
		}

		if err := t.store.Save(params.Key, params.Value, params.Category); err != nil {
			return "", fmt.Errorf("failed to save memory: %w", err)
		}

		category := params.Category
		if category == "" {
			category = "general"
		}
		return fmt.Sprintf("Saved to memory: [%s] %s = %s\n(Persisted across sessions)", category, params.Key, params.Value), nil

	case "recall":
		if params.Query == "" {
			return "", fmt.Errorf("'query' is required for recall action")
		}

		results := t.store.Recall(params.Query)
		if len(results) == 0 {
			return fmt.Sprintf("No memories found matching '%s'.", params.Query), nil
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Found %d memories matching '%s':\n\n", len(results), params.Query))
		for _, e := range results {
			sb.WriteString(fmt.Sprintf("  [%s] %s: %s\n", e.Category, e.Key, e.Value))
		}
		return sb.String(), nil

	case "forget":
		if params.Key == "" {
			return "", fmt.Errorf("'key' is required for forget action")
		}

		if err := t.store.Forget(params.Key); err != nil {
			return "", err
		}
		return fmt.Sprintf("Forgotten: %s\n(Removed from persistent memory)", params.Key), nil

	case "list":
		var entries []core.MemoryEntry
		if params.Category != "" {
			entries = t.store.ListByCategory(params.Category)
		} else {
			entries = t.store.List()
		}

		if len(entries) == 0 {
			if params.Category != "" {
				return fmt.Sprintf("No memories in category '%s'.", params.Category), nil
			}
			return "No memories stored yet.", nil
		}

		var sb strings.Builder
		if params.Category != "" {
			sb.WriteString(fmt.Sprintf("Memories in '%s' (%d):\n\n", params.Category, len(entries)))
		} else {
			sb.WriteString(fmt.Sprintf("All memories (%d):\n\n", len(entries)))
		}
		for _, e := range entries {
			sb.WriteString(fmt.Sprintf("  [%s] %s: %s\n", e.Category, e.Key, e.Value))
		}
		return sb.String(), nil

	case "update_knowledge":
		if params.Section == "" || params.Content == "" {
			return "", fmt.Errorf("'section' and 'content' are required for update_knowledge")
		}
		if err := t.store.UpdateKnowledge(params.Section, params.Content); err != nil {
			return "", err
		}
		// Validate that the write produced meaningful content
		falconMDPath := filepath.Join(t.store.FalconDir(), "falcon.md")
		if err := shared.ValidateFalconMD(falconMDPath); err != nil {
			return fmt.Sprintf("Updated '%s' in falcon.md but validation warning: %v", params.Section, err), nil
		}
		return fmt.Sprintf("Updated '%s' section in falcon.md (validated)", params.Section), nil

	default:
		return "", fmt.Errorf("unknown action '%s' (use: save, recall, forget, list, update_knowledge)", params.Action)
	}
}
