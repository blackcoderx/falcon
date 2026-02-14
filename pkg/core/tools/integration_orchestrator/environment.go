package integration_orchestrator

import (
	"strings"
)

// Environment manages the context and state across workflow steps.
type Environment struct {
	BaseURL string
	State   map[string]interface{}
}

// NewEnvironment creates a new environment context.
func NewEnvironment(baseURL string) *Environment {
	return &Environment{
		BaseURL: baseURL,
		State:   make(map[string]interface{}),
	}
}

// ResolveURL combines the base URL with a step path.
func (e *Environment) ResolveURL(path string) string {
	if strings.HasPrefix(path, "http") {
		return path
	}

	base := strings.TrimSuffix(e.BaseURL, "/")
	p := strings.TrimPrefix(path, "/")

	if base == "" {
		return p
	}
	return base + "/" + p
}

// Set stores a value in the shared state.
func (e *Environment) Set(key string, value interface{}) {
	e.State[key] = value
}

// Get retrieves a value from the shared state.
func (e *Environment) Get(key string) interface{} {
	return e.State[key]
}
