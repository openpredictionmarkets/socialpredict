package seed

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
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
				PublicUser: models.PublicUser{
					Username:              "admin",
					DisplayName:           "Administrator",
					UserType:              "ADMIN",
					InitialAccountBalance: config.Economics.User.InitialAccountBalance,
					AccountBalance:        config.Economics.User.InitialAccountBalance,
					PersonalEmoji:         "NONE",
					Description:           "Administrator",
				},
				PrivateUser: models.PrivateUser{
					Email:  "admin@example.com",
					APIKey: "NONE",
				},
				MustChangePassword: true,
			}

			adminUser.HashPassword(adminPassword)

			db.Create(&adminUser)

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

func SeedHomepage(db *gorm.DB, repoRoot string) error {
	var count int64
	if err := db.Model(&models.HomepageContent{}).
		Where("slug = ?", "home").Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	// default seed path lives next to frontend
	mdPath := filepath.Join(repoRoot, "frontend", "src", "content", "home.md")
	data, err := os.ReadFile(mdPath)
	if err != nil {
		// If file missing, seed with a trivial default
		data = []byte("# Welcome to BrierFoxForecast\n\nThis is the seeded home page.")
	}

	item := models.HomepageContent{
		Slug:     "home",
		Title:    "Home",
		Format:   "markdown",
		Markdown: string(data),
		HTML:     "", // rendered later by service
		Version:  1,
	}

	return db.Create(&item).Error
}
