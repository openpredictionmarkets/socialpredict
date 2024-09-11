package betutils

import (
	"errors"
	"socialpredict/models"

	"gorm.io/gorm"
)

func ValidateSale(db *gorm.DB, bet *models.Bet) error {
	var user models.User
	var market models.Market

	// Check if username exists
	if err := db.First(&user, "username = ?", bet.Username).Error; err != nil {
		return errors.New("invalid username")
	}

	// Check if market exists and is open
	if err := db.First(&market, "id = ? AND is_resolved = false", bet.MarketID).Error; err != nil {
		return errors.New("invalid or closed market")
	}

	// Check for valid amount: it should be less than or equal to -1
	if bet.Amount > -1 {
		return errors.New("Sale amount must be greater than or equal to 1")
	}

	// Validate bet outcome: it should be either 'YES' or 'NO'
	if bet.Outcome != "YES" && bet.Outcome != "NO" {
		return errors.New("bet outcome must be 'YES' or 'NO'")
	}

	return nil
}
