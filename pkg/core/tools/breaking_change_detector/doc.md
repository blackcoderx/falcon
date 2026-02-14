# Breaking Change Detector Module

## Overview
The Breaking Change Detector compares two versions of an API specification and identifies modifications that would break existing clients.

## Tools
- `detect_breaking_changes`: Compares two spec files (old vs new).

## Features
- **Endpoint Analysis**: Detects removed or renamed endpoints.
- **Contract Enforcement**: Identifies new required parameters or changed data types.
- **Categorization**: Groups changes into 'Breaking', 'Minor', and 'Patch'.

## Usage
```json
{
  "old_spec_path": "./v1/swagger.json",
  "new_spec_path": "./v2/swagger.json"
}
```
