package main

import (
	"bytes"
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"socialpredict/handlers"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestOpenAPISpecValidates(t *testing.T) {
	doc := loadOpenAPIDocument(t)
	if err := doc.Validate(context.Background()); err != nil {
		t.Fatalf("OpenAPI document validation failed: %v", err)
	}
}

func TestEmbeddedOpenAPISpecMatchesDocument(t *testing.T) {
	specPath := filepath.Join("docs", "openapi.yaml")

	specBytes, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read OpenAPI document (%s): %v", specPath, err)
	}

	if !bytes.Equal(openAPISpec, specBytes) {
		t.Fatalf("embedded OpenAPI spec does not match %s", specPath)
	}
}

func TestEmbeddedSwaggerUIAssetsPresent(t *testing.T) {
	requiredAssets := []string{
		"swagger-ui/index.html",
		"swagger-ui/swagger-initializer.js",
		"swagger-ui/swagger-ui.css",
	}

	for _, assetPath := range requiredAssets {
		t.Run(assetPath, func(t *testing.T) {
			info, err := fs.Stat(swaggerUIFS, assetPath)
			if err != nil {
				t.Fatalf("missing embedded swagger asset %s: %v", assetPath, err)
			}
			if info.IsDir() {
				t.Fatalf("expected swagger asset %s to be a file", assetPath)
			}
		})
	}
}

func TestOpenAPIOperationsMatchServerRoutes(t *testing.T) {
	doc := loadOpenAPIDocument(t)

	documentedOperations := openAPIOperationKeys(doc)
	implementedOperations := loadServerOperationKeys(t, filepath.Join("server", "server.go"))

	for _, routeKey := range []string{
		"GET /v0/marketprojection/{marketId}/{amount}/{outcome}/",
	} {
		delete(implementedOperations, routeKey)
	}

	if missingFromSpec := setDifference(implementedOperations, documentedOperations); len(missingFromSpec) > 0 {
		t.Fatalf("implemented routes missing from OpenAPI: %s", strings.Join(missingFromSpec, ", "))
	}
	if staleInSpec := setDifference(documentedOperations, implementedOperations); len(staleInSpec) > 0 {
		t.Fatalf("OpenAPI operations missing from server routes: %s", strings.Join(staleInSpec, ", "))
	}
}

func TestReasonResponseEnumMatchesSharedFailureVocabulary(t *testing.T) {
	doc := loadOpenAPIDocument(t)

	reasonSchemaRef := doc.Components.Schemas["ReasonResponse"]
	if reasonSchemaRef == nil || reasonSchemaRef.Value == nil {
		t.Fatal("components.schemas.ReasonResponse missing from OpenAPI document")
	}

	reasonProperty := reasonSchemaRef.Value.Properties["reason"]
	if reasonProperty == nil || reasonProperty.Value == nil {
		t.Fatal("components.schemas.ReasonResponse.properties.reason missing from OpenAPI document")
	}

	got := make([]string, 0, len(reasonProperty.Value.Enum))
	for _, value := range reasonProperty.Value.Enum {
		reason, ok := value.(string)
		if !ok {
			t.Fatalf("ReasonResponse reason enum contains non-string value: %#v", value)
		}
		got = append(got, reason)
	}

	want := make([]string, 0, len(handlers.PublicFailureReasons()))
	for _, reason := range handlers.PublicFailureReasons() {
		want = append(want, string(reason))
	}

	sort.Strings(got)
	sort.Strings(want)

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ReasonResponse enum mismatch\nwant: %v\ngot:  %v", want, got)
	}
}

func loadOpenAPIDocument(t *testing.T) *openapi3.T {
	t.Helper()

	specPath := filepath.Join("docs", "openapi.yaml")
	loader := &openapi3.Loader{IsExternalRefsAllowed: true}

	doc, err := loader.LoadFromFile(specPath)
	if err != nil {
		t.Fatalf("failed to load OpenAPI document (%s): %v", specPath, err)
	}

	return doc
}

func openAPIOperationKeys(doc *openapi3.T) map[string]struct{} {
	keys := make(map[string]struct{})
	for path, pathItem := range doc.Paths.Map() {
		for method := range pathItem.Operations() {
			keys[routeKey(method, path)] = struct{}{}
		}
	}
	return keys
}

func loadServerOperationKeys(t *testing.T, serverPath string) map[string]struct{} {
	t.Helper()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, serverPath, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", serverPath, err)
	}

	keys := make(map[string]struct{})
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		methodsSelector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || methodsSelector.Sel == nil || methodsSelector.Sel.Name != "Methods" {
			return true
		}

		routeCall, ok := methodsSelector.X.(*ast.CallExpr)
		if !ok {
			return true
		}

		routeSelector, ok := routeCall.Fun.(*ast.SelectorExpr)
		if !ok || routeSelector.Sel == nil {
			return true
		}
		if routeSelector.Sel.Name != "Handle" && routeSelector.Sel.Name != "HandleFunc" {
			return true
		}
		if len(routeCall.Args) == 0 {
			return true
		}

		path, ok := stringLiteral(routeCall.Args[0])
		if !ok {
			return true
		}

		for _, arg := range call.Args {
			method, ok := stringLiteral(arg)
			if !ok {
				continue
			}
			keys[routeKey(method, path)] = struct{}{}
		}

		return true
	})

	return keys
}

func stringLiteral(expr ast.Expr) (string, bool) {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}

	value, err := strconv.Unquote(lit.Value)
	if err != nil {
		return "", false
	}

	return value, true
}

func routeKey(method, path string) string {
	return strings.ToUpper(method) + " " + path
}

func setDifference(left, right map[string]struct{}) []string {
	var diff []string
	for key := range left {
		if _, ok := right[key]; ok {
			continue
		}
		diff = append(diff, key)
	}
	sort.Strings(diff)
	return diff
}
