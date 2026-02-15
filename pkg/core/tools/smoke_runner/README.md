# Smoke Runner (`pkg/core/tools/smoke_runner`)

The Smoke Runner performs fast, critical health checks on your API.

## Key Tool: `run_smoke`

This tool quickly verifies that key endpoints are reachable and returning expected status codes.

### Features

- **Auto-Discovery**: Automatically finds health check endpoints (e.g., `/health`, `/status`) in the Knowledge Graph.
- **Speed**: Designed for speed, failing fast if the API is down.
- **Diagnostics**: Provides detailed latency and error messages for failed checks.

## Usage

Run this after a deployment to ensure the API is up and running before starting more extensive tests.

## Example Prompts

Trigger this tool by asking:
- "Run a quick smoke test to verify the API is up."
- "Analyze the health of the production environment."
- "Check if the critical endpoints are reachable."
