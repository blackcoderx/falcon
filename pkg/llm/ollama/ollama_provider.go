package ollama

import "github.com/blackcoderx/falcon/pkg/llm"

func init() {
	llm.Register(&OllamaProvider{})
}

// OllamaProvider implements Provider for the Ollama backend.
// Supports both local (http://localhost:11434) and cloud (https://ollama.com) modes.
type OllamaProvider struct{}

func (p *OllamaProvider) ID() string          { return "ollama" }
func (p *OllamaProvider) DisplayName() string  { return "Ollama (local or cloud)" }
func (p *OllamaProvider) DefaultModel() string { return "llama3" }

func (p *OllamaProvider) SetupFields() []llm.SetupField {
	return []llm.SetupField{
		{
			Key:   "mode",
			Type:  llm.FieldSelect,
			Title: "Ollama mode",
			Description: "Local runs on your machine; Cloud uses Ollama's hosted service.",
			Options: []llm.FieldOption{
				{"Local (run on your machine)", "local"},
				{"Cloud (Ollama Cloud)", "cloud"},
			},
		},
		{
			Key:         "url",
			Title:       "Ollama URL",
			Description: "API endpoint. Local default: http://localhost:11434 · Cloud default: https://ollama.com",
			Placeholder: "http://localhost:11434",
		},
		{
			Key:         "api_key",
			Title:       "API Key",
			Description: "Required for cloud mode. Leave empty for local.",
			Placeholder: "Enter your Ollama Cloud API key...",
			Secret:      true,
			EnvFallback: "OLLAMA_API_KEY",
		},
	}
}

func (p *OllamaProvider) BuildClient(values map[string]string, model string) (llm.LLMClient, error) {
	url := values["url"]
	if url == "" {
		if values["mode"] == "cloud" {
			url = "https://ollama.com"
		} else {
			url = "http://localhost:11434"
		}
	}

	if model == "" {
		if values["mode"] == "cloud" {
			model = "qwen3-coder:480b-cloud"
		} else {
			model = p.DefaultModel()
		}
	}

	return NewOllamaClient(url, model, values["api_key"]), nil
}
