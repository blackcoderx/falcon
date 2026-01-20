package tools

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// SearchCodeTool searches for patterns in the codebase
type SearchCodeTool struct {
	workDir string
}

// NewSearchCodeTool creates a new code search tool
func NewSearchCodeTool(workDir string) *SearchCodeTool {
	if workDir == "" {
		workDir, _ = os.Getwd()
	}
	return &SearchCodeTool{workDir: workDir}
}

// Name returns the tool name
func (t *SearchCodeTool) Name() string {
	return "search_code"
}

// Description returns the tool description
func (t *SearchCodeTool) Description() string {
	return "Search for text/regex patterns in codebase. Returns matching files and lines."
}

// Parameters returns the tool parameter description
func (t *SearchCodeTool) Parameters() string {
	return `{"pattern": "string (required) - search pattern", "path": "string - directory to search", "file_pattern": "string - file glob like *.go"}`
}

// Execute searches for patterns in the codebase
func (t *SearchCodeTool) Execute(args string) (string, error) {
	var params struct {
		Pattern     string `json:"pattern"`
		Path        string `json:"path"`
		FilePattern string `json:"file_pattern"`
	}

	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	if params.Pattern == "" {
		return "", fmt.Errorf("pattern is required")
	}

	// Resolve search path
	searchPath := params.Path
	if searchPath == "" {
		searchPath = t.workDir
	} else if !filepath.IsAbs(searchPath) {
		searchPath = filepath.Join(t.workDir, searchPath)
	}

	// Security check
	absPath, err := filepath.Abs(searchPath)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	absWorkDir, _ := filepath.Abs(t.workDir)
	if !strings.HasPrefix(absPath, absWorkDir) {
		return "", fmt.Errorf("access denied: path outside project directory")
	}

	// Try ripgrep first (faster), fall back to native Go search
	result, err := t.searchWithRipgrep(params.Pattern, absPath, params.FilePattern)
	if err != nil {
		// Fallback to native search
		result, err = t.searchNative(params.Pattern, absPath, params.FilePattern)
		if err != nil {
			return "", err
		}
	}

	if result == "" {
		return "No matches found", nil
	}

	return result, nil
}

// searchWithRipgrep uses ripgrep for fast searching
func (t *SearchCodeTool) searchWithRipgrep(pattern, searchPath, filePattern string) (string, error) {
	args := []string{
		"-n",         // Line numbers
		"--no-heading", // No file headers
		"-M", "200",  // Max line length
		"--max-count", "10", // Max matches per file
	}

	// Add file pattern filter
	if filePattern != "" {
		args = append(args, "-g", filePattern)
	}

	// Exclude common directories
	args = append(args, "--glob", "!.git", "--glob", "!node_modules", "--glob", "!vendor")

	args = append(args, pattern, searchPath)

	cmd := exec.Command("rg", args...)
	output, err := cmd.Output()

	if err != nil {
		// Exit code 1 means no matches (not an error)
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return "", nil
		}
		return "", err
	}

	return t.formatSearchResults(string(output), searchPath)
}

// searchNative provides a pure Go search fallback
func (t *SearchCodeTool) searchNative(pattern, searchPath, filePattern string) (string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		// Fall back to literal search
		re = regexp.MustCompile(regexp.QuoteMeta(pattern))
	}

	var results []string
	maxMatches := 50
	matchCount := 0

	err = filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Skip hidden and common directories
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}

		// Check file pattern
		if filePattern != "" {
			matched, _ := filepath.Match(filePattern, info.Name())
			if !matched {
				return nil
			}
		}

		// Skip binary files (basic check)
		ext := strings.ToLower(filepath.Ext(info.Name()))
		binaryExts := map[string]bool{
			".exe": true, ".dll": true, ".so": true, ".dylib": true,
			".png": true, ".jpg": true, ".jpeg": true, ".gif": true,
			".pdf": true, ".zip": true, ".tar": true, ".gz": true,
		}
		if binaryExts[ext] {
			return nil
		}

		// Skip large files (> 1MB)
		if info.Size() > 1024*1024 {
			return nil
		}

		// Search file
		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		relPath, _ := filepath.Rel(t.workDir, path)
		scanner := bufio.NewScanner(file)
		lineNum := 0
		fileMatches := 0

		for scanner.Scan() {
			lineNum++
			line := scanner.Text()

			if re.MatchString(line) {
				fileMatches++
				if fileMatches <= 3 { // Max 3 matches per file
					// Truncate long lines
					if len(line) > 150 {
						line = line[:150] + "..."
					}
					results = append(results, fmt.Sprintf("%s:%d: %s", relPath, lineNum, line))
					matchCount++

					if matchCount >= maxMatches {
						results = append(results, fmt.Sprintf("... (stopped at %d matches)", maxMatches))
						return filepath.SkipAll
					}
				}
			}
		}

		return nil
	})

	if err != nil && err != filepath.SkipAll {
		return "", err
	}

	return strings.Join(results, "\n"), nil
}

// formatSearchResults formats ripgrep output
func (t *SearchCodeTool) formatSearchResults(output, searchPath string) (string, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		return "", nil
	}

	var results []string
	for i, line := range lines {
		if i >= 50 { // Limit results
			results = append(results, "... (more results truncated)")
			break
		}

		// Make paths relative
		if strings.HasPrefix(line, searchPath) {
			rel, err := filepath.Rel(t.workDir, strings.SplitN(line, ":", 2)[0])
			if err == nil {
				parts := strings.SplitN(line, ":", 3)
				if len(parts) >= 3 {
					line = fmt.Sprintf("%s:%s: %s", rel, parts[1], parts[2])
				}
			}
		}

		// Truncate long lines
		if len(line) > 200 {
			line = line[:200] + "..."
		}

		results = append(results, line)
	}

	return strings.Join(results, "\n"), nil
}
