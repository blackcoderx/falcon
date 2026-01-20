# ZAP Product Architecture

> **"The API debugger that reads your code"** — The intersection of Postman + Claude Code

---

## Product Brief

### Core Value Proposition
ZAP is a terminal-native API testing tool that reads your codebase to diagnose API errors and suggest fixes in context.

### Riskiest Assumptions
1. Developers will switch from Postman for AI-powered debugging (behavior change)
2. Local codebase indexing can be fast enough (<30s) to feel seamless
3. Error-to-code mapping accuracy will be high enough to be useful (>70%)
4. Developers trust AI to read their code locally

### Success Metrics for MVP
- Time to first "aha moment" < 5 minutes
- Error-to-code mapping accuracy > 70%
- GitHub stars > 1,000 in first month (viral potential indicator)
- User completes full debug cycle (request → error → fix) without leaving tool

---

## Market Validation

### Market Size
- **API Testing Tools:** $2-4 billion (2024), 21% CAGR
- **AI Coding Assistants:** $5 billion → $30 billion by 2032
- **TAM:** 25-33 million developers work with APIs
- **SAM:** 7-11 million prefer terminal-based workflows
- **SOM (Year 1):** 50,000-200,000 users, $2-10M revenue

### Competitive Gap
| Tool | Terminal-Native | API Testing | Codebase-Aware AI |
|------|-----------------|-------------|-------------------|
| Postman | ✗ | ✓ | ✗ |
| HTTPie | ✓ | ✓ | ✗ |
| Cursor | ✗ | ✗ | ✓ |
| Claude Code | ✓ | ✗ | ✓ |
| Bruno | ✓ | ✓ | ✗ |
| **ZAP** | ✓ | ✓ | ✓ |

### Why Now
- Postman forcing cloud sync → developer exodus (Bruno grew to 200k MAU)
- 80% of enterprises will integrate AI testing by 2027 (Gartner)
- CLI renaissance: terminal tools back in vogue
- Claude Code's $500M+ run-rate proves terminal AI adoption

---

## The ONE Killer Feature

### Error → Code → Fix Pipeline

This is not just a feature—it's the entire product thesis.

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│ HTTP Request│ ──► │ Error/Response│ ──► │ Code Search │ ──► │ Fix Suggest │
│   (exists)  │     │   Analysis   │     │  & Context  │     │   (AI)      │
└─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘
```

### The Demo That Sells
> **Run request → Get error → AI instantly shows the broken line in your codebase → Suggests the fix**

That's the demo. That's the tweet. That's the HN post.

---

## MVP Scope

### 🔴 MUST HAVE (MVP Core)

| Feature | Justification |
|---------|---------------|
| HTTP requests (GET/POST/PUT/DELETE) | Can't test APIs without this |
| Response display with syntax highlighting | Must see what went wrong |
| **File reading tool** | Core to value prop - read codebase |
| **Code search tool (grep/ripgrep)** | Find relevant code for errors |
| **Error analysis prompt** | AI connects error to code |
| Natural language → request | Differentiator, low effort |
| YAML request storage | Git-friendly, no cloud required |
| `.env` file support | Can't test real APIs without secrets |

### 🟡 SHOULD HAVE (Post-MVP, pre-launch)

| Feature | Justification |
|---------|---------------|
| Request collections/folders | Organization at scale |
| Environment switching (dev/staging/prod) | Real workflows need this |
| Response history | Compare responses over time |
| Bearer/API key auth helpers | 80% of APIs use these |
| Postman collection import | Zero switching cost |
| Fix suggestions with patches | Elevates diagnosis to solution |

### 🟢 NICE TO HAVE (V2+)

| Feature | Why Defer |
|---------|-----------|
| Full codebase indexing (embeddings/RAG) | Complex, grep works for MVP |
| OAuth 2.0 flow handling | Complex, users have tokens |
| Request chaining | Useful but not core |
| WebSocket/GraphQL/gRPC | Different protocols, expand later |
| Mock server | Different use case |
| Edge case generation | Requires deep schema understanding |
| Test suite generation | Different workflow |

### ⚪ OUT OF SCOPE (Not MVP)

| Feature | Why Out |
|---------|---------|
| Team collaboration | Individual developer first |
| Cloud sync | Anti-pattern to Postman |
| SSO/Enterprise features | No enterprise customers yet |
| CI/CD integration | Users can script CLI |
| IDE extensions | Terminal-first |
| Load testing | Different tool category |
| Security scanning | Different tool category |
| Database tools | Scope creep |

---

## Technical Decisions

### Decision 1: Codebase Understanding Approach

**Decision:** Grep-based search for MVP

**Rationale:**
- Proven: Claude Code uses grep + file reading successfully
- Zero setup time (no indexing wait)
- Works offline with local LLM
- Can evolve to RAG later

**Tradeoffs Accepted:**
- We lose: Semantic search capabilities
- We risk: Missing relevant code if naming is non-obvious
- Migration path: Add optional RAG indexing in v2

### Decision 2: LLM Strategy

**Decision:** BYO-Key + Ollama default

**Rationale:**
- Ollama default = works out of box, privacy-respecting
- BYO-Key = power users get Claude/GPT quality
- No server costs
- User controls spend

**Tradeoffs Accepted:**
- We lose: Consistent quality across users
- We risk: Bad experience if local model is weak
- Migration path: Hosted option later with usage-based pricing

### Decision 3: Request Storage Format

**Decision:** YAML

**Rationale:**
- Bruno validated this (200k+ users)
- Human-readable, editable anywhere
- Git diffs are meaningful
- Easy to implement in Go

### Decision 4: Tool Architecture

**Required Tools for MVP:**

| Tool | Purpose | Status |
|------|---------|--------|
| `http_request` | Make API calls | ✓ EXISTS |
| `read_file` | Read source code | TODO |
| `search_code` | Grep/ripgrep wrapper | TODO |
| `list_files` | Directory exploration | TODO |
| `write_file` | Apply fixes (with approval) | TODO |

---

## Risk Assessment

| Risk | Prob | Impact | Mitigation |
|------|------|--------|------------|
| Local LLM quality too low | HIGH | HIGH | Recommend good models, BYO-key for cloud |
| Error-code mapping inaccurate | MED | HIGH | Tune prompts, show confidence levels |
| Grep misses relevant code | MED | MED | Iterative search, multiple patterns |
| User doesn't have Ollama | HIGH | MED | Clear setup instructions |
| Competition copies feature | MED | MED | Move fast, build community |

---

## What Makes This YC-Worthy

1. **Clear demo** — 30-second video shows the magic
2. **Real market** — $2-4B TAM, validated by Bruno's 200k users
3. **Unoccupied niche** — No tool combines terminal + API + codebase AI
4. **Technical moat** — Prompt engineering + tool design is defensible
5. **PLG ready** — Individual devs adopt, pull into teams
6. **Clear monetization** — Free tier (limited AI) → Pro (unlimited)

---

## Pricing Strategy (Future)

| Tier | Price | Features |
|------|-------|----------|
| **Free** | $0 | Full CLI, 50 AI queries/month, local storage |
| **Pro** | $15/mo | Unlimited AI, advanced debugging, priority support |
| **Team** | $35/user/mo | Shared collections, team workspaces, admin controls |
| **Enterprise** | Custom | SSO, audit logs, self-hosted AI, dedicated support |

---

## Success Milestones

| Timeframe | Users | ARR | GitHub Stars |
|-----------|-------|-----|--------------|
| Month 3 | 10,000 | - | 5,000 |
| Month 6 | 25,000 | - | 10,000 |
| Month 12 | 100,000 | $500K | 25,000 |
| Month 18 | 250,000 | $2M | 50,000 |
