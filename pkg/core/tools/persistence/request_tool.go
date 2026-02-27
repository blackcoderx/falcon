package persistence

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/blackcoderx/falcon/pkg/core/tools/shared"
	"github.com/blackcoderx/falcon/pkg/storage"
)

// RequestTool is the unified request management tool that replaces save_request,
// load_request, and list_requests.
type RequestTool struct {
	manager *PersistenceManager
}

func NewRequestTool(manager *PersistenceManager) *RequestTool {
	return &RequestTool{manager: manager}
}

type RequestParams struct {
	Action  string            `json:"action"`           // "save", "load", "list"
	Name    string            `json:"name,omitempty"`
	Method  string            `json:"method,omitempty"`
	URL     string            `json:"url,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    interface{}       `json:"body,omitempty"`
}

func (t *RequestTool) Name() string { return "request" }

func (t *RequestTool) Description() string {
	return "Manage saved HTTP requests in .falcon/requests/. Actions: save (persist a request template), load (retrieve and substitute env vars), list (show all saved requests)"
}

func (t *RequestTool) Parameters() string {
	return `{
  "action": "save|load|list",
  "name":   "string — request name (required for save/load)",
  "method": "GET|POST|PUT|DELETE|PATCH (required for save)",
  "url":    "https://... or /path (required for save, supports {{VAR}})",
  "headers": {},
  "body":    {}
}`
}

func (t *RequestTool) Execute(args string) (string, error) {
	var params RequestParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	switch params.Action {
	case "save":
		return t.save(params)
	case "load":
		return t.load(params)
	case "list":
		return t.list()
	default:
		return "", fmt.Errorf("unknown action '%s' (use: save, load, list)", params.Action)
	}
}

func (t *RequestTool) save(params RequestParams) (string, error) {
	if params.Name == "" {
		return "", fmt.Errorf("name is required for save")
	}
	if params.Method == "" {
		return "", fmt.Errorf("method is required for save")
	}
	if params.URL == "" {
		return "", fmt.Errorf("url is required for save")
	}

	if secretErr := shared.ValidateRequestForSecrets(params.URL, params.Headers, params.Body); secretErr != "" {
		return "", fmt.Errorf("cannot save request: %s", secretErr)
	}

	req := storage.Request{
		Name:    params.Name,
		Method:  strings.ToUpper(params.Method),
		URL:     params.URL,
		Headers: params.Headers,
		Body:    params.Body,
	}

	filename := strings.ToLower(strings.ReplaceAll(params.Name, " ", "-")) + ".yaml"
	filePath := filepath.Join(storage.GetRequestsDir(t.manager.GetBaseDir()), filename)

	if err := storage.SaveRequest(req, filePath); err != nil {
		return "", err
	}

	info, statErr := os.Stat(filePath)
	if statErr != nil {
		return "", fmt.Errorf("file not found after save at %s: %w", filePath, statErr)
	}
	if info.Size() == 0 {
		return "", fmt.Errorf("file at %s is empty after save — write may have failed", filePath)
	}

	shared.UpdateManifestCounts(t.manager.GetBaseDir())
	return fmt.Sprintf("Request '%s' saved to %s (%d bytes)", params.Name, filePath, info.Size()), nil
}

func (t *RequestTool) load(params RequestParams) (string, error) {
	if params.Name == "" {
		return "", fmt.Errorf("name is required for load")
	}

	filename := params.Name
	if !strings.HasSuffix(filename, ".yaml") && !strings.HasSuffix(filename, ".yml") {
		filename = strings.ToLower(strings.ReplaceAll(filename, " ", "-")) + ".yaml"
	}

	filePath := filepath.Join(storage.GetRequestsDir(t.manager.GetBaseDir()), filename)
	req, err := storage.LoadRequest(filePath)
	if err != nil {
		return "", err
	}

	applied := storage.ApplyEnvironment(req, t.manager.GetEnvironment())

	result, _ := json.MarshalIndent(map[string]interface{}{
		"name":    applied.Name,
		"method":  applied.Method,
		"url":     applied.URL,
		"headers": applied.Headers,
		"body":    applied.Body,
	}, "", "  ")

	return string(result), nil
}

func (t *RequestTool) list() (string, error) {
	requests, err := storage.ListRequests(t.manager.GetBaseDir())
	if err != nil {
		return "", err
	}

	if len(requests) == 0 {
		return "No saved requests. Use request(action=\"save\") to save one.", nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Saved requests (%d):\n", len(requests)))
	for _, req := range requests {
		sb.WriteString("  - " + req + "\n")
	}
	return sb.String(), nil
}
