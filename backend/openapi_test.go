package main

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestOpenAPISpecValidates(t *testing.T) {
	specPath := filepath.Join("docs", "openapi.yaml")

	loader := &openapi3.Loader{
		IsExternalRefsAllowed: true,
	}

	doc, err := loader.LoadFromFile(specPath)
	if err != nil {
		t.Fatalf("failed to load OpenAPI document (%s): %v", specPath, err)
	}

	if err := doc.Validate(context.Background()); err != nil {
		t.Fatalf("OpenAPI document validation failed: %v", err)
	}
}
