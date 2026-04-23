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
	InitialMarketProbability   float64 `yaml:"initialMarketProbability" json:"initialMarketProbability"`
	InitialMarketSubsidization int64   `yaml:"initialMarketSubsidization" json:"initialMarketSubsidization"`
	InitialMarketYes           int64   `yaml:"initialMarketYes" json:"initialMarketYes"`
	InitialMarketNo            int64   `yaml:"initialMarketNo" json:"initialMarketNo"`
	MinimumFutureHours         float64 `yaml:"minimumFutureHours" json:"minimumFutureHours"`
}

type MarketIncentives struct {
	CreateMarketCost int64 `yaml:"createMarketCost" json:"createMarketCost"`
	TraderBonus      int64 `yaml:"traderBonus" json:"traderBonus"`
}

type User struct {
	InitialAccountBalance int64 `yaml:"initialAccountBalance" json:"initialAccountBalance"`
	MaximumDebtAllowed    int64 `yaml:"maximumDebtAllowed" json:"maximumDebtAllowed"`
}

type BetFees struct {
	InitialBetFee int64 `yaml:"initialBetFee" json:"initialBetFee"`
	BuySharesFee  int64 `yaml:"buySharesFee" json:"buySharesFee"`
	SellSharesFee int64 `yaml:"sellSharesFee" json:"sellSharesFee"`
}

type Betting struct {
	MinimumBet     int64   `yaml:"minimumBet" json:"minimumBet"`
	MaxDustPerSale int64   `yaml:"maxDustPerSale" json:"maxDustPerSale"`
	BetFees        BetFees `yaml:"betFees" json:"betFees"`
}

type Economics struct {
	MarketCreation   MarketCreation   `yaml:"marketcreation" json:"marketcreation"`
	MarketIncentives MarketIncentives `yaml:"marketincentives" json:"marketincentives"`
	User             User             `yaml:"user" json:"user"`
	Betting          Betting          `yaml:"betting" json:"betting"`
}

type FrontendCharts struct {
	SigFigs int `yaml:"sigFigs" json:"sigFigs"`
}

type Frontend struct {
	Charts FrontendCharts `yaml:"charts" json:"charts"`
}

type EconomicConfig struct {
	Economics Economics `yaml:"economics" json:"economics"`
	Frontend  Frontend  `yaml:"frontend" json:"frontend"`
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
