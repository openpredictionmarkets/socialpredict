package modelstesting

import (
	"socialpredict/models"

	"gorm.io/gorm"
)

// AdjustUserBalance applies a delta to the specified user's account balance in a transactional manner.
func AdjustUserBalance(db *gorm.DB, username string, delta int64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.Where("username = ?", username).First(&user).Error; err != nil {
			return err
		}
		user.AccountBalance += delta
		return tx.Save(&user).Error
	})
}

// SumAllUserBalances returns the aggregate account balance across all users in the database.
func SumAllUserBalances(db *gorm.DB) (int64, error) {
	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		return 0, err
	}

	var total int64
	for _, user := range users {
		total += user.AccountBalance
	}
	return total, nil
}

// LoadUserBalances returns a map of username to account balance for every user in the database.
func LoadUserBalances(db *gorm.DB) (map[string]int64, error) {
	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		return nil, err
	}

	result := make(map[string]int64, len(users))
	for _, user := range users {
		result[user.Username] = user.AccountBalance
	}
	return result, nil
}
