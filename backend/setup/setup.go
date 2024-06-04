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
			InitialMarketSubsidization int64   `yaml:"initialMarketSubsidization"`
			InitialMarketYes           int64   `yaml:"initialMarketYes"`
			InitialMarketNo            int64   `yaml:"initialMarketNo"`
		} `yaml:"marketcreation"`
		MarketIncentives struct {
			CreateMarketCost int64 `yaml:"createMarketCost"`
			TraderBonus      int64 `yaml:"traderBonus"`
		} `yaml:"marketincentives"`
		User struct {
			InitialAccountBalance int64 `yaml:"initialAccountBalance"`
			MaximumDebtAllowed    int64 `yaml:"maximumDebtAllowed"`
		} `yaml:"user"`
		Betting struct {
			MinimumBet int64 `yaml:"minimumBet"`
			BetFees    struct {
				InitialBetFee int64 `yaml:"initialBetFee"`
				EachBetFee    int64 `yaml:"eachBetFee"`
				SellSharesFee int64 `yaml:"sellSharesFee"`
			} `yaml:"betFees"`
		} `yaml:"betting"`
	} `yaml:"economics"`
}

var economicConfig *EconomicConfig

// load once as a singleton pattern
var once sync.Once

func LoadEconomicsConfig() (*EconomicConfig, error) {
	once.Do(func() {
		economicConfig = &EconomicConfig{}
		err := yaml.Unmarshal(setupYaml, economicConfig)
		if err != nil {
			log.Println("Error parsing YAML config:", err) // Log here or just pass the error up
			return
		}
	})
	return economicConfig, nil
}
