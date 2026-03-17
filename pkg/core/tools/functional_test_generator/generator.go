package functional_test_generator

import (
	"github.com/blackcoderx/falcon/pkg/core/tools/shared"
)

// TestGenerator executes generated test scenarios using the shared TestExecutor.
type TestGenerator struct {
	executor *shared.TestExecutor
}

// NewTestGenerator creates a new test generator.
func NewTestGenerator(executor *shared.TestExecutor) *TestGenerator {
	return &TestGenerator{executor: executor}
}

// ExecuteScenarios runs all test scenarios sequentially and returns the results.
// Scenarios must have fully qualified URLs (base_url already prepended).
func (g *TestGenerator) ExecuteScenarios(scenarios []shared.TestScenario) []shared.TestResult {
	// Use concurrency=1 with empty baseURL since scenarios already have full URLs
	return g.executor.RunScenarios(scenarios, "", 1)
}
