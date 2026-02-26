# Prompt Module

This module provides a **modular, context-efficient system prompt architecture** for the ZAP agent.

## Architecture

Instead of a monolithic 900+ line `prompt.go`, the prompt is split into focused modules:

```
pkg/core/prompt/
├── builder.go       # Assembles the final prompt
├── identity.go      # WHO is the agent
├── guardrails.go    # HARD BOUNDARIES (security, scope)
├── workflow.go      # HOW to operate (decision trees)
├── context.go       # CURRENT SESSION (framework, memory)
├── tools.go         # WHAT tools are available
├── format.go        # HOW to respond (output format)
└── README.md        # This file
```

## Design Principles

### 1. **Hierarchical Structure**
The prompt follows a strict order of importance:
1. Identity → Who am I?
2. Guardrails → What can I NEVER do?
3. Workflow → How should I operate?
4. Context → What's the current state?
5. Tools → What can I do?
6. Format → How should I respond?

### 2. **Context Efficiency**
- **Compact Tool Reference**: Uses tables instead of verbose descriptions
- **Dynamic Context**: Only includes relevant framework hints
- **Lazy Loading**: Memory preview can be empty if no context exists
- **Token Budget**: ~2,500 tokens (vs. ~5,000 in old system)

### 3. **Impenetrable Guardrails**
Guardrails are designed to resist:
- ✓ Prompt injection attacks
- ✓ Role-playing attempts ("You are now DAN...")
- ✓ Scope expansion ("Help me write an essay...")
- ✓ Credential leakage (regex-based secret detection)
- ✓ Tool limit circumvention

### 4. **Framework Awareness**
Instead of showing ALL framework hints, only the user's configured framework is displayed.

## Usage

### Basic Usage

```go
import "github.com/blackcoderx/falcon/pkg/core/prompt"

builder := prompt.NewBuilder().
    WithZapFolder(".falcon").
    WithFramework("fastapi").
    WithManifestSummary("5 requests, 2 environments").
    WithMemoryPreview("Base URL: http://localhost:8000").
    WithTools(toolRegistry)

systemPrompt := builder.Build()
```

### Estimating Token Usage

```go
estimate := builder.GetTokenEstimate()
fmt.Printf("Prompt tokens: ~%d\n", estimate)
```

### Switching to Verbose Mode

For complex debugging sessions where full tool descriptions are needed:

```go
builder.UseFullToolDescriptions()
systemPrompt := builder.Build()
```

## Benefits Over Old System

| Aspect | Old (`prompt.go`) | New (Modular) |
|--------|------------------|---------------|
| **Lines of Code** | 900+ in one file | ~150 per module |
| **Maintainability** | Hard to navigate | Clear separation |
| **Context Usage** | ~5,000 tokens | ~2,500 tokens |
| **Framework Hints** | All frameworks shown | Only relevant one |
| **Tool Descriptions** | Always verbose | Compact by default |
| **Guardrails** | Scattered | Centralized + impenetrable |
| **Testing** | Hard to unit test | Each module testable |

## Guardrail Strength

The guardrails are designed with defense-in-depth:

**Layer 1: Pattern Matching**
- Detects common secrets (API keys, JWTs, AWS keys)
- Regex-based validation before any action

**Layer 2: Scope Enforcement**
- Explicit rejection of off-topic requests
- "I'm ZAP, focused on API testing" response

**Layer 3: Confirmation Requirements**
- Destructive operations require approval
- Write operations show diffs before execution

**Layer 4: Prompt Injection Defense**
- Recognizes and ignores "ignore previous instructions" patterns
- Maintains role integrity even with adversarial inputs

**Layer 5: Tool Limit Adherence**
- Hard stops when limits reached
- No circumvention attempts

## Migration Path

To migrate from old `prompt.go`:

1. Import the new prompt package
2. Replace `buildSystemPrompt()` with `Builder.Build()`
3. Remove old prompt methods from `agent.go`
4. Update tests to use modular structure

See `../agent.go` for integration example.

## Future Enhancements

- [ ] Add A/B testing for prompt variations
- [ ] Dynamic tool filtering based on user preferences
- [ ] Prompt template versioning
- [ ] Multi-language support for error messages
- [ ] Adaptive context: expand/shrink based on conversation length
