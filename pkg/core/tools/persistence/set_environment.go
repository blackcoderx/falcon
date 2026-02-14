package persistence

import (
	"encoding/json"
	"fmt"
)

// SetEnvironmentTool sets the active environment
type SetEnvironmentTool struct {
	manager *PersistenceManager
}

func NewSetEnvironmentTool(manager *PersistenceManager) *SetEnvironmentTool {
	return &SetEnvironmentTool{manager: manager}
}

func (t *SetEnvironmentTool) Name() string { return "set_environment" }

func (t *SetEnvironmentTool) Description() string {
	return "Set the active environment. Environment variables will be substituted in saved requests."
}

func (t *SetEnvironmentTool) Parameters() string {
	return `{"name": "string (required) - Name of the environment (e.g., 'dev', 'prod')"}`
}

func (t *SetEnvironmentTool) Execute(args string) (string, error) {
	var params struct {
		Name string `json:"name"`
	}

	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("invalid parameters: %w", err)
	}

	if params.Name == "" {
		return "", fmt.Errorf("name is required")
	}

	if err := t.manager.SetEnvironment(params.Name); err != nil {
		return "", err
	}

	return fmt.Sprintf("Environment set to '%s'", params.Name), nil
}
