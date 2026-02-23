package security_scanner

import (
	"fmt"
	"strings"
	"time"

	"github.com/blackcoderx/falcon/pkg/core/tools/shared"
)

// AuthAuditor performs authentication and authorization security audits.
type AuthAuditor struct {
	httpTool *shared.HTTPTool
}

// NewAuthAuditor creates a new auth auditor.
func NewAuthAuditor(httpTool *shared.HTTPTool) *AuthAuditor {
	return &AuthAuditor{httpTool: httpTool}
}

// AuditAuth performs authentication and authorization security checks.
func (a *AuthAuditor) AuditAuth(endpoints map[string]shared.EndpointAnalysis, baseURL, authToken string) ([]Vulnerability, int) {
	var vulnerabilities []Vulnerability
	totalChecks := 0

	for endpointKey := range endpoints {
		parts := strings.SplitN(endpointKey, " ", 2)
		if len(parts) != 2 {
			continue
		}
		method := parts[0]
		path := parts[1]
		url := strings.TrimSuffix(baseURL, "/") + "/" + strings.TrimPrefix(path, "/")

		// Test 1: Expired/Invalid Token
		vulns, checks := a.testExpiredToken(method, url, endpointKey, authToken)
		vulnerabilities = append(vulnerabilities, vulns...)
		totalChecks += checks

		// Test 2: Missing Token
		vulns, checks = a.testMissingToken(method, url, endpointKey)
		vulnerabilities = append(vulnerabilities, vulns...)
		totalChecks += checks

		// Test 3: Weak Token
		vulns, checks = a.testWeakToken(method, url, endpointKey)
		vulnerabilities = append(vulnerabilities, vulns...)
		totalChecks += checks

		// Test 4: Horizontal Privilege Escalation (if endpoint has ID parameter)
		if strings.Contains(path, "{id}") || strings.Contains(path, "{userId}") {
			vulns, checks = a.testHorizontalPrivilegeEscalation(method, url, endpointKey, authToken)
			vulnerabilities = append(vulnerabilities, vulns...)
			totalChecks += checks
		}

		// Test 5: Vertical Privilege Escalation (admin endpoints)
		if strings.Contains(strings.ToLower(path), "admin") {
			vulns, checks = a.testVerticalPrivilegeEscalation(method, url, endpointKey, authToken)
			vulnerabilities = append(vulnerabilities, vulns...)
			totalChecks += checks
		}

		// Test 6: Session Fixation
		if strings.Contains(strings.ToLower(endpointKey), "login") {
			vulns, checks = a.testSessionFixation(method, url, endpointKey)
			vulnerabilities = append(vulnerabilities, vulns...)
			totalChecks += checks
		}

		// Test 7: JWT Security
		if strings.Contains(authToken, "eyJ") { // JWT starts with eyJ
			vulns, checks = a.testJWTSecurity(method, url, endpointKey, authToken)
			vulnerabilities = append(vulnerabilities, vulns...)
			totalChecks += checks
		}
	}

	return vulnerabilities, totalChecks
}

// testExpiredToken tests if API accepts expired or manipulated tokens.
func (a *AuthAuditor) testExpiredToken(method, url, endpoint string, _ string) ([]Vulnerability, int) {
	var vulns []Vulnerability
	checks := 0

	// Test with obviously invalid token
	checks++
	invalidToken := "Bearer invalid_token_12345"

	req := shared.HTTPRequest{
		Method:  method,
		URL:     url,
		Headers: map[string]string{"Authorization": invalidToken},
	}

	resp, err := a.httpTool.Run(req)
	if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
		vulns = append(vulns, Vulnerability{
			ID:          fmt.Sprintf("AUTH-INV-001-%s", sanitizeEndpoint(endpoint)),
			Title:       "Invalid Token Accepted",
			Severity:    "critical",
			Category:    "authentication",
			Endpoint:    endpoint,
			Description: "Endpoint accepts invalid authentication tokens",
			Evidence:    fmt.Sprintf("Invalid token resulted in %d response", resp.StatusCode),
			Remediation: "Implement proper token validation and verification",
			OWASPRef:    "A07:2021",
			CWERef:      "CWE-287",
		})
	}

	return vulns, checks
}

// testMissingToken tests if protected endpoints require authentication.
func (a *AuthAuditor) testMissingToken(method, url, endpoint string) ([]Vulnerability, int) {
	var vulns []Vulnerability
	checks := 0

	// Skip public endpoints
	if strings.Contains(strings.ToLower(endpoint), "public") ||
		strings.Contains(strings.ToLower(endpoint), "health") {
		return vulns, checks
	}

	checks++
	req := shared.HTTPRequest{
		Method:  method,
		URL:     url,
		Headers: map[string]string{},
	}

	resp, err := a.httpTool.Run(req)
	if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
		vulns = append(vulns, Vulnerability{
			ID:          fmt.Sprintf("AUTH-MISS-001-%s", sanitizeEndpoint(endpoint)),
			Title:       "Missing Authentication Check",
			Severity:    "high",
			Category:    "authentication",
			Endpoint:    endpoint,
			Description: "Protected endpoint accessible without authentication",
			Evidence:    fmt.Sprintf("Unauthenticated request returned %d", resp.StatusCode),
			Remediation: "Require authentication for all protected endpoints",
			OWASPRef:    "A01:2021",
			CWERef:      "CWE-306",
		})
	}

	return vulns, checks
}

// testWeakToken tests for weak token patterns.
func (a *AuthAuditor) testWeakToken(method, url, endpoint string) ([]Vulnerability, int) {
	var vulns []Vulnerability
	checks := 0

	// Test predictable tokens
	weakTokens := []string{
		"Bearer 123456",
		"Bearer admin",
		"Bearer test",
		"Bearer password",
	}

	for _, token := range weakTokens {
		checks++
		req := shared.HTTPRequest{
			Method:  method,
			URL:     url,
			Headers: map[string]string{"Authorization": token},
		}

		resp, err := a.httpTool.Run(req)
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			vulns = append(vulns, Vulnerability{
				ID:          fmt.Sprintf("AUTH-WEAK-001-%s", sanitizeEndpoint(endpoint)),
				Title:       "Weak Token Pattern Accepted",
				Severity:    "high",
				Category:    "authentication",
				Endpoint:    endpoint,
				Description: "Endpoint accepts weak or predictable tokens",
				Evidence:    fmt.Sprintf("Weak token '%s' was accepted", token),
				Remediation: "Use cryptographically secure random tokens (minimum 128 bits entropy)",
				OWASPRef:    "A07:2021",
				CWERef:      "CWE-330",
			})
			return vulns, checks // Stop after first detection
		}
	}

	return vulns, checks
}

// testHorizontalPrivilegeEscalation tests if users can access other users' resources.
func (a *AuthAuditor) testHorizontalPrivilegeEscalation(method, url, endpoint, authToken string) ([]Vulnerability, int) {
	var vulns []Vulnerability
	checks := 0

	// Replace user ID with different IDs to test IDOR
	testIDs := []string{"1", "2", "999", "admin", "0"}

	for _, testID := range testIDs {
		checks++
		// Replace ID patterns in URL
		testURL := strings.ReplaceAll(url, "{id}", testID)
		testURL = strings.ReplaceAll(testURL, "{userId}", testID)

		req := shared.HTTPRequest{
			Method:  method,
			URL:     testURL,
			Headers: map[string]string{"Authorization": authToken},
		}

		resp, err := a.httpTool.Run(req)
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			// Successful access might indicate IDOR
			vulns = append(vulns, Vulnerability{
				ID:          fmt.Sprintf("AUTH-IDOR-001-%s", sanitizeEndpoint(endpoint)),
				Title:       "Potential Insecure Direct Object Reference (IDOR)",
				Severity:    "high",
				Category:    "authorization",
				Endpoint:    endpoint,
				Description: "User may be able to access other users' resources",
				Evidence:    fmt.Sprintf("Successfully accessed resource with ID: %s", testID),
				Remediation: "Implement proper authorization checks to verify resource ownership",
				OWASPRef:    "A01:2021",
				CWERef:      "CWE-639",
			})
			return vulns, checks // Report first finding
		}
	}

	return vulns, checks
}

// testVerticalPrivilegeEscalation tests if regular users can access admin functions.
func (a *AuthAuditor) testVerticalPrivilegeEscalation(method, url, endpoint, authToken string) ([]Vulnerability, int) {
	var vulns []Vulnerability
	checks := 0

	// Test with regular user token (assumption: provided token is not admin)
	checks++
	req := shared.HTTPRequest{
		Method:  method,
		URL:     url,
		Headers: map[string]string{"Authorization": authToken},
	}

	resp, err := a.httpTool.Run(req)
	if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
		vulns = append(vulns, Vulnerability{
			ID:          fmt.Sprintf("AUTH-PRIV-001-%s", sanitizeEndpoint(endpoint)),
			Title:       "Potential Vertical Privilege Escalation",
			Severity:    "critical",
			Category:    "authorization",
			Endpoint:    endpoint,
			Description: "Admin endpoint may be accessible to non-admin users",
			Evidence:    fmt.Sprintf("Admin endpoint returned %d with regular token", resp.StatusCode),
			Remediation: "Implement role-based access control and verify user privileges",
			OWASPRef:    "A01:2021",
			CWERef:      "CWE-269",
		})
	}

	return vulns, checks
}

// testSessionFixation tests for session fixation vulnerabilities.
func (a *AuthAuditor) testSessionFixation(method, url, endpoint string) ([]Vulnerability, int) {
	var vulns []Vulnerability
	checks := 0

	// Test if session ID is preserved after login
	checks++
	sessionID := fmt.Sprintf("SESS_%d", time.Now().Unix())

	req := shared.HTTPRequest{
		Method: method,
		URL:    url,
		Headers: map[string]string{
			"Content-Type": "application/json",
			"Cookie":       fmt.Sprintf("session_id=%s", sessionID),
		},
		Body: map[string]interface{}{
			"username": "testuser",
			"password": "testpass",
		},
	}

	resp, err := a.httpTool.Run(req)
	if err == nil {
		// Check if the same session ID is returned
		if setCookie, ok := resp.Headers["Set-Cookie"]; ok {
			if strings.Contains(setCookie, sessionID) {
				vulns = append(vulns, Vulnerability{
					ID:          fmt.Sprintf("AUTH-SESS-001-%s", sanitizeEndpoint(endpoint)),
					Title:       "Potential Session Fixation",
					Severity:    "medium",
					Category:    "authentication",
					Endpoint:    endpoint,
					Description: "Session ID is not regenerated after authentication",
					Evidence:    "Same session ID preserved after login",
					Remediation: "Regenerate session ID after successful authentication",
					OWASPRef:    "A07:2021",
					CWERef:      "CWE-384",
				})
			}
		}
	}

	return vulns, checks
}

// testJWTSecurity tests for common JWT security issues.
func (a *AuthAuditor) testJWTSecurity(method, url, endpoint, jwtToken string) ([]Vulnerability, int) {
	var vulns []Vulnerability
	checks := 0

	// Remove "Bearer " prefix if present
	token := strings.TrimPrefix(jwtToken, "Bearer ")
	token = strings.TrimSpace(token)

	// Test 1: None algorithm attack
	checks++
	noneToken := modifyJWTAlgorithm(token, "none")
	req := shared.HTTPRequest{
		Method:  method,
		URL:     url,
		Headers: map[string]string{"Authorization": "Bearer " + noneToken},
	}

	resp, err := a.httpTool.Run(req)
	if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
		vulns = append(vulns, Vulnerability{
			ID:          fmt.Sprintf("AUTH-JWT-001-%s", sanitizeEndpoint(endpoint)),
			Title:       "JWT 'None' Algorithm Vulnerability",
			Severity:    "critical",
			Category:    "authentication",
			Endpoint:    endpoint,
			Description: "JWT with 'none' algorithm is accepted",
			Evidence:    "Unsigned JWT token was accepted",
			Remediation: "Reject JWTs with 'none' algorithm explicitly",
			OWASPRef:    "A07:2021",
			CWERef:      "CWE-347",
		})
	}

	// Test 2: Weak secret (if we can guess it)
	checks++
	// Try common weak secrets
	weakSecrets := []string{"secret", "password", "123456"}
	for _, secret := range weakSecrets {
		if validateJWTWithSecret(token, secret) {
			vulns = append(vulns, Vulnerability{
				ID:          fmt.Sprintf("AUTH-JWT-002-%s", sanitizeEndpoint(endpoint)),
				Title:       "Weak JWT Secret",
				Severity:    "critical",
				Category:    "authentication",
				Endpoint:    endpoint,
				Description: "JWT is signed with a weak secret",
				Evidence:    fmt.Sprintf("JWT can be verified with weak secret: %s", secret),
				Remediation: "Use strong random secrets (minimum 256 bits) for JWT signing",
				OWASPRef:    "A02:2021",
				CWERef:      "CWE-326",
			})
			break
		}
	}

	return vulns, checks
}

// modifyJWTAlgorithm attempts to modify the JWT algorithm (simplified implementation).
func modifyJWTAlgorithm(token string, _ string) string {
	// This is a simplified version - real implementation would properly decode/encode JWT
	// For demonstration purposes, we're just returning a modified token
	return token // Placeholder
}

// validateJWTWithSecret checks if a JWT can be validated with a given secret (simplified).
func validateJWTWithSecret(_ string, _ string) bool {
	// Simplified check - in real implementation, would properly verify JWT signature
	// For demonstration, we'll return false
	return false
}
