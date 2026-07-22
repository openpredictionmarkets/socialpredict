<!--
Sync Impact Report
==================
Version change: (template, unversioned) → 1.0.0
Rationale: Initial adoption — all template placeholders replaced with concrete
project principles derived from repository context (README, CONTRIBUTING, backend
architecture, CI workflows).

Modified principles: n/a (initial adoption; 5 principles defined)
Added sections:
  - Core Principles (I–V)
  - Technology & Platform Constraints
  - Development Workflow & Quality Gates
  - Governance
Removed sections: none (template slots all filled)

Templates requiring updates:
  - .specify/templates/plan-template.md — ✅ aligned (generic Constitution Check
    gates are derived from this file at plan time; no edits needed)
  - .specify/templates/spec-template.md — ✅ aligned (no constitution-specific
    references; mandatory sections unchanged)
  - .specify/templates/tasks-template.md — ✅ aligned (task categories already
    cover tests/integration work implied by Principles III–IV)
  - .claude/skills/speckit-*/SKILL.md — ✅ aligned (all reference the constitution
    generically by path; no agent-specific or outdated references found)
  - README.md / CONTRIBUTING.md — ✅ no changes required (workflow rules in
    Development Workflow section mirror CONTRIBUTING.md, not the reverse)

Follow-up TODOs: none
-->

# SocialPredict Constitution

## Core Principles

### I. Domain-Driven Backend Architecture

Business logic MUST live in domain packages under `backend/internal/domain/*`
(users, bets, markets, auth, analytics, wallet). Database models (`models.*`,
GORM) MUST remain separate from domain models; handlers and services MUST NOT
leak persistence types across domain boundaries. Dependencies are wired through
the DI container (`internal/app/container.go` → `BuildApplication()`); new
services MUST be registered there rather than constructed ad hoc.

*Rationale: The codebase is mid-migration from handler-centric to domain-driven
design. Every change that bypasses domain boundaries increases the migration
debt and couples business rules to storage details.*

### II. Narrow, Consumer-Defined Interfaces

Services MUST depend on the smallest interface that satisfies their need,
defined on the consumer side (e.g., `AuthService` depending on `UserReader` and
`CredentialRepository` rather than the full `users.Repository`). HTTP handlers
MUST be functions returning `http.HandlerFunc` that close over service
interfaces. Adding methods to a broad shared interface when a narrow local one
suffices is a constitution violation and MUST be justified in the plan's
Complexity Tracking section.

*Rationale: Interface bloat was the root cause of the auth-extraction effort;
narrow interfaces keep domains independently testable and replaceable.*

### III. Tests Accompany Every Change

Every behavioral change MUST ship with tests in the same PR. Test doubles are
hand-written fakes implementing the consumed interfaces directly — mocking
frameworks MUST NOT be introduced. Integration tests (database startup, seed,
cross-domain flows) are REQUIRED for changes touching persistence, migrations,
or market lifecycle. The backend CI workflow MUST pass before merge.

*Rationale: The project runs real-money-style accounting logic; untested
changes to it are indistinguishable from bugs.*

### IV. Financial Integrity (NON-NEGOTIABLE)

Market math, wallet balances, bet placement, and payout distribution MUST be
deterministic and conservation-checked: no operation may create or destroy
value except explicitly defined credits (e.g., account seeding). Changes to
betting, pricing, or resolution math REQUIRE tests demonstrating balance
conservation across the affected flows. Monetary amounts MUST use the existing
integer-points representation — floating-point money is forbidden.

*Rationale: A prediction market engine is an accounting system first; silent
value leaks destroy user trust and cannot be repaired retroactively.*

### V. Secure, Self-Hostable by Default

SocialPredict MUST remain deployable by a non-expert via the documented Docker
Compose path. Secrets MUST NOT be committed; configuration flows through env
files and the setup tooling. Authentication and authorization changes MUST go
through the `Authenticator` interface rather than package-level helpers.
Container images MUST pass the Trivy security scan in CI. Public API surface
changes MUST be reflected in the embedded OpenAPI spec.

*Rationale: The project's mission is open, self-hosted prediction-market
infrastructure — anything that only works in a bespoke environment or weakens
the security posture breaks the mission.*

## Technology & Platform Constraints

- Backend: Go (module in `backend/`), GORM ORM, PostgreSQL. Build and test with
  the system Go toolchain (`/usr/local/go/bin/go`) and clean `GOFLAGS`.
- Frontend: React (in `frontend/`), consuming the backend HTTP API.
- Deployment: Docker / Docker Compose (`docker/`, `deploy/`, `scripts/`);
  staging deploys on PR, production deploys via workflow run.
- API contract: OpenAPI spec embedded in the backend (`openapi_embed.go`) and
  served via Swagger UI; it is the single source of truth for HTTP contracts.
- License: MIT. Dependencies MUST be MIT-compatible.

## Development Workflow & Quality Gates

- Branch names MUST start with `feature/`, `fix/`, `refactor/`, or `doc/`
  followed by a short description (per CONTRIBUTING.md).
- Commits MUST be atomic, each addressing a single concern, with descriptive
  messages.
- All changes reach `main` via pull request; PR descriptions MUST state what
  changed and why.
- Merge gates: backend test workflow green, Trivy container scan green, and
  review approval. Constitution compliance is checked during planning
  (Constitution Check section of plan.md) and re-checked after design.
- Known pre-existing failures (e.g., quarantined tests) MUST be documented,
  not silently worked around.

## Governance

This constitution supersedes other written practices where they conflict.
Amendments are made by PR that edits this file, states the semantic version
bump and its rationale, and updates dependent templates in `.specify/templates/`
in the same change.

Versioning policy: MAJOR for removals or redefinitions of principles, MINOR
for new principles or materially expanded guidance, PATCH for clarifications
and wording fixes.

Compliance review: every feature plan MUST evaluate its design against the
Core Principles in its Constitution Check section; violations MUST be either
removed or explicitly justified in Complexity Tracking. Reviewers MUST treat
unjustified violations as blocking.

**Version**: 1.0.0 | **Ratified**: 2026-07-22 | **Last Amended**: 2026-07-22
