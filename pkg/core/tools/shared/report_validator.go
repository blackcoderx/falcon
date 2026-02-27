package shared

import (
	"fmt"
	"os"
	"strings"
)

// minReportBytes is the minimum number of bytes a valid report must contain.
const minReportBytes = 64

// minFalconMDBytes is the minimum size for falcon.md to be considered populated.
const minFalconMDBytes = 200

// ValidateReport checks that a report file exists and is non-empty.
// Call this immediately after writing any report file.
func ValidateReport(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("report was not created: %s", path)
		}
		return fmt.Errorf("failed to stat report file: %w", err)
	}

	if info.Size() == 0 {
		_ = os.Remove(path)
		return fmt.Errorf("report file was created but is empty (0 bytes): %s", path)
	}

	if info.Size() < minReportBytes {
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("report file is suspiciously small (%d bytes): %s", info.Size(), path)
		}
		if strings.TrimSpace(string(data)) == "" {
			_ = os.Remove(path)
			return fmt.Errorf("report file contains only whitespace: %s", path)
		}
	}

	return nil
}

// ValidateReportContent performs a deeper content check on a report file.
// It verifies the file has a heading, at least one result indicator, and no
// unresolved template placeholders. Returns a descriptive error if validation fails
// so the calling tool can surface it to the agent for a retry.
func ValidateReportContent(path string) error {
	if err := ValidateReport(path); err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read report for content validation: %w", err)
	}
	content := string(data)

	// Must have at least one Markdown heading
	if !strings.Contains(content, "# ") && !strings.Contains(content, "## ") {
		return fmt.Errorf("report is missing a Markdown heading (# or ##): %s", path)
	}

	// Must contain at least one result indicator
	resultIndicators := []string{"|", "```", "PASS", "FAIL", "ERROR", "OK", "✓", "✗", "passed", "failed"}
	hasResult := false
	for _, indicator := range resultIndicators {
		if strings.Contains(content, indicator) {
			hasResult = true
			break
		}
	}
	if !hasResult {
		return fmt.Errorf("report appears incomplete — no result indicators (table, code block, PASS/FAIL) found: %s", path)
	}

	// Must not contain unresolved template placeholders
	placeholders := []string{"{{", "TODO", "[placeholder]", "[INSERT", "[TBD]"}
	for _, p := range placeholders {
		if strings.Contains(content, p) {
			return fmt.Errorf("report contains unresolved placeholder '%s' — content may be incomplete: %s", p, path)
		}
	}

	return nil
}

// ValidateFalconMD checks that falcon.md is well-formed after an update_knowledge write.
// It verifies the file has the two required section headings and that the updated section
// is non-empty. Returns a descriptive error so the memory tool can alert the agent.
func ValidateFalconMD(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("falcon.md was not found at: %s", path)
		}
		return fmt.Errorf("failed to stat falcon.md: %w", err)
	}

	if info.Size() < minFalconMDBytes {
		return fmt.Errorf("falcon.md is too small (%d bytes) — it may not have been written correctly", info.Size())
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read falcon.md: %w", err)
	}
	content := string(data)

	// These two sections are the minimum required knowledge base entries
	requiredSections := []string{"# Base URLs", "# Known Endpoints"}
	for _, section := range requiredSections {
		if !strings.Contains(content, section) {
			return fmt.Errorf("falcon.md is missing required section '%s' — knowledge base may be corrupted", section)
		}
	}

	// Each required section must have at least one non-blank line of content after it
	for _, section := range requiredSections {
		idx := strings.Index(content, section)
		if idx == -1 {
			continue
		}
		afterSection := content[idx+len(section):]
		lines := strings.Split(afterSection, "\n")
		hasContent := false
		for _, line := range lines[1:] { // skip the heading line itself
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "#") {
				// Hit the next section — stop
				break
			}
			if trimmed != "" {
				hasContent = true
				break
			}
		}
		if !hasContent {
			return fmt.Errorf("section '%s' in falcon.md is empty — update_knowledge may not have written content", section)
		}
	}

	return nil
}
