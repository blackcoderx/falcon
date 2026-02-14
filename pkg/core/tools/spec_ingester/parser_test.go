package spec_ingester

import (
	"os"
	"testing"
)

// Minimal test to verify compilation and basic function references
func TestGraphBuilder(t *testing.T) {
	zapDir := os.TempDir()
	builder := NewGraphBuilder(zapDir)
	if builder == nil {
		t.Fatal("NewGraphBuilder returned nil")
	}
}

func TestParserDetection(t *testing.T) {
	openapi := &OpenAPIParser{}
	postman := &PostmanParser{}

	oaContent := []byte(`openapi: 3.0.0`)
	pmContent := []byte(`{"info": {"schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"}}`)

	if !openapi.DetectFormat(oaContent) {
		t.Error("OpenAPI parser failed to detect openapi content")
	}

	if !postman.DetectFormat(pmContent) {
		t.Error("Postman parser failed to detect postman content")
	}
}
