package debugging

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/blackcoderx/falcon/pkg/core"
	"github.com/blackcoderx/falcon/pkg/core/tools/shared"
)

// AutoFixTool orchestrates the full fix-and-verify loop:
// run test → analyze failure → find handler → propose fix → write (with confirmation) → re-run test.
// Retries up to MaxAttempts times if each fix attempt still fails.
type AutoFixTool struct {
	findHandler    *FindHandlerTool
	readFile       *ReadFileTool
	proposeFix     *ProposeFixTool
	writeFile      *WriteFileTool
	testExecutor   *shared.TestExecutor
	analyzeFailure *AnalyzeFailureTool
	eventCallback  core.EventCallback
}

// NewAutoFixTool creates a new auto_fix orchestrator.
func NewAutoFixTool(
	findHandler *FindHandlerTool,
	readFile *ReadFileTool,
	proposeFix *ProposeFixTool,
	writeFile *WriteFileTool,
	testExecutor *shared.TestExecutor,
	analyzeFailure *AnalyzeFailureTool,
) *AutoFixTool {
	return &AutoFixTool{
		findHandler:    findHandler,
		readFile:       readFile,
		proposeFix:     proposeFix,
		writeFile:      writeFile,
		testExecutor:   testExecutor,
		analyzeFailure: analyzeFailure,
	}
}

// SetEventCallback implements ConfirmableTool — propagates to WriteFileTool so the
// TUI confirmation dialog fires correctly when write_file is called internally.
func (t *AutoFixTool) SetEventCallback(callback core.EventCallback) {
	t.eventCallback = callback
}

// AutoFixParams defines input for auto_fix.
type AutoFixParams struct {
	Endpoint       string               `json:"endpoint"`          // e.g. "POST /api/users"
	BaseURL        string               `json:"base_url"`          // e.g. "http://localhost:8080"
	Scenario       *shared.TestScenario `json:"scenario,omitempty"` // optional pre-built scenario
	ExpectedStatus int                  `json:"expected_status,omitempty"` // default 200
	MaxAttempts    int                  `json:"max_attempts,omitempty"`    // default 3
}

func (t *AutoFixTool) Name() string {
	return "auto_fix"
}

func (t *AutoFixTool) Description() string {
	return "Autonomous fix-and-verify loop: confirms a test is failing, locates the handler file, generates a code fix, applies it (with user confirmation showing a diff), then re-runs the test to verify. Retries up to max_attempts times if the fix doesn't resolve the failure."
}

func (t *AutoFixTool) Parameters() string {
	return `{
  "endpoint": "POST /api/users",
  "base_url": "http://localhost:8080",
  "expected_status": 201,
  "max_attempts": 3
}`
}

func (t *AutoFixTool) Execute(args string) (string, error) {
	var params AutoFixParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	if params.BaseURL == "" {
		return "", fmt.Errorf("base_url is required")
	}
	if params.MaxAttempts <= 0 {
		params.MaxAttempts = 3
	}
	if params.ExpectedStatus == 0 {
		params.ExpectedStatus = 200
	}

	scenario := t.buildScenario(params)

	var report strings.Builder
	fmt.Fprintf(&report, "# Auto-Fix Report: %s\n\n", params.Endpoint)

	for attempt := 1; attempt <= params.MaxAttempts; attempt++ {
		fmt.Fprintf(&report, "## Attempt %d\n\n", attempt)

		// 1. Run test — if already passing, we're done
		result := t.testExecutor.RunScenario(scenario, params.BaseURL)
		if result.Passed {
			fmt.Fprintf(&report, "- Verification: PASSED ✓\n\n")
			fmt.Fprintf(&report, "**Final: Fixed in %d attempt(s).**\n", attempt)
			return report.String(), nil
		}
		fmt.Fprintf(&report, "- Test failed: %s\n", result.Error)

		// 2. Analyze failure and locate the handler
		rootCause, handlerFile := t.analyzeAndLocate(result, params.Endpoint, &report)
		if handlerFile == "" {
			fmt.Fprintf(&report, "- Could not locate handler file — stopping.\n")
			break
		}

		// 3. Propose and apply fix
		applied, done := t.applyFix(handlerFile, rootCause, result.Error, attempt, &report)
		if done {
			// User rejected or unrecoverable error
			return report.String(), nil
		}
		if !applied {
			break
		}
	}

	fmt.Fprintf(&report, "\n**Final: Could not resolve the failure after %d attempt(s).**\n", params.MaxAttempts)
	return report.String(), nil
}

// buildScenario returns the provided scenario or constructs a minimal one from params.
func (t *AutoFixTool) buildScenario(params AutoFixParams) shared.TestScenario {
	if params.Scenario != nil {
		return *params.Scenario
	}

	parts := strings.SplitN(params.Endpoint, " ", 2)
	method, path := "GET", params.Endpoint
	if len(parts) == 2 {
		method, path = parts[0], parts[1]
	}

	return shared.TestScenario{
		ID:     "auto_fix_verify",
		Name:   fmt.Sprintf("auto_fix: %s", params.Endpoint),
		Method: method,
		URL:    path,
		Expected: shared.TestExpectation{
			StatusCode: params.ExpectedStatus,
		},
	}
}

// analyzeAndLocate calls analyze_failure and find_handler, returning root cause and handler path.
func (t *AutoFixTool) analyzeAndLocate(result shared.TestResult, endpoint string, report *strings.Builder) (string, string) {
	// Analyze failure for root cause
	rootCause := result.Error
	failParams := AnalyzeFailureParams{
		TestResult:       result,
		ResponseBody:     result.ResponseBody,
		ExpectedBehavior: fmt.Sprintf("Status %d", result.ExpectedStatus),
	}
	failJSON, _ := json.Marshal(failParams)
	if failAnalysis, err := t.analyzeFailure.Execute(string(failJSON)); err == nil {
		var parsed map[string]interface{}
		if json.Unmarshal([]byte(failAnalysis), &parsed) == nil {
			if explanation, ok := parsed["explanation"].(string); ok && explanation != "" {
				rootCause = explanation
			}
		}
	}
	fmt.Fprintf(report, "- Root cause: %s\n", rootCause)

	// Find handler file
	parts := strings.SplitN(endpoint, " ", 2)
	method, path := "GET", endpoint
	if len(parts) == 2 {
		method, path = parts[0], parts[1]
	}
	findArgs, _ := json.Marshal(FindHandlerParams{
		Endpoint: endpoint,
		Method:   method,
		Path:     path,
	})
	handlerResult, err := t.findHandler.Execute(string(findArgs))
	if err != nil {
		return rootCause, ""
	}
	var handlerInfo HandlerInfo
	if json.Unmarshal([]byte(handlerResult), &handlerInfo) != nil || handlerInfo.File == "" {
		return rootCause, ""
	}
	fmt.Fprintf(report, "- Handler: %s\n", handlerInfo.File)
	return rootCause, handlerInfo.File
}

// applyFix proposes a fix and applies it via write_file.
// Returns (applied bool, done bool) where done=true means the loop should terminate early.
func (t *AutoFixTool) applyFix(handlerFile, rootCause, failureError string, attempt int, report *strings.Builder) (bool, bool) {
	fixParams := ProposeFixParams{
		File:          handlerFile,
		Vulnerability: rootCause,
		FailedTest:    failureError,
	}
	fixJSON, _ := json.Marshal(fixParams)
	fixResult, err := t.proposeFix.Execute(string(fixJSON))
	if err != nil {
		fmt.Fprintf(report, "- Fix proposal failed: %v\n", err)
		return false, false
	}

	var parsed map[string]interface{}
	if json.Unmarshal([]byte(fixResult), &parsed) != nil {
		fmt.Fprintf(report, "- Could not parse fix proposal.\n")
		return false, false
	}

	patchedContent, _ := parsed["patched_content"].(string)
	explanation, _ := parsed["explanation"].(string)
	if patchedContent == "" {
		fmt.Fprintf(report, "- propose_fix returned no patched_content.\n")
		return false, false
	}
	fmt.Fprintf(report, "- Proposed fix: %s\n", explanation)

	// Propagate eventCallback so TUI shows the diff + confirmation dialog
	if t.eventCallback != nil {
		t.writeFile.SetEventCallback(t.eventCallback)
	}

	writeArgs, _ := json.Marshal(WriteFileParams{
		Path:    handlerFile,
		Content: patchedContent,
	})
	writeResult, err := t.writeFile.Execute(string(writeArgs))
	if err != nil {
		fmt.Fprintf(report, "- Write failed: %v\n", err)
		return false, false
	}

	if strings.Contains(writeResult, "rejected") {
		fmt.Fprintf(report, "- Fix rejected by user.\n")
		fmt.Fprintf(report, "\n**Final: Stopped at attempt %d — user declined the fix.**\n", attempt)
		return false, true // done=true, stop the loop
	}

	fmt.Fprintf(report, "- Fix applied to %s\n", handlerFile)
	return true, false
}
