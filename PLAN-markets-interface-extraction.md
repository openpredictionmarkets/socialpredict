# Plan: Markets Interface Extraction

## Goal

Extract interfaces from `service.go` into a dedicated `interfaces.go` file, splitting the monolithic `ServiceInterface` into focused, single-responsibility interfaces.

## Current State

`service.go` (182 lines) contains:
- **Constants**: `MaxQuestionTitleLength`, `MaxDescriptionLength`, `MaxLabelLength`, `MinLabelLength`
- **Interfaces**: `Clock`, `Repository`, `UserService`, `ServiceInterface`
- **Config/DTOs**: `Config`, `ListFilters`, `SearchFilters`, `SearchResults`, `MarketOverview`, `Page`, `LeaderboardRow`, `ProbabilityProjectionRequest`, `ProbabilityProjection`, `BetDisplayInfo`, `CreatorSummary`, `ProbabilityPoint`
- **Service struct** + `NewService` constructor

The `ServiceInterface` has 12 methods spanning 5 different concerns:
1. Core CRUD: `CreateMarket`, `SetCustomLabels`, `GetMarket`, `GetPublicMarket`
2. Search/List: `ListMarkets`, `SearchMarkets`, `ListByStatus`
3. Leaderboard: `GetMarketLeaderboard`
4. Positions: `GetMarketPositions`, `GetUserPositionInMarket`
5. Projection/Display: `ProjectProbability`, `GetMarketDetails`, `GetMarketBets`, `CalculateMarketVolume`

## Target State

### File: `interfaces.go`

```go
package markets

import (
    "context"
    "time"
    users "socialpredict/internal/domain/users"
)

// Clock provides time functionality for testability.
type Clock interface {
    Now() time.Time
}

// Repository defines the interface for market data access.
type Repository interface {
    Create(ctx context.Context, market *Market) error
    GetByID(ctx context.Context, id int64) (*Market, error)
    UpdateLabels(ctx context.Context, id int64, yesLabel, noLabel string) error
    List(ctx context.Context, filters ListFilters) ([]*Market, error)
    ListByStatus(ctx context.Context, status string, p Page) ([]*Market, error)
    Search(ctx context.Context, query string, filters SearchFilters) ([]*Market, error)
    Delete(ctx context.Context, id int64) error
    ResolveMarket(ctx context.Context, id int64, resolution string) error
    GetUserPosition(ctx context.Context, marketID int64, username string) (*UserPosition, error)
    ListMarketPositions(ctx context.Context, marketID int64) (MarketPositions, error)
    ListBetsForMarket(ctx context.Context, marketID int64) ([]*Bet, error)
    CalculatePayoutPositions(ctx context.Context, marketID int64) ([]*PayoutPosition, error)
    GetPublicMarket(ctx context.Context, marketID int64) (*PublicMarket, error)
}

// UserService defines the interface for user-related operations.
type UserService interface {
    ValidateUserExists(ctx context.Context, username string) error
    ValidateUserBalance(ctx context.Context, username string, requiredAmount int64, maxDebt int64) error
    DeductBalance(ctx context.Context, username string, amount int64) error
    ApplyTransaction(ctx context.Context, username string, amount int64, transactionType string) error
    GetPublicUser(ctx context.Context, username string) (*users.PublicUser, error)
}

// --- Focused Service Interfaces ---

// CoreService handles core market CRUD operations.
type CoreService interface {
    CreateMarket(ctx context.Context, req MarketCreateRequest, creatorUsername string) (*Market, error)
    SetCustomLabels(ctx context.Context, marketID int64, yesLabel, noLabel string) error
    GetMarket(ctx context.Context, id int64) (*Market, error)
    GetPublicMarket(ctx context.Context, marketID int64) (*PublicMarket, error)
    ResolveMarket(ctx context.Context, marketID int64, resolution string, username string) error
}

// SearchService handles market search and listing.
type SearchService interface {
    ListMarkets(ctx context.Context, filters ListFilters) ([]*Market, error)
    SearchMarkets(ctx context.Context, query string, filters SearchFilters) (*SearchResults, error)
    ListByStatus(ctx context.Context, status string, p Page) ([]*Market, error)
}

// LeaderboardService handles market leaderboard calculations.
type LeaderboardService interface {
    GetMarketLeaderboard(ctx context.Context, marketID int64, p Page) ([]*LeaderboardRow, error)
}

// PositionsService handles user position queries.
type PositionsService interface {
    GetMarketPositions(ctx context.Context, marketID int64) (MarketPositions, error)
    GetUserPositionInMarket(ctx context.Context, marketID int64, username string) (*UserPosition, error)
}

// ProjectionService handles probability projections and market details.
type ProjectionService interface {
    ProjectProbability(ctx context.Context, req ProbabilityProjectionRequest) (*ProbabilityProjection, error)
    GetMarketDetails(ctx context.Context, marketID int64) (*MarketOverview, error)
    GetMarketBets(ctx context.Context, marketID int64) ([]*BetDisplayInfo, error)
    CalculateMarketVolume(ctx context.Context, marketID int64) (int64, error)
}

// ServiceInterface combines all market service capabilities.
// Deprecated: Prefer using focused interfaces (CoreService, SearchService, etc.) for new code.
type ServiceInterface interface {
    CoreService
    SearchService
    LeaderboardService
    PositionsService
    ProjectionService
}
```

### File: `service.go` (after extraction)

```go
package markets

// Constants for validation limits.
const (
    MaxQuestionTitleLength = 160
    MaxDescriptionLength   = 2000
    MaxLabelLength         = 20
    MinLabelLength         = 1
)

// Config holds configuration for the markets service.
type Config struct {
    MinimumFutureHours float64
    CreateMarketCost   int64
    MaximumDebtAllowed int64
}

// Service implements the core market business logic.
type Service struct {
    repo        Repository
    userService UserService
    clock       Clock
    config      Config
}

// NewService creates a new markets service.
func NewService(repo Repository, userService UserService, clock Clock, config Config) *Service {
    return &Service{
        repo:        repo,
        userService: userService,
        clock:       clock,
        config:      config,
    }
}

// Compile-time interface compliance checks.
var (
    _ ServiceInterface   = (*Service)(nil)
    _ CoreService        = (*Service)(nil)
    _ SearchService      = (*Service)(nil)
    _ LeaderboardService = (*Service)(nil)
    _ PositionsService   = (*Service)(nil)
    _ ProjectionService  = (*Service)(nil)
)
```

### File: `types.go` (new - extract DTOs)

```go
package markets

import "time"

// Page represents pagination parameters.
type Page struct {
    Limit  int
    Offset int
}

// ListFilters represents filters for listing markets.
type ListFilters struct {
    Status    string
    CreatedBy string
    Limit     int
    Offset    int
}

// SearchFilters represents filters for searching markets.
type SearchFilters struct {
    Status string
    Limit  int
    Offset int
}

// SearchResults represents the result of a market search with fallback.
type SearchResults struct {
    PrimaryResults  []*Market `json:"primaryResults"`
    FallbackResults []*Market `json:"fallbackResults"`
    Query           string    `json:"query"`
    PrimaryStatus   string    `json:"primaryStatus"`
    PrimaryCount    int       `json:"primaryCount"`
    FallbackCount   int       `json:"fallbackCount"`
    TotalCount      int       `json:"totalCount"`
    FallbackUsed    bool      `json:"fallbackUsed"`
}

// CreatorSummary captures lightweight information about a market creator.
type CreatorSummary struct {
    Username      string
    DisplayName   string
    PersonalEmoji string
}

// ProbabilityPoint records a market probability at a specific moment.
type ProbabilityPoint struct {
    Probability float64
    Timestamp   time.Time
}

// MarketOverview represents enriched market data with calculations.
type MarketOverview struct {
    Market             *Market
    Creator            *CreatorSummary
    ProbabilityChanges []ProbabilityPoint
    LastProbability    float64
    NumUsers           int
    TotalVolume        int64
    MarketDust         int64
}

// LeaderboardRow represents a single row in the market leaderboard.
type LeaderboardRow struct {
    Username       string
    Profit         int64
    CurrentValue   int64
    TotalSpent     int64
    Position       string
    YesSharesOwned int64
    NoSharesOwned  int64
    Rank           int
}

// ProbabilityProjectionRequest represents a request for probability projection.
type ProbabilityProjectionRequest struct {
    MarketID int64
    Amount   int64
    Outcome  string
}

// ProbabilityProjection represents the result of a probability projection.
type ProbabilityProjection struct {
    CurrentProbability   float64
    ProjectedProbability float64
}

// BetDisplayInfo represents a bet with probability information.
type BetDisplayInfo struct {
    Username    string    `json:"username"`
    Outcome     string    `json:"outcome"`
    Amount      int64     `json:"amount"`
    Probability float64   `json:"probability"`
    PlacedAt    time.Time `json:"placedAt"`
}
```

---

## Execution Steps

### Step 1: Create `interfaces.go`

1. Create new file `backend/internal/domain/markets/interfaces.go`
2. Move interfaces: `Clock`, `Repository`, `UserService`
3. Add new focused interfaces: `CoreService`, `SearchService`, `LeaderboardService`, `PositionsService`, `ProjectionService`
4. Redefine `ServiceInterface` as composition of focused interfaces

### Step 2: Create `types.go`

1. Create new file `backend/internal/domain/markets/types.go`
2. Move DTOs: `Page`, `ListFilters`, `SearchFilters`, `SearchResults`, `CreatorSummary`, `ProbabilityPoint`, `MarketOverview`, `LeaderboardRow`, `ProbabilityProjectionRequest`, `ProbabilityProjection`, `BetDisplayInfo`

### Step 3: Slim down `service.go`

1. Remove moved interfaces (now in `interfaces.go`)
2. Remove moved DTOs (now in `types.go`)
3. Keep: constants, `Config`, `Service` struct, `NewService`
4. Add compile-time interface checks

### Step 4: Verify compilation

```bash
cd backend && go build ./...
```

### Step 5: Run tests

```bash
cd backend && go test ./internal/domain/markets/...
```

---

## File Summary After Refactor

| File | Contents |
|------|----------|
| `interfaces.go` | All interfaces (`Clock`, `Repository`, `UserService`, focused service interfaces, `ServiceInterface`) |
| `types.go` | All DTOs and filter types |
| `models.go` | Domain models (`Market`, `MarketCreateRequest`, `UserPosition`, `Bet`, etc.) |
| `errors.go` | Error definitions |
| `service.go` | Constants, `Config`, `Service` struct, `NewService`, interface compliance checks |
| `market_*.go` | Method implementations (unchanged) |

---

## Benefits

1. **Interface segregation**: Consumers can depend on only the interfaces they need
2. **Clearer organization**: Interfaces, types, and implementation separated
3. **Easier testing**: Mock only the specific interface needed
4. **Documentation**: Each interface documents a cohesive capability
5. **Backward compatible**: `ServiceInterface` still exists as composition

---

## Risk Assessment

| Risk | Mitigation |
|------|------------|
| Import cycles | All types stay in same package - no risk |
| Breaking callers | `ServiceInterface` preserved as composition - no breaking change |
| Test failures | Run full test suite after each step |

**Risk Level: Low** - This is a pure reorganization within the same package.
