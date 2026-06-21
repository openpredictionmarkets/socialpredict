# Test Report — Multiple Choice Binary Markets

**Date:** 2026-06-15
**Original manual branch:** `feature/test-binary-market`
**Current follow-up branch:** `feature/multiple-choice-binary-markets`
**Environment:** `APP_ENV=development` (locally built Docker images only)

## Environment Setup

All containers were built from **local source** (no registry pulls):

| Image | Source | Built |
|-------|--------|-------|
| `socialpredict-dev-backend:latest` | `docker/backend/Dockerfile.dev` | local |
| `socialpredict-dev-frontend:latest` | `docker/frontend/Dockerfile.dev` | local |
| `postgres:16.6-alpine` | base image (DB only) | — |

Commands used:

```
./SocialPredict install -e development      # builds both dev images locally (--no-cache)
./SocialPredict up                          # docker-compose-dev.yaml
./SocialPredict dev-bootstrap-users         # admin + testuser01..10
```

> Note: the `localhost` and `production` environments pull pre-built images from `ghcr.io`. Only `development` builds locally, so it was used.

Containers verified healthy:

```
socialpredict-backend-container    Up   0.0.0.0:8080->8080
socialpredict-frontend-container   Up   0.0.0.0:5173->5173
socialpredict-postgres-container   Up (healthy)   0.0.0.0:5432->5432
```

Seed accounts (password `password`): `admin` (ADMIN), `testuser01` (active MODERATOR / market steward), `testuser02..10` (REGULAR). `testuser02` was promoted to MODERATOR in the DB to exercise the **non-steward moderator** path in TC12/TC13. Relevant economics from `setup.yaml`: `createMarketCost=10`, `initialBetFee=1`, `multipleChoiceBinary.addAnswerCost=2`, `hardAnswerSafetyCap=50`, `initialAccountBalance=0`, `maximumDebtAllowed=500`.

Methodology: each case was exercised through the public HTTP API (`/v0/...`) where reachable, with results cross-checked directly against the Postgres database and the domain source code. Game mode is `moderator` with `marketApprovalRequired=true`, so grouped markets are created in `proposed` status and were promoted to `published` via the admin approve endpoint before trading/resolution.

## Summary

| # | Test Case | Result |
|---|-----------|--------|
| 1 | Group Creation — Happy Path | ✅ PASS |
| 2 | Group Creation — Minimum Answers | ✅ PASS |
| 3 | Group Creation — Maximum Answers | ✅ PASS |
| 4 | Group Creation — Duplicate Answer Labels | ✅ PASS |
| 5 | Independent Trading on Multiple Answers | ✅ PASS |
| 6 | Opposing Bets Within a Group | ✅ PASS (see note on share rounding) |
| 7 | Independent Probabilities — No Normalization | ✅ PASS |
| 8 | Exclusive YES Resolution — One Winner | ✅ PASS |
| 9 | Manual Resolution — Multiple Winners | ✅ PASS |
| 10 | Steward Work Income Calculation | ✅ PASS (formula reconciled) |
| 11 | Answer Addition — Steward Auto-Approval | ✅ PASS (minor: amendment also on new child) |
| 12 | Answer Addition — Pending Review | ✅ PASS |
| 13 | Answer Addition — Rejection | ✅ PASS |
| 14 | Answer Addition — Duplicate of Existing | ✅ PASS |
| 15 | Answer Addition — Duplicate of Pending | ✅ PASS |
| 16 | Answer Addition After Resolution DateTime | ✅ PASS |
| 17 | Resolution Blocked by Unpublished Child | ✅ PASS |
| 18 | Grouped Leaderboard Aggregation | ✅ PASS |
| 19 | Grouped Positions View | ✅ PASS |
| 20 | Grouped Bets Activity Tab | ✅ PASS |

**20 PASS, 0 discrepancies.** No crashes or data corruption observed. TC6 and TC11 include informational notes, but no longer require a caveat.

## Current Automated Follow-Up

Follow-up checks run on 2026-06-16 against `feature/multiple-choice-binary-markets` confirm the implementation is passing with the reconciled work-profit behavior and the TC17 unpublished-child resolution error fix.

| Check | Command | Result |
|-------|---------|--------|
| Multiple-choice binary API scenario runner | `node integrationtest/scripts/multiple-choice-binary-api.mjs --base-url http://localhost:8080 --api-prefix /v0` | ✅ 17/17 checks passed |
| Read-only OpenAPI contract smoke | `MAX_EXAMPLES=1 integrationtest/scripts/schemathesis-read.sh` | ✅ 4/4 operations passed |
| Grouped resolution regression tests | `go test ./internal/domain/markets -run 'TestResolveMarketGroup' && go test ./handlers/markets -run 'TestResolveMarketGroupHandler'` | ✅ Passed |
| Backend test suite | `JWT_SIGNING_KEY=test-secret-key-for-testing go test ./...` | ✅ Passed |
| Frontend build report | `npm run build:report` | ✅ Passed with existing Browserslist/chunk-size warnings |

The API scenario runner currently covers seeded login, duplicate answer rejection, grouped market creation, child market invariants, independent trading, non-normalized probabilities, grouped bets/positions/leaderboard reads, steward answer addition, exclusive resolution, parent resolution state, and net grouped work-profit financial reporting.

Latest ignored machine-readable output:

```text
integrationtest/artifacts/multiple-choice-binary-latest.json
```

---

## Detailed Findings

### TC1 — Group Creation, Happy Path ✅
Created "Who wins Euro 2028?" with `[Spain, France, Germany, Brazil]` (HTTP 201). Verified:
- One `MarketGroup`, `groupType = MULTIPLE_CHOICE_BINARY`, `probabilityPolicy = INDEPENDENT_BINARY`, `resolutionPolicy = INDEPENDENT_CHILDREN`.
- Four child binary markets titled `"Who wins Euro 2028? - {Answer}"`, each `yesLabel=YES` / `noLabel=NO`, same `resolutionDateTime`.
- Child `proposal_cost = 0` for all four (DB confirmed); group `proposal_cost = 10`.
- Creator (`testuser01`) charged once: account balance `0 → -10`.
- All children `lifecycle_status = proposed`.

### TC2 — Minimum Answers ✅
2 answers → 201. 1 answer → 400. Minimum of 2 enforced (`MinMarketGroupAnswers = 2`).

### TC3 — Maximum Answers ✅
50 answers → 201. 51 answers → 400 (`hardAnswerSafetyCap = 50`).

> Side note: the handler DTO `CreateMarketGroupRequest.AnswerLabels` carries a `validate:"max=20"` struct tag, but `sanitizeMarketGroupRequest` does **not** run the validator — only string sanitization. The effective limit is the domain cap (50), so behavior is correct, but the `max=20` tag is dead/misleading and should be corrected to 50 (or removed) to avoid confusion.

### TC4 — Duplicate Answer Labels ✅
`["Red","Blue","red"]` → 400. Case-insensitive collision caught (`strings.ToLower` in `ValidateMarketGroupMembers`).

### TC5 — Independent Trading ✅
`testuser02` bet 50 YES on Spain and 30 YES on France.
- Two separate positions on two separate child markets (DB + positions view).
- Balance `0 → -82` = 80 staked + 2 × `initialBetFee` (1 per first-time market participation).
- Spain probability `0.50 → 0.917`, France `0.50 → 0.875`, Germany unchanged `0.50` — each child moves independently.
- Grouped positions view aggregates `yesSharesOwned = 80` with per-answer rows for Spain and France.

### TC6 — Opposing Bets ✅
A user holding YES on one answer and NO on another is accepted with no cross-market constraint, and the leaderboard badge is `MIXED`. Demonstrated with `testuser05` (YES Brazil 30 + NO Germany 15 → badge **MIXED**, both child probabilities moved in their respective directions).

> **Note (share rounding at extreme probability):** an earlier attempt with `testuser03` (20 YES on Spain when Spain was already ~92%) produced **0 shares** for the YES leg after the market's share redistribution, so that user's badge showed `NO` instead of `MIXED`. The `MIXED` aggregation logic is correct (`groupedLeaderboardPosition` checks `yesShares>0 && noShares>0`); the artifact is the underlying CPMM share math — a small stake against a near-certain outcome can round to zero shares. Worth being aware of when reasoning about badges, though not specific to grouped markets.

### TC7 — Independent Probabilities, No Normalization ✅
After heavy YES volume, Spain = 0.9375, France = 0.5833, Germany = 0.90. Sum of child YES probabilities = **2.42**, i.e. probabilities are **not** normalized to 1.0. Each child is a self-contained binary market.

### TC8 — Exclusive YES Resolution ✅
Resolved group 1 with `mode = exclusive_yes`, `winningMarketId = Spain`.
- Spain → YES; France, Germany, Brazil → NO; group `lifecycle_status = resolved`.
- Payouts (balances before → after):
  - `testuser02` (Spain YES winner + France YES loser): `-82 → -12` (net payout).
  - `testuser03` (France NO winner on a NO-resolved market): `-42 → +8`.
  - `testuser04` (Germany YES loser): `-41 → -41` (stake lost, no payout).
  - `testuser05` (Germany NO winner + Brazil YES loser): `-62 → +8`.
- Confirms: YES holders on winner paid; YES holders on losing answers lose stake; NO holders on losing answers paid.

### TC9 — Manual Resolution, Multiple Winners ✅
New 4-answer group resolved with `mode = manual`: Spain=YES, France=YES, Germany=NO, Brazil=NO. DB confirmed each child resolved to its specified outcome (2 YES, 2 NO) and group `lifecycle_status = resolved`.

### TC10 — Steward Work Income ✅ PASS
Group 1 had **4 unique participants** across its children. After resolution, steward `testuser01`'s balance moved `-30 → -26`, i.e. a `TransactionWorkProfit` of **+4 = 4 participants × initialBetFee(1)**.

Follow-up reconciliation on 2026-06-16 confirmed this is intended balance behavior. The resolution-time balance transaction pays gross work income:

```text
grossWorkIncome = uniqueParticipantsAcrossGroup * InitialBetFee
```

Financial reporting separately shows net work profit:

```text
netWorkProfit = grossWorkIncome - groupProposalCost
```

The proposal cost is charged up front at group creation, so subtracting it again from the resolution-time balance transaction would double-charge the steward/creator path. Unique-participant counting is correct: a user who trades multiple answers is counted once.

### TC11 — Answer Addition, Steward Auto-Approval ✅
Steward `testuser01` proposed "Yellow" on a published group → immediately `status = approved` (`reviewedBy = testuser01`). New child market created ("Favorite color? - Yellow", inheriting parent resolution datetime + description). Description amendments written to existing children. Proposer charged `addAnswerCost = 2` (`-46 → -48`).

> **Minor:** the amendment is also written to the **newly added** child market, not only the pre-existing ones (the new member is appended before the amendment loop runs). Harmless, but slightly more than "all existing child markets."

### TC12 — Answer Addition, Pending Review ✅
Non-steward moderator `testuser02` proposed "Orange" with auto-approve disabled → `status = pending`. No child market created (member count unchanged at 4). No charge applied (balance `-12 → -12`). Pending proposal is visible for review.

### TC13 — Answer Addition, Rejection ✅
Steward rejected the pending "Orange" with reason "not relevant" → `status = rejected`, `reviewedBy = testuser01`, `rejection_reason` stored. No child created, member count unchanged, proposer not charged.

### TC14 — Duplicate of Existing Answer ✅
Proposing "red" when "Red" exists → 400 (case-insensitive `answerLabelExists`).

### TC15 — Duplicate of Pending Proposal ✅
With "Orange" pending, proposing "orange" → 400 (case-insensitive `pendingAnswerLabelExists`).

### TC16 — Answer Addition After Resolution DateTime ✅
With the group's `resolution_date_time` set to the past, proposing a new answer → **409 INVALID_STATE**. (Past resolution can only occur via DB edit since creation requires a future datetime; the guard `!ResolutionDateTime.After(now)` fires correctly.)

### TC17 — Resolution Blocked by Unpublished Child ✅ PASS
A published 2-answer group with one child forced back to `proposed` → resolve returns **409 MARKET_GROUP_CHILD_UNPUBLISHED** and **neither** child is resolved (validation is all-or-nothing, before any mutation). Blocking works correctly.

Follow-up implementation on 2026-06-16 replaced the generic `MARKET_CLOSED` mapping with a specific `MARKET_GROUP_CHILD_UNPUBLISHED` failure. The response now identifies the blocking answer child:

```json
{
  "ok": false,
  "reason": "MARKET_GROUP_CHILD_UNPUBLISHED",
  "message": "market group child \"Away\" (market 102) is not published; current status is proposed",
  "details": {
    "marketId": 102,
    "answerLabel": "Away",
    "lifecycleStatus": "proposed"
  }
}
```

The domain error still unwraps to `ErrInvalidState` for internal compatibility, and validation remains all-or-nothing before any child resolution mutation.

### TC18 — Grouped Leaderboard Aggregation ✅
Leaderboard for group 1 (post-resolution):

```
rank 1 | testuser03 | badge NO    | profit  30 | curVal 50 | [France:30]
rank 2 | testuser05 | badge MIXED | profit  10 | curVal 70 | [Germany:40, Brazil:-30]
rank 3 | testuser02 | badge YES   | profit -10 | curVal 70 | [Spain:20, France:-30]
rank 4 | testuser04 | badge YES   | profit -40 | curVal  0 | [Germany:-40]
```

Confirms: total profit aggregated across children, per-answer profit breakdown present, badge reflects aggregate YES/NO exposure (`MIXED` for testuser05), and ranking is by total profit desc, then current value desc, then username. (The TESTCASE's `+20/-5/+10 = +25` figures are illustrative; the aggregation/ordering behavior matches.)

### TC19 — Grouped Positions View ✅
Positions for group 1 show aggregated `totYes`/`totNo` per user plus per-answer detail rows, sorted by total share count descending:

```
testuser02 | totYes 110 totNo  0 | [Spain(Y70/N0), France(Y40/N0)]
testuser04 | totYes  55 totNo  0 | [Germany(Y55/N0)]
testuser05 | totYes  30 totNo 15 | [Germany(Y0/N15), Brazil(Y30/N0)]
testuser03 | totYes   0 totNo 10 | [France(Y0/N10)]
```

### TC20 — Grouped Bets Activity Tab ✅
Bets endpoint returns all bets across all children, interleaved chronologically (newest first), not separated by answer. Each row carries username, answer label, outcome (YES/NO), amount, probability at time of bet, and timestamp:

```
total: 7
09:21:17 | testuser05 | Germany | NO  | amt 30 | prob 0.562
09:21:17 | testuser05 | Brazil  | YES | amt 30 | prob 0.875
09:19:33 | testuser04 | Germany | YES | amt 40 | prob 0.900
09:19:32 | testuser03 | France  | NO  | amt 20 | prob 0.583
09:19:32 | testuser03 | Spain   | YES | amt 20 | prob 0.938
09:19:12 | testuser02 | France  | YES | amt 30 | prob 0.875
09:19:12 | testuser02 | Spain   | YES | amt 50 | prob 0.917
```

---

## Recommendations

1. **TC3 — fix the dead DTO tag.** `CreateMarketGroupRequest.AnswerLabels` says `validate:"max=20"` but the enforced cap is 50 and the validator isn't even run; align it with `hardAnswerSafetyCap`.
2. **TC11 — confirm intended amendment scope.** Description amendments are written to the newly added child as well as the pre-existing ones; confirm this is desired.
3. **TC6 — document share-rounding behavior.** Small stakes against near-certain outcomes can round to 0 shares, which affects position/leaderboard badges. Not grouped-market-specific, but worth noting for QA.
