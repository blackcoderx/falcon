package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/blackcoderx/zap/pkg/core/tools/shared"
)

// MemoryEntry represents a single fact saved by the agent.
type MemoryEntry struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	Category  string `json:"category"`  // "preference", "endpoint", "error", "project", "general"
	Timestamp string `json:"timestamp"` // RFC3339
	Source    string `json:"source"`    // Session ID that created this
}

// memoryFile is the on-disk format of memory.json.
type memoryFile struct {
	Version int           `json:"version"`
	Entries []MemoryEntry `json:"entries"`
}

// MemoryStore manages persistent agent memory.
type MemoryStore struct {
	entries []MemoryEntry
	mu      sync.RWMutex
	zapDir  string
}

// NewMemoryStore creates a MemoryStore and loads existing memory.
func NewMemoryStore(zapDir string) *MemoryStore {
	ms := &MemoryStore{
		zapDir: zapDir,
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
		Source:    "",
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
// Returns empty string if no knowledge base content or memories exist.
func (ms *MemoryStore) GetCompactSummary() string {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	var sb strings.Builder
	hasContent := false

	// Inject falcon.md knowledge base
	falconPath := filepath.Join(ms.zapDir, "falcon.md")
	if falconData, err := os.ReadFile(falconPath); err == nil && len(falconData) > 0 {
		content := string(falconData)
		if strings.Contains(content, "##") {
			sb.WriteString("## API KNOWLEDGE BASE (falcon.md)\n")
			sb.WriteString("Use memory({\"action\":\"update_knowledge\", \"section\":\"...\", \"content\":\"...\"}) to update sections as you learn new API facts.\n\n")
			sb.WriteString(content)
			sb.WriteString("\n\n")
			hasContent = true
		}
	}

	// Remembered facts from memory.json
	if len(ms.entries) > 0 {
		sb.WriteString("## REMEMBERED FACTS\n")
		for _, e := range ms.entries {
			sb.WriteString(fmt.Sprintf("- [%s] %s: %s\n", e.Category, e.Key, e.Value))
		}
		sb.WriteString("\n")
		hasContent = true
	}

	if !hasContent {
		return ""
	}

	return sb.String()
}

// UpdateKnowledge rewrites a named section in falcon.md with new content.
// If the section heading does not exist, it is appended. If it exists,
// the content between that heading and the next H2 heading is replaced.
func (ms *MemoryStore) UpdateKnowledge(section, newContent string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	falconPath := filepath.Join(ms.zapDir, "falcon.md")
	data, err := os.ReadFile(falconPath)
	if err != nil {
		return fmt.Errorf("failed to read falcon.md: %w", err)
	}

	heading := "## " + section
	lines := strings.Split(string(data), "\n")
	sectionStart := -1
	for i, line := range lines {
		if strings.TrimRight(line, " ") == heading {
			sectionStart = i
			break
		}
	}

	var result string
	if sectionStart == -1 {
		// Append new section
		result = strings.TrimRight(string(data), "\n") + "\n\n" + heading + "\n\n" + strings.TrimSpace(newContent) + "\n"
	} else {
		// Find next ## heading
		sectionEnd := len(lines)
		for i := sectionStart + 1; i < len(lines); i++ {
			if strings.HasPrefix(lines[i], "## ") {
				sectionEnd = i
				break
			}
		}
		var sb strings.Builder
		for i := 0; i <= sectionStart; i++ {
			sb.WriteString(lines[i] + "\n")
		}
		sb.WriteString("\n" + strings.TrimSpace(newContent) + "\n")
		if sectionEnd < len(lines) {
			sb.WriteString("\n")
			for i := sectionEnd; i < len(lines); i++ {
				sb.WriteString(lines[i])
				if i < len(lines)-1 {
					sb.WriteString("\n")
				}
			}
		}
		result = sb.String()
	}

	return os.WriteFile(falconPath, []byte(result), 0644)
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
