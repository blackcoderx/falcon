# Spec Ingester (`pkg/core/tools/spec_ingester`)

The Spec Ingester Module is the entry point for ZAP's knowledge of your API.

## Key Tool: `ingest_spec`

This tool parses OpenAPI (Swagger) or Postman collection files and transforms them into ZAP's internal **API Knowledge Graph**.

### Features

- **Format Support**: Handles JSON/YAML OpenAPI v2/v3 and Postman Collections.
- **Graph Construction**: Builds a queryable graph of endpoints, schemas, and parameters.
- **Validation**: Checks the spec for basic syntax errors during ingestion.

## Usage

Run this tool *first* to teach ZAP about your API.

## Example Prompts

Trigger this tool by asking:
- "Ingest the OpenAPI spec from `docs/openapi.yaml`."
- "Load the Postman collection located at `./postman/v1.json`."
- "Parse the API specification to build the knowledge graph."
