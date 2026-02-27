# Functional Test Generator (`pkg/core/tools/functional_test_generator`)

The Functional Test Generator creates distinct test scenarios from the API Knowledge Graph using various strategies.

## Key Tool: `generate_functional_tests`

This tool generates and optionally executes comprehensive functional tests.

### features

- **Strategies**:
    - **Happy Path**: Valid requests with random valid data.
    - **Negative**: Invalid inputs, missing fields, wrong types.
    - **Boundary**: Edge cases, min/max values.
- **Filtering**: Target specific endpoints or strategies.
- **Export**: Automatically saves a Markdown report of all generated scenarios to `.falcon/reports/functional_report_<timestamp>.md`.

## Usage

Use this tool to rapidly create a test suite from your API specification.

## Example Prompts

Trigger this tool by asking:
- "Generate a full suite of functional tests for the User API."
- "Create negative test scenarios for the `/login` endpoint."
- "I need boundary tests for the simplified order processing flow."
- "Generate and run happy path tests for all GET endpoints."
