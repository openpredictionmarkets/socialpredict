---
title: Configuration Management
document_type: production-notes
domain: backend
author: Patrick Delaney
updated_at: 2026-04-25T13:30:00-05:00
updated_at_display: "Saturday, April 25, 2026 at 1:30 PM Central (CDT)"
update_reason: "Align production notes with actual progress and evolving architecture."
status: active
---

# Configuration Management

## Update Summary

This note was updated on Saturday, April 25, 2026 to replace an older greenfield-style configuration plan with guidance that reflects the current SocialPredict codebase and the architecture direction needed for high availability and fault tolerance.

| Topic | Prior to April 25, 2026 | After April 25, 2026 |
| --- | --- | --- |
| Core framing | Treated configuration as one generic subsystem | Splits configuration into runtime/bootstrap detail and application-policy input |
| Main proposal | Create a new centralized `config/` package with environment hierarchies and hot reload | Finish the existing runtime and config-service boundaries already present in the backend |
| Operational model | Considered hot reload as a production feature | Prefer validate once at startup, inject immutable snapshots, and roll forward via controlled restart/deployment |
| Source of truth | Implied a new file hierarchy would replace current behavior | Keep `backend/setup/setup.yaml` as the source asset for application policy during this migration |
| Current-state assumptions | Assumed `main.go` still used old env-loading patterns | Reflects current `internal/app/runtime`, `internal/service/config`, and explicit container wiring |
| HA / fault tolerance | Optimized for flexibility first | Optimizes for deterministic startup, replica consistency, and rollback safety |

## Executive Direction

SocialPredict should treat configuration as two concerns:

1. Runtime/bootstrap detail
2. Application-policy input

Runtime/bootstrap config should be loaded and validated at startup and injected immutably. Economics and frontend policy should be served through an owned internal configuration boundary with no globals and no direct `setup` leakage into use-case code.

For high-availability operation, the backend should:

- validate once at startup
- inject immutable snapshots
- roll forward with controlled restarts and deployments

Hot-reloading economics or accounting-relevant policy is not a goal. If economics can change while the system is live in an uncontrolled way, the platform can eliminate financial and accounting correctness. Financial correctness is binary. The objective is 100% financial and accounting correctness.

## Why This Matters

The older version of this note was useful as a generic checklist, but it no longer matched the codebase or the architecture direction. SocialPredict is already moving toward explicit runtime seams and a service-backed configuration boundary. The production notes need to document and accelerate that transition, not restart the discussion from a hypothetical greenfield configuration subsystem.

This matters even more under the broader production objective of high availability and fault tolerance:

- replicas must start from the same validated configuration snapshot
- rollouts must be deterministic and reversible
- economics and accounting behavior must not drift between instances
- business-policy inputs must remain stable enough to preserve market and ledger correctness

## Current Code Snapshot

As of 2026-04-25, the backend is in a transitional but meaningful intermediate state.

### Runtime/bootstrap concerns already have explicit seams

`backend/main.go` no longer follows the older `util.GetEnv()` pattern. Startup now uses explicit runtime helpers:

```go
dbCfg, err := appruntime.LoadDBConfigFromEnv()
db, err := appruntime.InitDB(dbCfg, appruntime.PostgresFactory{})
configService, err := appruntime.LoadConfigService()
server.Start(openAPISpec, swaggerUIFS, db, configService)
```

`backend/internal/app/runtime/db.go` already owns normalized DB bootstrap concerns:

```go
type DBConfig struct {
    Host     string
    Port     string
    User     string
    Password string
    Name     string
    SSLMode  string
    TimeZone string
}
```

This is the right category for environment-driven infrastructure configuration.

### A configuration service already exists

The backend already has an internal config service:

```go
type Service interface {
    Current() *AppConfig
    Economics() Economics
    Frontend() Frontend
    ChartSigFigs() int
}
```

The service is already wired through the application container and server route registration:

```go
func NewContainer(db *gorm.DB, configService configsvc.Service) *Container
func BuildApplicationWithConfigService(db *gorm.DB, configService configsvc.Service) *Container
```

Setup handlers also already consume the service explicitly:

```go
func GetSetupHandler(configService configsvc.Service) func(w http.ResponseWriter, r *http.Request)
func GetFrontendSetupHandler(configService configsvc.Service) func(w http.ResponseWriter, r *http.Request)
```

### `backend/setup` is still a transitional seam

`backend/setup/setup.go` is currently the source asset boundary for economics and frontend policy, but it still mixes two responsibilities:

1. policy-shaped configuration data
2. package-global loading and access behavior

Today it still contains:

- embedded `setup.yaml`
- package-global `economicConfig`
- `sync.Once`
- `init()` loading
- package-level convenience accessors

That means `setup` is not yet the final owned boundary. It is a source asset plus a legacy access path.

### The migration is incomplete

The current config service is still partly a thin wrapper around `setup`:

```go
type AppConfig = setup.EconomicConfig
type Economics = setup.Economics
type Frontend = setup.Frontend
```

And runtime loading still calls:

```go
return configsvc.NewService(configsvc.LoaderFunc(setup.LoadEconomicsConfig))
```

Some consumers are on the right path, but others still accept or retain raw setup-shaped configuration. For example:

- `backend/internal/domain/bets/service.go` still stores `*setup.EconomicConfig`
- `backend/internal/domain/analytics/service.go` still uses `setup.EconConfigLoader`
- `backend/handlers/stats/statshandler.go` still reaches through `configService.Current()` and reads the full tree
- `backend/internal/service/config/service_test.go` still builds test values directly from `setup` types

So the live state is neither the old monolithic global approach nor the final target state. It is an in-progress migration.

## Configuration Taxonomy

### 1. Runtime/Bootstrap Configuration

This category covers process and infrastructure detail, for example:

- DB host/user/password/name/port
- SSL mode and timezone for DB connectivity
- CORS environment flags
- bind/runtime environment selection
- deployment-time operational toggles

This configuration belongs in startup, infrastructure wiring, and the composition root.

Properties:

- loaded from env or runtime deployment sources
- validated at startup
- immutable for the running process
- not exposed as a general-purpose business-policy object

### 2. Application-Policy Configuration

This category covers policy inputs that influence product behavior, for example:

- market creation defaults
- subsidies and incentives
- user starting balances and debt limits
- betting fee rules
- frontend chart formatting policy where it reflects product policy

This configuration belongs behind an owned internal boundary and should be consumed through narrow interfaces or snapshots, not through package globals.

Properties:

- source asset currently remains `backend/setup/setup.yaml`
- loaded once into validated, explicit application-policy structures
- injected into the parts of the system that need it
- not mutated ad hoc during live operation

## Economics Freeze and Financial Correctness

Economics configuration is not merely cosmetic. It affects the correctness of market behavior, fee rules, balances, debt, and other accounting-relevant flows.

SocialPredict should therefore treat economic policy as frozen for the scope whose accounting it governs.

That means:

- running processes should not hot-reload economics policy
- deployments should roll forward with versioned, validated snapshots
- accounting-relevant values should be copied into durable domain state at the right workflow boundaries when required
- the design must preserve 100% financial and accounting correctness

This note intentionally rejects configuration hot reload as a default strategy for economics or accounting-sensitive behavior.

## Current Tree Versus Target Tree

### Current configuration-related tree

```text
backend/
├── main.go
├── setup/
│   ├── setup.go
│   └── setup.yaml
├── internal/
│   ├── app/
│   │   ├── container.go
│   │   └── runtime/
│   │       ├── config.go
│   │       └── db.go
│   └── service/
│       └── config/
│           ├── service.go
│           └── service_test.go
├── handlers/
│   ├── setup/
│   │   └── setuphandler.go
│   └── stats/
│       └── statshandler.go
└── server/
    └── server.go
```

### End-state objective

The objective is not to create a giant generic `config/` package. The objective is to finish the boundary split and move consumers to owned, narrow configuration interfaces.

```text
backend/
├── main.go
├── setup/
│   └── setup.yaml                    # source asset during migration
├── internal/
│   ├── app/
│   │   ├── container.go
│   │   └── runtime/
│   │       ├── config.go            # startup loading/validation of runtime/bootstrap config
│   │       └── db.go
│   ├── service/
│   │   └── config/
│   │       ├── types.go             # owned app-policy types, no setup aliases
│   │       ├── loader.go            # loads from setup source asset during migration
│   │       ├── service.go           # narrow interfaces / immutable snapshots
│   │       └── service_test.go
│   └── domain/
│       ├── markets/                 # receives narrow config structs
│       ├── bets/                    # no direct setup imports
│       └── analytics/               # no direct setup imports
├── handlers/
│   ├── setup/                       # read-only API exposure of approved config slices
│   └── stats/                       # depends on narrow policy views, not raw config tree
└── server/
    └── server.go
```

## Design Rules

The intended direction is:

- keep runtime/bootstrap concerns separate from application-policy concerns
- keep `backend/setup/setup.yaml` as the source asset during this migration
- remove `init()`-driven config loading from the long-term design
- remove package-global mutable configuration access
- stop handing the entire config tree to code that only needs a narrow slice
- stop importing `setup` directly from use-case code
- move from setup aliases to owned configuration types in `internal/service/config`
- validate configuration once at startup and fail fast on invalid state
- favor rolling deployment and restart over live mutation

The intended direction is not:

- one mega configuration service for every concern in the system
- early introduction of hot reload for accounting-sensitive rules
- externalizing the current singleton without first fixing the boundary
- confusing deployment-time settings with business-policy inputs

## Concrete Next Migration Goals

1. Keep using explicit runtime/bootstrap loading in `main.go` and `internal/app/runtime`.
2. Retain `backend/setup/setup.yaml` as the policy source asset for now.
3. Replace type aliases in `internal/service/config` with owned configuration types.
4. Move `setup` to a pure source-asset and translation role, not a package-global runtime access role.
5. Convert domain and handler consumers from whole-tree access to narrow config inputs.
6. Persist accounting-relevant economics at the appropriate workflow boundaries where replay and correctness require durable values.
7. Keep the runtime operational model simple: validate, inject, deploy, restart, verify.

## What This Note Replaces

This update replaces the older recommendation to:

- build a new top-level generic `config/` subsystem
- add environment-specific YAML inheritance as the main architecture move
- introduce hot reload as a featured direction
- treat configuration as primarily a tooling problem

SocialPredict’s real need is boundary completion, deterministic startup, and protection of financial and accounting correctness.
