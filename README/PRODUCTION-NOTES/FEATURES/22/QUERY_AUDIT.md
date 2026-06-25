---
title: GORM Query Scope Regex Audit
document_type: feature-audit
domain: features
author: Patrick Delaney
updated_at: 2026-06-25T00:00:00Z
updated_at_display: "Thursday, June 25, 2026"
update_reason: "Record repository-wide regex audit findings for GORM query scoping and first implementation actions."
status: in_progress
---

# GORM Query Scope Regex Audit

This audit records the repository-wide regex pass used for Feature 22. The goal is not to replace profiling. The goal is to catch obvious broad reads, missing predicates, unstable ordering, over-wide projections, and N+1 hydration patterns before deeper performance work.

## Search Commands

```bash
rg -n '\.Find\(&|\.First\(&|\.Last\(&|\.Take\(&|\.Scan\(&|\.Count\(&|\.Raw\(|\.Exec\(|\.Joins\(|\.Preload\(|\.Model\(&|\.Table\(' backend/internal backend/handlers backend/server backend/seed backend/migration backend/models -g '*.go'
rg -n 'models\.Bet|Table\("bets"\)|Model\(&models\.Bet|\[\]models\.Bet|Find\(&bets|Where\("market_id|Where\("username' backend -g '*.go'
rg -n '\.Find\(&|\.First\(&|\.Scan\(&|\.Count\(&|\.Raw\(|\.Table\("bets"\)|Model\(&models\.Bet|\[\]models\.Bet' backend/internal backend/handlers backend/cmd backend/seed backend/server -g '*.go' -g '!*_test.go'
```

## Findings

| Area | Query shape found | Risk | Action in PR #745 | Remaining action |
| --- | --- | --- | --- | --- |
| Market replay reads | `WHERE market_id = ?` with chronological replay in market, analytics, sale, and user-position adapters. | Correctly scoped, but timestamp ties could replay differently. | Added `placed_at ASC, id ASC`. Added `idx_bets_market_id_placed_at_id`. Added scoped-read tests. | Consider narrower replay projections for transaction adapters only after math-equivalence tests. |
| User bet history display | `WHERE username = ? ORDER BY placed_at DESC` hydrating full `models.Bet`. | Over-wide projection and no stable tie order. | Added `Select("market_id", "placed_at")`, `ORDER BY placed_at DESC, id DESC`, and `idx_bets_username_placed_at_id`. Expanded repository test. | If profile history grows large, add pagination at the repository/domain port. |
| User financial positions | Starts from `WHERE username = ?`, then loads only affected `market_id IN ?`. | Good scope; still replays full affected markets to preserve canonical position math. | Existing `idx_bets_username_market_id_placed_at_id` and `idx_bets_market_id_placed_at_id` support the two-phase read. | Do not shortcut transaction or position math without equivalence tests. |
| User-has-bet / first participation checks | `WHERE market_id = ? AND username = ? COUNT`. | Correctly scoped; needs composite index. | Added `idx_bets_market_id_username`. | A later aggregate method can count first positive participants without materializing rows. |
| Global analytics bet reads | `ListBetsOrdered` intentionally scans all bets for platform-wide metrics/leaderboards. | Legitimate global read, but expensive as platform grows. | Documented as intentionally global. Existing deterministic order already includes `market_id, placed_at, id`. | Move more global metrics to read-model snapshots or SQL aggregates under Feature 11/22 continuation. |
| Market group work-profit hydration | Groups were loaded, then members were queried once per group. | N+1 query pattern for grouped market financial metrics. | Collapsed to one `WHERE group_id IN ?` member query. Existing `idx_market_group_members_group_order` supports it. Added repository test. | None for this slice. |
| Market discovery `/markets` and `/markets/topic/:slug` | Loads matching markets, hydrates group/tag data, groups in Go, then paginates grouped rows. | Correct grouped pagination semantics, but can materialize too many rows for broad discovery pages. | Documented as the largest remaining scaling candidate. | Build a backend grouped discovery read model or SQL grouped-row query before high-volume production use. |
| Admin market review | Raw grouped page key query with `COUNT`, `LIMIT`, `OFFSET`, then hydrates selected keys. | Good server-side pagination shape; search uses `LOWER(... LIKE)` across text. | No code change needed in this slice. | Consider trigram/full-text search only if admin search latency becomes visible. |
| Description amendment review | Raw grouped page key query with `COUNT`, `LIMIT`, `OFFSET`, then hydrates selected keys. | Good server-side pagination shape; text search may grow. | No code change needed in this slice. | Same search-index caveat as admin review. |
| Answer addition review | GORM count + limited page query with optional group join. | Acceptable scoped admin/moderator queue. | Existing model indexes cover group/status and status/created. | Consider additional reviewer/steward indexes only after queue latency appears. |
| CMS/settings/read-model lookups | Slug/key lookups on small singleton/config tables. | Low volume and already scoped. | No change. | None. |
| Dev bootstrap/load-test seed commands | Utility queries by username/title/tag. | Not production request path. | No change. | None. |
| Tests/migrations | Broad GORM calls in test setup and migration plumbing. | Not production path. | Excluded from production-action list. | None. |

## Current Conclusion

The original concern that GORM might load the full `bets` table before applying `Where("market_id = ?")` is not accurate for these adapter calls. GORM generates SQL predicates and Postgres/SQLite apply those predicates in the database.

The useful optimizations from this pass are therefore more specific:

- keep per-market replay reads SQL-scoped;
- make replay ordering deterministic with `id` tie-breaks;
- add composite indexes that match the actual predicates and ordering;
- select fewer columns on display-only paths;
- remove obvious N+1 hydration where the result is still the same domain record;
- document intentionally global analytics reads rather than pretending every global scan is a bug.

## Follow-Up Candidates

| Candidate | Why not in this slice? | Entry criterion |
| --- | --- | --- |
| Grouped discovery SQL/read-model pagination | More invasive; must preserve grouped-market row semantics, tag filtering, and topic search. | `/markets` or `/markets/topic/:slug` p95/row materialization becomes a measured bottleneck. |
| Probability-history display read model | Changes display freshness path and needs invalidation/freshness metadata. | Single-market bet history grows enough that chart rendering repeatedly replays too many rows. |
| SQL aggregate first-participation counts | Safe if carefully named, but must match fee/work-profit policy around positive participation and grouped children. | System metrics or work-profit reporting shows global bet replay pressure. |
| Full-text/trigram search indexes | Deployment-specific Postgres feature choice. | Search latency is measured as a user/admin bottleneck. |
| Additional lifecycle/status/created indexes | Could help discovery/admin queues but adds write/storage cost. | Query plan or p95 shows status/time filters are scanning too much. |
