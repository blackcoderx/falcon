# pkg/llm

This package provides LLM (Large Language Model) client implementations and a provider registry. Falcon supports multiple LLM backends through a common interface and a self-registering provider system — adding a new provider requires no changes to the setup wizard or client factory.

## Package Overview

```
pkg/llm/
├── client.go                # LLMClient interface + Message/StreamCallback types
├── provider.go              # Provider interface + SetupField types
├── registry.go              # Global provider registry (Register, Get, All)
├── register_providers.go    # init() — registers all built-in providers
├── ollama.go                # Ollama HTTP client (local and cloud)
├── ollama_provider.go       # OllamaProvider — registry metadata + BuildClient
├── gemini.go                # Google Gemini client (official SDK)
├── gemini_provider.go       # GeminiProvider — registry metadata + BuildClient
├── openrouter.go            # OpenRouter HTTP client (OpenAI-compatible gateway)
└── openrouter_provider.go   # OpenRouterProvider — registry metadata + BuildClient
```

## LLMClient Interface

All provider clients implement this interface (`client.go`):

```go
type LLMClient interface {
    Chat(messages []Message) (string, error)
    ChatStream(messages []Message, callback StreamCallback) (string, error)
    CheckConnection() error
    GetModel() string
}

type Message struct {
    Role    string // "system", "user", or "assistant"
    Content string
}

type StreamCallback func(chunk string)
```

## Provider Interface

Every provider registration implements this interface (`provider.go`). It describes both how to show setup UI and how to build the client at runtime.

```go
type Provider interface {
    ID() string                // stable config key, e.g. "openrouter"
    DisplayName() string       // shown in setup wizard
    DefaultModel() string      // used when user leaves model field blank
    SetupFields() []SetupField // fields to collect during first-run setup
    BuildClient(values map[string]string, model string) (LLMClient, error)
}
```

### SetupField

Describes a single configuration field rendered by the setup wizard:

```go
type SetupField struct {
    Key         string        // config key, e.g. "api_key"
    Type        FieldType     // FieldInput (default) or FieldSelect
    Title       string        // label shown in wizard
    Description string
    Placeholder string
    Secret      bool          // echo as password
    Default     string        // applied when user leaves blank
    Options     []FieldOption // only for FieldSelect
    EnvFallback string        // env var checked at runtime when viper value is empty
}
```

## Supported Providers

### Ollama (`ollama.go` + `ollama_provider.go`)

Local and cloud Ollama instances. Uses Ollama's `/api/chat` endpoint with newline-delimited JSON streaming.

```go
client := llm.NewOllamaClient("http://localhost:11434", "llama3", "")
```

**Config format:**
```yaml
provider: ollama
default_model: llama3
provider_config:
  mode: local          # "local" or "cloud"
  url: http://localhost:11434
  api_key: ""          # required for cloud mode
```

**Setup fields:** mode (select), url, api_key
**Env fallback:** `OLLAMA_API_KEY`

---

### Google Gemini (`gemini.go` + `gemini_provider.go`)

Uses the official `google.golang.org/genai` SDK. Handles Gemini's role mapping (`"assistant"` → `"model"`) and system instruction extraction automatically.

```go
client, err := llm.NewGeminiClient("your-api-key", "gemini-2.5-flash-lite")
```

**Config format:**
```yaml
provider: gemini
default_model: gemini-2.5-flash-lite
provider_config:
  api_key: your-api-key
```

**Setup fields:** api_key
**Env fallback:** `GEMINI_API_KEY`

---

### OpenRouter (`openrouter.go` + `openrouter_provider.go`)

OpenAI-compatible gateway giving access to hundreds of models (Claude, GPT-4, Llama, Gemini, etc.) through a single API endpoint at `https://openrouter.ai/api/v1`. Uses SSE streaming.

```go
client := llm.NewOpenRouterClient("your-api-key", "google/gemini-2.5-flash-lite")
```

**Config format:**
```yaml
provider: openrouter
default_model: google/gemini-2.5-flash-lite
provider_config:
  api_key: your-api-key
```

**Setup fields:** api_key
**Env fallback:** `OPENROUTER_API_KEY`
**Browse models:** https://openrouter.ai/models

---

## Registry

The registry (`registry.go`) is a simple ordered map. All built-in providers are registered in `register_providers.go` via `init()`:

```go
func init() {
    Register(&OllamaProvider{})
    Register(&GeminiProvider{})
    Register(&OpenRouterProvider{})
}
```

**API:**

```go
llm.All()          // []Provider in registration order — used to build wizard UI
llm.Get("gemini")  // (Provider, bool) — used by the client factory at runtime
```

The setup wizard and client factory consume the registry — they never need to change when a provider is added.

## Usage

### Basic Chat

```go
client := llm.NewOllamaClient("http://localhost:11434", "llama3", "")

messages := []llm.Message{
    {Role: "system", Content: "You are a helpful assistant."},
    {Role: "user", Content: "Hello!"},
}

response, err := client.Chat(messages)
```

### Streaming Chat

```go
response, err := client.ChatStream(messages, func(chunk string) {
    fmt.Print(chunk)
})
```

### Connection Check

```go
if err := client.CheckConnection(); err != nil {
    fmt.Printf("LLM not available: %v\n", err)
}
```

## Adding a New Provider

Three steps, no changes to any other package:

### Step 1: Implement LLMClient

```go
// pkg/llm/myprovider.go
package llm

type MyProviderClient struct {
    apiKey string
    model  string
}

func NewMyProviderClient(apiKey, model string) *MyProviderClient { ... }

func (c *MyProviderClient) Chat(messages []Message) (string, error)                          { ... }
func (c *MyProviderClient) ChatStream(messages []Message, cb StreamCallback) (string, error) { ... }
func (c *MyProviderClient) CheckConnection() error                                            { ... }
func (c *MyProviderClient) GetModel() string                                                  { return c.model }
```

### Step 2: Implement Provider

```go
// pkg/llm/myprovider_provider.go
package llm

type MyProvider struct{}

func (p *MyProvider) ID() string          { return "myprovider" }
func (p *MyProvider) DisplayName() string  { return "My Provider" }
func (p *MyProvider) DefaultModel() string { return "my-default-model" }

func (p *MyProvider) SetupFields() []SetupField {
    return []SetupField{
        {
            Key:         "api_key",
            Title:       "API Key",
            Secret:      true,
            EnvFallback: "MYPROVIDER_API_KEY",
        },
    }
}

func (p *MyProvider) BuildClient(values map[string]string, model string) (LLMClient, error) {
    if model == "" {
        model = p.DefaultModel()
    }
    return NewMyProviderClient(values["api_key"], model), nil
}
```

### Step 3: Register

```go
// pkg/llm/register_providers.go — add one line:
func init() {
    Register(&OllamaProvider{})
    Register(&GeminiProvider{})
    Register(&OpenRouterProvider{})
    Register(&MyProvider{})   // ← add this
}
```

The provider now appears in the setup wizard, gets its own config section, and is instantiated at runtime. **Zero changes to `pkg/core/init.go` or `pkg/tui/init.go`.**

## Testing

Create a mock client for unit tests:

```go
type MockLLMClient struct {
    Response string
    Err      error
}

func (m *MockLLMClient) Chat(_ []Message) (string, error)                          { return m.Response, m.Err }
func (m *MockLLMClient) ChatStream(_ []Message, cb StreamCallback) (string, error) { cb(m.Response); return m.Response, m.Err }
func (m *MockLLMClient) CheckConnection() error                                     { return m.Err }
func (m *MockLLMClient) GetModel() string                                           { return "mock" }
```

```bash
go test ./pkg/llm/...
```
