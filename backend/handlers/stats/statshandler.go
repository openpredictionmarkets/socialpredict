package statshandlers

import (
	"encoding/json"
	"net/http"
	"socialpredict/models"
	"socialpredict/setup"
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
	TotalMoneyIssued           int64 `json:"totalMoneyIssued"` // Sum of all debt/credit issued
	DebtExtendedToRegularUsers int64 `json:"debtExtendedToRegularUsers"`
	AdditionalDebtExtended     int64 `json:"additionalDebtExtended"`
	DebtExtendedToMarketMakers int64 `json:"debtExtendedToMarketMakers"`
}

type RevenueDetails struct {
	TotalRevenue       int64 `json:"totalRevenue"` // Sum of all revenue components
	TransactionFees    int64 `json:"transactionFees"`
	MarketCreationFees int64 `json:"marketCreationFees"`
}

type ExpenditureDetails struct {
	TotalExpenditures int64 `json:"totalExpenditures"` // Sum of all expenditure components
	BonusesPaid       int64 `json:"bonusesPaid"`
	OtherExpenditures int64 `json:"otherExpenditures"` // Placeholder for other potential expenditures
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
	result.Revenue.TotalRevenue = calculateTotalRevenue(db)
	result.Expenditures.TotalExpenditures = calculateTotalExpenditures(db)
	result.FiscalBalance = result.Revenue.TotalRevenue - result.Expenditures.TotalExpenditures
	result.TotalEquity = result.Revenue.TotalRevenue - result.Liabilities // Liabilities are typically zero in this setup

	return result, nil
}

// calculateTotalMoney calculates the total initial money in the system based on the number of regular users.
func calculateTotalMoneyIssued(db *gorm.DB) (int64, error) {

	debtExtendedToUsersOnCreation, err := calculateDebtExtendedToUsersOnCreation(db)
	if err != nil {
		return debtExtendedToUsersOnCreation, err
	}

	// totalMoney will be the sum of all types of debt extended to accounts
	totalMoney := debtExtendedToUsersOnCreation

	return totalMoney, nil
}

func calculateDebtExtendedToUsersOnCreation(db *gorm.DB) (int64, error) {
	// Load economic configuration
	economicConfig, err := setup.LoadEconomicsConfig()
	if err != nil {
		return 0, err
	}

	// Count the number of regular users
	var userCount int64
	if err := db.Model(&models.User{}).Where("user_type = ?", "REGULAR").Count(&userCount).Error; err != nil {
		return 0, err
	}

	// Calculate total money based on the initial account balance and user count
	return economicConfig.Economics.User.MaximumDebtAllowed * userCount, nil

}

// Example placeholders for other calculations
func calculateDebtExtendedToRegularUsers(db *gorm.DB) int64 {
	// Implement actual database query to sum up debt to regular users
	return 0 // Placeholder
}

func calculateAdditionalDebtExtended(db *gorm.DB) int64 {
	// Implement actual database query to sum up additional debt issued
	return 0 // Placeholder
}

func calculateDebtExtendedToMarketMakers(db *gorm.DB) int64 {
	// Implement actual database query to sum up debt to market makers
	return 0 // Placeholder
}

func calculateTotalRevenue(db *gorm.DB) int64 {
	// Sum up all revenue components
	return 0 // Placeholder
}

func calculateTotalExpenditures(db *gorm.DB) int64 {
	// Sum up all expenditure components
	return 0 // Placeholder
}
