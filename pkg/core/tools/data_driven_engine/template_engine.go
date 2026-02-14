package data_driven_engine

import (
	"encoding/json"
	"strings"

	"github.com/blackcoderx/zap/pkg/core/tools/shared"
)

// TemplateEngine replaces placeholders in test scenarios with actual data.
type TemplateEngine struct{}

// Populate replaces {{var}} placeholders in the scenario with values from the data row.
func (e *TemplateEngine) Populate(template shared.TestScenario, data map[string]interface{}) shared.TestScenario {
	// Deep copy via JSON (simplest for demonstration)
	var scenario shared.TestScenario
	bytes, _ := json.Marshal(template)
	json.Unmarshal(bytes, &scenario)

	// Replace in URL
	for k, v := range data {
		placeholder := "{{" + k + "}}"
		valStr := ""
		if s, ok := v.(string); ok {
			valStr = s
		} else {
			valStr = strings.Trim(string(bytes), "\"") // simplified
		}

		scenario.URL = strings.ReplaceAll(scenario.URL, placeholder, valStr)
	}

	// Replace in Body (if map)
	if scenario.Body != nil {
		bodyBytes, _ := json.Marshal(scenario.Body)
		bodyStr := string(bodyBytes)
		for k, v := range data {
			placeholder := "{{" + k + "}}"
			valBytes, _ := json.Marshal(v)
			valStr := string(valBytes)
			// Avoid double quotes if value is already a string
			if strings.HasPrefix(valStr, "\"") && strings.HasSuffix(valStr, "\"") {
				valStr = valStr[1 : len(valStr)-1]
			}
			bodyStr = strings.ReplaceAll(bodyStr, placeholder, valStr)
		}
		json.Unmarshal([]byte(bodyStr), &scenario.Body)
	}

	return scenario
}
