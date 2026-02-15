# Dependency Mapper (`pkg/core/tools/dependency_mapper`)

The Dependency Mapper understands the logical relationships between your API resources.

## Key Tool: `map_dependencies`

This tool analyzes traffic and specifications to understand resource lifecycles.

### Features

- **Relationship Discovery**: Identifies that `POST /users` returns an ID required by `GET /users/{id}`.
- **Order of Operations**: Helps the agent understand that it must create a user *before* it can update their profile.
- **Graph Visualization**: Outputs a dependency graph of your API surface.

## Usage

Fundamental for generating valid integration tests and complex workflows.

## Example Prompts

Trigger this tool by asking:
- "Map the dependencies between the API endpoints."
- "Identify which resources need to be created before testing the orders API."
- "Show me the relationship graph for the user module."
