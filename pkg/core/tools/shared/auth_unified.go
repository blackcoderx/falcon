package shared

import (
	"encoding/json"
	"fmt"
)

// AuthTool is the unified authentication tool that replaces auth_bearer, auth_basic,
// auth_oauth2, and auth_helper. Use the "action" field to select the auth method.
type AuthTool struct {
	bearer  *BearerTool
	basic   *BasicTool
	oauth2  *OAuth2Tool
	helper  *HelperTool
}

// NewAuthTool creates the unified auth tool backed by the existing auth implementations.
func NewAuthTool(responseManager *ResponseManager, varStore *VariableStore) *AuthTool {
	return &AuthTool{
		bearer: NewBearerTool(varStore),
		basic:  NewBasicTool(varStore),
		oauth2: NewOAuth2Tool(varStore),
		helper: NewHelperTool(responseManager, varStore),
	}
}

// AuthParams dispatches to the correct auth implementation based on action.
type AuthParams struct {
	// Action is required: "bearer", "basic", "oauth2", "parse_jwt", "decode_basic"
	Action string `json:"action"`

	// For "bearer": token, save_as
	Token  string `json:"token,omitempty"`
	SaveAs string `json:"save_as,omitempty"`

	// For "basic": username, password, save_as
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`

	// For "oauth2": flow, token_url, client_id, client_secret, scopes, save_token_as
	Flow         string   `json:"flow,omitempty"`
	TokenURL     string   `json:"token_url,omitempty"`
	ClientID     string   `json:"client_id,omitempty"`
	ClientSecret string   `json:"client_secret,omitempty"`
	Scopes       []string `json:"scopes,omitempty"`
	SaveTokenAs  string   `json:"save_token_as,omitempty"`

	// For "parse_jwt" / "decode_basic": token
	FromBody string `json:"from_body,omitempty"`
}

func (t *AuthTool) Name() string { return "auth" }

func (t *AuthTool) Description() string {
	return "Unified authentication tool. Actions: bearer (set Bearer token header), basic (HTTP Basic auth), oauth2 (client_credentials or password flow), parse_jwt (decode JWT claims), decode_basic (decode Basic credentials)"
}

func (t *AuthTool) Parameters() string {
	return `{
  "action": "bearer|basic|oauth2|parse_jwt|decode_basic",
  "token": "...",                   // bearer / parse_jwt / decode_basic
  "save_as": "auth_header",         // bearer / basic
  "username": "...",                // basic
  "password": "...",                // basic
  "flow": "client_credentials",     // oauth2
  "token_url": "https://...",       // oauth2
  "client_id": "...",               // oauth2
  "client_secret": "...",           // oauth2
  "scopes": ["api:read"],           // oauth2
  "save_token_as": "oauth_token"    // oauth2
}`
}

func (t *AuthTool) Execute(args string) (string, error) {
	var params AuthParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse auth parameters: %w", err)
	}

	switch params.Action {
	case "bearer":
		return t.bearer.Execute(args)

	case "basic":
		return t.basic.Execute(args)

	case "oauth2":
		return t.oauth2.Execute(args)

	case "parse_jwt":
		helperArgs, _ := json.Marshal(map[string]string{
			"action": "parse_jwt",
			"token":  params.Token,
		})
		return t.helper.Execute(string(helperArgs))

	case "decode_basic":
		helperArgs, _ := json.Marshal(map[string]string{
			"action": "decode_basic",
			"token":  params.Token,
		})
		return t.helper.Execute(string(helperArgs))

	default:
		return "", fmt.Errorf("unknown auth action '%s' (use: bearer, basic, oauth2, parse_jwt, decode_basic)", params.Action)
	}
}
