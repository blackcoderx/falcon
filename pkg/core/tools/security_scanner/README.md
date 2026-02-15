# Security Scanner (`pkg/core/tools/security_scanner`)

The Security Scanner Module performs automated technical security audits on your API.

## Key Tool: `scan_security`

This tool runs a battery of security tests including OWASP Top 10 checks, fuzzing, and authentication auditing.

### Features

- **OWASP Checks**: Validates against common vulnerabilities like Injection, XSS, and Security Misconfiguration.
- **Fuzzing**: Sends malformed data to endpoints to detect crashes or improper error handling.
- **Auth Audit**: Checks for weak tokens, missing authorization checks, and privilege escalation risks.

## Usage

Use this tool to actively probe your API for vulnerabilities.

## Example Prompts

Trigger this tool by asking:
- "Run a security scan on the API."
- "Check the `/auth/login` endpoint for vulnerabilities."
- "Perform a fuzz test on the user input fields."
- "Audit the API for OWASP Top 10 issues."
