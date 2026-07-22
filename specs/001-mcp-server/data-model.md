# Data Model: MCP Server

**Feature**: [spec.md](spec.md) | **Date**: 2026-07-22

## Persistence

**None.** This feature introduces no new database tables, GORM models,
migrations, or writes. All data is read through the existing `dmarkets` domain
service. (Constitution I/IV: no persistence types leak; no financial state
touched.)

## Consumer-Defined Interface (adapter boundary)

The MCP adapter depends on a narrow, consumer-side interface (constitution II),
satisfied by the existing markets domain service:

```text
MarketReader (defined in backend/handlers/mcp/interfaces.go)
├── ListMarkets(ctx, dmarkets.ListFilters) ([]*dmarkets.Market, error)
├── GetMarketDetails(ctx, marketID int64) (*dmarkets.MarketOverview, error)
├── SearchMarkets(ctx, query string, dmarkets.SearchFilters) (*dmarkets.SearchResults, error)
└── GetMarketBets(ctx, marketID int64) ([]*dmarkets.BetDisplayInfo, error)
```

Source domain types (existing, unchanged): `dmarkets.Market`,
`dmarkets.MarketOverview` (LastProbability, NumUsers, TotalVolume, Creator),
`dmarkets.SearchResults`, `dmarkets.BetDisplayInfo`
(`backend/internal/domain/markets/models.go`, `service.go`).

## Output DTOs (MCP tool results)

Adapter-owned structs, JSON-serialized into MCP tool results. They expose only
fields visible to a logged-out website visitor (FR-003) — internal moderation
fields (`ApprovedBy`, `RejectedBy`, `RejectionReason`, `ProposalCost`, etc.)
are deliberately excluded.

### MarketSummary (element of `list_markets` / `search_markets` results)

| Field | Type | Source |
|-------|------|--------|
| id | int64 | Market.ID |
| question | string | Market.QuestionTitle |
| status | string | Market.Status (`active` / `closed` / `resolved`) |
| outcomeType | string | Market.OutcomeType (e.g. `BINARY`) |
| yesLabel / noLabel | string | Market.YesLabel / NoLabel |
| creatorUsername | string | Market.CreatorUsername |
| resolutionDateTime | RFC 3339 string | Market.ResolutionDateTime |
| resolutionResult | string | Market.ResolutionResult (empty until resolved) |
| createdAt | RFC 3339 string | Market.CreatedAt |

### MarketDetail (result of `get_market`)

MarketSummary fields, plus:

| Field | Type | Source |
|-------|------|--------|
| description | string | Market.Description |
| lastProbability | float64 | MarketOverview.LastProbability |
| totalVolume | int64 | MarketOverview.TotalVolume |
| numUsers | int | MarketOverview.NumUsers |
| creator | CreatorSummary | MarketOverview.Creator (username, displayName, personalEmoji) |

### BetInfo (element of `list_market_bets` result)

| Field | Type | Source |
|-------|------|--------|
| username | string | BetDisplayInfo.Username |
| outcome | string | BetDisplayInfo.Outcome |
| amount | int64 | BetDisplayInfo.Amount (integer points — constitution IV) |
| probability | float64 | BetDisplayInfo.Probability |
| placedAt | RFC 3339 string | BetDisplayInfo.PlacedAt |

### SearchResult (result of `search_markets`)

| Field | Type | Source |
|-------|------|--------|
| query | string | SearchResults.Query |
| results | []MarketSummary | SearchResults.PrimaryResults |
| fallbackResults | []MarketSummary | SearchResults.FallbackResults |
| totalCount | int | SearchResults.TotalCount |

## Validation Rules

- `get_market` / `list_market_bets`: `id` must be a positive integer; unknown
  id → tool error "market not found" (no stack traces, no internals — FR-004).
- `list_markets`: optional `status` must be one of `active|closed|resolved|all`
  (default `all`); optional `limit` clamped to [1, 100].
- `search_markets`: `query` required, non-empty after trimming; optional
  `status`, `limit` as above.
- All tools: reject unknown arguments per JSON Schema (`additionalProperties: false`).

## State Transitions

None owned by this feature. Market status transitions happen elsewhere; MCP
reads whatever the domain service returns at call time (spec edge case:
no stale-state guarantees beyond the website's).

## Invariants (tested)

1. Tool catalog contains exactly: `list_markets`, `get_market`,
   `search_markets`, `list_market_bets` — nothing else (FR-003, SC-003).
2. No DTO field maps from moderation/internal-only domain fields.
3. Zero writes: the adapter holds no reference to any interface with mutating
   methods (`MarketReader` is read-only by construction).
