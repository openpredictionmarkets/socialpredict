package analytics

// FinancialSnapshot captures a user's financial aggregates.
type FinancialSnapshot struct {
	AccountBalance     int64
	MaximumDebtAllowed int64
	AmountInPlay       int64
	AmountBorrowed     int64
	RetainedEarnings   int64
	Equity             int64
	TradingProfits     int64
	WorkProfits        int64
	TotalProfits       int64

	AmountInPlayActive int64
	TotalSpent         int64
	TotalSpentInPlay   int64
	RealizedProfits    int64
	PotentialProfits   int64
	RealizedValue      int64
	PotentialValue     int64
}

// FinancialSnapshotRequest is the input for computing user financials.
type FinancialSnapshotRequest struct {
	Username       string
	AccountBalance int64
}

// MetricWithExplanation documents metric derivations.
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
