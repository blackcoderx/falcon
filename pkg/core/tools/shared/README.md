# Shared Tools (`pkg/core/tools/shared`)

This directory contains the foundational tools and managers used across Falcon. These tools provide the low-level capabilities for HTTP communication, authentication, assertion, state management, and .falcon artifact handling.

## Core Services

- **ResponseManager**: Stores and shares the last HTTP response across tools.
- **VariableStore**: Manages session-scoped and global variables (with {{VAR}} substitution).
- **ConfirmationManager**: Handles human-in-the-loop approval for destructive operations.

## Core Tools (5)

Essential for every interaction:
- **`http_request`**: Make HTTP requests (GET, POST, PUT, DELETE, PATCH) with headers, auth, body
- **`variable`**: Get/set variables in session scope (cleared on exit) or global scope (persistent)
- **`auth`**: Unified authentication — bearer, basic, OAuth2, JWT parsing, basic auth decoding
- **`wait`**: Delay between requests (seconds, backoff, polling)
- **`retry`**: Retry failed tool calls with exponential backoff

## Validation & Extraction Tools (3)

Test individual responses:
- **`assert_response`**: Validate HTTP status, response body content, JSON paths, headers
- **`extract_value`**: Extract values from response (JSON path, header, cookie, regex) into variables for chaining
- **`validate_json_schema`**: Strict JSON Schema validation against spec

## .falcon Artifact Tools (3)

Manage persistent artifacts in the .falcon folder:
- **`falcon_write`**: Write validated YAML/JSON/Markdown to .falcon/ (with path safety: blocks traversal, protected files, syntax validation)
- **`falcon_read`**: Read artifacts from .falcon/ (reports, flows, specs) — scoped to .falcon only
- **`session_log`**: Create session audit trail — start/end timestamps, summary, searchable history

## Managers & Helpers

Tools rely on these internal managers for consistency:
- **ReportValidator**: Validates reports (`ValidateReportContent`) and falcon.md (`ValidateFalconMD`) after writes
- **AuthManager**: Delegates to BearerTool, BasicTool, OAuth2Tool internally

## Usage

These tools are the building blocks for higher-level autonomous workflows. They ensure consistency in how Falcon interacts with APIs, validates responses, and persists knowledge.

## Example Prompts

Trigger these tools by:
- "Make a GET request to `/health`."
- "Assert that the response status code is 200."
- "Extract the `token` from the login response and save it as `auth_token`."
- "Authenticate using Bearer token `eyJ...`."
- "Save this response to `.falcon/flows/unit_get_users.yaml`."
- "Recall what we know about the API from falcon.md."
- "Log this session — we tested the auth endpoints and found a CORS bug."
