# Public Market MCP Design

Date: 2026-07-18
Repository: SocialPredict
Canonical SHA: a1f6af4004abceb4643e6752788cc7382e38f1d6

## Goal

Build a first-version MCP server for SocialPredict that lets agents discover, inspect, and compare public market data without requiring API keys or user authentication.

The first release focuses on public, read-only market discovery:

- Search and list markets using the same status and market tag filters as the public REST API.
- List and inspect market tags.
- Inspect individual markets, summaries, volume, probability history, activity pages, positions, and leaderboards.
- Quote hypothetical probability movement without placing bets.

The server must leave clear room for future private tools, API keys, JWT-backed user auth, and service identity, but none of those auth features are implemented in v1.

## Non-Goals

V1 does not include:

- Placing bets, selling positions, resolving markets, or any other write action.
- User-generated API key CRUD.
- OAuth, client credentials, delegated user auth, or service identity.
- Private profile, balance, personal portfolio, or authenticated user position tools.
- Admin, moderator, lifecycle review, content management, existing operations/status APIs, or Swagger endpoints as MCP tools.
- Agent recommendations such as "where should I bet?" The MCP returns facts; consuming agents reason over them.

## Architecture

Add a separate Go binary:

```text
backend/cmd/mcpserver/main.go
```

Add an MCP runtime package:

```text
backend/internal/mcpserver
```

The MCP binary runs as a separate process/container from the main HTTP API but uses the same Go module, configuration style, database connection, repositories, and domain services. It should not duplicate market math, tag normalization, search behavior, or read-model logic.

The request flow is:

```text
MCP tool request
  -> tool argument validation and normalization
  -> existing SocialPredict service/read-model method
  -> MCP response DTO
```

The MCP server should prefer direct service calls over HTTP calls back into the normal API. Tool schemas should still mirror the public REST/OpenAPI surface so behavior stays recognizable and auditable.

## Runtime Endpoints

The MCP container exposes:

- `/health` for process liveness.
- `/readyz` for dependency readiness.
- `/mcp` for MCP transport.

Exact MCP transport details can follow the Go MCP library chosen during implementation. If the library supports streamable HTTP, `/mcp` should be the stable route exposed through the deployment proxy.

## Authorization Boundary

V1 public tools run without API keys.

The MCP code should still classify tools by access level so auth can be added later without reshaping every tool:

```text
AccessPublicRead
AccessUserRead
AccessUserWrite
AccessServiceRead
```

All v1 tools are `AccessPublicRead`. The auth resolver can return an anonymous principal for v1. Future API-key, JWT, or service identity work can replace that resolver and register private/write tools under stricter access levels.

## Status And Tag Semantics

The canonical backend market statuses are:

- `active`
- `closed`
- `resolved`
- `all`

MCP should accept `open` as an agent-friendly alias for `active`, but responses should preserve canonical backend status names.

Market taxonomy uses market tags, not categories. Tools should use `tagSlug` to filter markets and should normalize tag slugs the same way the backend handlers do: lower-case, trim whitespace, and trim leading/trailing hyphens.

## Public Discovery Tools

### `list_market_tags`

Returns active public market tags.

Arguments: none.

Returns tags with:

- `id`
- `slug`
- `displayName`
- `description`
- `colorKey`
- `sortOrder`
- `isActive`

### `get_market_tag`

Returns one active market tag by slug.

Arguments:

- `slug`

Implementation can call the same tag service as `list_market_tags` and filter by normalized slug. No new REST endpoint is required.

### `list_markets`

Lists public market overviews.

Arguments:

- `status` optional: `active`, `closed`, `resolved`, `all`, or `open`
- `tagSlug` optional
- `createdBy` optional, mapped to existing `created_by`
- `limit` optional
- `offset` optional

Returned market overviews must include existing public fields, including:

- market identity and labels
- status and lifecycle status
- market tags
- market group link/aggregate metadata when present
- creator summary
- last probability
- user count
- `totalVolume`
- `marketDust`

### `search_markets`

Searches public market overviews by question/title using the existing search behavior.

Arguments:

- `query` required
- `status` optional: `active`, `closed`, `resolved`, `all`, or `open`
- `tagSlug` optional
- `limit` optional
- `offset` optional

The MCP tool should preserve the backend's primary/fallback search semantics. Returned results should include counts and fallback metadata where available.

### `get_market_discovery`

Returns the public market discovery read model for the top page or a tag/topic slug.

Arguments:

- `slug` required; `markets` means the top discovery page
- `status` optional: `active`, `closed`, `resolved`, `all`, or `open`
- `tagSlug` optional
- `limit` optional
- `offset` optional

Returns:

- layout/page metadata
- topic navigation when available
- market overview rows
- pinned market details
- total when available
- freshness metadata

This is useful for agents that want the same curated discovery shape the product uses.

### `get_market`

Returns public market detail by market ID.

Arguments:

- `marketId`

Returns market detail fields, creator summary, probability changes, user count, `totalVolume`, `marketDust`, and approved description amendments where available.

### `get_market_summary`

Returns the public market summary read model by market ID.

Arguments:

- `marketId`

Returns a compact summary with probability history, user count, `totalVolume`, `marketDust`, and freshness metadata.

## Public Activity And Analytics Tools

Every list-style activity tool must be paginated. MCP must not expose the old unbounded HTTP behavior for market bets or positions.

### `list_market_bets`

Arguments:

- `marketId`
- `limit` optional
- `offset` optional

Returns recent market bet rows, newest first, including username, outcome, amount, probability, and placed time.

### `list_market_positions`

Arguments:

- `marketId`
- `limit` optional
- `offset` optional

Returns active market positions sorted by total shares, including username, YES/NO shares, value, total spent, in-play spent, resolution state, and result.

### `get_market_user_position`

Arguments:

- `marketId`
- `username`

Returns one public user position in a market. This is not the authenticated user's private position endpoint.

### `get_market_leaderboard`

Arguments:

- `marketId`
- `limit` optional
- `offset` optional

Returns leaderboard rows with username, profit, current value, total spent, position, YES/NO shares, and rank.

### `list_market_group_bets`

Arguments:

- `groupId`
- `limit` optional
- `offset` optional

Returns one globally sorted bet feed across child markets in the group.

### `list_market_group_positions`

Arguments:

- `groupId`
- `limit` optional
- `offset` optional

Returns positions aggregated across child markets with answer-level breakdowns.

### `get_market_group_leaderboard`

Arguments:

- `groupId`
- `limit` optional
- `offset` optional

Returns aggregate grouped leaderboard rows with answer-level profit breakdowns.

### `quote_market_probability`

Quotes a hypothetical probability movement. This is read-only and must not persist a bet.

Arguments:

- `marketId`
- `amount`
- `outcome`: `YES` or `NO`, case-insensitive

Returns:

- `marketId`
- `currentProbability`
- `projectedProbability`
- `amount`
- `outcome`

This tool is decision support, not basic discovery. It belongs in v1 because it is public/read-only and already exists as projection logic, but it should be named as a quote to avoid implying a mutation.

## Pagination Contract

All paginated MCP tools use:

- default `limit`: `20`
- maximum `limit`: `100`
- default `offset`: `0`
- minimum `offset`: `0`

Responses use a consistent wrapper:

```json
{
  "items": [],
  "page": {
    "limit": 20,
    "offset": 0,
    "count": 0,
    "nextOffset": null,
    "hasMore": false,
    "total": null
  }
}
```

Rules:

- `count` is the number of returned items.
- `nextOffset` is `offset + count` when `hasMore` is true; otherwise `null`.
- `hasMore` is true when `count == limit` and a backend total is not available.
- `total` is set only when the underlying service knows the full total.
- For grouped market activity tools, use the service-provided total.
- For single-market bets, positions, and leaderboard tools, do not invent a full total if the service does not return one.

## Volume Semantics

Market responses should expose:

- `totalVolume`
- `marketDust`

The backend currently maps public volume from accounting `VolumeWithDust`, with `marketDust` separately available. For grouped market discovery rows, the public overview aggregates child market volume and dust. V1 should expose these fields but should not add server-side volume sorting or filters.

Agents can rank returned markets locally by `totalVolume`. Server-side arguments such as `minVolume`, `maxVolume`, or `sort=volume` are future extensions.

## Error Handling

Tool errors should be structured and stable:

- Invalid IDs, statuses, outcomes, limits, offsets, or tag slugs return validation errors.
- Missing market/group/tag returns not-found errors.
- Closed markets can return a conflict-style error for `quote_market_probability`.
- Unexpected service/repository failures return internal errors without leaking database details.

MCP errors should include a stable code and a short message. Use backend domain errors as the source of truth for mapping.

## Deployment

Use a separate container/service for MCP:

- Same repository and Go module.
- Same backend image can include both `server` and `mcpserver` binaries.
- Compose/deployment config starts a distinct MCP process.
- Proxy routes `/mcp` to the MCP service.

This keeps MCP sessions, streaming behavior, rate limits, and failures separate from the main web API while avoiding duplicate build contexts.

## Testing

Implementation should include:

- Unit tests for status normalization, including `open -> active`.
- Unit tests for tag slug normalization and `get_market_tag` not-found behavior.
- Tool handler tests for each public discovery tool.
- Pagination tests for default limit, max limit, offset, `hasMore`, `nextOffset`, and total behavior.
- Service-adapter tests proving MCP calls the existing market/tag/discovery services instead of duplicating logic.
- Error mapping tests for validation, not found, conflict, and internal errors.
- Build test for the new `mcpserver` binary.

Integration tests should run only if the existing test infrastructure can boot the backend container/database reliably. V1 does not require external network access.

## Implementation Order

1. Add the MCP server binary and runtime package skeleton.
2. Wire configuration, database, repositories, and domain services using existing backend patterns.
3. Add shared MCP argument normalization and page response helpers.
4. Register public market tag and discovery/list/search tools.
5. Register market detail and summary tools.
6. Register paginated activity and leaderboard tools.
7. Register `quote_market_probability`.
8. Add Docker/compose/proxy wiring.
9. Add focused tests and run the relevant backend test suite.

## Future Extension Points

Future work can add:

- API key storage and web API endpoints for key management.
- User-authenticated MCP tools backed by JWT or API keys.
- Service identity for liquidity/service integrations.
- Private portfolio, balance, and authenticated user-position tools.
- Write tools for placing bets or selling positions, gated behind explicit private auth.
- Server-side volume sorting/filtering for large discovery result sets.
