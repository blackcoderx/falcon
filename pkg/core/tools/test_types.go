package tools

// TestScenario represents a single test case with full categorization
type TestScenario struct {
	ID          string            `json:"id"`          // "sec-001", "val-001", etc.
	Name        string            `json:"name"`        // "SQL injection in promo_code"
	Category    string            `json:"category"`    // "security" | "validation" | "happy_path" | "performance" | "edge_case"
	Severity    string            `json:"severity"`    // "critical" | "high" | "medium" | "low"
	Description string            `json:"description"` // Detailed test description
	Method      string            `json:"method"`      // "POST", "GET", etc.
	URL         string            `json:"url"`         // "/api/checkout"
	Headers     map[string]string `json:"headers,omitempty"`
	Body        interface{}       `json:"body,omitempty"`
	Expected    TestExpectation   `json:"expected"`
	OWASPRef    string            `json:"owasp_ref,omitempty"` // "A03:2021 - Injection"
	CWERef      string            `json:"cwe_ref,omitempty"`   // "CWE-89"
}

type TestExpectation struct {
	StatusCode      int                    `json:"status_code"`
	StatusCodeRange *StatusCodeRange       `json:"status_code_range,omitempty"` // Alternative to exact match
	BodyContains    []string               `json:"body_contains,omitempty"`
	BodyNotContains []string               `json:"body_not_contains,omitempty"`
	HeaderContains  map[string]string      `json:"header_contains,omitempty"`
	MaxDurationMs   int                    `json:"max_duration_ms,omitempty"`
	JSONPath        map[string]interface{} `json:"json_path,omitempty"` // JSONPath assertions
}

type StatusCodeRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

type TestResult struct {
	ScenarioID     string   `json:"scenario_id"`
	ScenarioName   string   `json:"scenario_name"`
	Category       string   `json:"category"`
	Passed         bool     `json:"passed"`
	ActualStatus   int      `json:"actual_status"`
	ExpectedStatus int      `json:"expected_status"`
	Error          string   `json:"error,omitempty"`
	DurationMs     int64    `json:"duration_ms"`
	ResponseBody   string   `json:"response_body,omitempty"`
	Logs           []string `json:"logs"`
	Severity       string   `json:"severity,omitempty"`
	OWASPRef       string   `json:"owasp_ref,omitempty"`
	Timestamp      string   `json:"timestamp"`
}
