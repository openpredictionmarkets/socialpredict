---
title: API Conformance Placeholder
document_type: production-notes
domain: api
author: Patrick Delaney
updated_at: 2026-05-24T00:00:00Z
updated_at_display: "Sunday, May 24, 2026"
update_reason: "Placeholder for a future API conformance harness combining Go kin-openapi checks and Schemathesis runtime probing."
status: placeholder
---

# API Conformance Placeholder

## Purpose

This is a placeholder for a future implementation PR that adds an API conformance harness to SocialPredict.

The goal is to combine two complementary checks:

1. Static OpenAPI validation with Go and `kin-openapi`.
2. Runtime OpenAPI probing with Schemathesis against a running backend.

No runtime tooling is implemented in this placeholder.

## Proposed Shape

Suggested future structure:

```text
backend/
  openapi_test.go
  scripts/
    api-conformance.sh
  docs/
    README.md
    openapi.yaml
```

The future script should run the existing in-repo Go OpenAPI checks first, then optionally run Schemathesis when a backend base URL and Schemathesis CLI are available.

## Static Contract Layer

The existing static layer already lives in the backend:

- `backend/docs/openapi.yaml` is the canonical OpenAPI document.
- `backend/openapi_test.go` uses `github.com/getkin/kin-openapi/openapi3`.
- `backend/server/server_contract_test.go` checks backend-served docs publishing.

Baseline command:

```bash
cd backend && go test ./...
```

Focused checks:

```bash
cd backend && go test . -run 'TestOpenAPI|TestRouteFamily|TestReasonResponse|TestEmbedded|TestDocsPublishing'
cd backend && go test ./server
```

## Runtime Contract Layer

The runtime layer should use Schemathesis only after static OpenAPI validation passes.

Initial posture:

- opt-in locally
- bounded request volume
- endpoint-family scoped where practical
- not a hard CI gate until stable
- respectful of authentication, seed-data, and rate-limit constraints

Sketch command:

```bash
SCHEMATHESIS_BASE_URL=http://localhost:8080 \
  schemathesis run \
    --url http://localhost:8080 \
    --workers 1 \
    --rate-limit 30/m \
    --max-examples 10 \
    --generation-database none \
    --request-timeout 10 \
    backend/docs/openapi.yaml
```

## Future Acceptance Criteria

- [ ] Add an in-repo script that runs Go OpenAPI checks.
- [ ] Add optional Schemathesis runtime probing behind an explicit base URL.
- [ ] Document required local setup and failure triage.
- [ ] Keep auth-heavy or stateful endpoints scoped until test data and credentials are explicit.
- [ ] Do not make Schemathesis a required CI gate until false positives and rate-limit effects are understood.
