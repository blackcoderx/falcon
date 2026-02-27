package shared

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// FalconReadTool reads any file inside .falcon/ with path safety enforcement.
// Use this instead of read_file when accessing .falcon artifacts.
type FalconReadTool struct {
	falconDir string
}

func NewFalconReadTool(falconDir string) *FalconReadTool {
	return &FalconReadTool{falconDir: falconDir}
}

type FalconReadParams struct {
	// Path relative to .falcon/ — e.g. "flows/unit_get_users.yaml"
	Path string `json:"path"`
	// Format: "yaml", "json", or "raw" (default). Parsed output returned as formatted string.
	Format string `json:"format,omitempty"`
}

func (t *FalconReadTool) Name() string { return "falcon_read" }

func (t *FalconReadTool) Description() string {
	return "Read any file inside .falcon/ safely. Use this to inspect flows, spec.yaml, reports, or baselines. Format 'yaml'/'json' parses and re-formats the content; 'raw' returns it as-is."
}

func (t *FalconReadTool) Parameters() string {
	return `{
  "path":   "flows/unit_get_users.yaml",
  "format": "raw|yaml|json"
}`
}

func (t *FalconReadTool) Execute(args string) (string, error) {
	var params FalconReadParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	if params.Path == "" {
		return "", fmt.Errorf("path is required")
	}

	// Security: no directory traversal
	if strings.Contains(params.Path, "..") {
		return "", fmt.Errorf("path must not contain '..': %s", params.Path)
	}
	if filepath.IsAbs(params.Path) {
		return "", fmt.Errorf("path must be relative to .falcon/, not absolute: %s", params.Path)
	}

	fullPath := filepath.Join(t.falconDir, params.Path)

	data, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("file not found in .falcon/: %s", params.Path)
		}
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	content := string(data)
	format := strings.ToLower(params.Format)

	switch format {
	case "yaml", "yml":
		var parsed interface{}
		if err := yaml.Unmarshal(data, &parsed); err != nil {
			return "", fmt.Errorf("file is not valid YAML: %w", err)
		}
		formatted, err := yaml.Marshal(parsed)
		if err != nil {
			return content, nil // return raw on marshal error
		}
		return string(formatted), nil

	case "json":
		var parsed interface{}
		if err := json.Unmarshal(data, &parsed); err != nil {
			return "", fmt.Errorf("file is not valid JSON: %w", err)
		}
		formatted, err := json.MarshalIndent(parsed, "", "  ")
		if err != nil {
			return content, nil
		}
		return string(formatted), nil

	default:
		// raw — return as-is
		return content, nil
	}
}
