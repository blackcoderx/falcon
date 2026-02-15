# Unit Test Scaffolder (`pkg/core/tools/unit_test_scaffolder`)

The Unit Test Scaffolder accelerates development by generating boilerplate test code.

## Key Tool: `scaffold_unit_tests`

This tool scans your source code (Go, Python, TS, etc.) and generates unit test files.

### Features

- **Mock Generation**: Automatically creates mocks for interfaces and dependencies.
- **Context Awareness**: understands the function signatures and generates relevant test cases.
- **Language Support**: Designed to be language-agnostic, with specific templates for major languages.

## Usage

Running this on a legacy codebase gives you a massive headstart on test coverage.

## Example Prompts

Trigger this tool by asking:
- "Scaffold unit tests for the `pkg/auth` directory."
- "Generate unit tests and mocks for `user_service.go`."
- "Create test boilerplate for the new payment controller."
