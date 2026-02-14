package spec_ingester

import (
	"fmt"
	"strings"

	"github.com/pb33f/libopenapi"
	validator "github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// OpenAPIParser implements the SpecParser for OpenAPI 3.x and Swagger 2.0
type OpenAPIParser struct{}

func (p *OpenAPIParser) DetectFormat(content []byte) bool {
	// Simple heuristic: check for "openapi" or "swagger" strings
	s := string(content)
	return strings.Contains(s, "openapi") || strings.Contains(s, "swagger")
}

func (p *OpenAPIParser) Parse(content []byte) (*ParsedSpec, error) {
	// Parse the document
	document, err := libopenapi.NewDocument(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse spec document: %w", err)
	}

	// We primarily target V3 for now
	model, err := document.BuildV3Model()
	if err != nil {
		// If V3 fails, could try V2, but for now let's just return the error
		return nil, fmt.Errorf("failed to build OpenAPI v3 model: %w", err)
	}

	spec := &ParsedSpec{
		Format:  "openapi3",
		Version: model.Model.Info.Version,
	}

	// Iterate Paths using ordered map iterator
	for pair := model.Model.Paths.PathItems.First(); pair != nil; pair = pair.Next() {
		path := pair.Key()
		pathItem := pair.Value()

		ops := map[string]*v3.Operation{
			"GET":    pathItem.Get,
			"POST":   pathItem.Post,
			"PUT":    pathItem.Put,
			"DELETE": pathItem.Delete,
			"PATCH":  pathItem.Patch,
		}

		for method, op := range ops {
			if op == nil {
				continue
			}

			endpoint := ParsedEndpoint{
				Method:      method,
				Path:        path,
				Summary:     op.Summary,
				Description: op.Description,
				HasBody:     op.RequestBody != nil,
			}

			// Extract Parameters
			for _, param := range p.resolveParameters(op.Parameters) {
				endpoint.Parameters = append(endpoint.Parameters, ParsedParameter{
					Name:     param.Name,
					In:       param.In,
					Required: param.Required != nil && *param.Required,
					Type:     p.extractType(param.Schema),
				})
			}

			// Extract Response Codes
			for pair := op.Responses.Codes.First(); pair != nil; pair = pair.Next() {
				status := pair.Key()
				// val := pair.Value()

				var code int
				if n, err := fmt.Sscanf(status, "%d", &code); err == nil && n == 1 {
					endpoint.Responses = append(endpoint.Responses, code)
				}
			}

			spec.Endpoints = append(spec.Endpoints, endpoint)
		}
	}

	return spec, nil
}

func (p *OpenAPIParser) resolveParameters(params []*v3.Parameter) []*v3.Parameter {
	// Simplified: libopenapi high-level model should already resolve refs if configured,
	// but here we just iterate provided slice.
	// For a more robust implementation we'd handle shared params at PathItem level too.
	return params
}

func (p *OpenAPIParser) extractType(schema *validator.SchemaProxy) string {
	if schema == nil || schema.Schema() == nil {
		return "unknown"
	}
	s := schema.Schema()
	if len(s.Type) > 0 {
		return s.Type[0]
	}
	return "object"
}
