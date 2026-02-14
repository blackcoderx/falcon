package functional_test_generator

import (
	"encoding/json"
	"fmt"

	"github.com/blackcoderx/zap/pkg/core/tools/shared"
	"github.com/blackcoderx/zap/pkg/core/tools/spec_ingester"
)

// FunctionalTestGeneratorTool generates comprehensive functional tests
// from the API Knowledge Graph using various testing strategies.
type FunctionalTestGeneratorTool struct {
	zapDir         string
	httpTool       *shared.HTTPTool
	assertTool     *shared.AssertTool
	strategyEngine *StrategyEngine
	generator      *TestGenerator
}

// NewFunctionalTestGeneratorTool creates a new functional test generator tool.
func NewFunctionalTestGeneratorTool(zapDir string, httpTool *shared.HTTPTool, assertTool *shared.AssertTool) *FunctionalTestGeneratorTool {
	return &FunctionalTestGeneratorTool{
		zapDir:         zapDir,
		httpTool:       httpTool,
		assertTool:     assertTool,
		strategyEngine: NewStrategyEngine(),
		generator:      NewTestGenerator(httpTool, assertTool),
	}
}

// GenerateParams defines the parameters for test generation.
type GenerateParams struct {
	BaseURL    string   `json:"base_url"`             // Base URL for API (e.g., http://localhost:3000)
	Strategies []string `json:"strategies,omitempty"` // List of strategies (happy, negative, boundary). Empty = all
	Endpoints  []string `json:"endpoints,omitempty"`  // Specific endpoints to test. Empty = all endpoints
	Execute    bool     `json:"execute"`              // Whether to execute tests immediately
	Export     bool     `json:"export,omitempty"`     // Whether to export test scenarios to file
}

// GenerateResult represents the output of test generation.
type GenerateResult struct {
	TotalScenarios int                 `json:"total_scenarios"`
	GeneratedBy    map[string]int      `json:"generated_by_strategy"` // strategy -> count
	Results        []shared.TestResult `json:"results,omitempty"`     // Only if Execute=true
	ExportPath     string              `json:"export_path,omitempty"` // Only if Export=true
	Summary        string              `json:"summary"`
}

// Name returns the tool name.
func (t *FunctionalTestGeneratorTool) Name() string {
	return "generate_functional_tests"
}

// Description returns the tool description.
func (t *FunctionalTestGeneratorTool) Description() string {
	return "Generate and optionally execute comprehensive functional tests from the API Knowledge Graph using happy path, negative, and boundary strategies"
}

// Parameters returns the tool parameter description.
func (t *FunctionalTestGeneratorTool) Parameters() string {
	return `{
  "base_url": "http://localhost:3000",
  "strategies": ["happy", "negative", "boundary"],
  "endpoints": ["GET /api/users", "POST /api/users"],
  "execute": true,
  "export": false
}`
}

// Execute generates functional tests from the Knowledge Graph.
func (t *FunctionalTestGeneratorTool) Execute(args string) (string, error) {
	var params GenerateParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	// Validate required parameters
	if params.BaseURL == "" {
		return "", fmt.Errorf("base_url is required")
	}

	// Default to all strategies if none specified
	if len(params.Strategies) == 0 {
		params.Strategies = []string{"happy", "negative", "boundary"}
	}

	// 1. Load API Knowledge Graph
	builder := spec_ingester.NewGraphBuilder(t.zapDir)
	graph, err := builder.LoadGraph()
	if err != nil {
		return "", fmt.Errorf("failed to load API Knowledge Graph: %w", err)
	}
	if graph == nil || len(graph.Endpoints) == 0 {
		return "", fmt.Errorf("API Knowledge Graph is empty. Run 'ingest_spec' first to index an API specification")
	}

	// 2. Filter endpoints if specific ones requested
	endpointsToTest := graph.Endpoints
	if len(params.Endpoints) > 0 {
		filtered := make(map[string]shared.EndpointAnalysis)
		for _, endpoint := range params.Endpoints {
			if analysis, ok := graph.Endpoints[endpoint]; ok {
				filtered[endpoint] = analysis
			}
		}
		if len(filtered) == 0 {
			return "", fmt.Errorf("none of the requested endpoints found in Knowledge Graph")
		}
		endpointsToTest = filtered
	}

	// 3. Generate test scenarios using selected strategies
	scenarios, strategyBreakdown := t.generateScenarios(endpointsToTest, params.BaseURL, params.Strategies)

	if len(scenarios) == 0 {
		return "", fmt.Errorf("no test scenarios generated")
	}

	// 4. Execute tests if requested
	var results []shared.TestResult
	if params.Execute {
		results = t.generator.ExecuteScenarios(scenarios)
	}

	// 5. Export to file if requested
	exportPath := ""
	if params.Export {
		path, err := ExportScenarios(t.zapDir, scenarios)
		if err != nil {
			return "", fmt.Errorf("failed to export scenarios: %w", err)
		}
		exportPath = path
	}

	// 6. Format and return result
	result := GenerateResult{
		TotalScenarios: len(scenarios),
		GeneratedBy:    strategyBreakdown,
		Results:        results,
		ExportPath:     exportPath,
		Summary:        t.formatSummary(len(scenarios), strategyBreakdown, results, exportPath),
	}

	return result.Summary, nil
}

// generateScenarios creates test scenarios from endpoints using the specified strategies.
func (t *FunctionalTestGeneratorTool) generateScenarios(
	endpoints map[string]shared.EndpointAnalysis,
	baseURL string,
	strategies []string,
) ([]shared.TestScenario, map[string]int) {
	var allScenarios []shared.TestScenario
	strategyBreakdown := make(map[string]int)

	// Generate scenarios for each endpoint using each strategy
	for endpointKey, analysis := range endpoints {
		for _, strategyName := range strategies {
			scenarios := t.strategyEngine.Generate(endpointKey, analysis, baseURL, strategyName)
			allScenarios = append(allScenarios, scenarios...)
			strategyBreakdown[strategyName] += len(scenarios)
		}
	}

	return allScenarios, strategyBreakdown
}

// formatSummary creates a human-readable summary of the generation and execution results.
func (t *FunctionalTestGeneratorTool) formatSummary(
	totalScenarios int,
	breakdown map[string]int,
	results []shared.TestResult,
	exportPath string,
) string {
	summary := fmt.Sprintf("✓ Generated %d test scenarios\n\n", totalScenarios)

	// Strategy breakdown
	summary += "Strategy Breakdown:\n"
	for strategy, count := range breakdown {
		summary += fmt.Sprintf("  • %s: %d scenarios\n", strategy, count)
	}

	// Execution results if available
	if len(results) > 0 {
		passed := 0
		failed := 0
		for _, result := range results {
			if result.Passed {
				passed++
			} else {
				failed++
			}
		}

		summary += "\nExecution Results:\n"
		summary += fmt.Sprintf("  • Passed: %d/%d (%.1f%%)\n", passed, totalScenarios, float64(passed)/float64(totalScenarios)*100)
		summary += fmt.Sprintf("  • Failed: %d/%d (%.1f%%)\n", failed, totalScenarios, float64(failed)/float64(totalScenarios)*100)

		// Show first few failures
		if failed > 0 {
			summary += "\nSample Failures:\n"
			failureCount := 0
			for _, result := range results {
				if !result.Passed && failureCount < 3 {
					summary += fmt.Sprintf("  • %s: %s\n", result.ScenarioName, result.Error)
					failureCount++
				}
			}
			if failed > 3 {
				summary += fmt.Sprintf("  ... and %d more failures\n", failed-3)
			}
		}
	}

	// Export path if available
	if exportPath != "" {
		summary += fmt.Sprintf("\nExported scenarios to: %s\n", exportPath)
	}

	return summary
}
