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
	"socialpredict/security"

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

func TestEmbeddedSwaggerUIUsesCanonicalOpenAPIEndpoint(t *testing.T) {
	initializer, err := fs.ReadFile(swaggerUIFS, "swagger-ui/swagger-initializer.js")
	if err != nil {
		t.Fatalf("read embedded swagger initializer: %v", err)
	}

	if !bytes.Contains(initializer, []byte(`url: "/openapi.yaml"`)) {
		t.Fatalf("swagger initializer must load the backend-served /openapi.yaml contract: %s", initializer)
	}
}

func TestDocsPublishingProxyTemplatesExposeBackendDocs(t *testing.T) {
	templates := []string{
		filepath.Join("..", "data", "nginx", "vhosts", "dev", "default.conf.template"),
		filepath.Join("..", "data", "nginx", "vhosts", "prod", "default.conf.template"),
	}
	requiredSnippets := []string{
		"location = /openapi.yaml",
		"location = /swagger",
		"location /swagger/",
		"proxy_pass http://backend:8080;",
	}

	for _, template := range templates {
		t.Run(template, func(t *testing.T) {
			content, err := os.ReadFile(template)
			if err != nil {
				t.Fatalf("read proxy template: %v", err)
			}
			for _, snippet := range requiredSnippets {
				if !strings.Contains(string(content), snippet) {
					t.Fatalf("proxy template missing %q", snippet)
				}
			}
		})
	}

	traefikTemplate := filepath.Join("..", "data", "traefik", "config", "traefik.template")
	content, err := os.ReadFile(traefikTemplate)
	if err != nil {
		t.Fatalf("read traefik template: %v", err)
	}
	for _, snippet := range []string{"public host and TLS edge", "/swagger/", "/openapi.yaml", "explicit in nginx"} {
		if !strings.Contains(string(content), snippet) {
			t.Fatalf("traefik template missing docs publishing note %q", snippet)
		}
	}
}

func TestOpenAPIOperationsMatchServerRoutes(t *testing.T) {
	doc := loadOpenAPIDocument(t)

	documentedOperations := openAPIOperationKeys(doc)
	implementedOperations := loadServerOperationKeys(t, filepath.Join("server", "server.go"))

	if missingFromSpec := setDifference(implementedOperations, documentedOperations); len(missingFromSpec) > 0 {
		t.Fatalf("implemented routes missing from OpenAPI: %s", strings.Join(missingFromSpec, ", "))
	}
	if staleInSpec := setDifference(documentedOperations, implementedOperations); len(staleInSpec) > 0 {
		t.Fatalf("OpenAPI operations missing from server routes: %s", strings.Join(staleInSpec, ", "))
	}
}

func TestRouteFamilyMigrationMatrixMatchesTouchedContractSlice(t *testing.T) {
	doc := loadOpenAPIDocument(t)
	matrix := extensionMap(t, doc.Extensions, "x-route-family-migration-matrix")

	wantSourceOrder := []string{
		"backend/server/server.go",
		"touched handlers and DTOs under backend/handlers/**",
		"backend/docs/openapi.yaml",
		"backend/docs/API-ISSUES.md",
	}
	if got := stringSlice(t, matrix["source_of_truth_order"], "source_of_truth_order"); !reflect.DeepEqual(got, wantSourceOrder) {
		t.Fatalf("source_of_truth_order mismatch\nwant: %v\ngot:  %v", wantSourceOrder, got)
	}

	if got := matrix["plain_text_error_response_state"]; got != "migration_state_not_target_contract" {
		t.Fatalf("plain_text_error_response_state = %v, want migration_state_not_target_contract", got)
	}

	wantReasons := make([]string, 0, len(handlers.PublicFailureReasons()))
	for _, reason := range handlers.PublicFailureReasons() {
		wantReasons = append(wantReasons, string(reason))
	}
	if got := stringSlice(t, matrix["public_reason_values"], "public_reason_values"); !reflect.DeepEqual(got, wantReasons) {
		t.Fatalf("public_reason_values mismatch\nwant: %v\ngot:  %v", wantReasons, got)
	}

	expectedFamilies := map[string]struct{}{
		"infra-probes":              {},
		"infra-docs":                {},
		"runtime-middleware":        {},
		"auth":                      {},
		"setup":                     {},
		"reporting":                 {},
		"markets":                   {},
		"market-search":             {},
		"market-bets-and-positions": {},
		"public-users":              {},
		"private-users":             {},
		"private-actions":           {},
		"admin-and-content":         {},
	}
	implementedPaths := loadServerPaths(t, filepath.Join("server", "server.go"))
	implementedPaths["/swagger"] = struct{}{}
	implementedPaths["/swagger/"] = struct{}{}

	for _, rawFamily := range anySlice(t, matrix["route_families"], "route_families") {
		family := rawFamily.(map[string]any)
		name := family["family"].(string)
		if _, ok := expectedFamilies[name]; !ok {
			t.Fatalf("unexpected route family %q", name)
		}
		delete(expectedFamilies, name)

		if family["migration_state"] == "" {
			t.Fatalf("route family %q missing migration_state", name)
		}
		if family["success_contract"] == "" || family["failure_contract"] == "" {
			t.Fatalf("route family %q missing success or failure contract", name)
		}

		for _, path := range stringSlice(t, family["paths"], name+".paths") {
			if path == "all registered routes" {
				continue
			}
			if doc.Paths.Find(path) == nil {
				t.Fatalf("route family %q references path %q missing from OpenAPI paths", name, path)
			}
			if _, ok := implementedPaths[path]; !ok {
				t.Fatalf("route family %q references path %q missing from server routes", name, path)
			}
		}
	}

	if len(expectedFamilies) > 0 {
		t.Fatalf("missing expected route families: %v", sortedKeys(expectedFamilies))
	}
}

func TestPlainTextErrorResponseDocumentsMigrationState(t *testing.T) {
	doc := loadOpenAPIDocument(t)
	schemaRef := doc.Components.Schemas["PlainTextErrorResponse"]
	if schemaRef == nil || schemaRef.Value == nil {
		t.Fatal("components.schemas.PlainTextErrorResponse missing from OpenAPI document")
	}

	description := schemaRef.Value.Description
	for _, want := range []string{"migration-state", "not the target API contract", "ReasonResponse"} {
		if !strings.Contains(description, want) {
			t.Fatalf("PlainTextErrorResponse description missing %q: %s", want, description)
		}
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

func TestRuntimeFailureReasonsRemainInPublicOpenAPIVocabulary(t *testing.T) {
	doc := loadOpenAPIDocument(t)
	reasonEnum := reasonResponseEnum(t, doc)
	publicReasons := make(map[string]struct{})
	for _, reason := range handlers.PublicFailureReasons() {
		publicReasons[string(reason)] = struct{}{}
	}

	for _, reason := range []string{
		security.RuntimeReasonMethodNotAllowed,
		security.RuntimeReasonRateLimited,
		security.RuntimeReasonLoginRateLimited,
		security.RuntimeReasonInternalError,
	} {
		if _, ok := publicReasons[reason]; !ok {
			t.Fatalf("runtime reason %q missing from handlers.PublicFailureReasons", reason)
		}
		if _, ok := reasonEnum[reason]; !ok {
			t.Fatalf("runtime reason %q missing from OpenAPI ReasonResponse enum", reason)
		}
	}
}

func TestOpenAPIDocumentsPasswordChangeAuthExceptions(t *testing.T) {
	doc := loadOpenAPIDocument(t)

	changePassword := operationFor(t, doc, "/v0/changepassword", httpMethodPost)
	changePasswordDescription := changePassword.Description
	for _, want := range []string{"token-only authentication", "mustChangePassword"} {
		if !strings.Contains(changePasswordDescription, want) {
			t.Fatalf("/v0/changepassword description missing %q: %s", want, changePasswordDescription)
		}
	}
	if _, ok := changePassword.Responses.Map()["403"]; ok {
		t.Fatal("/v0/changepassword must not document PASSWORD_CHANGE_REQUIRED as a blocking 403 response")
	}

	for _, route := range []struct {
		path   string
		method string
	}{
		{path: "/v0/bet", method: httpMethodPost},
		{path: "/v0/sell", method: httpMethodPost},
		{path: "/v0/userposition/{marketId}", method: httpMethodGet},
	} {
		t.Run(route.path, func(t *testing.T) {
			operation := operationFor(t, doc, route.path, route.method)
			if !strings.Contains(operation.Description, "PASSWORD_CHANGE_REQUIRED") {
				t.Fatalf("%s description missing PASSWORD_CHANGE_REQUIRED: %s", route.path, operation.Description)
			}
			response := operation.Responses.Map()["403"]
			if response == nil || response.Value == nil {
				t.Fatalf("%s missing documented 403 password-change response", route.path)
			}
			if response.Value.Description == nil || !strings.Contains(*response.Value.Description, "Password change") {
				description := ""
				if response.Value.Description != nil {
					description = *response.Value.Description
				}
				t.Fatalf("%s 403 response does not describe password-change gate: %s", route.path, description)
			}
		})
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

const (
	httpMethodGet  = "GET"
	httpMethodPost = "POST"
)

func operationFor(t *testing.T, doc *openapi3.T, path string, method string) *openapi3.Operation {
	t.Helper()

	pathItem := doc.Paths.Find(path)
	if pathItem == nil {
		t.Fatalf("OpenAPI path %s missing", path)
	}

	var operation *openapi3.Operation
	switch method {
	case httpMethodGet:
		operation = pathItem.Get
	case httpMethodPost:
		operation = pathItem.Post
	default:
		t.Fatalf("unsupported test method %s", method)
	}
	if operation == nil {
		t.Fatalf("OpenAPI operation %s %s missing", method, path)
	}
	return operation
}

func reasonResponseEnum(t *testing.T, doc *openapi3.T) map[string]struct{} {
	t.Helper()

	reasonSchemaRef := doc.Components.Schemas["ReasonResponse"]
	if reasonSchemaRef == nil || reasonSchemaRef.Value == nil {
		t.Fatal("components.schemas.ReasonResponse missing from OpenAPI document")
	}

	reasonProperty := reasonSchemaRef.Value.Properties["reason"]
	if reasonProperty == nil || reasonProperty.Value == nil {
		t.Fatal("components.schemas.ReasonResponse.properties.reason missing from OpenAPI document")
	}

	enum := make(map[string]struct{}, len(reasonProperty.Value.Enum))
	for _, value := range reasonProperty.Value.Enum {
		reason, ok := value.(string)
		if !ok {
			t.Fatalf("ReasonResponse reason enum contains non-string value: %#v", value)
		}
		enum[reason] = struct{}{}
	}
	return enum
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

func loadServerPaths(t *testing.T, serverPath string) map[string]struct{} {
	t.Helper()

	operations := loadServerOperationKeys(t, serverPath)
	paths := make(map[string]struct{})
	for key := range operations {
		_, path, found := strings.Cut(key, " ")
		if found {
			paths[path] = struct{}{}
		}
	}
	return paths
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

		path, ok := routePathFromMethodsReceiver(methodsSelector.X)
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

func routePathFromMethodsReceiver(expr ast.Expr) (string, bool) {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return "", false
	}

	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel == nil {
		return "", false
	}

	switch selector.Sel.Name {
	case "Handle", "HandleFunc", "PathPrefix":
		if len(call.Args) == 0 {
			return "", false
		}
		return stringLiteral(call.Args[0])
	case "Handler":
		return routePathFromMethodsReceiver(selector.X)
	default:
		return "", false
	}
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

func extensionMap(t *testing.T, extensions map[string]any, key string) map[string]any {
	t.Helper()

	raw, ok := extensions[key]
	if !ok {
		t.Fatalf("OpenAPI extension %s missing", key)
	}
	value, ok := raw.(map[string]any)
	if !ok {
		t.Fatalf("OpenAPI extension %s has type %T, want map[string]any", key, raw)
	}
	return value
}

func anySlice(t *testing.T, raw any, field string) []any {
	t.Helper()

	value, ok := raw.([]any)
	if !ok {
		t.Fatalf("%s has type %T, want []any", field, raw)
	}
	return value
}

func stringSlice(t *testing.T, raw any, field string) []string {
	t.Helper()

	values := anySlice(t, raw, field)
	result := make([]string, 0, len(values))
	for _, value := range values {
		stringValue, ok := value.(string)
		if !ok {
			t.Fatalf("%s contains %T, want string", field, value)
		}
		result = append(result, stringValue)
	}
	return result
}

func sortedKeys[T any](values map[string]T) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
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
