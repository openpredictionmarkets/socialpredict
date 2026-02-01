package users_test

import (
	"context"
	"testing"

	analytics "socialpredict/internal/domain/analytics"
	users "socialpredict/internal/domain/users"
	rusers "socialpredict/internal/repository/users"
	"socialpredict/models/modelstesting"
	"socialpredict/security"
	"socialpredict/setup"
)

type fakeAnalyticsService struct{}

func (fakeAnalyticsService) ComputeUserFinancials(ctx context.Context, req analytics.FinancialSnapshotRequest) (*analytics.FinancialSnapshot, error) {
	return &analytics.FinancialSnapshot{}, nil
}

func TestServiceApplyTransaction(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := rusers.NewGormRepository(db)
	service := users.NewService(repo, fakeAnalyticsService{}, security.NewSecurityService().Sanitizer)

	user := modelstesting.GenerateUser("tx_user", 0)
	user.AccountBalance = 100
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	tests := []struct {
		name        string
		txType      string
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

func TestServiceGetUserCredit(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := rusers.NewGormRepository(db)
	service := users.NewService(repo, fakeAnalyticsService{}, security.NewSecurityService().Sanitizer)

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
	modelstesting.SeedWPAMFromConfig(modelstesting.GenerateEconomicConfig())
	repo := rusers.NewGormRepository(db)
	service := users.NewService(repo, fakeAnalyticsService{}, security.NewSecurityService().Sanitizer)

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
	_, _ = modelstesting.UseStandardTestEconomics(t)
	repo := rusers.NewGormRepository(db)
	config := modelstesting.GenerateEconomicConfig()
	loader := func() *setup.EconomicConfig { return config }
	analyticsSvc := analytics.NewService(analytics.NewGormRepository(db), loader)
	service := users.NewService(repo, analyticsSvc, security.NewSecurityService().Sanitizer)

	user := modelstesting.GenerateUser("financial_user", 0)
	user.AccountBalance = 300
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	creator := modelstesting.GenerateUser("creator_financial", 0)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("create creator: %v", err)
	}

	market := modelstesting.GenerateMarket(6101, creator.Username)
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("create market: %v", err)
	}

	bet := modelstesting.GenerateBet(80, "YES", user.Username, uint(market.ID), 0)
	if err := db.Create(&bet).Error; err != nil {
		t.Fatalf("create bet: %v", err)
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

func TestServiceDeductBalance(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := rusers.NewGormRepository(db)
	service := users.NewService(repo, fakeAnalyticsService{}, security.NewSecurityService().Sanitizer)

	user := modelstesting.GenerateUser("deduct_user", 0)
	user.AccountBalance = 500
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	ctx := context.Background()

	tests := []struct {
		name        string
		amount      int64
		wantBalance int64
		wantErr     bool
	}{
		{"deduct positive amount", 100, 400, false},
		{"deduct to zero", 400, 0, false},
		{"deduct into negative", 50, -50, false},
		{"deduct zero amount", 0, -50, false},
		{"deduct negative amount adds funds", -100, 50, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.DeductBalance(ctx, user.Username, tt.amount)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("DeductBalance returned error: %v", err)
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

func TestServiceDeductBalanceMissingUser(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := rusers.NewGormRepository(db)
	service := users.NewService(repo, fakeAnalyticsService{}, security.NewSecurityService().Sanitizer)

	ctx := context.Background()
	err := service.DeductBalance(ctx, "nonexistent_user", 100)
	if err != users.ErrUserNotFound {
		t.Fatalf("DeductBalance() error = %v, want ErrUserNotFound", err)
	}
}

func TestServiceApplyTransactionZeroAmount(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := rusers.NewGormRepository(db)
	service := users.NewService(repo, fakeAnalyticsService{}, security.NewSecurityService().Sanitizer)

	user := modelstesting.GenerateUser("zero_tx_user", 0)
	user.AccountBalance = 100
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	ctx := context.Background()

	// Zero amount should not change balance
	if err := service.ApplyTransaction(ctx, user.Username, 0, users.TransactionBuy); err != nil {
		t.Fatalf("ApplyTransaction with zero amount returned error: %v", err)
	}

	var balance int64
	if err := db.Model(&user).Select("account_balance").Where("username = ?", user.Username).Scan(&balance).Error; err != nil {
		t.Fatalf("scan balance: %v", err)
	}
	if balance != 100 {
		t.Fatalf("balance after zero transaction = %d, want 100", balance)
	}
}

func TestServiceApplyTransactionNegativeAmount(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := rusers.NewGormRepository(db)
	service := users.NewService(repo, fakeAnalyticsService{}, security.NewSecurityService().Sanitizer)

	user := modelstesting.GenerateUser("neg_tx_user", 0)
	user.AccountBalance = 100
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	ctx := context.Background()

	// BUY with negative amount should ADD funds (subtracts negative)
	if err := service.ApplyTransaction(ctx, user.Username, -50, users.TransactionBuy); err != nil {
		t.Fatalf("ApplyTransaction with negative amount returned error: %v", err)
	}

	var balance int64
	if err := db.Model(&user).Select("account_balance").Where("username = ?", user.Username).Scan(&balance).Error; err != nil {
		t.Fatalf("scan balance: %v", err)
	}
	// BUY subtracts, so subtracting -50 = adding 50
	if balance != 150 {
		t.Fatalf("balance after negative BUY = %d, want 150", balance)
	}
}

func TestServiceValidateUserBalance(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := rusers.NewGormRepository(db)
	service := users.NewService(repo, fakeAnalyticsService{}, security.NewSecurityService().Sanitizer)

	user := modelstesting.GenerateUser("validate_bal_user", 0)
	user.AccountBalance = 100
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	ctx := context.Background()

	tests := []struct {
		name           string
		balance        int64
		requiredAmount int64
		maxDebt        int64
		wantErr        bool
	}{
		{
			name:           "sufficient balance no debt needed",
			balance:        100,
			requiredAmount: 50,
			maxDebt:        500,
			wantErr:        false,
		},
		{
			name:           "exactly at balance",
			balance:        100,
			requiredAmount: 100,
			maxDebt:        500,
			wantErr:        false,
		},
		{
			name:           "requires some debt within limit",
			balance:        100,
			requiredAmount: 200,
			maxDebt:        500,
			wantErr:        false,
		},
		{
			name:           "exactly at max debt limit",
			balance:        100,
			requiredAmount: 600,
			maxDebt:        500,
			wantErr:        false,
		},
		{
			name:           "exceeds max debt by 1",
			balance:        100,
			requiredAmount: 601,
			maxDebt:        500,
			wantErr:        true,
		},
		{
			name:           "way over max debt",
			balance:        100,
			requiredAmount: 1000,
			maxDebt:        500,
			wantErr:        true,
		},
		{
			name:           "zero required amount",
			balance:        100,
			requiredAmount: 0,
			maxDebt:        500,
			wantErr:        false,
		},
		{
			name:           "negative balance within debt limit",
			balance:        -200,
			requiredAmount: 100,
			maxDebt:        500,
			wantErr:        false,
		},
		{
			name:           "negative balance at debt limit",
			balance:        -400,
			requiredAmount: 100,
			maxDebt:        500,
			wantErr:        false,
		},
		{
			name:           "negative balance exceeds debt limit",
			balance:        -400,
			requiredAmount: 101,
			maxDebt:        500,
			wantErr:        true,
		},
		{
			name:           "zero max debt with positive balance",
			balance:        100,
			requiredAmount: 100,
			maxDebt:        0,
			wantErr:        false,
		},
		{
			name:           "zero max debt cannot go negative",
			balance:        100,
			requiredAmount: 101,
			maxDebt:        0,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Update user balance to test value
			if err := db.Model(&user).Update("account_balance", tt.balance).Error; err != nil {
				t.Fatalf("update balance: %v", err)
			}

			err := service.ValidateUserBalance(ctx, user.Username, tt.requiredAmount, tt.maxDebt)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ValidateUserBalance() expected error, got nil")
				}
				if err != users.ErrInsufficientBalance {
					t.Fatalf("ValidateUserBalance() error = %v, want ErrInsufficientBalance", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("ValidateUserBalance() unexpected error: %v", err)
			}
		})
	}
}

func TestServiceValidateUserBalanceMissingUser(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := rusers.NewGormRepository(db)
	service := users.NewService(repo, fakeAnalyticsService{}, security.NewSecurityService().Sanitizer)

	ctx := context.Background()
	err := service.ValidateUserBalance(ctx, "nonexistent_user", 100, 500)
	if err != users.ErrUserNotFound {
		t.Fatalf("ValidateUserBalance() error = %v, want ErrUserNotFound", err)
	}
}

func TestServiceGetUserCreditWithNegativeBalance(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := rusers.NewGormRepository(db)
	service := users.NewService(repo, fakeAnalyticsService{}, security.NewSecurityService().Sanitizer)

	user := modelstesting.GenerateUser("neg_credit_user", 0)
	user.AccountBalance = -200
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	ctx := context.Background()

	// Credit = maxDebt + balance = 500 + (-200) = 300
	credit, err := service.GetUserCredit(ctx, user.Username, 500)
	if err != nil {
		t.Fatalf("GetUserCredit returned error: %v", err)
	}
	if credit != 300 {
		t.Fatalf("credit = %d, want 300", credit)
	}
}

func TestServiceGetUserCreditAtDebtLimit(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := rusers.NewGormRepository(db)
	service := users.NewService(repo, fakeAnalyticsService{}, security.NewSecurityService().Sanitizer)

	user := modelstesting.GenerateUser("debt_limit_user", 0)
	user.AccountBalance = -500 // At max debt limit
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	ctx := context.Background()

	// Credit = maxDebt + balance = 500 + (-500) = 0
	credit, err := service.GetUserCredit(ctx, user.Username, 500)
	if err != nil {
		t.Fatalf("GetUserCredit returned error: %v", err)
	}
	if credit != 0 {
		t.Fatalf("credit = %d, want 0", credit)
	}
}
