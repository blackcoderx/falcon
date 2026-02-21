package prompt

// OutputFormat defines the exact formatting rules for tool calls and responses.
// This is critical for the LLM to produce parseable output.
const OutputFormat = `# OUTPUT FORMAT - CRITICAL

## One Tool Per Turn
Call **exactly one tool per response**. Wait for the observation before deciding the next action.
Never emit two ACTION: lines in a single response.

## ReAct Cycle (repeat until done)
` + "```" + `
ACTION: tool_name({"param": "value"})
--- observation arrives ---
ACTION: next_tool(...)
--- repeat ---
Final Answer: <concise response to the user>
` + "```" + `

## Tool Call Format

**Syntax**: ACTION: tool_name({"param": "value"})

**Rules**:
1. ACTION must be on its own line
2. No space before opening parenthesis
3. JSON must use double quotes (not single)
4. No trailing commas in JSON
5. No comments in JSON

## Valid Examples

` + "```" + `
ACTION: http_request({"method": "GET", "url": "http://localhost:8000/api/users"})
` + "```" + `

` + "```" + `
ACTION: search_code({"pattern": "/api/users", "file_pattern": "*.go"})
` + "```" + `

` + "```" + `
ACTION: variable({"action": "set", "name": "token", "value": "abc123", "scope": "session"})
` + "```" + `

## Invalid Examples (DO NOT DO)

❌ Missing quotes:
` + "```" + `
ACTION: http_request({method: "GET", url: "http://localhost:8000"})
` + "```" + `

❌ Single quotes:
` + "```" + `
ACTION: http_request({'method': 'GET'})
` + "```" + `

❌ Trailing comma:
` + "```" + `
ACTION: http_request({"method": "GET",})
` + "```" + `

❌ Space before parenthesis:
` + "```" + `
ACTION: http_request ({"method": "GET"})
` + "```" + `

❌ Multiple tool calls in one response (NEVER do this):
` + "```" + `
ACTION: search_code({"pattern": "test"})
ACTION: read_file({"path": "test.py"})
` + "```" + `

## Final Answer Format

When you're done analyzing and ready to respond to the user, use:

` + "```" + `
Final Answer: The API returned 200 OK. User created successfully with ID 123.
` + "```" + `

**Always use ` + "`" + `Final Answer:` + "`" + ` as the termination signal.** This tells the system to stop the ReAct loop and display your response.

## Diagnosis Format

When diagnosing errors, include:
- **File**: path/to/file.go:42
- **Cause**: Validation missing for 'email' field
- **Fix**: Add email validator in Pydantic model

Be concise and actionable.

`
