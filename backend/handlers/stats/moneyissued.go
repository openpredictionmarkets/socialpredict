package statshandlers

import (
	"socialpredict/repository"
	"socialpredict/setup"

	"gorm.io/gorm"
)

// calculateTotalMoney calculates the total initial money in the system based on the number of regular users.
func calculateTotalMoneyIssued(db *gorm.DB) (int64, error) {

	debtExtendedToUsersOnCreation := calculateDebtExtendedToRegularUsers(db)

	// totalMoney will be the sum of all types of debt extended to accounts
	totalMoney := debtExtendedToUsersOnCreation

	return totalMoney, nil
}

// Example placeholders for other calculations
func calculateDebtExtendedToRegularUsers(db *gorm.DB) int64 {
	// Implement actual database query to sum up debt to regular users
	// Catchall for any type of universal debt besides debt upon creation

	debtExtendedToRegularUsersOnCreation, err := calculateDebtExtendedToRegularUsersOnCreation(db)
	if err != nil {
		return 0
	}

	totalRegularUserDebt := debtExtendedToRegularUsersOnCreation

	return totalRegularUserDebt
}

func calculateDebtExtendedToRegularUsersOnCreation(db *gorm.DB) (int64, error) {

	economicConfig, err := setup.LoadEconomicsConfig()
	if err != nil {
		return 0, err
	}

	// Count the number of regular users
	var userCount int64

	gormDatabase := &repository.GormDatabase{DB: db}

	userRepo := repository.NewUserRepository(gormDatabase)
	userCount, err = userRepo.CountRegularUsers()
	if err != nil {
		return 0, err
	}

	// Calculate total money based on the initial account balance and user count
	return economicConfig.Economics.User.MaximumDebtAllowed * userCount, nil

}

func calculateAdditionalDebtExtended(db *gorm.DB) int64 {
	// Implement actual database query to sum up additional debt issued
	return 0 // Placeholder
}

func calculateDebtExtendedToMarketMakers(db *gorm.DB) int64 {
	// Implement actual database query to sum up debt to market makers
	return 0 // Placeholder
}
