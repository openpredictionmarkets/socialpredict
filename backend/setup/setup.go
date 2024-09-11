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
	EachBetFee    int64 `yaml:"eachBetFee"`
	SellSharesFee int64 `yaml:"sellSharesFee"`
}

type Betting struct {
	MinimumBet int64   `yaml:"minimumBet"`
	BetFees    BetFees `yaml:"betFees"`
}

type Economics struct {
	MarketCreation   MarketCreation   `yaml:"marketcreation"`
	MarketIncentives MarketIncentives `yaml:"marketincentives"`
	User             User             `yaml:"user"`
	Betting          Betting          `yaml:"betting"`
}

type EconomicConfig struct {
	Economics Economics `yaml:"economics"`
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
