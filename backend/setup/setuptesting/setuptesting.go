// package setuptesting provides testing helpers and support for testing doubles
package setuptesting

import (
	"testing"

	"socialpredict/setup"
)

// BuildInitialMarketAppConfig builds the MarketCreation portion of the app config for use in tests that require an EconomicConfig
func BuildInitialMarketAppConfig(t *testing.T, probability float64, subsidization, yes, no int64) *setup.EconomicConfig {
	t.Helper()
	return &setup.EconomicConfig{
		Economics: setup.Economics{
			MarketCreation: setup.MarketCreation{
				InitialMarketProbability:   probability,
				InitialMarketSubsidization: subsidization,
				InitialMarketYes:           yes,
				InitialMarketNo:            no,
			},
		},
	}
}

func MockEconomicConfig() *setup.EconomicConfig {
	return &setup.EconomicConfig{
		Economics: setup.Economics{
			MarketCreation: setup.MarketCreation{
				InitialMarketProbability:   0.5,
				InitialMarketSubsidization: 10,
				InitialMarketYes:           0,
				InitialMarketNo:            0,
			},
			MarketIncentives: setup.MarketIncentives{
				CreateMarketCost: 10,
				TraderBonus:      1,
			},
			User: setup.User{
				InitialAccountBalance: 1000,
				MaximumDebtAllowed:    500,
			},
			Betting: setup.Betting{
				MinimumBet: 1,
				BetFees: setup.BetFees{
					InitialBetFee: 1,
					EachBetFee:    1,
					SellSharesFee: 0,
				},
			},
		},
	}
}
