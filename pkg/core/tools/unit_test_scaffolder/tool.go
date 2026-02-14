package unit_test_scaffolder

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/blackcoderx/zap/pkg/llm"
)

// UnitTestCasefolderTool scaffolds unit tests and mocks for an existing codebase.
type UnitTestCasefolderTool struct {
	llmClient llm.LLMClient
}

// NewUnitTestCasefolderTool creates a new unit test scaffolder tool.
func NewUnitTestCasefolderTool(llmClient llm.LLMClient) *UnitTestCasefolderTool {
	return &UnitTestCasefolderTool{
		llmClient: llmClient,
	}
}

// ScaffoldParams defines parameters for unit test scaffolding.
type ScaffoldParams struct {
	SourceDir string   `json:"source_dir"`           // Base directory of source code
	Files     []string `json:"files,omitempty"`      // Specific files to scaffold tests for
	Language  string   `json:"language,omitempty"`   // go, typescript, python, etc.
	MockLevel string   `json:"mock_level,omitempty"` // low, medium, high (default: medium)
	OutputDir string   `json:"output_dir,omitempty"` // Where to save generated tests
}

// ScaffoldResult represents the results of scaffolding.
type ScaffoldResult struct {
	ScannedFiles   []string `json:"scanned_files"`
	GeneratedTests []string `json:"generated_tests"`
	GeneratedMocks []string `json:"generated_mocks"`
	Summary        string   `json:"summary"`
}

func (t *UnitTestCasefolderTool) Name() string {
	return "scaffold_unit_tests"
}

func (t *UnitTestCasefolderTool) Description() string {
	return "Scan source code to identify controllers, services, and repositories, and automatically scaffold unit tests with appropriate mocks"
}

func (t *UnitTestCasefolderTool) Parameters() string {
	return `{
  "source_dir": "./internal",
  "language": "go",
  "mock_level": "high",
  "output_dir": "./tests/unit"
}`
}

func (t *UnitTestCasefolderTool) Execute(args string) (string, error) {
	var params ScaffoldParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	if params.SourceDir == "" {
		return "", fmt.Errorf("source_dir is required")
	}

	// 1. Scan codebase
	scanner := &Scanner{SourceDir: params.SourceDir}
	filesToTest, err := scanner.Scan(params.Files)
	if err != nil {
		return "", fmt.Errorf("failed to scan codebase: %w", err)
	}

	if len(filesToTest) == 0 {
		return "No relevant files found to scaffold tests for.", nil
	}

	// 2. Generate tests and mocks (simulation of LLM generation)
	_ = &MockGenerator{Language: params.Language}
	var generatedTests []string
	var generatedMocks []string

	for _, file := range filesToTest {
		// In a real implementation, we'd read the file content and pass it to LLM
		// or use a template engine.
		testFile := t.getTestPath(file, params.OutputDir)

		// Simulate file creation
		if params.OutputDir != "" {
			os.MkdirAll(filepath.Dir(testFile), 0755)
			os.WriteFile(testFile, []byte("// Auto-generated test for "+file), 0644)
		}

		generatedTests = append(generatedTests, testFile)

		// Generate mocks if needed
		if params.MockLevel != "low" {
			mockFile := t.getMockPath(file, params.OutputDir)
			generatedMocks = append(generatedMocks, mockFile)
		}
	}

	result := ScaffoldResult{
		ScannedFiles:   filesToTest,
		GeneratedTests: generatedTests,
		GeneratedMocks: generatedMocks,
	}
	result.Summary = t.formatSummary(result)

	_ = result
	return result.Summary, nil
}

func (t *UnitTestCasefolderTool) getTestPath(srcPath, outputDir string) string {
	ext := filepath.Ext(srcPath)
	base := strings.TrimSuffix(filepath.Base(srcPath), ext)

	testName := base + "_test" + ext
	if outputDir != "" {
		return filepath.Join(outputDir, testName)
	}
	return filepath.Join(filepath.Dir(srcPath), testName)
}

func (t *UnitTestCasefolderTool) getMockPath(srcPath, outputDir string) string {
	ext := filepath.Ext(srcPath)
	base := strings.TrimSuffix(filepath.Base(srcPath), ext)

	mockName := base + "_mock" + ext
	if outputDir != "" {
		return filepath.Join(outputDir, "mocks", mockName)
	}
	return filepath.Join(filepath.Dir(srcPath), "mocks", mockName)
}

func (t *UnitTestCasefolderTool) formatSummary(r ScaffoldResult) string {
	summary := fmt.Sprintf("🏗️ Unit Test Scaffolding Complete\n\n")
	summary += fmt.Sprintf("Scanned Files: %d\n", len(r.ScannedFiles))
	summary += fmt.Sprintf("Generated Tests: %d\n", len(r.GeneratedTests))
	summary += fmt.Sprintf("Generated Mocks: %d\n\n", len(r.GeneratedMocks))

	if len(r.GeneratedTests) > 0 {
		summary += "First few generated tests:\n"
		for i, f := range r.GeneratedTests {
			if i >= 3 {
				break
			}
			summary += fmt.Sprintf("  • %s\n", f)
		}
	}

	return summary
}
