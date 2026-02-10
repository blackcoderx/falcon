# Package core

The `core` package provides the central agent logic, tool management, and ReAct loop implementation for the ZAP API debugging assistant.

## Overview

This package implements the AI agent that powers ZAP. It uses a ReAct (Reason+Act) pattern where the agent:

1. **Receives** a user message
2. **Reasons** about the task using the LLM
3. **Acts** by selecting and executing tools as needed
4. **Observes** the results of tool execution
5. **Continues** the cycle until a final answer is reached

## Key Components

### Autonomous Testing Workflow (`react.go`, `prompt.go`)
ZAP uses a specialized ReAct loop combined with AI-driven testing tools:
1. **Analyze**: Understand endpoint requirements via `analyze_endpoint`.
2. **Generate**: Create thousands of tests covering security, validation, and edge cases.
3. **Execute**: Run parallelized tests with `run_tests`.
4. **Fix**: Locate handlers with `find_handler` and propose secure code changes with `propose_fix`.
5. **Report**: Aggregate findings into professional security reports.

### Tool Interface (`types.go`, `test_types.go`)
Standard tool interface coupled with core testing data structures:
- `TestScenario`: Defines a single test case (HTTP details + expectations).
- `TestResult`: Captures execution details, duration, and failure analysis.

### System Prompts (`prompt.go`)
Constructs dynamic instructions with:
- **Auto-Test Workflows**: Step-by-step guidance for autonomous testing.
- **Code Fixing Workflows**: Guided remediation patterns.
- **Reporting Workflows**: Professional assessment generation.

## Event System

The agent emits events during processing for real-time UI updates:

| Event Type | Description |
|------------|-------------|
| `thinking` | Agent is reasoning |
| `tool_call` | Executing a tool (e.g., `auto_test`) |
| `observation` | Tool returned a result |
| `answer` | Final answer ready |
| `streaming` | LLM response chunk |
| `confirmation_required` | File write or code fix needs approval |

## File Structure

```
pkg/core/
├── doc.md          # This file
├── types.go        # Core interfaces
├── agent.go        # Agent and tool management
├── react.go        # ReAct loop implementation
├── prompt.go       # Autonomous workflow prompts
├── memory.go       # Persistent memory store
├── init.go         # Initialization and config
└── tools/          # Tool implementations
    ├── test_types.go    # Core testing data structures
    ├── analyze.go       # AI endpoint & failure analysis
    ├── generate.go      # AI test generation
    ├── orchestrate.go   # Test parallelization & auto-flow
    ├── handler.go       # Codebase handler discovery
    ├── fix.go           # AI code fix generation
    ├── test_gen.go      # Regression test generation
    └── report.go        # Security scoring & reporting
```
