package llm

func init() {
	Register(&OllamaProvider{})
	Register(&GeminiProvider{})
	Register(&OpenRouterProvider{})
}
