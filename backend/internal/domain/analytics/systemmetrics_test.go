package analytics

import (
	"context"
	"testing"

	"socialpredict/models"
	"socialpredict/models/modelstesting"
	"socialpredict/setup"
)

func TestComputeSystemMetrics_EmptyDatabase(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	econ := modelstesting.GenerateEconomicConfig()

	svc := NewService(NewGormRepository(db), func() *setup.EconomicConfig { return econ })

	metrics, err := svc.ComputeSystemMetrics(context.Background())
	if err != nil {
		t.Fatalf("ComputeSystemMetrics returned error: %v", err)
	}

	if val, ok := metrics.MoneyCreated.UserDebtCapacity.Value.(int64); !ok || val != 0 {
		t.Fatalf("expected user debt capacity 0, got %v", metrics.MoneyCreated.UserDebtCapacity.Value)
	}
	if val, ok := metrics.MoneyUtilized.TotalUtilized.Value.(int64); !ok || val != 0 {
		t.Fatalf("expected total utilized 0, got %v", metrics.MoneyUtilized.TotalUtilized.Value)
	}
}

func TestComputeSystemMetrics_WithData(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	econ := modelstesting.GenerateEconomicConfig()
	econ.Economics.MarketIncentives.CreateMarketCost = 50
	econ.Economics.Betting.BetFees.InitialBetFee = 5
	econ.Economics.User.MaximumDebtAllowed = 500

	users := []models.User{
		modelstesting.GenerateUser("user1", 950),
		modelstesting.GenerateUser("user2", -100),
	}
	for i := range users {
		if err := db.Create(&users[i]).Error; err != nil {
			t.Fatalf("create user: %v", err)
		}
	}

	market := modelstesting.GenerateMarket(1, users[0].Username)
	market.IsResolved = false
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("create market: %v", err)
	}

	bets := []models.Bet{
		modelstesting.GenerateBet(50, "YES", "user1", uint(market.ID), 0),
		modelstesting.GenerateBet(30, "YES", "user2", uint(market.ID), 0),
	}
	for _, bet := range bets {
		if err := db.Create(&bet).Error; err != nil {
			t.Fatalf("create bet: %v", err)
		}
	}

	svc := NewService(NewGormRepository(db), func() *setup.EconomicConfig { return econ })

	metrics, err := svc.ComputeSystemMetrics(context.Background())
	if err != nil {
		t.Fatalf("ComputeSystemMetrics returned error: %v", err)
	}

	if val, _ := metrics.MoneyCreated.UserDebtCapacity.Value.(int64); val != 1000 {
		t.Errorf("expected user debt capacity 1000, got %d", val)
	}
	if val, _ := metrics.MoneyUtilized.UnusedDebt.Value.(int64); val != 900 {
		t.Errorf("expected unused debt 900, got %d", val)
	}
	if val, _ := metrics.MoneyUtilized.MarketCreationFees.Value.(int64); val != 50 {
		t.Errorf("expected market creation fees 50, got %d", val)
	}
	if val, _ := metrics.MoneyUtilized.ParticipationFees.Value.(int64); val != 10 {
		t.Errorf("expected participation fees 10, got %d", val)
	}
}
