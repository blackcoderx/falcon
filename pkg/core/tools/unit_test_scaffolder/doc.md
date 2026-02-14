# Unit Test Scaffolder Module

## Overview
The Unit Test Scaffolder analyzes your source code to identify business logic and automatically generates test boilerplates and mocks.

## Tools
- `scaffold_unit_tests`: Scans directories and generates test files.

## Features
- **Code Analysis**: Identifies controllers, services, repositories, and handlers.
- **Mock Generation**: Creates mock structures for dependencies.
- **Multi-Language**: Supports Go, TypeScript, and Python.

## Usage
```json
{
  "source_dir": "./pkg/logic",
  "language": "go",
  "output_dir": "./tests/unit"
}
```
