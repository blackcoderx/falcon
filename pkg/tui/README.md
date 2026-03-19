# pkg/tui

This package implements Falcon's terminal user interface using the [Charm](https://charm.sh/) ecosystem — Bubble Tea for the application framework, Lip Gloss for styling, and Glamour for markdown rendering.

## Package Overview

```
pkg/tui/
├── app.go          # Entry point: Run() creates the Bubble Tea program
├── model.go        # Model struct — all UI state
├── init.go         # InitialModel(): LLM client, agent, tools, confirmation manager
├── update.go       # Bubble Tea Update() — handles all tea.Msg types
├── view.go         # Bubble Tea View() — renders the full TUI layout
├── keys.go         # Keyboard bindings and input history navigation
├── modelpicker.go  # In-session model switcher UI (/model command)
├── envpicker.go    # In-session environment switcher UI (/env command)
├── slash.go        # Slash command processor
├── styles.go       # Lip Gloss color palette and style definitions
└── highlight.go    # JSON syntax highlighting utility
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

---

## Model

The `Model` struct holds all UI state:

```go
type Model struct {
    // UI components
    viewport  viewport.Model    // Scrollable output area
    textinput textinput.Model   // User input field
    spinner   spinner.Model     // Loading animation (harmonica spring)

    // Message log
    logs         []logEntry     // Message history displayed in viewport
    status       string         // "idle", "thinking", "streaming", "tool:name"
    inputHistory []string       // Previous commands for Shift+↑/↓ navigation
    historyIdx   int            // Current position in input history
    savedInput   string         // Saved input when navigating history

    // Agent
    agent          *core.Agent  // The LLM agent instance
    modelName      string       // Current LLM model name for badge display
    streamingBuffer string      // Accumulates streaming content

    // Tool usage display
    toolUsage     []ToolUsageDisplay // Per-tool call stats
    totalCalls    int                // Total tool calls in session
    totalLimit    int                // Total limit cap
    lastToolName  string
    toolStartTime time.Time

    // File write confirmation
    confirmationMode    bool
    pendingConfirmation *core.FileConfirmation
    confirmManager      *shared.ConfirmationManager

    // Slash command state
    slashState SlashState

    // Model picker (/model command)
    modelPickerActive bool
    modelPickerItems  []modelEntry  // Configured providers from GlobalConfig
    modelPickerIdx    int

    // Environment picker (/env command)
    envPickerActive bool
    envPickerItems  []string  // Names from .falcon/environments/
    envPickerIdx    int

    // Active environment
    activeEnv string
    envVars   map[string]string

    // Layout
    width  int
    height int
    ready  bool
    webPort int  // Web dashboard port (0 if disabled)
}
```

### Log Entry Types

| Type | Description | Display |
|------|-------------|---------|
| `user` | User input | Blue `>` prefix |
| `thinking` | Agent reasoning | Hidden |
| `streaming` | Partial LLM response | Appended live to current entry |
| `tool` | Tool invocation | Green `○` prefix with args and usage count |
| `observation` | Tool result | Dimmed `→` prefix |
| `response` | Final answer | Glamour-rendered markdown |
| `error` | Error message | Red `✗` prefix |
| `splash` | Startup Falcon ASCII art | Indented brand art |
| `separator` | Visual break | Hidden |

---

## Initialization

`InitialModel()` in `init.go`:

1. Loads global config from `~/.falcon/config.yaml`
2. Runs the setup wizard (Huh forms) if no provider is configured
3. Builds the LLM client via the provider registry
4. Creates the `Agent` and registers all 28+ tools via the central `Registry`
5. Applies tool limits from `.falcon/config.yaml`
6. Initializes UI components (viewport, textinput, spinner)
7. Displays the Falcon ASCII splash screen with version, working directory, and web UI URL

---

## Event Handling

### Keyboard — Normal Mode

| Key | Action |
|-----|--------|
| `Enter` | Send message to agent |
| `Shift+↑` | Navigate to previous command in history |
| `Shift+↓` | Navigate to next command in history |
| `PgUp / ↑` | Scroll output up |
| `PgDown / ↓` | Scroll output down |
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
| `PgUp / PgDown` | Scroll the unified diff |
| `Esc` | Reject and continue |

### Keyboard — Model Picker

| Key | Action |
|-----|--------|
| `↑ / ↓` | Move selection |
| `Enter` | Select model and switch LLM client |
| `Esc` | Cancel and close picker |

### Keyboard — Environment Picker

| Key | Action |
|-----|--------|
| `↑ / ↓` | Move selection |
| `Enter` | Load environment and apply variables |
| `Esc` | Cancel and close picker |

### Agent Events

The agent emits `AgentEvent` values via callback. The TUI handles them in `update.go`:

| Event Type | TUI Action |
|------------|------------|
| `streaming` | Append chunk to current log entry (real-time display) |
| `tool_call` | Add tool log entry, update status line |
| `observation` | Add dimmed observation entry with duration |
| `tool_usage` | Update per-tool call counters |
| `answer` | Render final answer as Glamour markdown |
| `error` | Add red error entry |
| `confirmation_required` | Enter confirmation mode, show diff viewport |

---

## Slash Commands

Processed by `slash.go` before the input is sent to the agent:

| Command | Action |
|---------|--------|
| `/model` | Open the model picker panel |
| `/env` | Open the environment picker panel |
| `/flow <file>` | Load and execute a YAML workflow file |

---

## Model Picker

Implemented in `modelpicker.go`. Activated by typing `/model`.

- Reads configured providers from `~/.falcon/config.yaml`
- Displays a list of `modelEntry` items: `{ProviderID, DisplayName, Model, Config}`
- On selection, calls `BuildClient()` on the provider and hot-swaps the agent's LLM client
- The `modelName` badge in the footer updates immediately

---

## Environment Picker

Implemented in `envpicker.go`. Activated by typing `/env`.

- Reads `.yaml` files from `.falcon/environments/`
- On selection, loads the environment and updates `envVars` in the model
- Variable substitution in tool calls will use the newly active environment

---

## Rendering

### Layout

```
┌──────────────────────────────────┐
│          Viewport                │  ← Scrollable output (logs)
│                                  │
├──────────────────────────────────┤
│  > input field         [spinner] │  ← Text input
├──────────────────────────────────┤
│  model · env · calls   help text │  ← Footer (status badges)
└──────────────────────────────────┘
```

When a model picker or env picker is active, an overlay panel renders above the input line.

When in confirmation mode, a diff viewport replaces the input area.

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
UserPrefix   = Accent.Render("> ")
ToolPrefix   = Tool.Render("○ ")
ErrorPrefix  = Error.Render("✗ ")
ResultPrefix = Dim.Render("→ ")
```

---

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

---

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

---

## Performance Notes

- Viewport content is only re-rendered when `logs` changes
- Streaming chunks are appended directly to the last log entry — no full re-render per chunk
- Log entries are capped at 1000 to prevent memory growth in long sessions
- Markdown rendering (Glamour) is done once per `response` entry, not on every frame
