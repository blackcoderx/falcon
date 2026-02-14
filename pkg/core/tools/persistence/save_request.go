package persistence

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/blackcoderx/zap/pkg/core"
	"github.com/blackcoderx/zap/pkg/storage"
)

// SaveRequestTool saves requests to YAML files
type SaveRequestTool struct {
	manager *PersistenceManager
}

func NewSaveRequestTool(manager *PersistenceManager) *SaveRequestTool {
	return &SaveRequestTool{manager: manager}
}

func (t *SaveRequestTool) Name() string { return "save_request" }

func (t *SaveRequestTool) Description() string {
	return "Save an API request to a YAML file for later use. Saved requests can be loaded and executed with load_request."
}

func (t *SaveRequestTool) Parameters() string {
	return `{
  "name": "string (required) - Name for the request",
  "method": "string (required) - HTTP method (GET, POST, PUT, DELETE)",
  "url": "string (required) - Request URL (can use {{VAR}} placeholders)",
  "headers": "object (optional) - Request headers",
  "body": "object (optional) - Request body for POST/PUT"
}`
}

func (t *SaveRequestTool) Execute(args string) (string, error) {
	var params struct {
		Name    string            `json:"name"`
		Method  string            `json:"method"`
		URL     string            `json:"url"`
		Headers map[string]string `json:"headers"`
		Body    interface{}       `json:"body"`
	}

	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("invalid parameters: %w", err)
	}

	if params.Name == "" {
		return "", fmt.Errorf("name is required")
	}
	if params.Method == "" {
		return "", fmt.Errorf("method is required")
	}
	if params.URL == "" {
		return "", fmt.Errorf("url is required")
	}

	// Validate for plaintext secrets
	if secretErr := core.ValidateRequestForSecrets(params.URL, params.Headers, params.Body); secretErr != "" {
		return "", fmt.Errorf("cannot save request: %s", secretErr)
	}

	req := storage.Request{
		Name:    params.Name,
		Method:  strings.ToUpper(params.Method),
		URL:     params.URL,
		Headers: params.Headers,
		Body:    params.Body,
	}

	// Generate filename from name
	filename := strings.ToLower(strings.ReplaceAll(params.Name, " ", "-")) + ".yaml"
	filePath := filepath.Join(storage.GetRequestsDir(t.manager.GetBaseDir()), filename)

	if err := storage.SaveRequest(req, filePath); err != nil {
		return "", err
	}

	// Update manifest counts
	core.UpdateManifestCounts(t.manager.GetBaseDir())

	return fmt.Sprintf("Request saved to %s", filePath), nil
}
