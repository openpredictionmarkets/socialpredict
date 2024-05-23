package seed

import (
	"log"
	"socialpredict/models"
	"socialpredict/setup"
	"time"

	"gorm.io/gorm"
)

func SeedUsers(db *gorm.DB) {

	// load the config constants
	config := setup.LoadEconomicsConfig()

	// Specific time: October 31st, 2023 at 11:59 PM CST
	loc, err := time.LoadLocation("America/Chicago") // CST location
	if err != nil {
		log.Printf("Error loading location: %v", err)
		return
	}
	specificTime := time.Date(2023, time.October, 31, 23, 59, 0, 0, loc)

	// Check if the admin user already exists
	var count int64
	db.Model(&models.User{}).Where("username = ?", "admin").Count(&count)
	if count == 0 {
		// No admin user found, create one
		adminUser := models.User{
			Username:              "admin",
			DisplayName:           "Administrator",
			Email:                 "admin@example.com",
			UserType:              "ADMIN",
			InitialAccountBalance: config.Economics.User.InitialAccountBalance,
			AccountBalance:        config.Economics.User.InitialAccountBalance,
			ApiKey:                "NONE",
			PersonalEmoji:         "NONE",
			Description:           "Administrator",
			MustChangePassword:    false,
		}
		adminUser.HashPassword("password")

		db.Create(&adminUser)
		// Then, update the CreatedAt field for debugging purposes
		db.Model(&adminUser).Update("CreatedAt", specificTime)

	}

}

func SeedMarket(db *gorm.DB) {
	var count int64
	db.Model(&models.Market{}).Count(&count) // Count all markets

	// Specific time: November 1st, 2023 at 11:59 PM CST
	loc, err := time.LoadLocation("America/Chicago") // CST location
	if err != nil {
		log.Printf("Error loading location: %v", err)
		return
	}
	specificTime := time.Date(2023, time.November, 1, 23, 59, 0, 0, loc)

	// load the config constants
	config := setup.LoadEconomicsConfig()
	// Use the config as needed
	initialProbability := config.Economics.MarketCreation.InitialMarketProbability

	if count == 0 {
		// No markets found, create a couple
		market1 := models.Market{
			// ... initialize the market fields ...
			QuestionTitle:      "Will Atlantis Invade Aqua City by the End of 2027?",
			Description:        "This is a sample market description.",
			OutcomeType:        "Binary",
			ResolutionDateTime: time.Now().AddDate(0, 1, 0), // e.g., one month from now
			UTCOffset:          0,
			IsResolved:         false,
			InitialProbability: initialProbability,
			CreatorUsername:    "user1",
		}

		result := db.Create(&market1)
		if result.Error != nil {
			log.Printf("Error seeding market: %v", result.Error)
		}

		// Then, update the CreatedAt field for debugging purposes
		db.Model(&market1).Update("CreatedAt", specificTime)

		market2 := models.Market{
			// ... initialize the market fields ...
			QuestionTitle:      "Will Humans Harvest Anything from Ancient DNA >1MYA by 2030?",
			Description:        "This is a sample market description.",
			OutcomeType:        "Binary",
			ResolutionDateTime: time.Now().AddDate(0, 1, 0), // e.g., one month from now
			UTCOffset:          0,
			IsResolved:         false,
			InitialProbability: 0.5,
			CreatorUsername:    "user1",
		}

		result = db.Create(&market2)
		if result.Error != nil {
			log.Printf("Error seeding market: %v", result.Error)
		}

		// Then, update the CreatedAt field for debugging purposes
		db.Model(&market2).Update("CreatedAt", specificTime)

	}
}
