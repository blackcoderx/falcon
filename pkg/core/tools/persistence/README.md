# Persistence (`pkg/core/tools/persistence`)

This directory contains tools for managing Falcon's persistent state, including saved requests and environment variables.

## Key Features

- **Request Storage**: Save and load complex HTTP requests as YAML files with `{{VAR}}` placeholders.
- **Environment Management**: Switch between different environments (dev, prod, staging) with specific variable sets.
- **Variable Scope**: Session-scoped variables (cleared on exit) or global-scoped variables (persistent in `.falcon/variables.json`).

## Merged Tools (2)

To reduce confusion, multiple tools were merged into two unified tools with action parameters:

### `request` (replaces: save_request, load_request, list_requests)

```json
{"action": "save", "name": "create_user", "method": "POST", "url": "...", "headers": {}, "body": "..."}
{"action": "load", "name": "create_user"}
{"action": "list"}
```

Persists to `.falcon/requests/<name>.yaml`. Requests can include `{{VAR}}` placeholders for variable substitution.

### `environment` (replaces: set_environment, list_environments)

```json
{"action": "set", "name": "prod", "variables": {"API_KEY": "...", "BASE_URL": "..."}}
{"action": "list"}
```

Persists to `.falcon/environments/<name>.yaml`. Switch environments to change which variables are active.

## Variable Scope

- **Session scope**: Cleared when the conversation ends. Use for temporary tokens, test IDs.
  - Set: `variable({"action": "set", "name": "auth_token", "value": "...", "scope": "session"})`
  - Get: `variable({"action": "get", "name": "auth_token", "scope": "session"})`

- **Global scope**: Persisted in `.falcon/variables.json` across sessions. Use for API keys, base URLs, framework names.
  - Set: `variable({"action": "set", "name": "API_KEY", "value": "...", "scope": "global"})`
  - Get: `variable({"action": "get", "name": "API_KEY", "scope": "global"})`

## Example Prompts

Trigger these tools by asking:
- "Save this request as 'create_user' for later use."
- "Load and run the saved 'create_user' request."
- "List all my saved requests."
- "Set the environment to 'production'."
- "Set a global variable `API_KEY` to `12345`."
- "Get the current value of `auth_token`."
