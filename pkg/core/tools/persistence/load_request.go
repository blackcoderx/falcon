package persistence

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/blackcoderx/falcon/pkg/storage"
)

// LoadRequestTool loads requests from YAML files
type LoadRequestTool struct {
	manager *PersistenceManager
}

func NewLoadRequestTool(manager *PersistenceManager) *LoadRequestTool {
	return &LoadRequestTool{manager: manager}
}

func (t *LoadRequestTool) Name() string { return "load_request" }

func (t *LoadRequestTool) Description() string {
	return "Load a saved request from a YAML file. Returns the request details with environment variables substituted."
}

func (t *LoadRequestTool) Parameters() string {
	return `{"name": "string (required) - Name or filename of the saved request"}`
}

func (t *LoadRequestTool) Execute(args string) (string, error) {
	var params struct {
		Name string `json:"name"`
	}

	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("invalid parameters: %w", err)
	}

	if params.Name == "" {
		return "", fmt.Errorf("name is required")
	}

	// Try to find the file
	filename := params.Name
	if !strings.HasSuffix(filename, ".yaml") && !strings.HasSuffix(filename, ".yml") {
		filename = strings.ToLower(strings.ReplaceAll(filename, " ", "-")) + ".yaml"
	}

	filePath := filepath.Join(storage.GetRequestsDir(t.manager.GetBaseDir()), filename)
	req, err := storage.LoadRequest(filePath)
	if err != nil {
		return "", err
	}

	// Apply environment variables
	applied := storage.ApplyEnvironment(req, t.manager.GetEnvironment())

	// Format output
	result, _ := json.MarshalIndent(map[string]interface{}{
		"name":    applied.Name,
		"method":  applied.Method,
		"url":     applied.URL,
		"headers": applied.Headers,
		"body":    applied.Body,
	}, "", "  ")

	return string(result), nil
}
