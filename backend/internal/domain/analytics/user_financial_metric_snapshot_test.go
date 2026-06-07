package analytics

import (
	"testing"
	"time"

	positionsmath "socialpredict/internal/domain/math/positions"
)

func TestUserFinancialMetricSnapshotCalculatorMatchesFinancialSnapshotMath(t *testing.T) {
	generatedAt := time.Date(2026, 6, 7, 10, 30, 0, 0, time.UTC)
	req := FinancialSnapshotRequest{
		Username:       "alice",
		AccountBalance: 500,
	}
	positions := []positionsmath.MarketPosition{
		{
			Username:         "alice",
			MarketID:         1,
			Value:            140,
			TotalSpent:       100,
			TotalSpentInPlay: 80,
			IsResolved:       false,
		},
		{
			Username:         "alice",
			MarketID:         2,
			Value:            90,
			TotalSpent:       120,
			TotalSpentInPlay: 0,
			IsResolved:       true,
			ResolutionResult: "YES",
		},
	}

	snapshot := NewUserFinancialMetricSnapshotCalculator(Config{MaximumDebtAllowed: 250}).
		Calculate(req, positions, generatedAt)

	if snapshot.Username != req.Username {
		t.Fatalf("username = %s, want %s", snapshot.Username, req.Username)
	}
	if !snapshot.GeneratedAt.Equal(generatedAt) {
		t.Fatalf("generated at = %s, want %s", snapshot.GeneratedAt, generatedAt)
	}
	if snapshot.PositionCount != len(positions) {
		t.Fatalf("position count = %d, want %d", snapshot.PositionCount, len(positions))
	}
	if snapshot.TransactionSafeRead {
		t.Fatalf("user financial metric snapshot must not be transaction safe")
	}
	if snapshot.Source != "read_model" {
		t.Fatalf("source = %q, want read_model", snapshot.Source)
	}

	financial := snapshot.Financial
	if financial.AccountBalance != 500 ||
		financial.MaximumDebtAllowed != 250 ||
		financial.AmountInPlay != 230 ||
		financial.TotalSpent != 220 ||
		financial.TotalSpentInPlay != 80 ||
		financial.TradingProfits != 10 ||
		financial.PotentialProfits != 40 ||
		financial.RealizedProfits != -30 ||
		financial.PotentialValue != 140 ||
		financial.RealizedValue != 90 ||
		financial.AmountInPlayActive != 140 ||
		financial.RetainedEarnings != 270 ||
		financial.Equity != 500 ||
		financial.TotalProfits != 10 {
		t.Fatalf("unexpected financial snapshot: %+v", financial)
	}
}

func TestUserFinancialMetricSnapshotCalculatorNegativeBalance(t *testing.T) {
	snapshot := NewUserFinancialMetricSnapshotCalculator(Config{MaximumDebtAllowed: 100}).
		Calculate(FinancialSnapshotRequest{Username: "borrower", AccountBalance: -25}, nil, time.Time{})

	if snapshot.Financial.AmountBorrowed != 25 {
		t.Fatalf("amount borrowed = %d, want 25", snapshot.Financial.AmountBorrowed)
	}
	if snapshot.Financial.Equity != -50 {
		t.Fatalf("equity = %d, want -50", snapshot.Financial.Equity)
	}
}
