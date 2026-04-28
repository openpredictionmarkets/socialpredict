package setup

import (
	_ "embed"
	"fmt"
	"log"
	"sync"

	"gopkg.in/yaml.v3"
)

//go:embed setup.yaml
var setupYaml []byte

// Source exposes setup asset bytes for explicit configuration seams.
type Source interface {
	Bytes() ([]byte, error)
}

// EmbeddedSource serves the embedded setup asset without exposing package globals to callers.
type EmbeddedSource struct{}

func (EmbeddedSource) Bytes() ([]byte, error) {
	return append([]byte(nil), setupYaml...), nil
}

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

// Clone returns a defensive copy of the legacy startup snapshot.
func (cfg *EconomicConfig) Clone() *EconomicConfig {
	if cfg == nil {
		return &EconomicConfig{}
	}

	cfgCopy := *cfg
	return &cfgCopy
}

var economicConfig *EconomicConfig

const (
	minChartSigFigs     = 2
	maxChartSigFigs     = 9
	defaultChartSigFigs = 4
)

var legacyLoadState struct {
	once sync.Once
	err  error
}

// ParseEconomicConfig decodes the setup asset into the legacy config shape.
func ParseEconomicConfig(data []byte) (*EconomicConfig, error) {
	cfg := &EconomicConfig{}
	if len(data) == 0 {
		return cfg, nil
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// LoadEconomicConfigFromSource loads the legacy config explicitly from the provided asset source.
func LoadEconomicConfigFromSource(source Source) (*EconomicConfig, error) {
	if source == nil {
		return nil, fmt.Errorf("setup source is nil")
	}

	data, err := source.Bytes()
	if err != nil {
		return nil, err
	}

	return ParseEconomicConfig(data)
}

// LoadEconomicsConfig validates the embedded config once and returns a defensive copy.
func LoadEconomicsConfig() (*EconomicConfig, error) {
	legacyLoadState.once.Do(func() {
		economicConfig, legacyLoadState.err = LoadEconomicConfigFromSource(EmbeddedSource{})
	})
	if legacyLoadState.err != nil {
		return nil, legacyLoadState.err
	}
	return economicConfig.Clone(), nil
}

// EconConfigLoader allows functions to use this type as a parameter to load an EconomicConfig Dependency
type EconConfigLoader func() *EconomicConfig

// EconomicsConfig returns a defensive copy of the startup-loaded economics snapshot.
func EconomicsConfig() *EconomicConfig {
	cfg, err := LoadEconomicsConfig()
	if err != nil {
		log.Fatal("Error parsing YAML config:", err) // If the config cannot be loaded, the application cannot recover.
	}
	return cfg
}

// ChartSigFigs returns a clamped significant figures value for chart formatting.
func ChartSigFigs() int {
	cfg := EconomicsConfig()

	sigFigs := cfg.Frontend.Charts.SigFigs
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
