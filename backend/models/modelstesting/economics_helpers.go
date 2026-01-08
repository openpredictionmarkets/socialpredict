package modelstesting

import (
	"testing"

	"socialpredict/internal/domain/math/probabilities/wpam"
	"socialpredict/setup"
)

// UseStandardTestEconomics replaces the global economics configuration with the standard
// testing values for the duration of the test. The original configuration is restored
// automatically via t.Cleanup.
func UseStandardTestEconomics(t *testing.T) (*setup.EconomicConfig, func() *setup.EconomicConfig) {
	t.Helper()

	econConfig := setup.EconomicsConfig()
	original := econConfig.Economics
	econConfig.Economics = GenerateEconomicConfig().Economics

	t.Cleanup(func() {
		econConfig.Economics = original
	})

	return econConfig, func() *setup.EconomicConfig {
		return econConfig
	}
}

// SeedWPAMFromConfig configures WPAM seeds using the provided economics config.
func SeedWPAMFromConfig(config *setup.EconomicConfig) {
	if config == nil {
		return
	}
	wpam.SetSeeds(wpam.Seeds{
		InitialProbability:     config.Economics.MarketCreation.InitialMarketProbability,
		InitialSubsidization:   config.Economics.MarketCreation.InitialMarketSubsidization,
		InitialYesContribution: config.Economics.MarketCreation.InitialMarketYes,
		InitialNoContribution:  config.Economics.MarketCreation.InitialMarketNo,
	})
}
