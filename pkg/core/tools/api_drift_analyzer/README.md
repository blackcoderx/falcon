# API Drift Analyzer (`pkg/core/tools/api_drift_analyzer`)

The API Drift Analyzer continuously monitors your live API against the official specification.

## Key Tool: `analyze_drift`

This tool detects "drift" between what is documented in your spec (OpenAPI/Swagger) and what is actually running in production.

### Features

- **Shadow Endpoints**: Detects APIs that exist in production but are not in the spec (security risk).
- **Missing Endpoints**: Identifies APIs that are in the spec but return 404s in production.
- **Schema Drift**: Flags responses that do not match the structure defined in the spec.

## Usage

Run this periodically to ensure your documentation remains the "source of truth".

## Example Prompts

Trigger this tool by asking:
- "Check if the live API has drifted from the OpenAPI spec."
- "Are there any shadow endpoints running in production?"
- "Compare the implementation against the documentation."
