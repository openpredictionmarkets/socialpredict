# Implementation Plan: MCP Server

**Branch**: `feature/mcp-server` | **Date**: 2026-07-22 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/001-mcp-server/spec.md`

## Summary

Expose SocialPredict's public market data to AI assistants through an MCP
(Model Context Protocol) endpoint. v1 is strictly read-only: a small catalog of
tools (list markets, market details, search, market bets) served over MCP's
Streamable HTTP transport, mounted inside the existing backend binary at
`/v0/mcp`. Because it rides the existing HTTP stack and nginx proxy, the same
endpoint works unchanged in dev, localhost, and production VPS deployments —
reachable at `{origin}/api/v0/mcp` everywhere.

## Technical Context

**Language/Version**: Go 1.25.0 (existing `backend/` module `socialpredict`)

**Primary Dependencies**: gorilla/mux (routing), GORM/PostgreSQL (existing
persistence, untouched), rs/cors, `github.com/modelcontextprotocol/go-sdk`
(NEW — official MCP Go SDK, Streamable HTTP transport)

**Storage**: PostgreSQL via existing domain services; this feature adds NO
storage, migrations, or models

**Testing**: `go test` (`/usr/local/go/bin/go`, clean `GOFLAGS`), `httptest`
against the mounted MCP handler, hand-written fakes for the markets service
interface (no mocking frameworks, per constitution)

**Target Platform**: Linux server in Docker (dev / localhost / production VPS
behind nginx + HTTPS); dev also exposes the backend port directly

**Project Type**: Web service — backend-only change; no frontend work

**Performance Goals**: Same as existing public read endpoints; MCP adds a thin
JSON-RPC layer over the same domain service calls. Stateless request handling —
no per-session server state that would break behind a proxy

**Constraints**: Strictly read-only tool surface (FR-003); must function behind
nginx reverse proxy with HTTPS in production without sticky sessions; no new
container or deployment step; rate limiting via existing `securityMiddleware`

**Scale/Scope**: Self-hosted instances (hundreds of users); a handful of
concurrent MCP clients; 4 tools in v1

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| # | Principle | Assessment | Status |
|---|-----------|------------|--------|
| I | Domain-Driven Backend Architecture | MCP adapter is a transport layer in `backend/handlers/mcp/`; it consumes the existing `dmarkets` domain service via the DI container. No business logic in the adapter, no persistence types leaked — tools map domain structs to their own output DTOs. Wired in `BuildApplication()`/`server.go` like every other handler. | PASS |
| II | Narrow, Consumer-Defined Interfaces | The MCP package defines its own consumer-side interface (`MarketReader`, ~4 methods: ListMarkets, GetMarketDetails, SearchMarkets, market bets read) instead of depending on the full markets service interface. Handler follows the `http.Handler` mounting pattern. | PASS |
| III | Tests Accompany Every Change | Unit tests per tool with a hand-written fake `MarketReader`; contract test driving the real JSON-RPC endpoint via `httptest` (initialize → tools/list → tools/call); catalog test asserting no write tools are exposed. No mocking frameworks. | PASS |
| IV | Financial Integrity (NON-NEGOTIABLE) | Feature is read-only by requirement (FR-003). No bet/balance/market mutation paths exist; the catalog test enforces this structurally. No monetary math added. | PASS |
| V | Secure, Self-Hostable by Default | No new deployment step: endpoint rides existing backend + nginx `/api/` proxy in all three environments. Wrapped in existing `securityMiddleware` (headers + rate limiting). No secrets, no auth surface added. `/v0/mcp` documented in the embedded OpenAPI spec. New dependency (official MCP Go SDK) is permissively licensed, MIT-compatible. | PASS |

**Initial gate**: PASS — no violations, Complexity Tracking empty.

**Post-design re-check (after Phase 1)**: PASS — design artifacts introduce no
new entities, no writes, no deployment divergence; tool DTOs defined in the
adapter keep domain types unleaked.

## Project Structure

### Documentation (this feature)

```text
specs/001-mcp-server/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md        # Phase 1 output (/speckit-plan command)
├── quickstart.md        # Phase 1 output (/speckit-plan command)
├── contracts/           # Phase 1 output (/speckit-plan command)
│   └── mcp-tools.md
└── tasks.md             # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
backend/
├── handlers/
│   └── mcp/                     # NEW: MCP transport adapter
│       ├── server.go            # MCP server construction, tool registration,
│       │                        #   http.Handler via Streamable HTTP transport
│       ├── interfaces.go        # Narrow MarketReader interface (consumer-defined)
│       ├── tools_markets.go     # list_markets / get_market / search_markets /
│       │                        #   list_market_bets tool implementations + DTOs
│       ├── tools_markets_test.go# Unit tests with hand-written fake MarketReader
│       └── server_contract_test.go # JSON-RPC contract test via httptest
├── server/
│   └── server.go                # MODIFIED: mount /v0/mcp route (securityMiddleware)
├── internal/app/container.go    # MODIFIED (if needed): expose markets service to wiring
├── openapi.yaml (embedded)      # MODIFIED: document /v0/mcp endpoint
└── go.mod / go.sum              # MODIFIED: add MCP Go SDK

scripts/                         # UNCHANGED: no compose changes needed
data/nginx/vhosts/               # UNCHANGED: /api/ proxy already covers /v0/mcp
```

**Structure Decision**: Backend-only. New `backend/handlers/mcp/` package as a
transport adapter beside existing handler packages; single route added in
`backend/server/server.go`. Deployment artifacts untouched — environment parity
(dev/localhost/prod) comes from reusing the existing entry path.

## Complexity Tracking

No constitution violations — table intentionally empty.
