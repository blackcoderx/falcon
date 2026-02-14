# Documentation Validator Module

## Overview
The Documentation Validator ensures that your human-readable docs (READMEs, Swagger-UI descriptions) match the reality of the API code.

## Tools
- `validate_docs`: Scans documentation files for consistency errors.

## Features
- **Markdown Parsing**: Extracts endpoint definitions from markdown files.
- **Cross-Reference**: Checks documented parameters against those found in the API Knowledge Graph.
- **Truth Verification**: Prevents "stale documentation" bugs by flagging missing or incorrect info.

## Usage
```json
{
  "doc_path": "README.md"
}
```
