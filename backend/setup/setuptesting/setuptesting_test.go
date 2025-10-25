package setuptesting

import (
	"testing"
)

func TestBuildInitialMarketAppConfig(t *testing.T) {
	cfg := BuildInitialMarketAppConfig(t, 0.7, 15, 5, 3)

	if cfg.Economics.MarketCreation.InitialMarketProbability != 0.7 {
		t.Fatalf("expected probability 0.7, got %f", cfg.Economics.MarketCreation.InitialMarketProbability)
	}
	if cfg.Economics.MarketCreation.InitialMarketSubsidization != 15 {
		t.Fatalf("expected subsidization 15, got %d", cfg.Economics.MarketCreation.InitialMarketSubsidization)
	}
	if cfg.Economics.MarketCreation.InitialMarketYes != 5 {
		t.Fatalf("expected initial yes 5, got %d", cfg.Economics.MarketCreation.InitialMarketYes)
	}
	if cfg.Economics.MarketCreation.InitialMarketNo != 3 {
		t.Fatalf("expected initial no 3, got %d", cfg.Economics.MarketCreation.InitialMarketNo)
	}
}

func TestMockEconomicConfig(t *testing.T) {
	cfg := MockEconomicConfig()

	if cfg.Economics.User.MaximumDebtAllowed != 500 {
		t.Fatalf("expected max debt 500, got %d", cfg.Economics.User.MaximumDebtAllowed)
	}

	if cfg.Economics.MarketIncentives.CreateMarketCost != 10 {
		t.Fatalf("expected create market cost 10, got %d", cfg.Economics.MarketIncentives.CreateMarketCost)
	}

	if cfg.Economics.Betting.BetFees.InitialBetFee != 1 {
		t.Fatalf("expected initial bet fee 1, got %d", cfg.Economics.Betting.BetFees.InitialBetFee)
	}
}
