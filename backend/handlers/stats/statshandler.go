package statshandlers

import (
	"encoding/json"
	"net/http"
	"socialpredict/util"

	"gorm.io/gorm"
)

type FinancialStats struct {
	MoneyIssued   MoneyIssuedDetails `json:"moneyIssued"`
	Revenue       RevenueDetails     `json:"revenue"`
	Expenditures  ExpenditureDetails `json:"expenditures"`
	FiscalBalance int64              `json:"fiscalBalance"` // Calculated as Total Revenue - Total Expenditures
	TotalEquity   int64              `json:"totalEquity"`   // Total Assets (including Money Issued) - Liabilities
	Liabilities   int64              `json:"liabilities"`   // Typically zero in this setup
}

type MoneyIssuedDetails struct {
	TotalMoneyIssued           int64 `json:"totalMoneyIssued"`
	DebtExtendedToRegularUsers int64 `json:"debtExtendedToRegularUsers"`
	AdditionalDebtExtended     int64 `json:"additionalDebtExtended"`
	DebtExtendedToMarketMakers int64 `json:"debtExtendedToMarketMakers"`
}

type RevenueDetails struct {
	TotalRevenue       int64 `json:"totalRevenue"`
	TransactionFees    int64 `json:"transactionFees"`
	MarketCreationFees int64 `json:"marketCreationFees"`
}

type ExpenditureDetails struct {
	TotalExpenditures int64 `json:"totalExpenditures"`
	BonusesPaid       int64 `json:"bonusesPaid"`
	OtherExpenditures int64 `json:"otherExpenditures"`
}

// Handle Circulation Details as a separate struct so we can use it as a checksum
// May be computationally intensive
type CirculationDetails struct {
	UnusedCredits            int64 `json:"unusedCredits"`            // Credits on user accounts not currently used
	InvestmentsInMarkets     int64 `json:"investmentsInMarkets"`     // Active investments from trades
	PendingMarketFees        int64 `json:"pendingMarketFees"`        // Initial trading fees pending resolution
	FinalFeesFromTrading     int64 `json:"finalFeesFromTrading"`     // Final fees collected from trading activities
	FeesFromCancelledMarkets int64 `json:"feesFromCancelledMarkets"` // Fees collected from cancelled markets
	BonusesAfterResolution   int64 `json:"bonusesAfterResolution"`   // Bonuses paid out at market resolution
	TotalInCirculation       int64 `json:"totalInCirculation"`       // Should match TotalMoneyIssued if all calculations are correct
}

// StatsHandler handles requests for financial stats
func StatsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		db := util.GetDB()
		// Call the calculateFinancialStats function
		stats, err := calculateFinancialStats(db)
		if err != nil {
			http.Error(w, "Failed to calculate financial stats: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(stats); err != nil {
			http.Error(w, "Failed to encode financial stats: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

func calculateFinancialStats(db *gorm.DB) (FinancialStats, error) {
	var result FinancialStats

	// Assuming calculateTotalMoneyIssued sums up all forms of issued money
	totalMoneyIssued, err := calculateTotalMoneyIssued(db)
	if err != nil {
		return result, err // Handle errors appropriately
	}

	// Fill in the MoneyIssuedDetails struct within FinancialStats
	result.MoneyIssued.TotalMoneyIssued = totalMoneyIssued
	result.MoneyIssued.DebtExtendedToRegularUsers = calculateDebtExtendedToRegularUsers(db)
	result.MoneyIssued.AdditionalDebtExtended = calculateAdditionalDebtExtended(db)
	result.MoneyIssued.DebtExtendedToMarketMakers = calculateDebtExtendedToMarketMakers(db)

	// Calculate and populate other financial details
	// Revenue
	result.Revenue.TotalRevenue = calculateTotalRevenue(db)
	result.Revenue.MarketCreationFees = sumAllMarketCreationFees(db)
	// Expenditures
	result.Expenditures.TotalExpenditures = calculateTotalExpenditures(db)
	// Balance
	result.FiscalBalance = result.Revenue.TotalRevenue - result.Expenditures.TotalExpenditures
	result.TotalEquity = result.Revenue.TotalRevenue - result.Liabilities // Liabilities are typically zero in this setup

	return result, nil
}
