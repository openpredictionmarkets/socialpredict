package financials

import (
	"errors"

	marketmath "socialpredict/handlers/math/market"
	"socialpredict/handlers/tradingdata"
	"socialpredict/models"
	"socialpredict/setup"

	"gorm.io/gorm"
)

type MetricWithExplanation struct {
	Value       interface{} `json:"value"`
	Formula     string      `json:"formula,omitempty"`
	Explanation string      `json:"explanation"`
}

type MoneyCreated struct {
	UserDebtCapacity MetricWithExplanation `json:"userDebtCapacity"`
	NumUsers         MetricWithExplanation `json:"numUsers"`
}

type MoneyUtilized struct {
	UnusedDebt         MetricWithExplanation `json:"unusedDebt"`
	ActiveBetVolume    MetricWithExplanation `json:"activeBetVolume"`
	MarketCreationFees MetricWithExplanation `json:"marketCreationFees"`
	ParticipationFees  MetricWithExplanation `json:"participationFees"`
	BonusesPaid        MetricWithExplanation `json:"bonusesPaid"`
	TotalUtilized      MetricWithExplanation `json:"totalUtilized"`
}

type Verification struct {
	Balanced MetricWithExplanation `json:"balanced"`
	Surplus  MetricWithExplanation `json:"surplus"`
}

type SystemMetrics struct {
	MoneyCreated  MoneyCreated  `json:"moneyCreated"`
	MoneyUtilized MoneyUtilized `json:"moneyUtilized"`
	Verification  Verification  `json:"verification"`
}

// ComputeSystemMetrics is stateless/read-only and uses existing models only.
func ComputeSystemMetrics(db *gorm.DB, loadEcon setup.EconConfigLoader) (SystemMetrics, error) {
	if db == nil {
		return SystemMetrics{}, errors.New("nil db")
	}
	econ := loadEcon()

	// Users (count, unused debt calculation)
	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		return SystemMetrics{}, err
	}

	var (
		userCount  = int64(len(users))
		unusedDebt int64 // remaining borrowing capacity
	)

	for i := range users {
		balance := users[i].PublicUser.AccountBalance

		// Calculate unused debt capacity for this user
		// Formula: maxDebtAllowed - max(0, -balance)
		usedDebt := int64(0)
		if balance < 0 {
			usedDebt = -balance
		}
		unusedDebt += econ.Economics.User.MaximumDebtAllowed - usedDebt
	}

	// Total debt capacity
	totalDebtCapacity := econ.Economics.User.MaximumDebtAllowed * userCount

	// Markets data
	var markets []models.Market
	if err := db.Find(&markets).Error; err != nil {
		return SystemMetrics{}, err
	}

	// Market creation fees
	marketCreationFees := int64(len(markets)) * econ.Economics.MarketIncentives.CreateMarketCost

	// Active bet volume: sum of unresolved market volumes (pure bet volume only, excludes subsidization)
	var activeBetVolume int64
	for i := range markets {
		if !markets[i].IsResolved {
			bets := tradingdata.GetBetsForMarket(db, uint(markets[i].ID))
			vol := marketmath.GetMarketVolume(bets)
			activeBetVolume += vol
		}
	}

	// Participation fees: first-time user participation per market
	var bets []models.Bet
	if err := db.Order("market_id ASC, placed_at ASC, id ASC").Find(&bets).Error; err != nil {
		return SystemMetrics{}, err
	}

	type userMarket struct {
		marketID uint
		username string
	}
	seen := make(map[userMarket]bool)
	var participationFees int64

	for i := range bets {
		b := bets[i]
		if b.Amount > 0 { // Only count BUY bets for first-time participation
			key := userMarket{marketID: b.MarketID, username: b.Username}
			if !seen[key] {
				participationFees += econ.Economics.Betting.BetFees.InitialBetFee
				seen[key] = true
			}
		}
	}

	// Bonuses (future feature)
	bonusesPaid := int64(0)

	// Total utilized (corrected calculation without moneyInWallets)
	totalUtilized := unusedDebt + activeBetVolume + marketCreationFees + participationFees + bonusesPaid

	// Verification
	surplus := totalDebtCapacity - totalUtilized
	balanced := surplus == 0

	// Build response with embedded documentation
	return SystemMetrics{
		MoneyCreated: MoneyCreated{
			UserDebtCapacity: MetricWithExplanation{
				Value:       totalDebtCapacity,
				Formula:     "numUsers × maxDebtPerUser",
				Explanation: "Total credit capacity made available to all users",
			},
			NumUsers: MetricWithExplanation{
				Value:       userCount,
				Explanation: "Total number of registered users",
			},
		},
		MoneyUtilized: MoneyUtilized{
			UnusedDebt: MetricWithExplanation{
				Value:       unusedDebt,
				Formula:     "Σ(maxDebtPerUser - max(0, -balance))",
				Explanation: "Remaining borrowing capacity available to users",
			},
			ActiveBetVolume: MetricWithExplanation{
				Value:       activeBetVolume,
				Formula:     "Σ(unresolved_market_volumes)",
				Explanation: "Total value of bets currently active in unresolved markets (excludes fees and subsidies)",
			},
			MarketCreationFees: MetricWithExplanation{
				Value:       marketCreationFees,
				Formula:     "number_of_markets × creation_fee_per_market",
				Explanation: "Fees collected from users creating new markets",
			},
			ParticipationFees: MetricWithExplanation{
				Value:       participationFees,
				Formula:     "Σ(first_bet_per_user_per_market × participation_fee)",
				Explanation: "Fees collected from first-time participation in each market",
			},
			BonusesPaid: MetricWithExplanation{
				Value:       bonusesPaid,
				Explanation: "System bonuses paid to users (future feature)",
			},
			TotalUtilized: MetricWithExplanation{
				Value:       totalUtilized,
				Formula:     "unusedDebt + activeBetVolume + marketCreationFees + participationFees + bonusesPaid",
				Explanation: "Total debt capacity that has been utilized across all categories",
			},
		},
		Verification: Verification{
			Balanced: MetricWithExplanation{
				Value:       balanced,
				Explanation: "Whether total created equals total utilized (perfect accounting balance)",
			},
			Surplus: MetricWithExplanation{
				Value:       surplus,
				Formula:     "userDebtCapacity - totalUtilized",
				Explanation: "Positive = unused capacity, Negative = over-utilization (indicates accounting error)",
			},
		},
	}, nil
}
