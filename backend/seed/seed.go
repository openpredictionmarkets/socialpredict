package seed

import (
	"fmt"
	"log"
	"os"
	"socialpredict/models"
	"socialpredict/setup"
	"time"

	"gorm.io/gorm"
)

func getEnv(key string) (string, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return "", fmt.Errorf("environment variable %s not set", key)
	}
	return value, nil
}

func SeedUsers(db *gorm.DB) {

	config, err := setup.LoadEconomicsConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	adminPassword, err := getEnv("ADMIN_PASSWORD")
	if err != nil {
		log.Fatalf("Error retrieving ADMIN_PASSWORD: %v", err)
	}
	if adminPassword == "" {
		log.Fatalf("ADMIN_PASSWORD is set but empty")
	} else {
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

			adminUser.HashPassword(adminPassword)

			db.Create(&adminUser)

			// Database default for must_change_password is true, the following force resets to false.
			if result := db.Model(&adminUser).Update("must_change_password", false); result.Error != nil {
				log.Printf("Failed to update MustChangePassword: %v", result.Error)
				return
			}
		}
	}

}

func EnsureDBReady(db *gorm.DB, maxAttempts int) error {
	var err error
	for attempts := 1; attempts <= maxAttempts; attempts++ {
		// Attempt to perform a simple operation like pinging the database
		sqlDB, err := db.DB()
		if err != nil {
			log.Printf("Unable to get database/sql DB from GORM DB: %v", err)
			time.Sleep(time.Second * 5)
			continue
		}

		err = sqlDB.Ping()
		if err != nil {
			log.Printf("Failed to connect to the database, attempt %d/%d: %v", attempts, maxAttempts, err)
			time.Sleep(time.Second * 5) // Wait before retrying
			continue
		}

		log.Println("Database is ready.")
		return nil
	}

	return fmt.Errorf("database is not ready after %d attempts: %v", maxAttempts, err)
}
