# Documentation Validator (`pkg/core/tools/documentation_validator`)

The Documentation Validator ensures your external docs (READMEs, Wikis) match the actual API implementation.

## Key Tool: `validate_docs`

This tool parses markdown documentation and verifies referenced endpoints against the API Knowledge Graph.

### Features

- **Consistency Check**: Ensures valid HTTP methods and paths are used in examples.
- **Drift Detection**: Flags documentation that refers to deprecated or removed endpoints.

## Usage

Keep your developer portal and READMEs accurate without manual review.

## Example Prompts

Trigger this tool by asking:
- "Validate the API examples in `README.md`."
- "Check if the documentation is up to date with the code."
- "Verify that all endpoints mentioned in the docs actually exist."
