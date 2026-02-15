# Regression Watchdog (`pkg/core/tools/regression_watchdog`)

The Regression Watchdog guards against unintended changes by comparing current API behavior against a known stable baseline.

## Key Tool: `check_regression`

This tool executes validation checks against a previous "snapshot" of the API.

### Features

- **Baseline Comparison**: Compares status codes, headers, and body structure against a stored baseline.
- **Drift Detection**: Alerts on changes in response times or payload sizes beyond a threshold.
- **Snapshot Management**: Automatically updates the baseline when changes are approved.

## Usage

Use this as a safety net before a deployment to catch "silent" breakage in existing endpoints.

## Example Prompts

Trigger this tool by asking:
- "Check for regressions against the 'stable_v1' baseline."
- "Verify that the latest changes didn't break existing functionality."
- "Run a regression test on the user endpoints."
