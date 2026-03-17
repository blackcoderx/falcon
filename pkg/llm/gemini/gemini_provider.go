package gemini

import "github.com/blackcoderx/falcon/pkg/llm"

func init() {
	llm.Register(&GeminiProvider{})
}

// GeminiProvider implements Provider for Google's Gemini backend.
type GeminiProvider struct{}

func (p *GeminiProvider) ID() string          { return "gemini" }
func (p *GeminiProvider) DisplayName() string  { return "Gemini (Google AI)" }
func (p *GeminiProvider) DefaultModel() string { return "gemini-2.5-flash-lite" }

func (p *GeminiProvider) SetupFields() []llm.SetupField {
	return []llm.SetupField{
		{
			Key:         "api_key",
			Title:       "Gemini API Key",
			Description: "Get your API key from aistudio.google.com.",
			Placeholder: "Enter your Gemini API key...",
			Secret:      true,
			EnvFallback: "GEMINI_API_KEY",
		},
	}
}

func (p *GeminiProvider) BuildClient(values map[string]string, model string) (llm.LLMClient, error) {
	if model == "" {
		model = p.DefaultModel()
	}
	return NewGeminiClient(values["api_key"], model)
}
