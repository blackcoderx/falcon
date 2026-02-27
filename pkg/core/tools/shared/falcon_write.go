package shared

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// FalconWriteTool is a validated writer for .falcon/ artifacts (flows, spec, etc.).
// It enforces path safety, format validation, and protects critical config files.
type FalconWriteTool struct {
	falconDir string
}

func NewFalconWriteTool(falconDir string) *FalconWriteTool {
	return &FalconWriteTool{falconDir: falconDir}
}

type FalconWriteParams struct {
	// Path relative to .falcon/ — e.g. "flows/unit_get_users.yaml"
	Path    string `json:"path"`
	Content string `json:"content"`
	// Format: "yaml", "json", or "markdown" (default). Parsed before writing.
	Format string `json:"format,omitempty"`
}

// protectedFiles are .falcon root files that must never be overwritten by the agent.
var protectedFiles = map[string]bool{
	"config.yaml":   true,
	"manifest.json": true,
	"memory.json":   true,
}

func (t *FalconWriteTool) Name() string { return "falcon_write" }

func (t *FalconWriteTool) Description() string {
	return "Write validated artifacts into .falcon/ (flows, spec, custom notes). Enforces path safety, validates YAML/JSON syntax before writing. Use this to save test flows or API notes — NOT for reports (tools write those automatically)."
}

func (t *FalconWriteTool) Parameters() string {
	return `{
  "path":    "flows/unit_get_users.yaml",
  "content": "name: get_users\n...",
  "format":  "yaml|json|markdown"
}`
}

func (t *FalconWriteTool) Execute(args string) (string, error) {
	var params FalconWriteParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	if params.Path == "" {
		return "", fmt.Errorf("path is required")
	}
	if params.Content == "" {
		return "", fmt.Errorf("content is required")
	}

	// Security: no directory traversal
	if strings.Contains(params.Path, "..") {
		return "", fmt.Errorf("path must not contain '..': %s", params.Path)
	}

	// Security: no absolute paths
	if filepath.IsAbs(params.Path) {
		return "", fmt.Errorf("path must be relative to .falcon/, not absolute: %s", params.Path)
	}

	// Protect critical files
	base := filepath.Base(params.Path)
	if protectedFiles[base] && filepath.Dir(params.Path) == "." {
		return "", fmt.Errorf("cannot overwrite protected file '%s'", params.Path)
	}

	// Validate content format before touching the filesystem
	format := strings.ToLower(params.Format)
	switch format {
	case "yaml", "yml":
		var out interface{}
		if err := yaml.Unmarshal([]byte(params.Content), &out); err != nil {
			return "", fmt.Errorf("invalid YAML content: %w", err)
		}
	case "json":
		var out interface{}
		if err := json.Unmarshal([]byte(params.Content), &out); err != nil {
			return "", fmt.Errorf("invalid JSON content: %w", err)
		}
	default:
		// markdown or raw — no structural validation needed
	}

	fullPath := filepath.Join(t.falconDir, params.Path)

	// Ensure parent directory exists (only inside .falcon)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(fullPath, []byte(params.Content), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	info, _ := os.Stat(fullPath)
	size := int64(0)
	if info != nil {
		size = info.Size()
	}

	return fmt.Sprintf("Written %d bytes to .falcon/%s", size, params.Path), nil
}
