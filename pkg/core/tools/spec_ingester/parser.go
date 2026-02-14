package spec_ingester

// SpecParser defines the interface for parsing API specifications
type SpecParser interface {
	// DetectFormat returns true if the parser can handle the given content
	DetectFormat(content []byte) bool

	// Parse converts the raw content into a unified APIKnowledgeGraph structure
	// It basically extracts endpoints and models but returns them in a generic map first
	// to be later processed by the GraphBuilder
	Parse(content []byte) (*ParsedSpec, error)
}

// ParsedSpec is an intermediate representation of a parsed API spec
type ParsedSpec struct {
	Format    string // "openapi3", "swagger2", "postman2.1"
	Version   string
	Endpoints []ParsedEndpoint
}

// ParsedEndpoint represents a single API operation found in the spec
type ParsedEndpoint struct {
	Method      string
	Path        string
	Summary     string
	Description string
	Parameters  []ParsedParameter
	// Simplified representation of request/response for initial indexing
	HasBody   bool
	Responses []int
}

type ParsedParameter struct {
	Name     string
	In       string // query, path, header, etc.
	Required bool
	Type     string
}
