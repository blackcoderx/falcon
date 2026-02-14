# Dependency Mapper Module

## Overview
The Dependency Mapper build a logical graph of how API endpoints relate to each other based on the flow of data and resources.

## Tools
- `map_dependencies`: Analyzes the Knowledge Graph to find resource links.

## Features
- **Identifier Flow**: Automatically maps `POST` results to `GET/{id}` requirements.
- **Dependency Chains**: Visualizes multi-step processes (e.g. Create -> Approve -> List -> Delete).
- **Resource Discovery**: Identifies shared keys across different API modules.

## Usage
```json
{
  "focus": "resources"
}
```
