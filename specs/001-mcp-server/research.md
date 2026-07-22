# Phase 0 Research: MCP Server

**Feature**: [spec.md](spec.md) | **Date**: 2026-07-22

All Technical Context unknowns resolved below. No NEEDS CLARIFICATION markers remain.

## R1. MCP SDK choice

- **Decision**: `github.com/modelcontextprotocol/go-sdk` — the official MCP Go SDK.
- **Rationale**: Official SDK maintained under the modelcontextprotocol org
  (with Google/Anthropic involvement), tracks the current MCP specification,
  ships a Streamable HTTP server transport that returns a standard
  `http.Handler` — exactly what mounting inside the existing gorilla/mux
  router requires. Permissive license, MIT-compatible (constitution V).
- **Alternatives considered**:
  - `mark3labs/mcp-go` — popular community SDK, but third-party; official SDK
    supersedes it for long-term spec tracking.
  - Hand-rolled JSON-RPC 2.0 implementation — full control, zero deps, but
    re-implements protocol negotiation, capability discovery, and streaming;
    high maintenance cost against an evolving spec for no benefit.

## R2. Transport

- **Decision**: Streamable HTTP transport, stateless mode, mounted at
  `/v0/mcp` on the existing backend HTTP server.
- **Rationale**: Streamable HTTP is the current MCP remote transport (single
  endpoint, POST for client→server messages). Stateless mode means no
  server-side session affinity — safe behind the nginx reverse proxy in
  production and identical in dev/localhost. Read-only tools need no
  server-initiated streams.
- **Alternatives considered**:
  - stdio transport — local-only; fails the VPS/production requirement outright.
  - HTTP+SSE (legacy 2024-11-05 transport) — deprecated in favor of Streamable
    HTTP; long-lived SSE sessions are fragile behind proxies.
  - WebSocket — not a standard MCP transport.

## R3. Process/deployment topology

- **Decision**: Embed the MCP handler in the existing backend binary and
  route table; no new container, port, or compose service.
- **Rationale**: The nginx vhosts for dev and prod already proxy
  `location /api/ → backend:8080/`, so `/v0/mcp` is automatically reachable at
  `{origin}/api/v0/mcp` in every environment — env parity (FR-005, SC-002)
  falls out of the existing deploy path with zero operator steps. Dev also
  exposes the backend port directly (`${BACKEND_PORT}:8080`) for
  `http://localhost:{port}/v0/mcp`. A separate service would need new compose
  entries, vhost rules, TLS wiring, and health checks in all three
  environments — pure cost against constitution V.
- **Alternatives considered**:
  - Sidecar MCP container proxying the REST API — extra hop, extra deploy
    surface, duplicate DTOs.
  - Standalone binary under `backend/cmd/` — same code, worse operations story.

## R4. Tool catalog (v1, read-only)

- **Decision**: Four tools mapping 1:1 onto existing public reads:
  `list_markets`, `get_market`, `search_markets`, `list_market_bets`.
- **Rationale**: Covers spec Story 1 (list, details, discoverability) plus bet
  history that the market page shows publicly. Everything maps to existing
  `dmarkets` service methods already exposed unauthenticated on the website
  (`/v0/markets`, `/v0/markets/{id}`, `/v0/markets/search`,
  `/v0/markets/bets/{marketId}`), so "public data only" (FR-003) is inherited,
  not re-derived. Small catalog keeps the strictly-read-only guarantee easy to
  test structurally.
- **Alternatives considered**:
  - MCP resources/resource templates in addition to tools — weaker client
    support than tools; adds surface without adding data. Deferred.
  - Exposing stats/leaderboards — public but secondary; easy follow-up once
    the pattern exists. Deferred to keep v1 minimal.
  - Write tools (place bet, etc.) — explicitly out of scope per user decision
    recorded in spec Assumptions.

## R5. Authentication & abuse protection

- **Decision**: No authentication in v1 (anonymous, public data only).
  Wrap the MCP route in the existing `securityMiddleware` (security headers +
  rate limiting via the existing `security` package).
- **Rationale**: v1 exposes only what a logged-out visitor sees; adding auth
  would contradict the agreed scope. Reusing `securityMiddleware` satisfies
  FR-006 (same abuse protections as equivalent public web traffic) with no new
  code paths. Future account linkage can reuse the platform's API-key work in
  the auth domain.
- **Alternatives considered**:
  - New MCP-specific rate limiter — duplicate of existing `RateManager`; rejected.
  - Requiring an API key even for public reads — friction with zero data-protection
    benefit in a read-only surface.

## R6. OpenAPI / API surface documentation

- **Decision**: Add a minimal `/v0/mcp` path entry to the embedded
  `openapi.yaml` describing it as the MCP (JSON-RPC 2.0) endpoint, with a
  pointer to the MCP specification and the tool contract document.
- **Rationale**: Constitution V requires public API surface changes to be
  reflected in the embedded OpenAPI spec. MCP is JSON-RPC, not REST, so the
  entry documents the endpoint's existence, method (POST), and content type
  rather than modeling every RPC message; the detailed tool contract lives in
  [contracts/mcp-tools.md](contracts/mcp-tools.md).
- **Alternatives considered**: Fully modeling JSON-RPC envelopes in OpenAPI —
  poor fit, high noise, duplicates the MCP schema definitions.

## R7. Testing approach

- **Decision**: Three layers, all standard `go test`:
  1. Unit tests per tool against a hand-written fake `MarketReader`.
  2. Contract test spinning up the real handler via `httptest` and driving the
     MCP lifecycle (initialize → tools/list → tools/call) as JSON-RPC.
  3. Catalog guard test asserting the advertised tool list contains exactly
     the four read-only tools (enforces FR-003/SC-003 structurally).
- **Rationale**: Matches constitution III (fakes, no mock frameworks) and the
  existing `server_contract_test.go` pattern in `backend/server/`.
- **Alternatives considered**: End-to-end Docker test across all three compose
  files — valuable but belongs in the existing integration-test tier;
  quickstart.md documents the manual verification per environment.
