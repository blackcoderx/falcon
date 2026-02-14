package security_scanner

import (
	"fmt"
	"strings"

	"github.com/blackcoderx/zap/pkg/core/tools/shared"
)

// Fuzzer performs input fuzzing with various injection payloads.
type Fuzzer struct {
	httpTool *shared.HTTPTool
}

// NewFuzzer creates a new fuzzer.
func NewFuzzer(httpTool *shared.HTTPTool) *Fuzzer {
	return &Fuzzer{httpTool: httpTool}
}

// FuzzEndpoints performs fuzzing attacks on endpoints.
func (f *Fuzzer) FuzzEndpoints(endpoints map[string]shared.EndpointAnalysis, baseURL string, maxPayload int) ([]Vulnerability, int) {
	var vulnerabilities []Vulnerability
	totalChecks := 0

	for endpointKey := range endpoints {
		parts := strings.SplitN(endpointKey, " ", 2)
		if len(parts) != 2 {
			continue
		}
		method := parts[0]
		path := parts[1]

		// Only fuzz endpoints that accept input
		if method != "POST" && method != "PUT" && method != "PATCH" {
			continue
		}

		url := strings.TrimSuffix(baseURL, "/") + "/" + strings.TrimPrefix(path, "/")

		// SQL Injection fuzzing
		vulns, checks := f.fuzzSQLInjection(method, url, endpointKey)
		vulnerabilities = append(vulnerabilities, vulns...)
		totalChecks += checks

		// NoSQL Injection fuzzing
		vulns, checks = f.fuzzNoSQLInjection(method, url, endpointKey)
		vulnerabilities = append(vulnerabilities, vulns...)
		totalChecks += checks

		// Command Injection fuzzing
		vulns, checks = f.fuzzCommandInjection(method, url, endpointKey)
		vulnerabilities = append(vulnerabilities, vulns...)
		totalChecks += checks

		// XSS fuzzing
		vulns, checks = f.fuzzXSS(method, url, endpointKey)
		vulnerabilities = append(vulnerabilities, vulns...)
		totalChecks += checks

		// Path Traversal fuzzing
		vulns, checks = f.fuzzPathTraversal(method, url, endpointKey)
		vulnerabilities = append(vulnerabilities, vulns...)
		totalChecks += checks

		// XXE fuzzing
		vulns, checks = f.fuzzXXE(method, url, endpointKey)
		vulnerabilities = append(vulnerabilities, vulns...)
		totalChecks += checks

		// Buffer Overflow / Large Payload fuzzing
		vulns, checks = f.fuzzLargePayload(method, url, endpointKey, maxPayload)
		vulnerabilities = append(vulnerabilities, vulns...)
		totalChecks += checks
	}

	return vulnerabilities, totalChecks
}

// fuzzSQLInjection tests for SQL injection vulnerabilities.
func (f *Fuzzer) fuzzSQLInjection(method, url, endpoint string) ([]Vulnerability, int) {
	var vulns []Vulnerability
	checks := 0

	sqlPayloads := []string{
		"' OR '1'='1",
		"admin'--",
		"' OR 1=1--",
		"1' UNION SELECT NULL--",
		"' AND 1=0 UNION ALL SELECT 'admin', '81dc9bdb52d04dc20036dbd8313ed055",
	}

	for _, payload := range sqlPayloads {
		checks++
		testPayload := map[string]interface{}{
			"username": payload,
			"email":    payload,
			"id":       payload,
		}

		req := shared.HTTPRequest{
			Method:  method,
			URL:     url,
			Headers: map[string]string{"Content-Type": "application/json"},
			Body:    testPayload,
		}

		resp, err := f.httpTool.Run(req)
		if err == nil {
			// Check for SQL error messages
			errorIndicators := []string{
				"SQL syntax", "mysql", "postgresql", "sqlite", "oracle",
				"ORA-", "PG::", "syntax error", "unclosed quotation",
			}

			for _, indicator := range errorIndicators {
				if strings.Contains(strings.ToLower(resp.Body), strings.ToLower(indicator)) {
					vulns = append(vulns, Vulnerability{
						ID:          fmt.Sprintf("FUZZ-SQL-001-%s", sanitizeEndpoint(endpoint)),
						Title:       "SQL Injection Vulnerability Detected",
						Severity:    "critical",
						Category:    "injection",
						Endpoint:    endpoint,
						Description: "Endpoint is vulnerable to SQL injection attacks",
						Evidence:    fmt.Sprintf("SQL error detected with payload: %s", payload),
						Remediation: "Use parameterized queries or prepared statements. Never concatenate user input into SQL queries.",
						OWASPRef:    "A03:2021",
						CWERef:      "CWE-89",
					})
					return vulns, checks // Stop after first detection
				}
			}
		}
	}

	return vulns, checks
}

// fuzzNoSQLInjection tests for NoSQL injection vulnerabilities.
func (f *Fuzzer) fuzzNoSQLInjection(method, url, endpoint string) ([]Vulnerability, int) {
	var vulns []Vulnerability
	checks := 0

	noSQLPayloads := []map[string]interface{}{
		{"username": map[string]interface{}{"$ne": ""}},
		{"username": map[string]interface{}{"$gt": ""}},
		{"password": map[string]interface{}{"$regex": ".*"}},
	}

	for _, payload := range noSQLPayloads {
		checks++
		req := shared.HTTPRequest{
			Method:  method,
			URL:     url,
			Headers: map[string]string{"Content-Type": "application/json"},
			Body:    payload,
		}

		resp, err := f.httpTool.Run(req)
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			// If accepted and successful, might be vulnerable
			vulns = append(vulns, Vulnerability{
				ID:          fmt.Sprintf("FUZZ-NOSQL-001-%s", sanitizeEndpoint(endpoint)),
				Title:       "Potential NoSQL Injection",
				Severity:    "high",
				Category:    "injection",
				Endpoint:    endpoint,
				Description: "Endpoint may be vulnerable to NoSQL injection",
				Evidence:    "NoSQL operator payload was accepted",
				Remediation: "Sanitize inputs and use schema validation for MongoDB/NoSQL queries",
				OWASPRef:    "A03:2021",
				CWERef:      "CWE-943",
			})
			return vulns, checks
		}
	}

	return vulns, checks
}

// fuzzCommandInjection tests for OS command injection.
func (f *Fuzzer) fuzzCommandInjection(method, url, endpoint string) ([]Vulnerability, int) {
	var vulns []Vulnerability
	checks := 0

	cmdPayloads := []string{
		"; ls -la",
		"| whoami",
		"& dir",
		"`id`",
		"$(cat /etc/passwd)",
	}

	for _, payload := range cmdPayloads {
		checks++
		testPayload := map[string]interface{}{
			"filename": payload,
			"path":     payload,
			"command":  payload,
		}

		req := shared.HTTPRequest{
			Method:  method,
			URL:     url,
			Headers: map[string]string{"Content-Type": "application/json"},
			Body:    testPayload,
		}

		resp, err := f.httpTool.Run(req)
		if err == nil {
			// Check for command output indicators
			cmdIndicators := []string{"root:", "bin/bash", "uid=", "gid=", "total "}
			for _, indicator := range cmdIndicators {
				if strings.Contains(resp.Body, indicator) {
					vulns = append(vulns, Vulnerability{
						ID:          fmt.Sprintf("FUZZ-CMD-001-%s", sanitizeEndpoint(endpoint)),
						Title:       "Command Injection Vulnerability",
						Severity:    "critical",
						Category:    "injection",
						Endpoint:    endpoint,
						Description: "Endpoint is vulnerable to OS command injection",
						Evidence:    fmt.Sprintf("Command output detected with payload: %s", payload),
						Remediation: "Never pass user input to shell commands. Use safe APIs and validate inputs.",
						OWASPRef:    "A03:2021",
						CWERef:      "CWE-78",
					})
					return vulns, checks
				}
			}
		}
	}

	return vulns, checks
}

// fuzzXSS tests for Cross-Site Scripting vulnerabilities.
func (f *Fuzzer) fuzzXSS(method, url, endpoint string) ([]Vulnerability, int) {
	var vulns []Vulnerability
	checks := 0

	xssPayloads := []string{
		"<script>alert('XSS')</script>",
		"<img src=x onerror=alert('XSS')>",
		"javascript:alert('XSS')",
		"<svg onload=alert('XSS')>",
		"\"><script>alert(String.fromCharCode(88,83,83))</script>",
	}

	for _, payload := range xssPayloads {
		checks++
		testPayload := map[string]interface{}{
			"name":    payload,
			"comment": payload,
			"message": payload,
		}

		req := shared.HTTPRequest{
			Method:  method,
			URL:     url,
			Headers: map[string]string{"Content-Type": "application/json"},
			Body:    testPayload,
		}

		resp, err := f.httpTool.Run(req)
		if err == nil {
			// Check if payload is reflected unescaped
			if strings.Contains(resp.Body, "<script>") || strings.Contains(resp.Body, "onerror=") {
				vulns = append(vulns, Vulnerability{
					ID:          fmt.Sprintf("FUZZ-XSS-001-%s", sanitizeEndpoint(endpoint)),
					Title:       "Cross-Site Scripting (XSS) Vulnerability",
					Severity:    "high",
					Category:    "injection",
					Endpoint:    endpoint,
					Description: "Endpoint reflects user input without proper encoding",
					Evidence:    "XSS payload reflected in response",
					Remediation: "Encode all user input before displaying. Use Content-Security-Policy headers.",
					OWASPRef:    "A03:2021",
					CWERef:      "CWE-79",
				})
				return vulns, checks
			}
		}
	}

	return vulns, checks
}

// fuzzPathTraversal tests for path traversal vulnerabilities.
func (f *Fuzzer) fuzzPathTraversal(method, url, endpoint string) ([]Vulnerability, int) {
	var vulns []Vulnerability
	checks := 0

	pathPayloads := []string{
		"../../../etc/passwd",
		"..\\..\\..\\windows\\system32\\config\\sam",
		"....//....//....//etc/passwd",
		"%2e%2e%2f%2e%2e%2f%2e%2e%2fetc%2fpasswd",
	}

	for _, payload := range pathPayloads {
		checks++
		testPayload := map[string]interface{}{
			"file":     payload,
			"path":     payload,
			"filename": payload,
		}

		req := shared.HTTPRequest{
			Method:  method,
			URL:     url,
			Headers: map[string]string{"Content-Type": "application/json"},
			Body:    testPayload,
		}

		resp, err := f.httpTool.Run(req)
		if err == nil {
			// Check for file system indicators
			if strings.Contains(resp.Body, "root:x:0:0") || strings.Contains(resp.Body, "Windows Registry") {
				vulns = append(vulns, Vulnerability{
					ID:          fmt.Sprintf("FUZZ-PATH-001-%s", sanitizeEndpoint(endpoint)),
					Title:       "Path Traversal Vulnerability",
					Severity:    "critical",
					Category:    "path_traversal",
					Endpoint:    endpoint,
					Description: "Endpoint allows reading arbitrary files from the server",
					Evidence:    "System file content detected in response",
					Remediation: "Validate file paths, use allowlists, and restrict file access to specific directories",
					OWASPRef:    "A01:2021",
					CWERef:      "CWE-22",
				})
				return vulns, checks
			}
		}
	}

	return vulns, checks
}

// fuzzXXE tests for XML External Entity vulnerabilities.
func (f *Fuzzer) fuzzXXE(method, url, endpoint string) ([]Vulnerability, int) {
	var vulns []Vulnerability
	checks := 0

	xxePayload := `<?xml version="1.0"?>
<!DOCTYPE foo [
<!ENTITY xxe SYSTEM "file:///etc/passwd">
]>
<foo>&xxe;</foo>`

	checks++
	req := shared.HTTPRequest{
		Method:  method,
		URL:     url,
		Headers: map[string]string{"Content-Type": "application/xml"},
		Body:    xxePayload,
	}

	resp, err := f.httpTool.Run(req)
	if err == nil {
		if strings.Contains(resp.Body, "root:x:0:0") {
			vulns = append(vulns, Vulnerability{
				ID:          fmt.Sprintf("FUZZ-XXE-001-%s", sanitizeEndpoint(endpoint)),
				Title:       "XML External Entity (XXE) Vulnerability",
				Severity:    "critical",
				Category:    "injection",
				Endpoint:    endpoint,
				Description: "Endpoint is vulnerable to XXE attacks",
				Evidence:    "External entity was processed and file content returned",
				Remediation: "Disable XML external entity processing in your XML parser",
				OWASPRef:    "A05:2021",
				CWERef:      "CWE-611",
			})
		}
	}

	return vulns, checks
}

// fuzzLargePayload tests for buffer overflow and DoS with large payloads.
func (f *Fuzzer) fuzzLargePayload(method, url, endpoint string, maxSize int) ([]Vulnerability, int) {
	var vulns []Vulnerability
	checks := 0

	// Create a large payload
	largeString := strings.Repeat("A", maxSize)
	largePayload := map[string]interface{}{
		"data":    largeString,
		"content": largeString,
		"message": largeString,
	}

	checks++
	req := shared.HTTPRequest{
		Method:  method,
		URL:     url,
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    largePayload,
	}

	resp, err := f.httpTool.Run(req)
	if err != nil {
		// If request failed, might indicate DoS vulnerability
		vulns = append(vulns, Vulnerability{
			ID:          fmt.Sprintf("FUZZ-DOS-001-%s", sanitizeEndpoint(endpoint)),
			Title:       "Potential Denial of Service via Large Payload",
			Severity:    "medium",
			Category:    "availability",
			Endpoint:    endpoint,
			Description: "Endpoint failed to handle large payload gracefully",
			Evidence:    fmt.Sprintf("Request with %d bytes failed: %v", maxSize, err),
			Remediation: "Implement payload size limits and rate limiting",
			OWASPRef:    "A05:2021",
			CWERef:      "CWE-400",
		})
	} else if resp.StatusCode == 500 {
		vulns = append(vulns, Vulnerability{
			ID:          fmt.Sprintf("FUZZ-DOS-002-%s", sanitizeEndpoint(endpoint)),
			Title:       "Server Error on Large Payload",
			Severity:    "medium",
			Category:    "availability",
			Endpoint:    endpoint,
			Description: "Endpoint returns 500 error when processing large payload",
			Evidence:    fmt.Sprintf("500 error with %d byte payload", maxSize),
			Remediation: "Add proper error handling and payload size validation",
			OWASPRef:    "A05:2021",
			CWERef:      "CWE-400",
		})
	}

	return vulns, checks
}
