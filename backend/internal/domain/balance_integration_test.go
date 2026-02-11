package domain_test

import (
	"context"
	"testing"
	"time"

	analytics "socialpredict/internal/domain/analytics"
	bets "socialpredict/internal/domain/bets"
	markets "socialpredict/internal/domain/markets"
	users "socialpredict/internal/domain/users"
	dwallet "socialpredict/internal/domain/wallet"
	rbets "socialpredict/internal/repository/bets"
	rmarkets "socialpredict/internal/repository/markets"
	rusers "socialpredict/internal/repository/users"
	rwallet "socialpredict/internal/repository/wallet"
	"socialpredict/models/modelstesting"
	"socialpredict/security"
)

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

type fakeAnalyticsService struct{}

func (fakeAnalyticsService) ComputeUserFinancials(ctx context.Context, req analytics.FinancialSnapshotRequest) (*analytics.FinancialSnapshot, error) {
	return &analytics.FinancialSnapshot{}, nil
}

func TestIntegration_BetPlacement_DeductsBalance(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	_, _ = modelstesting.UseStandardTestEconomics(t)
	modelstesting.SeedWPAMFromConfig(modelstesting.GenerateEconomicConfig())

	// Setup services with real repositories
	userRepo := rusers.NewGormRepository(db)
	userService := users.NewService(userRepo, fakeAnalyticsService{}, security.NewSecurityService().Sanitizer)
	walletRepo := rwallet.NewGormRepository(db)
	walletService := dwallet.NewService(walletRepo, realClock{})

	marketRepo := rmarkets.NewGormRepository(db)
	marketConfig := markets.Config{
		MinimumFutureHours: 1,
		CreateMarketCost:   10,
		MaximumDebtAllowed: 500,
	}
	marketService := markets.NewServiceWithWallet(marketRepo, userService, walletService, realClock{}, marketConfig)

	betRepo := rbets.NewGormRepository(db)
	econ := modelstesting.GenerateEconomicConfig()
	betService := bets.NewServiceWithWallet(betRepo, marketService, walletService, econ, realClock{})

	ctx := context.Background()

	// Create user with known balance
	user := modelstesting.GenerateUser("bet_integration_user", 1000)
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	// Create market creator
	creator := modelstesting.GenerateUser("market_creator", 1000)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("create creator: %v", err)
	}

	// Create market
	market := modelstesting.GenerateMarket(9001, creator.Username)
	market.ResolutionDateTime = time.Now().Add(48 * time.Hour)
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("create market: %v", err)
	}

	initialBalance := user.AccountBalance

	// Place a bet
	betAmount := int64(100)
	_, err := betService.Place(ctx, bets.PlaceRequest{
		Username: user.Username,
		MarketID: uint(market.ID),
		Amount:   betAmount,
		Outcome:  "YES",
	})
	if err != nil {
		t.Fatalf("Place bet returned error: %v", err)
	}

	// Verify balance was deducted
	var finalBalance int64
	if err := db.Model(&user).Select("account_balance").Where("username = ?", user.Username).Scan(&finalBalance).Error; err != nil {
		t.Fatalf("scan balance: %v", err)
	}

	expectedFees := econ.Economics.Betting.BetFees.InitialBetFee + econ.Economics.Betting.BetFees.BuySharesFee
	expectedDeduction := betAmount + int64(expectedFees)
	expectedBalance := initialBalance - expectedDeduction

	if finalBalance != expectedBalance {
		t.Fatalf("balance after bet = %d, want %d (initial=%d, bet=%d, fees=%d)",
			finalBalance, expectedBalance, initialBalance, betAmount, expectedFees)
	}
}

func TestIntegration_BetPlacement_InsufficientBalance_Fails(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	_, _ = modelstesting.UseStandardTestEconomics(t)
	modelstesting.SeedWPAMFromConfig(modelstesting.GenerateEconomicConfig())

	userRepo := rusers.NewGormRepository(db)
	userService := users.NewService(userRepo, fakeAnalyticsService{}, security.NewSecurityService().Sanitizer)
	walletRepo := rwallet.NewGormRepository(db)
	walletService := dwallet.NewService(walletRepo, realClock{})

	marketRepo := rmarkets.NewGormRepository(db)
	marketConfig := markets.Config{
		MinimumFutureHours: 1,
		CreateMarketCost:   10,
		MaximumDebtAllowed: 100, // Low debt limit
	}
	marketService := markets.NewServiceWithWallet(marketRepo, userService, walletService, realClock{}, marketConfig)

	betRepo := rbets.NewGormRepository(db)
	econ := modelstesting.GenerateEconomicConfig()
	econ.Economics.User.MaximumDebtAllowed = 100 // Match the market config
	betService := bets.NewServiceWithWallet(betRepo, marketService, walletService, econ, realClock{})

	ctx := context.Background()

	// Create user with low balance
	user := modelstesting.GenerateUser("poor_user", 0)
	user.AccountBalance = 50
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	creator := modelstesting.GenerateUser("creator2", 1000)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("create creator: %v", err)
	}

	market := modelstesting.GenerateMarket(9002, creator.Username)
	market.ResolutionDateTime = time.Now().Add(48 * time.Hour)
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("create market: %v", err)
	}

	// Try to place a bet exceeding available credit (balance + maxDebt = 50 + 100 = 150)
	_, err := betService.Place(ctx, bets.PlaceRequest{
		Username: user.Username,
		MarketID: uint(market.ID),
		Amount:   200, // Exceeds available credit
		Outcome:  "YES",
	})

	if err == nil {
		t.Fatalf("expected error for insufficient balance, got nil")
	}

	// Verify balance was NOT changed
	var finalBalance int64
	if err := db.Model(&user).Select("account_balance").Where("username = ?", user.Username).Scan(&finalBalance).Error; err != nil {
		t.Fatalf("scan balance: %v", err)
	}

	if finalBalance != 50 {
		t.Fatalf("balance should be unchanged at 50, got %d", finalBalance)
	}
}

func TestIntegration_BetPlacement_AtDebtLimit(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	_, _ = modelstesting.UseStandardTestEconomics(t)
	modelstesting.SeedWPAMFromConfig(modelstesting.GenerateEconomicConfig())

	userRepo := rusers.NewGormRepository(db)
	userService := users.NewService(userRepo, fakeAnalyticsService{}, security.NewSecurityService().Sanitizer)
	walletRepo := rwallet.NewGormRepository(db)
	walletService := dwallet.NewService(walletRepo, realClock{})

	marketRepo := rmarkets.NewGormRepository(db)
	marketConfig := markets.Config{
		MinimumFutureHours: 1,
		CreateMarketCost:   10,
		MaximumDebtAllowed: 100,
	}
	marketService := markets.NewServiceWithWallet(marketRepo, userService, walletService, realClock{}, marketConfig)

	betRepo := rbets.NewGormRepository(db)
	econ := modelstesting.GenerateEconomicConfig()
	econ.Economics.User.MaximumDebtAllowed = 100
	econ.Economics.Betting.BetFees.InitialBetFee = 0
	econ.Economics.Betting.BetFees.BuySharesFee = 0
	betService := bets.NewServiceWithWallet(betRepo, marketService, walletService, econ, realClock{})

	ctx := context.Background()

	// User at exactly the debt limit
	user := modelstesting.GenerateUser("debt_limit_user", 0)
	user.AccountBalance = -100 // At max debt
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	creator := modelstesting.GenerateUser("creator3", 1000)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("create creator: %v", err)
	}

	market := modelstesting.GenerateMarket(9003, creator.Username)
	market.ResolutionDateTime = time.Now().Add(48 * time.Hour)
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("create market: %v", err)
	}

	// Try to place any bet - should fail since already at debt limit
	_, err := betService.Place(ctx, bets.PlaceRequest{
		Username: user.Username,
		MarketID: uint(market.ID),
		Amount:   1, // Even smallest amount should fail
		Outcome:  "YES",
	})

	if err == nil {
		t.Fatalf("expected error when at debt limit, got nil")
	}
}

func TestIntegration_MarketCreation_DeductsBalance(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	_, _ = modelstesting.UseStandardTestEconomics(t)
	modelstesting.SeedWPAMFromConfig(modelstesting.GenerateEconomicConfig())

	userRepo := rusers.NewGormRepository(db)
	userService := users.NewService(userRepo, fakeAnalyticsService{}, security.NewSecurityService().Sanitizer)
	walletRepo := rwallet.NewGormRepository(db)
	walletService := dwallet.NewService(walletRepo, realClock{})

	marketRepo := rmarkets.NewGormRepository(db)
	createCost := int64(50)
	marketConfig := markets.Config{
		MinimumFutureHours: 1,
		CreateMarketCost:   createCost,
		MaximumDebtAllowed: 500,
	}
	marketService := markets.NewServiceWithWallet(marketRepo, userService, walletService, realClock{}, marketConfig)

	ctx := context.Background()

	// Create user with known balance
	user := modelstesting.GenerateUser("market_creator_user", 1000)
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	initialBalance := user.AccountBalance

	// Create a market
	_, err := marketService.CreateMarket(ctx, markets.MarketCreateRequest{
		QuestionTitle:      "Will this test pass?",
		Description:        "Integration test for market creation",
		OutcomeType:        "BINARY",
		ResolutionDateTime: time.Now().Add(48 * time.Hour),
	}, user.Username)

	if err != nil {
		t.Fatalf("CreateMarket returned error: %v", err)
	}

	// Verify balance was deducted
	var finalBalance int64
	if err := db.Model(&user).Select("account_balance").Where("username = ?", user.Username).Scan(&finalBalance).Error; err != nil {
		t.Fatalf("scan balance: %v", err)
	}

	expectedBalance := initialBalance - createCost
	if finalBalance != expectedBalance {
		t.Fatalf("balance after market creation = %d, want %d", finalBalance, expectedBalance)
	}
}

func TestIntegration_MarketCreation_InsufficientBalance_Fails(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	_, _ = modelstesting.UseStandardTestEconomics(t)

	userRepo := rusers.NewGormRepository(db)
	userService := users.NewService(userRepo, fakeAnalyticsService{}, security.NewSecurityService().Sanitizer)
	walletRepo := rwallet.NewGormRepository(db)
	walletService := dwallet.NewService(walletRepo, realClock{})

	marketRepo := rmarkets.NewGormRepository(db)
	createCost := int64(100)
	marketConfig := markets.Config{
		MinimumFutureHours: 1,
		CreateMarketCost:   createCost,
		MaximumDebtAllowed: 50, // Low debt limit
	}
	marketService := markets.NewServiceWithWallet(marketRepo, userService, walletService, realClock{}, marketConfig)

	ctx := context.Background()

	// Create user with insufficient balance
	user := modelstesting.GenerateUser("poor_creator", 0)
	user.AccountBalance = 20 // Less than create cost and debt limit combined
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	// Try to create a market
	_, err := marketService.CreateMarket(ctx, markets.MarketCreateRequest{
		QuestionTitle:      "This should fail",
		Description:        "Not enough funds",
		OutcomeType:        "BINARY",
		ResolutionDateTime: time.Now().Add(48 * time.Hour),
	}, user.Username)

	if err == nil {
		t.Fatalf("expected error for insufficient balance, got nil")
	}

	// Verify balance was NOT changed
	var finalBalance int64
	if err := db.Model(&user).Select("account_balance").Where("username = ?", user.Username).Scan(&finalBalance).Error; err != nil {
		t.Fatalf("scan balance: %v", err)
	}

	if finalBalance != 20 {
		t.Fatalf("balance should be unchanged at 20, got %d", finalBalance)
	}
}

func TestIntegration_MultipleBets_CumulativeBalanceChange(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	_, _ = modelstesting.UseStandardTestEconomics(t)
	modelstesting.SeedWPAMFromConfig(modelstesting.GenerateEconomicConfig())

	userRepo := rusers.NewGormRepository(db)
	userService := users.NewService(userRepo, fakeAnalyticsService{}, security.NewSecurityService().Sanitizer)
	walletRepo := rwallet.NewGormRepository(db)
	walletService := dwallet.NewService(walletRepo, realClock{})

	marketRepo := rmarkets.NewGormRepository(db)
	marketConfig := markets.Config{
		MinimumFutureHours: 1,
		CreateMarketCost:   10,
		MaximumDebtAllowed: 500,
	}
	marketService := markets.NewServiceWithWallet(marketRepo, userService, walletService, realClock{}, marketConfig)

	betRepo := rbets.NewGormRepository(db)
	econ := modelstesting.GenerateEconomicConfig()
	econ.Economics.Betting.BetFees.InitialBetFee = 0
	econ.Economics.Betting.BetFees.BuySharesFee = 0
	betService := bets.NewServiceWithWallet(betRepo, marketService, walletService, econ, realClock{})

	ctx := context.Background()

	user := modelstesting.GenerateUser("multi_bet_user", 1000)
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	creator := modelstesting.GenerateUser("creator4", 1000)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("create creator: %v", err)
	}

	market := modelstesting.GenerateMarket(9004, creator.Username)
	market.ResolutionDateTime = time.Now().Add(48 * time.Hour)
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("create market: %v", err)
	}

	// Place multiple bets
	betAmounts := []int64{100, 150, 200}
	var totalBet int64

	for i, amount := range betAmounts {
		_, err := betService.Place(ctx, bets.PlaceRequest{
			Username: user.Username,
			MarketID: uint(market.ID),
			Amount:   amount,
			Outcome:  "YES",
		})
		if err != nil {
			t.Fatalf("bet %d failed: %v", i+1, err)
		}
		totalBet += amount
	}

	// Verify cumulative balance change
	var finalBalance int64
	if err := db.Model(&user).Select("account_balance").Where("username = ?", user.Username).Scan(&finalBalance).Error; err != nil {
		t.Fatalf("scan balance: %v", err)
	}

	// Account for initial bet fee on first bet only (fees are 0 in this test)
	expectedBalance := int64(1000) - totalBet
	if finalBalance != expectedBalance {
		t.Fatalf("balance after multiple bets = %d, want %d", finalBalance, expectedBalance)
	}
}

func TestIntegration_BalanceGoesNegative_WithinDebtLimit(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	_, _ = modelstesting.UseStandardTestEconomics(t)
	modelstesting.SeedWPAMFromConfig(modelstesting.GenerateEconomicConfig())

	userRepo := rusers.NewGormRepository(db)
	userService := users.NewService(userRepo, fakeAnalyticsService{}, security.NewSecurityService().Sanitizer)
	walletRepo := rwallet.NewGormRepository(db)
	walletService := dwallet.NewService(walletRepo, realClock{})

	marketRepo := rmarkets.NewGormRepository(db)
	marketConfig := markets.Config{
		MinimumFutureHours: 1,
		CreateMarketCost:   10,
		MaximumDebtAllowed: 500,
	}
	marketService := markets.NewServiceWithWallet(marketRepo, userService, walletService, realClock{}, marketConfig)

	betRepo := rbets.NewGormRepository(db)
	econ := modelstesting.GenerateEconomicConfig()
	econ.Economics.Betting.BetFees.InitialBetFee = 0
	econ.Economics.Betting.BetFees.BuySharesFee = 0
	betService := bets.NewServiceWithWallet(betRepo, marketService, walletService, econ, realClock{})

	ctx := context.Background()

	// User with small positive balance
	user := modelstesting.GenerateUser("will_go_negative_user", 50)
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	creator := modelstesting.GenerateUser("creator5", 1000)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("create creator: %v", err)
	}

	market := modelstesting.GenerateMarket(9005, creator.Username)
	market.ResolutionDateTime = time.Now().Add(48 * time.Hour)
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("create market: %v", err)
	}

	// Place a bet that will put balance negative
	betAmount := int64(200) // Will result in -150 balance
	_, err := betService.Place(ctx, bets.PlaceRequest{
		Username: user.Username,
		MarketID: uint(market.ID),
		Amount:   betAmount,
		Outcome:  "YES",
	})

	if err != nil {
		t.Fatalf("bet should succeed within debt limit: %v", err)
	}

	var finalBalance int64
	if err := db.Model(&user).Select("account_balance").Where("username = ?", user.Username).Scan(&finalBalance).Error; err != nil {
		t.Fatalf("scan balance: %v", err)
	}

	expectedBalance := int64(50) - betAmount // -150
	if finalBalance != expectedBalance {
		t.Fatalf("balance = %d, want %d (negative within debt limit)", finalBalance, expectedBalance)
	}

	if finalBalance >= 0 {
		t.Fatalf("expected negative balance, got %d", finalBalance)
	}
}
