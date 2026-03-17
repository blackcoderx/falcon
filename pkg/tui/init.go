package tui

import (
	"fmt"
	"os"

	"github.com/blackcoderx/falcon/pkg/core"
	"github.com/blackcoderx/falcon/pkg/core/tools"
	"github.com/blackcoderx/falcon/pkg/core/tools/shared"
	"github.com/blackcoderx/falcon/pkg/llm"
	"github.com/blackcoderx/falcon/pkg/llm/ollama"
	_ "github.com/blackcoderx/falcon/pkg/llm/gemini"
	_ "github.com/blackcoderx/falcon/pkg/llm/openrouter"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/harmonica"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"
)

// configureToolLimits sets up per-tool call limits from config file.
// Falls back to sensible defaults if config values are missing.
// High-risk tools (network I/O, side effects) have lower limits.
// Low-risk tools (in-memory, no side effects) have higher limits.
func configureToolLimits(agent *core.Agent) {
	// Default limits (used if config doesn't specify)
	// We use the core defaults as the source of truth
	defaultLimits := core.DefaultToolLimits

	// Set global limits from config (with defaults)
	defaultLimit := viper.GetInt("tool_limits.default_limit")
	if defaultLimit <= 0 {
		defaultLimit = 50
	}
	agent.SetDefaultLimit(defaultLimit)

	totalLimit := viper.GetInt("tool_limits.total_limit")
	if totalLimit <= 0 {
		totalLimit = 200
	}
	agent.SetTotalLimit(totalLimit)

	// Apply default per-tool limits first
	for toolName, limit := range defaultLimits {
		agent.SetToolLimit(toolName, limit)
	}

	// Override with config values if present
	perToolConfig := viper.GetStringMap("tool_limits.per_tool")
	for toolName, limitVal := range perToolConfig {
		// viper returns interface{}, need to convert to int
		var limit int
		switch v := limitVal.(type) {
		case int:
			limit = v
		case int64:
			limit = int(v)
		case float64:
			limit = int(v)
		default:
			continue // Skip invalid values
		}
		if limit > 0 {
			agent.SetToolLimit(toolName, limit)
		}
	}
}

// registerTools adds all tools to the agent.
// This includes codebase tools, persistence tools, and testing tools from all sprints.
// registerTools adds all tools to the agent using the central registry.
// This switches Falcon to use the new modular tool packages (shared, debugging, persistence, agent).
func registerTools(agent *core.Agent, falconDir, workDir string, confirmManager *shared.ConfirmationManager, memStore *core.MemoryStore) {
	registry := tools.NewRegistry(agent, agent.LLMClient(), workDir, falconDir, memStore, confirmManager)
	registry.RegisterAllTools()
}

// newLLMClient creates and configures the LLM client from Viper config.
// Provider selection and instantiation are fully driven by the llm.Provider
// registry — adding a new provider requires no changes here.
func newLLMClient() llm.LLMClient {
	providerID := viper.GetString("provider")
	model := viper.GetString("default_model")

	p, ok := llm.Get(providerID)
	if !ok {
		// Unknown provider — fall back to legacy Ollama config
		return newOllamaClientFallback(model)
	}

	values := collectProviderValues(p)
	client, err := p.BuildClient(values, model)
	if err != nil {
		return newOllamaClientFallback(model)
	}
	return client
}

// collectProviderValues reads provider_config from viper and applies env-variable
// fallbacks for any fields whose value is empty.
func collectProviderValues(p llm.Provider) map[string]string {
	values := viper.GetStringMapString("provider_config")
	for _, f := range p.SetupFields() {
		if values[f.Key] == "" && f.EnvFallback != "" {
			values[f.Key] = os.Getenv(f.EnvFallback)
		}
	}
	return values
}

// newOllamaClientFallback creates an Ollama client using legacy top-level config
// fields. Used for backward compatibility when the provider registry lookup fails.
func newOllamaClientFallback(model string) *ollama.OllamaClient {
	url := viper.GetString("ollama_url")
	if url == "" {
		url = "http://localhost:11434"
	}

	apiKey := viper.GetString("ollama_api_key")
	if apiKey == "" {
		apiKey = os.Getenv("OLLAMA_API_KEY")
	}

	if model == "" {
		model = "llama3"
	}

	return ollama.NewOllamaClient(url, model, apiKey)
}

// newSpinner creates a spinner with the Falcon style (points animation).
func newSpinner() spinner.Model {
	sp := spinner.New()
	sp.Spinner = spinner.Points
	sp.Style = lipgloss.NewStyle().Foreground(AccentColor)
	return sp
}

const FalconASCII = `███████╗ █████╗ ██╗      ██████╗ ██████╗ ███╗   ██╗
██╔════╝██╔══██╗██║     ██╔════╝██╔═══██╗████╗  ██║
█████╗  ███████║██║     ██║     ██║   ██║██╔██╗ ██║
██╔══╝  ██╔══██║██║     ██║     ██║   ██║██║╚██╗██║
██║     ██║  ██║███████╗╚██████╗╚██████╔╝██║ ╚████║
╚═╝     ╚═╝  ╚═╝╚══════╝ ╚═════╝ ╚═════╝ ╚═╝  ╚═══╝`

// newTextInput creates a text input with the Falcon style.
// No prompt prefix - clean input area.
// init.go

// newTextInput creates a text input with the Falcon style.
func newTextInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "Ask me anything..."
	ti.Focus()
	ti.CharLimit = 2000
	ti.Width = 80
	ti.Prompt = "" // No prompt prefix

	// --- FIX STARTS HERE ---

	// We need to match the textinput background to the container background
	// defined in your tui.go (InputAreaBg)

	// 1. The text you type
	ti.TextStyle = lipgloss.NewStyle().
		Foreground(TextColor).
		Background(InputAreaBg)

	// 2. The placeholder text ("Ask me anything...")
	ti.PlaceholderStyle = lipgloss.NewStyle().
		Foreground(DimColor).
		Background(InputAreaBg)

	// 3. The blinking cursor
	ti.Cursor.Style = lipgloss.NewStyle().
		Foreground(AccentColor).
		Background(InputAreaBg)

	// --- FIX ENDS HERE ---

	return ti
}

// newGlamourRenderer creates a glamour renderer for markdown.
func newGlamourRenderer() *glamour.TermRenderer {
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	return renderer
}

// updateGlamourWidth recreates the glamour renderer with a new word wrap width.
// This is called when the terminal is resized to ensure markdown renders correctly.
func (m *Model) updateGlamourWidth(width int) {
	if width < 40 {
		width = 40
	}
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err == nil {
		m.renderer = renderer
	}
}

// InitialModel creates and returns the initial TUI model.
// webPort is the port the web UI is listening on (0 if disabled).
func InitialModel(webPort int) Model {
	// Get current working directory for codebase tools
	workDir, _ := os.Getwd()

	// Get .falcon directory path
	falconDir := core.FalconFolderName

	// Get model name for display
	modelName := viper.GetString("default_model")
	if modelName == "" {
		modelName = "llama3"
	}

	client := newLLMClient()
	agent := core.NewAgent(client)

	// Set framework from config for context-aware assistance
	framework := viper.GetString("framework")
	if framework == "" {
		// Fallback: read directly from config file (for first-run scenarios)
		framework = core.GetConfigFramework()
	}
	agent.SetFramework(framework)

	// Configure per-tool call limits before registering tools
	configureToolLimits(agent)

	// Create confirmation manager for file write approvals (shared between tool and TUI)
	confirmManager := shared.NewConfirmationManager()

	// Set up timeout callback to notify TUI when confirmation times out
	confirmManager.SetTimeoutCallback(func() {
		globalProgram.Send(confirmationTimeoutMsg{})
	})

	// Create memory store for persistent agent memory
	memStore := core.NewMemoryStore(falconDir)
	agent.SetMemoryStore(memStore)

	registerTools(agent, falconDir, workDir, confirmManager, memStore)

	m := Model{
		textinput:        newTextInput(),
		spinner:          newSpinner(),
		logs:             []logEntry{},
		thinking:         false,
		agent:            agent,
		ready:            false,
		renderer:         newGlamourRenderer(),
		inputHistory:     []string{},
		historyIdx:       -1,
		savedInput:       "",
		status:           "idle",
		currentTool:      "",
		streamingBuffer:  "",
		modelName:        modelName,
		confirmManager:   confirmManager,
		confirmationMode: false,
		memoryStore:      memStore,

		// Initialize harmonica spring for pulsing animation
		// frequency=5.0 (moderate oscillation speed), damping=0.3 (keeps bouncing)
		animSpring: harmonica.NewSpring(harmonica.FPS(30), 5.0, 0.3),
		animPos:    0.0,
		animVel:    0.0,
		animTarget: 1.0,
	}

	// Add splash screen to logs
	m.logs = append(m.logs, logEntry{
		Type:    "splash",
		Content: SplashStyle.Render(FalconASCII),
	})

	version := "1.0.0"
	webUIInfo := ""
	if webPort > 0 {
		webUIInfo = fmt.Sprintf(" • Web UI: %s",
			SplashVersionStyle.Render(fmt.Sprintf("http://localhost:%d", webPort)),
		)
	}
	splashInfo := fmt.Sprintf("Falcon v%s • Current dir: %s%s",
		SplashVersionStyle.Render(version),
		SplashInfoStyle.Render(workDir),
		webUIInfo,
	)

	m.logs = append(m.logs, logEntry{
		Type:    "splash",
		Content: SplashInfoStyle.Render(splashInfo),
	})

	m.logs = append(m.logs, logEntry{
		Type:    "splash",
		Content: "\n",
	})

	return m
}

// Init initializes the Bubble Tea model.
// This is called once when the program starts.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		textinput.Blink,
		m.spinner.Tick,
		animTick(), // Start harmonica spring animation loop
	)
}
