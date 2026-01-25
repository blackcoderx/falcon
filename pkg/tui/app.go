package tui

import (
	"os"
	"strings"
	"sync"

	"github.com/atotto/clipboard"
	"github.com/blackcoderx/zap/pkg/core"
	"github.com/blackcoderx/zap/pkg/core/tools"
	"github.com/blackcoderx/zap/pkg/llm"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"
)

// logEntry represents a single log line in the UI
type logEntry struct {
	Type    string // "user", "thinking", "tool", "observation", "response", "error", "separator", "streaming"
	Content string
}

// model is the Bubble Tea model
type model struct {
	viewport        viewport.Model
	textinput       textinput.Model
	spinner         spinner.Model
	logs            []logEntry
	thinking        bool
	width           int
	height          int
	agent           *core.Agent
	ready           bool
	renderer        *glamour.TermRenderer
	inputHistory    []string // history of user inputs
	historyIdx      int      // current position in history (-1 = new input)
	savedInput      string   // saved input when navigating history
	status          string   // current status: "idle", "thinking", "tool:name", "streaming"
	currentTool     string   // name of tool currently being executed
	streamingBuffer string   // buffer for accumulating streaming content
	modelName       string   // current LLM model name for badge display
}

// agentEventMsg wraps an agent event for the TUI
type agentEventMsg struct {
	event core.AgentEvent
}

// agentDoneMsg signals the agent has finished
type agentDoneMsg struct {
	err error
}

// programRef holds the program reference for sending messages from goroutines.
// Using a struct with mutex for thread-safe access instead of a bare global variable.
type programRef struct {
	mu      sync.RWMutex
	program *tea.Program
}

func (p *programRef) Set(prog *tea.Program) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.program = prog
}

func (p *programRef) Send(msg tea.Msg) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.program != nil {
		p.program.Send(msg)
	}
}

// Global program reference with thread-safe accessors.
// This is still a package-level variable but access is now synchronized.
var globalProgram = &programRef{}

// registerTools adds all tools to the agent.
func registerTools(agent *core.Agent, zapDir, workDir string) {
	// Initialize shared components
	responseManager := tools.NewResponseManager()
	varStore := tools.NewVariableStore(zapDir)

	// Register codebase tools
	httpTool := tools.NewHTTPTool(responseManager, varStore)
	agent.RegisterTool(httpTool)
	agent.RegisterTool(tools.NewReadFileTool(workDir))
	agent.RegisterTool(tools.NewListFilesTool(workDir))
	agent.RegisterTool(tools.NewSearchCodeTool(workDir))

	// Register persistence tools
	persistence := tools.NewPersistenceTool(zapDir)
	agent.RegisterTool(tools.NewSaveRequestTool(persistence))
	agent.RegisterTool(tools.NewLoadRequestTool(persistence))
	agent.RegisterTool(tools.NewListRequestsTool(persistence))
	agent.RegisterTool(tools.NewListEnvironmentsTool(persistence))
	agent.RegisterTool(tools.NewSetEnvironmentTool(persistence))

	// Register Sprint 1 testing tools
	assertTool := tools.NewAssertTool(responseManager)
	extractTool := tools.NewExtractTool(responseManager, varStore)
	agent.RegisterTool(assertTool)
	agent.RegisterTool(extractTool)
	agent.RegisterTool(tools.NewVariableTool(varStore))
	agent.RegisterTool(tools.NewWaitTool())
	agent.RegisterTool(tools.NewRetryTool(agent))

	// Register Sprint 2 tools
	agent.RegisterTool(tools.NewSchemaValidationTool(responseManager))
	agent.RegisterTool(tools.NewAuthBearerTool(varStore))
	agent.RegisterTool(tools.NewAuthBasicTool(varStore))
	agent.RegisterTool(tools.NewAuthHelperTool(responseManager, varStore))
	agent.RegisterTool(tools.NewTestSuiteTool(httpTool, assertTool, extractTool, responseManager, varStore, zapDir))
	agent.RegisterTool(tools.NewCompareResponsesTool(responseManager, zapDir))

	// Register Sprint 3 tools (MVP)
	agent.RegisterTool(tools.NewPerformanceTool(httpTool, varStore))
	agent.RegisterTool(tools.NewWebhookListenerTool(varStore))
	agent.RegisterTool(tools.NewAuthOAuth2Tool(varStore))
}

// newLLMClient creates and configures the LLM client from Viper config.
func newLLMClient() *llm.OllamaClient {
	// Get config from viper
	ollamaURL := viper.GetString("ollama_url")
	if ollamaURL == "" {
		ollamaURL = "https://ollama.com"
	}

	ollamaAPIKey := viper.GetString("ollama_api_key")
	if ollamaAPIKey == "" {
		ollamaAPIKey = viper.GetString("OLLAMA_API_KEY")
	}

	defaultModel := viper.GetString("default_model")
	if defaultModel == "" {
		defaultModel = "llama3"
	}

	return llm.NewOllamaClient(ollamaURL, defaultModel, ollamaAPIKey)
}

func initialModel() model {
	// Get current working directory for codebase tools
	workDir, _ := os.Getwd()

	// Get .zap directory path
	zapDir := core.ZapFolderName

	// Get model name for display
	modelName := viper.GetString("default_model")
	if modelName == "" {
		modelName = "llama3"
	}

	client := newLLMClient()
	agent := core.NewAgent(client)

	registerTools(agent, zapDir, workDir)

	// Create text input (single line, auto-wraps visually)
	ti := textinput.New()
	ti.Placeholder = "Ask me anything..."
	ti.Focus()
	ti.CharLimit = 2000
	ti.Width = 80
	ti.Prompt = PromptStyle.Render(UserPrefix)

	// Create spinner
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(AccentColor)

	// Create glamour renderer for markdown
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)

	return model{
		textinput:       ti,
		spinner:         sp,
		logs:            []logEntry{},
		thinking:        false,
		agent:           agent,
		ready:           false,
		renderer:        renderer,
		inputHistory:    []string{},
		historyIdx:      -1,
		savedInput:      "",
		status:          "idle",
		currentTool:     "",
		streamingBuffer: "",
		modelName:       modelName,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		textinput.Blink,
		m.spinner.Tick,
	)
}

// runAgentAsync starts the agent in a goroutine and sends events via the program
func runAgentAsync(agent *core.Agent, input string) tea.Cmd {
	return func() tea.Msg {
		// Run agent in goroutine so we can send intermediate events
		go func() {
			callback := func(event core.AgentEvent) {
				globalProgram.Send(agentEventMsg{event: event})
			}

			_, err := agent.ProcessMessageWithEvents(input, callback)
			globalProgram.Send(agentDoneMsg{err: err})
		}()

		// Return nil - actual results come via program.Send
		return nil
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "ctrl+l":
			// Clear screen
			m.logs = []logEntry{}
			m.streamingBuffer = ""
			m.updateViewportContent()
			return m, nil
		case "ctrl+y":
			// Copy last response to clipboard
			var lastResponse string
			for i := len(m.logs) - 1; i >= 0; i-- {
				if m.logs[i].Type == "response" {
					lastResponse = m.logs[i].Content
					break
				}
			}
			if lastResponse != "" {
				_ = clipboard.WriteAll(lastResponse)
				// Determine command to flash status?
				// For now, we'll just rely on the user knowing it worked,
				// or maybe we briefly change the status text?
				// Since status is "idle", we can't easily override it without a timer.
				// Let's just do it silently for now or we could add a temporary "copied" state.
			}
			return m, nil
		case "ctrl+u":
			// Clear input
			m.textinput.SetValue("")
			m.historyIdx = -1
			return m, nil
		case "up":
			// Navigate history backwards
			if !m.thinking && len(m.inputHistory) > 0 {
				if m.historyIdx == -1 {
					// Save current input before navigating
					m.savedInput = m.textinput.Value()
					m.historyIdx = len(m.inputHistory) - 1
				} else if m.historyIdx > 0 {
					m.historyIdx--
				}
				m.textinput.SetValue(m.inputHistory[m.historyIdx])
				m.textinput.CursorEnd()
				return m, nil
			}
		case "down":
			// Navigate history forwards
			if !m.thinking && m.historyIdx != -1 {
				if m.historyIdx < len(m.inputHistory)-1 {
					m.historyIdx++
					m.textinput.SetValue(m.inputHistory[m.historyIdx])
				} else {
					// Return to saved input
					m.historyIdx = -1
					m.textinput.SetValue(m.savedInput)
				}
				m.textinput.CursorEnd()
				return m, nil
			}
		case "enter":
			// Send message with enter
			if m.textinput.Value() != "" && !m.thinking {
				userInput := strings.TrimSpace(m.textinput.Value())
				if userInput == "" {
					return m, nil
				}
				// Add separator if there are previous logs
				if len(m.logs) > 0 {
					m.logs = append(m.logs, logEntry{Type: "separator", Content: ""})
				}
				m.logs = append(m.logs, logEntry{Type: "user", Content: userInput})
				// Add to history
				m.inputHistory = append(m.inputHistory, userInput)
				m.historyIdx = -1
				m.savedInput = ""
				m.textinput.SetValue("")
				m.thinking = true
				m.status = "thinking"
				m.streamingBuffer = ""
				m.updateViewportContent()

				return m, tea.Batch(
					m.spinner.Tick,
					runAgentAsync(m.agent, userInput),
				)
			}
		case "pgup", "pgdown", "home", "end":
			// Let viewport handle these for scrolling
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// No header - maximize viewport space
		inputHeight := 1
		footerHeight := 1
		margins := 3

		viewportHeight := m.height - inputHeight - footerHeight - margins
		if viewportHeight < 5 {
			viewportHeight = 5
		}

		if !m.ready {
			m.viewport = viewport.New(m.width-2, viewportHeight)
			m.viewport.SetContent("")
			m.ready = true
		} else {
			m.viewport.Width = m.width - 2
			m.viewport.Height = viewportHeight
		}

		// Account for model badge width in input
		badgeWidth := lipgloss.Width(ModelBadgeStyle.Render(m.modelName))
		m.textinput.Width = m.width - badgeWidth - 10

	case agentEventMsg:
		// Handle agent events
		switch msg.event.Type {
		case "thinking":
			// Clear streaming buffer when starting new thinking
			if m.streamingBuffer != "" {
				m.streamingBuffer = ""
			}
			m.logs = append(m.logs, logEntry{Type: "thinking", Content: msg.event.Content})
			m.status = "thinking"
		case "streaming":
			// Append chunk to streaming buffer and update display
			m.streamingBuffer += msg.event.Content
			m.status = "streaming"
			// Update or add streaming log entry
			if len(m.logs) > 0 && m.logs[len(m.logs)-1].Type == "streaming" {
				m.logs[len(m.logs)-1].Content = m.streamingBuffer
			} else {
				m.logs = append(m.logs, logEntry{Type: "streaming", Content: m.streamingBuffer})
			}
		case "tool_call":
			// Clear streaming when tool is called
			m.streamingBuffer = ""
			m.logs = append(m.logs, logEntry{Type: "tool", Content: msg.event.Content})
			m.status = "tool"
			m.currentTool = msg.event.Content
		case "observation":
			m.logs = append(m.logs, logEntry{Type: "observation", Content: msg.event.Content})
			m.status = "thinking"
			m.currentTool = ""
		case "answer":
			// Replace streaming entry with final response if exists
			if len(m.logs) > 0 && m.logs[len(m.logs)-1].Type == "streaming" {
				m.logs[len(m.logs)-1] = logEntry{Type: "response", Content: msg.event.Content}
			} else {
				m.logs = append(m.logs, logEntry{Type: "response", Content: msg.event.Content})
			}
			m.streamingBuffer = ""
			m.status = "idle"
		case "error":
			m.logs = append(m.logs, logEntry{Type: "error", Content: msg.event.Content})
			m.streamingBuffer = ""
			m.status = "idle"
		}
		m.updateViewportContent()
		cmds = append(cmds, m.spinner.Tick)

	case agentDoneMsg:
		m.thinking = false
		m.status = "idle"
		m.currentTool = ""
		if msg.err != nil {
			m.logs = append(m.logs, logEntry{Type: "error", Content: msg.err.Error()})
		}
		m.updateViewportContent()

	case spinner.TickMsg:
		if m.thinking {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	// Update textinput
	if !m.thinking {
		var cmd tea.Cmd
		m.textinput, cmd = m.textinput.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Update viewport
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *model) updateViewportContent() {
	var content strings.Builder

	for _, entry := range m.logs {
		line := m.formatLogEntry(entry)
		content.WriteString(line)
		content.WriteString("\n")
	}

	// Check if we were at the bottom before updating
	atBottom := m.viewport.AtBottom()

	m.viewport.SetContent(content.String())

	// Only auto-scroll to bottom if we were already at the bottom
	// This allows users to scroll up and read history
	if atBottom || m.thinking {
		m.viewport.GotoBottom()
	}
}

func (m *model) formatLogEntry(entry logEntry) string {
	contentWidth := m.width - 6
	if contentWidth < 40 {
		contentWidth = 40
	}

	switch entry.Type {
	case "user":
		// Blue left border + gray background (OpenCode style)
		return UserMessageStyle.Width(contentWidth).Render(entry.Content)

	case "thinking":
		// Hide thinking entries for cleaner display
		return ""

	case "tool":
		// Circle prefix + tool name: args (dimmed)
		return ToolCallStyle.Render(ToolCallPrefix + entry.Content)

	case "observation":
		// Tool results - show in dimmed text, truncated
		content := entry.Content
		if len(content) > 500 {
			content = content[:400] + "\n... (truncated)"
		}
		// If contains markdown code blocks, render with glamour
		if strings.Contains(entry.Content, "```") && m.renderer != nil {
			rendered, err := m.renderer.Render(entry.Content)
			if err == nil {
				return strings.TrimSpace(rendered)
			}
		}
		return ToolCallStyle.Render("  " + content) // Indent results, no circle

	case "streaming":
		return AgentMessageStyle.Render(entry.Content)

	case "response":
		if m.renderer != nil {
			rendered, err := m.renderer.Render(entry.Content)
			if err == nil {
				return strings.TrimSpace(rendered)
			}
		}
		return AgentMessageStyle.Render(entry.Content)

	case "error":
		return ErrorStyle.Render("  Error: " + entry.Content)

	case "separator":
		return "" // No visible separators in OpenCode style

	default:
		return entry.Content
	}
}

func (m model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	var b strings.Builder

	// Viewport (messages) - no header, maximize space
	b.WriteString(m.viewport.View())
	b.WriteString("\n")

	// Input area
	b.WriteString(m.renderInputArea())
	b.WriteString("\n")

	// Footer
	b.WriteString(m.renderFooter())

	return b.String()
}

// renderStatus renders the current agent status
func (m model) renderStatus() string {
	switch m.status {
	case "thinking":
		return StatusActiveStyle.Render(m.spinner.View() + " thinking...")
	case "streaming":
		return StatusActiveStyle.Render(m.spinner.View() + " streaming...")
	case "tool":
		return StatusToolStyle.Render(m.spinner.View() + " executing " + m.currentTool)
	default:
		return StatusIdleStyle.Render("ready")
	}
}

// renderInputArea renders the OpenCode-style input area with model badge
func (m model) renderInputArea() string {
	if m.thinking {
		statusText := m.renderStatus()
		return InputAreaStyle.Width(m.width - 2).Render(statusText)
	}

	inputView := m.textinput.View()
	badge := ModelBadgeStyle.Render(m.modelName)

	// Calculate spacing to push badge to right
	spacing := m.width - lipgloss.Width(inputView) - lipgloss.Width(badge) - 6
	if spacing < 1 {
		spacing = 1
	}

	inputLine := inputView + strings.Repeat(" ", spacing) + badge
	return InputAreaStyle.Width(m.width - 2).Render(inputLine)
}

// renderFooter renders the footer with keyboard shortcuts (OpenCode style)
func (m model) renderFooter() string {
	var parts []string

	if m.thinking {
		parts = append(parts, ShortcutKeyStyle.Render("esc")+ShortcutDescStyle.Render(" interrupt"))
	} else {
		parts = append(parts, ShortcutKeyStyle.Render("esc")+ShortcutDescStyle.Render(" quit"))
		parts = append(parts, ShortcutKeyStyle.Render("↑↓")+ShortcutDescStyle.Render(" history"))
	}
	parts = append(parts, ShortcutKeyStyle.Render("ctrl+l")+ShortcutDescStyle.Render(" clear"))
	parts = append(parts, ShortcutKeyStyle.Render("ctrl+y")+ShortcutDescStyle.Render(" copy"))

	return FooterStyle.Width(m.width).Render(strings.Join(parts, " | "))
}

// Run starts the TUI application.
func Run() error {
	m := initialModel()
	prog := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())

	// Store program reference for goroutines to send messages
	globalProgram.Set(prog)

	_, err := prog.Run()

	// Clear program reference after run completes
	globalProgram.Set(nil)

	return err
}
