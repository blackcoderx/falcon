package openrouter

import "github.com/blackcoderx/falcon/pkg/llm"

func init() {
	llm.Register(&OpenRouterProvider{})
}

// OpenRouterProvider implements Provider for the OpenRouter gateway.
// OpenRouter is an OpenAI-compatible API that provides access to hundreds
// of models (Claude, GPT-4, Llama, Gemini, etc.) through a single endpoint.
// Find available models at https://openrouter.ai/models
type OpenRouterProvider struct{}

func (p *OpenRouterProvider) ID() string          { return "openrouter" }
func (p *OpenRouterProvider) DisplayName() string  { return "OpenRouter (Claude, GPT-4, Llama, and 100s more)" }
func (p *OpenRouterProvider) DefaultModel() string { return "google/gemini-2.5-flash-lite" }

func (p *OpenRouterProvider) SetupFields() []llm.SetupField {
	return []llm.SetupField{
		{
			Key:         "api_key",
			Title:       "OpenRouter API Key",
			Description: "Get your API key from openrouter.ai/keys.",
			Placeholder: "Enter your OpenRouter API key...",
			Secret:      true,
			EnvFallback: "OPENROUTER_API_KEY",
		},
	}
}

func (p *OpenRouterProvider) BuildClient(values map[string]string, model string) (llm.LLMClient, error) {
	if model == "" {
		model = p.DefaultModel()
	}
	return NewOpenRouterClient(values["api_key"], model), nil
}
