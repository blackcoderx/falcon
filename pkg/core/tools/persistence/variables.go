package persistence

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/blackcoderx/falcon/pkg/core/tools/shared"
)

// VariableTool provides variable get/set/list operations
type VariableTool struct {
	store *shared.VariableStore
}

// NewVariableTool creates a new variable tool
func NewVariableTool(store *shared.VariableStore) *VariableTool {
	return &VariableTool{store: store}
}

// VariableParams defines variable operations
type VariableParams struct {
	Action string `json:"action"` // "set", "get", "delete", "list"
	Name   string `json:"name,omitempty"`
	Value  string `json:"value,omitempty"`
	Scope  string `json:"scope,omitempty"` // "session" (default) or "global"
}

// Name returns the tool name
func (t *VariableTool) Name() string {
	return "variable"
}

// Description returns the tool description
func (t *VariableTool) Description() string {
	return "Manage session and global variables for storing values across requests. Actions: set, get, delete, list"
}

// Parameters returns the tool parameter description
func (t *VariableTool) Parameters() string {
	return `{
  "action": "set|get|delete|list",
  "name": "variable_name",
  "value": "variable_value",
  "scope": "session|global"
}`
}

// Execute performs variable operations
func (t *VariableTool) Execute(args string) (string, error) {
	var params VariableParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	switch params.Action {
	case "set":
		if params.Name == "" {
			return "", fmt.Errorf("'name' is required for set action")
		}
		if params.Value == "" {
			return "", fmt.Errorf("'value' is required for set action")
		}

		if params.Scope == "global" {
			warning, err := t.store.SetGlobal(params.Name, params.Value)
			if err != nil {
				return "", fmt.Errorf("failed to set global variable: %w", err)
			}
			result := fmt.Sprintf("Set global variable: {{%s}} = '%s'\n(Persisted to disk)", params.Name, shared.MaskSecret(params.Value))
			if warning != "" {
				result = warning + "\n\n" + result
			}
			return result, nil
		}

		t.store.Set(params.Name, params.Value)
		return fmt.Sprintf("Set session variable: {{%s}} = '%s'\n(Available until Falcon exits)", params.Name, shared.MaskSecret(params.Value)), nil

	case "get":
		if params.Name == "" {
			return "", fmt.Errorf("'name' is required for get action")
		}

		value, ok := t.store.Get(params.Name)
		if !ok {
			return "", fmt.Errorf("variable '{{%s}}' not found", params.Name)
		}
		return fmt.Sprintf("Variable {{%s}} = '%s'", params.Name, value), nil

	case "delete":
		if params.Name == "" {
			return "", fmt.Errorf("'name' is required for delete action")
		}

		t.store.Delete(params.Name)
		return fmt.Sprintf("Deleted variable: {{%s}}", params.Name), nil

	case "list":
		vars := t.store.List()
		if len(vars) == 0 {
			return "No variables stored.", nil
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Stored variables (%d):\n\n", len(vars)))
		for name, value := range vars {
			sb.WriteString(fmt.Sprintf("  {{%s}} = %s\n", name, value))
		}
		return sb.String(), nil

	default:
		return "", fmt.Errorf("unknown action '%s' (use: set, get, delete, list)", params.Action)
	}
}
