package dependency_mapper

import (
	"fmt"
	"strings"

	"github.com/blackcoderx/zap/pkg/core/tools/shared"
	"github.com/blackcoderx/zap/pkg/core/tools/spec_ingester"
)

// DependencyMapperTool maps relationships between API endpoints.
type DependencyMapperTool struct {
	zapDir string
}

// NewDependencyMapperTool creates a new dependency mapper tool.
func NewDependencyMapperTool(zapDir string) *DependencyMapperTool {
	return &DependencyMapperTool{
		zapDir: zapDir,
	}
}

// MapResult represents the resource dependency graph.
type MapResult struct {
	Dependencies []Dependency `json:"dependencies"`
	Summary      string       `json:"summary"`
}

// Dependency represents a relationship where one endpoint depends on another.
type Dependency struct {
	FromNode string `json:"from_endpoint"`
	ToNode   string `json:"to_endpoint"`
	Type     string `json:"dependency_type"` // creates_resource, requires_resource
	Resource string `json:"resource"`        // e.g. "user_id"
}

func (t *DependencyMapperTool) Name() string {
	return "map_dependencies"
}

func (t *DependencyMapperTool) Description() string {
	return "Map logical dependencies between API endpoints by analyzing resource creation and consumption patterns (e.g., POST /users creates ID used by GET /users/{id})"
}

func (t *DependencyMapperTool) Parameters() string {
	return "{}"
}

func (t *DependencyMapperTool) Execute(_ string) (string, error) {
	// 1. Load Knowledge Graph
	builder := spec_ingester.NewGraphBuilder(t.zapDir)
	graph, err := builder.LoadGraph()
	if err != nil {
		return "", fmt.Errorf("failed to load API Knowledge Graph: %w", err)
	}

	// 2. Map dependencies
	// This is a complex logic that would look for:
	// - Shared parameter names (e.g. "userId" in one, "userId" in another)
	// - Response schemas vs Request schemas

	dependencies := t.performMapping(graph)

	result := MapResult{
		Dependencies: dependencies,
	}
	result.Summary = t.formatSummary(result)

	return result.Summary, nil
}

func (t *DependencyMapperTool) performMapping(graph *shared.APIKnowledgeGraph) []Dependency {
	var deps []Dependency

	// Simplified logic for simulation
	for epKey, analysis := range graph.Endpoints {
		if strings.HasPrefix(epKey, "POST") {
			// Find corresponding GET/PUT/DELETE
			resourceName := t.inferResourceName(epKey)
			for otherKey := range graph.Endpoints {
				if otherKey != epKey && strings.Contains(otherKey, "{"+resourceName+"}") {
					deps = append(deps, Dependency{
						FromNode: epKey,
						ToNode:   otherKey,
						Type:     "provides_identifier",
						Resource: resourceName,
					})
				}
			}
		}
		_ = analysis
	}

	return deps
}

func (t *DependencyMapperTool) inferResourceName(epKey string) string {
	parts := strings.Split(epKey, "/")
	last := parts[len(parts)-1]
	// if plural, try to singularize
	if strings.HasSuffix(last, "s") {
		return last[:len(last)-1] + "Id"
	}
	return last + "Id"
}

func (t *DependencyMapperTool) formatSummary(r MapResult) string {
	summary := "🔗 API Dependency Map\n\n"
	summary += fmt.Sprintf("Identified Relationships: %d\n\n", len(r.Dependencies))

	for _, d := range r.Dependencies {
		summary += fmt.Sprintf("  • %s → %s (via %s)\n", d.FromNode, d.ToNode, d.Resource)
	}

	if len(r.Dependencies) == 0 {
		summary += "No clear resource dependencies identified in the current Knowledge Graph."
	}

	return summary
}
