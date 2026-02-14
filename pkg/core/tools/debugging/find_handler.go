package debugging

import (
	"encoding/json"
	"fmt"
	"strings"
)

// FindHandlerTool locates endpoint handlers in the codebase
type FindHandlerTool struct {
	workDir    string
	searchTool *SearchCodeTool
}

// NewFindHandlerTool creates a new find_handler tool
// It internally creates a SearchCodeTool, which is also in the debugging package.
func NewFindHandlerTool(workDir string) *FindHandlerTool {
	return &FindHandlerTool{
		workDir:    workDir,
		searchTool: NewSearchCodeTool(workDir),
	}
}

// FindHandlerParams defines input for find_handler
type FindHandlerParams struct {
	Endpoint  string `json:"endpoint"`
	Method    string `json:"method"`
	Path      string `json:"path"`
	Framework string `json:"framework"`
}

// HandlerInfo contains details about a discovered handler
type HandlerInfo struct {
	File         string   `json:"file"`
	Line         int      `json:"line"`
	Content      string   `json:"content"`
	RelatedFiles []string `json:"related_files"`
	Analysis     string   `json:"analysis"`
}

func (t *FindHandlerTool) Name() string {
	return "find_handler"
}

func (t *FindHandlerTool) Description() string {
	return "Search codebase to locate the exact file and function that handles a specific API endpoint, analyze the code to understand current implementation, and trace data flow."
}

func (t *FindHandlerTool) Parameters() string {
	return `{
  "endpoint": "POST /api/checkout",
  "method": "POST",
  "path": "/api/checkout",
  "framework": "gin"
}`
}

func (t *FindHandlerTool) Execute(args string) (string, error) {
	var params FindHandlerParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	// 1. Identify search patterns based on framework
	patterns := t.getSearchPatterns(params)
	if len(patterns) == 0 {
		// Generic fallback: just search the path
		patterns = []string{params.Path}
	}

	var results []string
	for _, pattern := range patterns {
		searchArgs := fmt.Sprintf(`{"pattern": "%s"}`, pattern)
		res, err := t.searchTool.Execute(searchArgs)
		if err == nil && !strings.Contains(res, "No matches found") {
			results = append(results, res)
		}
	}

	if len(results) == 0 {
		return "Could not find handler for the specified endpoint.", nil
	}

	// For MVP, return first match summary and let agent read file
	return strings.Join(results, "\n---\n"), nil
}

func (t *FindHandlerTool) getSearchPatterns(params FindHandlerParams) []string {
	method := strings.ToUpper(params.Method)
	path := params.Path

	switch strings.ToLower(params.Framework) {
	case "gin":
		return []string{
			fmt.Sprintf(`%s("%s"`, method, path),
			fmt.Sprintf(`%s( "%s"`, method, path),
		}
	case "echo":
		return []string{
			fmt.Sprintf(`%s("%s"`, method, path),
		}
	case "fastapi":
		return []string{
			fmt.Sprintf(`@app.%s("%s"`, strings.ToLower(method), path),
			fmt.Sprintf(`@router.%s("%s"`, strings.ToLower(method), path),
		}
	case "express":
		return []string{
			fmt.Sprintf(`\.%s(['"]%s['"]`, strings.ToLower(method), path),
			fmt.Sprintf(`app\.%s(['"]%s['"]`, strings.ToLower(method), path),
			fmt.Sprintf(`router\.%s(['"]%s['"]`, strings.ToLower(method), path),
		}
	default:
		return nil
	}
}
