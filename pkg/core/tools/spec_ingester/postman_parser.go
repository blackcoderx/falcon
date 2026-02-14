package spec_ingester

import (
	"fmt"
	"strings"

	"github.com/rbretecher/go-postman-collection"
)

// PostmanParser implements the SpecParser for Postman Collections v2.1
type PostmanParser struct{}

func (p *PostmanParser) DetectFormat(content []byte) bool {
	// Simple heuristic: check for "postman" and "info"
	s := string(content)
	return strings.Contains(s, "_postman_id") || (strings.Contains(s, "info") && strings.Contains(s, "schema"))
}

func (p *PostmanParser) Parse(content []byte) (*ParsedSpec, error) {
	// Parse the document
	r := strings.NewReader(string(content))
	collection, err := postman.ParseCollection(r)
	if err != nil {
		return nil, fmt.Errorf("failed to parse postman collection: %w", err)
	}

	spec := &ParsedSpec{
		Format:  "postman2.1",
		Version: collection.Info.Version,
	}

	// Iterate items (Postman structure is recursive: Folders -> Items -> Requests)
	p.processItems(collection.Items, spec)

	return spec, nil
}

func (p *PostmanParser) processItems(items []*postman.Items, spec *ParsedSpec) {
	for _, item := range items {
		if item.IsGroup() {
			p.processItems(item.Items, spec)
			continue
		}

		if item.Request != nil {
			req := item.Request
			endpoint := ParsedEndpoint{
				Method:      string(req.Method),
				Summary:     item.Name,
				Description: item.Description, // postman desc is often a struct, but library might stringify
			}

			// Extract URL
			if req.URL != nil {
				endpoint.Path = req.URL.Raw // Use raw URL for now
			}

			// Extract Body
			if req.Body != nil {
				endpoint.HasBody = true
			}

			// Extract Headers as Parameters
			for _, h := range req.Header {
				endpoint.Parameters = append(endpoint.Parameters, ParsedParameter{
					Name:     h.Key,
					In:       "header",
					Required: false, // headers in postman are usually optional toggle-able
					Type:     "string",
				})
			}

			// Extract Query Params
			if req.URL != nil {
				for _, q := range req.URL.Query {
					endpoint.Parameters = append(endpoint.Parameters, ParsedParameter{
						Name:     q.Key,
						In:       "query",
						Required: false,
						Type:     "string",
					})
				}
			}

			spec.Endpoints = append(spec.Endpoints, endpoint)
		}
	}
}
