# API Drift Analyzer Module

## Overview
The API Drift Analyzer detects the "drift" between what an API is *intended* to do (the spec) and what it *actually* does in production.

## Tools
- `analyze_drift`: Compares live traffic/behavior against the specification.

## Features
- **Shadow Endpoint Detection**: Identifies APIs implemented in code but missing from the documentation/spec.
- **Behavioral Drift**: Detects changes in response formats or status codes that aren't reflected in the graph.
- **Live Sync**: Provides insights into the "truthiness" of your API descriptors.

## Usage
```json
{
  "base_url": "https://api.myapp.com"
}
```
