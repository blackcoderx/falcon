package shared

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ReportWriter handles the boilerplate of writing Markdown reports to .falcon/reports/.
type ReportWriter struct {
	FalconDir string
}

// NewReportWriter creates a new ReportWriter.
func NewReportWriter(falconDir string) *ReportWriter {
	return &ReportWriter{FalconDir: falconDir}
}

// Write creates a report file in .falcon/reports/ with the given content.
// If reportName is empty, a name is generated from defaultPrefix + timestamp.
// Returns the full path to the written report file.
func (w *ReportWriter) Write(reportName, defaultPrefix, content string) (string, error) {
	reportsDir := filepath.Join(w.FalconDir, "reports")
	if err := os.MkdirAll(reportsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create reports directory: %w", err)
	}

	name := reportName
	if name == "" {
		name = fmt.Sprintf("%s_%s", defaultPrefix, time.Now().Format("20060102_150405"))
	}
	name = strings.ReplaceAll(name, " ", "_")
	if !strings.HasSuffix(name, ".md") {
		name += ".md"
	}

	reportPath := filepath.Join(reportsDir, name)

	if err := os.WriteFile(reportPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write report: %w", err)
	}

	if err := ValidateReport(reportPath); err != nil {
		return "", err
	}

	return reportPath, nil
}
