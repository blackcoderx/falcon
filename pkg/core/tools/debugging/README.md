# Debugging Tools (`pkg/core/tools/debugging`)

The Debugging Module provides a suite of tools to analyze errors, locate faulty code, and propose fixes.

## Key Tools

### 1. `analyze_failure`

Uses an LLM to interpret detailed error logs and test failure reports to explain *why* something went wrong.

### 2. `find_handler`

Locates the exact file and function in your codebase that handles a specific API endpoint (e.g., finds `HandleLogin` for `POST /login`).

### 3. `propose_fix`

Generates a code patch to resolve a specific bug or vulnerability found during testing.

## Usage

These tools are typically used in response to a failed test or a user report.

## Example Prompts

Trigger these tools by asking:
- "Why did the login test fail?" (**analyze_failure**)
- "Find the code responsible for the `/orders` endpoint." (**find_handler**)
- "Propose a fix for the nil pointer exception in `auth_service.go`." (**propose_fix**)
- "Debug the 500 error on the registration page."
