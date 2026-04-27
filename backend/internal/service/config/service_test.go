package config

import (
	"errors"
	"testing"

	"socialpredict/setup"
)

func TestNewServiceLoadsCurrentConfig(t *testing.T) {
	cfg := &AppConfig{
		Economics: Economics{
			User: User{MaximumDebtAllowed: 500},
		},
		Frontend: Frontend{
			Charts: FrontendCharts{SigFigs: 7},
		},
	}

	svc, err := NewService(LoaderFunc(func() (*AppConfig, error) {
		return cfg, nil
	}))
	if err != nil {
		t.Fatalf("NewService returned error: %v", err)
	}

	current := svc.Current()
	if current == cfg {
		t.Fatalf("Current should return a defensive copy")
	}
	if got := current.Economics.User.MaximumDebtAllowed; got != 500 {
		t.Fatalf("Current returned wrong maximum debt: got %d", got)
	}
	if got := svc.Economics().User.MaximumDebtAllowed; got != 500 {
		t.Fatalf("Economics returned wrong maximum debt: got %d", got)
	}
	if got := svc.ChartSigFigs(); got != 7 {
		t.Fatalf("ChartSigFigs returned %d, want 7", got)
	}
}

func TestNewStaticServiceClonesInputSnapshot(t *testing.T) {
	cfg := &AppConfig{
		Economics: Economics{
			User: User{MaximumDebtAllowed: 500},
		},
	}

	svc := NewStaticService(cfg)
	cfg.Economics.User.MaximumDebtAllowed = 999

	if got := svc.Economics().User.MaximumDebtAllowed; got != 500 {
		t.Fatalf("Economics returned %d after source mutation, want frozen 500", got)
	}
}

func TestCurrentReturnsDefensiveCopy(t *testing.T) {
	svc := NewStaticService(&AppConfig{
		Economics: Economics{
			User: User{MaximumDebtAllowed: 500},
		},
	})

	current := svc.Current()
	current.Economics.User.MaximumDebtAllowed = 999

	if got := svc.Economics().User.MaximumDebtAllowed; got != 500 {
		t.Fatalf("Economics returned %d after Current mutation, want frozen 500", got)
	}
}

func TestNewServicePropagatesLoaderError(t *testing.T) {
	wantErr := errors.New("boom")

	svc, err := NewService(LoaderFunc(func() (*AppConfig, error) {
		return nil, wantErr
	}))
	if !errors.Is(err, wantErr) {
		t.Fatalf("NewService error = %v, want %v", err, wantErr)
	}
	if svc != nil {
		t.Fatalf("expected nil service on loader error")
	}
}

func TestClampChartSigFigs(t *testing.T) {
	tests := []struct {
		name     string
		frontend Frontend
		want     int
	}{
		{name: "defaults zero", frontend: Frontend{}, want: 4},
		{
			name: "clamps low",
			frontend: Frontend{
				Charts: FrontendCharts{SigFigs: 1},
			},
			want: 2,
		},
		{
			name: "clamps high",
			frontend: Frontend{
				Charts: FrontendCharts{SigFigs: 99},
			},
			want: 9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ClampChartSigFigs(tt.frontend); got != tt.want {
				t.Fatalf("ClampChartSigFigs() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestSetupCompatibilityTranslation(t *testing.T) {
	legacy := &setup.EconomicConfig{
		Economics: setup.Economics{
			MarketCreation: setup.MarketCreation{
				InitialMarketProbability:   0.42,
				InitialMarketSubsidization: 250,
				InitialMarketYes:           5,
				InitialMarketNo:            7,
				MinimumFutureHours:         24,
			},
			MarketIncentives: setup.MarketIncentives{
				CreateMarketCost: 15,
				TraderBonus:      20,
			},
			User: setup.User{
				InitialAccountBalance: 900,
				MaximumDebtAllowed:    300,
			},
			Betting: setup.Betting{
				MinimumBet:     10,
				MaxDustPerSale: 3,
				BetFees: setup.BetFees{
					InitialBetFee: 1,
					BuySharesFee:  2,
					SellSharesFee: 4,
				},
			},
		},
		Frontend: setup.Frontend{
			Charts: setup.FrontendCharts{SigFigs: 6},
		},
	}

	owned := FromSetup(legacy)
	if owned.Economics.User.InitialAccountBalance != 900 {
		t.Fatalf("owned initial account balance = %d, want 900", owned.Economics.User.InitialAccountBalance)
	}
	if owned.Frontend.Charts.SigFigs != 6 {
		t.Fatalf("owned sig figs = %d, want 6", owned.Frontend.Charts.SigFigs)
	}

	roundTrip := owned.ToSetup()
	if roundTrip.Economics.MarketIncentives.TraderBonus != 20 {
		t.Fatalf("round trip trader bonus = %d, want 20", roundTrip.Economics.MarketIncentives.TraderBonus)
	}
	if roundTrip.Economics.Betting.BetFees.SellSharesFee != 4 {
		t.Fatalf("round trip sell shares fee = %d, want 4", roundTrip.Economics.Betting.BetFees.SellSharesFee)
	}
}

func TestNewStaticServiceClonesLegacySnapshot(t *testing.T) {
	legacy := &setup.EconomicConfig{
		Economics: setup.Economics{
			User: setup.User{
				MaximumDebtAllowed: 300,
			},
		},
	}

	svc := NewStaticService(legacy)
	legacy.Economics.User.MaximumDebtAllowed = 700

	if got := svc.Economics().User.MaximumDebtAllowed; got != 300 {
		t.Fatalf("Economics returned %d after legacy mutation, want frozen 300", got)
	}
}
