package tools

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

// SchemaValidationTool validates JSON responses against JSON Schema
type SchemaValidationTool struct {
	responseManager *ResponseManager
}

// NewSchemaValidationTool creates a new schema validation tool
func NewSchemaValidationTool(responseManager *ResponseManager) *SchemaValidationTool {
	return &SchemaValidationTool{
		responseManager: responseManager,
	}
}

// SchemaParams defines schema validation parameters
type SchemaParams struct {
	Schema       interface{} `json:"schema"`                  // Inline schema or file path
	SchemaURL    string      `json:"schema_url,omitempty"`    // Schema from URL
	ResponseBody string      `json:"response_body,omitempty"` // Or use last_response
}

// Name returns the tool name
func (t *SchemaValidationTool) Name() string {
	return "validate_json_schema"
}

// Description returns the tool description
func (t *SchemaValidationTool) Description() string {
	return "Validate JSON response body against a JSON Schema specification (draft-07, draft-2020-12)"
}

// Parameters returns the tool parameter description
func (t *SchemaValidationTool) Parameters() string {
	return `{
  "schema": {
    "type": "object",
    "required": ["id", "name"],
    "properties": {
      "id": {"type": "integer"},
      "name": {"type": "string"},
      "email": {"type": "string", "format": "email"}
    }
  }
}`
}

// Execute validates the response against the schema
func (t *SchemaValidationTool) Execute(args string) (string, error) {
	var params SchemaParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	// Get the response body to validate
	var responseBody string
	if params.ResponseBody != "" {
		responseBody = params.ResponseBody
	} else {
		lastResponse := t.responseManager.GetHTTPResponse()
		if lastResponse == nil {
			return "", fmt.Errorf("no HTTP response available - make an http_request first")
		}
		responseBody = lastResponse.Body
	}

	// Load the schema
	var schemaLoader gojsonschema.JSONLoader

	if params.SchemaURL != "" {
		// Load schema from URL
		schemaLoader = gojsonschema.NewReferenceLoader(params.SchemaURL)
	} else if params.Schema != nil {
		// Load inline schema
		schemaJSON, err := json.Marshal(params.Schema)
		if err != nil {
			return "", fmt.Errorf("failed to marshal schema: %w", err)
		}
		schemaLoader = gojsonschema.NewBytesLoader(schemaJSON)
	} else {
		return "", fmt.Errorf("either 'schema' or 'schema_url' must be provided")
	}

	// Load the document to validate
	documentLoader := gojsonschema.NewStringLoader(responseBody)

	// Validate
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return "", fmt.Errorf("schema validation error: %w", err)
	}

	// Format results
	var sb strings.Builder

	if result.Valid() {
		sb.WriteString("✓ JSON Schema validation passed\n\n")
		sb.WriteString("The response body conforms to the provided schema.")
	} else {
		sb.WriteString("✗ JSON Schema validation failed\n\n")
		sb.WriteString(fmt.Sprintf("Found %d validation error(s):\n\n", len(result.Errors())))

		for i, err := range result.Errors() {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, formatValidationError(err)))
			sb.WriteString(fmt.Sprintf("   Context: %s\n", err.Context().String()))
			if err.Details() != nil && len(err.Details()) > 0 {
				sb.WriteString(fmt.Sprintf("   Details: %v\n", err.Details()))
			}
			sb.WriteString("\n")
		}

		// Add helpful summary
		sb.WriteString("Common fixes:\n")
		sb.WriteString("- Check field types match schema (string, integer, boolean, etc.)\n")
		sb.WriteString("- Ensure all required fields are present\n")
		sb.WriteString("- Validate format constraints (email, uri, date-time, etc.)\n")
		sb.WriteString("- Check numeric ranges (minimum, maximum)\n")
		sb.WriteString("- Verify string lengths (minLength, maxLength)\n")
	}

	return sb.String(), nil
}

// formatValidationError formats a validation error for display
func formatValidationError(err gojsonschema.ResultError) string {
	desc := err.Description()

	// Make error messages more user-friendly
	switch err.Type() {
	case "required":
		return fmt.Sprintf("Required field missing: %s", err.Field())
	case "invalid_type":
		if err.Details() != nil {
			expected := err.Details()["expected"]
			actual := err.Details()["given"]
			return fmt.Sprintf("Type mismatch at '%s': expected %v, got %v", err.Field(), expected, actual)
		}
	case "number_any_of", "number_one_of":
		return fmt.Sprintf("Value at '%s' doesn't match any allowed schemas", err.Field())
	case "format":
		return fmt.Sprintf("Format validation failed at '%s': %s", err.Field(), desc)
	case "enum":
		return fmt.Sprintf("Value at '%s' not in allowed enum values: %s", err.Field(), desc)
	case "minimum":
		return fmt.Sprintf("Value at '%s' below minimum: %s", err.Field(), desc)
	case "maximum":
		return fmt.Sprintf("Value at '%s' above maximum: %s", err.Field(), desc)
	case "min_length":
		return fmt.Sprintf("String at '%s' too short: %s", err.Field(), desc)
	case "max_length":
		return fmt.Sprintf("String at '%s' too long: %s", err.Field(), desc)
	case "pattern":
		return fmt.Sprintf("String at '%s' doesn't match pattern: %s", err.Field(), desc)
	case "additional_properties_false":
		return fmt.Sprintf("Unexpected property '%s' (additional properties not allowed)", err.Field())
	}

	return desc
}
