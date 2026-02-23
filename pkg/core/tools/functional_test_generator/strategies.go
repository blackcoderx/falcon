package functional_test_generator

import (
	"fmt"
	"strings"

	"github.com/blackcoderx/falcon/pkg/core/tools/shared"
)

// StrategyEngine is responsible for generating test scenarios based on different strategies.
type StrategyEngine struct {
	strategies map[string]Strategy
}

// Strategy defines the interface for test generation strategies.
type Strategy interface {
	// Generate creates test scenarios for a given endpoint.
	Generate(endpointKey string, analysis shared.EndpointAnalysis, baseURL string) []shared.TestScenario
	Name() string
}

// NewStrategyEngine creates a new strategy engine with all available strategies.
func NewStrategyEngine() *StrategyEngine {
	return &StrategyEngine{
		strategies: map[string]Strategy{
			"happy":    &HappyPathStrategy{},
			"negative": &NegativeStrategy{},
			"boundary": &BoundaryStrategy{},
		},
	}
}

// Generate creates test scenarios using the specified strategy.
func (e *StrategyEngine) Generate(
	endpointKey string,
	analysis shared.EndpointAnalysis,
	baseURL string,
	strategyName string,
) []shared.TestScenario {
	strategy, ok := e.strategies[strategyName]
	if !ok {
		return nil
	}
	return strategy.Generate(endpointKey, analysis, baseURL)
}

// HappyPathStrategy generates valid test cases with correct data.
type HappyPathStrategy struct{}

func (s *HappyPathStrategy) Name() string {
	return "happy"
}

// Generate creates happy path scenarios with valid data for every endpoint.
func (s *HappyPathStrategy) Generate(endpointKey string, analysis shared.EndpointAnalysis, baseURL string) []shared.TestScenario {
	// Parse method and path from endpoint key (e.g., "GET /api/users")
	parts := strings.SplitN(endpointKey, " ", 2)
	if len(parts) != 2 {
		return nil
	}
	method := parts[0]
	path := parts[1]

	// Build URL
	url := strings.TrimSuffix(baseURL, "/") + "/" + strings.TrimPrefix(path, "/")

	// Generate valid body based on parameters
	body := s.generateValidBody(analysis.Parameters)

	// Determine expected status code
	expectedStatus := s.getExpectedSuccessStatus(method, analysis.Responses)

	scenario := shared.TestScenario{
		ID:          fmt.Sprintf("happy_%s_%s", method, sanitizePath(path)),
		Name:        fmt.Sprintf("Happy Path: %s %s", method, path),
		Category:    "functional",
		Severity:    "high",
		Description: fmt.Sprintf("Valid request to %s with correct parameters", endpointKey),
		Method:      method,
		URL:         url,
		Headers:     map[string]string{"Content-Type": "application/json"},
		Body:        body,
		Expected: shared.TestExpectation{
			StatusCodeRange: &shared.StatusCodeRange{Min: 200, Max: 299},
			StatusCode:      expectedStatus,
		},
	}

	return []shared.TestScenario{scenario}
}

// generateValidBody creates a valid request body based on parameters.
func (s *HappyPathStrategy) generateValidBody(params []shared.Parameter) map[string]interface{} {
	if len(params) == 0 {
		return nil
	}

	body := make(map[string]interface{})
	for _, param := range params {
		// Only include body parameters
		if strings.Contains(param.Description, "in: body") || strings.Contains(param.Description, "in: requestBody") {
			body[param.Name] = s.generateValidValue(param.Type, param.Name)
		}
	}

	if len(body) == 0 {
		return nil
	}
	return body
}

// generateValidValue creates a valid value based on the parameter type.
func (s *HappyPathStrategy) generateValidValue(paramType string, paramName string) interface{} {
	switch paramType {
	case "string":
		// Use semantic naming for better test data
		if strings.Contains(strings.ToLower(paramName), "email") {
			return "user@example.com"
		}
		if strings.Contains(strings.ToLower(paramName), "name") {
			return "Test User"
		}
		if strings.Contains(strings.ToLower(paramName), "phone") {
			return "+1234567890"
		}
		return "valid_string_value"
	case "integer", "int", "int32", "int64":
		return 123
	case "number", "float", "double":
		return 123.45
	case "boolean", "bool":
		return true
	case "array":
		return []interface{}{"item1", "item2"}
	case "object":
		return map[string]interface{}{"key": "value"}
	default:
		return "default_value"
	}
}

// getExpectedSuccessStatus determines the expected success status code.
func (s *HappyPathStrategy) getExpectedSuccessStatus(method string, responses []shared.Response) int {
	// Check if there's a success response code (2xx) in the spec
	for _, resp := range responses {
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return resp.StatusCode
		}
	}

	// Default based on HTTP method
	switch method {
	case "POST":
		return 201
	case "DELETE":
		return 204
	case "PUT", "PATCH":
		return 200
	default:
		return 200
	}
}

// NegativeStrategy generates invalid test cases with missing/wrong data.
type NegativeStrategy struct{}

func (s *NegativeStrategy) Name() string {
	return "negative"
}

// Generate creates negative test scenarios with invalid data.
func (s *NegativeStrategy) Generate(endpointKey string, analysis shared.EndpointAnalysis, baseURL string) []shared.TestScenario {
	parts := strings.SplitN(endpointKey, " ", 2)
	if len(parts) != 2 {
		return nil
	}
	method := parts[0]
	path := parts[1]

	url := strings.TrimSuffix(baseURL, "/") + "/" + strings.TrimPrefix(path, "/")

	var scenarios []shared.TestScenario

	// Test 1: Missing required fields
	if len(analysis.Parameters) > 0 {
		requiredParams := s.getRequiredParams(analysis.Parameters)
		if len(requiredParams) > 0 {
			scenarios = append(scenarios, shared.TestScenario{
				ID:          fmt.Sprintf("negative_missing_required_%s_%s", method, sanitizePath(path)),
				Name:        fmt.Sprintf("Negative: Missing Required Fields - %s %s", method, path),
				Category:    "functional",
				Severity:    "high",
				Description: fmt.Sprintf("Request to %s with missing required fields", endpointKey),
				Method:      method,
				URL:         url,
				Headers:     map[string]string{"Content-Type": "application/json"},
				Body:        map[string]interface{}{}, // Empty body
				Expected: shared.TestExpectation{
					StatusCode:      400,
					StatusCodeRange: &shared.StatusCodeRange{Min: 400, Max: 499},
				},
			})
		}
	}

	// Test 2: Wrong data types
	if len(analysis.Parameters) > 0 {
		wrongTypeBody := s.generateWrongTypeBody(analysis.Parameters)
		if wrongTypeBody != nil {
			scenarios = append(scenarios, shared.TestScenario{
				ID:          fmt.Sprintf("negative_wrong_type_%s_%s", method, sanitizePath(path)),
				Name:        fmt.Sprintf("Negative: Wrong Data Types - %s %s", method, path),
				Category:    "functional",
				Severity:    "medium",
				Description: fmt.Sprintf("Request to %s with incorrect data types", endpointKey),
				Method:      method,
				URL:         url,
				Headers:     map[string]string{"Content-Type": "application/json"},
				Body:        wrongTypeBody,
				Expected: shared.TestExpectation{
					StatusCode:      400,
					StatusCodeRange: &shared.StatusCodeRange{Min: 400, Max: 499},
				},
			})
		}
	}

	// Test 3: Invalid values
	scenarios = append(scenarios, shared.TestScenario{
		ID:          fmt.Sprintf("negative_invalid_values_%s_%s", method, sanitizePath(path)),
		Name:        fmt.Sprintf("Negative: Invalid Values - %s %s", method, path),
		Category:    "functional",
		Severity:    "medium",
		Description: fmt.Sprintf("Request to %s with invalid field values", endpointKey),
		Method:      method,
		URL:         url,
		Headers:     map[string]string{"Content-Type": "application/json"},
		Body:        s.generateInvalidBody(analysis.Parameters),
		Expected: shared.TestExpectation{
			StatusCode:      400,
			StatusCodeRange: &shared.StatusCodeRange{Min: 400, Max: 499},
		},
	})

	return scenarios
}

// getRequiredParams filters parameters to find required ones.
func (s *NegativeStrategy) getRequiredParams(params []shared.Parameter) []shared.Parameter {
	var required []shared.Parameter
	for _, param := range params {
		if param.Required {
			required = append(required, param)
		}
	}
	return required
}

// generateWrongTypeBody creates a body with wrong data types.
func (s *NegativeStrategy) generateWrongTypeBody(params []shared.Parameter) map[string]interface{} {
	body := make(map[string]interface{})
	for _, param := range params {
		if strings.Contains(param.Description, "in: body") {
			// Intentionally use wrong type (string instead of number, etc.)
			switch param.Type {
			case "integer", "number", "float":
				body[param.Name] = "not_a_number"
			case "boolean":
				body[param.Name] = "not_a_boolean"
			case "array":
				body[param.Name] = "not_an_array"
			default:
				body[param.Name] = 12345 // number instead of string
			}
		}
	}
	if len(body) == 0 {
		return nil
	}
	return body
}

// generateInvalidBody creates a body with semantically invalid values.
func (s *NegativeStrategy) generateInvalidBody(params []shared.Parameter) map[string]interface{} {
	body := make(map[string]interface{})
	for _, param := range params {
		if strings.Contains(param.Description, "in: body") {
			switch param.Type {
			case "string":
				if strings.Contains(strings.ToLower(param.Name), "email") {
					body[param.Name] = "invalid-email"
				} else {
					body[param.Name] = ""
				}
			case "integer", "number":
				body[param.Name] = -999999
			default:
				body[param.Name] = nil
			}
		}
	}
	return body
}

// BoundaryStrategy generates test cases with boundary values.
type BoundaryStrategy struct{}

func (s *BoundaryStrategy) Name() string {
	return "boundary"
}

// Generate creates boundary test scenarios.
func (s *BoundaryStrategy) Generate(endpointKey string, analysis shared.EndpointAnalysis, baseURL string) []shared.TestScenario {
	parts := strings.SplitN(endpointKey, " ", 2)
	if len(parts) != 2 {
		return nil
	}
	method := parts[0]
	path := parts[1]

	url := strings.TrimSuffix(baseURL, "/") + "/" + strings.TrimPrefix(path, "/")

	var scenarios []shared.TestScenario

	// Test 1: Empty strings
	scenarios = append(scenarios, shared.TestScenario{
		ID:          fmt.Sprintf("boundary_empty_strings_%s_%s", method, sanitizePath(path)),
		Name:        fmt.Sprintf("Boundary: Empty Strings - %s %s", method, path),
		Category:    "functional",
		Severity:    "medium",
		Description: fmt.Sprintf("Request to %s with empty string values", endpointKey),
		Method:      method,
		URL:         url,
		Headers:     map[string]string{"Content-Type": "application/json"},
		Body:        s.generateEmptyStringBody(analysis.Parameters),
		Expected: shared.TestExpectation{
			StatusCodeRange: &shared.StatusCodeRange{Min: 400, Max: 499},
		},
	})

	// Test 2: Maximum values
	scenarios = append(scenarios, shared.TestScenario{
		ID:          fmt.Sprintf("boundary_max_values_%s_%s", method, sanitizePath(path)),
		Name:        fmt.Sprintf("Boundary: Maximum Values - %s %s", method, path),
		Category:    "functional",
		Severity:    "medium",
		Description: fmt.Sprintf("Request to %s with maximum boundary values", endpointKey),
		Method:      method,
		URL:         url,
		Headers:     map[string]string{"Content-Type": "application/json"},
		Body:        s.generateMaxValueBody(analysis.Parameters),
		Expected: shared.TestExpectation{
			StatusCodeRange: &shared.StatusCodeRange{Min: 200, Max: 499},
		},
	})

	// Test 3: Large payload
	scenarios = append(scenarios, shared.TestScenario{
		ID:          fmt.Sprintf("boundary_large_payload_%s_%s", method, sanitizePath(path)),
		Name:        fmt.Sprintf("Boundary: Large Payload - %s %s", method, path),
		Category:    "functional",
		Severity:    "low",
		Description: fmt.Sprintf("Request to %s with very large payload", endpointKey),
		Method:      method,
		URL:         url,
		Headers:     map[string]string{"Content-Type": "application/json"},
		Body:        s.generateLargePayload(analysis.Parameters),
		Expected: shared.TestExpectation{
			StatusCodeRange: &shared.StatusCodeRange{Min: 200, Max: 499},
		},
	})

	return scenarios
}

// generateEmptyStringBody creates a body with empty strings.
func (s *BoundaryStrategy) generateEmptyStringBody(params []shared.Parameter) map[string]interface{} {
	body := make(map[string]interface{})
	for _, param := range params {
		if strings.Contains(param.Description, "in: body") {
			switch param.Type {
			case "string":
				body[param.Name] = ""
			case "integer", "number":
				body[param.Name] = 0
			case "array":
				body[param.Name] = []interface{}{}
			default:
				body[param.Name] = ""
			}
		}
	}
	return body
}

// generateMaxValueBody creates a body with maximum values.
func (s *BoundaryStrategy) generateMaxValueBody(params []shared.Parameter) map[string]interface{} {
	body := make(map[string]interface{})
	for _, param := range params {
		if strings.Contains(param.Description, "in: body") {
			switch param.Type {
			case "string":
				// Very long string (1000 characters)
				body[param.Name] = strings.Repeat("x", 1000)
			case "integer":
				body[param.Name] = 2147483647 // Max int32
			case "number", "float":
				body[param.Name] = 1.7976931348623157e+308 // Max float64
			case "array":
				// Large array
				arr := make([]interface{}, 100)
				for i := range arr {
					arr[i] = i
				}
				body[param.Name] = arr
			default:
				body[param.Name] = strings.Repeat("x", 1000)
			}
		}
	}
	return body
}

// generateLargePayload creates a very large payload.
func (s *BoundaryStrategy) generateLargePayload(_ []shared.Parameter) map[string]interface{} {
	body := make(map[string]interface{})

	// Create a large nested structure
	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("field_%d", i)
		body[key] = strings.Repeat("Large payload data ", 100)
	}

	return body
}

// sanitizePath converts a path to a safe identifier.
func sanitizePath(path string) string {
	s := strings.ReplaceAll(path, "/", "_")
	s = strings.ReplaceAll(s, ":", "")
	s = strings.ReplaceAll(s, "{", "")
	s = strings.ReplaceAll(s, "}", "")
	s = strings.Trim(s, "_")
	return s
}
