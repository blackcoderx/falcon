package tools

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// AuthBearerTool creates Bearer token authorization headers (for JWT, API tokens, etc.)
type AuthBearerTool struct {
	varStore *VariableStore
}

// NewAuthBearerTool creates a new Bearer auth tool
func NewAuthBearerTool(varStore *VariableStore) *AuthBearerTool {
	return &AuthBearerTool{varStore: varStore}
}

// AuthBearerParams defines Bearer auth parameters
type AuthBearerParams struct {
	Token  string `json:"token"`             // Token value (can use {{VAR}})
	SaveAs string `json:"save_as,omitempty"` // Variable name to save header
}

// Name returns the tool name
func (t *AuthBearerTool) Name() string {
	return "auth_bearer"
}

// Description returns the tool description
func (t *AuthBearerTool) Description() string {
	return "Create Bearer token authorization header (for JWT tokens, API tokens). Saves 'Authorization: Bearer <token>' to a variable for use in requests."
}

// Parameters returns the tool parameter description
func (t *AuthBearerTool) Parameters() string {
	return `{
  "token": "{{AUTH_TOKEN}}",
  "save_as": "auth_header"
}`
}

// Execute creates a Bearer authorization header
func (t *AuthBearerTool) Execute(args string) (string, error) {
	// Substitute variables in args
	if t.varStore != nil {
		args = t.varStore.Substitute(args)
	}

	var params AuthBearerParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	if params.Token == "" {
		return "", fmt.Errorf("'token' parameter is required")
	}

	// Create Bearer header
	authHeader := fmt.Sprintf("Bearer %s", params.Token)

	// Save to variable if requested
	if params.SaveAs != "" {
		t.varStore.Set(params.SaveAs, authHeader)
		return fmt.Sprintf("Created Bearer token authorization header.\nSaved as: {{%s}}\n\nUse in requests:\n{\n  \"headers\": {\"Authorization\": \"{{%s}}\"}\n}",
			params.SaveAs, params.SaveAs), nil
	}

	return fmt.Sprintf("Bearer token: %s\n\nUse in requests:\n{\n  \"headers\": {\"Authorization\": \"%s\"}\n}",
		authHeader, authHeader), nil
}

// AuthBasicTool creates HTTP Basic authentication headers
type AuthBasicTool struct {
	varStore *VariableStore
}

// NewAuthBasicTool creates a new Basic auth tool
func NewAuthBasicTool(varStore *VariableStore) *AuthBasicTool {
	return &AuthBasicTool{varStore: varStore}
}

// AuthBasicParams defines Basic auth parameters
type AuthBasicParams struct {
	Username string `json:"username"`          // Username
	Password string `json:"password"`          // Password
	SaveAs   string `json:"save_as,omitempty"` // Variable name to save header
}

// Name returns the tool name
func (t *AuthBasicTool) Name() string {
	return "auth_basic"
}

// Description returns the tool description
func (t *AuthBasicTool) Description() string {
	return "Create HTTP Basic authentication header. Encodes username:password in base64 and saves 'Authorization: Basic <encoded>' to a variable."
}

// Parameters returns the tool parameter description
func (t *AuthBasicTool) Parameters() string {
	return `{
  "username": "admin",
  "password": "secret123",
  "save_as": "auth_header"
}`
}

// Execute creates a Basic authentication header
func (t *AuthBasicTool) Execute(args string) (string, error) {
	// Substitute variables in args
	if t.varStore != nil {
		args = t.varStore.Substitute(args)
	}

	var params AuthBasicParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	if params.Username == "" {
		return "", fmt.Errorf("'username' parameter is required")
	}

	if params.Password == "" {
		return "", fmt.Errorf("'password' parameter is required")
	}

	// Encode credentials as base64
	credentials := fmt.Sprintf("%s:%s", params.Username, params.Password)
	encoded := base64.StdEncoding.EncodeToString([]byte(credentials))
	authHeader := fmt.Sprintf("Basic %s", encoded)

	// Save to variable if requested
	if params.SaveAs != "" {
		t.varStore.Set(params.SaveAs, authHeader)
		return fmt.Sprintf("Created HTTP Basic authentication header.\nUsername: %s\nSaved as: {{%s}}\n\nUse in requests:\n{\n  \"headers\": {\"Authorization\": \"{{%s}}\"}\n}",
			params.Username, params.SaveAs, params.SaveAs), nil
	}

	return fmt.Sprintf("Basic auth header: %s\n\nUse in requests:\n{\n  \"headers\": {\"Authorization\": \"%s\"}\n}",
		authHeader, authHeader), nil
}

// AuthHelperTool provides general auth utilities and token parsing
type AuthHelperTool struct {
	responseManager *ResponseManager
	varStore        *VariableStore
}

// NewAuthHelperTool creates a new auth helper tool
func NewAuthHelperTool(responseManager *ResponseManager, varStore *VariableStore) *AuthHelperTool {
	return &AuthHelperTool{
		responseManager: responseManager,
		varStore:        varStore,
	}
}

// AuthHelperParams defines auth helper parameters
type AuthHelperParams struct {
	Action   string `json:"action"`             // "parse_jwt", "decode_basic"
	Token    string `json:"token,omitempty"`    // Token to parse
	FromBody string `json:"from_body,omitempty"` // Extract from response body field
}

// Name returns the tool name
func (t *AuthHelperTool) Name() string {
	return "auth_helper"
}

// Description returns the tool description
func (t *AuthHelperTool) Description() string {
	return "Auth utilities: parse JWT tokens, decode Basic auth, extract tokens from responses"
}

// Parameters returns the tool parameter description
func (t *AuthHelperTool) Parameters() string {
	return `{
  "action": "parse_jwt",
  "token": "{{JWT_TOKEN}}"
}`
}

// Execute performs auth helper actions
func (t *AuthHelperTool) Execute(args string) (string, error) {
	// Substitute variables
	if t.varStore != nil {
		args = t.varStore.Substitute(args)
	}

	var params AuthHelperParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	switch params.Action {
	case "parse_jwt":
		return t.parseJWT(params.Token)
	case "decode_basic":
		return t.decodeBasic(params.Token)
	default:
		return "", fmt.Errorf("unknown action '%s' (use: parse_jwt, decode_basic)", params.Action)
	}
}

// parseJWT decodes and displays JWT token claims
func (t *AuthHelperTool) parseJWT(token string) (string, error) {
	if token == "" {
		return "", fmt.Errorf("'token' parameter is required")
	}

	// Remove "Bearer " prefix if present
	token = strings.TrimPrefix(token, "Bearer ")

	// JWT has 3 parts: header.payload.signature
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid JWT format (expected 3 parts, got %d)", len(parts))
	}

	var sb strings.Builder
	sb.WriteString("JWT Token Analysis:\n\n")

	// Decode header
	headerJSON, err := base64DecodeJWTPart(parts[0])
	if err != nil {
		sb.WriteString(fmt.Sprintf("Header: (decode error: %v)\n", err))
	} else {
		sb.WriteString("Header:\n")
		sb.WriteString(formatJSON(headerJSON))
		sb.WriteString("\n\n")
	}

	// Decode payload (claims)
	payloadJSON, err := base64DecodeJWTPart(parts[1])
	if err != nil {
		sb.WriteString(fmt.Sprintf("Payload: (decode error: %v)\n", err))
	} else {
		sb.WriteString("Payload (Claims):\n")
		sb.WriteString(formatJSON(payloadJSON))
		sb.WriteString("\n\n")

		// Parse common claims
		var claims map[string]interface{}
		if err := json.Unmarshal([]byte(payloadJSON), &claims); err == nil {
			if exp, ok := claims["exp"].(float64); ok {
				sb.WriteString(fmt.Sprintf("Expires: %v (Unix timestamp)\n", exp))
			}
			if iat, ok := claims["iat"].(float64); ok {
				sb.WriteString(fmt.Sprintf("Issued At: %v (Unix timestamp)\n", iat))
			}
			if sub, ok := claims["sub"].(string); ok {
				sb.WriteString(fmt.Sprintf("Subject: %s\n", sub))
			}
		}
	}

	sb.WriteString("\nSignature: " + parts[2] + " (not verified)\n")
	sb.WriteString("\nNote: Signature verification requires the secret key and is not performed by this tool.")

	return sb.String(), nil
}

// decodeBasic decodes Basic auth credentials
func (t *AuthHelperTool) decodeBasic(authHeader string) (string, error) {
	if authHeader == "" {
		return "", fmt.Errorf("'token' parameter is required")
	}

	// Remove "Basic " prefix if present
	encoded := strings.TrimPrefix(authHeader, "Basic ")

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("failed to decode Basic auth: %w", err)
	}

	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid Basic auth format (expected username:password)")
	}

	return fmt.Sprintf("Basic Auth Decoded:\nUsername: %s\nPassword: %s", parts[0], parts[1]), nil
}

// base64DecodeJWTPart decodes a JWT part with URL-safe base64
func base64DecodeJWTPart(part string) (string, error) {
	// JWT uses URL-safe base64 without padding
	// Add padding if needed
	switch len(part) % 4 {
	case 2:
		part += "=="
	case 3:
		part += "="
	}

	decoded, err := base64.RawURLEncoding.DecodeString(part)
	if err != nil {
		return "", err
	}

	return string(decoded), nil
}

// formatJSON pretty-prints JSON
func formatJSON(jsonStr string) string {
	var obj interface{}
	if err := json.Unmarshal([]byte(jsonStr), &obj); err != nil {
		return jsonStr
	}

	pretty, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return jsonStr
	}

	return string(pretty)
}
