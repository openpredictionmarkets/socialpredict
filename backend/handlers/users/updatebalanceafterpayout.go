package handlers

import (
	"fmt"
	"socialpredict/models"

	"gorm.io/gorm"
)

// updateUserBalance updates the user's account balance for winnings or refunds
func UpdateUserBalance(username string, amount float64, db *gorm.DB, transactionType string) error {
	var user models.User

	// Retrieve the user from the database using the username
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		return err
	}

	// Adjust the user's account balance
	switch transactionType {
	case "win":
		// Add the winning amount to the user's balance
		user.AccountBalance += amount
	case "refund":
		// Refund the bet amount to the user's balance
		user.AccountBalance += amount
	default:
		// Handle unknown transaction types if necessary
		return fmt.Errorf("unknown transaction type")
	}

	// Save the updated user record
	if err := db.Save(&user).Error; err != nil {
		return err
	}

	return nil
}
