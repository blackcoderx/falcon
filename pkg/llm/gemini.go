package llm

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/genai"
)

// GeminiClient handles communication with Google's Gemini API.
type GeminiClient struct {
	client *genai.Client
	model  string
	apiKey string
}

// NewGeminiClient creates a new Gemini client with the given API key and model.
// The default model is "gemini-2.5-flash-lite" if none is specified.
func NewGeminiClient(apiKey, model string) (*GeminiClient, error) {
	if model == "" {
		model = "gemini-2.5-flash-lite"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &GeminiClient{
		client: client,
		model:  model,
		apiKey: apiKey,
	}, nil
}

// convertMessages converts our Message type to Gemini Content format.
// Gemini uses "user" and "model" roles (not "assistant").
func (c *GeminiClient) convertMessages(messages []Message) []*genai.Content {
	var contents []*genai.Content

	for _, msg := range messages {
		role := msg.Role
		// Gemini uses "model" instead of "assistant"
		if role == "assistant" {
			role = "model"
		}

		contents = append(contents, &genai.Content{
			Role:  role,
			Parts: []*genai.Part{genai.NewPartFromText(msg.Content)},
		})
	}

	return contents
}

// extractSystemInstruction extracts the system message (if any) from messages.
// Returns the system instruction and remaining messages.
func (c *GeminiClient) extractSystemInstruction(messages []Message) (string, []Message) {
	var systemInstruction string
	var remaining []Message

	for _, msg := range messages {
		if msg.Role == "system" {
			// Concatenate system messages if there are multiple
			if systemInstruction != "" {
				systemInstruction += "\n\n"
			}
			systemInstruction += msg.Content
		} else {
			remaining = append(remaining, msg)
		}
	}

	return systemInstruction, remaining
}

// Chat sends a non-streaming chat request and returns the complete response.
func (c *GeminiClient) Chat(messages []Message) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Extract system instruction from messages
	systemInstruction, conversationMessages := c.extractSystemInstruction(messages)

	// Convert messages to Gemini format
	contents := c.convertMessages(conversationMessages)

	// Build config with system instruction
	var config *genai.GenerateContentConfig
	if systemInstruction != "" {
		config = &genai.GenerateContentConfig{
			SystemInstruction: &genai.Content{
				Parts: []*genai.Part{genai.NewPartFromText(systemInstruction)},
			},
		}
	}

	// Generate content
	response, err := c.client.Models.GenerateContent(ctx, c.model, contents, config)
	if err != nil {
		return "", fmt.Errorf("gemini (model: %s) request failed: %w", c.model, err)
	}

	// Extract text from response
	text := response.Text()
	return text, nil
}

// ChatStream sends a streaming chat request and calls callback for each chunk.
// Returns the complete response when streaming finishes.
func (c *GeminiClient) ChatStream(messages []Message, callback StreamCallback) (string, error) {
	ctx := context.Background() // No timeout for streaming

	// Extract system instruction from messages
	systemInstruction, conversationMessages := c.extractSystemInstruction(messages)

	// Convert messages to Gemini format
	contents := c.convertMessages(conversationMessages)

	// Build config with system instruction
	var config *genai.GenerateContentConfig
	if systemInstruction != "" {
		config = &genai.GenerateContentConfig{
			SystemInstruction: &genai.Content{
				Parts: []*genai.Part{genai.NewPartFromText(systemInstruction)},
			},
		}
	}

	// Stream content
	var fullContent string
	for response, err := range c.client.Models.GenerateContentStream(ctx, c.model, contents, config) {
		if err != nil {
			// If we have partial content, return it with the error
			if fullContent != "" {
				return fullContent, fmt.Errorf("streaming interrupted: %w", err)
			}
			return "", fmt.Errorf("gemini streaming failed: %w", err)
		}

		// Extract text from this chunk
		chunk := response.Text()
		if chunk != "" {
			fullContent += chunk
			if callback != nil {
				callback(chunk)
			}
		}
	}

	return fullContent, nil
}

// CheckConnection verifies that the Gemini API is accessible.
func (c *GeminiClient) CheckConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Try a simple generate request to verify the API is accessible
	contents := []*genai.Content{
		{
			Role:  "user",
			Parts: []*genai.Part{genai.NewPartFromText("Hello")},
		},
	}

	_, err := c.client.Models.GenerateContent(ctx, c.model, contents, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to Gemini API: %w", err)
	}

	return nil
}

// GetModel returns the name of the model being used.
func (c *GeminiClient) GetModel() string {
	return c.model
}
