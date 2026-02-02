# pkg/core/tools/auth

This subpackage contains authentication-related tools for ZAP. These tools help with creating auth headers, managing tokens, and performing OAuth2 flows.

## Package Overview

```
pkg/core/tools/auth/
├── bearer.go   # Bearer token auth (JWT, API tokens)
├── basic.go    # HTTP Basic authentication
├── oauth2.go   # OAuth2 flows (client_credentials, password)
└── helper.go   # JWT parsing, auth decoding utilities
```

## Tools

### auth_bearer

Creates Bearer token Authorization headers.

**Parameters:**

```json
{
  "token": "your-jwt-or-api-token"
}
```

**Usage:**

```
> Create a Bearer token header with my JWT
```

**Output:**

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

### auth_basic

Creates HTTP Basic Authentication headers (base64 encoded).

**Parameters:**

```json
{
  "username": "user",
  "password": "pass"
}
```

**Usage:**

```
> Create Basic auth for user admin with password secret123
```

**Output:**

```
Authorization: Basic YWRtaW46c2VjcmV0MTIz
```

### auth_oauth2

Performs OAuth2 authentication flows.

**Supported Flows:**

| Flow | Description |
|------|-------------|
| `client_credentials` | Server-to-server auth (no user interaction) |
| `password` | Resource Owner Password Credentials (username/password) |

**Parameters:**

```json
{
  "flow": "client_credentials",
  "token_url": "https://auth.example.com/oauth/token",
  "client_id": "your-client-id",
  "client_secret": "your-client-secret",
  "scope": "read write"
}
```

For password flow, also include:

```json
{
  "flow": "password",
  "username": "user@example.com",
  "password": "userpassword"
}
```

**Usage:**

```
> Get an OAuth2 token using client credentials from https://auth.example.com/token
```

**Output:**

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "scope": "read write"
}
```

### auth_helper

Parses and decodes authentication tokens.

**Operations:**

| Operation | Description |
|-----------|-------------|
| `parse_jwt` | Decode JWT token, show claims, expiration |
| `decode_basic` | Decode Basic auth header to username:password |

**Parameters:**

```json
{
  "operation": "parse_jwt",
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Usage:**

```
> Parse this JWT token and show me the claims
```

**Output (parse_jwt):**

```json
{
  "header": {
    "alg": "HS256",
    "typ": "JWT"
  },
  "payload": {
    "sub": "1234567890",
    "name": "John Doe",
    "iat": 1516239022,
    "exp": 1516242622
  },
  "expired": false,
  "expires_at": "2024-01-18T12:30:22Z"
}
```

## Implementation Details

### Bearer Token (bearer.go)

Simple header construction:

```go
func (t *BearerTool) Execute(args string) (string, error) {
    var params struct {
        Token string `json:"token"`
    }
    json.Unmarshal([]byte(args), &params)

    return fmt.Sprintf("Authorization: Bearer %s", params.Token), nil
}
```

### Basic Auth (basic.go)

Base64 encoding of `username:password`:

```go
func (t *BasicTool) Execute(args string) (string, error) {
    var params struct {
        Username string `json:"username"`
        Password string `json:"password"`
    }
    json.Unmarshal([]byte(args), &params)

    credentials := fmt.Sprintf("%s:%s", params.Username, params.Password)
    encoded := base64.StdEncoding.EncodeToString([]byte(credentials))

    return fmt.Sprintf("Authorization: Basic %s", encoded), nil
}
```

### OAuth2 (oauth2.go)

Uses `golang.org/x/oauth2` package:

```go
func (t *OAuth2Tool) clientCredentialsFlow(params OAuth2Params) (string, error) {
    config := &clientcredentials.Config{
        ClientID:     params.ClientID,
        ClientSecret: params.ClientSecret,
        TokenURL:     params.TokenURL,
        Scopes:       strings.Split(params.Scope, " "),
    }

    token, err := config.Token(context.Background())
    if err != nil {
        return "", err
    }

    return formatTokenResponse(token), nil
}
```

### JWT Parsing (helper.go)

Decodes JWT without verification (for inspection):

```go
func (t *HelperTool) parseJWT(token string) (string, error) {
    parts := strings.Split(token, ".")
    if len(parts) != 3 {
        return "", errors.New("invalid JWT format")
    }

    header, _ := base64.RawURLEncoding.DecodeString(parts[0])
    payload, _ := base64.RawURLEncoding.DecodeString(parts[1])

    // Parse expiration
    var claims map[string]interface{}
    json.Unmarshal(payload, &claims)

    expired := false
    if exp, ok := claims["exp"].(float64); ok {
        expired = time.Unix(int64(exp), 0).Before(time.Now())
    }

    return formatJWTInfo(header, payload, expired), nil
}
```

## Usage Patterns

### Chaining Auth with Requests

```
> Get an OAuth2 token from https://api.example.com/oauth/token
> Store the access_token as API_TOKEN
> GET https://api.example.com/users with Bearer {{API_TOKEN}}
```

### Debugging Token Issues

```
> Parse this JWT: eyJhbGciOiJIUzI1NiIs...
> Check if the token is expired
> Show me the claims
```

### Environment-Based Auth

Store tokens in environments:

```yaml
# .zap/environments/dev.yaml
API_TOKEN: dev-token-abc123

# .zap/environments/prod.yaml
API_TOKEN: prod-token-xyz789
```

Then use:

```
> switch to prod environment
> GET /api/users with Bearer {{API_TOKEN}}
```

## Security Notes

1. **Tokens are not stored** - Auth tools return headers, they don't persist tokens
2. **Use variables for persistence** - Store tokens in session/global variables if needed
3. **Environment isolation** - Different tokens per environment
4. **JWT parsing is unsigned** - `parse_jwt` decodes without verification (for debugging)

## Adding New Auth Methods

To add a new authentication method:

1. Create a new file (e.g., `apikey.go`)
2. Implement the `Tool` interface
3. Register in `pkg/tui/init.go`

Example for API Key auth:

```go
type APIKeyTool struct{}

func NewAPIKeyTool() *APIKeyTool {
    return &APIKeyTool{}
}

func (t *APIKeyTool) Name() string {
    return "auth_apikey"
}

func (t *APIKeyTool) Description() string {
    return "Create API key header (X-API-Key or custom header)"
}

func (t *APIKeyTool) Parameters() string {
    return `{
        "type": "object",
        "properties": {
            "key": {"type": "string", "description": "The API key"},
            "header_name": {"type": "string", "default": "X-API-Key"}
        },
        "required": ["key"]
    }`
}

func (t *APIKeyTool) Execute(args string) (string, error) {
    var params struct {
        Key        string `json:"key"`
        HeaderName string `json:"header_name"`
    }
    json.Unmarshal([]byte(args), &params)

    if params.HeaderName == "" {
        params.HeaderName = "X-API-Key"
    }

    return fmt.Sprintf("%s: %s", params.HeaderName, params.Key), nil
}
```
