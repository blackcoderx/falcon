# Idempotency Verifier (`pkg/core/tools/idempotency_verifier`)

The Idempotency Verifier checks if API endpoints correctly handle repeated requests.

## Key Tool: `verify_idempotency`

This tool repeats requests to non-safe endpoints (POST, PUT, PATCH) to detect unintended side effects.

### Features

- **Double-Submit Detection**: Checks if sending the same request twice creates two records (when it shouldn't).
- **State Integrity**: Verifies resource state remains consistent after multiple identical calls.

## Usage

Critical for payment APIs and order processing systems where duplicate transactions are dangerous.

## Example Prompts

Trigger this tool by asking:
- "Verify that the payment endpoint is idempotent."
- "Check if submitting the order twice creates duplicate records."
- "Ensure that retrying a POST request doesn't cause side effects."
