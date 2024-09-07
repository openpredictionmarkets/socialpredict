package setup

import "testing"

// BuildInitialMarketAppConfig builds the MarketCreation portion of the app config for use in tests that require an EconomicConfig
func BuildInitialMarketAppConfig(t *testing.T, probability float64, subsidization, yes, no int64) *EconomicConfig {
	t.Helper()
	return &EconomicConfig{
		Economics: Economics{
			MarketCreation: MarketCreation{
				InitialMarketProbability:   probability,
				InitialMarketSubsidization: subsidization,
				InitialMarketYes:           yes,
				InitialMarketNo:            no,
			},
		},
	}
}
