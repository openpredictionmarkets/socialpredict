---
title: Load Testing And Release Dossier Implementation Plan
document_type: feature-plan
domain: features
author: Patrick Delaney
updated_at: 2026-05-28T00:00:00Z
updated_at_display: "Thursday, May 28, 2026"
update_reason: "Add the implementation sequence for staging load testing, hot-market betting simulations, and release dossier evidence capture."
status: draft
---

# Load Testing And Release Dossier Implementation Plan

## Purpose

This plan turns [DESIGN.md](./DESIGN.md) and [04-load-testing.md](./04-load-testing.md) into an implementation sequence.

The plan is intentionally split into reviewable slices. Load testing touches staging safety, API behavior, database write paths, observability, and release evidence. The first implementation should produce evidence without changing production architecture.

Agents implementing this feature should mark checklist items as they complete them and leave unchecked items in place when intentionally deferred.

## Branching Rule

This feature plan should branch from `main` unless it explicitly depends on a still-open infrastructure or API PR.

Implementation PRs should stay small enough to review independently:

- docs and scenario design
- safe seed/reset support
- k6 smoke and baseline scripts
- hot-market burst scripts
- dossier builder/output schema
- optional workflow/manual dispatch

The first implementation should keep the CLI and k6 scripts in this repository under top-level `loadtest/`. This is an early-development convenience and contract-cohesion choice, not a permanent ownership claim. Do not create a separate repository unless the harness later needs independent versioning or reuse outside SocialPredict.

## Planning Principles

- Treat load testing as release evidence, not as a product feature visible to normal users.
- Keep seed/reset behavior staging-only and guarded by explicit configuration.
- Use normal API/auth/betting behavior for traffic generation.
- Do not introduce a god key for betting.
- Measure before optimizing.
- Keep raw results separate from summarized release dossier artifacts.
- Record deployment topology with every meaningful result.
- Prefer k6 for the first API-load implementation before inventing custom tooling.
- Keep HostOps as optional host observation support, not the owner of load scenarios.
- Keep DigitalOcean authentication separate from SocialPredict fake-user API authentication.

## Progress Ledger

- [x] 01. Feature artifact and design alignment
- [ ] 02. Staging safety and environment gates
- [ ] 03. Seed/reset fixture command
- [x] 04. Credential/token fixture export scaffold
- [x] 05. k6 smoke scenario scaffold
- [x] 06. k6 baseline browse and bet scenario scaffold
- [x] 07. k6 hot-market burst scenario scaffold
- [x] 08. Soak scenario and failure classification scaffold
- [x] 09. Dossier artifact schema and summarizer scaffold
- [ ] 10. Operator runbook and first staging evidence capture

## Implementation Checklist

### 01. Feature Artifact And Design Alignment

Status: complete for the feature docs and initial loadtest harness scaffold.

Checklist:

- [x] Create `README/PRODUCTION-NOTES/FEATURES/04/`.
- [x] Add feature overview document.
- [x] Add `DESIGN.md` aligned with canonical runtime, API, betting, database, observability, and release boundaries.
- [x] Add `PLAN.md` as an agent-usable implementation sequence.
- [x] Update production-note index links.
- [x] Keep the first PR documentation-first, then add an initial portable harness scaffold in the same branch.

Exit criteria:

- [x] Documentation explains why betting traffic must use normal auth and `/v0/bet`.
- [x] Documentation records staging-only seed/reset safety requirements.
- [x] Documentation defines release dossier evidence expectations.

Validation:

- [x] `git diff --check`

### 02. Staging Safety And Environment Gates

Service ownership: Runtime Bootstrap and Infrastructure plus Release and Deployment Control.

Status: planned.

Checklist:

- [ ] Add a typed runtime flag such as `LOAD_TEST_ENABLED` or equivalent.
- [ ] Ensure the flag defaults to disabled in every environment.
- [ ] Decide whether the flag lives only in backend runtime config or is also surfaced in `./SocialPredict` help.
- [ ] Ensure production install/deploy documentation does not enable load-test mutation paths.
- [ ] Add tests that seed/reset commands refuse to run unless load testing is explicitly enabled.

Exit criteria:

- [ ] Accidental production fixture creation is blocked by default.
- [ ] Operators can see how to enable load testing on staging or a dedicated load-test droplet.

Validation:

- [ ] Backend config tests.
- [ ] Manual command refusal check with flag absent.

### 03. Seed/Reset Fixture Command

Service ownership: Participant Account Context, Prediction Market Context, Runtime Bootstrap.

Status: planned.

Checklist:

- [ ] Decide command shape: `./SocialPredict load seed`, backend `cmd/loadseed`, or both.
- [ ] Create fictional users with deterministic prefix and `must_change_password=false`.
- [ ] Create fictional moderators when moderator mode is enabled.
- [ ] Create configurable market count.
- [ ] Mark or export selected hot-market IDs.
- [ ] Provide a reset/wipe option scoped only to load-test fixtures.
- [ ] Prevent deletion of non-load-test data.

Exit criteria:

- [x] A staging operator can seed 50,000 users, 100 moderators, and 1,000 markets repeatably.
- [x] Fixture data is identifiable and removable.

Validation:

- [ ] Local development seed smoke test.
- [ ] Staging seed dry run or limited-count test.

### 04. Credential/Token Fixture Export

Service ownership: API and Auth Contract Boundary plus Participant Account Context.

Status: seeded fixture generation is implemented; token export remains intentionally deferred.

Checklist:

- [x] Decide whether k6 logs in during setup or consumes exported credentials/tokens for the first scaffold.
- [x] Keep the first runner CLI under `loadtest/cli` in this repository.
- [x] Keep k6 scenarios under `loadtest/k6`.
- [x] Keep release dossier schema and curated summaries under `loadtest/dossier`.
- [x] Keep HostOps/DigitalOcean/Linux observation captures under `loadtest/hostops`.
- [x] Document that k6 needs only SocialPredict fake-user credentials for API traffic.
- [x] Document that DigitalOcean credentials are only for host observation or provisioning a separate load-generator droplet.
- [x] If exporting credentials, keep files out of git by default.
- [ ] If exporting tokens later, ensure they are short-lived or staging-only.
- [x] Ensure fictional passwords are not reused for real users in documented examples.
- [x] Document secure local paths for generated fixtures.

Exit criteria:

- [x] k6 can authenticate user pools without requiring a god key.
- [x] Generated fixture files are clearly ephemeral.

Validation:

- [ ] Smoke script logs in and places a small number of bets as normal users.

### 05. k6 Smoke Scenario

Service ownership: Release and Deployment Control plus API and Auth Contract Boundary.

Status: scaffolded.

Checklist:

- [x] Add `loadtest/k6/smoke.js`.
- [x] Run k6 from outside the target app/database droplet by default.
- [x] Target the public staging URL so traffic crosses DNS, TLS, reverse proxy, backend, and database.
- [x] Check `/health`, `/readyz`, and `/ops/status`.
- [x] Log in one or more fictional users.
- [x] Read market list and market details.
- [x] Place a small number of bets on a known market.
- [x] Emit JSON summary output.

Exit criteria:

- [x] Smoke test scaffold proves environment and credentials can be wired before heavy load.

Validation:

- [ ] Local or staging smoke run with low traffic.

### 06. k6 Baseline Browse And Bet Scenarios

Service ownership: API and Auth Contract Boundary, Prediction Market Context, Betting and Position Ledger Context.

Status: scaffolded.

Checklist:

- [x] Add browse scenario for market list and detail reads.
- [x] Add moderate bet scenario for normal write load.
- [x] Parameterize base URL, users file, market IDs, duration, and request rate.
- [x] Record p50/p95/p99, throughput, and error classifications.
- [ ] Keep thresholds advisory until enough evidence exists.

Exit criteria:

- [ ] Operators can run a repeatable baseline against staging.

Validation:

- [ ] Staging baseline run recorded in a draft dossier.

### 07. k6 Hot-Market Burst Scenario

Service ownership: Betting and Position Ledger Context and Database Runtime.

Status: scaffolded.

Checklist:

- [x] Add `hot-market-burst.js` using request-rate based execution.
- [x] Concentrate writes on hot markets.
- [x] Support configurable target rate and duration.
- [x] Alternate YES/NO outcomes.
- [x] Track bet success and failed bet counts.
- [x] Capture setup-time `/readyz` and `/ops/status` snapshots.

Exit criteria:

- [ ] The test can expose hot-market write contention without bypassing normal betting behavior.

Validation:

- [ ] Staging hot-market run at a small rate.
- [ ] Higher-rate run only after confirming reset and monitoring procedure.

### 08. Soak Scenario And Failure Classification

Service ownership: Observability Boundary plus Release and Deployment Control.

Status: scaffolded.

Checklist:

- [x] Add a longer steady-load scenario.
- [ ] Classify failures by HTTP status and application reason where possible.
- [x] Track dropped iterations and client-side timeout categories through k6 summary output.
- [x] Document host observation capture location for memory/disk state.

Exit criteria:

- [ ] Operators can distinguish transient overload from slow resource leaks.

Validation:

- [ ] Short soak run in staging.

### 09. Dossier Artifact Schema And Summarizer

Service ownership: Documentation and Release Evidence.

Status: scaffolded.

Checklist:

- [x] Add `loadtest/dossier/schema.example.json`.
- [x] Add a summarizer that converts k6 JSON output plus operator-supplied topology metadata into a dossier JSON file.
- [x] Include release, commit SHA, base URL, scenario, seed counts, traffic shape, result metrics, infra observations, decision, and known risks.
- [x] Keep raw result files ignored unless intentionally retained.
- [x] Document how the dossier can be used with the release dossier dashboard.

Exit criteria:

- [ ] A release candidate can carry a compact load-test evidence artifact.

Validation:

- [x] Generate a sample dossier from a synthetic k6 summary.
- [ ] Generate a sample dossier from a smoke run.

### 10. Operator Runbook And First Staging Evidence Capture

Service ownership: Release and Deployment Control plus Host Operations.

Status: planned.

Checklist:

- [ ] Document how to run from macOS.
- [ ] Document when to use a separate load-generator droplet.
- [ ] Document that same-droplet k6 runs are smoke checks only, not capacity evidence.
- [ ] Document DigitalOcean metrics to capture.
- [ ] Document basic host checks such as `docker stats`, `df -h`, and safe Postgres observation commands.
- [ ] Record first staging results and capacity notes.
- [ ] Decide next architecture action from evidence.

Exit criteria:

- [ ] A non-author can reproduce the first load test and understand the dossier.

Validation:

- [ ] Staging runbook followed once end-to-end.

## Review Checklist

Before implementation PRs merge, reviewers should confirm:

- [ ] Load-test mutation commands are staging-only by default.
- [ ] Betting traffic uses normal auth and normal `/v0/bet` behavior.
- [ ] Fixture files and raw results are not accidentally committed.
- [ ] OpenAPI is used as the API source of truth when scripts depend on routes or payloads.
- [ ] Failure categories are not collapsed into a single opaque error count.
- [ ] The dossier records topology, because capacity numbers without topology are misleading.
- [ ] No architecture optimization is bundled without evidence from a prior test run.
