# API Design Implementation Plan

## Overview
Standardize the API design with consistent patterns, proper versioning, comprehensive documentation, and improved response formats while maintaining backward compatibility.

## Current State Analysis
- API endpoints use `/v0/` versioning prefix
- Inconsistent response formats across endpoints
- Basic REST patterns but not fully RESTful
- Limited API documentation
- No OpenAPI/Swagger specification
- Mixed error response formats

## Implementation Steps

### Step 1: API Design Standards
**Timeline: 2-3 days**

Establish comprehensive API design standards:

```go
// api/standards.go
type APIResponse struct {
    Data       interface{}       `json:"data,omitempty"`
    Meta       *ResponseMeta     `json:"meta,omitempty"`
    Links      *ResponseLinks    `json:"links,omitempty"`
    Error      *ErrorResponse    `json:"error,omitempty"`
    Success    bool              `json:"success"`
    Timestamp  time.Time         `json:"timestamp"`
    Version    string            `json:"version"`
}

type ResponseMeta struct {
    Page         int `json:"page,omitempty"`
    PerPage      int `json:"per_page,omitempty"`
    Total        int `json:"total,omitempty"`
    TotalPages   int `json:"total_pages,omitempty"`
}

type ResponseLinks struct {
    Self     string `json:"self,omitempty"`
    Next     string `json:"next,omitempty"`
    Previous string `json:"previous,omitempty"`
    First    string `json:"first,omitempty"`
    Last     string `json:"last,omitempty"`
}
```

**Design principles:**
- Consistent response format
- RESTful resource naming
- Proper HTTP status codes
- Pagination standards
- Resource relationships
- HATEOAS implementation

### Step 2: API Versioning Strategy
**Timeline: 1-2 days**

Implement comprehensive API versioning:

```go
// api/versioning.go
type APIVersion struct {
    Major         int       `json:"major"`
    Minor         int       `json:"minor"`
    Patch         int       `json:"patch"`
    Deprecated    bool      `json:"deprecated"`
    SunsetDate    *time.Time `json:"sunset_date,omitempty"`
    SupportedUntil *time.Time `json:"supported_until,omitempty"`
}

type VersionManager struct {
    versions map[string]APIVersion
    handlers map[string]map[string]http.Handler // version -> endpoint -> handler
}

func (vm *VersionManager) RegisterHandler(version, endpoint string, handler http.Handler) {
    if vm.handlers[version] == nil {
        vm.handlers[version] = make(map[string]http.Handler)
    }
    vm.handlers[version][endpoint] = handler
}

func (vm *VersionManager) GetHandler(r *http.Request) (http.Handler, error) {
    version := vm.extractVersion(r)
    endpoint := vm.extractEndpoint(r)

    if handler, exists := vm.handlers[version][endpoint]; exists {
        return handler, nil
    }

    return nil, fmt.Errorf("handler not found for version %s, endpoint %s", version, endpoint)
}
```

**Versioning features:**
- Header-based versioning (Accept header)
- URL path versioning (/v1/, /v2/)
- Version deprecation warnings
- Sunset date management
- Backward compatibility support

### Step 3: OpenAPI Specification
**Timeline: 3-4 days**

Generate comprehensive OpenAPI documentation:

```go
// api/swagger.go
type SwaggerGenerator struct {
    config SwaggerConfig
    spec   *openapi3.T
}

func (sg *SwaggerGenerator) GenerateSpec() *openapi3.T {
    spec := &openapi3.T{
        OpenAPI: "3.0.3",
        Info: &openapi3.Info{
            Title:       "SocialPredict API",
            Description: "Prediction market API for social predictions",
            Version:     "1.0.0",
        },
        Servers: openapi3.Servers{
            {URL: "https://api.socialpredict.com/v1"},
            {URL: "https://staging-api.socialpredict.com/v1"},
        },
    }

    // Add paths, components, security schemes
    sg.addPaths(spec)
    sg.addComponents(spec)
    sg.addSecurity(spec)

    return spec
}

// Generate OpenAPI annotations
//go:generate swaggo init -g main.go -o ./docs
```

**Documentation features:**
- Complete endpoint documentation
- Request/response schemas
- Authentication requirements
- Error response documentation
- Interactive API explorer
- Code generation support

### Step 4: Request/Response Middleware
**Timeline: 2 days**

Standardize request/response handling:

```go
// api/middleware/response.go
func ResponseFormatterMiddleware() mux.MiddlewareFunc {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            rw := &responseWrapper{
                ResponseWriter: w,
                request:       r,
                statusCode:    http.StatusOK,
            }

            next.ServeHTTP(rw, r)

            // Format response consistently
            rw.formatResponse()
        })
    }
}

type responseWrapper struct {
    http.ResponseWriter
    request    *http.Request
    statusCode int
    body       []byte
}

func (rw *responseWrapper) formatResponse() {
    response := APIResponse{
        Success:   rw.statusCode < 400,
        Timestamp: time.Now(),
        Version:   getAPIVersion(rw.request),
    }

    if rw.statusCode >= 400 {
        // Handle error response
        response.Error = parseErrorResponse(rw.body)
    } else {
        // Handle success response
        response.Data = parseSuccessResponse(rw.body)
        response.Meta = generateMeta(rw.request)
        response.Links = generateLinks(rw.request)
    }

    rw.Header().Set("Content-Type", "application/json")
    json.NewEncoder(rw).Encode(response)
}
```

### Step 5: Pagination and Filtering
**Timeline: 2-3 days**

Implement standardized pagination and filtering:

```go
// api/pagination.go
type PaginationRequest struct {
    Page     int `json:"page" validate:"min=1"`
    PerPage  int `json:"per_page" validate:"min=1,max=100"`
    Sort     string `json:"sort"`
    Order    string `json:"order" validate:"oneof=asc desc"`
}

type FilterRequest struct {
    Filters map[string]interface{} `json:"filters"`
    Search  string                 `json:"search"`
    DateRange *DateRangeFilter     `json:"date_range"`
}

func (pr *PaginationRequest) GetOffset() int {
    return (pr.Page - 1) * pr.PerPage
}

func (pr *PaginationRequest) GetLimit() int {
    return pr.PerPage
}

// Usage in handlers
func ListMarketsHandler(w http.ResponseWriter, r *http.Request) {
    pagination, err := ParsePaginationRequest(r)
    if err != nil {
        WriteErrorResponse(w, err)
        return
    }

    filters, err := ParseFilterRequest(r)
    if err != nil {
        WriteErrorResponse(w, err)
        return
    }

    markets, total, err := marketService.List(r.Context(), pagination, filters)
    if err != nil {
        WriteErrorResponse(w, err)
        return
    }

    response := PaginatedResponse{
        Data: markets,
        Meta: ResponseMeta{
            Page:       pagination.Page,
            PerPage:    pagination.PerPage,
            Total:      total,
            TotalPages: (total + pagination.PerPage - 1) / pagination.PerPage,
        },
        Links: GeneratePaginationLinks(r, pagination, total),
    }

    WriteSuccessResponse(w, response)
}
```

### Step 6: Content Negotiation
**Timeline: 1-2 days**

Implement content negotiation for different response formats:

```go
// api/content_negotiation.go
type ContentNegotiator struct {
    supportedFormats map[string]ResponseFormatter
}

type ResponseFormatter interface {
    Format(data interface{}) ([]byte, error)
    ContentType() string
}

type JSONFormatter struct{}
func (jf *JSONFormatter) Format(data interface{}) ([]byte, error) {
    return json.Marshal(data)
}
func (jf *JSONFormatter) ContentType() string {
    return "application/json"
}

type XMLFormatter struct{}
func (xf *XMLFormatter) Format(data interface{}) ([]byte, error) {
    return xml.Marshal(data)
}
func (xf *XMLFormatter) ContentType() string {
    return "application/xml"
}

func (cn *ContentNegotiator) Negotiate(r *http.Request) ResponseFormatter {
    acceptHeader := r.Header.Get("Accept")

    switch {
    case strings.Contains(acceptHeader, "application/json"):
        return cn.supportedFormats["json"]
    case strings.Contains(acceptHeader, "application/xml"):
        return cn.supportedFormats["xml"]
    default:
        return cn.supportedFormats["json"] // default
    }
}
```

### Step 7: API Rate Documentation and Examples
**Timeline: 2 days**

Create comprehensive API documentation with examples:

```markdown
# SocialPredict API Documentation

## Authentication
All API requests require authentication using Bearer tokens:

```bash
curl -H "Authorization: Bearer your-token-here" \
     https://api.socialpredict.com/v1/markets
```

## Rate Limiting
- 1000 requests per hour per authenticated user
- 100 requests per hour per IP address
- Rate limit headers included in responses

## Pagination
All list endpoints support pagination:

```bash
curl "https://api.socialpredict.com/v1/markets?page=2&per_page=20"
```

## Filtering and Search
Most list endpoints support filtering:

```bash
curl "https://api.socialpredict.com/v1/markets?status=active&search=crypto"
```
```

## Directory Structure
```
api/
├── standards.go           # API design standards
├── versioning.go          # Version management
├── swagger.go             # OpenAPI generation
├── pagination.go          # Pagination utilities
├── filtering.go           # Filtering utilities
├── content_negotiation.go # Content negotiation
├── middleware/
│   ├── response.go        # Response formatting
│   ├── versioning.go      # Version selection
│   └── content_type.go    # Content type handling
├── handlers/
│   ├── base.go           # Base handler with common functionality
│   └── validators.go     # Request validation
├── docs/
│   ├── swagger.json      # Generated OpenAPI spec
│   ├── swagger.yaml      # Generated OpenAPI spec (YAML)
│   └── examples/         # API usage examples
└── client/
    ├── go/               # Generated Go client
    ├── javascript/       # Generated JS client
    └── python/           # Generated Python client
```

## API Standards Documentation
```yaml
# api-standards.yaml
response_format:
  success:
    required_fields: ["data", "success", "timestamp", "version"]
    optional_fields: ["meta", "links"]
  error:
    required_fields: ["error", "success", "timestamp", "version"]

pagination:
  default_per_page: 20
  max_per_page: 100
  required_params: ["page", "per_page"]

versioning:
  strategy: "url_path"  # /v1/, /v2/
  deprecation_notice: "6_months"
  sunset_period: "12_months"

status_codes:
  success: [200, 201, 202, 204]
  client_error: [400, 401, 403, 404, 409, 422, 429]
  server_error: [500, 502, 503, 504]
```

## Testing Strategy
- API contract testing with OpenAPI specs
- Response format validation tests
- Pagination and filtering tests
- Content negotiation tests
- Version compatibility tests
- Documentation accuracy tests

## Migration Strategy
1. Implement new API standards alongside existing endpoints
2. Create v1 endpoints with new standards
3. Deprecate v0 endpoints with sunset timeline
4. Update client libraries and documentation
5. Monitor usage and migrate users gradually

## Benefits
- Consistent developer experience
- Improved API discoverability
- Better tooling support
- Easier client library generation
- Standardized error handling
- Professional API documentation

## Integration Points
- **Authentication**: JWT token validation in all endpoints
- **Logging**: All API calls logged with request/response details
- **Metrics**: API usage metrics by endpoint, version, and user
- **Caching**: Response caching for read-heavy endpoints
- **Security**: Rate limiting and input validation on all endpoints