# Security Scanner Module

## Overview
The Security Scanner provides a multi-layered security analysis of API endpoints, mapping results to the OWASP Top 10 vulnerabilities.

## Tools
- `scan_security`: Executes OWASP checks, input fuzzing, and auth auditing.

## Checks
- **OWASP Top 10**: Access control, injection, crypto failures, misconfigurations, etc.
- **Fuzzing**: SQLi, XSS, Command Injection, Path Traversal, SSRF, XXE.
- **Auth Audit**: Horizontal/Vertical privilege escalation, JWT security, Session Fixation.

## Usage
```json
{
  "base_url": "http://localhost:3000",
  "scan_types": ["owasp", "fuzz", "auth"],
  "depth": "deep"
}
```
