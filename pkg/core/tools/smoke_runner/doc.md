# Smoke Runner Module

## Overview
The Smoke Runner provides ultra-fast health checks to verify API reachability and core functionality. Use this as a "pre-flight" check before running heavy test suites.

## Tools
- `run_smoke`: Executes pings and health/status checks.

## Features
- **Reachability**: Verifies the API is up and responding.
- **Health Mapping**: Automatically identifies `/health`, `/status`, and `/ping` endpoints.
- **Fast Execution**: Designed for < 5s execution time.

## Usage
```json
{
  "base_url": "http://localhost:3000",
  "detailed": true
}
```
