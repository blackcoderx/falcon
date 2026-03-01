package llm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const openRouterBaseURL = "https://openrouter.ai/api/v1"

// openRouterRequest is the OpenAI-compatible request body used by OpenRouter.
type openRouterRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

// openRouterChoice represents a single choice in a non-streaming response.
type openRouterChoice struct {
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// openRouterResponse is the non-streaming response body from OpenRouter.
type openRouterResponse struct {
	ID      string             `json:"id"`
	Model   string             `json:"model"`
	Choices []openRouterChoice `json:"choices"`
	Error   *openRouterError   `json:"error,omitempty"`
}

// openRouterError is returned by OpenRouter when the request fails.
type openRouterError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// openRouterStreamDelta holds the partial content from a streaming chunk.
type openRouterStreamDelta struct {
	Content string `json:"content"`
}

// openRouterStreamChoice is one choice entry in a streaming chunk.
type openRouterStreamChoice struct {
	Delta        openRouterStreamDelta `json:"delta"`
	FinishReason *string               `json:"finish_reason"`
}

// openRouterStreamChunk is a single SSE data payload during streaming.
type openRouterStreamChunk struct {
	ID      string                   `json:"id"`
	Choices []openRouterStreamChoice `json:"choices"`
	Error   *openRouterError         `json:"error,omitempty"`
}

// OpenRouterClient handles communication with the OpenRouter API.
// OpenRouter is an OpenAI-compatible gateway that provides access to
// hundreds of models from providers like Anthropic, OpenAI, Meta, etc.
type OpenRouterClient struct {
	apiKey          string
	model           string
	httpClient      *http.Client // For regular requests (with timeout)
	streamingClient *http.Client // For streaming (no timeout)
}

// NewOpenRouterClient creates a new OpenRouter client.
// The default model is "google/gemini-2.5-flash-lite" if none is specified.
// Find available models at https://openrouter.ai/models
func NewOpenRouterClient(apiKey, model string) *OpenRouterClient {
	if model == "" {
		model = "google/gemini-2.5-flash-lite"
	}
	return &OpenRouterClient{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
		streamingClient: &http.Client{
			Timeout: 0, // No timeout for streaming
		},
	}
}

// newRequest builds an authenticated HTTP POST request for the OpenRouter chat endpoint.
func (c *OpenRouterClient) newRequest(body []byte, stream bool) (*http.Request, error) {
	url := openRouterBaseURL + "/chat/completions"
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	// Recommended by OpenRouter for analytics / rate-limit attribution
	req.Header.Set("X-Title", "Falcon")
	if stream {
		req.Header.Set("Accept", "text/event-stream")
	}
	return req, nil
}

// Chat sends a non-streaming chat request and returns the complete response.
func (c *OpenRouterClient) Chat(messages []Message) (string, error) {
	payload := openRouterRequest{
		Model:    c.model,
		Messages: messages,
		Stream:   false,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := c.newRequest(body, false)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("openrouter request failed: %w", err)
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("openrouter (model: %s) returned status %d: %s", c.model, resp.StatusCode, string(rawBody))
	}

	var result openRouterResponse
	if err := json.Unmarshal(rawBody, &result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("openrouter error (code %d): %s", result.Error.Code, result.Error.Message)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("openrouter returned no choices")
	}

	return result.Choices[0].Message.Content, nil
}

// ChatStream sends a streaming chat request using SSE and calls callback for each chunk.
// Returns the complete response when streaming finishes.
func (c *OpenRouterClient) ChatStream(messages []Message, callback StreamCallback) (string, error) {
	payload := openRouterRequest{
		Model:    c.model,
		Messages: messages,
		Stream:   true,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := c.newRequest(body, true)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.streamingClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("openrouter streaming request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		rawBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("openrouter (model: %s) streaming returned status %d: %s", c.model, resp.StatusCode, string(rawBody))
	}

	// OpenRouter streams using Server-Sent Events (SSE).
	// Each line is either:
	//   "data: <json>"  — a chunk
	//   "data: [DONE]"  — end of stream
	//   ""              — blank separator between events
	var fullContent string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		if !strings.HasPrefix(line, "data: ") {
			continue // skip comment lines, blank lines, etc.
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var chunk openRouterStreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			// Non-fatal: skip malformed SSE lines
			continue
		}

		if chunk.Error != nil {
			return fullContent, fmt.Errorf("openrouter stream error (code %d): %s", chunk.Error.Code, chunk.Error.Message)
		}

		if len(chunk.Choices) == 0 {
			continue
		}

		text := chunk.Choices[0].Delta.Content
		if text != "" {
			fullContent += text
			if callback != nil {
				callback(text)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fullContent, fmt.Errorf("error reading openrouter stream: %w", err)
	}

	return fullContent, nil
}

// CheckConnection verifies that the OpenRouter API is reachable and the key is valid
// by fetching the models list (a cheap, read-only endpoint).
func (c *OpenRouterClient) CheckConnection() error {
	req, err := http.NewRequest(http.MethodGet, openRouterBaseURL+"/models", nil)
	if err != nil {
		return fmt.Errorf("failed to create check request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to OpenRouter: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("openrouter: invalid API key")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("openrouter returned status %d", resp.StatusCode)
	}

	return nil
}

// GetModel returns the model identifier being used.
func (c *OpenRouterClient) GetModel() string {
	return c.model
}
