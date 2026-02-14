package shared

// TestScenario represents a single test case with full categorization.
type TestScenario struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Category    string            `json:"category"`
	Severity    string            `json:"severity"`
	Description string            `json:"description"`
	Method      string            `json:"method"`
	URL         string            `json:"url"`
	Headers     map[string]string `json:"headers,omitempty"`
	Body        interface{}       `json:"body,omitempty"`
	Expected    TestExpectation   `json:"expected"`
	OWASPRef    string            `json:"owasp_ref,omitempty"`
	CWERef      string            `json:"cwe_ref,omitempty"`
}

// TestExpectation defines what a test expects from the response.
type TestExpectation struct {
	StatusCode      int                    `json:"status_code"`
	StatusCodeRange *StatusCodeRange       `json:"status_code_range,omitempty"`
	BodyContains    []string               `json:"body_contains,omitempty"`
	BodyNotContains []string               `json:"body_not_contains,omitempty"`
	HeaderContains  map[string]string      `json:"header_contains,omitempty"`
	MaxDurationMs   int                    `json:"max_duration_ms,omitempty"`
	JSONPath        map[string]interface{} `json:"json_path,omitempty"`
}

// StatusCodeRange allows matching a range of status codes.
type StatusCodeRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

// TestResult represents the outcome of executing a test scenario.
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

// EndpointAnalysis represents the structured output of analysis
type EndpointAnalysis struct {
	Summary    string          `json:"summary"`
	Parameters []Parameter     `json:"parameters"`
	AuthType   string          `json:"auth_type"`
	Responses  []Response      `json:"responses"`
	Security   []SecurityRisks `json:"security_risks"`
}

type Parameter struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

type Response struct {
	StatusCode  int    `json:"status_code"`
	Description string `json:"description"`
}

type SecurityRisks struct {
	Risk        string `json:"risk"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
}
