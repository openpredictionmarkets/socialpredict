package setup

import (
	_ "embed"
	"log"
	"sync"

	"gopkg.in/yaml.v3"
)

//go:embed setup.yaml
var setupYaml []byte

type EconomicConfig struct {
	Economics struct {
		MarketCreation struct {
			InitialMarketProbability   float64 `yaml:"initialMarketProbability"`
			InitialMarketSubsidization float64 `yaml:"initialMarketSubsidization"`
			CreateMarketCost           float64 `yaml:"createMarketCost"`
			TraderBonus                float64 `yaml:"traderbonus"`
		} `yaml:"marketcreation"`
		User struct {
			InitialAccountBalance float64 `yaml:"initialAccountBalance"`
			MaximumDebtAllowed    float64 `yaml:"maximumDebtAllowed"`
		} `yaml:"user"`
		Betting struct {
			MinimumBet    float64 `yaml:"minimumBet"`
			BetFee        float64 `yaml:"betFee"`
			SellSharesFee float64 `yaml:"sellSharesFee"`
		} `yaml:"betting"`
	} `yaml:"economics"`
}

var economicConfig *EconomicConfig

// load once as a singleton pattern
var once sync.Once

func LoadEconomicsConfig() *EconomicConfig {
	once.Do(func() {
		economicConfig = &EconomicConfig{}
		err := yaml.Unmarshal(setupYaml, economicConfig)
		if err != nil {
			log.Fatalf("Error parsing YAML config: %v", err)
		}
	})
	return economicConfig
}
