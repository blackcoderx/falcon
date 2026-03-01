package llm

import "fmt"

var (
	providerRegistry []Provider
	providerMap      = map[string]Provider{}
)

// Register adds a provider to the global registry.
// Call this from an init() function in register_providers.go.
// Panics if a provider with the same ID is registered twice.
func Register(p Provider) {
	if _, exists := providerMap[p.ID()]; exists {
		panic(fmt.Sprintf("llm: provider %q already registered", p.ID()))
	}
	providerRegistry = append(providerRegistry, p)
	providerMap[p.ID()] = p
}

// Get returns the provider with the given ID, or (nil, false) if not found.
func Get(id string) (Provider, bool) {
	p, ok := providerMap[id]
	return p, ok
}

// All returns all registered providers in registration order.
func All() []Provider {
	return providerRegistry
}
