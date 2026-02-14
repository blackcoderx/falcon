# Schema Conformance Module

## Overview
The Schema Conformance module ensures that your API actually does what the specification says. It performs deep validation of response bodies against the JSON Schemas defined in the API Knowledge Graph.

## Tools
- `verify_schema_conformance`: Runs conformance checks against live endpoints.

## Features
- **Strict Mode**: Fails if the API returns fields not documented in the schema.
- **Type Checking**: Validates all field types (string, number, boolean, object, array).
- **Graph Integration**: Pulls latest schemas directly from the spec ingester's output.

## Usage
```json
{
  "base_url": "http://localhost:3000",
  "strict": true
}
```
