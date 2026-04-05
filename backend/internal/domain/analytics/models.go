package analytics

// FinancialSnapshotBalanceReader exposes balance-oriented financial snapshot fields.
type FinancialSnapshotBalanceReader interface {
	AccountBalanceValue() int64
	MaximumDebtAllowedValue() int64
	AmountBorrowedValue() int64
	RetainedEarningsValue() int64
	EquityValue() int64
}

// FinancialSnapshotExposureReader exposes position and exposure fields.
type FinancialSnapshotExposureReader interface {
	AmountInPlayValue() int64
	AmountInPlayActiveValue() int64
	TotalSpentValue() int64
	TotalSpentInPlayValue() int64
	RealizedValueValue() int64
	PotentialValueValue() int64
}

// FinancialSnapshotProfitReader exposes profitability-oriented fields.
type FinancialSnapshotProfitReader interface {
	TradingProfitsValue() int64
	WorkProfitsValue() int64
	TotalProfitsValue() int64
	RealizedProfitsValue() int64
	PotentialProfitsValue() int64
}

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

func (s FinancialSnapshot) AccountBalanceValue() int64     { return s.AccountBalance }
func (s FinancialSnapshot) MaximumDebtAllowedValue() int64 { return s.MaximumDebtAllowed }
func (s FinancialSnapshot) AmountBorrowedValue() int64     { return s.AmountBorrowed }
func (s FinancialSnapshot) RetainedEarningsValue() int64   { return s.RetainedEarnings }
func (s FinancialSnapshot) EquityValue() int64             { return s.Equity }
func (s FinancialSnapshot) AmountInPlayValue() int64       { return s.AmountInPlay }
func (s FinancialSnapshot) AmountInPlayActiveValue() int64 { return s.AmountInPlayActive }
func (s FinancialSnapshot) TotalSpentValue() int64         { return s.TotalSpent }
func (s FinancialSnapshot) TotalSpentInPlayValue() int64   { return s.TotalSpentInPlay }
func (s FinancialSnapshot) RealizedValueValue() int64      { return s.RealizedValue }
func (s FinancialSnapshot) PotentialValueValue() int64     { return s.PotentialValue }
func (s FinancialSnapshot) TradingProfitsValue() int64     { return s.TradingProfits }
func (s FinancialSnapshot) WorkProfitsValue() int64        { return s.WorkProfits }
func (s FinancialSnapshot) TotalProfitsValue() int64       { return s.TotalProfits }
func (s FinancialSnapshot) RealizedProfitsValue() int64    { return s.RealizedProfits }
func (s FinancialSnapshot) PotentialProfitsValue() int64   { return s.PotentialProfits }

// FinancialSnapshotRequestReader exposes only the request data needed to compute a snapshot.
type FinancialSnapshotRequestReader interface {
	UsernameValue() string
	AccountBalanceValue() int64
}

// FinancialSnapshotRequest is the input for computing user financials.
type FinancialSnapshotRequest struct {
	Username       string
	AccountBalance int64
}

func (r FinancialSnapshotRequest) UsernameValue() string      { return r.Username }
func (r FinancialSnapshotRequest) AccountBalanceValue() int64 { return r.AccountBalance }

// MetricExplanationReader exposes metric explanation fields without requiring concrete DTO access.
type MetricExplanationReader interface {
	FormulaValue() string
	ExplanationValue() string
}

// MetricExplanation documents how a metric is derived.
type MetricExplanation struct {
	Formula     string `json:"formula,omitempty"`
	Explanation string `json:"explanation"`
}

func (m MetricExplanation) FormulaValue() string     { return m.Formula }
func (m MetricExplanation) ExplanationValue() string { return m.Explanation }

// Int64MetricReader exposes only the integer metric value contract.
type Int64MetricReader interface {
	Int64Value() int64
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

// BoolMetricReader exposes only the boolean metric value contract.
type BoolMetricReader interface {
	BoolValue() bool
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

// MoneyCreatedReader exposes only the created-money values a consumer needs.
type MoneyCreatedReader interface {
	UserDebtCapacityValue() int64
	NumUsersValue() int64
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

// MoneyUtilizedReader exposes only the utilized-money values a consumer needs.
type MoneyUtilizedReader interface {
	UnusedDebtValue() int64
	ActiveBetVolumeValue() int64
	MarketCreationFeesValue() int64
	ParticipationFeesValue() int64
	BonusesPaidValue() int64
	TotalUtilizedValue() int64
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

// VerificationReader exposes only the verification values a consumer needs.
type VerificationReader interface {
	BalancedValue() bool
	SurplusValue() int64
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

// SystemMetricsReader exposes the segmented metric groups without requiring concrete DTO access.
type SystemMetricsReader interface {
	MoneyCreatedMetrics() MoneyCreatedReader
	MoneyUtilizedMetrics() MoneyUtilizedReader
	VerificationMetrics() VerificationReader
}

func (m SystemMetrics) MoneyCreatedMetrics() MoneyCreatedReader   { return m.MoneyCreated }
func (m SystemMetrics) MoneyUtilizedMetrics() MoneyUtilizedReader { return m.MoneyUtilized }
func (m SystemMetrics) VerificationMetrics() VerificationReader   { return m.Verification }

var (
	_ FinancialSnapshotBalanceReader  = FinancialSnapshot{}
	_ FinancialSnapshotExposureReader = FinancialSnapshot{}
	_ FinancialSnapshotProfitReader   = FinancialSnapshot{}
	_ FinancialSnapshotRequestReader  = FinancialSnapshotRequest{}
	_ MetricExplanationReader         = MetricExplanation{}
	_ Int64MetricReader               = Int64Metric{}
	_ BoolMetricReader                = BoolMetric{}
	_ MoneyCreatedReader              = MoneyCreated{}
	_ MoneyUtilizedReader             = MoneyUtilized{}
	_ VerificationReader              = Verification{}
	_ SystemMetricsReader             = SystemMetrics{}
)
