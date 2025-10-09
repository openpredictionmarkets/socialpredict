package seed

import (
	"fmt"
	"log"
	"os"
	"socialpredict/models"
	"socialpredict/setup"
	"time"

	"socialpredict/handlers/cms/homepage"

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

	// Use embedded content to avoid filesystem path issues
	var data []byte
	if len(defaultHomeMD) > 0 {
		data = defaultHomeMD
	} else {
		// Fallback only if embedding failed
		data = []byte("# Welcome to BrierFoxForecast\n\nThis is the seeded home page.")
	}

	// Create renderer for sanitization
	renderer := homepage.NewDefaultRenderer()

	// Since the content is pure HTML, treat it as HTML format
	htmlContent := string(data)

	// Sanitize the HTML directly (no markdown conversion needed)
	sanitizedHTML := renderer.SanitizeHTML(htmlContent)

	item := models.HomepageContent{
		Slug:     "home",
		Title:    "Home",
		Format:   "html", // Changed to html since content is pure HTML
		Markdown: "",     // Empty since we're using HTML format
		HTML:     sanitizedHTML,
		Version:  1,
	}

	return db.Create(&item).Error
}
