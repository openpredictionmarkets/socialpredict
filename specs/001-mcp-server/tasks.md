# Tasks: MCP Server

**Input**: Design documents from `/specs/001-mcp-server/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/mcp-tools.md, quickstart.md

**Tests**: INCLUDED — constitution Principle III makes tests accompanying every
behavioral change mandatory (hand-written fakes, no mocking frameworks).

**Organization**: Tasks grouped by user story. US1 = read-only market data tools
(MVP). US2 = environment parity (dev / localhost / production VPS).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: US1 or US2 (from spec.md)
- Build/test command: `cd backend && GOFLAGS="" /usr/local/go/bin/go test ./...`
  (known pre-existing failure: `internal/domain/balance_integration_test.go` — unrelated)

## Path Conventions

Web app, backend-only feature: all code under `backend/` per plan.md structure.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Dependency + package skeleton

- [X] T001 Add `github.com/modelcontextprotocol/go-sdk` to `backend/go.mod` (`go get`), verify `go build ./...` passes
- [X] T002 Create package skeleton `backend/handlers/mcp/` with `server.go` containing package declaration and doc comment describing the read-only MCP transport adapter

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Adapter boundary + endpoint mount — required before any tool work

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T003 Define narrow `MarketReader` interface (ListMarkets, GetMarketDetails, SearchMarkets, GetMarketBets — exactly the 4 methods from data-model.md) in `backend/handlers/mcp/interfaces.go`, with compile-time assertion that the existing `dmarkets` service satisfies it
- [ ] T004 [P] Write hand-written fake `MarketReader` (seedable markets/bets, error injection; no mocking framework) in `backend/handlers/mcp/fake_reader_test.go`
- [ ] T005 Implement MCP server constructor in `backend/handlers/mcp/server.go`: official SDK server (serverInfo name `socialpredict`), Streamable HTTP handler in stateless mode, tools capability only, returns `http.Handler`; no tools registered yet (per research.md R1/R2)
- [ ] T006 Mount `POST /v0/mcp` in `backend/server/server.go` wrapped in existing `securityMiddleware`, wiring the markets service from the DI container (`internal/app/container.go` — extend wiring only if the service is not already exposed)

**Checkpoint**: `/v0/mcp` answers `initialize` and returns an empty `tools/list` — foundation ready

---

## Phase 3: User Story 1 - AI Assistant Reads Market Data (Priority: P1) 🎯 MVP

**Goal**: Four read-only MCP tools serving public market data identical to the website

**Independent Test**: Connect a standard MCP client (or curl JSON-RPC) to a running instance, list markets, fetch details — data matches the website (spec US1 acceptance scenarios 1–4)

### Tests for User Story 1 (write FIRST, must fail before implementation)

- [ ] T007 [P] [US1] Contract test in `backend/handlers/mcp/server_contract_test.go`: drive real handler via `httptest` through initialize → tools/list → tools/call lifecycle per `contracts/mcp-tools.md`, including unknown-tool and unknown-market negative cases
- [ ] T008 [P] [US1] Catalog guard test in `backend/handlers/mcp/catalog_test.go`: `tools/list` returns exactly {`list_markets`, `get_market`, `search_markets`, `list_market_bets`} — nothing else (FR-003/SC-003 structural enforcement)
- [ ] T009 [P] [US1] Unit tests in `backend/handlers/mcp/tools_markets_test.go` against fake reader: per-tool happy paths, status filter, limit clamping [1,100], empty instance → `[]`, invalid id → "market not found", unknown arguments rejected

### Implementation for User Story 1

- [ ] T010 [P] [US1] Define output DTOs (`MarketSummary`, `MarketDetail`, `BetInfo`, `SearchResult`) with mappers from domain structs in `backend/handlers/mcp/tools_markets.go` — public fields only, moderation/internal fields excluded per data-model.md
- [ ] T011 [US1] Implement `list_markets` tool (input schema w/ status enum + limit/offset, `additionalProperties: false`) in `backend/handlers/mcp/tools_markets.go`
- [ ] T012 [US1] Implement `get_market` tool (id ≥ 1 validation, not-found tool error, MarketOverview → MarketDetail) in `backend/handlers/mcp/tools_markets.go`
- [ ] T013 [US1] Implement `search_markets` tool (required non-empty query, primary + fallback results) in `backend/handlers/mcp/tools_markets.go`
- [ ] T014 [US1] Implement `list_market_bets` tool (id validation, BetDisplayInfo → BetInfo) in `backend/handlers/mcp/tools_markets.go`
- [ ] T015 [US1] Register all four tools in the server constructor in `backend/handlers/mcp/server.go`; confirm T007–T009 now pass
- [ ] T016 [US1] Run full backend suite `cd backend && GOFLAGS="" /usr/local/go/bin/go test ./...`; everything green except documented pre-existing failure

**Checkpoint**: US1 fully functional — quickstart.md §1–2 pass on dev

---

## Phase 4: User Story 2 - Operator Enables MCP on Any Deployment (Priority: P2)

**Goal**: Same endpoint works in dev, localhost, and production VPS with zero extra operator steps

**Independent Test**: Run documented deployment per environment; same client config (address only changes) reaches all MCP capabilities (spec US2 acceptance scenarios 1–3)

### Implementation for User Story 2

- [ ] T017 [P] [US2] Document `/v0/mcp` endpoint in embedded OpenAPI spec (`backend/openapi.yaml` or wherever `openapi_embed.go` sources it): POST, JSON-RPC 2.0/MCP content type, pointer to contract doc; verify `openapi_test.go` passes (constitution V)
- [ ] T018 [P] [US2] Verify dev environment per quickstart.md §2: `./SocialPredict up`, curl `http://localhost:${BACKEND_PORT}/v0/mcp` (direct) and `http://localhost/api/v0/mcp` (nginx) — record results in quickstart or PR notes
- [ ] T019 [US2] Verify localhost deployment per quickstart.md §3 (`scripts/docker-compose-local.yaml` path): `http://localhost/api/v0/mcp` returns identical tool catalog
- [ ] T020 [US2] Confirm production path requires no changes: audit `data/nginx/vhosts/prod/` (`/api/` proxy already covers `/v0/mcp`) and `scripts/docker-compose-prod.yaml`; document VPS verification steps (quickstart.md §4) — no compose/vhost edits expected
- [ ] T021 [P] [US2] Write operator/user docs: MCP connection guide in `backend/docs/` (or repo `README.md` section) covering client setup (`claude mcp add`, MCP Inspector), all three environment URLs, and read-only scope — satisfies SC-001 five-minute path

**Checkpoint**: All three environments verified (VPS steps documented); both stories independently functional

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: End-to-end validation against spec success criteria

- [ ] T022 [P] Real-client smoke test per quickstart.md §5: `claude mcp add --transport http socialpredict <dev-url>` (or MCP Inspector), ask for open markets, verify live data returned (SC-001)
- [ ] T023 Negative + parity checks per quickstart.md §6–7: unknown tool, unknown market id, zero state changes; `get_market` output matches website values for same market (SC-003/SC-004/SC-005)
- [ ] T024 Final pass: rate limiting confirmed on `/v0/mcp` (securityMiddleware active — FR-006), full test suite green, quickstart.md checklist complete, spec checklists updated in `specs/001-mcp-server/checklists/requirements.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: none — start immediately (T001 → T002 sequential: skeleton needs module state)
- **Foundational (Phase 2)**: needs Phase 1. T003 → T005 → T006; T004 parallel with T005/T006. BLOCKS all stories
- **US1 (Phase 3)**: needs Phase 2. Tests T007–T009 first (parallel), then T010 → T011–T014 (same file, sequential) → T015 → T016
- **US2 (Phase 4)**: needs Phase 2 only for T017; T018–T020 need US1 complete to verify real tool catalog (verification tasks, not code). T017 ∥ T021 anytime after Phase 2
- **Polish (Phase 5)**: needs US1 + US2

### User Story Dependencies

- **US1 (P1)**: independent after Foundational
- **US2 (P2)**: code-independent (T017/T021 parallel to US1); environment verification (T018–T020) meaningfully runs after US1 delivers tools

### Parallel Opportunities

```text
Phase 2:  T004 ∥ (T005 → T006)
Phase 3:  T007 ∥ T008 ∥ T009 (test files), then T010 before T011–T014
Phase 4:  T017 ∥ T021 ∥ (T018 → T019 → T020)
Cross-story: T017/T021 (US2 docs) ∥ T011–T016 (US1 tools) — different files
```

## Parallel Example: User Story 1

```bash
# Launch all US1 test-authoring tasks together:
Task: "Contract test in backend/handlers/mcp/server_contract_test.go"
Task: "Catalog guard test in backend/handlers/mcp/catalog_test.go"
Task: "Unit tests in backend/handlers/mcp/tools_markets_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Phase 1: Setup (T001–T002)
2. Phase 2: Foundational (T003–T006) — CRITICAL gate
3. Phase 3: US1 (T007–T016), tests first
4. **STOP and VALIDATE**: quickstart.md §1–2 on dev — working MCP endpoint with live market data
5. Demo-able MVP

### Incremental Delivery

1. Setup + Foundational → `/v0/mcp` answers handshake
2. US1 → four tools live → validate on dev → MVP
3. US2 → OpenAPI + docs + three-environment verification → production-ready
4. Polish → real-client smoke + negative/parity checks → done

---

## Notes

- T011–T014 share `tools_markets.go` — sequential, no [P]
- Constitution IV guard: no task adds any write path; T008 enforces structurally
- Commit after each task or logical group (branch `feature/mcp-server`)
- Verify T007–T009 FAIL before starting T010
