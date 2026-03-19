# Falcon — Coding & Debugging Capability Gaps

This document captures the gaps between what Falcon's debugging tools claim to do and what they actually implement, along with concrete proposals for making Falcon genuinely good at code fixing and debugging.

---

## Current State Summary

Falcon's debugging pipeline has two tiers:

| Tier | Tools | Reality |
|------|-------|---------|
| **Real** | `read_file`, `list_files`, `search_code`, `write_file` | Actual disk I/O and filesystem operations |
| **LLM wrapper** | `propose_fix`, `analyze_endpoint`, `analyze_failure`, `create_test_file` | Build a prompt, call `llmClient.Chat()`, return raw LLM text |

The file I/O layer is solid. The intelligence layer is just prompting.

---

## Gap 1: `propose_fix` Never Applies the Fix

### What it does today

1. Receives vulnerability description + code snippet as input parameters
2. Sends a single LLM prompt asking for a JSON with a `diff` field
3. Strips markdown fences from the response
4. Returns the JSON string to the agent

The diff never touches the filesystem. The agent receives fix text and stops.

### What it should do

The fix pipeline should be:

```
propose_fix(file, issue)
  → read_file(path)          ← get actual current code
  → LLM(code + issue)        ← generate unified diff against real content
  → write_file(path, patch)  ← apply with user confirmation
  → re-run failed test       ← verify the fix worked
```

### Proposed implementation

**Option A — Inline application (simplest):**

In `propose_fix.go`, after generating the diff, call `write_file` directly:

```go
// After LLM returns the diff:
if params.Apply && params.FilePath != "" {
    writeArgs := fmt.Sprintf(`{"path": %q, "content": %q}`, params.FilePath, patchedContent)
    result, err := t.writeFile.Execute(writeArgs)
    // write_file already handles confirmation + disk write
}
```

**Option B — Two-phase tool (better UX):**

Split into `propose_fix` (show diff, ask approval) and `apply_fix` (write file). The agent calls both. This keeps each tool single-purpose and the confirmation visible.

**Option C — Auto-patch loop (best for autonomous debugging):**

Add an `auto_fix` orchestrator that:
1. Calls `propose_fix` to generate the patch
2. Calls `write_file` with the patched content (triggers user confirmation)
3. Re-runs the failing test
4. If the test still fails, loops back to step 1 with the new error (max 3 iterations)

### Impact

Without this, `propose_fix` is a suggestion engine. With it, Falcon becomes a code-patching agent.

---

## Gap 2: `create_test_file` Returns Content but Never Writes It

### What it does today

Returns `{"filename": "test_users.go", "content": "...test code..."}` as a JSON string. The file never exists on disk unless the agent separately calls `write_file`.

### What it should do

Write the file immediately after generation (with confirmation), or at minimum emit a `write_pending` event so the TUI prompts the user.

### Proposed implementation

```go
func (t *CreateTestFileTool) Execute(args string) (string, error) {
    // ... existing LLM call to generate content ...

    // After generation:
    writeArgs := fmt.Sprintf(`{"path": %q, "content": %q}`, filename, content)
    return t.writeFile.Execute(writeArgs) // triggers confirmation + disk write
}
```

`CreateTestFileTool` already has access to `workDir`. It just needs a reference to `WriteFileTool` injected via the registry.

---

## Gap 3: Analysis Tools Don't Read the Actual Code

### What they do today

`analyze_endpoint` and `analyze_failure` receive structured parameters (method, URL, status code, response body) and ask the LLM to reason about them. They never look at the source code.

### The problem

The LLM is guessing. It has no visibility into:
- The actual handler implementation
- Middleware chain
- Database queries
- Auth logic

This produces generic advice ("check your input validation") rather than specific fixes.

### Proposed implementation

**`analyze_endpoint` should read the handler first:**

```go
func (t *AnalyzeEndpointTool) Execute(args string) (string, error) {
    // 1. Find the handler file
    handlerResult, _ := t.findHandler.Execute(fmt.Sprintf(`{"path": %q, "framework": %q}`, params.Path, t.framework))

    // 2. Read the handler file
    fileContent, _ := t.readFile.Execute(fmt.Sprintf(`{"path": %q}`, extractedFilePath))

    // 3. Send actual code to LLM
    prompt := buildPromptWithCode(params, fileContent)
    response, _ := t.llmClient.Chat([]llm.Message{{Role: "user", Content: prompt}})
    return response, nil
}
```

**`analyze_failure` should use the stack trace:**

```go
func (t *AnalyzeFailureTool) Execute(args string) (string, error) {
    // 1. Parse stack trace from failure output
    locations := core.ParseStackTrace(params.ErrorOutput)

    // 2. Read each relevant file
    var codeContext strings.Builder
    for _, loc := range locations {
        content, _ := t.readFile.Execute(fmt.Sprintf(`{"path": %q, "line": %d, "context": 20}`, loc.File, loc.Line))
        codeContext.WriteString(content)
    }

    // 3. Analyze with real code in context
    prompt := buildFailurePrompt(params, codeContext.String())
    response, _ := t.llmClient.Chat(...)
    return response, nil
}
```

`core.ParseStackTrace()` already exists in `analysis.go` and handles Python, Go, and JS traces. It just isn't being used by the analysis tools.

---

## Gap 4: `find_handler` Only Knows 4 Frameworks

### What it does today

Has specific regex patterns for Gin, Echo, FastAPI, and Express. Returns a generic path search for everything else.

### What it should cover

All 15 frameworks in `SupportedFrameworks`:

| Framework | Pattern to add |
|-----------|---------------|
| `chi` | `r.Method("GET", "/path"`, `r.Get("/path"` |
| `fiber` | `app.Get("/path"`, `app.Post("/path"` |
| `flask` | `@app.route("/path"`, `@blueprint.route(` |
| `django` | `path('/path'`, `re_path(r'/path'` |
| `nestjs` | `@Get('/path')`, `@Post('/path')`, `@Controller(` |
| `hono` | `app.get('/path'`, `app.post('/path'` |
| `spring` | `@GetMapping("/path")`, `@PostMapping(`, `@RequestMapping(` |
| `laravel` | `Route::get('/path'`, `Route::post('/path'` |
| `rails` | `get '/path'`, `resources :name` |
| `actix` | `web::get().to(`, `.route("/path"` |
| `axum` | `get(handler)`, `Router::new().route("/path"` |

### Proposed implementation

Extend `getSearchPatterns()` in `find_handler.go` with a `case` block per framework. The patterns already work — the method just needs more entries.

---

## Gap 5: No Verify-After-Fix Loop

### What's missing

After `write_file` applies a patch, Falcon has no mechanism to:
1. Re-run the test that was failing
2. Confirm the fix actually worked
3. Roll back if the fix made things worse

### Proposed `auto_fix` orchestrator

A new `auto_fix` tool (or extension of `auto_tester`) that closes the loop:

```
auto_fix(endpoint, test_scenario)
  │
  ├─ 1. run_tests(scenario)          → get failure details
  ├─ 2. analyze_failure(failure)     → get root cause
  ├─ 3. find_handler(endpoint)       → locate source file
  ├─ 4. read_file(handler_file)      → get actual code
  ├─ 5. propose_fix(code + cause)    → generate patch
  ├─ 6. write_file(patch)            → apply with confirmation
  ├─ 7. run_tests(scenario)          → verify fix
  │
  └─ if still failing and attempt < 3:
       loop back to step 2 with new error
```

This is the core loop that makes a debugging agent actually useful. `auto_tester` already does steps 1, 2, and 7 for generation. The missing piece is 3–6.

---

## Gap 6: `propose_fix` Works on Snippets, Not Full Files

### What it does today

Receives a `code_snippet` parameter — the caller is responsible for providing the relevant code. The LLM has no context about surrounding functions, imports, or the full file.

### The problem

Fixes that require import changes, type adjustments, or changes to multiple functions will be wrong or incomplete because the LLM sees only a fragment.

### Proposed fix

Before calling the LLM, always read the full file:

```go
// In propose_fix.go Execute():
fullContent, err := os.ReadFile(filepath.Join(t.workDir, params.FilePath))
if err == nil {
    params.CodeSnippet = string(fullContent) // replace snippet with full file
}
```

Then the diff the LLM generates will be correct against the full file context, and `write_file` can apply it cleanly.

---

## Recommended Priority Order

| Priority | Change | Effort | Impact |
|----------|--------|--------|--------|
| ~~1~~ | ~~Fix `run_performance` mock — call `httpTool.Execute()` in `executeRequest()`~~ | ~~Very Low~~ | ~~Very High~~ — **RESOLVED** |
| 2 | `propose_fix` reads full file before calling LLM | Low | High — fixes context problem immediately |
| 3 | `propose_fix` calls `write_file` to apply the patch | Low | High — closes the biggest debugging gap |
| 4 | `create_test_file` calls `write_file` after generation | Low | Medium — removes a manual step |
| 5 | `analyze_failure` reads files from parsed stack trace | Medium | High — analysis becomes grounded in real code |
| 6 | `analyze_endpoint` reads handler before analysis | Medium | High — advice becomes specific, not generic |
| 7 | Expand `find_handler` patterns to all 15 frameworks | Low | Medium — broader project support |
| 8 | Add `auto_fix` orchestrator with verify loop | High | Very high — turns Falcon into a real debugger |

---

## What a Good Debugging Session Looks Like After These Changes

```
User: "POST /api/users is returning 500"

Falcon:
  1. find_handler("POST /api/users")         → users_handler.go:42
  2. read_file("users_handler.go")            → full handler code
  3. run_tests("POST /api/users smoke")       → 500, NilPointerException line 67
  4. analyze_failure(error + code context)    → "db.Users is nil — missing DB injection"
  5. propose_fix(full file + root cause)      → unified diff: add DB nil check + return 503
  6. write_file(patch)                        → [shows diff, waits for Y/N]
  7. run_tests("POST /api/users smoke")       → 201 Created ✓
  8. create_test_file("nil DB scenario")      → writes test_users_db_nil_test.go

Final Answer: "Fixed. The handler was missing a nil check on the DB connection.
               Patch applied to users_handler.go. Regression test written."
```

Today Falcon does steps 1–4 and stops. Steps 5–8 require manual follow-up.
