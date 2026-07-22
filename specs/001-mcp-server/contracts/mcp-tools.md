# MCP Contract: SocialPredict Tools

**Feature**: [../spec.md](../spec.md) | **Date**: 2026-07-22

## Endpoint

| Property | Value |
|----------|-------|
| Path (backend) | `POST /v0/mcp` |
| Path (public, all environments) | `POST {origin}/api/v0/mcp` (nginx `/api/` → backend) |
| Path (dev direct) | `POST http://localhost:{BACKEND_PORT}/v0/mcp` |
| Protocol | MCP over Streamable HTTP (JSON-RPC 2.0), stateless |
| Content-Type | `application/json` (client MUST also send `Accept: application/json, text/event-stream`) |
| Auth | None (anonymous; public data only in v1) |
| Rate limiting | Same policy as public web endpoints (existing security middleware) |

Server info advertised on `initialize`: name `socialpredict`, version = backend
version. Capabilities: `tools` only (no resources, prompts, or sampling in v1).

## Lifecycle (contract test sequence)

1. `initialize` request → result includes `protocolVersion`, `serverInfo`,
   `capabilities.tools`.
2. `notifications/initialized` → accepted (202/no error).
3. `tools/list` → exactly the four tools below, each with a JSON Schema
   `inputSchema` (`additionalProperties: false`).
4. `tools/call` per tool → `content` containing JSON text per the result
   shapes in [../data-model.md](../data-model.md); `isError: true` with a
   human-readable message on invalid input.

## Tools

### 1. `list_markets`

List markets, optionally filtered by status.

**Input schema**:

```json
{
  "type": "object",
  "properties": {
    "status": { "type": "string", "enum": ["active", "closed", "resolved", "all"], "default": "all" },
    "limit":  { "type": "integer", "minimum": 1, "maximum": 100, "default": 50 },
    "offset": { "type": "integer", "minimum": 0, "default": 0 }
  },
  "additionalProperties": false
}
```

**Result**: JSON array of `MarketSummary` (see data-model.md).
Empty instance → `[]`, not an error (spec edge case).

### 2. `get_market`

Full public detail for one market.

**Input schema**:

```json
{
  "type": "object",
  "properties": {
    "id": { "type": "integer", "minimum": 1 }
  },
  "required": ["id"],
  "additionalProperties": false
}
```

**Result**: `MarketDetail` JSON object.
**Errors**: unknown id → tool error `market not found`.

### 3. `search_markets`

Keyword search over markets.

**Input schema**:

```json
{
  "type": "object",
  "properties": {
    "query":  { "type": "string", "minLength": 1 },
    "status": { "type": "string", "enum": ["active", "closed", "resolved", "all"], "default": "all" },
    "limit":  { "type": "integer", "minimum": 1, "maximum": 100, "default": 50 }
  },
  "required": ["query"],
  "additionalProperties": false
}
```

**Result**: `SearchResult` JSON object (primary + fallback results).

### 4. `list_market_bets`

Public bet history for one market (what the market page shows).

**Input schema**:

```json
{
  "type": "object",
  "properties": {
    "id": { "type": "integer", "minimum": 1 }
  },
  "required": ["id"],
  "additionalProperties": false
}
```

**Result**: JSON array of `BetInfo`.
**Errors**: unknown id → tool error `market not found`.

## Negative Contract (guaranteed absences — FR-003/FR-004, SC-003)

- `tools/list` result contains **no** tool other than the four above; the
  contract test asserts the exact set.
- No tool mutates platform state; no capability exposes data invisible to a
  logged-out website visitor.
- Calling an undefined tool name → JSON-RPC/tool error, no information leakage.
- `resources/list`, `prompts/list` → empty or method-not-supported per
  advertised capabilities; never data.

## Example (curl, stateless round-trip)

```bash
curl -s -X POST "$ORIGIN/api/v0/mcp" \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json, text/event-stream' \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/call",
       "params":{"name":"list_markets","arguments":{"status":"active"}}}'
```

(Exact header/session requirements follow the MCP Streamable HTTP spec as
implemented by the official Go SDK; the contract test is the executable
source of truth.)
