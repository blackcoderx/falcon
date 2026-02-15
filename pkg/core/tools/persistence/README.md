# Persistence (`pkg/core/tools/persistence`)

This directory contains tools for managing ZAP's persistent state, including saved requests and environment variables.

## Key Features

- **Request Storage**: Save and load complex HTTP requests as YAML files with `{{VAR}}` placeholders.
- **Environment Management**: Switch between different environments (dev, prod, staging) with specific variable sets.
- **Global Variables**: Manage persistent global variables across the entire ZAP session.

## Tools

- **save_request / load_request**: The primary way to bookmark and reuse API calls.
- **list_requests**: View all saved requests in the current workspace.
- **set_environment / list_environments**: Interface for managing deployment contexts.
- **variable (Global/Session)**: Tool for manual variable manipulation.

## Example Prompts

Trigger these tools by asking:
- "Save this request as 'create_user' for later use."
- "Load the 'production' environment variables."
- "List all my saved requests."
- "Set a global variable `API_KEY` to `12345`."
