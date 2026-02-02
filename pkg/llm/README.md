# pkg/llm

This package provides LLM (Large Language Model) client implementations. ZAP supports multiple LLM providers through a common interface.

## Package Overview

```
pkg/llm/
├── client.go    # LLMClient interface definition
├── ollama.go    # Ollama client (local and cloud)
└── gemini.go    # Google Gemini client
```

## LLMClient Interface

All LLM providers must implement this interface:

```go
type LLMClient interface {
    // Chat sends messages and returns the complete response
    Chat(messages []Message) (string, error)

    // ChatStream sends messages and streams the response via callback
    ChatStream(messages []Message, callback StreamCallback) (string, error)

    // CheckConnection verifies the LLM is accessible
    CheckConnection() error

    // GetModel returns the current model name
    GetModel() string
}

type Message struct {
    Role    string // "system", "user", or "assistant"
    Content string
}

type StreamCallback func(chunk string)
```

## Supported Providers

### Ollama (ollama.go)

Supports both local and cloud Ollama instances.

**Local Mode:**

```go
client := llm.NewOllamaClient("http://localhost:11434", "llama3", "")
```

**Cloud Mode (with API key):**

```go
client := llm.NewOllamaClient("https://ollama.example.com", "llama3", "your-api-key")
```

**Features:**

- Streaming support for real-time response display
- Bearer token authentication for cloud instances
- Two HTTP clients: regular (60s timeout) and streaming (no timeout)
- Automatic retry on connection errors

**Configuration:**

```json
{
  "provider": "ollama",
  "ollama": {
    "mode": "local",
    "url": "http://localhost:11434",
    "api_key": ""
  },
  "default_model": "llama3"
}
```

### Google Gemini (gemini.go)

Uses the Google Generative AI API.

```go
client := llm.NewGeminiClient("your-api-key", "gemini-pro")
```

**Features:**

- Streaming support
- Automatic content safety handling
- Supports Gemini Pro and Gemini Pro Vision

**Configuration:**

```json
{
  "provider": "gemini",
  "gemini": {
    "api_key": "your-api-key"
  },
  "default_model": "gemini-pro"
}
```

## Usage

### Basic Chat

```go
client := llm.NewOllamaClient("http://localhost:11434", "llama3", "")

messages := []llm.Message{
    {Role: "system", Content: "You are a helpful assistant."},
    {Role: "user", Content: "Hello!"},
}

response, err := client.Chat(messages)
if err != nil {
    log.Fatal(err)
}
fmt.Println(response)
```

### Streaming Chat

```go
response, err := client.ChatStream(messages, func(chunk string) {
    // Called for each chunk of the response
    fmt.Print(chunk)
})
```

### Connection Check

```go
if err := client.CheckConnection(); err != nil {
    fmt.Printf("LLM not available: %v\n", err)
}
```

## Implementation Details

### Ollama Client

The Ollama client makes POST requests to the `/api/chat` endpoint:

```go
type ollamaRequest struct {
    Model    string    `json:"model"`
    Messages []Message `json:"messages"`
    Stream   bool      `json:"stream"`
}

type ollamaResponse struct {
    Message struct {
        Content string `json:"content"`
    } `json:"message"`
    Done bool `json:"done"`
}
```

**Streaming:**

For streaming, the client reads newline-delimited JSON:

```go
scanner := bufio.NewScanner(resp.Body)
for scanner.Scan() {
    var chunk ollamaResponse
    json.Unmarshal(scanner.Bytes(), &chunk)
    callback(chunk.Message.Content)
    if chunk.Done {
        break
    }
}
```

### Gemini Client

Uses the official Google Generative AI SDK:

```go
import "google.golang.org/genai"

client, _ := genai.NewClient(ctx, option.WithAPIKey(apiKey))
model := client.GenerativeModel(modelName)

resp, _ := model.GenerateContent(ctx, genai.Text(prompt))
```

**Streaming:**

```go
iter := model.GenerateContentStream(ctx, genai.Text(prompt))
for {
    resp, err := iter.Next()
    if err == iterator.Done {
        break
    }
    callback(resp.Candidates[0].Content.Parts[0].(genai.Text))
}
```

## Adding a New Provider

### Step 1: Create the File

```go
// pkg/llm/newprovider.go
package llm

type NewProviderClient struct {
    apiKey string
    model  string
    // Add other fields
}

func NewNewProviderClient(apiKey, model string) *NewProviderClient {
    return &NewProviderClient{
        apiKey: apiKey,
        model:  model,
    }
}
```

### Step 2: Implement the Interface

```go
func (c *NewProviderClient) Chat(messages []Message) (string, error) {
    // Convert messages to provider format
    // Make API call
    // Return response
}

func (c *NewProviderClient) ChatStream(messages []Message, callback StreamCallback) (string, error) {
    // Similar to Chat but stream chunks via callback
}

func (c *NewProviderClient) CheckConnection() error {
    // Make a lightweight API call to verify connectivity
}

func (c *NewProviderClient) GetModel() string {
    return c.model
}
```

### Step 3: Update Configuration

Add provider config in `pkg/core/init.go`:

```go
type Config struct {
    Provider    string `json:"provider"`
    NewProvider struct {
        APIKey string `json:"api_key"`
    } `json:"new_provider"`
    // ...
}
```

### Step 4: Update Setup Wizard

Add option in `pkg/tui/setup/` for new provider selection.

### Step 5: Update TUI Initialization

In `pkg/tui/init.go`:

```go
switch config.Provider {
case "ollama":
    client = llm.NewOllamaClient(...)
case "gemini":
    client = llm.NewGeminiClient(...)
case "newprovider":
    client = llm.NewNewProviderClient(...)
}
```

## Error Handling

All clients should return descriptive errors:

```go
func (c *OllamaClient) Chat(messages []Message) (string, error) {
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return "", fmt.Errorf("failed to connect to Ollama at %s: %w", c.url, err)
    }

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return "", fmt.Errorf("Ollama error (status %d): %s", resp.StatusCode, body)
    }

    // ...
}
```

## Testing

Create mock clients for testing:

```go
type MockLLMClient struct {
    Response string
    Error    error
}

func (m *MockLLMClient) Chat(messages []Message) (string, error) {
    return m.Response, m.Error
}

func (m *MockLLMClient) ChatStream(messages []Message, callback StreamCallback) (string, error) {
    callback(m.Response)
    return m.Response, m.Error
}

func (m *MockLLMClient) CheckConnection() error {
    return m.Error
}

func (m *MockLLMClient) GetModel() string {
    return "mock"
}
```

Run tests:

```bash
go test ./pkg/llm/...
```
