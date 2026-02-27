package functional_test_generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/blackcoderx/falcon/pkg/core/tools/shared"
)

// ExportScenarios exports test scenarios to a Markdown report.
func ExportScenarios(falconDir string, scenarios []shared.TestScenario) (string, error) {
	// Save into shared reports directory
	reportsDir := filepath.Join(falconDir, "reports")
	if err := os.MkdirAll(reportsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create reports directory: %w", err)
	}

	// Generate filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("functional_report_%s.md", timestamp)
	reportPath := filepath.Join(reportsDir, filename)

	// Build markdown report
	var sb strings.Builder

	fmt.Fprintf(&sb, "# Functional Test Report\n\n")
	fmt.Fprintf(&sb, "**Generated:** %s\n\n", time.Now().Format(time.RFC1123))
	fmt.Fprintf(&sb, "**Total Scenarios:** %d\n\n", len(scenarios))

	if len(scenarios) == 0 {
		fmt.Fprintf(&sb, "No test scenarios were generated.\n")
	} else {
		fmt.Fprintf(&sb, "## Scenarios\n\n")
		for _, s := range scenarios {
			fmt.Fprintf(&sb, "### %s\n\n", s.Name)
			fmt.Fprintf(&sb, "- **ID:** %s\n", s.ID)
			fmt.Fprintf(&sb, "- **Category:** %s\n", s.Category)
			fmt.Fprintf(&sb, "- **Severity:** %s\n", s.Severity)
			fmt.Fprintf(&sb, "- **Description:** %s\n", s.Description)
			fmt.Fprintf(&sb, "- **Method:** %s\n", s.Method)
			fmt.Fprintf(&sb, "- **URL:** %s\n", s.URL)
			if s.OWASPRef != "" {
				fmt.Fprintf(&sb, "- **OWASP Ref:** %s\n", s.OWASPRef)
			}
			if s.CWERef != "" {
				fmt.Fprintf(&sb, "- **CWE Ref:** %s\n", s.CWERef)
			}
			fmt.Fprintf(&sb, "\n")
		}
	}

	if err := os.WriteFile(reportPath, []byte(sb.String()), 0644); err != nil {
		return "", fmt.Errorf("failed to write report: %w", err)
	}

	if err := shared.ValidateReport(reportPath); err != nil {
		return "", err
	}

	return reportPath, nil
}

// ExportTemplate generates exportable test code templates.
// This is a placeholder for future implementation where we can generate
// actual test code files (e.g., Go test files, Jest tests, pytest, etc.)
type ExportTemplate struct {
	Language string // go, javascript, python, etc.
	Format   string // framework-specific (e.g., jest, pytest, etc.)
}

// GenerateTestFile creates executable test code from scenarios.
// Future implementation for Sprint 5.4
func (t *ExportTemplate) GenerateTestFile(scenarios []shared.TestScenario, outputPath string) error {
	// TODO: Implement code generation based on language/format
	return fmt.Errorf("code template generation not yet implemented")
}

// ScenarioTemplate represents a template for a test scenario.
type ScenarioTemplate struct {
	Name        string
	Description string
	Method      string
	URL         string
	Headers     map[string]string
	Body        string
	Assertions  []string
}

// ToGoTest generates Go test code (example for future implementation).
func (st *ScenarioTemplate) ToGoTest() string {
	return fmt.Sprintf(`
func Test%s(t *testing.T) {
	// %s
	req := httptest.NewRequest("%s", "%s", strings.NewReader(%s))
	// Add headers
	// Add assertions
}
`, st.Name, st.Description, st.Method, st.URL, st.Body)
}
