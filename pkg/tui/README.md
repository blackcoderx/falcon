# pkg/tui

This package implements Falcon's terminal user interface using the [Charm](https://charm.sh/) ecosystem — Bubble Tea for the application framework, Lip Gloss for styling, and Glamour for markdown rendering.

## Package Overview

```
pkg/tui/
├── app.go        # Entry point: Run() creates the Bubble Tea program
├── model.go      # Model struct — all UI state
├── init.go       # InitialModel(): LLM client, agent, tools, confirmation manager
├── update.go     # Bubble Tea Update() — handles all tea.Msg types
├── view.go       # Bubble Tea View() — renders the full TUI layout
├── keys.go       # Keyboard bindings and input history navigation
├── styles.go     # Lip Gloss color palette and style definitions
└── highlight.go  # JSON syntax highlighting utility
```

## Architecture

Falcon's TUI follows the Elm Architecture (Model-View-Update):

```
┌──────────────────────────────────┐
│             Run()                │
│   Creates Bubble Tea program     │
└──────────┬───────────────────────┘
           │
    ┌──────┼──────┐
    ▼      ▼      ▼
  Init  Update  View
```

- **Init** (`init.go`, `model.go`) — builds the initial model
- **Update** (`update.go`, `keys.go`) — handles events and produces new model state
- **View** (`view.go`, `styles.go`) — renders the model to a string each frame

## Model

The `Model` struct holds all UI state:

```go
type Model struct {
    // UI components
    viewport  viewport.Model    // Scrollable output area
    textinput textinput.Model   // User input field
    spinner   spinner.Model     // Loading animation (harmonica spring)

    // State
    logs         []logEntry     // Message history displayed in viewport
    status       string         // "idle", "thinking", "streaming", "tool:name"
    inputHistory []string       // Previous commands for Shift+↑/↓ navigation
    historyIndex int            // Current position in input history

    // Agent
    agent        *core.Agent   // The LLM agent instance
    agentRunning bool          // True while agent is processing
    stopRequested bool         // Set to true when user presses Esc

    // File write confirmation
    confirmationMode bool                  // True when awaiting Y/N approval
    pendingConfirm   *core.FileConfirmation // Details of the pending file change
    confirmViewport  viewport.Model        // Viewport for scrolling the diff

    // Display
    width  int  // Terminal width
    height int  // Terminal height
    ready  bool // True once viewport is initialized

    // Web UI
    webPort int  // Port the web dashboard is listening on (0 if disabled)
}
```

### Log Entry Types

| Type | Description | Display |
|------|-------------|---------|
| `user` | User input | Blue `>` prefix |
| `thinking` | Agent reasoning | Hidden |
| `streaming` | Partial LLM response | Appended live to last entry |
| `tool` | Tool invocation | Dimmed with `○` prefix |
| `observation` | Tool result | Dimmed result text |
| `response` | Final answer | Glamour-rendered markdown |
| `error` | Error message | Red `✗` prefix |
| `splash` | Startup Falcon ASCII art | Indented brand art |
| `separator` | Visual break | Hidden |

## Initialization

`InitialModel()` in `init.go`:

1. Loads `config.yaml` from `.zap/`
2. Runs the setup wizard (Huh forms) if no config exists
3. Creates the LLM client (Ollama or Gemini)
4. Creates the `Agent` and registers all 40+ tools via the central `Registry`
5. Applies tool limits from config
6. Initializes UI components (viewport, textinput, spinner)
7. Displays the Falcon ASCII splash screen with version, working directory, and web UI URL

## Event Handling

### Keyboard — Normal Mode

| Key | Action |
|-----|--------|
| `Enter` | Send message to agent |
| `Shift+↑` | Navigate to previous command |
| `Shift+↓` | Navigate to next command |
| `PgUp` | Scroll output up |
| `PgDown` | Scroll output down |
| `Ctrl+L` | Clear screen |
| `Ctrl+U` | Clear input line |
| `Ctrl+Y` | Copy last response to clipboard |
| `Esc` | Stop running agent / Quit if idle |
| `Ctrl+C` | Quit |

### Keyboard — Confirmation Mode

When Falcon proposes a file change, the TUI enters confirmation mode:

| Key | Action |
|-----|--------|
| `Y` | Approve — write the file |
| `N` | Reject — discard the change |
| `PgUp/PgDown` | Scroll the unified diff |
| `Esc` | Reject and continue |

### Agent Events

The agent emits `AgentEvent` values via callback. The TUI handles them in `update.go`:

| Event Type | TUI Action |
|------------|------------|
| `streaming` | Append chunk to current log entry (real-time display) |
| `tool_call` | Add tool log entry, update status line |
| `observation` | Add dimmed observation entry |
| `tool_usage` | Update per-tool call counters |
| `answer` | Render final answer as Glamour markdown |
| `error` | Add red error entry |
| `confirmation_required` | Enter confirmation mode, show diff viewport |

## Rendering

### Layout

```
┌──────────────────────────────────┐
│          Viewport                │  ← Scrollable output (logs)
│                                  │
├──────────────────────────────────┤
│  > input field                   │  ← Text input with spinner
├──────────────────────────────────┤
│  status          help text       │  ← Footer
└──────────────────────────────────┘
```

### Styling

Defined in `styles.go` using Lip Gloss:

```go
var (
    Dim     = lipgloss.NewStyle().Foreground(lipgloss.Color("#6c6c6c"))
    Text    = lipgloss.NewStyle().Foreground(lipgloss.Color("#e0e0e0"))
    Accent  = lipgloss.NewStyle().Foreground(lipgloss.Color("#7aa2f7"))
    Error   = lipgloss.NewStyle().Foreground(lipgloss.Color("#f7768e"))
    Tool    = lipgloss.NewStyle().Foreground(lipgloss.Color("#9ece6a"))
    Success = lipgloss.NewStyle().Foreground(lipgloss.Color("#73daca"))
)
```

Log entry prefixes:

```go
UserPrefix  = Accent.Render("> ")
ToolPrefix  = Tool.Render("○ ")
ErrorPrefix = Error.Render("✗ ")
ResultPrefix = Dim.Render("→ ")
```

## Agent Integration

The agent runs in a goroutine and sends events back to the Bubble Tea program via `program.Send()`:

```go
func runAgentAsync(m Model) tea.Cmd {
    return func() tea.Msg {
        err := m.agent.ProcessMessageWithEvents(ctx, input, func(event core.AgentEvent) {
            globalProgram.Send(agentEventMsg(event))
        })
        return agentDoneMsg{err: err}
    }
}
```

A thread-safe `globalProgram` reference allows the agent callback goroutine to safely send messages into the Bubble Tea event loop.

## Testing

```bash
go test ./pkg/tui/...
```

Example:

```go
func TestModel_Update_WindowResize(t *testing.T) {
    m := InitialModel(0)
    updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
    result := updated.(Model)
    if result.width != 120 {
        t.Error("width not updated")
    }
}
```

## Performance Notes

- Viewport content is only re-rendered when `logs` changes
- Streaming chunks are appended directly to the last log entry (no full re-render per chunk)
- Log entries are capped at 1000 to prevent memory growth in long sessions
- Markdown rendering (Glamour) is done once per `response` entry, not on every frame
