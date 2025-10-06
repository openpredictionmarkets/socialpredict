package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"socialpredict/models"
	"socialpredict/seed"
	"socialpredict/util"

	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

// SeedMarkets populates the database with example prediction markets
func SeedMarkets(db *gorm.DB) error {
	fmt.Println("Starting market seeding process...")

	// Check if markets already exist
	var existingCount int64
	db.Model(&models.Market{}).Count(&existingCount)
	
	if existingCount > 0 {
		fmt.Printf("Found %d existing markets in database.\n", existingCount)
		fmt.Print("Do you want to add example markets anyway? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Aborting market seeding.")
			return nil
		}
	}

	// Define example markets
	markets := []models.Market{
		{
			QuestionTitle:           "Will OpenAI release GPT-5 by end of 2025?",
			Description:             "This market resolves to YES if OpenAI officially announces and releases a model specifically named \"GPT-5\" or \"GPT-5.0\" by December 31, 2025, 11:59 PM UTC. The model must be publicly available or in limited beta testing. Early access programs count as release. Renamed versions (like GPT-4.5 Turbo Ultra) do not count unless explicitly called GPT-5.",
			OutcomeType:             "BINARY",
			ResolutionDateTime:      time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
			FinalResolutionDateTime: time.Date(2026, 1, 5, 23, 59, 59, 0, time.UTC),
			UTCOffset:               0,
			IsResolved:              false,
			InitialProbability:      0.35,
			YesLabel:                "RELEASED",
			NoLabel:                 "NOT RELEASED",
			CreatorUsername:         "admin",
		},
		{
			QuestionTitle:           "Will Lionel Messi score 15+ goals in MLS 2025 regular season?",
			Description:             "This market resolves to YES if Lionel Messi scores 15 or more goals during the 2025 MLS regular season for Inter Miami CF. Only goals scored in official MLS regular season matches count. Playoff goals, friendly matches, international matches, and cup competitions do not count. Market resolves when the regular season ends or if Messi reaches 15 goals earlier.",
			OutcomeType:             "BINARY",
			ResolutionDateTime:      time.Date(2025, 11, 15, 23, 59, 59, 0, time.UTC),
			FinalResolutionDateTime: time.Date(2025, 11, 20, 23, 59, 59, 0, time.UTC),
			UTCOffset:               -5,
			IsResolved:              false,
			InitialProbability:      0.62,
			YesLabel:                "15+ GOALS",
			NoLabel:                 "UNDER 15",
			CreatorUsername:         "admin",
		},
		{
			QuestionTitle:           "Will Donald Trump be the Republican nominee for President in 2028?",
			Description:             "This market resolves to YES if Donald Trump is officially nominated as the Republican Party candidate for President of the United States in the 2028 election. The nomination must be confirmed at the Republican National Convention. If Trump does not run, withdraws, or loses the nomination to another candidate, this resolves to NO.",
			OutcomeType:             "BINARY",
			ResolutionDateTime:      time.Date(2028, 8, 31, 23, 59, 59, 0, time.UTC),
			FinalResolutionDateTime: time.Date(2028, 9, 5, 23, 59, 59, 0, time.UTC),
			UTCOffset:               -4,
			IsResolved:              false,
			InitialProbability:      0.45,
			YesLabel:                "TRUMP",
			NoLabel:                 "OTHER GOP",
			CreatorUsername:         "admin",
		},
		{
			QuestionTitle:           "Will Bitcoin price exceed $150,000 by end of 2025?",
			Description:             "This market resolves to YES if Bitcoin (BTC) price reaches or exceeds $150,000 USD on any major exchange (Coinbase, Binance, Kraken, or Bitstamp) at any point before December 31, 2025, 11:59 PM UTC. The price must be sustained for at least 1 hour on the exchange. Flash crashes or technical glitches lasting less than 1 hour do not count.",
			OutcomeType:             "BINARY",  
			ResolutionDateTime:      time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
			FinalResolutionDateTime: time.Date(2026, 1, 2, 23, 59, 59, 0, time.UTC),
			UTCOffset:               0,
			IsResolved:              false,
			InitialProbability:      0.28,
			YesLabel:                "BULL 🚀",
			NoLabel:                 "BEAR 📉",
			CreatorUsername:         "admin",
		},
		{
			QuestionTitle:           "Will Taylor Swift announce a new studio album in 2025?",
			Description:             "This market resolves to YES if Taylor Swift officially announces a new studio album (not re-recording, compilation, or live album) during 2025. The announcement must come from Taylor Swift herself, her official social media accounts, or her official representatives. The album does not need to be released in 2025, only announced.",
			OutcomeType:             "BINARY",
			ResolutionDateTime:      time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
			FinalResolutionDateTime: time.Date(2026, 1, 5, 23, 59, 59, 0, time.UTC),
			UTCOffset:               0,
			IsResolved:              false,
			InitialProbability:      0.72,
			YesLabel:                "NEW ALBUM 🎵",
			NoLabel:                 "NO ALBUM",
			CreatorUsername:         "admin",
		},
		{
			QuestionTitle:           "Will a cure for Type 1 Diabetes receive FDA approval by 2030?",
			Description:             "This market resolves to YES if the FDA approves a treatment that effectively cures Type 1 Diabetes (not just manages it) by December 31, 2030. The treatment must eliminate the need for insulin injections in at least 80% of patients for at least 2 years. Gene therapy, cell therapy, or artificial pancreas systems that meet these criteria count as cures.",
			OutcomeType:             "BINARY",
			ResolutionDateTime:      time.Date(2030, 12, 31, 23, 59, 59, 0, time.UTC),
			FinalResolutionDateTime: time.Date(2031, 1, 7, 23, 59, 59, 0, time.UTC),
			UTCOffset:               0,
			IsResolved:              false,
			InitialProbability:      0.15,
			YesLabel:                "CURE FOUND",
			NoLabel:                 "NO CURE",
			CreatorUsername:         "admin",
		},
		{
			QuestionTitle:           "Will 2025 be the hottest year on record globally?",
			Description:             "This market resolves to YES if 2025 is declared the hottest year on record for global average temperature by NASA GISS, NOAA, or the UK Met Office. At least two of these three organizations must confirm 2025 as the hottest year. The announcement typically comes in January of the following year.",
			OutcomeType:             "BINARY",
			ResolutionDateTime:      time.Date(2026, 2, 28, 23, 59, 59, 0, time.UTC),
			FinalResolutionDateTime: time.Date(2026, 3, 5, 23, 59, 59, 0, time.UTC),
			UTCOffset:               0,
			IsResolved:              false,
			InitialProbability:      0.55,
			YesLabel:                "HOTTEST 🔥",
			NoLabel:                 "COOLER",
			CreatorUsername:         "admin",
		},
		{
			QuestionTitle:           "Will SpaceX successfully land humans on Mars by 2030?",
			Description:             "This market resolves to YES if SpaceX successfully lands at least one human being on the surface of Mars and that person survives the landing by December 31, 2030. The mission must be primarily operated by SpaceX, though partnerships with NASA or other organizations are allowed. The person must be alive for at least 24 hours after landing.",
			OutcomeType:             "BINARY",
			ResolutionDateTime:      time.Date(2030, 12, 31, 23, 59, 59, 0, time.UTC),
			FinalResolutionDateTime: time.Date(2031, 1, 7, 23, 59, 59, 0, time.UTC),
			UTCOffset:               0,
			IsResolved:              false,
			InitialProbability:      0.25,
			YesLabel:                "MARS 🚀",
			NoLabel:                 "EARTH BOUND",
			CreatorUsername:         "admin",
		},
		{
			QuestionTitle:           "Will Tesla stock price exceed $500 per share in 2025?",
			Description:             "This market resolves to YES if Tesla Inc. (TSLA) stock price reaches or exceeds $500.00 per share on NASDAQ during regular trading hours at any point in 2025. The price must be sustained for at least 15 minutes during market hours. After-hours trading and pre-market trading do not count. Stock splits will be adjusted accordingly.",
			OutcomeType:             "BINARY",
			ResolutionDateTime:      time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
			FinalResolutionDateTime: time.Date(2026, 1, 2, 23, 59, 59, 0, time.UTC),
			UTCOffset:               -5,
			IsResolved:              false,
			InitialProbability:      0.38,
			YesLabel:                "MOON 📈",
			NoLabel:                 "GROUNDED",
			CreatorUsername:         "admin",
		},
		{
			QuestionTitle:           "Will Grand Theft Auto 6 be released in 2025?",
			Description:             "This market resolves to YES if Rockstar Games releases Grand Theft Auto 6 (GTA 6) for any gaming platform in 2025. The game must be available for purchase by consumers, not just announced or in beta. Early access programs count as release. The game must be specifically titled \"Grand Theft Auto 6\" or \"Grand Theft Auto VI\".",
			OutcomeType:             "BINARY",
			ResolutionDateTime:      time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
			FinalResolutionDateTime: time.Date(2026, 1, 5, 23, 59, 59, 0, time.UTC),
			UTCOffset:               0,
			IsResolved:              false,
			InitialProbability:      0.42,
			YesLabel:                "RELEASED 🎮",
			NoLabel:                 "DELAYED",
			CreatorUsername:         "admin",
		},
	}

	// Insert markets into database
	fmt.Printf("Inserting %d example markets...\n", len(markets))
	
	for i, market := range markets {
		if err := db.Create(&market).Error; err != nil {
			return fmt.Errorf("failed to create market %d (%s): %v", i+1, market.QuestionTitle, err)
		}
		fmt.Printf("✓ Created market: %s\n", market.QuestionTitle)
	}

	// Count total markets after insertion
	var totalCount int64
	db.Model(&models.Market{}).Count(&totalCount)
	fmt.Printf("\nSuccessfully added %d example markets!\n", len(markets))
	fmt.Printf("Database now contains %d total markets.\n", totalCount)

	return nil
}

// loadEnvFile loads environment variables from .env file
func loadEnvFile() {
	// Get the directory where this script is located
	scriptDir, err := os.Getwd()
	if err != nil {
		log.Printf("Warning: Could not get current directory: %v", err)
		return
	}

	// Look for .env file in the parent directory (project root)
	projectRoot := filepath.Dir(scriptDir)
	envFile := filepath.Join(projectRoot, ".env")

	// Try to load .env file
	if err := godotenv.Load(envFile); err != nil {
		fmt.Printf("Warning: Could not load .env file from %s: %v\n", envFile, err)
		fmt.Println("Proceeding with existing environment variables or defaults...")
	} else {
		fmt.Printf("✓ Loaded environment variables from %s\n", envFile)
	}
}

func main() {
	fmt.Println("SocialPredict Market Seeder")
	fmt.Println("===========================")

	// Load .env file if it exists
	loadEnvFile()

	// Initialize database connection
	util.InitDB()
	db := util.GetDB()

	// Ensure database is ready
	maxAttempts := 5
	if err := seed.EnsureDBReady(db, maxAttempts); err != nil {
		log.Fatalf("Database not ready: %v", err)
	}

	// Check if admin user exists
	var adminCount int64
	db.Model(&models.User{}).Where("username = ?", "admin").Count(&adminCount)
	if adminCount == 0 {
		fmt.Println("Warning: Admin user not found!")
		fmt.Println("The example markets reference 'admin' as creator_username.")
		fmt.Println("Please create an admin user first by running the main application.")
		fmt.Print("Do you want to continue anyway? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Aborting market seeding.")
			os.Exit(1)
		}
	}

	// Seed markets
	if err := SeedMarkets(db); err != nil {
		log.Fatalf("Failed to seed markets: %v", err)
	}

	fmt.Println("\nMarket seeding completed successfully!")
}