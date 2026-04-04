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

// MetricExplanation documents how a metric is derived.
type MetricExplanation struct {
	Formula     string `json:"formula,omitempty"`
	Explanation string `json:"explanation"`
}

// Int64Metric documents an integer metric derivation.
type Int64Metric struct {
	Value int64 `json:"value"`
	MetricExplanation
}

// NewInt64Metric builds a typed integer metric.
func NewInt64Metric(value int64, formula string, explanation string) Int64Metric {
	return Int64Metric{
		Value: value,
		MetricExplanation: MetricExplanation{
			Formula:     formula,
			Explanation: explanation,
		},
	}
}

// Int64Value returns the integer metric value.
func (m Int64Metric) Int64Value() int64 {
	return m.Value
}

// BoolMetric documents a boolean metric derivation.
type BoolMetric struct {
	Value bool `json:"value"`
	MetricExplanation
}

// NewBoolMetric builds a typed boolean metric.
func NewBoolMetric(value bool, explanation string) BoolMetric {
	return BoolMetric{
		Value: value,
		MetricExplanation: MetricExplanation{
			Explanation: explanation,
		},
	}
}

// BoolValue returns the boolean metric value.
func (m BoolMetric) BoolValue() bool {
	return m.Value
}

type MoneyCreated struct {
	UserDebtCapacity Int64Metric `json:"userDebtCapacity"`
	NumUsers         Int64Metric `json:"numUsers"`
}

func (m MoneyCreated) UserDebtCapacityValue() int64 {
	return m.UserDebtCapacity.Int64Value()
}

func (m MoneyCreated) NumUsersValue() int64 {
	return m.NumUsers.Int64Value()
}

type MoneyUtilized struct {
	UnusedDebt         Int64Metric `json:"unusedDebt"`
	ActiveBetVolume    Int64Metric `json:"activeBetVolume"`
	MarketCreationFees Int64Metric `json:"marketCreationFees"`
	ParticipationFees  Int64Metric `json:"participationFees"`
	BonusesPaid        Int64Metric `json:"bonusesPaid"`
	TotalUtilized      Int64Metric `json:"totalUtilized"`
}

func (m MoneyUtilized) UnusedDebtValue() int64 {
	return m.UnusedDebt.Int64Value()
}

func (m MoneyUtilized) ActiveBetVolumeValue() int64 {
	return m.ActiveBetVolume.Int64Value()
}

func (m MoneyUtilized) MarketCreationFeesValue() int64 {
	return m.MarketCreationFees.Int64Value()
}

func (m MoneyUtilized) ParticipationFeesValue() int64 {
	return m.ParticipationFees.Int64Value()
}

func (m MoneyUtilized) BonusesPaidValue() int64 {
	return m.BonusesPaid.Int64Value()
}

func (m MoneyUtilized) TotalUtilizedValue() int64 {
	return m.TotalUtilized.Int64Value()
}

type Verification struct {
	Balanced BoolMetric  `json:"balanced"`
	Surplus  Int64Metric `json:"surplus"`
}

func (v Verification) BalancedValue() bool {
	return v.Balanced.BoolValue()
}

func (v Verification) SurplusValue() int64 {
	return v.Surplus.Int64Value()
}

type SystemMetrics struct {
	MoneyCreated  MoneyCreated  `json:"moneyCreated"`
	MoneyUtilized MoneyUtilized `json:"moneyUtilized"`
	Verification  Verification  `json:"verification"`
}
