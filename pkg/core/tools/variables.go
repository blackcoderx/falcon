package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/blackcoderx/zap/pkg/core"
)

// VariableStore manages session and global variables
type VariableStore struct {
	session map[string]string // In-memory session variables
	global  map[string]string // Persistent global variables
	mu      sync.RWMutex
	zapDir  string // Path to .zap directory
}

// NewVariableStore creates a new variable store
func NewVariableStore(zapDir string) *VariableStore {
	store := &VariableStore{
		session: make(map[string]string),
		global:  make(map[string]string),
		zapDir:  zapDir,
	}
	store.loadGlobalVariables()
	return store
}

// Set stores a variable (default: session scope)
func (vs *VariableStore) Set(name, value string) {
	vs.mu.Lock()
	defer vs.mu.Unlock()
	vs.session[name] = value
}

// SetGlobal stores a global variable (persisted to disk)
// Warns if the value appears to be a secret (should use session scope instead)
func (vs *VariableStore) SetGlobal(name, value string) (warning string, err error) {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	// Warn on potential secrets
	if core.IsSecret(name, value) {
		warning = fmt.Sprintf("WARNING: '%s' appears to be a secret. Consider using session scope instead (secrets are cleared on exit for security).", name)
	}

	vs.global[name] = value
	return warning, vs.saveGlobalVariables()
}

// Get retrieves a variable (checks session first, then global)
func (vs *VariableStore) Get(name string) (string, bool) {
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	// Check session first
	if value, ok := vs.session[name]; ok {
		return value, true
	}

	// Then check global
	if value, ok := vs.global[name]; ok {
		return value, true
	}

	return "", false
}

// Delete removes a variable
func (vs *VariableStore) Delete(name string) {
	vs.mu.Lock()
	defer vs.mu.Unlock()
	delete(vs.session, name)
	delete(vs.global, name)
	vs.saveGlobalVariables()
}

// List returns all variables (session + global)
func (vs *VariableStore) List() map[string]string {
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	result := make(map[string]string)
	// Global first
	for k, v := range vs.global {
		result[k] = v + " (global)"
	}
	// Session overrides global
	for k, v := range vs.session {
		result[k] = v + " (session)"
	}
	return result
}

// Substitute replaces {{VAR}} placeholders in text with variable values
func (vs *VariableStore) Substitute(text string) string {
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	result := text
	// Replace session variables
	for name, value := range vs.session {
		placeholder := "{{" + name + "}}"
		result = strings.ReplaceAll(result, placeholder, value)
	}
	// Replace global variables
	for name, value := range vs.global {
		placeholder := "{{" + name + "}}"
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

// loadGlobalVariables reads global variables from disk
func (vs *VariableStore) loadGlobalVariables() error {
	varFile := filepath.Join(vs.zapDir, "variables.json")
	data, err := os.ReadFile(varFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet, that's ok
		}
		return err
	}

	return json.Unmarshal(data, &vs.global)
}

// saveGlobalVariables writes global variables to disk
func (vs *VariableStore) saveGlobalVariables() error {
	varFile := filepath.Join(vs.zapDir, "variables.json")
	data, err := json.MarshalIndent(vs.global, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(varFile, data, 0644); err != nil {
		return err
	}

	// Update manifest counts
	core.UpdateManifestCounts(vs.zapDir)
	return nil
}

// VariableTool provides variable get/set/list operations
type VariableTool struct {
	store *VariableStore
}

// NewVariableTool creates a new variable tool
func NewVariableTool(store *VariableStore) *VariableTool {
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
			result := fmt.Sprintf("Set global variable: {{%s}} = '%s'\n(Persisted to disk)", params.Name, core.MaskSecret(params.Value))
			if warning != "" {
				result = warning + "\n\n" + result
			}
			return result, nil
		}

		t.store.Set(params.Name, params.Value)
		return fmt.Sprintf("Set session variable: {{%s}} = '%s'\n(Available until ZAP exits)", params.Name, core.MaskSecret(params.Value)), nil

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
