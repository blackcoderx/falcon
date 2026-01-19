# ⚡ ZAP

> AI-powered API testing in your terminal

**ZAP** is the developer assistant that lives where you work—your terminal. It bridges the gap between coding, testing, and fixing by giving you an autonomous agent that understands your code and can interact with your APIs naturally.

## 🚀 Features

- **Context-Aware**: Reads your actual code to understand API endpoints
- **AI-Powered**: Uses local LLMs (Ollama) or cloud providers (OpenAI/Anthropic)
- **Beautiful TUI**: Modern, vibrant terminal interface built with Charm
- **Secure**: API keys stored in `.env`, never in plain text
- **Memory**: Persistent conversation history and learned facts

## 🎯 Vision

Stop context-switching between your editor, Postman, and terminal. ZAP brings them together with an AI that understands your specific API implementation.

## 🛠️ Tech Stack

- **Language**: Go 1.23+
- **CLI Framework**: Cobra + Viper
- **TUI Stack**: Bubble Tea, Lip Gloss, Huh, Bubbles, Glamour
- **LLM Interface**: Raw HTTP Client (no LangChain)
- **Local LLM**: Ollama (with cloud provider fallback)

## 📦 Installation

### Prerequisites

- Go 1.23 or higher
- [Ollama](https://ollama.ai/) (optional, for local AI)

### Build from Source

```bash
git clone https://github.com/blackcoderx/zap.git
cd zap
go build -o zap.exe ./cmd/zap
```

## 🚦 Quick Start

Run ZAP:

```bash
./zap
```

On first run, ZAP will create a `.zap` folder with:
- `config.json` - Your preferences
- `history.jsonl` - Conversation log
- `memory.json` - Agent memory

## ⚙️ Configuration

Edit `.zap/config.json`:

```json
{
  "ollama_url": "http://localhost:11434",
  "default_model": "llama3",
  "theme": "dark"
}
```

For API keys, create a `.env` file:

```env
OPENAI_API_KEY=your_key_here
ANTHROPIC_API_KEY=your_key_here
```

## 🎨 Design Philosophy

1. **Context is King** - The agent sees the actual code, not guesses
2. **Human in the Loop** - Dangerous operations require approval
3. **Fail Loudly** - Errors should be visible and helpful
4. **Local First** - Everything works offline except LLM calls
5. **Beautiful UX** - Production-quality interface, not a prototype

## 🗺️ Roadmap

### Phase 1: The "Smart Curl" (MVP) ✅ In Progress
- [x] Scaffold Go project
- [x] Beautiful TUI with Charm stack
- [x] `.zap` folder initialization
- [ ] Connect to Ollama
- [ ] HTTP client tool

### Phase 2: Security & Context
- [ ] `.env` loader
- [ ] File reading tool
- [ ] Code search tool
- [ ] Conversation history

### Phase 3: The "Fixer" & Extension
- [ ] Code editing with approval
- [ ] VS Code extension

## 📝 License

MIT

## 🤝 Contributing

Contributions welcome! This is an active development project.

## 🙏 Acknowledgments

Built with the amazing [Charm](https://charm.sh/) ecosystem.