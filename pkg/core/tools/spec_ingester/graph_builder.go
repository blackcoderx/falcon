package spec_ingester

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/blackcoderx/falcon/pkg/core/tools/shared"
)

// GraphBuilder is responsible for transforming intermediate ParsedSpec into
// the final APIKnowledgeGraph used by Falcon
type GraphBuilder struct {
	ZapDir string
}

func NewGraphBuilder(zapDir string) *GraphBuilder {
	return &GraphBuilder{
		ZapDir: zapDir,
	}
}

// BuildGraph takes parsing results and "fuses" them into a single graph
func (b *GraphBuilder) BuildGraph(spec *ParsedSpec, context shared.ProjectContext) (*shared.APIKnowledgeGraph, error) {
	graph := &shared.APIKnowledgeGraph{
		Endpoints: make(map[string]shared.EndpointAnalysis),
		Models:    make(map[string]shared.ModelDefinition),
		Context:   context,
		Version:   spec.Version,
	}

	for _, endpoint := range spec.Endpoints {
		uniqueID := fmt.Sprintf("%s %s", endpoint.Method, endpoint.Path)

		// Map ParsedEndpoint to shared.EndpointAnalysis
		analysis := shared.EndpointAnalysis{
			Summary:    endpoint.Summary,
			Parameters: b.mapParameters(endpoint.Parameters),
			Responses:  b.mapResponses(endpoint.Responses),
			// AuthType and Security will need enrichment from framework scan later
		}

		graph.Endpoints[uniqueID] = analysis
	}

	// TODO: Map Models (Schemas) from spec once we implement schema extraction deep-dive

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

// SaveGraph persists the graph to .zap/api_graph.json
func (b *GraphBuilder) SaveGraph(graph *shared.APIKnowledgeGraph) error {
	data, err := json.MarshalIndent(graph, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal graph: %w", err)
	}

	path := fmt.Sprintf("%s/api_graph.json", b.ZapDir)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to save API graph: %w", err)
	}
	return nil
}

// LoadGraph loads the graph from .zap/api_graph.json
func (b *GraphBuilder) LoadGraph() (*shared.APIKnowledgeGraph, error) {
	path := fmt.Sprintf("%s/api_graph.json", b.ZapDir)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // Return nil if no graph exists yet
		}
		return nil, fmt.Errorf("failed to read API graph: %w", err)
	}

	var graph shared.APIKnowledgeGraph
	if err := json.Unmarshal(data, &graph); err != nil {
		return nil, fmt.Errorf("failed to parse API graph: %w", err)
	}

	return &graph, nil
}
