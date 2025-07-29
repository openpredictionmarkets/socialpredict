package usershandlers

import (
	"fmt"
	"socialpredict/models"

	"gorm.io/gorm"
)

const (
	TransactionWin    = "WIN"
	TransactionRefund = "REFUND"
)

// ApplyTransactionToUser credits the user's balance for a specific transaction type (WIN, REFUND, etc.)
func ApplyTransactionToUser(username string, amount int64, db *gorm.DB, transactionType string) error {
	var user models.User

	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		return fmt.Errorf("user lookup failed: %w", err)
	}

	switch transactionType {
	case TransactionWin, TransactionRefund:
		user.AccountBalance += amount
	default:
		return fmt.Errorf("unknown transaction type: %s", transactionType)
	}

	if err := db.Save(&user).Error; err != nil {
		return fmt.Errorf("failed to update user balance: %w", err)
	}

	return nil
}
