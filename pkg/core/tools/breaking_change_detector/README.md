# Breaking Change Detector (`pkg/core/tools/breaking_change_detector`)

The Breaking Change Detector analyzes two API specifications to identify backward-incompatible changes.

## Key Tool: `detect_breaking_changes`

This tool compares an "old" spec version against a "new" spec version.

### Features

- **Categorization**: Classifies changes as:
    - **Breaking** (🔴): Removed endpoints, new required params, type changes.
    - **Minor** (🟡): New optional params, new endpoints.
    - **Patch** (🟢): Description updates, typo fixes.
- **Summary**: Provides a clear change log.

## Usage

Use this in CI pipelines before merging PRs to prevent accidental breaking changes for clients.

## Example Prompts

Trigger this tool by asking:
- "Check for breaking changes between `v1.json` and `v2.json`."
- "Compare the old and new/current API specs."
- "Did we introduce any backward-incompatible changes in the latest update?"
