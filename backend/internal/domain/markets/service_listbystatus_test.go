package markets_test

import (
	"context"
	"errors"
	"sort"
	"testing"
	"time"

	markets "socialpredict/internal/domain/markets"
	"socialpredict/internal/domain/math/probabilities/wpam"
	dusers "socialpredict/internal/domain/users"
	rmarkets "socialpredict/internal/repository/markets"
	"socialpredict/models"
	"socialpredict/models/modelstesting"

	"gorm.io/gorm"
)

var errUnexpectedMarketsTestCall = errors.New("unexpected markets test call")

type noopUserService struct {
	validateUserExistsFunc  func(context.Context, string) error
	validateUserBalanceFunc func(context.Context, string, int64, int64) error
	deductBalanceFunc       func(context.Context, string, int64) error
	applyTransactionFunc    func(context.Context, string, int64, string) error
	getPublicUserFunc       func(context.Context, string) (*dusers.PublicUser, error)
}

func marketsTestTime() time.Time {
	return time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
}

func newNoopUserService(opts ...func(*noopUserService)) noopUserService {
	service := noopUserService{
		validateUserExistsFunc:  func(context.Context, string) error { return nil },
		validateUserBalanceFunc: func(context.Context, string, int64, int64) error { return nil },
		deductBalanceFunc:       func(context.Context, string, int64) error { return nil },
		applyTransactionFunc:    func(context.Context, string, int64, string) error { return nil },
		getPublicUserFunc:       func(context.Context, string) (*dusers.PublicUser, error) { return nil, nil },
	}
	for _, opt := range opts {
		opt(&service)
	}
	return service
}

func (s noopUserService) ValidateUserExists(ctx context.Context, username string) error {
	if s.validateUserExistsFunc == nil {
		return errUnexpectedMarketsTestCall
	}
	return s.validateUserExistsFunc(ctx, username)
}

func (s noopUserService) ValidateUserBalance(ctx context.Context, username string, requiredAmount int64, maxDebt int64) error {
	if s.validateUserBalanceFunc == nil {
		return errUnexpectedMarketsTestCall
	}
	return s.validateUserBalanceFunc(ctx, username, requiredAmount, maxDebt)
}

func (s noopUserService) DeductBalance(ctx context.Context, username string, amount int64) error {
	if s.deductBalanceFunc == nil {
		return errUnexpectedMarketsTestCall
	}
	return s.deductBalanceFunc(ctx, username, amount)
}

func (s noopUserService) ApplyTransaction(ctx context.Context, username string, amount int64, transactionType string) error {
	if s.applyTransactionFunc == nil {
		return errUnexpectedMarketsTestCall
	}
	return s.applyTransactionFunc(ctx, username, amount, transactionType)
}

func (s noopUserService) GetPublicUser(ctx context.Context, username string) (*dusers.PublicUser, error) {
	if s.getPublicUserFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return s.getPublicUserFunc(ctx, username)
}

type fixedClock struct {
	nowFunc func() time.Time
}

func newFixedClock(now time.Time) fixedClock {
	return fixedClock{
		nowFunc: func() time.Time { return now },
	}
}

func (f fixedClock) Now() time.Time {
	if f.nowFunc == nil {
		return marketsTestTime()
	}
	return f.nowFunc()
}

func setupServiceWithDB(t *testing.T) (*markets.Service, *gorm.DB, wpam.ProbabilityCalculator) {
	t.Helper()

	econ := modelstesting.GenerateEconomicConfig()
	calculator := wpam.NewProbabilityCalculator(wpam.StaticSeedProvider{Value: wpam.Seeds{
		InitialProbability:     econ.Economics.MarketCreation.InitialMarketProbability,
		InitialSubsidization:   econ.Economics.MarketCreation.InitialMarketSubsidization,
		InitialYesContribution: econ.Economics.MarketCreation.InitialMarketYes,
		InitialNoContribution:  econ.Economics.MarketCreation.InitialMarketNo,
	}})

	db := modelstesting.NewFakeDB(t)
	repo := rmarkets.NewGormRepository(db)
	clock := newFixedClock(marketsTestTime())
	cfg := markets.Config{}

	service := markets.NewService(
		repo,
		newNoopUserService(),
		clock,
		cfg,
		markets.WithProbabilityEngine(markets.DefaultProbabilityEngine(calculator)),
	)
	return service, db, calculator
}

func buildStatusMarkets(username string, now time.Time) []models.Market {
	return []models.Market{
		{
			ID:                 1,
			QuestionTitle:      "Active Market",
			Description:        "Active",
			OutcomeType:        "BINARY",
			ResolutionDateTime: now.Add(24 * time.Hour),
			IsResolved:         false,
			InitialProbability: 0.5,
			CreatorUsername:    username,
		},
		{
			ID:                 2,
			QuestionTitle:      "Closed Market",
			Description:        "Closed",
			OutcomeType:        "BINARY",
			ResolutionDateTime: now.Add(-24 * time.Hour),
			IsResolved:         false,
			InitialProbability: 0.5,
			CreatorUsername:    username,
		},
		{
			ID:                      3,
			QuestionTitle:           "Resolved Market",
			Description:             "Resolved",
			OutcomeType:             "BINARY",
			ResolutionDateTime:      now.Add(-48 * time.Hour),
			FinalResolutionDateTime: now.Add(-24 * time.Hour),
			IsResolved:              true,
			ResolutionResult:        "YES",
			InitialProbability:      0.5,
			CreatorUsername:         username,
		},
	}
}

func assertSortedIDs(t *testing.T, got []int64, want []int64) {
	t.Helper()
	sort.Slice(got, func(i, j int) bool { return got[i] < got[j] })
	if len(got) != len(want) {
		t.Fatalf("expected %d markets, got %d (ids=%v)", len(want), len(got), got)
	}
	for i, id := range got {
		if id != want[i] {
			t.Fatalf("expected ids %v, got %v", want, got)
		}
	}
}

func TestServiceListByStatusFiltersMarkets(t *testing.T) {
	service, db, _ := setupServiceWithDB(t)

	now := time.Now()

	user := modelstesting.GenerateUser("testuser", 1000)
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	for _, market := range buildStatusMarkets(user.Username, now) {
		if err := db.Create(&market).Error; err != nil {
			t.Fatalf("create market %s: %v", market.QuestionTitle, err)
		}
	}

	tests := []struct {
		name        string
		status      string
		expectedIDs []int64
	}{
		{
			name:        "Active Markets",
			status:      "active",
			expectedIDs: []int64{1},
		},
		{
			name:        "Closed Markets",
			status:      "closed",
			expectedIDs: []int64{2},
		},
		{
			name:        "Resolved Markets",
			status:      "resolved",
			expectedIDs: []int64{3},
		},
		{
			name:        "All Markets",
			status:      "all",
			expectedIDs: []int64{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := service.ListByStatus(context.Background(), tt.status, markets.Page{Limit: 10})
			if err != nil {
				t.Fatalf("ListByStatus returned error: %v", err)
			}

			var ids []int64
			for _, market := range results {
				ids = append(ids, market.ID)
			}
			assertSortedIDs(t, ids, tt.expectedIDs)
		})
	}

	if got := (fixedClock{}).Now(); !got.Equal(marketsTestTime()) {
		t.Fatalf("expected zero-value clock fallback, got %v", got)
	}
}

func TestServiceListByStatusInvalidStatus(t *testing.T) {
	service, _, _ := setupServiceWithDB(t)

	_, err := service.ListByStatus(context.Background(), "unknown", markets.Page{})
	if !errors.Is(err, markets.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}

	if err := (noopUserService{}).ValidateUserExists(context.Background(), "alice"); !errors.Is(err, errUnexpectedMarketsTestCall) {
		t.Fatalf("expected zero-value user service to fail predictably, got %v", err)
	}
}
