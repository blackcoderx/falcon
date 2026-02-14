# Functional Test Generator Module

## Overview

The Functional Test Generator module provides comprehensive, spec-driven test generation capabilities for APIs. It reads from the API Knowledge Graph (created by the Spec Ingester) and automatically generates test scenarios using multiple testing strategies.

## Purpose

This module replaces the older `generate_tests` tool with a more sophisticated, spec-aware test generator that:
- Generates tests directly from API specifications (OpenAPI, Swagger, Postman)
- Uses multiple testing strategies (happy path, negative, boundary)
- Executes tests automatically and validates responses
- Exports test scenarios for use in other testing frameworks

## Tool: `generate_functional_tests`

### Description
Generates and optionally executes comprehensive functional tests from the API Knowledge Graph using happy path, negative, and boundary strategies.

### Parameters

```json
{
  "base_url": "http://localhost:3000",
  "strategies": ["happy", "negative", "boundary"],
  "endpoints": ["GET /api/users", "POST /api/users"],
  "execute": true,
  "export": false
}
```

- **base_url** (required): Base URL for the API being tested
- **strategies** (optional): Array of strategies to use. Available: `happy`, `negative`, `boundary`. Default: all strategies
- **endpoints** (optional): Specific endpoints to test. If empty, tests all endpoints from the Knowledge Graph
- **execute** (required): Whether to execute the generated tests immediately
- **export** (optional): Whether to export test scenarios to a JSON file

### Example Usage

```
Generate happy path tests for all endpoints:
{
  "base_url": "http://localhost:3000",
  "strategies": ["happy"],
  "execute": true
}

Generate and export all test types:
{
  "base_url": "http://localhost:3000",
  "execute": false,
  "export": true
}

Generate negative tests for specific endpoints:
{
  "base_url": "http://localhost:3000",
  "strategies": ["negative"],
  "endpoints": ["POST /api/users", "PUT /api/users/{id}"],
  "execute": true
}
```

## Testing Strategies

### 1. Happy Path Strategy
Generates valid test cases with correct data for successful operations.

**What it tests:**
- Valid requests with all required fields
- Correct data types for all parameters
- Semantically meaningful values (e.g., valid emails, phone numbers)
- Expected success status codes (200, 201, 204, etc.)

**Use case:** Verify that the API works correctly with valid inputs.

### 2. Negative Strategy
Generates invalid test cases to ensure proper error handling.

**What it tests:**
- Missing required fields
- Wrong data types (e.g., string instead of number)
- Invalid values (e.g., malformed email, negative numbers where positive expected)
- Expected error status codes (400, 422, etc.)

**Use case:** Verify that the API properly validates inputs and returns appropriate error responses.

### 3. Boundary Strategy
Generates test cases with edge case values.

**What it tests:**
- Empty strings and zero values
- Maximum boundary values (very long strings, max integers)
- Large payloads
- Empty arrays and objects
- Expected responses for edge cases (may be success or error)

**Use case:** Verify that the API handles edge cases gracefully without crashes or unexpected behavior.

## Architecture

### Components

1. **tool.go**: Main tool implementation
   - Orchestrates test generation and execution
   - Loads API Knowledge Graph
   - Manages strategy selection and endpoint filtering
   - Formats results for display

2. **strategies.go**: Strategy implementations
   - `HappyPathStrategy`: Generates valid test data
   - `NegativeStrategy`: Generates invalid/missing data
   - `BoundaryStrategy`: Generates edge case data
   - Each strategy implements the `Strategy` interface

3. **generator.go**: Test execution engine
   - Executes test scenarios using `shared/http.go`
   - Validates responses against expectations
   - Collects and reports results

4. **templates.go**: Export functionality
   - Exports scenarios to JSON files
   - Placeholder for future code generation (Go, Jest, pytest, etc.)

### Data Flow

```
1. User calls generate_functional_tests
   ↓
2. Load API Knowledge Graph (.zap/api_graph.json)
   ↓
3. Filter endpoints (if specific ones requested)
   ↓
4. For each endpoint + strategy combination:
   - Generate test scenarios
   ↓
5. If execute=true:
   - Run each scenario
   - Validate responses
   - Collect results
   ↓
6. If export=true:
   - Export scenarios to .zap/exports/
   ↓
7. Format and return summary
```

## Integration with Other Modules

### Depends On:
- **Spec Ingester**: Reads the API Knowledge Graph created by `ingest_spec`
- **shared/http.go**: Executes HTTP requests
- **shared/assertions.go**: Validates responses (future enhancement)
- **shared/types.go**: Uses `TestScenario`, `TestResult`, `APIKnowledgeGraph` structs

### Used By:
- Future integration orchestrator modules (Sprint 6+)
- Regression testing workflows
- CI/CD pipelines (via export functionality)

## Future Enhancements

1. **Code Generation (templates.go)**
   - Generate executable test files in Go, JavaScript, Python
   - Framework-specific templates (Jest, pytest, etc.)

2. **Advanced Strategies**
   - Security testing strategy
   - Performance testing strategy
   - Fuzz testing strategy

3. **Smart Data Generation**
   - Use schema constraints (min/max length, pattern, enum)
   - Faker integration for realistic test data
   - Reference data from example responses in spec

4. **Test Dependencies**
   - Support for test scenarios that depend on other scenarios
   - Automatic ordering (e.g., POST before GET)

## Best Practices

1. **Always run `ingest_spec` first**: The generator requires an API Knowledge Graph
2. **Start with happy path**: Verify basic functionality before negative testing
3. **Use execute=false first**: Review generated scenarios before executing
4. **Export for CI/CD**: Export scenarios to integrate with existing test frameworks
5. **Combine strategies**: Use all three strategies for comprehensive coverage

## Example Workflow

```
1. Index the API specification:
   ingest_spec({"action": "index", "source": "./docs/openapi.yaml"})

2. Generate and preview tests:
   generate_functional_tests({
     "base_url": "http://localhost:3000",
     "execute": false,
     "export": true
   })

3. Review exported scenarios in .zap/exports/

4. Execute tests:
   generate_functional_tests({
     "base_url": "http://localhost:3000",
     "execute": true
   })

5. Fix any failures, then re-run
```

## Files

- `tool.go`: Main tool implementation (245 lines)
- `strategies.go`: Strategy implementations (415 lines)
- `generator.go`: Test execution engine (145 lines)
- `templates.go`: Export functionality (62 lines)

**Total:** ~867 lines of code
