package main

import (
	"log"
	"net/http"
	"socialpredict/middleware"
	"socialpredict/models"
	"socialpredict/server"
	"socialpredict/setup"
	"socialpredict/util"
	"time"

	"gorm.io/gorm"
)

func main() {

	// Secure routes
	http.Handle("/secure", middleware.Authenticate(http.HandlerFunc(secureEndpoint)))

	// Load environment variables
	err := util.GetEnv()
	if err != nil {
		log.Fatalf("Failed to load environment variables: %v", err)
	}

	// Initialize the database connection
	util.InitDB()

	// Now you can safely use the database connection
	db := util.GetDB()

	// Migrate the database
	migrateDB(db)

	// Seed the admin user
	// seedUsers(db)
	// seedMarket(db)
	// seedBets(db)

	server.Start()
}

func secureEndpoint(w http.ResponseWriter, r *http.Request) {
	// This is a secure endpoint, only accessible if Authenticate middleware passes
}

func migrateDB(db *gorm.DB) {
	// Migrate the User and Market models first
	err := db.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatalf("Error migrating User and Market models: %v", err)
	}

	// Then, migrate the Bet model
	err = db.AutoMigrate(&models.Market{})
	if err != nil {
		log.Fatalf("Error migrating Bet model: %v", err)
	}

	// Then, migrate the Bet model
	err = db.AutoMigrate(&models.Bet{})
	if err != nil {
		log.Fatalf("Error migrating Bet model: %v", err)
	}
}

func seedUsers(db *gorm.DB) {

	// load the config constants
	config := setup.LoadEconomicsConfig()
	// Use the config as needed
	initialAccountBalance := config.Economics.User.InitialAccountBalance

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
			Username:    "admin",
			DisplayName: "Administrator",
			Email:       "admin@example.com",
			UserType:    "ADMIN",
			ApiKey:      util.GenerateUniqueApiKey(),
		}
		adminUser.HashPassword("password") // Always use a strong, hashed password

		db.Create(&adminUser)
		// Then, update the CreatedAt field for debugging purposes
		db.Model(&adminUser).Update("CreatedAt", specificTime)

	}

	// Check to see if user1 already exists
	db.Model(&models.User{}).Where("username = ?", "user1").Count(&count)
	if count == 0 {
		// No user1 user found, create one
		user1 := models.User{
			Username:              "user1",
			DisplayName:           "Eegabeep",
			Email:                 "eegabeep@example.com",
			UserType:              "REGULAR",
			InitialAccountBalance: initialAccountBalance,
			ApiKey:                util.GenerateUniqueApiKey(),
			PersonalEmoji:         "😅",
			Description:           "I like predicting things.",
			PersonalLink1:         "https://www.google.com",
		}
		user1.HashPassword("password") // Always use a strong, hashed password

		db.Create(&user1)
		// Then, update the CreatedAt field for debugging purposes
		db.Model(&user1).Update("CreatedAt", specificTime)
	}

	// Check to see if user2 already exists
	db.Model(&models.User{}).Where("username = ?", "user2").Count(&count)
	if count == 0 {
		// No user2 user found, create one
		user2 := models.User{
			Username:              "user2",
			DisplayName:           "Boom Bam",
			Email:                 "BoomBam@example.com",
			UserType:              "REGULAR",
			InitialAccountBalance: initialAccountBalance,
			ApiKey:                util.GenerateUniqueApiKey(),
			PersonalEmoji:         "😅",
			Description:           "Just a typical user. Like to predict stuff.",
			PersonalLink1:         "https://www.duckduckgo.com",
		}
		user2.HashPassword("password") // Always use a strong, hashed password

		db.Create(&user2)
		// Then, update the CreatedAt field for debugging purposes
		db.Model(&user2).Update("CreatedAt", specificTime)

	}
}

func seedMarket(db *gorm.DB) {
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

func seedBets(db *gorm.DB) {
	// Define the initial time for the bets
	loc, err := time.LoadLocation("America/Chicago") // CST location
	if err != nil {
		log.Printf("Error loading location: %v", err)
		return
	}
	initialTime := time.Date(2023, time.November, 10, 23, 59, 0, 0, loc)

	// User IDs and Market IDs (assuming they exist in your database)
	userNames := []string{"user1", "user2"} // User1 and User2
	marketIDs := []uint{1, 2}               // Market1 and Market2
	outcomes := []string{"YES", "NO"}

	// Initialize betTime with the initial time
	betTime := initialTime

	for _, userName := range userNames {
		for _, marketID := range marketIDs {
			for _, outcome := range outcomes {
				// Set the amount based on the outcome
				var amount float64
				if outcome == "YES" {
					amount = 20
				} else {
					amount = 10
				}

				bet := models.Bet{
					Username: userName,
					MarketID: marketID,
					Amount:   amount,
					Outcome:  outcome,
					PlacedAt: betTime,
				}

				result := db.Create(&bet)
				if result.Error != nil {
					log.Printf("Error seeding bet: %v", result.Error)
				}

				// Increment betTime by 15 minutes for the next bet
				betTime = betTime.Add(15 * time.Minute)
			}
		}
	}
}
