package core

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/blackcoderx/zap/pkg/core/tools/shared"
	"github.com/blackcoderx/zap/pkg/llm"
)

// MemoryEntry represents a single fact saved by the agent.
type MemoryEntry struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	Category  string `json:"category"`  // "preference", "endpoint", "error", "project", "general"
	Timestamp string `json:"timestamp"` // RFC3339
	Source    string `json:"source"`    // Session ID that created this
}

// SessionEntry represents a summary of a past session.
type SessionEntry struct {
	SessionID string   `json:"session_id"`
	StartTime string   `json:"start_time"`
	EndTime   string   `json:"end_time"`
	Summary   string   `json:"summary"`
	Topics    []string `json:"topics"`
	ToolsUsed []string `json:"tools_used"`
	TurnCount int      `json:"turn_count"`
}

// memoryFile is the on-disk format of memory.json.
type memoryFile struct {
	Version int           `json:"version"`
	Entries []MemoryEntry `json:"entries"`
}

// MemoryStore manages persistent agent memory and session tracking.
type MemoryStore struct {
	entries   []MemoryEntry
	mu        sync.RWMutex
	zapDir    string
	sessionID string
	startTime time.Time
	topics    map[string]bool
	toolsUsed map[string]bool
	turnCount int
}

// NewMemoryStore creates a MemoryStore, loads existing memory, and generates a session ID.
func NewMemoryStore(zapDir string) *MemoryStore {
	ms := &MemoryStore{
		zapDir:    zapDir,
		sessionID: fmt.Sprintf("session_%s", time.Now().Format("20060102_150405")),
		startTime: time.Now(),
		topics:    make(map[string]bool),
		toolsUsed: make(map[string]bool),
	}
	ms.loadMemory()
	return ms
}

// Save upserts a memory entry (updates if key exists, inserts otherwise) and persists to disk.
// Returns an error if attempting to save secrets to memory.
func (ms *MemoryStore) Save(key, value, category string) error {
	// Check for secrets - prevent saving sensitive data to memory
	if shared.IsSecret(key, value) {
		return fmt.Errorf("cannot save secrets to memory. Use the 'variable' tool with session scope instead for sensitive values like tokens and passwords")
	}

	ms.mu.Lock()
	defer ms.mu.Unlock()

	if category == "" {
		category = "general"
	}

	entry := MemoryEntry{
		Key:       key,
		Value:     value,
		Category:  category,
		Timestamp: time.Now().Format(time.RFC3339),
		Source:    ms.sessionID,
	}

	// Upsert: replace if key exists
	found := false
	for i, e := range ms.entries {
		if e.Key == key {
			ms.entries[i] = entry
			found = true
			break
		}
	}
	if !found {
		ms.entries = append(ms.entries, entry)
	}

	return ms.saveMemory()
}

// Recall searches memory entries by substring match across key, value, and category.
func (ms *MemoryStore) Recall(query string) []MemoryEntry {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	query = strings.ToLower(query)
	var results []MemoryEntry
	for _, e := range ms.entries {
		if strings.Contains(strings.ToLower(e.Key), query) ||
			strings.Contains(strings.ToLower(e.Value), query) ||
			strings.Contains(strings.ToLower(e.Category), query) {
			results = append(results, e)
		}
	}
	return results
}

// Forget removes a memory entry by key and persists the change.
func (ms *MemoryStore) Forget(key string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	for i, e := range ms.entries {
		if e.Key == key {
			ms.entries = append(ms.entries[:i], ms.entries[i+1:]...)
			return ms.saveMemory()
		}
	}
	return fmt.Errorf("memory key '%s' not found", key)
}

// List returns all memory entries.
func (ms *MemoryStore) List() []MemoryEntry {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	result := make([]MemoryEntry, len(ms.entries))
	copy(result, ms.entries)
	return result
}

// ListByCategory returns entries matching the given category.
func (ms *MemoryStore) ListByCategory(category string) []MemoryEntry {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	var results []MemoryEntry
	for _, e := range ms.entries {
		if strings.EqualFold(e.Category, category) {
			results = append(results, e)
		}
	}
	return results
}

// GetCompactSummary generates a compact string for injection into the system prompt.
// Returns empty string if no memories or sessions exist.
func (ms *MemoryStore) GetCompactSummary() string {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	var sb strings.Builder

	// Get recent sessions
	sessions := ms.getRecentSessionsUnlocked(3)

	if len(ms.entries) == 0 && len(sessions) == 0 {
		return ""
	}

	sb.WriteString("## AGENT MEMORY\n")
	sb.WriteString("The following are facts from previous sessions. Use the `memory` tool to save new facts or recall details.\n\n")

	// Recent sessions summary
	if len(sessions) > 0 {
		last := sessions[len(sessions)-1]
		fmt.Fprintf(&sb, "Recent sessions: %d sessions, last: \"%s\"\n\n", len(sessions), last.Summary)
	}

	// Remembered facts
	if len(ms.entries) > 0 {
		sb.WriteString("Remembered facts:\n")
		for _, e := range ms.entries {
			sb.WriteString(fmt.Sprintf("- [%s] %s: %s\n", e.Category, e.Key, e.Value))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("Save important discoveries with memory tool. Forget outdated info when things change.\n\n")

	return sb.String()
}

// TrackTurn increments the session turn count.
func (ms *MemoryStore) TrackTurn() {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.turnCount++
}

// TrackTool records that a tool was used in this session.
func (ms *MemoryStore) TrackTool(name string) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.toolsUsed[name] = true
}

// TrackTopic records a topic discussed in this session.
func (ms *MemoryStore) TrackTopic(topic string) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.topics[topic] = true
}

// SaveSessionSummary generates a session summary from the conversation history
// and appends it to history.jsonl.
// SaveSessionSummary generates a session summary from the conversation history
// and appends it to history.jsonl.
func (ms *MemoryStore) SaveSessionSummary(history []llm.Message) error {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if ms.turnCount == 0 && len(history) == 0 {
		return nil // Nothing happened in this session
	}

	// Build summary deterministically from first user message + topics + tools
	summary := ms.buildSessionSummary(history)

	// Collect topics
	topics := make([]string, 0, len(ms.topics))
	for t := range ms.topics {
		topics = append(topics, t)
	}

	// Collect tools used
	toolsList := make([]string, 0, len(ms.toolsUsed))
	for t := range ms.toolsUsed {
		toolsList = append(toolsList, t)
	}

	entry := SessionEntry{
		SessionID: ms.sessionID,
		StartTime: ms.startTime.Format(time.RFC3339),
		EndTime:   time.Now().Format(time.RFC3339),
		Summary:   summary,
		Topics:    topics,
		ToolsUsed: toolsList,
		TurnCount: ms.turnCount,
	}

	// Append to history.jsonl
	historyPath := filepath.Join(ms.zapDir, "history.jsonl")
	f, err := os.OpenFile(historyPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open history.jsonl: %w", err)
	}
	defer f.Close()

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal session entry: %w", err)
	}

	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("failed to write session entry: %w", err)
	}
	if _, err := f.Write([]byte("\n")); err != nil {
		return fmt.Errorf("failed to write newline: %w", err)
	}

	return nil
}

// GetRecentSessions reads the last N sessions from history.jsonl.
func (ms *MemoryStore) GetRecentSessions(n int) []SessionEntry {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.getRecentSessionsUnlocked(n)
}

// getRecentSessionsUnlocked reads sessions without acquiring the lock (caller must hold it).
func (ms *MemoryStore) getRecentSessionsUnlocked(n int) []SessionEntry {
	historyPath := filepath.Join(ms.zapDir, "history.jsonl")
	f, err := os.Open(historyPath)
	if err != nil {
		return nil
	}
	defer f.Close()

	var all []SessionEntry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var entry SessionEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue // Skip malformed lines
		}
		all = append(all, entry)
	}

	if len(all) <= n {
		return all
	}
	return all[len(all)-n:]
}

// buildSessionSummary creates a compact summary from conversation history.
// Deterministic (no LLM call): extracts first user message + topics + tools.
func (ms *MemoryStore) buildSessionSummary(history []llm.Message) string {
	// Find first user message
	firstMsg := ""
	for _, msg := range history {
		if msg.Role == "user" && !strings.HasPrefix(msg.Content, "Observation:") {
			firstMsg = msg.Content
			break
		}
	}

	// Truncate first message for summary
	if len(firstMsg) > 80 {
		firstMsg = firstMsg[:80] + "..."
	}

	var parts []string
	if firstMsg != "" {
		parts = append(parts, firstMsg)
	}

	// Add topic info
	if len(ms.topics) > 0 {
		topicList := make([]string, 0, len(ms.topics))
		for t := range ms.topics {
			topicList = append(topicList, t)
		}
		parts = append(parts, fmt.Sprintf("topics: %s", strings.Join(topicList, ", ")))
	}

	// Add tool count
	if len(ms.toolsUsed) > 0 {
		parts = append(parts, fmt.Sprintf("used %d tools", len(ms.toolsUsed)))
	}

	if len(parts) == 0 {
		return "Empty session"
	}

	return strings.Join(parts, "; ")
}

// loadMemory reads memory.json from disk, handling both old ({}) and new (versioned) formats.
func (ms *MemoryStore) loadMemory() {
	memPath := filepath.Join(ms.zapDir, "memory.json")
	data, err := os.ReadFile(memPath)
	if err != nil {
		return // File doesn't exist yet
	}

	// Try new versioned format first
	var mf memoryFile
	if err := json.Unmarshal(data, &mf); err == nil && mf.Version > 0 {
		ms.entries = mf.Entries
		return
	}

	// Handle old empty {} format - start fresh
	ms.entries = []MemoryEntry{}
}

// saveMemory writes memory entries to memory.json (must be called with lock held).
func (ms *MemoryStore) saveMemory() error {
	mf := memoryFile{
		Version: 1,
		Entries: ms.entries,
	}

	data, err := json.MarshalIndent(mf, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal memory: %w", err)
	}

	memPath := filepath.Join(ms.zapDir, "memory.json")
	return os.WriteFile(memPath, data, 0644)
}
