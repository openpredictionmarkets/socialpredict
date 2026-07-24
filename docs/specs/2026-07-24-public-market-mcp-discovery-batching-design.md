# Public Market MCP Discovery Batching Design

## Context

The public MCP discovery tools currently enrich each market with two service calls:

1. `GetMarketSummaryReadModel` loads the market, accounting snapshot, and creator.
2. `GetMarketGroupForMarket` reloads group membership.

A request with `limit=100` can produce hundreds of database queries. The group lookup also discards every error, so repository failures can return incomplete successful responses.

The discovery repository already groups markets and loads each row's parent group before the MCP layer receives it. The MCP layer should reuse that data and obtain the remaining display data through a batched read-model API.

## Decision

Add a batched discovery-enrichment API to the markets domain and repository layers.

The repository will load accounting snapshots and public creator fields for all requested markets with bounded `IN` queries. The domain service will assemble `MarketSummaryReadModel` values using the hydrated markets supplied by discovery. It will preserve the existing behavior for missing accounting snapshots by refreshing those markets once and adding the new snapshots to the result.

The MCP runtime will request one enrichment set for each discovery result set. It will map summaries by market ID and derive group links from `MarketDiscoveryRow.Group`. It will stop calling `GetMarketSummaryReadModel` and `GetMarketGroupForMarket` inside the row loop.

## Interfaces

The markets domain will expose a display-only method shaped like:

```go
GetMarketDiscoverySummaries(
    ctx context.Context,
    markets []*Market,
) (map[int64]*MarketSummaryReadModel, error)
```

The repository capability behind it will accept deduplicated market IDs and creator usernames and return:

- accounting snapshots keyed by market ID;
- public creator summaries keyed by username.

The capability remains separate from transaction repository interfaces. Betting, selling, resolution, payout, and refund code must not gain access to display snapshots.

The MCP `MarketService` interface will consume the new domain method. Existing single-market summary and group lookup methods remain available for tools that inspect one market.

## Data Flow

1. `ListMarketDiscovery` or `SearchMarketDiscovery` returns grouped discovery rows.
2. The MCP runtime collects unique child and standalone markets from those rows.
3. The markets service asks the repository for accounting snapshots and creator data in batches.
4. The service refreshes only missing snapshots, preserving the current cold-read contract.
5. The service returns summaries keyed by market ID.
6. The MCP mapper joins summaries to rows in memory.
7. Grouped rows use their existing `MarketDiscoveryRow.Group` value to build each child market's group link.

Normal requests use a fixed number of repository queries independent of page size. Only missing snapshots trigger per-market refresh work, and subsequent reads use the stored snapshots.

## Error Handling

Repository batch failures return to the MCP boundary and pass through `MapError`. The MCP response must not silently omit accounting, creator, or group data after an unexpected repository failure.

A missing summary for a valid market after refresh is an internal-state error. Invalid or nil markets remain safe to map as empty values, matching current mapper behavior.

The MCP layer performs no group lookup during enrichment. The discovery repository already returns errors when group membership or group loading fails, so those failures propagate from the original discovery call.

## Merge With Main

The branch will merge current `origin/main`. The known conflict is `.gitignore`; resolution will retain both main's entries and the branch's agent-artifact ignores. Main's README and stargazer changes will remain unchanged.

## Testing

Tests will establish these behaviors before production changes:

- repository batch reads return snapshots and creators keyed by identity;
- domain batching deduplicates inputs and preserves hydrated market values;
- missing snapshots refresh once and appear in the returned map;
- MCP discovery uses one batch-enrichment call for a multi-market page;
- grouped child links come from the discovery row without group service calls;
- batch failures return mapped MCP errors;
- list, search, and discovery-page outputs retain pagination, volume, dust, creator, and group fields.

After focused tests pass, run the markets domain, markets repository, MCP package, both backend builds, and the full backend test suite.

## Scope

This change does not add caching, rate limiting, authentication, write tools, new endpoints, or transaction behavior. Request limiting and production query metrics remain deployment follow-ups.
