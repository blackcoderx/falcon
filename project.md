# ZAP: Project Analysis & Master Plan

## 1. Product Brief
**Vision:** ZAP is the AI-powered developer assistant that lives where you work—your terminal. It bridges the gap between coding (Go codebase), testing (API requests), and fixing (debugging) by giving you an autonomous agent that understands your code and can interact with your APIs naturally.

**Value Proposition:** "Stop context-switching between your editor, Postman, and terminal. ZAP brings them together with an AI that understands your specific API implementation."

**Core Pillars:**
1.  **Context-Awareness:** It doesn't just guess; it reads your `main.go` / `routes` to know the *actual* endpoints.
2.  **Action-Oriented:** It doesn't just chat; it executes HTTP requests and (eventually) fixes code.
3.  **Developer-Centric:** CLI/TUI first, high performance, local-first memory, security-conscious.

---

## 2. Technical Architecture Strategy

### High-Level Pattern: Hybrid Agentic CLI
We will adopt a **Modular Monolith** architecture for the CLI, designed to be "server-ready" for future Extension integration.

#### The "Cursor vs Claude Code" Decision
**Recommendation: Start with Agentic Exploration (Claude Code style), Evolve to Indexing.**

*   **Why Agentic First (MVP)?**
    *   **Simplicity:** Building a reliable vector sync engine (Cursor style) is a massive engineering undertaking.
    *   **Freshness:** An agent that runs `grep` or `ls` sees the code *exactly* as it is now. No stale indexes.
    *   **Fit:** For API testing, you usually care about a specific slice of code (the handler), not the entire repo at once.
*   **When to Switch?** When the codebase is >10k files or "thinking" takes >10s due to search latency.

### System Components
1.  **Core Engine (`pkg/core`):**
    *   **Context Manager:** Manages the `.zap` folder context (Project memory).
    *   **Tool Registry:** Handles execution of `http`, `file_read`, `search`.
    *   **LLM Interface:** **Raw HTTP Client** (No LangChain) connecting to Ollama/OpenAI.
2.  **TUI Layer (`cmd/zap`):**
    *   **CLI Framework:** `cobra` (commands) + `viper` (config).
    *   **UI Toolkit:** The **Charm Ecosystem**:
        *   `bubbletea`: The ELM architecture runtime.
        *   `lipgloss`: Modern, beautiful styling (colors, layouts).
        *   `huh`: Interactive forms (inputs, selects) for user prompts.
        *   `bubbles`: Reusable components (spinners, paginators).
        *   `glamour`: Beautiful Markdown rendering for LLM responses.
3.  **Memory & Security Store (`.zap/`):**
    *   **Initialization:** Auto-created on first run if missing.
    *   **Security:** `secure storage` / `.env` integration for API keys. **Never** store secrets in plain JSON.
    *   `history.jsonl`: Conversation log.
    *   `context.json`: Learned facts.

---

## 3. Agent Design (The "Brain")

**Architecture:** **ReAct Loop (Reason -> Act -> Observe)**

### Primary Tools
| Tool Name | Purpose | Example Usage |
|-----------|---------|---------------|
| `FileSystem` | Read code to understand API definitions. | "Read `handlers/user.go` to see the JSON struct." |
| `CodeSearch` | Locate relevant files. | "Search for 'func Login' to find the auth handler." |
| `HttpClient` | Execute API requests. | "POST /login with username='admin'..." |
| `Memory` | Recall user preferences/vars. | "Remember that the base URL is localhost:8080." |
| `EnvManager` | Securely access secrets. | "Read `API_KEY` from `.env`." |

### Context Strategy
*   **System Prompt:** Defines the "Senior Backend Engineer" persona.
*   **Dynamic Context:** When a user asks about "Login", the agent searches for "Login", reads the top 2 files, and injects that code into the context window *before* answering.
*   **Memory:** Persistent variables stored in `.zap/memory.json` (e.g., `JWT_TOKEN=xyz`).

---

## 4. Implementation Roadmap

### Phase 1: The "Smart Curl" (MVP)
**Goal:** A TUI that replaces curl/Postman for basic tasks.
*   [ ] **Scaffold:** Go project with `cobra` + `viper` + `bubbletea`.
*   [ ] **Initialization:** Check/Create `.zap` folder on startup.
*   [ ] **LLM Integration:** **Raw HTTP Client** to Ollama (no LangChain).
*   [ ] **UI:** Beautiful interface using `lipgloss` & `glamour`.
*   [ ] **Input:** Use `huh` for natural language input forms.
*   [ ] **Tooling:** Implement `HttpClient` (GET/POST/PUT/DELETE).
*   [ ] **Feature:** "Natural Language Request" ("Test the google homepage" -> Agent executes GET google.com).

### Phase 2: Security & Context
**Goal:** The tool is safe and understands your API.
*   [ ] **Security:** Implement `.env` loader and secure key storage.
*   [ ] **Tooling:** Implement `FileSystem` (Read) and `CodeSearch` (Grep).
*   [ ] **Feature:** "Analyze `routes.go` and list all endpoints."
*   [ ] **Feature:** "Test the /login endpoint I just wrote."
*   [ ] **Memory:** Save conversation history to `.zap/history`.

### Phase 3: The "Fixer" & Extension
**Goal:** Close the loop (Edit code) and integrate with VS Code.
*   [ ] **Tooling:** Implement `FileEdit` (Diff/Patch).
*   [ ] **Safety:** Add "Human Approval" step before any write operation.
*   [ ] **Extension:** Create a VS Code extension that spawns `zap serve` and communicates via JSON-RPC.

---

## 5. Technology Stack Decisions (Confirmed)
*   **Language:** Go 1.23+
*   **CLI Framework:** Cobra + Viper.
*   **TUI Stack:** Bubble Tea, Lip Gloss, Huh, Bubbles, Glamour.
*   **LLM Interface:** Raw HTTP Client (Custom implementation).
*   **Local LLM:** Ollama (Models: `llama3`, `mistral`, `codellama`).
*   **Storage:** JSON files (Context) + Secure Storage (Secrets).

## 6. Risk Assessment
*   **Risk:** Local LLMs (Ollama) may be too dumb for complex debugging.
    *   *Mitigation:* Design the prompt carefully; allow easy config switch to OpenAI/Anthropic API keys if user wants "Pro" mode.
*   **Risk:** Context Window Limits.
    *   *Mitigation:* Strict aggressive filtering. Only read relevant functions, not whole files.


## MY Thoughts 
To ensure that the security and privacy of the user is maintained we will use .env or secure file storage to store the API keys and other sensitive information and the agent can access it when needed through variables. So the agent will instruct the user to store the API keys in the .env file or secure file storage and the agent will use it when needed. 

I want the UI to be like claude code style but beautiful and modern. 

using your search tool , check glamour, bubbles, huh , lipgloss,  all from the charm library, You can check their github pages [text](https://github.com/orgs/charmbracelet/repositories)
let us use cobra and viper for cli and creation of .zap upon init( once). when the user closes and opens the terminal to run zap again , the zap is checked and if it is not present then it is created. 

NO langchain, we will use raw http client to call ollama api. 
