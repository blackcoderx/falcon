package core

import (
	"regexp"
	"strings"
)

// SecretPatterns contains regex patterns for detecting sensitive values
var SecretPatterns = []*regexp.Regexp{
	// API Keys and tokens
	regexp.MustCompile(`(?i)^(sk|pk|api|key|token|secret|password|passwd|pwd|auth|bearer|jwt|access|refresh)[-_]?[a-zA-Z0-9]{8,}`),
	regexp.MustCompile(`(?i)[a-zA-Z0-9]{32,}`), // Long random strings (likely tokens)

	// Specific provider patterns
	regexp.MustCompile(`sk-[a-zA-Z0-9]{20,}`),                           // OpenAI
	regexp.MustCompile(`(?i)^bearer\s+[a-zA-Z0-9_\-\.]+`),               // Bearer tokens
	regexp.MustCompile(`(?i)^basic\s+[a-zA-Z0-9+/=]+`),                  // Basic auth
	regexp.MustCompile(`ghp_[a-zA-Z0-9]{36}`),                           // GitHub PAT
	regexp.MustCompile(`gho_[a-zA-Z0-9]{36}`),                           // GitHub OAuth
	regexp.MustCompile(`github_pat_[a-zA-Z0-9_]{22,}`),                  // GitHub PAT (new)
	regexp.MustCompile(`xox[baprs]-[a-zA-Z0-9\-]+`),                     // Slack tokens
	regexp.MustCompile(`(?i)^ey[a-zA-Z0-9_\-]+\.[a-zA-Z0-9_\-]+\.`),     // JWT
	regexp.MustCompile(`AKIA[0-9A-Z]{16}`),                              // AWS Access Key
	regexp.MustCompile(`(?i)^[a-z0-9]{32}$`),                            // Generic 32-char hex
	regexp.MustCompile(`(?i)^[a-f0-9]{40}$`),                            // SHA-1 (40 hex chars)
	regexp.MustCompile(`(?i)^[a-f0-9]{64}$`),                            // SHA-256 (64 hex chars)
	regexp.MustCompile(`AIza[0-9A-Za-z_\-]{35}`),                        // Google API Key
	regexp.MustCompile(`(?i)^SG\.[a-zA-Z0-9_\-]+\.[a-zA-Z0-9_\-]+`),     // SendGrid API Key
	regexp.MustCompile(`(?i)^sk_live_[a-zA-Z0-9]{24,}`),                 // Stripe Live Key
	regexp.MustCompile(`(?i)^sk_test_[a-zA-Z0-9]{24,}`),                 // Stripe Test Key
	regexp.MustCompile(`(?i)^rk_live_[a-zA-Z0-9]{24,}`),                 // Stripe Restricted Key
	regexp.MustCompile(`(?i)^rk_test_[a-zA-Z0-9]{24,}`),                 // Stripe Restricted Test Key
	regexp.MustCompile(`sq0[a-z]{3}-[a-zA-Z0-9_\-]{22,}`),               // Square
	regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`), // UUID (sometimes used as API keys)
}

// SensitiveKeyPatterns contains patterns for keys that typically hold sensitive values
var SensitiveKeyPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(api[_-]?key|apikey)`),
	regexp.MustCompile(`(?i)(secret[_-]?key|secretkey)`),
	regexp.MustCompile(`(?i)(access[_-]?key|accesskey)`),
	regexp.MustCompile(`(?i)(auth[_-]?token|authtoken)`),
	regexp.MustCompile(`(?i)(bearer[_-]?token|bearertoken)`),
	regexp.MustCompile(`(?i)(password|passwd|pwd)`),
	regexp.MustCompile(`(?i)(private[_-]?key|privatekey)`),
	regexp.MustCompile(`(?i)(client[_-]?secret|clientsecret)`),
	regexp.MustCompile(`(?i)(jwt[_-]?token|jwttoken)`),
	regexp.MustCompile(`(?i)(refresh[_-]?token|refreshtoken)`),
	regexp.MustCompile(`(?i)(access[_-]?token|accesstoken)`),
	regexp.MustCompile(`(?i)^token$`),
	regexp.MustCompile(`(?i)^secret$`),
	regexp.MustCompile(`(?i)^credentials?$`),
	regexp.MustCompile(`(?i)authorization`),
}

// VariablePlaceholderPattern matches {{VAR}} placeholders
var VariablePlaceholderPattern = regexp.MustCompile(`\{\{[A-Za-z_][A-Za-z0-9_]*\}\}`)

// IsSecret checks if a key/value pair appears to be sensitive.
// Returns true if:
// - The key matches a sensitive key pattern, OR
// - The value matches a known secret pattern
func IsSecret(key, value string) bool {
	// Check if key indicates a sensitive value
	for _, pattern := range SensitiveKeyPatterns {
		if pattern.MatchString(key) {
			return true
		}
	}

	// Check if value looks like a secret
	return isSecretValue(value)
}

// isSecretValue checks if a value looks like a secret
func isSecretValue(value string) bool {
	// Skip empty or very short values
	if len(value) < 8 {
		return false
	}

	// Skip if it's just a placeholder
	if ContainsVariablePlaceholder(value) && !hasNonPlaceholderContent(value) {
		return false
	}

	// Check against secret patterns
	for _, pattern := range SecretPatterns {
		if pattern.MatchString(value) {
			return true
		}
	}

	return false
}

// hasNonPlaceholderContent checks if a string has content beyond just placeholders
func hasNonPlaceholderContent(value string) bool {
	stripped := VariablePlaceholderPattern.ReplaceAllString(value, "")
	stripped = strings.TrimSpace(stripped)
	// If after removing placeholders there's still significant content, it might have hardcoded secrets
	return len(stripped) > 10
}

// MaskSecret returns a masked version of a secret value.
// For values 12+ chars: shows first 4 and last 4 chars (e.g., "sk-12...cdef")
// For shorter values: shows "****"
func MaskSecret(value string) string {
	if len(value) <= 8 {
		return "****"
	}
	if len(value) < 12 {
		return value[:2] + "..." + value[len(value)-2:]
	}
	return value[:4] + "..." + value[len(value)-4:]
}

// HasPlaintextSecret checks if text contains hardcoded secrets without {{VAR}} placeholders.
// This is used to validate that saved requests use environment variables instead of hardcoded secrets.
func HasPlaintextSecret(text string) bool {
	// Skip empty text
	if text == "" {
		return false
	}

	// If the entire text is just a placeholder, it's fine
	if isOnlyPlaceholder(text) {
		return false
	}

	// Check for common secret patterns that aren't wrapped in placeholders
	// First, extract non-placeholder parts
	nonPlaceholderParts := extractNonPlaceholderParts(text)

	for _, part := range nonPlaceholderParts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Check if this part matches any secret pattern
		for _, pattern := range SecretPatterns {
			if pattern.MatchString(part) {
				return true
			}
		}
	}

	return false
}

// isOnlyPlaceholder checks if the text is just a placeholder (with optional prefix like "Bearer ")
func isOnlyPlaceholder(text string) bool {
	// Common patterns like "Bearer {{TOKEN}}" are OK
	text = strings.TrimSpace(text)

	// Remove common auth prefixes
	prefixes := []string{"Bearer ", "bearer ", "Basic ", "basic ", "Token ", "token "}
	for _, prefix := range prefixes {
		text = strings.TrimPrefix(text, prefix)
	}

	// Check if what remains is just placeholders
	stripped := VariablePlaceholderPattern.ReplaceAllString(text, "")
	stripped = strings.TrimSpace(stripped)
	return stripped == ""
}

// extractNonPlaceholderParts extracts parts of text that are not {{VAR}} placeholders
func extractNonPlaceholderParts(text string) []string {
	// Split by placeholders and return non-empty parts
	parts := VariablePlaceholderPattern.Split(text, -1)
	var result []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		// Skip common non-sensitive prefixes
		if part == "" || part == "Bearer" || part == "bearer" ||
			part == "Basic" || part == "basic" || part == "Token" || part == "token" {
			continue
		}
		result = append(result, part)
	}
	return result
}

// ContainsVariablePlaceholder checks if text contains {{VAR}} syntax
func ContainsVariablePlaceholder(text string) bool {
	return VariablePlaceholderPattern.MatchString(text)
}

// ValidateRequestForSecrets checks a request's URL, headers, and body for plaintext secrets.
// Returns an error message describing what was found, or empty string if clean.
func ValidateRequestForSecrets(url string, headers map[string]string, body interface{}) string {
	// Check URL
	if HasPlaintextSecret(url) {
		return "URL contains plaintext secret. Use {{VAR}} placeholder instead.\nExample: {{BASE_URL}}/api/users?key={{API_KEY}}"
	}

	// Check headers
	for key, value := range headers {
		if HasPlaintextSecret(value) {
			return "Header '" + key + "' contains plaintext secret. Use {{VAR}} instead.\nExample: Authorization: Bearer {{API_TOKEN}}"
		}
	}

	// Check body if it's a string
	if bodyStr, ok := body.(string); ok {
		if HasPlaintextSecret(bodyStr) {
			return "Request body contains plaintext secret. Use {{VAR}} placeholder instead."
		}
	}

	// Check body if it's a map
	if bodyMap, ok := body.(map[string]interface{}); ok {
		for key, val := range bodyMap {
			if strVal, ok := val.(string); ok {
				// Check if the key suggests a sensitive value
				for _, pattern := range SensitiveKeyPatterns {
					if pattern.MatchString(key) {
						if !ContainsVariablePlaceholder(strVal) && len(strVal) > 0 {
							return "Body field '" + key + "' appears to contain a secret. Use {{VAR}} placeholder instead.\nExample: \"" + key + "\": \"{{" + strings.ToUpper(key) + "}}\""
						}
					}
				}
				// Also check the value itself
				if HasPlaintextSecret(strVal) {
					return "Body field '" + key + "' contains plaintext secret. Use {{VAR}} placeholder instead."
				}
			}
		}
	}

	return ""
}
