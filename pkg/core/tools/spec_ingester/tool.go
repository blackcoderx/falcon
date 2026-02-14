package spec_ingester

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/blackcoderx/zap/pkg/core/tools/shared"
	"github.com/blackcoderx/zap/pkg/llm"
)

// IngestSpecTool provides commands to index API specifications
type IngestSpecTool struct {
	llmClient llm.LLMClient
	zapDir    string
}

// NewIngestSpecTool creates a new spec ingestion tool
func NewIngestSpecTool(llmClient llm.LLMClient, zapDir string) *IngestSpecTool {
	return &IngestSpecTool{
		llmClient: llmClient,
		zapDir:    zapDir,
	}
}

// IngestParams checks inputs for file path or URL
type IngestParams struct {
	Action string `json:"action"` // "index", "update", "status"
	Source string `json:"source"` // file path or URL
}

func (t *IngestSpecTool) Name() string {
	return "ingest_spec"
}

func (t *IngestSpecTool) Description() string {
	return "Ingest API specifications (OpenAPI/Swagger/Postman) to build a Knowledge Graph for autonomous testing. Use 'index' to start a fresh scan."
}

func (t *IngestSpecTool) Parameters() string {
	return `{
  "action": "index",
  "source": "./docs/openapi.yaml"
}`
}

func (t *IngestSpecTool) Execute(args string) (string, error) {
	var params IngestParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("invalid params: %w", err)
	}

	if params.Action == "status" {
		// Just check if we have a graph
		builder := NewGraphBuilder(t.zapDir)
		graph, err := builder.LoadGraph()
		if err != nil {
			return "No API Knowledge Graph found. Use 'index' to create one.", nil
		}
		if graph == nil {
			return "API Knowledge Graph is empty.", nil
		}
		return fmt.Sprintf("API Knowledge Graph v%s contains %d endpoints.", graph.Version, len(graph.Endpoints)), nil
	}

	if params.Source == "" {
		return "", fmt.Errorf("source is required for index action")
	}

	// 1. Fetch Content
	content, err := t.fetchContent(params.Source)
	if err != nil {
		return "", fmt.Errorf("failed to read source: %w", err)
	}

	// 2. Detect & Parse
	var parser SpecParser
	openapi := &OpenAPIParser{}
	postman := &PostmanParser{}

	if openapi.DetectFormat(content) {
		parser = openapi
	} else if postman.DetectFormat(content) {
		parser = postman
	} else {
		return "", fmt.Errorf("unsupported spec format")
	}

	parsedSpec, err := parser.Parse(content)
	if err != nil {
		return "", fmt.Errorf("parsing failed: %w", err)
	}

	// 3. Build Graph
	// TODO: Integrate framework scanning here (Sprint 4 Step 2)
	// For now, use empty context or inferred from file ext?
	ctx := shared.ProjectContext{
		SpecPath: params.Source,
	}

	builder := NewGraphBuilder(t.zapDir)
	graph, err := builder.BuildGraph(parsedSpec, ctx)
	if err != nil {
		return "", fmt.Errorf("graph build failed: %w", err)
	}

	// 4. Save
	if err := builder.SaveGraph(graph); err != nil {
		return "", fmt.Errorf("failed to save graph: %w", err)
	}

	return fmt.Sprintf("Successfully indexed API from %s. Found %d endpoints.", params.Source, len(graph.Endpoints)), nil
}

func (t *IngestSpecTool) fetchContent(source string) ([]byte, error) {
	if strings.HasPrefix(source, "http") {
		resp, err := http.Get(source)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		return io.ReadAll(resp.Body)
	}
	return os.ReadFile(source)
}
