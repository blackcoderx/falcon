# Data Driven Engine (`pkg/core/tools/data_driven_engine`)

The Data Driven Engine executes test scenarios using external data sources.

## Key Tool: `run_data_driven`

This tool iterates over a dataset (CSV, JSON) and injects values into a test scenario template.

### Features

- **Data Sources**: Supports CSV and JSON files, or simulated "fake" data.
- **Variable Mapping**: Maps column names to request templates (e.g., `{{email}}` -> `user@example.com`).
- **Batch Processing**: Executes the scenario for every row in the dataset.

## Reports

After every run, `run_data_driven` automatically writes a Markdown report to `.falcon/reports/`. Pass `report_name` to set the filename (e.g. `data_driven_report_users`). If omitted, the filename defaults to `data_driven_report_<timestamp>.md`. A validator confirms the file has content before the tool returns success.

## Usage

Ideal for testing bulk creation endpoints or checking how an API handles a wide variety of inputs.

## Example Prompts

Trigger this tool by asking:
- "Test the registration endpoint using the user data in `users.csv`."
- "Run data-driven tests on the login API using the provided JSON dataset."
- "Verify valid and invalid email formats by iterating through `emails.csv`."
