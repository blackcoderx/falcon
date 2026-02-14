# Idempotency Verifier Module

## Overview
The Idempotency Verifier ensures that repeating the same request doesn't cause unexpected state changes or duplicate data.

## Tools
- `verify_idempotency`: Tests POST/PUT/PATCH/DELETE endpoints for idempotency.

## Features
- **Repeat Execution**: Automatically repeats the same request multiple times.
- **Side Effect Detection**: Compares status codes and response bodies to find consistency issues.
- **Duplicate Detection**: Specifically looks for "201 Created" on repeated POST requests.

## Usage
```json
{
  "base_url": "http://localhost:3000",
  "endpoints": ["POST /api/orders"],
  "repeat_count": 2
}
```
