package users_test

import (
	"context"
	"errors"
	"testing"

	analytics "socialpredict/internal/domain/analytics"
	users "socialpredict/internal/domain/users"
	rusers "socialpredict/internal/repository/users"
	"socialpredict/models/modelstesting"
	"socialpredict/security"
)

type staleUserReader struct {
	user *users.User
}

func (r staleUserReader) GetByUsername(context.Context, string) (*users.User, error) {
	if r.user == nil {
		return nil, users.ErrUserNotFound
	}
	return r.user, nil
}

type authoritativeBalanceRepo struct {
	user          *users.User
	updatedTo     int64
	updateCalled  bool
	failOnMissing bool
}

func (r *authoritativeBalanceRepo) GetByUsername(context.Context, string) (*users.User, error) {
	if r.user == nil {
		return nil, users.ErrUserNotFound
	}
	return r.user, nil
}

func (r *authoritativeBalanceRepo) UpdateBalance(_ context.Context, _ string, newBalance int64) error {
	if r.failOnMissing {
		return errors.New("unexpected balance update")
	}
	r.updatedTo = newBalance
	r.updateCalled = true
	return nil
}

type fakeAnalyticsService struct {
	computeUserFinancialsFn func(context.Context, analytics.FinancialSnapshotRequest) (*analytics.FinancialSnapshot, error)
}

var _ users.AnalyticsService = fakeAnalyticsService{}

func (f fakeAnalyticsService) ComputeUserFinancials(ctx context.Context, req analytics.FinancialSnapshotRequest) (*analytics.FinancialSnapshot, error) {
	if f.computeUserFinancialsFn != nil {
		return f.computeUserFinancialsFn(ctx, req)
	}
	return &analytics.FinancialSnapshot{}, nil
}

func serviceDependencies(repo users.Repository) users.ServiceDependencies {
	return users.ServiceDependencies{
		Reader:      repo,
		BalanceRepo: repo,
		Writer:      repo,
		Lister:      repo,
		Portfolio:   repo,
		Markets:     repo,
		Credentials: repo,
	}
}

func TestServiceApplyTransaction(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := rusers.NewGormRepository(db)
	service := users.NewServiceWithDependencies(serviceDependencies(repo), fakeAnalyticsService{}, security.NewSecurityService().Sanitizer)

	user := modelstesting.GenerateUser("tx_user", 0)
	user.AccountBalance = 100
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	tests := []struct {
		name        string
		txType      users.TransactionType
		amount      int64
		wantBalance int64
		wantErr     bool
	}{
		{"win adds funds", users.TransactionWin, 50, 150, false},
		{"refund adds funds", users.TransactionRefund, 30, 180, false},
		{"sale adds funds", users.TransactionSale, 20, 200, false},
		{"buy subtracts funds", users.TransactionBuy, 40, 160, false},
		{"fee subtracts funds", users.TransactionFee, 10, 150, false},
		{"invalid type", "UNKNOWN", 5, 150, true},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ApplyTransaction(ctx, user.Username, tt.amount, tt.txType)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("ApplyTransaction returned error: %v", err)
			}

			var updatedBalance int64
			if err := db.Model(&user).Select("account_balance").Where("username = ?", user.Username).Scan(&updatedBalance).Error; err != nil {
				t.Fatalf("scan balance: %v", err)
			}
			if updatedBalance != tt.wantBalance {
				t.Fatalf("balance = %d, want %d", updatedBalance, tt.wantBalance)
			}
		})
	}
}

func TestServiceApplyTransactionUsesBalanceRepositoryAsAuthoritativeSource(t *testing.T) {
	balanceRepo := &authoritativeBalanceRepo{
		user: &users.User{
			Username:       "tx_user",
			AccountBalance: 300,
		},
	}
	service := users.NewServiceWithDependencies(users.ServiceDependencies{
		Reader: staleUserReader{
			user: &users.User{
				Username:       "tx_user",
				AccountBalance: 100,
			},
		},
		BalanceRepo: balanceRepo,
	}, fakeAnalyticsService{}, security.NewSecurityService().Sanitizer)

	if err := service.ApplyTransaction(context.Background(), "tx_user", 50, users.TransactionBuy); err != nil {
		t.Fatalf("ApplyTransaction returned error: %v", err)
	}
	if !balanceRepo.updateCalled {
		t.Fatalf("expected authoritative balance repository update")
	}
	if balanceRepo.updatedTo != 250 {
		t.Fatalf("balance update used stale reader value; got %d, want 250", balanceRepo.updatedTo)
	}
}

func TestServiceGetUserCredit(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := rusers.NewGormRepository(db)
	service := users.NewServiceWithDependencies(serviceDependencies(repo), fakeAnalyticsService{}, security.NewSecurityService().Sanitizer)

	user := modelstesting.GenerateUser("credit_user", 0)
	user.AccountBalance = 200
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	ctx := context.Background()

	credit, err := service.GetUserCredit(ctx, user.Username, 500)
	if err != nil {
		t.Fatalf("GetUserCredit returned error: %v", err)
	}
	if credit != 700 {
		t.Fatalf("credit = %d, want 700", credit)
	}

	credit, err = service.GetUserCredit(ctx, "missing_user", 500)
	if err != nil {
		t.Fatalf("expected no error for missing user, got %v", err)
	}
	if credit != 500 {
		t.Fatalf("credit for missing user = %d, want 500", credit)
	}
}

func TestServiceGetUserPortfolio(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	_, _ = modelstesting.UseStandardTestEconomics(t)
	_ = modelstesting.SeedWPAMFromConfig(modelstesting.GenerateEconomicConfig())
	repo := rusers.NewGormRepository(db)
	service := users.NewServiceWithDependencies(serviceDependencies(repo), fakeAnalyticsService{}, security.NewSecurityService().Sanitizer)

	user := modelstesting.GenerateUser("portfolio_user", 0)
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	creator := modelstesting.GenerateUser("creator", 0)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("create creator: %v", err)
	}

	market := modelstesting.GenerateMarket(5001, creator.Username)
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("create market: %v", err)
	}

	bet := modelstesting.GenerateBet(100, "YES", user.Username, uint(market.ID), 0)
	if err := db.Create(&bet).Error; err != nil {
		t.Fatalf("create bet: %v", err)
	}

	ctx := context.Background()
	portfolio, err := service.GetUserPortfolio(ctx, user.Username)
	if err != nil {
		t.Fatalf("GetUserPortfolio returned error: %v", err)
	}

	if portfolio == nil || len(portfolio.Items) != 1 {
		t.Fatalf("expected 1 portfolio item, got %+v", portfolio)
	}

	item := portfolio.Items[0]
	if item.MarketID != uint(market.ID) {
		t.Fatalf("expected market id %d, got %d", market.ID, item.MarketID)
	}
	if item.QuestionTitle != market.QuestionTitle {
		t.Fatalf("expected question title %q, got %q", market.QuestionTitle, item.QuestionTitle)
	}
	if portfolio.TotalSharesOwned == 0 {
		t.Fatalf("expected non-zero total shares, got %d", portfolio.TotalSharesOwned)
	}

	portfolio, err = service.GetUserPortfolio(ctx, "unknown")
	if err != nil {
		t.Fatalf("expected no error for user without bets, got %v", err)
	}
	if len(portfolio.Items) != 0 || portfolio.TotalSharesOwned != 0 {
		t.Fatalf("expected empty portfolio, got %+v", portfolio)
	}
}

func TestServiceGetUserFinancials(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := rusers.NewGormRepository(db)
	analyticsSvc := fakeAnalyticsService{
		computeUserFinancialsFn: func(_ context.Context, req analytics.FinancialSnapshotRequest) (*analytics.FinancialSnapshot, error) {
			return &analytics.FinancialSnapshot{
				AccountBalance:     req.AccountBalance,
				MaximumDebtAllowed: 500,
				AmountInPlay:       80,
				AmountBorrowed:     20,
				RetainedEarnings:   40,
				Equity:             820,
				TradingProfits:     10,
				WorkProfits:        30,
				TotalProfits:       40,
				AmountInPlayActive: 60,
				TotalSpent:         120,
				TotalSpentInPlay:   80,
				RealizedProfits:    15,
				PotentialProfits:   25,
				RealizedValue:      315,
				PotentialValue:     405,
			}, nil
		},
	}
	service := users.NewServiceWithDependencies(serviceDependencies(repo), analyticsSvc, security.NewSecurityService().Sanitizer)

	user := modelstesting.GenerateUser("financial_user", 0)
	user.AccountBalance = 300
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	ctx := context.Background()
	snapshot, err := service.GetUserFinancials(ctx, user.Username)
	if err != nil {
		t.Fatalf("GetUserFinancials returned error: %v", err)
	}

	if snapshot == nil || len(snapshot) == 0 {
		t.Fatalf("expected financial snapshot, got %v", snapshot)
	}
	if _, ok := snapshot["accountBalance"]; !ok {
		t.Fatalf("expected accountBalance in snapshot, got %v", snapshot)
	}

	// Ensure missing user still returns error (since the service expects existing users)
	if _, err := service.GetUserFinancials(ctx, "unknown"); err == nil {
		t.Fatal("expected error for unknown user")
	}
}
