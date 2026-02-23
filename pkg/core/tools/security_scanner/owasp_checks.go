package security_scanner

import (
	"fmt"
	"strings"

	"github.com/blackcoderx/falcon/pkg/core/tools/shared"
)

// OWASPChecker performs OWASP Top 10 security checks.
type OWASPChecker struct {
	httpTool *shared.HTTPTool
}

// NewOWASPChecker creates a new OWASP checker.
func NewOWASPChecker(httpTool *shared.HTTPTool) *OWASPChecker {
	return &OWASPChecker{httpTool: httpTool}
}

// RunChecks executes OWASP Top 10 checks on the endpoints.
func (c *OWASPChecker) RunChecks(endpoints map[string]shared.EndpointAnalysis, baseURL string) ([]Vulnerability, int) {
	var vulnerabilities []Vulnerability
	totalChecks := 0

	for endpointKey := range endpoints {
		// Parse endpoint
		parts := strings.SplitN(endpointKey, " ", 2)
		if len(parts) != 2 {
			continue
		}
		method := parts[0]
		path := parts[1]
		url := strings.TrimSuffix(baseURL, "/") + "/" + strings.TrimPrefix(path, "/")

		// A01:2021 - Broken Access Control
		vulns, checks := c.checkBrokenAccessControl(method, url, endpointKey)
		vulnerabilities = append(vulnerabilities, vulns...)
		totalChecks += checks

		// A02:2021 - Cryptographic Failures
		vulns, checks = c.checkCryptographicFailures(method, url, endpointKey)
		vulnerabilities = append(vulnerabilities, vulns...)
		totalChecks += checks

		// A03:2021 - Injection
		vulns, checks = c.checkInjection(method, url, endpointKey)
		vulnerabilities = append(vulnerabilities, vulns...)
		totalChecks += checks

		// A04:2021 - Insecure Design (check for sensitive data exposure)
		vulns, checks = c.checkInsecureDesign(method, url, endpointKey)
		vulnerabilities = append(vulnerabilities, vulns...)
		totalChecks += checks

		// A05:2021 - Security Misconfiguration
		vulns, checks = c.checkSecurityMisconfiguration(method, url, endpointKey)
		vulnerabilities = append(vulnerabilities, vulns...)
		totalChecks += checks

		// A07:2021 - Identification and Authentication Failures
		vulns, checks = c.checkAuthenticationFailures(method, url, endpointKey)
		vulnerabilities = append(vulnerabilities, vulns...)
		totalChecks += checks

		// A10:2021 - Server-Side Request Forgery (SSRF)
		vulns, checks = c.checkSSRF(method, url, endpointKey)
		vulnerabilities = append(vulnerabilities, vulns...)
		totalChecks += checks
	}

	return vulnerabilities, totalChecks
}

// A01:2021 - Broken Access Control
func (c *OWASPChecker) checkBrokenAccessControl(method, url, endpoint string) ([]Vulnerability, int) {
	var vulns []Vulnerability
	checks := 0

	// Test: Access without authentication
	checks++
	req := shared.HTTPRequest{
		Method:  method,
		URL:     url,
		Headers: map[string]string{},
	}

	resp, err := c.httpTool.Run(req)
	if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Endpoint accessible without auth - potential issue
		vulns = append(vulns, Vulnerability{
			ID:          fmt.Sprintf("BAC-001-%s", sanitizeEndpoint(endpoint)),
			Title:       "Potential Broken Access Control",
			Severity:    "high",
			Category:    "access_control",
			Endpoint:    endpoint,
			Description: "Endpoint is accessible without authentication",
			Evidence:    fmt.Sprintf("Request without auth returned %d", resp.StatusCode),
			Remediation: "Implement proper authentication and authorization checks",
			OWASPRef:    "A01:2021",
			CWERef:      "CWE-284",
		})
	}

	return vulns, checks
}

// A02:2021 - Cryptographic Failures
func (c *OWASPChecker) checkCryptographicFailures(_, url, endpoint string) ([]Vulnerability, int) {
	var vulns []Vulnerability
	checks := 0

	// Test: Check if using HTTPS
	checks++
	if strings.HasPrefix(url, "http://") {
		vulns = append(vulns, Vulnerability{
			ID:          fmt.Sprintf("CF-001-%s", sanitizeEndpoint(endpoint)),
			Title:       "Insecure Transport (HTTP instead of HTTPS)",
			Severity:    "high",
			Category:    "crypto",
			Endpoint:    endpoint,
			Description: "API endpoint uses unencrypted HTTP protocol",
			Evidence:    fmt.Sprintf("URL: %s", url),
			Remediation: "Use HTTPS to encrypt data in transit",
			OWASPRef:    "A02:2021",
			CWERef:      "CWE-319",
		})
	}

	return vulns, checks
}

// A03:2021 - Injection
func (c *OWASPChecker) checkInjection(method, url, endpoint string) ([]Vulnerability, int) {
	var vulns []Vulnerability
	checks := 0

	if method == "POST" || method == "PUT" || method == "PATCH" {
		// SQL Injection test
		checks++
		sqlPayload := map[string]interface{}{
			"username": "admin' OR '1'='1",
			"password": "password",
		}

		req := shared.HTTPRequest{
			Method:  method,
			URL:     url,
			Headers: map[string]string{"Content-Type": "application/json"},
			Body:    sqlPayload,
		}

		resp, err := c.httpTool.Run(req)
		if err == nil {
			// Check for SQL error messages in response
			if strings.Contains(resp.Body, "SQL") || strings.Contains(resp.Body, "syntax") ||
				strings.Contains(resp.Body, "mysql") || strings.Contains(resp.Body, "postgresql") {
				vulns = append(vulns, Vulnerability{
					ID:          fmt.Sprintf("INJ-001-%s", sanitizeEndpoint(endpoint)),
					Title:       "Potential SQL Injection",
					Severity:    "critical",
					Category:    "injection",
					Endpoint:    endpoint,
					Description: "Endpoint may be vulnerable to SQL injection",
					Evidence:    "SQL error messages detected in response",
					Remediation: "Use parameterized queries / prepared statements",
					OWASPRef:    "A03:2021",
					CWERef:      "CWE-89",
				})
			}
		}

		// XSS test
		checks++
		xssPayload := map[string]interface{}{
			"name": "<script>alert('XSS')</script>",
		}

		req = shared.HTTPRequest{
			Method:  method,
			URL:     url,
			Headers: map[string]string{"Content-Type": "application/json"},
			Body:    xssPayload,
		}

		resp, err = c.httpTool.Run(req)
		if err == nil && strings.Contains(resp.Body, "<script>") {
			vulns = append(vulns, Vulnerability{
				ID:          fmt.Sprintf("INJ-002-%s", sanitizeEndpoint(endpoint)),
				Title:       "Potential Cross-Site Scripting (XSS)",
				Severity:    "high",
				Category:    "injection",
				Endpoint:    endpoint,
				Description: "Endpoint may reflect user input without proper sanitization",
				Evidence:    "Script tag reflected in response",
				Remediation: "Sanitize and encode all user inputs before displaying",
				OWASPRef:    "A03:2021",
				CWERef:      "CWE-79",
			})
		}
	}

	return vulns, checks
}

// A04:2021 - Insecure Design
func (c *OWASPChecker) checkInsecureDesign(method, url, endpoint string) ([]Vulnerability, int) {
	var vulns []Vulnerability
	checks := 0

	// Test for sensitive data exposure
	checks++
	req := shared.HTTPRequest{
		Method:  method,
		URL:     url,
		Headers: map[string]string{},
	}

	resp, err := c.httpTool.Run(req)
	if err == nil {
		// Check for sensitive keywords in response
		sensitiveKeywords := []string{"password", "secret", "api_key", "token", "ssn", "credit_card"}
		for _, keyword := range sensitiveKeywords {
			if strings.Contains(strings.ToLower(resp.Body), keyword) {
				vulns = append(vulns, Vulnerability{
					ID:          fmt.Sprintf("ID-001-%s", sanitizeEndpoint(endpoint)),
					Title:       "Potential Sensitive Data Exposure",
					Severity:    "medium",
					Category:    "insecure_design",
					Endpoint:    endpoint,
					Description: fmt.Sprintf("Response may contain sensitive data: %s", keyword),
					Evidence:    "Sensitive keywords found in response body",
					Remediation: "Remove sensitive data from responses or implement proper access controls",
					OWASPRef:    "A04:2021",
					CWERef:      "CWE-200",
				})
				break // Only report once per endpoint
			}
		}
	}

	return vulns, checks
}

// A05:2021 - Security Misconfiguration
func (c *OWASPChecker) checkSecurityMisconfiguration(method, url, endpoint string) ([]Vulnerability, int) {
	var vulns []Vulnerability
	checks := 0

	// Check for verbose error messages
	checks++
	req := shared.HTTPRequest{
		Method:  method,
		URL:     url,
		Headers: map[string]string{},
	}

	resp, err := c.httpTool.Run(req)
	if err == nil {
		// Check for security headers
		checks++
		securityHeaders := map[string]string{
			"X-Content-Type-Options":  "nosniff",
			"X-Frame-Options":         "DENY",
			"Content-Security-Policy": "",
		}

		for header := range securityHeaders {
			if _, ok := resp.Headers[header]; !ok {
				vulns = append(vulns, Vulnerability{
					ID:          fmt.Sprintf("SM-001-%s-%s", sanitizeEndpoint(endpoint), header),
					Title:       fmt.Sprintf("Missing Security Header: %s", header),
					Severity:    "low",
					Category:    "misconfiguration",
					Endpoint:    endpoint,
					Description: fmt.Sprintf("Response missing recommended security header: %s", header),
					Evidence:    "Header not present in HTTP response",
					Remediation: fmt.Sprintf("Add %s header to all responses", header),
					OWASPRef:    "A05:2021",
					CWERef:      "CWE-16",
				})
			}
		}

		// Check for stack traces in error responses
		if resp.StatusCode >= 500 {
			if strings.Contains(resp.Body, "Traceback") || strings.Contains(resp.Body, "Stack trace") {
				vulns = append(vulns, Vulnerability{
					ID:          fmt.Sprintf("SM-002-%s", sanitizeEndpoint(endpoint)),
					Title:       "Stack Trace Exposure",
					Severity:    "medium",
					Category:    "misconfiguration",
					Endpoint:    endpoint,
					Description: "Server error response contains stack trace",
					Evidence:    "Stack trace detected in error response",
					Remediation: "Disable debug mode in production and return generic error messages",
					OWASPRef:    "A05:2021",
					CWERef:      "CWE-209",
				})
			}
		}
	}

	return vulns, checks
}

// A07:2021 - Identification and Authentication Failures
func (c *OWASPChecker) checkAuthenticationFailures(method, url, endpoint string) ([]Vulnerability, int) {
	var vulns []Vulnerability
	checks := 0

	// Only check auth endpoints
	if !strings.Contains(strings.ToLower(endpoint), "login") &&
		!strings.Contains(strings.ToLower(endpoint), "auth") &&
		!strings.Contains(strings.ToLower(endpoint), "signin") {
		return vulns, checks
	}

	if method == "POST" {
		// Test for weak password acceptance
		checks++
		weakPasswordPayload := map[string]interface{}{
			"username": "testuser",
			"password": "123",
		}

		req := shared.HTTPRequest{
			Method:  method,
			URL:     url,
			Headers: map[string]string{"Content-Type": "application/json"},
			Body:    weakPasswordPayload,
		}

		resp, err := c.httpTool.Run(req)
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			vulns = append(vulns, Vulnerability{
				ID:          fmt.Sprintf("AUTH-001-%s", sanitizeEndpoint(endpoint)),
				Title:       "Weak Password Policy",
				Severity:    "high",
				Category:    "authentication",
				Endpoint:    endpoint,
				Description: "Endpoint accepts weak passwords",
				Evidence:    "Very weak password '123' was accepted",
				Remediation: "Implement strong password policy (minimum length, complexity requirements)",
				OWASPRef:    "A07:2021",
				CWERef:      "CWE-521",
			})
		}
	}

	return vulns, checks
}

// A10:2021 - Server-Side Request Forgery
func (c *OWASPChecker) checkSSRF(method, url, endpoint string) ([]Vulnerability, int) {
	var vulns []Vulnerability
	checks := 0

	if method == "POST" || method == "PUT" {
		// Test for SSRF by trying to access internal URLs
		checks++
		ssrfPayload := map[string]interface{}{
			"url":      "http://localhost:8080/admin",
			"callback": "http://169.254.169.254/latest/meta-data",
		}

		req := shared.HTTPRequest{
			Method:  method,
			URL:     url,
			Headers: map[string]string{"Content-Type": "application/json"},
			Body:    ssrfPayload,
		}

		resp, err := c.httpTool.Run(req)
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			// Check if response contains evidence of internal resource access
			if strings.Contains(resp.Body, "localhost") || strings.Contains(resp.Body, "169.254") {
				vulns = append(vulns, Vulnerability{
					ID:          fmt.Sprintf("SSRF-001-%s", sanitizeEndpoint(endpoint)),
					Title:       "Potential Server-Side Request Forgery",
					Severity:    "critical",
					Category:    "ssrf",
					Endpoint:    endpoint,
					Description: "Endpoint may be vulnerable to SSRF attacks",
					Evidence:    "Internal URL access detected",
					Remediation: "Validate and sanitize all user-provided URLs, use allowlist of permitted domains",
					OWASPRef:    "A10:2021",
					CWERef:      "CWE-918",
				})
			}
		}
	}

	return vulns, checks
}

// sanitizeEndpoint removes special characters for ID generation.
func sanitizeEndpoint(endpoint string) string {
	s := strings.ReplaceAll(endpoint, "/", "_")
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, ":", "")
	s = strings.ReplaceAll(s, "{", "")
	s = strings.ReplaceAll(s, "}", "")
	return s
}
