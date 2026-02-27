package security_scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/blackcoderx/falcon/pkg/core/tools/shared"
)

// GenerateSecurityReport persists the vulnerabilities and scan parameters into a Markdown report.
func GenerateSecurityReport(falconDir string, vulns []Vulnerability, params ScanParams) (string, error) {
	// Save into shared reports directory
	reportsDir := filepath.Join(falconDir, "reports")
	if err := os.MkdirAll(reportsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create reports directory: %w", err)
	}

	// Generate filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("security_report_%s.md", timestamp)
	reportPath := filepath.Join(reportsDir, filename)

	// Build markdown report
	severity := categorizeBySeverity(vulns)
	var sb strings.Builder

	fmt.Fprintf(&sb, "# Security Scan Report\n\n")
	fmt.Fprintf(&sb, "**Date:** %s\n\n", time.Now().Format(time.RFC1123))
	fmt.Fprintf(&sb, "**Target:** %s\n\n", params.BaseURL)
	fmt.Fprintf(&sb, "**Scan Types:** %s\n\n", strings.Join(params.ScanTypes, ", "))
	fmt.Fprintf(&sb, "## Summary\n\n")
	fmt.Fprintf(&sb, "| Severity | Count |\n|----------|-------|\n")
	fmt.Fprintf(&sb, "| Critical | %d |\n", severity["critical"])
	fmt.Fprintf(&sb, "| High     | %d |\n", severity["high"])
	fmt.Fprintf(&sb, "| Medium   | %d |\n", severity["medium"])
	fmt.Fprintf(&sb, "| Low      | %d |\n\n", severity["low"])

	if len(vulns) > 0 {
		fmt.Fprintf(&sb, "## Vulnerabilities\n\n")
		for _, v := range vulns {
			fmt.Fprintf(&sb, "### [%s] %s\n\n", strings.ToUpper(v.Severity), v.Title)
			fmt.Fprintf(&sb, "- **Category:** %s\n", v.Category)
			fmt.Fprintf(&sb, "- **Endpoint:** %s\n", v.Endpoint)
			fmt.Fprintf(&sb, "- **Description:** %s\n", v.Description)
			if v.Evidence != "" {
				fmt.Fprintf(&sb, "- **Evidence:** %s\n", v.Evidence)
			}
			fmt.Fprintf(&sb, "- **Remediation:** %s\n", v.Remediation)
			if v.OWASPRef != "" {
				fmt.Fprintf(&sb, "- **OWASP Ref:** %s\n", v.OWASPRef)
			}
			if v.CWERef != "" {
				fmt.Fprintf(&sb, "- **CWE Ref:** %s\n", v.CWERef)
			}
			fmt.Fprintf(&sb, "\n")
		}
	} else {
		fmt.Fprintf(&sb, "## Result\n\nNo vulnerabilities detected.\n")
	}

	if err := os.WriteFile(reportPath, []byte(sb.String()), 0644); err != nil {
		return "", fmt.Errorf("failed to save security report: %w", err)
	}

	if err := shared.ValidateReport(reportPath); err != nil {
		return "", err
	}

	return reportPath, nil
}
