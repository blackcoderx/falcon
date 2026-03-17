package llm

// Provider registration has moved to each provider's own init() function.
// To activate providers, import the subpackages with blank imports from
// the consuming package (e.g., tui/init.go or cmd/falcon/main.go):
//
//   _ "github.com/blackcoderx/falcon/pkg/llm/ollama"
//   _ "github.com/blackcoderx/falcon/pkg/llm/gemini"
//   _ "github.com/blackcoderx/falcon/pkg/llm/openrouter"
