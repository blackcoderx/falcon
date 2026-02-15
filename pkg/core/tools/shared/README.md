# Shared Tools (`pkg/core/tools/shared`)

This directory contains the foundational tools and managers used across the ZAP ecosystem. These tools provide the low-level capabilities for HTTP communication, authentication, assertion, and state management.

## Core Services

- **ResponseManager**: Stores and shares the last HTTP response across tools.
- **VariableStore**: Manages session and global variables (with {{VAR}} substitution).
- **ConfirmationManager**: Handles human-in-the-loop approval for destructive operations.

## Foundational Tools

- **http_request**: The primary engine for making API calls.
- **assertions**: Validates status codes, headers, and body content.
- **extraction**: Pulls values from responses (JSON path, regex) into variables.
- **auth**: Implements Bearer, Basic, and OAuth2 authentication flows.
- **schema_validator**: Ensures responses conform to JSON Schema definitions.
- **timing**: Provides `wait` and `retry` (with exponential backoff) logic.
- **test_suite**: Orchestrates multiple related tests into a single execution.
- **webhook_listener**: Captures incoming callbacks for asynchronous API testing.

## Usage

These tools are typically used as the building blocks for higher-level autonomous workflows. They ensure consistency in how ZAP interacts with APIs and handles data.

## Example Prompts

Trigger these tools by asking:
- "Make a GET request to `/health`."
- "Assert that the response status code is 200."
- "Extract the `token` from the login response and save it as a variable."
- "Authenticate using Bearer token `eyJ...`."
