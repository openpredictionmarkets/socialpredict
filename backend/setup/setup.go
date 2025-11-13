package setup

import (
	_ "embed"
	"log"
	"sync"

	"gopkg.in/yaml.v3"
)

//go:embed setup.yaml
var setupYaml []byte

type MarketCreation struct {
	InitialMarketProbability   float64 `yaml:"initialMarketProbability"`
	InitialMarketSubsidization int64   `yaml:"initialMarketSubsidization"`
	InitialMarketYes           int64   `yaml:"initialMarketYes"`
	InitialMarketNo            int64   `yaml:"initialMarketNo"`
	MinimumFutureHours         float64 `yaml:"minimumFutureHours"`
}

type MarketIncentives struct {
	CreateMarketCost int64 `yaml:"createMarketCost"`
	TraderBonus      int64 `yaml:"traderBonus"`
}

type User struct {
	InitialAccountBalance int64 `yaml:"initialAccountBalance"`
	MaximumDebtAllowed    int64 `yaml:"maximumDebtAllowed"`
}

type BetFees struct {
	InitialBetFee int64 `yaml:"initialBetFee"`
	BuySharesFee  int64 `yaml:"buySharesFee"`
	SellSharesFee int64 `yaml:"sellSharesFee"`
}

type Betting struct {
	MinimumBet     int64   `yaml:"minimumBet"`
	MaxDustPerSale int64   `yaml:"maxDustPerSale"`
	BetFees        BetFees `yaml:"betFees"`
}

type Economics struct {
	MarketCreation   MarketCreation   `yaml:"marketcreation"`
	MarketIncentives MarketIncentives `yaml:"marketincentives"`
	User             User             `yaml:"user"`
	Betting          Betting          `yaml:"betting"`
}

type FrontendCharts struct {
	SigFigs int `yaml:"sigFigs"`
}

type Frontend struct {
	Charts FrontendCharts `yaml:"charts"`
}

type EconomicConfig struct {
	Economics Economics `yaml:"economics"`
	Frontend  Frontend  `yaml:"frontend"`
}

var economicConfig *EconomicConfig

const (
	minChartSigFigs     = 2
	maxChartSigFigs     = 9
	defaultChartSigFigs = 4
)

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

// EconConfigLoader allows functions to use this type as a parameter to load an EconomicConfig Dependency
type EconConfigLoader func() *EconomicConfig

// EconomicsConfig returns the entire config for the applications economics
func EconomicsConfig() *EconomicConfig {
	return economicConfig
}

func mustLoadEconomicsConfig() {
	economicConfig = &EconomicConfig{}
	err := yaml.Unmarshal(setupYaml, economicConfig)
	if err != nil {
		log.Fatal("Error parsing YAML config:", err) // If the config cannot be loaded, the application cannot recover.
	}
}

func init() {
	mustLoadEconomicsConfig()
}

// ChartSigFigs returns a clamped significant figures value for chart formatting.
func ChartSigFigs() int {
	if economicConfig == nil {
		mustLoadEconomicsConfig()
	}

	sigFigs := economicConfig.Frontend.Charts.SigFigs
	if sigFigs == 0 {
		sigFigs = defaultChartSigFigs
	}

	if sigFigs < minChartSigFigs {
		return minChartSigFigs
	}
	if sigFigs > maxChartSigFigs {
		return maxChartSigFigs
	}
	return sigFigs
}
