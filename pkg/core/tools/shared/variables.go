package shared

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// VariableStore manages session and global variables
type VariableStore struct {
	session map[string]string // In-memory session variables
	global  map[string]string // Persistent global variables
	mu      sync.RWMutex
	falconDir string // Path to .falcon directory
}

// NewVariableStore creates a new variable store
func NewVariableStore(falconDir string) *VariableStore {
	store := &VariableStore{
		session: make(map[string]string),
		global:  make(map[string]string),
		falconDir: falconDir,
	}
	store.loadGlobalVariables()
	return store
}

// Set stores a variable (default: session scope)
func (vs *VariableStore) Set(name, value string) {
	vs.mu.Lock()
	defer vs.mu.Unlock()
	vs.session[name] = value
}

// SetGlobal stores a global variable (persisted to disk)
// Warns if the value appears to be a secret (should use session scope instead)
func (vs *VariableStore) SetGlobal(name, value string) (warning string, err error) {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	// Warn on potential secrets
	// Warn on potential secrets
	if IsSecret(name, value) {
		warning = fmt.Sprintf("WARNING: '%s' appears to be a secret. Consider using session scope instead (secrets are cleared on exit for security).", name)
	}

	vs.global[name] = value
	return warning, vs.saveGlobalVariables()
}

// Get retrieves a variable (checks session first, then global)
func (vs *VariableStore) Get(name string) (string, bool) {
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	// Check session first
	if value, ok := vs.session[name]; ok {
		return value, true
	}

	// Then check global
	if value, ok := vs.global[name]; ok {
		return value, true
	}

	return "", false
}

// Delete removes a variable
func (vs *VariableStore) Delete(name string) {
	vs.mu.Lock()
	defer vs.mu.Unlock()
	delete(vs.session, name)
	delete(vs.global, name)
	vs.saveGlobalVariables()
}

// List returns all variables (session + global)
func (vs *VariableStore) List() map[string]string {
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	result := make(map[string]string)
	// Global first
	for k, v := range vs.global {
		result[k] = v + " (global)"
	}
	// Session overrides global
	for k, v := range vs.session {
		result[k] = v + " (session)"
	}
	return result
}

// Substitute replaces {{VAR}} placeholders in text with variable values
func (vs *VariableStore) Substitute(text string) string {
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	result := text
	// Replace session variables
	for name, value := range vs.session {
		placeholder := "{{" + name + "}}"
		result = strings.ReplaceAll(result, placeholder, value)
	}
	// Replace global variables
	for name, value := range vs.global {
		placeholder := "{{" + name + "}}"
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

// loadGlobalVariables reads global variables from disk
func (vs *VariableStore) loadGlobalVariables() error {
	varFile := filepath.Join(vs.falconDir, "variables.json")
	data, err := os.ReadFile(varFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet, that's ok
		}
		return err
	}

	return json.Unmarshal(data, &vs.global)
}

// saveGlobalVariables writes global variables to disk
func (vs *VariableStore) saveGlobalVariables() error {
	varFile := filepath.Join(vs.falconDir, "variables.json")
	data, err := json.MarshalIndent(vs.global, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(varFile, data, 0644); err != nil {
		return err
	}

	// Update manifest counts
	UpdateManifestCounts(vs.falconDir)
	return nil
}

