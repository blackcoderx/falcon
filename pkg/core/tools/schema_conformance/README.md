# Schema Conformance (`pkg/core/tools/schema_conformance`)

The Schema Conformance module enforces strict adherence to the API specification.

## Key Tool: `verify_schema_conformance`

This tool validates every API response against its definition in the Knowledge Graph (OpenAPI spec).

### Features

- **Strict Validation**: Fails on undefined fields or incorrect data types.
- **Header checks**: Verifies required headers are present.
- **Reporting**: detailed breakdown of every schema violation found.

## Usage

Ensures your implementation doesn't drift away from the contract defined in your OpenAPI/Swagger spec.

## Example Prompts

Trigger this tool by asking:
- "Verify that all API responses strictly follow the OpenAPI schema."
- "Check for schema violations in the `/products` endpoint."
- "Ensure the API output matches the defined specification."
