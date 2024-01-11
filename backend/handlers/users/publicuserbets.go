package handlers

import (
	"errors"
	"log"
	"socialpredict/models"

	"gorm.io/gorm"
)

// UserBets retrieves all bets made by a specific user, identified by username.
func GetPublicUserBets(db *gorm.DB, username string) ([]models.Bet, error) {
	publicUserInfo := GetPublicUserInfo(db, username)
	if publicUserInfo.Username == "" {
		log.Printf("User not found for username: %v", username)
		return nil, errors.New("user not found")
	}

	var bets []models.Bet
	err := db.Where("username = ?", publicUserInfo.Username).Order("placed_at DESC").Find(&bets).Error
	if err != nil {
		log.Printf("Error fetching bets for user %s: %v", publicUserInfo.Username, err)
		return nil, err
	}

	return bets, nil
}
