package security_scanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// GenerateSecurityReport persists the vulnerabilities and scan parameters into a JSON report.
func GenerateSecurityReport(zapDir string, vulns []Vulnerability, params ScanParams) (string, error) {
	// Create reports directory
	reportsDir := filepath.Join(zapDir, "security_reports")
	if err := os.MkdirAll(reportsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create security reports directory: %w", err)
	}

	// Generate filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("security_scan_%s.json", timestamp)
	reportPath := filepath.Join(reportsDir, filename)

	// Build report structure
	reportData := struct {
		Timestamp       string          `json:"timestamp"`
		Parameters      ScanParams      `json:"parameters"`
		Vulnerabilities []Vulnerability `json:"vulnerabilities"`
		Summary         map[string]int  `json:"summary"`
	}{
		Timestamp:       time.Now().Format(time.RFC3339),
		Parameters:      params,
		Vulnerabilities: vulns,
		Summary:         categorizeBySeverity(vulns),
	}

	// Marshal and save
	data, err := json.MarshalIndent(reportData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal security report: %w", err)
	}

	if err := os.WriteFile(reportPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to save security report: %w", err)
	}

	return reportPath, nil
}
