package math

import "gorm.io/gorm"

func CalculateSharesOwned(db *gorm.DB, userID, marketID uint) (float64, error) {
	// Query database to sum up the amounts of active bets for the user in this market
	// Return the total shares owned
}

func CalculateTotalValueOfShares(sharesOwned, marketPrice float64) float64 {
	// Multiply the total shares owned by the current market price per share
	return sharesOwned * marketPrice
}

func ConvertAmountToShares(sellAmount, marketPrice float64) float64 {
	// Divide the sell amount by the market price to get the number of shares
	return sellAmount / marketPrice
}
