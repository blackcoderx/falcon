package shared

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// ExtractTool extracts values from HTTP responses for use in subsequent requests
type ExtractTool struct {
	responseManager *ResponseManager
	variables       *VariableStore // Shared with VariableTool
}

// NewExtractTool creates a new extraction tool
func NewExtractTool(responseManager *ResponseManager, varStore *VariableStore) *ExtractTool {
	return &ExtractTool{
		responseManager: responseManager,
		variables:       varStore,
	}
}

// ExtractParams defines what to extract and where to save it
type ExtractParams struct {
	JSONPath   string `json:"json_path,omitempty"`   // e.g., "$.data.user.id"
	Header     string `json:"header,omitempty"`      // e.g., "X-Request-Id"
	Cookie     string `json:"cookie,omitempty"`      // e.g., "session_token"
	Regex      string `json:"regex,omitempty"`       // e.g., "token=([a-z0-9]+)"
	RegexGroup int    `json:"regex_group,omitempty"` // Which capture group to use (default: 1)
	SaveAs     string `json:"save_as"`               // Variable name to save extracted value
}

// Name returns the tool name
func (t *ExtractTool) Name() string {
	return "extract_value"
}

// Description returns the tool description
func (t *ExtractTool) Description() string {
	return "Extract values from the last HTTP response (JSON path, headers, cookies, regex) and save as a variable for use in subsequent requests"
}

// Parameters returns the tool parameter description
func (t *ExtractTool) Parameters() string {
	return `{
  "json_path": "$.data.user.id",
  "header": "X-Request-Id",
  "cookie": "session_token",
  "regex": "token=([a-z0-9]+)",
  "regex_group": 1,
  "save_as": "user_id"
}`
}

// Execute extracts a value from the last response
func (t *ExtractTool) Execute(args string) (string, error) {
	lastResponse := t.responseManager.GetHTTPResponse()
	if lastResponse == nil {
		return "", fmt.Errorf("no HTTP response available - make an http_request first")
	}

	var params ExtractParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse extraction parameters: %w", err)
	}

	if params.SaveAs == "" {
		return "", fmt.Errorf("'save_as' parameter is required")
	}

	var extractedValue string
	var extractionMethod string

	// Try each extraction method (only one should be specified)
	if params.JSONPath != "" {
		value, err := t.extractFromJSONPath(params.JSONPath, lastResponse)
		if err != nil {
			return "", fmt.Errorf("JSON path extraction failed: %w", err)
		}
		extractedValue = value
		extractionMethod = "JSON path"
	} else if params.Header != "" {
		value, ok := lastResponse.Headers[params.Header]
		if !ok {
			return "", fmt.Errorf("header '%s' not found in response", params.Header)
		}
		extractedValue = value
		extractionMethod = "header"
	} else if params.Cookie != "" {
		value, err := t.extractCookie(params.Cookie, lastResponse)
		if err != nil {
			return "", err
		}
		extractedValue = value
		extractionMethod = "cookie"
	} else if params.Regex != "" {
		group := params.RegexGroup
		if group == 0 {
			group = 1 // Default to first capture group
		}
		value, err := t.extractFromRegex(params.Regex, group, lastResponse)
		if err != nil {
			return "", fmt.Errorf("regex extraction failed: %w", err)
		}
		extractedValue = value
		extractionMethod = "regex"
	} else {
		return "", fmt.Errorf("no extraction method specified (json_path, header, cookie, or regex)")
	}

	// Save to variables
	t.variables.Set(params.SaveAs, extractedValue)

	return fmt.Sprintf("Extracted value from %s: '%s'\nSaved as variable: {{%s}}\n\nYou can now use {{%s}} in subsequent requests.",
		extractionMethod, extractedValue, params.SaveAs, params.SaveAs), nil
}

// extractFromJSONPath extracts a value using JSON path notation
func (t *ExtractTool) extractFromJSONPath(path string, lastResponse *HTTPResponse) (string, error) {
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(lastResponse.Body), &jsonData); err != nil {
		return "", fmt.Errorf("response body is not valid JSON: %w", err)
	}

	value, err := getJSONPath(jsonData, path)
	if err != nil {
		return "", err
	}

	// Convert value to string
	switch v := value.(type) {
	case string:
		return v, nil
	case float64:
		return fmt.Sprintf("%.0f", v), nil
	case bool:
		return fmt.Sprintf("%t", v), nil
	default:
		// For complex types, return JSON representation
		jsonBytes, _ := json.Marshal(v)
		return string(jsonBytes), nil
	}
}

// extractCookie extracts a cookie value from Set-Cookie headers
func (t *ExtractTool) extractCookie(cookieName string, lastResponse *HTTPResponse) (string, error) {
	setCookie, ok := lastResponse.Headers["Set-Cookie"]
	if !ok {
		return "", fmt.Errorf("no Set-Cookie header found")
	}

	// Parse Set-Cookie header (can be comma-separated list)
	cookies := strings.Split(setCookie, ",")
	for _, cookie := range cookies {
		// Cookie format: "name=value; Path=/; HttpOnly"
		parts := strings.Split(strings.TrimSpace(cookie), ";")
		if len(parts) > 0 {
			nameValue := strings.SplitN(parts[0], "=", 2)
			if len(nameValue) == 2 && strings.TrimSpace(nameValue[0]) == cookieName {
				return strings.TrimSpace(nameValue[1]), nil
			}
		}
	}

	return "", fmt.Errorf("cookie '%s' not found in Set-Cookie header", cookieName)
}

// extractFromRegex extracts a value using regex pattern matching
func (t *ExtractTool) extractFromRegex(pattern string, group int, lastResponse *HTTPResponse) (string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", fmt.Errorf("invalid regex pattern: %w", err)
	}

	matches := re.FindStringSubmatch(lastResponse.Body)
	if matches == nil {
		return "", fmt.Errorf("regex pattern '%s' did not match response body", pattern)
	}

	if group < 0 || group >= len(matches) {
		return "", fmt.Errorf("capture group %d not found (pattern has %d groups)", group, len(matches)-1)
	}

	return matches[group], nil
}
