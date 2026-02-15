# Integration Orchestrator (`pkg/core/tools/integration_orchestrator`)

The Integration Orchestrator Module manages complex, multi-step API workflows.

## Key Tool: `orchestrate_integration`

This tool executes sequences of dependent API calls, passing data (like IDs or tokens) from one step to the next.

### Features

- **State Management**: captures variables from responses (e.g., `userId` from a create response) and uses them in subsequent requests.
- **Workflow Definitions**: Supports defining complex user journeys (e.g., Register -> Login -> Create Order -> Check History).
- **Assertions**: Validates the success of the entire chain, not just individual requests.

## Usage

Use this tool when you need to verify a complete business process rather than a single endpoint.

## Example Prompts

Trigger this tool by asking:
- "Test the full user registration and login flow."
- "Run an integration test for the checkout process (add item -> cart -> pay)."
- "Verify that a user can be created, updated, and then deleted."
