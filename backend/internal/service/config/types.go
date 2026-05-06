package config

import "socialpredict/setup"

const (
	minChartSigFigs     = 2
	maxChartSigFigs     = 9
	defaultChartSigFigs = 4
)

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

type AppConfig struct {
	Economics Economics `yaml:"economics" json:"economics"`
	Frontend  Frontend  `yaml:"frontend" json:"frontend"`
}

// Clone returns a detached copy of the application policy snapshot.
func (cfg *AppConfig) Clone() *AppConfig {
	if cfg == nil {
		return &AppConfig{}
	}

	cfgCopy := *cfg
	return &cfgCopy
}

// FromSetup converts the legacy setup snapshot into the owned config service types.
func FromSetup(cfg *setup.EconomicConfig) *AppConfig {
	if cfg == nil {
		return &AppConfig{}
	}

	return &AppConfig{
		Economics: Economics{
			MarketCreation: MarketCreation{
				InitialMarketProbability:   cfg.Economics.MarketCreation.InitialMarketProbability,
				InitialMarketSubsidization: cfg.Economics.MarketCreation.InitialMarketSubsidization,
				InitialMarketYes:           cfg.Economics.MarketCreation.InitialMarketYes,
				InitialMarketNo:            cfg.Economics.MarketCreation.InitialMarketNo,
				MinimumFutureHours:         cfg.Economics.MarketCreation.MinimumFutureHours,
			},
			MarketIncentives: MarketIncentives{
				CreateMarketCost: cfg.Economics.MarketIncentives.CreateMarketCost,
				TraderBonus:      cfg.Economics.MarketIncentives.TraderBonus,
			},
			User: User{
				InitialAccountBalance: cfg.Economics.User.InitialAccountBalance,
				MaximumDebtAllowed:    cfg.Economics.User.MaximumDebtAllowed,
			},
			Betting: Betting{
				MinimumBet:     cfg.Economics.Betting.MinimumBet,
				MaxDustPerSale: cfg.Economics.Betting.MaxDustPerSale,
				BetFees: BetFees{
					InitialBetFee: cfg.Economics.Betting.BetFees.InitialBetFee,
					BuySharesFee:  cfg.Economics.Betting.BetFees.BuySharesFee,
					SellSharesFee: cfg.Economics.Betting.BetFees.SellSharesFee,
				},
			},
		},
		Frontend: Frontend{
			Charts: FrontendCharts{
				SigFigs: cfg.Frontend.Charts.SigFigs,
			},
		},
	}
}

// ToSetup converts the owned snapshot into the legacy setup shape for unmigrated consumers.
func (cfg *AppConfig) ToSetup() *setup.EconomicConfig {
	if cfg == nil {
		return &setup.EconomicConfig{}
	}

	return &setup.EconomicConfig{
		Economics: setup.Economics{
			MarketCreation: setup.MarketCreation{
				InitialMarketProbability:   cfg.Economics.MarketCreation.InitialMarketProbability,
				InitialMarketSubsidization: cfg.Economics.MarketCreation.InitialMarketSubsidization,
				InitialMarketYes:           cfg.Economics.MarketCreation.InitialMarketYes,
				InitialMarketNo:            cfg.Economics.MarketCreation.InitialMarketNo,
				MinimumFutureHours:         cfg.Economics.MarketCreation.MinimumFutureHours,
			},
			MarketIncentives: setup.MarketIncentives{
				CreateMarketCost: cfg.Economics.MarketIncentives.CreateMarketCost,
				TraderBonus:      cfg.Economics.MarketIncentives.TraderBonus,
			},
			User: setup.User{
				InitialAccountBalance: cfg.Economics.User.InitialAccountBalance,
				MaximumDebtAllowed:    cfg.Economics.User.MaximumDebtAllowed,
			},
			Betting: setup.Betting{
				MinimumBet:     cfg.Economics.Betting.MinimumBet,
				MaxDustPerSale: cfg.Economics.Betting.MaxDustPerSale,
				BetFees: setup.BetFees{
					InitialBetFee: cfg.Economics.Betting.BetFees.InitialBetFee,
					BuySharesFee:  cfg.Economics.Betting.BetFees.BuySharesFee,
					SellSharesFee: cfg.Economics.Betting.BetFees.SellSharesFee,
				},
			},
		},
		Frontend: setup.Frontend{
			Charts: setup.FrontendCharts{
				SigFigs: cfg.Frontend.Charts.SigFigs,
			},
		},
	}
}

// ClampChartSigFigs bounds the frontend chart precision to the supported range.
func ClampChartSigFigs(frontend Frontend) int {
	sigFigs := frontend.Charts.SigFigs
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
