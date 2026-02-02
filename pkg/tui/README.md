# pkg/tui

This package implements ZAP's terminal user interface using the [Charm](https://charm.sh/) ecosystem, specifically Bubble Tea for the application framework.

## Package Overview

```
pkg/tui/
├── app.go         # Entry point: Run() creates program and starts event loop
├── model.go       # Model struct containing all UI state
├── init.go        # Initialization: creates model, registers tools, configures LLM
├── update.go      # Event handling: keyboard, agent events, window resize
├── view.go        # Rendering: viewport, input area, footer, log formatting
├── keys.go        # Keyboard handling: shortcuts and bindings
├── styles.go      # Visual styling: colors, prefixes, spacing
├── highlight.go   # JSON syntax highlighting utility
└── setup/         # Setup wizard components
```

## Architecture

ZAP's TUI follows the Elm Architecture (Model-View-Update):

```
┌─────────────────────────────────────────────┐
│                    Run()                     │
│         Creates Bubble Tea program           │
└────────────────────┬────────────────────────┘
                     │
         ┌───────────┼───────────┐
         │           │           │
         ▼           ▼           ▼
    ┌─────────┐ ┌─────────┐ ┌─────────┐
    │  Init   │ │ Update  │ │  View   │
    │         │ │         │ │         │
    │ model.go│ │update.go│ │ view.go │
    │ init.go │ │ keys.go │ │styles.go│
    └─────────┘ └─────────┘ └─────────┘
```

## Core Components

### Model (model.go)

The Model struct holds all UI state:

```go
type Model struct {
    // UI Components
    viewport  viewport.Model    // Scrollable message area
    textinput textinput.Model   // User input field
    spinner   spinner.Model     // Loading animation

    // State
    logs         []logEntry       // Message history
    status       string           // "idle", "thinking", "streaming", "tool:name"
    inputHistory []string         // Command history for navigation
    historyIndex int              // Current position in history

    // Agent
    agent           *core.Agent   // LLM agent instance
    agentRunning    bool          // Is agent currently processing
    stopRequested   bool          // User requested stop (Esc)

    // Confirmation Mode
    confirmationMode bool               // File write approval mode
    pendingConfirm   *core.FileConfirmation  // Pending file change
    confirmViewport  viewport.Model     // Diff display viewport

    // Display
    width   int                  // Terminal width
    height  int                  // Terminal height
    ready   bool                 // Viewport initialized
}
```

### Log Entry Types

```go
type logEntry struct {
    Type    string  // Entry type (see below)
    Content string  // Text content
}
```

| Type | Description | Display |
|------|-------------|---------|
| `user` | User input | Blue `>` prefix |
| `thinking` | Agent reasoning | Hidden (not shown) |
| `streaming` | Partial response | Appended to last entry |
| `tool` | Tool execution | Dimmed with `○` prefix |
| `observation` | Tool result | Dimmed result text |
| `response` | Final answer | Markdown rendered |
| `error` | Error message | Red `✗` prefix |
| `separator` | Visual break | Hidden |

### Initialization (init.go)

The `Init` function sets up:

1. Load configuration from `.zap/config.json`
2. Run setup wizard if needed
3. Create LLM client (Ollama or Gemini)
4. Create agent with tools
5. Configure tool limits
6. Initialize UI components

```go
func initialModel() Model {
    // Load config
    config := core.LoadOrCreateConfig()

    // Create LLM client
    var llmClient llm.LLMClient
    switch config.Provider {
    case "ollama":
        llmClient = llm.NewOllamaClient(...)
    case "gemini":
        llmClient = llm.NewGeminiClient(...)
    }

    // Create agent
    agent := core.NewAgent(llmClient)
    agent.SetFramework(config.Framework)

    // Register tools
    responseManager := tools.NewResponseManager()
    confirmManager := tools.NewConfirmationManager()

    agent.RegisterTool(tools.NewHTTPTool(responseManager))
    agent.RegisterTool(tools.NewReadFileTool())
    agent.RegisterTool(tools.NewWriteFileTool(confirmManager))
    // ... more tools

    // Configure limits
    for name, limit := range config.ToolLimits.PerTool {
        agent.SetToolLimit(name, limit)
    }

    return Model{
        agent:     agent,
        viewport:  viewport.New(80, 20),
        textinput: textinput.New(),
        spinner:   spinner.New(),
        // ...
    }
}
```

## Event Handling (update.go)

### Message Types

```go
// Agent completed
type agentDoneMsg struct {
    response string
    err      error
}

// Agent event (streaming, tool call, etc.)
type agentEventMsg core.AgentEvent

// Window resized
type tea.WindowSizeMsg struct {
    Width  int
    Height int
}
```

### Update Function

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {

    case tea.KeyMsg:
        return handleKeyPress(m, msg)

    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        m.viewport.Width = msg.Width
        m.viewport.Height = msg.Height - 4  // Reserve for input/footer
        return m, nil

    case agentEventMsg:
        return handleAgentEvent(m, core.AgentEvent(msg))

    case agentDoneMsg:
        m.agentRunning = false
        m.status = "idle"
        if msg.err != nil {
            m.logs = append(m.logs, logEntry{Type: "error", Content: msg.err.Error()})
        }
        return m, nil

    case spinner.TickMsg:
        if m.agentRunning {
            m.spinner, _ = m.spinner.Update(msg)
        }
        return m, nil
    }

    return m, nil
}
```

### Agent Event Handling

```go
func handleAgentEvent(m Model, event core.AgentEvent) (Model, tea.Cmd) {
    switch event.Type {
    case "streaming":
        // Append to current response
        if len(m.logs) > 0 && m.logs[len(m.logs)-1].Type == "streaming" {
            m.logs[len(m.logs)-1].Content += event.Content
        } else {
            m.logs = append(m.logs, logEntry{Type: "streaming", Content: event.Content})
        }

    case "tool_call":
        m.status = "tool:" + event.Content
        m.logs = append(m.logs, logEntry{
            Type:    "tool",
            Content: fmt.Sprintf("%s(%s)", event.Content, event.ToolArgs),
        })

    case "observation":
        m.logs = append(m.logs, logEntry{Type: "observation", Content: event.Content})

    case "answer":
        m.logs = append(m.logs, logEntry{Type: "response", Content: event.Content})

    case "confirmation_required":
        m.confirmationMode = true
        m.pendingConfirm = event.FileConfirmation
        // Setup diff viewport
    }

    updateViewportContent(&m)
    return m, nil
}
```

## Keyboard Handling (keys.go)

### Normal Mode

| Key | Handler |
|-----|---------|
| `Enter` | `handleEnter()` - Send message |
| `Shift+↑` | `handleHistoryUp()` - Previous command |
| `Shift+↓` | `handleHistoryDown()` - Next command |
| `PgUp` | Scroll viewport up |
| `PgDown` | Scroll viewport down |
| `Ctrl+L` | `handleClearScreen()` - Clear logs |
| `Ctrl+U` | `handleClearInput()` - Clear input line |
| `Ctrl+Y` | `handleCopyResponse()` - Copy to clipboard |
| `Esc` | Stop agent or quit |
| `Ctrl+C` | Quit |

### Confirmation Mode

| Key | Handler |
|-----|---------|
| `Y` | Approve file change |
| `N` | Reject file change |
| `PgUp/PgDown` | Scroll diff |
| `Esc` | Reject and continue |

### Implementation

```go
func handleKeyPress(m Model, msg tea.KeyMsg) (Model, tea.Cmd) {
    if m.confirmationMode {
        return handleConfirmationKey(m, msg)
    }

    switch {
    case key.Matches(msg, keys.Enter):
        return handleEnter(m)

    case key.Matches(msg, keys.HistoryUp):
        return handleHistoryUp(m)

    case key.Matches(msg, keys.HistoryDown):
        return handleHistoryDown(m)

    case key.Matches(msg, keys.ClearScreen):
        return handleClearScreen(m)

    case key.Matches(msg, keys.CopyResponse):
        return handleCopyResponse(m)

    case key.Matches(msg, keys.Quit):
        if m.agentRunning {
            m.stopRequested = true
            return m, nil
        }
        return m, tea.Quit
    }

    // Pass to text input
    m.textinput, _ = m.textinput.Update(msg)
    return m, nil
}
```

## Rendering (view.go)

### Main View

```go
func (m Model) View() string {
    if !m.ready {
        return "Initializing..."
    }

    if m.confirmationMode {
        return renderConfirmationView(m)
    }

    return lipgloss.JoinVertical(
        lipgloss.Left,
        renderViewport(m),
        renderInputArea(m),
        renderFooter(m),
    )
}
```

### Viewport Content

```go
func updateViewportContent(m *Model) {
    var lines []string

    for _, entry := range m.logs {
        switch entry.Type {
        case "user":
            lines = append(lines, styles.UserPrefix+entry.Content)

        case "tool":
            lines = append(lines, styles.ToolPrefix+styles.Dim(entry.Content))

        case "response":
            rendered, _ := glamour.Render(entry.Content, "dark")
            lines = append(lines, rendered)

        case "error":
            lines = append(lines, styles.ErrorPrefix+styles.Error(entry.Content))

        // ... other types
        }
    }

    m.viewport.SetContent(strings.Join(lines, "\n"))
    m.viewport.GotoBottom()
}
```

### Input Area

```go
func renderInputArea(m Model) string {
    prompt := "> "
    if m.agentRunning {
        prompt = m.spinner.View() + " "
    }

    return lipgloss.NewStyle().
        BorderStyle(lipgloss.NormalBorder()).
        BorderTop(true).
        Render(prompt + m.textinput.View())
}
```

### Footer / Status Line

```go
func renderFooter(m Model) string {
    status := m.status
    if m.agentRunning {
        status = styles.Accent(status)
    }

    help := "enter:send  shift+↑↓:history  ctrl+y:copy  esc:quit"

    return lipgloss.JoinHorizontal(
        lipgloss.Top,
        status,
        strings.Repeat(" ", m.width-len(status)-len(help)),
        styles.Dim(help),
    )
}
```

## Styling (styles.go)

### Color Palette

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

### Prefixes

```go
var (
    UserPrefix   = Accent.Render("> ")
    ToolPrefix   = Tool.Render("○ ")
    ErrorPrefix  = Error.Render("✗ ")
    ResultPrefix = Dim.Render("→ ")
)
```

## Agent Integration

### Running the Agent

```go
func runAgentAsync(m Model) tea.Cmd {
    return func() tea.Msg {
        input := m.textinput.Value()

        response, err := m.agent.ProcessMessageWithEvents(input, func(event core.AgentEvent) {
            // Send events to TUI
            globalProgram.Send(agentEventMsg(event))
        })

        return agentDoneMsg{response: response, err: err}
    }
}
```

### Global Program Reference

For sending events from the agent callback:

```go
var globalProgram *tea.Program

func Run() error {
    m := initialModel()
    globalProgram = tea.NewProgram(m, tea.WithAltScreen())
    _, err := globalProgram.Run()
    return err
}
```

## Setup Wizard (setup/)

The setup wizard uses [Huh](https://github.com/charmbracelet/huh) for forms:

```go
func RunSetupWizard() *core.Config {
    var provider string
    var ollamaURL string
    var apiKey string
    var framework string

    form := huh.NewForm(
        huh.NewGroup(
            huh.NewSelect[string]().
                Title("Select LLM Provider").
                Options(
                    huh.NewOption("Ollama (local)", "ollama-local"),
                    huh.NewOption("Ollama (cloud)", "ollama-cloud"),
                    huh.NewOption("Gemini", "gemini"),
                ).
                Value(&provider),
        ),
        // More groups for URL, API key, framework...
    )

    form.Run()

    return &core.Config{
        Provider:  provider,
        Framework: framework,
        // ...
    }
}
```

## Testing

For testing TUI components:

```go
func TestModel_Update(t *testing.T) {
    m := initialModel()

    // Simulate window resize
    m, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
    if m.width != 100 {
        t.Error("width not updated")
    }

    // Simulate key press
    m, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlL})
    if len(m.logs) != 0 {
        t.Error("logs not cleared")
    }
}
```

Run tests:

```bash
go test ./pkg/tui/...
```

## Performance Considerations

1. **Viewport updates** - Only update when logs change
2. **Markdown rendering** - Cache rendered content
3. **Streaming** - Batch updates during rapid streaming
4. **Large outputs** - Truncate very long responses

```go
const maxLogEntries = 1000

func addLog(m *Model, entry logEntry) {
    m.logs = append(m.logs, entry)
    if len(m.logs) > maxLogEntries {
        m.logs = m.logs[len(m.logs)-maxLogEntries:]
    }
}
```
