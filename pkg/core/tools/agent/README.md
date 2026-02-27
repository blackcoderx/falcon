# Agent Core & Orchestration (`pkg/core/tools/agent`)

This directory contains the high-level tools that govern the ZAP agent's "brain" and autonomous testing workflows.

## Primary Tools

- **memory**: Recall and save project-specific knowledge (base URLs, auth patterns) across sessions.
- **auto_test**: The "Magic" autonomous testing engine. It orchestrates analysis, generation, and execution for a target endpoint.
- **run_tests / run_single_test**: High-level test runners that execute generated scenarios in parallel. Automatically saves a Markdown report to `.falcon/reports/` after every run — pass `report_name` to control the filename (e.g. `test_report_users_api`).

## Reports

`run_tests` writes directly to `.falcon/reports/<report_name>.md`. No separate export step is needed. A validator confirms the file has content before the tool returns success.

## Usage

These tools are usually invoked by the agent itself during autonomous cycles, or by the user when requesting a high-level summary or a full scan of a system.

## Example Prompts

Trigger these tools by asking:
- "Run a full autonomous test on the user service."
- "Remember that the base URL for the staging environment is `https://api.staging.example.com`."
- "Run the generated test suite for the payment module."
