package analytics

import "fmt"

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

// Int64Value returns the metric as an int64 when the underlying value matches.
func (m MetricWithExplanation) Int64Value() (int64, error) {
	value, ok := m.Value.(int64)
	if !ok {
		return 0, fmt.Errorf("metric value is %T, want int64", m.Value)
	}
	return value, nil
}

// BoolValue returns the metric as a bool when the underlying value matches.
func (m MetricWithExplanation) BoolValue() (bool, error) {
	value, ok := m.Value.(bool)
	if !ok {
		return false, fmt.Errorf("metric value is %T, want bool", m.Value)
	}
	return value, nil
}

type MoneyCreated struct {
	UserDebtCapacity MetricWithExplanation `json:"userDebtCapacity"`
	NumUsers         MetricWithExplanation `json:"numUsers"`
}

func (m MoneyCreated) UserDebtCapacityValue() (int64, error) {
	return m.UserDebtCapacity.Int64Value()
}

func (m MoneyCreated) NumUsersValue() (int64, error) {
	return m.NumUsers.Int64Value()
}

type MoneyUtilized struct {
	UnusedDebt         MetricWithExplanation `json:"unusedDebt"`
	ActiveBetVolume    MetricWithExplanation `json:"activeBetVolume"`
	MarketCreationFees MetricWithExplanation `json:"marketCreationFees"`
	ParticipationFees  MetricWithExplanation `json:"participationFees"`
	BonusesPaid        MetricWithExplanation `json:"bonusesPaid"`
	TotalUtilized      MetricWithExplanation `json:"totalUtilized"`
}

func (m MoneyUtilized) UnusedDebtValue() (int64, error) {
	return m.UnusedDebt.Int64Value()
}

func (m MoneyUtilized) ActiveBetVolumeValue() (int64, error) {
	return m.ActiveBetVolume.Int64Value()
}

func (m MoneyUtilized) MarketCreationFeesValue() (int64, error) {
	return m.MarketCreationFees.Int64Value()
}

func (m MoneyUtilized) ParticipationFeesValue() (int64, error) {
	return m.ParticipationFees.Int64Value()
}

func (m MoneyUtilized) BonusesPaidValue() (int64, error) {
	return m.BonusesPaid.Int64Value()
}

func (m MoneyUtilized) TotalUtilizedValue() (int64, error) {
	return m.TotalUtilized.Int64Value()
}

type Verification struct {
	Balanced MetricWithExplanation `json:"balanced"`
	Surplus  MetricWithExplanation `json:"surplus"`
}

func (v Verification) BalancedValue() (bool, error) {
	return v.Balanced.BoolValue()
}

func (v Verification) SurplusValue() (int64, error) {
	return v.Surplus.Int64Value()
}

type SystemMetrics struct {
	MoneyCreated  MoneyCreated  `json:"moneyCreated"`
	MoneyUtilized MoneyUtilized `json:"moneyUtilized"`
	Verification  Verification  `json:"verification"`
}
