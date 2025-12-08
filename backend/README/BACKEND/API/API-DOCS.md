# SocialPredict API Documentation

## Overview

This directory contains the API documentation for the SocialPredict prediction markets platform.

## Files

- `openapi.yaml` - OpenAPI 3.0.3 specification for the SocialPredict API
- `API-DOCS.md` - This file, providing an overview and instructions
- `API-DESIGN-REPORT.md` - Current API state and roadmap

## Using the API Documentation

### Built-in Swagger UI and Spec

When running the backend locally (on port 8080 by default), the server exposes:

- `GET /swagger/` – Embedded Swagger UI, preconfigured to load the current OpenAPI spec.
- `GET /openapi.yaml` – The bundled OpenAPI 3.0.3 specification served directly from the binary.
- `GET /health` – Plain-text health check that returns `ok` when the backend is up.

For example:

```bash
curl http://localhost:8080/health
curl http://localhost:8080/openapi.yaml
```

Open `http://localhost:8080/swagger` in a browser to interact with the backend routes.

### Building Docs using Redoc

```bash
# Install redoc-cli globally
npm install -g redoc-cli

# Generate static HTML documentation
redoc-cli build openapi.yaml --output api-docs.html

# Serve the documentation
redoc-cli serve openapi.yaml --port 8082
```

Then visit <http://localhost:8082>

### API Base URLs

- **Production**: `https://api.socialpredict.com/v0`
- **Staging**: `https://staging-api.socialpredict.com/v0`
- **Development**: `http://localhost:8080/v0`

## Authentication

Most API endpoints require authentication using Bearer tokens:

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -X GET \
     "https://api.socialpredict.com/v0/markets"
```

## Current API Coverage

### ✅ Implemented Endpoints

- `GET /markets` - List markets with filtering
- `POST /markets` - Create new markets (authenticated)
- `GET /markets/{id}` - Get market details
- `GET /markets/search` - Search markets

### 🚧 In Progress

The following endpoints are being migrated to the new clean architecture:

- User management endpoints
- Betting/position endpoints
- Administrative endpoints
- Metrics and statistics endpoints

### 📋 Planned Endpoints

See `API-DESIGN-REPORT.md` for the complete roadmap.

## Making API Requests

### Example: List Markets

```bash
curl -X GET "http://localhost:8080/v0/markets?status=active&limit=10"
```

### Example: Create Market

```bash
curl -X POST "http://localhost:8080/v0/markets" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "questionTitle": "Will it rain tomorrow?",
    "description": "Market resolves based on local weather station data",
    "outcomeType": "binary",
    "resolutionDateTime": "2024-12-01T12:00:00Z",
    "yesLabel": "Rain",
    "noLabel": "No Rain"
  }'
```

## Error Handling

All API endpoints return consistent error responses:

```json
{
  "error": "Human readable error message",
  "code": "ERROR_CODE",
  "details": "Additional context if available"
}
```

Common HTTP status codes:

- `200` - Success
- `201` - Created
- `400` - Bad Request
- `401` - Unauthorized
- `404` - Not Found
- `500` - Internal Server Error

## Development

### Updating the API Documentation

1. Modify `openapi.yaml` as needed
2. Validate the OpenAPI spec:

   ```bash
   npx @apidevtools/swagger-parser validate openapi.yaml
   ```

3. Update this documentation if needed
4. Test the changes with Swagger UI

### Code Generation

You can generate client SDKs and server stubs from the OpenAPI specification:

```bash
# Generate Go client
openapi-generator generate -i openapi.yaml -g go -o ./go-client

# Generate TypeScript client
openapi-generator generate -i openapi.yaml -g typescript-axios -o ./ts-client

# Generate Python client
openapi-generator generate -i openapi.yaml -g python -o ./python-client
```

## Support

For API support or questions:

- Create an issue in the project repository
- Contact: <support@socialpredict.com>
