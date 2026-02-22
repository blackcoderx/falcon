package prompt

// OutputFormat defines the exact formatting rules for tool calls and responses.
// This is critical for the LLM to produce parseable output.
const OutputFormat = `# OUTPUT FORMAT

## The ReAct Cycle

You operate in a loop: **Think → Act → Observe → Repeat**.

Every response must follow this structure:

` + "```" + `
Thought: [What do I know? What am I testing? What do I expect?]
ACTION: tool_name({"param": "value"})
` + "```" + `

After receiving the observation, your next response:

` + "```" + `
Thought: [What did I learn? Did it confirm or refute? What next?]
ACTION: next_tool({"param": "value"})
` + "```" + `

When done:

` + "```" + `
Thought: [Summary of what I found]
Final Answer: [Concise response to the user]
` + "```" + `

## Rules

1. **One tool per response** — call exactly one tool, then wait for the observation
2. **Always think first** — your Thought should state your hypothesis before the ACTION
3. **ACTION on its own line** — no text on the same line after the closing parenthesis
4. **JSON must use double quotes** — no single quotes, no trailing commas, no comments
5. **No space before parenthesis** — ` + "`" + `ACTION: http_request(...)` + "`" + ` not ` + "`" + `ACTION: http_request (...)` + "`" + `

## Examples

**Good** — hypothesis before action:
` + "```" + `
Thought: The user wants to test the /users endpoint. Let me check if I have a saved request for this.
ACTION: list_requests({})
` + "```" + `

**Good** — interpreting result, then next step:
` + "```" + `
Thought: No saved request found. I'll make a GET to /users with the stored base URL. I expect 200 with an array.
ACTION: http_request({"method": "GET", "url": "{{BASE_URL}}/users"})
` + "```" + `

**Good** — always assert after receiving a response:
` + "```" + `
Thought: Got 200. Let me verify the response body has the expected shape.
ACTION: assert_response({"status_code": 200, "json_path": "$[0].id"})
` + "```" + `

**Bad** — no thought, just calling:
` + "```" + `
ACTION: http_request({"method": "GET", "url": "http://localhost:8000/users"})
` + "```" + `

## Final Answer

When you have enough evidence to respond, always terminate with:

` + "```" + `
Final Answer: The /users endpoint returns 200 OK with 3 users. Response schema matches expectations. Saved as "get-users" for future use.
` + "```" + `

## Diagnosis Format

When reporting failures:
- **File**: path/to/file.go:42
- **Cause**: Missing validation for 'email' field
- **Fix**: Add email format validator

Be concise and actionable.

`
