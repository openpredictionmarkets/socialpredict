# Public Market MCP Discovery Batching Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `$subagent-driven-development` (recommended, if installed) or `$executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove per-market MCP discovery queries by adding a batched display read model, propagate enrichment failures, synchronize the PR branch with `main`, and restore mergeability.

**Architecture:** The markets repository loads accounting snapshots and public creator fields with two bounded queries. The markets service joins those results to discovery markets and refreshes only missing snapshots. MCP discovery requests call the batch service once and derive group links from `MarketDiscoveryRow.Group`.

**Tech Stack:** Go 1.25, GORM, SQLite/Postgres-compatible queries, existing markets domain services, MCP Go SDK, Go tests, GitHub PR #783.

## Global Constraints

- Keep V1 public tools anonymous and read-only.
- Keep display read models outside transaction repository interfaces.
- Preserve existing single-market summary behavior.
- Preserve missing-snapshot refresh behavior.
- Preserve market tags, canonical statuses, pagination, `VolumeWithDust` as `totalVolume`, and separate `marketDust`.
- Do not add caching, rate limiting, authentication, write tools, endpoints, or schema migrations.
- Do not reintroduce `.codebase-graph/`, `.superpowers/`, or `docs/superpowers/`.
- Resolve `main` conflicts by retaining both upstream changes and branch artifact-ignore rules.

---

## Task 1: Synchronize The PR Branch With Main

**Files:**
- Modify: `.gitignore`
- Preserve: upstream README, metrics, workflow, and stargazer files

**Interfaces:**
- Consumes: `origin/main`
- Produces: a merge commit with no unresolved files

- [ ] **Step 1: Fetch and inspect the conflict**

```bash
git fetch origin main
BASE="$(git merge-base HEAD origin/main)"
git merge-tree "$BASE" HEAD origin/main
```

Expected: `.gitignore` is the only file changed on both sides. Upstream README and stargazer files are additions or changes from `main`.

- [ ] **Step 2: Merge current main**

```bash
git merge --no-edit origin/main
```

Expected: Git stops on `.gitignore`.

- [ ] **Step 3: Resolve `.gitignore`**

Retain every upstream entry and these branch entries:

```gitignore
.worktrees/

# Agent workflow artifacts
/.codebase-graph/
/.superpowers/
/docs/superpowers/
```

Do not edit upstream README, metrics, workflow, or stargazer files.

- [ ] **Step 4: Verify and finish the merge**

```bash
git diff --check
git status --short
git add .gitignore
git commit
```

Expected: the merge commit completes with no unresolved paths.

- [ ] **Step 5: Run the MCP baseline**

```bash
cd backend
GOCACHE=/private/tmp/socialpredict-mcp-batch-gocache go test ./internal/mcpserver ./internal/domain/markets ./internal/repository/markets -count=1
```

Expected: PASS before batching changes.

---

## Task 2: Add Repository Batch Enrichment

**Files:**
- Modify: `backend/internal/domain/markets/service.go`
- Modify: `backend/internal/repository/markets/market_accounting_snapshots.go`
- Modify: `backend/internal/repository/markets/market_accounting_snapshots_test.go`

**Interfaces:**
- Produces:

```go
type MarketDiscoveryEnrichment struct {
    AccountingByMarketID map[int64]MarketAccountingSnapshot
    CreatorsByUsername    map[string]CreatorSummary
}

type MarketDiscoveryEnrichmentRepository interface {
    GetMarketDiscoveryEnrichment(
        ctx context.Context,
        marketIDs []int64,
        creatorUsernames []string,
    ) (*MarketDiscoveryEnrichment, error)
}
```

- [ ] **Step 1: Write failing repository tests**

Add tests that create two users, two markets, and two accounting snapshots, then call:

```go
got, err := repo.GetMarketDiscoveryEnrichment(
    context.Background(),
    []int64{marketA.ID, marketB.ID, marketA.ID, 0},
    []string{creatorA.Username, creatorB.Username, creatorA.Username, " "},
)
```

Assert:

```go
if err != nil {
    t.Fatalf("GetMarketDiscoveryEnrichment returned error: %v", err)
}
if len(got.AccountingByMarketID) != 2 {
    t.Fatalf("accounting count = %d, want 2", len(got.AccountingByMarketID))
}
if got.AccountingByMarketID[marketA.ID].VolumeWithDust != snapshotA.VolumeWithDust {
    t.Fatalf("market A accounting = %#v", got.AccountingByMarketID[marketA.ID])
}
if len(got.CreatorsByUsername) != 2 {
    t.Fatalf("creator count = %d, want 2", len(got.CreatorsByUsername))
}
if got.CreatorsByUsername[creatorA.Username].DisplayName != creatorA.DisplayName {
    t.Fatalf("creator A = %#v", got.CreatorsByUsername[creatorA.Username])
}
```

Add a second test:

```go
got, err := repo.GetMarketDiscoveryEnrichment(context.Background(), nil, nil)
if err != nil {
    t.Fatalf("empty enrichment returned error: %v", err)
}
if got == nil || got.AccountingByMarketID == nil || got.CreatorsByUsername == nil {
    t.Fatalf("empty enrichment = %#v", got)
}
```

- [ ] **Step 2: Verify RED**

```bash
cd backend
GOCACHE=/private/tmp/socialpredict-mcp-batch-gocache go test ./internal/repository/markets -run TestGormRepositoryGetMarketDiscoveryEnrichment -count=1
```

Expected: FAIL because the type and method do not exist.

- [ ] **Step 3: Add the display-only domain types**

In `backend/internal/domain/markets/service.go`, add the exact types from the Interfaces block beside `MarketAccountingSnapshotRepository`. Do not embed the new interface in `Repository`.

- [ ] **Step 4: Implement the repository method**

Add:

```go
func (r *GormRepository) GetMarketDiscoveryEnrichment(
    ctx context.Context,
    marketIDs []int64,
    creatorUsernames []string,
) (*dmarkets.MarketDiscoveryEnrichment, error) {
    out := &dmarkets.MarketDiscoveryEnrichment{
        AccountingByMarketID: map[int64]dmarkets.MarketAccountingSnapshot{},
        CreatorsByUsername:    map[string]dmarkets.CreatorSummary{},
    }

    marketIDs = uniquePositiveMarketIDs(marketIDs)
    if len(marketIDs) > 0 {
        var snapshots []models.MarketAccountingSnapshot
        if err := r.db.WithContext(ctx).
            Where("market_id IN ?", marketIDs).
            Find(&snapshots).Error; err != nil {
            return nil, err
        }
        for i := range snapshots {
            snapshot := modelAccountingSnapshotToDomain(&snapshots[i])
            out.AccountingByMarketID[snapshot.MarketID] = *snapshot
        }
    }

    creatorUsernames = uniqueNonBlankStrings(creatorUsernames)
    if len(creatorUsernames) > 0 {
        var users []models.User
        if err := r.db.WithContext(ctx).
            Select("username", "display_name", "personal_emoji").
            Where("username IN ?", creatorUsernames).
            Find(&users).Error; err != nil {
            return nil, err
        }
        for _, user := range users {
            out.CreatorsByUsername[user.Username] = dmarkets.CreatorSummary{
                Username: user.Username,
                DisplayName: user.DisplayName,
                PersonalEmoji: user.PersonalEmoji,
            }
        }
    }
    return out, nil
}
```

Add focused helpers in the same file:

```go
func uniquePositiveMarketIDs(values []int64) []int64 {
    seen := map[int64]struct{}{}
    out := make([]int64, 0, len(values))
    for _, value := range values {
        if value <= 0 {
            continue
        }
        if _, ok := seen[value]; ok {
            continue
        }
        seen[value] = struct{}{}
        out = append(out, value)
    }
    return out
}

func uniqueNonBlankStrings(values []string) []string {
    seen := map[string]struct{}{}
    out := make([]string, 0, len(values))
    for _, value := range values {
        value = strings.TrimSpace(value)
        if value == "" {
            continue
        }
        if _, ok := seen[value]; ok {
            continue
        }
        seen[value] = struct{}{}
        out = append(out, value)
    }
    return out
}
```

Add `strings` to the file imports.

- [ ] **Step 5: Verify GREEN**

```bash
cd backend
GOCACHE=/private/tmp/socialpredict-mcp-batch-gocache go test ./internal/repository/markets -run TestGormRepositoryGetMarketDiscoveryEnrichment -count=1
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/domain/markets/service.go backend/internal/repository/markets/market_accounting_snapshots.go backend/internal/repository/markets/market_accounting_snapshots_test.go
git commit -m "feat: batch market discovery enrichment reads"
```

---

## Task 3: Add The Domain Batch Summary Service

**Files:**
- Modify: `backend/internal/domain/markets/market_accounting_snapshot_refresh.go`
- Modify: `backend/internal/domain/markets/market_accounting_snapshot_refresh_test.go`

**Interfaces:**
- Consumes: `MarketDiscoveryEnrichmentRepository`
- Produces:

```go
func (s *Service) GetMarketDiscoverySummaries(
    ctx context.Context,
    markets []*Market,
) (map[int64]*MarketSummaryReadModel, error)
```

- [ ] **Step 1: Write failing service tests**

Create a focused fake repository that implements `MarketDiscoveryEnrichmentRepository`, records its arguments, and returns one stored snapshot and creator.

Test deduplication and hydrated-market preservation:

```go
got, err := service.GetMarketDiscoverySummaries(ctx, []*markets.Market{
    marketA,
    marketB,
    marketA,
    nil,
})
if err != nil {
    t.Fatalf("GetMarketDiscoverySummaries returned error: %v", err)
}
if len(fake.marketIDs) != 2 || len(fake.creatorUsernames) != 2 {
    t.Fatalf("batch args = %#v %#v", fake.marketIDs, fake.creatorUsernames)
}
if got[marketA.ID].Market != marketA {
    t.Fatalf("service replaced hydrated market")
}
if got[marketA.ID].Creator.DisplayName != "Alice" {
    t.Fatalf("creator = %#v", got[marketA.ID].Creator)
}
```

Add a database-backed test with one missing snapshot. Call the method twice and assert that the first call creates the snapshot and the second returns the stored value without changing its generated timestamp.

Add an error test where the batch repository returns `errBatch`; assert `errors.Is(err, errBatch)`.

- [ ] **Step 2: Verify RED**

```bash
cd backend
GOCACHE=/private/tmp/socialpredict-mcp-batch-gocache go test ./internal/domain/markets -run TestServiceGetMarketDiscoverySummaries -count=1
```

Expected: FAIL because the method does not exist.

- [ ] **Step 3: Implement the service method**

Add:

```go
func (s *Service) GetMarketDiscoverySummaries(
    ctx context.Context,
    input []*Market,
) (map[int64]*MarketSummaryReadModel, error) {
    repo, ok := s.repo.(MarketDiscoveryEnrichmentRepository)
    if !ok {
        return nil, ErrInvalidState
    }

    marketsByID := map[int64]*Market{}
    marketIDs := make([]int64, 0, len(input))
    creators := make([]string, 0, len(input))
    for _, market := range input {
        if market == nil || market.ID <= 0 {
            continue
        }
        if _, exists := marketsByID[market.ID]; exists {
            continue
        }
        marketsByID[market.ID] = market
        marketIDs = append(marketIDs, market.ID)
        creators = append(creators, market.CreatorUsername)
    }

    enrichment, err := repo.GetMarketDiscoveryEnrichment(ctx, marketIDs, creators)
    if err != nil {
        return nil, err
    }
    if enrichment == nil {
        return nil, ErrInvalidState
    }

    out := make(map[int64]*MarketSummaryReadModel, len(marketsByID))
    for _, marketID := range marketIDs {
        market := marketsByID[marketID]
        accounting, found := enrichment.AccountingByMarketID[marketID]
        if !found {
            refreshed, err := s.RefreshMarketAccountingSnapshot(ctx, marketID)
            if err != nil {
                return nil, err
            }
            accounting = *refreshed
        }
        creator := CreatorSummary{Username: market.CreatorUsername}
        if enriched, found := enrichment.CreatorsByUsername[market.CreatorUsername]; found {
            creator = enriched
        }
        out[marketID] = &MarketSummaryReadModel{
            Market: market,
            Creator: &creator,
            Accounting: accounting,
        }
    }
    return out, nil
}
```

- [ ] **Step 4: Verify GREEN**

```bash
cd backend
GOCACHE=/private/tmp/socialpredict-mcp-batch-gocache go test ./internal/domain/markets -run 'TestServiceGetMarketDiscoverySummaries|TestServiceGetMarketSummaryReadModel' -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/domain/markets/market_accounting_snapshot_refresh.go backend/internal/domain/markets/market_accounting_snapshot_refresh_test.go
git commit -m "feat: assemble discovery summaries in batches"
```

---

## Task 4: Switch MCP Discovery To Batch Enrichment

**Files:**
- Modify: `backend/internal/mcpserver/runtime.go`
- Modify: `backend/internal/mcpserver/discovery_tools.go`
- Modify: `backend/internal/mcpserver/discovery_tools_test.go`
- Modify test doubles that implement `mcpserver.MarketService`

**Interfaces:**
- Consumes:

```go
GetMarketDiscoverySummaries(
    context.Context,
    []*dmarkets.Market,
) (map[int64]*dmarkets.MarketSummaryReadModel, error)
```

- Produces: discovery mapping with one batch call and no per-market group lookup

- [ ] **Step 1: Write failing MCP tests**

Extend `discoveryToolMarketService` with:

```go
batchCalls   int
batchMarkets []*dmarkets.Market
batchErr     error
groupCalls   int
```

Implement its new interface method:

```go
func (s *discoveryToolMarketService) GetMarketDiscoverySummaries(
    _ context.Context,
    markets []*dmarkets.Market,
) (map[int64]*dmarkets.MarketSummaryReadModel, error) {
    s.batchCalls++
    s.batchMarkets = append([]*dmarkets.Market(nil), markets...)
    if s.batchErr != nil {
        return nil, s.batchErr
    }
    out := map[int64]*dmarkets.MarketSummaryReadModel{}
    for _, market := range markets {
        out[market.ID] = s.summaryFor(market)
    }
    return out, nil
}
```

Add a multi-market grouped test that asserts:

```go
if svc.batchCalls != 1 {
    t.Fatalf("batch calls = %d, want 1", svc.batchCalls)
}
if svc.groupCalls != 0 {
    t.Fatalf("group calls = %d, want 0", svc.groupCalls)
}
if got.Results.Items[0].ChildMarkets[0].Market.MarketGroup.ID != group.ID {
    t.Fatalf("group link missing: %#v", got.Results.Items[0])
}
```

Add an error test:

```go
svc.batchErr = errors.New("batch failed")
_, _, err := NewRuntime(svc, nil).ListMarkets(ctx, nil, MarketListInput{})
if got := MapError(err); got.Code != "internal_error" {
    t.Fatalf("error = %#v", got)
}
```

- [ ] **Step 2: Verify RED**

```bash
cd backend
GOCACHE=/private/tmp/socialpredict-mcp-batch-gocache go test ./internal/mcpserver -run 'TestListMarketsUsesSingleBatchEnrichment|TestListMarketsMapsBatchEnrichmentError' -count=1
```

Expected: FAIL because runtime still performs per-market calls and the interface lacks the batch method.

- [ ] **Step 3: Extend `MarketService`**

Add the exact batch signature from the Interfaces block to `backend/internal/mcpserver/runtime.go`. Keep `GetMarketSummaryReadModel` for `get_market_summary`. Remove `GetMarketGroupForMarket` only if no remaining MCP production code uses it; otherwise keep it without calling it from discovery.

- [ ] **Step 4: Replace row-loop service calls**

Add a collector:

```go
func discoveryMarkets(rows []dmarkets.MarketDiscoveryRow) []*dmarkets.Market {
    out := make([]*dmarkets.Market, 0, len(rows))
    seen := map[int64]struct{}{}
    add := func(market *dmarkets.Market) {
        if market == nil || market.ID <= 0 {
            return
        }
        if _, ok := seen[market.ID]; ok {
            return
        }
        seen[market.ID] = struct{}{}
        out = append(out, market)
    }
    for _, row := range rows {
        if row.Group != nil && row.Group.ID > 0 {
            for _, child := range row.Children {
                add(child)
            }
            if len(row.Children) == 0 {
                add(row.Market)
            }
            continue
        }
        add(row.Market)
    }
    return out
}
```

Change `discoveryRowOutputs` to call the service once:

```go
summaries, err := rt.markets.GetMarketDiscoverySummaries(ctx, discoveryMarkets(rows))
if err != nil {
    return nil, err
}
```

Pass `summaries` into pure row mapping. For each valid market, require a summary:

```go
summary, ok := summaries[market.ID]
if !ok || summary == nil {
    return MarketOverviewOutput{}, dmarkets.ErrInvalidState
}
```

Build the overview from the summary accounting fields. When `row.Group != nil`, set:

```go
overview.Market.MarketGroup = MarketGroupLinkOutputFromDomain(row.Group, market.ID)
```

Delete the per-market `GetMarketSummaryReadModel` and `GetMarketGroupForMarket` calls from `marketOverviewOutput`.

- [ ] **Step 5: Update all MCP test doubles**

Use this safe default in test services that do not exercise discovery:

```go
func (s *serviceStub) GetMarketDiscoverySummaries(
    context.Context,
    []*dmarkets.Market,
) (map[int64]*dmarkets.MarketSummaryReadModel, error) {
    return map[int64]*dmarkets.MarketSummaryReadModel{}, nil
}
```

Do not add per-market calls to test doubles.

- [ ] **Step 6: Verify GREEN**

```bash
cd backend
GOCACHE=/private/tmp/socialpredict-mcp-batch-gocache go test ./internal/mcpserver -count=1
```

Expected: PASS, including list, search, discovery layout, pagination, total volume, dust, creator, and group-link assertions.

- [ ] **Step 7: Commit**

```bash
git add backend/internal/mcpserver
git commit -m "fix: batch public MCP discovery enrichment"
```

---

## Task 5: Final Verification And PR Update

**Files:**
- Modify only files from Tasks 1-4 if verification finds a real defect

**Interfaces:**
- Consumes: merged `origin/main` plus completed batching changes
- Produces: green PR branch and mergeable PR #783

- [ ] **Step 1: Run focused tests**

```bash
cd backend
GOCACHE=/private/tmp/socialpredict-mcp-batch-gocache go test \
  ./internal/repository/markets \
  ./internal/domain/markets \
  ./internal/mcpserver \
  -count=1
```

Expected: PASS.

- [ ] **Step 2: Build both binaries**

```bash
cd backend
GOCACHE=/private/tmp/socialpredict-mcp-batch-gocache go build .
GOCACHE=/private/tmp/socialpredict-mcp-batch-gocache go build ./cmd/mcpserver
rm -f socialpredict mcpserver
```

Expected: both builds exit 0 and generated binaries are removed.

- [ ] **Step 3: Run the full backend suite**

```bash
cd backend
GOCACHE=/private/tmp/socialpredict-mcp-batch-gocache go test ./... -count=1
```

Expected: PASS.

- [ ] **Step 4: Inspect branch hygiene**

```bash
git status --short
git diff --check origin/main..HEAD
git ls-files .codebase-graph .superpowers docs/superpowers
```

Expected: clean status, no whitespace errors, and no tracked agent artifacts.

- [ ] **Step 5: Push and verify the PR**

```bash
git push origin codex/public-market-mcp-impl
gh pr view 783 --repo openpredictionmarkets/socialpredict \
  --json url,state,mergeable,statusCheckRollup
```

Expected: PR remains open and GitHub no longer reports `CONFLICTING`. Checks may be queued or in progress after the push.

## Plan Self-Review

- **Spec coverage:** Tasks cover the `main` merge, repository batching, domain assembly, missing-snapshot refresh, MCP integration, error propagation, group links, verification, and PR update.
- **Placeholder scan:** The plan contains no incomplete markers or unspecified code steps.
- **Type consistency:** `GetMarketDiscoverySummaries` uses the same map type in the markets service and MCP interface. Repository maps use `int64` market IDs and string usernames throughout.
- **Scope:** The plan adds no schema, endpoint, auth, caching, rate-limit, or transaction changes.
