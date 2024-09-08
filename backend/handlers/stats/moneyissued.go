package statshandlers

import (
	"socialpredict/models"
	"socialpredict/setup"

	"gorm.io/gorm"
)

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
