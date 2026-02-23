package security_scanner

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/blackcoderx/falcon/pkg/core/tools/shared"
	"github.com/blackcoderx/falcon/pkg/core/tools/spec_ingester"
)

// SecurityScannerTool performs comprehensive security scans on APIs.
type SecurityScannerTool struct {
	zapDir       string
	httpTool     *shared.HTTPTool
	owaspChecker *OWASPChecker
	fuzzer       *Fuzzer
	authAuditor  *AuthAuditor
}

// NewSecurityScannerTool creates a new security scanner tool.
func NewSecurityScannerTool(zapDir string, httpTool *shared.HTTPTool) *SecurityScannerTool {
	return &SecurityScannerTool{
		zapDir:       zapDir,
		httpTool:     httpTool,
		owaspChecker: NewOWASPChecker(httpTool),
		fuzzer:       NewFuzzer(httpTool),
		authAuditor:  NewAuthAuditor(httpTool),
	}
}

// ScanParams defines parameters for security scanning.
type ScanParams struct {
	BaseURL    string   `json:"base_url"`              // Base URL of the API
	Endpoints  []string `json:"endpoints,omitempty"`   // Specific endpoints to scan (empty = all)
	ScanTypes  []string `json:"scan_types,omitempty"`  // Types of scans: owasp, fuzz, auth (empty = all)
	AuthToken  string   `json:"auth_token,omitempty"`  // Auth token for authenticated endpoints
	Depth      string   `json:"depth,omitempty"`       // Scan depth: quick, standard, deep (default: standard)
	MaxPayload int      `json:"max_payload,omitempty"` // Max payload size for fuzzing (default: 10000)
}

// ScanResult represents the output of a security scan.
type ScanResult struct {
	TotalChecks     int             `json:"total_checks"`
	VulnFound       int             `json:"vulnerabilities_found"`
	Critical        int             `json:"critical"`
	High            int             `json:"high"`
	Medium          int             `json:"medium"`
	Low             int             `json:"low"`
	Vulnerabilities []Vulnerability `json:"vulnerabilities"`
	ScanDuration    string          `json:"scan_duration"`
	Summary         string          `json:"summary"`
	ReportPath      string          `json:"report_path,omitempty"`
}

// Vulnerability represents a security finding.
type Vulnerability struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Severity    string `json:"severity"` // critical, high, medium, low
	Category    string `json:"category"` // owasp, auth, injection, etc.
	Endpoint    string `json:"endpoint"`
	Description string `json:"description"`
	Evidence    string `json:"evidence,omitempty"`
	Remediation string `json:"remediation"`
	OWASPRef    string `json:"owasp_ref,omitempty"`
	CWERef      string `json:"cwe_ref,omitempty"`
}

// Name returns the tool name.
func (t *SecurityScannerTool) Name() string {
	return "scan_security"
}

// Description returns the tool description.
func (t *SecurityScannerTool) Description() string {
	return "Perform comprehensive security scans including OWASP Top 10 checks, input fuzzing, and authentication/authorization testing"
}

// Parameters returns the tool parameter description.
func (t *SecurityScannerTool) Parameters() string {
	return `{
  "base_url": "http://localhost:3000",
  "endpoints": ["POST /api/login", "GET /api/users"],
  "scan_types": ["owasp", "fuzz", "auth"],
  "auth_token": "Bearer eyJ0eXAiOiJKV1QiLCJhbGc...",
  "depth": "standard",
  "max_payload": 10000
}`
}

// Execute performs the security scan.
func (t *SecurityScannerTool) Execute(args string) (string, error) {
	var params ScanParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	// Validate required parameters
	if params.BaseURL == "" {
		return "", fmt.Errorf("base_url is required")
	}

	// Default values
	if len(params.ScanTypes) == 0 {
		params.ScanTypes = []string{"owasp", "fuzz", "auth"}
	}
	if params.Depth == "" {
		params.Depth = "standard"
	}
	if params.MaxPayload == 0 {
		params.MaxPayload = 10000
	}

	startTime := time.Now()

	// 1. Load endpoints (from Knowledge Graph or use specified ones)
	endpoints, err := t.getEndpoints(params.Endpoints)
	if err != nil {
		return "", fmt.Errorf("failed to get endpoints: %w", err)
	}

	if len(endpoints) == 0 {
		return "", fmt.Errorf("no endpoints to scan")
	}

	// 2. Execute scans based on scan types
	var allVulnerabilities []Vulnerability
	totalChecks := 0

	for _, scanType := range params.ScanTypes {
		switch scanType {
		case "owasp":
			vulns, checks := t.owaspChecker.RunChecks(endpoints, params.BaseURL)
			allVulnerabilities = append(allVulnerabilities, vulns...)
			totalChecks += checks

		case "fuzz":
			vulns, checks := t.fuzzer.FuzzEndpoints(endpoints, params.BaseURL, params.MaxPayload)
			allVulnerabilities = append(allVulnerabilities, vulns...)
			totalChecks += checks

		case "auth":
			if params.AuthToken != "" {
				vulns, checks := t.authAuditor.AuditAuth(endpoints, params.BaseURL, params.AuthToken)
				allVulnerabilities = append(allVulnerabilities, vulns...)
				totalChecks += checks
			}
		}
	}

	// 3. Categorize by severity
	severityCounts := categorizeBySeverity(allVulnerabilities)

	// 4. Generate report
	reportPath, err := GenerateSecurityReport(t.zapDir, allVulnerabilities, params)
	if err != nil {
		// Non-fatal, continue
		reportPath = ""
	}

	// 5. Build result
	result := ScanResult{
		TotalChecks:     totalChecks,
		VulnFound:       len(allVulnerabilities),
		Critical:        severityCounts["critical"],
		High:            severityCounts["high"],
		Medium:          severityCounts["medium"],
		Low:             severityCounts["low"],
		Vulnerabilities: allVulnerabilities,
		ScanDuration:    time.Since(startTime).String(),
		ReportPath:      reportPath,
		Summary:         t.formatSummary(totalChecks, len(allVulnerabilities), severityCounts, allVulnerabilities),
	}
	_ = result // Suppress unused write to field info lint
	return result.Summary, nil
}

// getEndpoints retrieves endpoints either from the Knowledge Graph or the provided list.
func (t *SecurityScannerTool) getEndpoints(specifiedEndpoints []string) (map[string]shared.EndpointAnalysis, error) {
	if len(specifiedEndpoints) > 0 {
		// Use specified endpoints - create minimal analysis
		endpoints := make(map[string]shared.EndpointAnalysis)
		for _, ep := range specifiedEndpoints {
			endpoints[ep] = shared.EndpointAnalysis{
				Summary: "Specified endpoint for security scan",
			}
		}
		return endpoints, nil
	}

	// Load from Knowledge Graph
	builder := spec_ingester.NewGraphBuilder(t.zapDir)
	graph, err := builder.LoadGraph()
	if err != nil {
		return nil, err
	}
	if graph == nil || len(graph.Endpoints) == 0 {
		return nil, fmt.Errorf("no endpoints in Knowledge Graph - run ingest_spec first or specify endpoints")
	}

	return graph.Endpoints, nil
}

// categorizeBySeverity counts vulnerabilities by severity level.
func categorizeBySeverity(vulns []Vulnerability) map[string]int {
	counts := map[string]int{
		"critical": 0,
		"high":     0,
		"medium":   0,
		"low":      0,
	}

	for _, v := range vulns {
		counts[v.Severity]++
	}

	return counts
}

// formatSummary creates a human-readable summary of the scan results.
func (t *SecurityScannerTool) formatSummary(totalChecks, vulnCount int, severityCounts map[string]int, vulns []Vulnerability) string {
	summary := "🔒 Security Scan Complete\n\n"
	summary += fmt.Sprintf("Total Checks: %d\n", totalChecks)
	summary += fmt.Sprintf("Vulnerabilities Found: %d\n\n", vulnCount)

	if vulnCount == 0 {
		summary += "✓ No vulnerabilities detected! API appears secure.\n"
		return summary
	}

	// Severity breakdown
	summary += "Severity Breakdown:\n"
	if severityCounts["critical"] > 0 {
		summary += fmt.Sprintf("  🔴 Critical: %d\n", severityCounts["critical"])
	}
	if severityCounts["high"] > 0 {
		summary += fmt.Sprintf("  🟠 High: %d\n", severityCounts["high"])
	}
	if severityCounts["medium"] > 0 {
		summary += fmt.Sprintf("  🟡 Medium: %d\n", severityCounts["medium"])
	}
	if severityCounts["low"] > 0 {
		summary += fmt.Sprintf("  🟢 Low: %d\n", severityCounts["low"])
	}

	// Show top vulnerabilities
	summary += "\nTop Vulnerabilities:\n"
	count := 0
	for _, v := range vulns {
		if count >= 5 {
			break
		}
		if v.Severity == "critical" || v.Severity == "high" {
			summary += fmt.Sprintf("  • [%s] %s - %s\n", v.Severity, v.Title, v.Endpoint)
			count++
		}
	}

	if vulnCount > 5 {
		summary += fmt.Sprintf("  ... and %d more findings\n", vulnCount-5)
	}

	return summary
}
