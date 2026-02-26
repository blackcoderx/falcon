package shared

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CompareResponsesTool compares API responses for regression testing
type CompareResponsesTool struct {
	responseManager *ResponseManager
	falconDir       string
}

// NewCompareResponsesTool creates a new response comparison tool
func NewCompareResponsesTool(responseManager *ResponseManager, falconDir string) *CompareResponsesTool {
	return &CompareResponsesTool{
		responseManager: responseManager,
		falconDir:       falconDir,
	}
}

// CompareParams defines comparison parameters
type CompareParams struct {
	Baseline     string   `json:"baseline"`                // Baseline response ID or "last_response"
	Current      string   `json:"current,omitempty"`       // Current response or "last_response"
	IgnoreFields []string `json:"ignore_fields,omitempty"` // Fields to ignore (e.g., "timestamp")
	IgnoreOrder  bool     `json:"ignore_order,omitempty"`  // Ignore array order
	Tolerance    float64  `json:"tolerance,omitempty"`     // Numeric tolerance (0.01 = 1%)
	SaveBaseline bool     `json:"save_baseline,omitempty"` // Save current as new baseline
}

// ComparisonResult represents the comparison outcome
type ComparisonResult struct {
	Match       bool     `json:"match"`
	Differences []string `json:"differences,omitempty"`
	Summary     string   `json:"summary"`
}

// Baseline stores a saved response
type Baseline struct {
	Name      string            `json:"name"`
	CreatedAt time.Time         `json:"created_at"`
	Response  string            `json:"response"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// Name returns the tool name
func (t *CompareResponsesTool) Name() string {
	return "compare_responses"
}

// Description returns the tool description
func (t *CompareResponsesTool) Description() string {
	return "Compare two API responses for regression testing. Detects added, removed, or changed fields."
}

// Parameters returns the tool parameter description
func (t *CompareResponsesTool) Parameters() string {
	return `{
  "baseline": "baseline_name",
  "current": "last_response",
  "ignore_fields": ["timestamp", "request_id"],
  "ignore_order": true,
  "tolerance": 0.01
}`
}

// Execute compares two responses
func (t *CompareResponsesTool) Execute(args string) (string, error) {
	var params CompareParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	// Save baseline if requested
	if params.SaveBaseline {
		return t.saveBaseline(params.Baseline)
	}

	// Load baseline
	baselineData, err := t.loadResponse(params.Baseline)
	if err != nil {
		return "", fmt.Errorf("failed to load baseline: %w", err)
	}

	// Load current response
	currentData, err := t.loadResponse(params.Current)
	if err != nil {
		return "", fmt.Errorf("failed to load current response: %w", err)
	}

	// Parse as JSON
	var baselineJSON, currentJSON interface{}
	if err := json.Unmarshal([]byte(baselineData), &baselineJSON); err != nil {
		return "", fmt.Errorf("baseline is not valid JSON: %w", err)
	}
	if err := json.Unmarshal([]byte(currentData), &currentJSON); err != nil {
		return "", fmt.Errorf("current response is not valid JSON: %w", err)
	}

	// Remove ignored fields
	if len(params.IgnoreFields) > 0 {
		baselineJSON = t.removeFields(baselineJSON, params.IgnoreFields)
		currentJSON = t.removeFields(currentJSON, params.IgnoreFields)
	}

	// Compare
	result := t.compareJSON(baselineJSON, currentJSON, "", params)

	// Format output
	return t.formatComparison(result), nil
}

// loadResponse loads a response (baseline file or last_response)
func (t *CompareResponsesTool) loadResponse(source string) (string, error) {
	if source == "" || source == "last_response" {
		lastResp := t.responseManager.GetHTTPResponse()
		if lastResp == nil {
			return "", fmt.Errorf("no HTTP response available")
		}
		return lastResp.Body, nil
	}

	// Load from baseline file
	baselinesDir := filepath.Join(t.falconDir, "baselines")
	baselinePath := filepath.Join(baselinesDir, source+".json")

	data, err := os.ReadFile(baselinePath)
	if err != nil {
		return "", fmt.Errorf("baseline '%s' not found", source)
	}

	var baseline Baseline
	if err := json.Unmarshal(data, &baseline); err != nil {
		return "", fmt.Errorf("invalid baseline file: %w", err)
	}

	return baseline.Response, nil
}

// saveBaseline saves the current response as a baseline
func (t *CompareResponsesTool) saveBaseline(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("baseline name is required")
	}

	lastResp := t.responseManager.GetHTTPResponse()
	if lastResp == nil {
		return "", fmt.Errorf("no HTTP response available to save")
	}

	// Create baselines directory
	baselinesDir := filepath.Join(t.falconDir, "baselines")
	if err := os.MkdirAll(baselinesDir, 0755); err != nil {
		return "", err
	}

	// Create baseline
	baseline := Baseline{
		Name:      name,
		CreatedAt: time.Now(),
		Response:  lastResp.Body,
		Metadata: map[string]string{
			"status_code": fmt.Sprintf("%d", lastResp.StatusCode),
		},
	}

	// Save to file
	data, err := json.MarshalIndent(baseline, "", "  ")
	if err != nil {
		return "", err
	}

	baselinePath := filepath.Join(baselinesDir, name+".json")
	if err := os.WriteFile(baselinePath, data, 0644); err != nil {
		return "", err
	}

	return fmt.Sprintf("Saved baseline: '%s'\nPath: %s\n\nUse in comparisons:\n{\n  \"baseline\": \"%s\",\n  \"current\": \"last_response\"\n}",
		name, baselinePath, name), nil
}

// removeFields removes specified fields from JSON
func (t *CompareResponsesTool) removeFields(data interface{}, fields []string) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			// Check if this field should be ignored
			ignored := false
			for _, ignoreField := range fields {
				if key == ignoreField {
					ignored = true
					break
				}
			}
			if !ignored {
				result[key] = t.removeFields(value, fields)
			}
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = t.removeFields(item, fields)
		}
		return result
	default:
		return data
	}
}

// compareJSON compares two JSON values
func (t *CompareResponsesTool) compareJSON(baseline, current interface{}, path string, params CompareParams) ComparisonResult {
	result := ComparisonResult{Match: true}

	switch baselineVal := baseline.(type) {
	case map[string]interface{}:
		currentMap, ok := current.(map[string]interface{})
		if !ok {
			result.Match = false
			result.Differences = append(result.Differences,
				fmt.Sprintf("Type mismatch at '%s': expected object, got %T", path, current))
			return result
		}

		// Check for missing fields in current
		for key := range baselineVal {
			keyPath := path + "." + key
			if path == "" {
				keyPath = key
			}

			if _, exists := currentMap[key]; !exists {
				result.Match = false
				result.Differences = append(result.Differences,
					fmt.Sprintf("Field removed: '%s'", keyPath))
			} else {
				// Recursively compare
				subResult := t.compareJSON(baselineVal[key], currentMap[key], keyPath, params)
				if !subResult.Match {
					result.Match = false
					result.Differences = append(result.Differences, subResult.Differences...)
				}
			}
		}

		// Check for new fields in current
		for key := range currentMap {
			if _, exists := baselineVal[key]; !exists {
				keyPath := path + "." + key
				if path == "" {
					keyPath = key
				}
				result.Match = false
				result.Differences = append(result.Differences,
					fmt.Sprintf("Field added: '%s'", keyPath))
			}
		}

	case []interface{}:
		currentArray, ok := current.([]interface{})
		if !ok {
			result.Match = false
			result.Differences = append(result.Differences,
				fmt.Sprintf("Type mismatch at '%s': expected array, got %T", path, current))
			return result
		}

		if len(baselineVal) != len(currentArray) {
			result.Match = false
			result.Differences = append(result.Differences,
				fmt.Sprintf("Array length mismatch at '%s': baseline has %d items, current has %d",
					path, len(baselineVal), len(currentArray)))
		}

		// Compare array elements
		minLen := len(baselineVal)
		if len(currentArray) < minLen {
			minLen = len(currentArray)
		}

		for i := 0; i < minLen; i++ {
			itemPath := fmt.Sprintf("%s[%d]", path, i)
			subResult := t.compareJSON(baselineVal[i], currentArray[i], itemPath, params)
			if !subResult.Match {
				result.Match = false
				result.Differences = append(result.Differences, subResult.Differences...)
			}
		}

	case float64:
		currentFloat, ok := current.(float64)
		if !ok {
			result.Match = false
			result.Differences = append(result.Differences,
				fmt.Sprintf("Type mismatch at '%s': expected number, got %T", path, current))
			return result
		}

		// Apply tolerance if specified
		if params.Tolerance > 0 {
			diff := math.Abs(baselineVal - currentFloat)
			allowedDiff := math.Abs(baselineVal * params.Tolerance)
			if diff > allowedDiff {
				result.Match = false
				result.Differences = append(result.Differences,
					fmt.Sprintf("Numeric difference at '%s': baseline=%.2f, current=%.2f (diff=%.2f, tolerance=%.2f%%)",
						path, baselineVal, currentFloat, diff, params.Tolerance*100))
			}
		} else if baselineVal != currentFloat {
			result.Match = false
			result.Differences = append(result.Differences,
				fmt.Sprintf("Value changed at '%s': baseline=%.2f, current=%.2f",
					path, baselineVal, currentFloat))
		}

	case string:
		currentStr, ok := current.(string)
		if !ok {
			result.Match = false
			result.Differences = append(result.Differences,
				fmt.Sprintf("Type mismatch at '%s': expected string, got %T", path, current))
			return result
		}

		if baselineVal != currentStr {
			result.Match = false
			result.Differences = append(result.Differences,
				fmt.Sprintf("Value changed at '%s': baseline='%s', current='%s'",
					path, baselineVal, currentStr))
		}

	case bool:
		currentBool, ok := current.(bool)
		if !ok {
			result.Match = false
			result.Differences = append(result.Differences,
				fmt.Sprintf("Type mismatch at '%s': expected boolean, got %T", path, current))
			return result
		}

		if baselineVal != currentBool {
			result.Match = false
			result.Differences = append(result.Differences,
				fmt.Sprintf("Value changed at '%s': baseline=%t, current=%t",
					path, baselineVal, currentBool))
		}

	case nil:
		if current != nil {
			result.Match = false
			result.Differences = append(result.Differences,
				fmt.Sprintf("Value changed at '%s': baseline=null, current=%v",
					path, current))
		}

	default:
		// Fallback to simple equality
		if baseline != current {
			result.Match = false
			result.Differences = append(result.Differences,
				fmt.Sprintf("Value changed at '%s': baseline=%v, current=%v",
					path, baseline, current))
		}
	}

	return result
}

// formatComparison formats the comparison result
func (t *CompareResponsesTool) formatComparison(result ComparisonResult) string {
	var sb strings.Builder

	if result.Match {
		sb.WriteString("✓ Responses Match\n\n")
		sb.WriteString("No differences detected between baseline and current response.\n")
	} else {
		sb.WriteString("✗ Responses Differ\n\n")
		sb.WriteString(fmt.Sprintf("Found %d difference(s):\n\n", len(result.Differences)))

		for i, diff := range result.Differences {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, diff))
		}

		sb.WriteString("\nTips:\n")
		sb.WriteString("- Use 'ignore_fields' to skip dynamic fields like timestamps\n")
		sb.WriteString("- Use 'tolerance' for numeric comparisons (e.g., 0.01 for 1%)\n")
		sb.WriteString("- Use 'ignore_order' for arrays where order doesn't matter\n")
	}

	return sb.String()
}
