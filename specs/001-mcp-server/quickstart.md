# Quickstart: MCP Server Validation

**Feature**: [spec.md](spec.md) | Contract: [contracts/mcp-tools.md](contracts/mcp-tools.md)

Validation scenarios proving the feature end-to-end in all three environments
(spec Story 2 / SC-002). Same client config everywhere — only the address changes.

## Prerequisites

- Docker + Docker Compose V2
- Repo checked out; `./SocialPredict install` completed for the target env
- An MCP-compatible client (Claude Code, MCP Inspector, or plain `curl`)

## 1. Automated tests (fastest signal)

```bash
cd backend
GOFLAGS="" /usr/local/go/bin/go test ./handlers/mcp/... ./server/...
```

**Expected**: unit tests (tools against fake reader), contract test
(initialize → tools/list → tools/call), and catalog guard test all pass.
Note: pre-existing failure in `internal/domain/balance_integration_test.go`
is unrelated.

## 2. Dev environment

```bash
./SocialPredict up   # dev compose
```

Direct to backend:

```bash
curl -s -X POST "http://localhost:${BACKEND_PORT:-8080}/v0/mcp" \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json, text/event-stream' \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list"}'
```

Through nginx (same path shape as prod):

```bash
curl -s -X POST "http://localhost/api/v0/mcp" ...same body...
```

**Expected**: JSON-RPC result listing exactly `list_markets`, `get_market`,
`search_markets`, `list_market_bets`.

## 3. Localhost / prod-like deployment

Bring up the localhost compose (`scripts/docker-compose-local.yaml` via the
installer's local mode), then repeat the nginx-path call:

```bash
curl -s -X POST "http://localhost/api/v0/mcp" ...same body...
```

**Expected**: identical response to dev.

## 4. Production VPS

After the documented HTTPS deployment:

```bash
curl -s -X POST "https://<your-domain>/api/v0/mcp" ...same body...
```

**Expected**: identical response over TLS. No extra operator steps beyond the
standard deployment (Story 2, scenario 3).

## 5. Real MCP client (SC-001: under 5 minutes)

Claude Code:

```bash
claude mcp add --transport http socialpredict https://<your-domain>/api/v0/mcp
```

Then ask: *"What markets are open on socialpredict?"*

**Expected**: assistant calls `list_markets` and answers with live market data
matching the website (SC-004).

MCP Inspector alternative:

```bash
npx @modelcontextprotocol/inspector
# connect: Streamable HTTP → https://<your-domain>/api/v0/mcp
```

## 6. Negative checks (FR-003/FR-004, SC-003)

```bash
# Unknown tool → error, no leakage
curl -s -X POST "$ORIGIN/api/v0/mcp" -H 'Content-Type: application/json' \
  -H 'Accept: application/json, text/event-stream' \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/call",
       "params":{"name":"place_bet","arguments":{}}}'

# Unknown market id → "market not found"
curl -s -X POST "$ORIGIN/api/v0/mcp" -H 'Content-Type: application/json' \
  -H 'Accept: application/json, text/event-stream' \
  -d '{"jsonrpc":"2.0","id":3,"method":"tools/call",
       "params":{"name":"get_market","arguments":{"id":999999}}}'
```

**Expected**: clean tool/JSON-RPC errors; zero state changes anywhere
(verify: no new bets/balance changes in the UI).

## 7. Data parity spot-check (SC-004/SC-005)

Pick one market. Compare `get_market` output (probability, volume, status)
against the market's web page at the same moment. Values must match.
