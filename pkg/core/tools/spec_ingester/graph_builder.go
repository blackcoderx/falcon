package spec_ingester

import (
	"fmt"
	"os"

	"github.com/blackcoderx/falcon/pkg/core/tools/shared"
	"gopkg.in/yaml.v3"
)

// GraphBuilder is responsible for transforming intermediate ParsedSpec into
// the final APIKnowledgeGraph used by Falcon.
type GraphBuilder struct {
	FalconDir string
}

func NewGraphBuilder(falconDir string) *GraphBuilder {
	return &GraphBuilder{FalconDir: falconDir}
}

// BuildGraph takes parsing results and fuses them into a single graph.
func (b *GraphBuilder) BuildGraph(spec *ParsedSpec, context shared.ProjectContext) (*shared.APIKnowledgeGraph, error) {
	graph := &shared.APIKnowledgeGraph{
		Endpoints: make(map[string]shared.EndpointAnalysis),
		Models:    make(map[string]shared.ModelDefinition),
		Context:   context,
		Version:   spec.Version,
	}

	for _, endpoint := range spec.Endpoints {
		uniqueID := fmt.Sprintf("%s %s", endpoint.Method, endpoint.Path)
		graph.Endpoints[uniqueID] = shared.EndpointAnalysis{
			Summary:    endpoint.Summary,
			Parameters: b.mapParameters(endpoint.Parameters),
			Responses:  b.mapResponses(endpoint.Responses),
		}
	}

	return graph, nil
}

func (b *GraphBuilder) mapParameters(params []ParsedParameter) []shared.Parameter {
	var result []shared.Parameter
	for _, p := range params {
		result = append(result, shared.Parameter{
			Name:        p.Name,
			Type:        p.Type,
			Required:    p.Required,
			Description: fmt.Sprintf("in: %s", p.In),
		})
	}
	return result
}

func (b *GraphBuilder) mapResponses(codes []int) []shared.Response {
	var result []shared.Response
	for _, c := range codes {
		result = append(result, shared.Response{
			StatusCode:  c,
			Description: "Derived from spec",
		})
	}
	return result
}

// SaveGraph persists the graph to .falcon/spec.yaml (human-readable YAML).
func (b *GraphBuilder) SaveGraph(graph *shared.APIKnowledgeGraph) error {
	data, err := yaml.Marshal(graph)
	if err != nil {
		return fmt.Errorf("failed to marshal graph to YAML: %w", err)
	}

	path := fmt.Sprintf("%s/spec.yaml", b.FalconDir)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to save API spec: %w", err)
	}
	return nil
}

// LoadGraph loads the graph from .falcon/spec.yaml.
func (b *GraphBuilder) LoadGraph() (*shared.APIKnowledgeGraph, error) {
	path := fmt.Sprintf("%s/spec.yaml", b.FalconDir)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // Return nil if no spec has been ingested yet
		}
		return nil, fmt.Errorf("failed to read spec.yaml: %w", err)
	}

	var graph shared.APIKnowledgeGraph
	if err := yaml.Unmarshal(data, &graph); err != nil {
		return nil, fmt.Errorf("failed to parse spec.yaml: %w", err)
	}

	return &graph, nil
}
