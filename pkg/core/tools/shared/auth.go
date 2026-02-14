// Package shared provides merged authentication tools for the ZAP agent.
// This file combines Bearer token, Basic auth, OAuth2, and auth helper tools.
package shared

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// =============================================================================
// Bearer Token Authentication
// =============================================================================

// BearerTool creates Bearer token authorization headers for JWT, API tokens, etc.
// It wraps a token in the standard "Bearer <token>" format and optionally saves it to a variable.
type BearerTool struct {
	varStore *VariableStore
}

// NewBearerTool creates a new Bearer auth tool with the given variable store.
func NewBearerTool(varStore *VariableStore) *BearerTool {
	return &BearerTool{varStore: varStore}
}

// BearerParams defines the parameters for Bearer token authentication.
type BearerParams struct {
	// Token is the actual token value (can use {{VAR}} for variable substitution)
	Token string `json:"token"`
	// SaveAs is the optional variable name to save the Authorization header
	SaveAs string `json:"save_as,omitempty"`
}

// Name returns the tool name.
func (t *BearerTool) Name() string {
	return "auth_bearer"
}

// Description returns a human-readable description of the tool.
func (t *BearerTool) Description() string {
	return "Create Bearer token authorization header (for JWT tokens, API tokens). Saves 'Authorization: Bearer <token>' to a variable for use in requests."
}

// Parameters returns an example of the JSON parameters this tool accepts.
func (t *BearerTool) Parameters() string {
	return `{
  "token": "{{AUTH_TOKEN}}",
  "save_as": "auth_header"
}`
}

// Execute creates a Bearer authorization header from the provided token.
// If save_as is specified, the header is saved to a variable for later use.
func (t *BearerTool) Execute(args string) (string, error) {
	// Substitute variables in args
	if t.varStore != nil {
		args = t.varStore.Substitute(args)
	}

	var params BearerParams
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

// =============================================================================
// Basic Authentication
// =============================================================================

// BasicTool creates HTTP Basic authentication headers with base64 encoding.
// It encodes username:password and wraps it in the standard "Basic <encoded>" format.
type BasicTool struct {
	varStore *VariableStore
}

// NewBasicTool creates a new Basic auth tool with the given variable store.
func NewBasicTool(varStore *VariableStore) *BasicTool {
	return &BasicTool{varStore: varStore}
}

// BasicParams defines the parameters for HTTP Basic authentication.
type BasicParams struct {
	// Username for authentication
	Username string `json:"username"`
	// Password for authentication
	Password string `json:"password"`
	// SaveAs is the optional variable name to save the Authorization header
	SaveAs string `json:"save_as,omitempty"`
}

// Name returns the tool name.
func (t *BasicTool) Name() string {
	return "auth_basic"
}

// Description returns a human-readable description of the tool.
func (t *BasicTool) Description() string {
	return "Create HTTP Basic authentication header. Encodes username:password in base64 and saves 'Authorization: Basic <encoded>' to a variable."
}

// Parameters returns an example of the JSON parameters this tool accepts.
func (t *BasicTool) Parameters() string {
	return `{
  "username": "admin",
  "password": "secret123",
  "save_as": "auth_header"
}`
}

// Execute creates a Basic authentication header from the provided credentials.
// The credentials are base64-encoded in the format "username:password".
// If save_as is specified, the header is saved to a variable for later use.
func (t *BasicTool) Execute(args string) (string, error) {
	// Substitute variables in args
	if t.varStore != nil {
		args = t.varStore.Substitute(args)
	}

	var params BasicParams
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

// =============================================================================
// OAuth2 Authentication
// =============================================================================

// OAuth2Tool handles OAuth2 authentication flows.
// It supports client_credentials and password grant types, obtaining access tokens
// and automatically saving them as variables for use in subsequent requests.
type OAuth2Tool struct {
	varStore *VariableStore
}

// NewOAuth2Tool creates a new OAuth2 auth tool with the given variable store.
func NewOAuth2Tool(varStore *VariableStore) *OAuth2Tool {
	return &OAuth2Tool{varStore: varStore}
}

// OAuth2Params defines the parameters for OAuth2 authentication.
type OAuth2Params struct {
	// Flow specifies the OAuth2 grant type: "client_credentials", "password"
	Flow string `json:"flow"`
	// TokenURL is the OAuth2 token endpoint URL
	TokenURL string `json:"token_url"`
	// ClientID is the OAuth2 client identifier
	ClientID string `json:"client_id"`
	// ClientSecret is the OAuth2 client secret
	ClientSecret string `json:"client_secret"`
	// Scopes are the requested OAuth2 scopes (optional)
	Scopes []string `json:"scopes,omitempty"`
	// Username is required for password flow
	Username string `json:"username,omitempty"`
	// Password is required for password flow
	Password string `json:"password,omitempty"`
	// AuthURL is for authorization_code flow (not supported in CLI mode)
	AuthURL string `json:"auth_url,omitempty"`
	// RedirectURL is for authorization_code flow (not supported in CLI mode)
	RedirectURL string `json:"redirect_url,omitempty"`
	// Code is for authorization_code flow (not supported in CLI mode)
	Code string `json:"code,omitempty"`
	// SaveTokenAs is the variable name to save the access token
	SaveTokenAs string `json:"save_token_as,omitempty"`
}

// Name returns the tool name.
func (t *OAuth2Tool) Name() string {
	return "auth_oauth2"
}

// Description returns a human-readable description of the tool.
func (t *OAuth2Tool) Description() string {
	return "Perform OAuth2 authentication flows (client_credentials, password). Obtains access token and saves to variable."
}

// Parameters returns an example of the JSON parameters this tool accepts.
func (t *OAuth2Tool) Parameters() string {
	return `{
  "flow": "client_credentials",
  "token_url": "https://auth.example.com/token",
  "client_id": "{{CLIENT_ID}}",
  "client_secret": "{{CLIENT_SECRET}}",
  "scopes": ["api:read", "api:write"],
  "save_token_as": "oauth_token"
}`
}

// Execute performs the OAuth2 authentication flow.
// Supported flows:
//   - client_credentials: Server-to-server authentication using client ID and secret
//   - password: User authentication using username and password (Resource Owner Password Credentials)
func (t *OAuth2Tool) Execute(args string) (string, error) {
	// Substitute variables in args
	if t.varStore != nil {
		args = t.varStore.Substitute(args)
	}

	var params OAuth2Params
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	// Validate common parameters
	if params.TokenURL == "" {
		return "", fmt.Errorf("'token_url' parameter is required")
	}
	if params.ClientID == "" {
		return "", fmt.Errorf("'client_id' parameter is required")
	}
	if params.ClientSecret == "" {
		return "", fmt.Errorf("'client_secret' parameter is required")
	}

	// Execute flow based on type
	switch params.Flow {
	case "client_credentials":
		return t.clientCredentialsFlow(params)
	case "password":
		return t.passwordFlow(params)
	case "authorization_code":
		return "", fmt.Errorf("authorization_code flow requires manual browser interaction and is not supported in CLI mode. Use 'client_credentials' or 'password' flows instead")
	default:
		return "", fmt.Errorf("unknown flow '%s' (supported: client_credentials, password)", params.Flow)
	}
}

// clientCredentialsFlow performs OAuth2 client credentials flow.
// This flow is used for server-to-server authentication where the client authenticates
// using its own credentials (client_id and client_secret).
func (t *OAuth2Tool) clientCredentialsFlow(params OAuth2Params) (string, error) {
	config := clientcredentials.Config{
		ClientID:     params.ClientID,
		ClientSecret: params.ClientSecret,
		TokenURL:     params.TokenURL,
		Scopes:       params.Scopes,
	}

	ctx := context.Background()
	token, err := config.Token(ctx)
	if err != nil {
		return "", fmt.Errorf("OAuth2 client_credentials flow failed: %w", err)
	}

	return t.formatTokenResponse(token, params)
}

// passwordFlow performs OAuth2 password (Resource Owner Password Credentials) flow.
// This flow is used when the client has the user's credentials and exchanges them
// for an access token.
func (t *OAuth2Tool) passwordFlow(params OAuth2Params) (string, error) {
	if params.Username == "" {
		return "", fmt.Errorf("'username' parameter is required for password flow")
	}
	if params.Password == "" {
		return "", fmt.Errorf("'password' parameter is required for password flow")
	}

	config := oauth2.Config{
		ClientID:     params.ClientID,
		ClientSecret: params.ClientSecret,
		Endpoint: oauth2.Endpoint{
			TokenURL: params.TokenURL,
		},
		Scopes: params.Scopes,
	}

	ctx := context.Background()
	token, err := config.PasswordCredentialsToken(ctx, params.Username, params.Password)
	if err != nil {
		return "", fmt.Errorf("OAuth2 password flow failed: %w", err)
	}

	return t.formatTokenResponse(token, params)
}

// formatTokenResponse formats the OAuth2 token response and saves it to variables.
// If save_token_as is specified, both the raw token and a Bearer header are saved.
func (t *OAuth2Tool) formatTokenResponse(token *oauth2.Token, params OAuth2Params) (string, error) {
	var sb strings.Builder

	sb.WriteString("OAuth2 Authentication Successful!\n\n")
	sb.WriteString(fmt.Sprintf("Access Token: %s\n", token.AccessToken))
	sb.WriteString(fmt.Sprintf("Token Type: %s\n", token.TokenType))

	if token.RefreshToken != "" {
		sb.WriteString(fmt.Sprintf("Refresh Token: %s\n", token.RefreshToken))
	}

	if !token.Expiry.IsZero() {
		sb.WriteString(fmt.Sprintf("Expires: %s\n", token.Expiry.Format("2006-01-02 15:04:05")))
	}

	// Save token to variable if requested
	if params.SaveTokenAs != "" && t.varStore != nil {
		t.varStore.Set(params.SaveTokenAs, token.AccessToken)
		sb.WriteString(fmt.Sprintf("\nToken saved as: {{%s}}\n", params.SaveTokenAs))

		// Also save as Bearer header for convenience
		authHeaderVar := params.SaveTokenAs + "_header"
		bearerHeader := fmt.Sprintf("Bearer %s", token.AccessToken)
		t.varStore.Set(authHeaderVar, bearerHeader)
		sb.WriteString(fmt.Sprintf("Bearer header saved as: {{%s}}\n", authHeaderVar))

		sb.WriteString("\nUse in requests:\n")
		sb.WriteString("{\n")
		sb.WriteString(fmt.Sprintf("  \"headers\": {\"Authorization\": \"{{%s}}\"}\n", authHeaderVar))
		sb.WriteString("}\n")
	}

	return sb.String(), nil
}

// =============================================================================
// Auth Helper (JWT parsing, Basic auth decoding)
// =============================================================================

// HelperTool provides authentication utilities including JWT parsing and Basic auth decoding.
// It helps developers inspect and debug authentication tokens.
type HelperTool struct {
	responseManager *ResponseManager
	varStore        *VariableStore
}

// NewHelperTool creates a new auth helper tool.
func NewHelperTool(responseManager *ResponseManager, varStore *VariableStore) *HelperTool {
	return &HelperTool{
		responseManager: responseManager,
		varStore:        varStore,
	}
}

// HelperParams defines the parameters for auth helper operations.
type HelperParams struct {
	// Action specifies the operation: "parse_jwt", "decode_basic"
	Action string `json:"action"`
	// Token is the token to parse or decode
	Token string `json:"token,omitempty"`
	// FromBody extracts the token from a response body field (optional)
	FromBody string `json:"from_body,omitempty"`
}

// Name returns the tool name.
func (t *HelperTool) Name() string {
	return "auth_helper"
}

// Description returns a human-readable description of the tool.
func (t *HelperTool) Description() string {
	return "Auth utilities: parse JWT tokens, decode Basic auth, extract tokens from responses"
}

// Parameters returns an example of the JSON parameters this tool accepts.
func (t *HelperTool) Parameters() string {
	return `{
  "action": "parse_jwt",
  "token": "{{JWT_TOKEN}}"
}`
}

// Execute performs the requested auth helper action.
// Supported actions:
//   - parse_jwt: Decode and display JWT token claims (header, payload, signature)
//   - decode_basic: Decode Base64-encoded Basic auth credentials
func (t *HelperTool) Execute(args string) (string, error) {
	// Substitute variables
	if t.varStore != nil {
		args = t.varStore.Substitute(args)
	}

	var params HelperParams
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

// parseJWT decodes and displays JWT token claims.
// JWT tokens have 3 parts: header.payload.signature
// This function decodes the header and payload to show their contents.
func (t *HelperTool) parseJWT(token string) (string, error) {
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

// decodeBasic decodes Base64-encoded Basic auth credentials.
// The input should be in the format "Basic <base64>" or just the base64 string.
func (t *HelperTool) decodeBasic(authHeader string) (string, error) {
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

// =============================================================================
// Helper Functions
// =============================================================================

// base64DecodeJWTPart decodes a JWT part with URL-safe base64.
// JWT tokens use URL-safe base64 encoding without padding.
func base64DecodeJWTPart(part string) (string, error) {
	// First try RawURLEncoding (no padding, which is standard for JWT)
	decoded, err := base64.RawURLEncoding.DecodeString(part)
	if err == nil {
		return string(decoded), nil
	}

	// If that fails, try with padding added (some encoders add padding)
	switch len(part) % 4 {
	case 2:
		part += "=="
	case 3:
		part += "="
	}

	decoded, err = base64.URLEncoding.DecodeString(part)
	if err != nil {
		return "", fmt.Errorf("failed to decode JWT part: %w", err)
	}

	return string(decoded), nil
}

// formatJSON pretty-prints a JSON string.
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
