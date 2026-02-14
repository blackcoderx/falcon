# Data-Driven Engine Module

## Overview
The Data-Driven Engine allows you to run a single test scenario hundreds of times using data from external files or generators.

## Tools
- `run_data_driven`: Maps CSV/JSON data to test templates.

## Features
- **Variable Interpolation**: Uses `{{var}}` syntax in URL, body, and headers.
- **Multi-Source**: Load data from CSV, JSON, or use the "fake" data generator.
- **Bulk Execution**: Processes many rows in a single tool call.

## Usage
```json
{
  "scenario": {
    "method": "POST",
    "url": "http://localhost:3000/api/users",
    "body": {"name": "{{name}}", "email": "{{email}}"}
  },
  "data_source": "fake",
  "variables": ["name", "email"],
  "max_rows": 5
}
```
