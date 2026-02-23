package persistence

import (
	"strings"

	"github.com/blackcoderx/falcon/pkg/storage"
)

// ListEnvironmentsTool lists available environments
type ListEnvironmentsTool struct {
	manager *PersistenceManager
}

func NewListEnvironmentsTool(manager *PersistenceManager) *ListEnvironmentsTool {
	return &ListEnvironmentsTool{manager: manager}
}

func (t *ListEnvironmentsTool) Name() string { return "list_environments" }

func (t *ListEnvironmentsTool) Description() string {
	return "List all available environments in the .zap/environments directory."
}

func (t *ListEnvironmentsTool) Parameters() string {
	return `{}`
}

func (t *ListEnvironmentsTool) Execute(args string) (string, error) {
	envs, err := storage.ListEnvironments(t.manager.GetBaseDir())
	if err != nil {
		return "", err
	}

	if len(envs) == 0 {
		return "No environments found. Create YAML files in .zap/environments/ directory.", nil
	}

	var sb strings.Builder
	sb.WriteString("Available environments:\n")
	for _, env := range envs {
		marker := ""
		if env == t.manager.GetCurrentEnv() {
			marker = " (active)"
		}
		sb.WriteString("  - " + env + marker + "\n")
	}

	return sb.String(), nil
}
