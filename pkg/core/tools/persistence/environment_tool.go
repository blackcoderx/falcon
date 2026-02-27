package persistence

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/blackcoderx/falcon/pkg/storage"
)

// EnvironmentTool is the unified environment management tool that replaces
// set_environment and list_environments.
type EnvironmentTool struct {
	manager *PersistenceManager
}

func NewEnvironmentTool(manager *PersistenceManager) *EnvironmentTool {
	return &EnvironmentTool{manager: manager}
}

type EnvironmentParams struct {
	Action    string            `json:"action"`              // "set", "list"
	Name      string            `json:"name,omitempty"`
	Variables map[string]string `json:"variables,omitempty"` // optional: define env vars
}

func (t *EnvironmentTool) Name() string { return "environment" }

func (t *EnvironmentTool) Description() string {
	return "Manage environments in .falcon/environments/. Actions: set (activate environment and optionally persist its variables), list (show all available environments and which is active)"
}

func (t *EnvironmentTool) Parameters() string {
	return `{
  "action": "set|list",
  "name":      "dev|staging|prod|...",
  "variables": {"BASE_URL": "http://localhost:3000", "API_KEY": "{{API_KEY}}"}
}`
}

func (t *EnvironmentTool) Execute(args string) (string, error) {
	var params EnvironmentParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	switch params.Action {
	case "set":
		return t.set(params)
	case "list":
		return t.list()
	default:
		return "", fmt.Errorf("unknown action '%s' (use: set, list)", params.Action)
	}
}

func (t *EnvironmentTool) set(params EnvironmentParams) (string, error) {
	if params.Name == "" {
		return "", fmt.Errorf("name is required for set")
	}

	result := fmt.Sprintf("Environment set to '%s'", params.Name)

	// If variables are provided, persist them first so SetEnvironment can load them
	if len(params.Variables) > 0 {
		envPath := filepath.Join(storage.GetEnvironmentsDir(t.manager.GetBaseDir()), params.Name+".yaml")
		if err := storage.SaveEnvironment(params.Variables, envPath); err != nil {
			return "", fmt.Errorf("failed to save environment variables: %w", err)
		}
		result += fmt.Sprintf(" with %d variables", len(params.Variables))
	}

	// Activate the environment (loads variables into memory)
	if err := t.manager.SetEnvironment(params.Name); err != nil {
		// If the file doesn't exist yet and no variables were given, that's an error
		return "", err
	}

	return result, nil
}

func (t *EnvironmentTool) list() (string, error) {
	envs, err := storage.ListEnvironments(t.manager.GetBaseDir())
	if err != nil {
		return "", err
	}

	if len(envs) == 0 {
		return "No environments found. Use environment(action=\"set\", name=\"dev\", variables={...}) to create one.", nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Available environments (%d):\n", len(envs)))
	for _, env := range envs {
		marker := ""
		if env == t.manager.GetCurrentEnv() {
			marker = " (active)"
		}
		sb.WriteString("  - " + env + marker + "\n")
	}
	return sb.String(), nil
}
