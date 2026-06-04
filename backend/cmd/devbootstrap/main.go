package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	appruntime "socialpredict/internal/app/runtime"
	"socialpredict/models"
	"socialpredict/setup"

	"gorm.io/gorm"
)

const (
	defaultPassword  = "password"
	defaultUserCount = 10
	defaultPrefix    = "testuser"
)

type bootstrapUser struct {
	username    string
	displayName string
	email       string
	apiKey      string
	userType    string
	emoji       string
	description string
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "dev bootstrap failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	appEnv := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
	if appEnv != "development" {
		return fmt.Errorf("refusing to run outside APP_ENV=development; current APP_ENV=%q", appEnv)
	}

	password := envString("DEV_BOOTSTRAP_PASSWORD", defaultPassword)
	count, err := envInt("DEV_BOOTSTRAP_USER_COUNT", defaultUserCount)
	if err != nil {
		return err
	}
	prefix := envString("DEV_BOOTSTRAP_USER_PREFIX", defaultPrefix)

	dbCfg, err := appruntime.LoadDBConfigFromEnv()
	if err != nil {
		return err
	}
	db, err := appruntime.InitDB(dbCfg, appruntime.PostgresFactory{})
	if err != nil {
		return err
	}
	defer func() {
		_ = appruntime.CloseDB(db)
	}()

	config, err := appruntime.LoadConfigService(setup.EmbeddedSource{})
	if err != nil {
		return err
	}
	initialBalance := config.Economics().User.InitialAccountBalance

	users := []bootstrapUser{
		{
			username:    "admin",
			displayName: "Dev Admin",
			email:       "admin+dev@example.com",
			apiKey:      "dev-admin-api-key",
			userType:    "ADMIN",
			emoji:       "NONE",
			description: "Development admin user",
		},
	}
	for i := 1; i <= count; i++ {
		username := fmt.Sprintf("%s%02d", prefix, i)
		users = append(users, bootstrapUser{
			username:    username,
			displayName: fmt.Sprintf("Dev %s User %02d", prefix, i),
			email:       fmt.Sprintf("%s%02d@example.com", prefix, i),
			apiKey:      fmt.Sprintf("dev-%s%02d-api-key", prefix, i),
			userType:    "REGULAR",
			emoji:       "😀",
			description: "Development test user",
		})
	}

	for _, user := range users {
		if err := upsertBootstrapUser(db, user, password, initialBalance); err != nil {
			return err
		}
	}

	fmt.Printf("Development bootstrap complete.\n")
	fmt.Printf("Password: %s\n", password)
	fmt.Printf("Admin: admin\n")
	fmt.Printf("Users: %s01 through %s%02d\n", prefix, prefix, count)
	fmt.Printf("MustChangePassword: false\n")
	return nil
}

func upsertBootstrapUser(db *gorm.DB, seed bootstrapUser, password string, initialBalance int64) error {
	user := models.User{
		PublicUser: models.PublicUser{
			Username:              seed.username,
			DisplayName:           seed.displayName,
			UserType:              seed.userType,
			InitialAccountBalance: initialBalance,
			AccountBalance:        initialBalance,
			PersonalEmoji:         seed.emoji,
			Description:           seed.description,
		},
		PrivateUser: models.PrivateUser{
			Email:  seed.email,
			APIKey: seed.apiKey,
		},
		ModeratorGovernance: models.ModeratorGovernance{
			ModeratorStatus: "none",
		},
		MustChangePassword: false,
	}
	if err := user.HashPassword(password); err != nil {
		return fmt.Errorf("hash password for %s: %w", seed.username, err)
	}

	var existing models.User
	err := db.Where("username = ?", seed.username).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := db.Create(&user).Error; err != nil {
			return fmt.Errorf("create %s: %w", seed.username, err)
		}
		if err := db.Model(&models.User{}).Where("username = ?", seed.username).Update("must_change_password", false).Error; err != nil {
			return fmt.Errorf("clear password-change flag for %s: %w", seed.username, err)
		}
		fmt.Printf("created %s (%s)\n", seed.username, seed.userType)
		return nil
	}
	if err != nil {
		return fmt.Errorf("load %s: %w", seed.username, err)
	}

	updates := map[string]any{
		"display_name":            seed.displayName,
		"user_type":               seed.userType,
		"initial_account_balance": initialBalance,
		"account_balance":         initialBalance,
		"personal_emoji":          seed.emoji,
		"description":             seed.description,
		"email":                   seed.email,
		"api_key":                 seed.apiKey,
		"password":                user.Password,
		"must_change_password":    false,
		"moderator_status":        "none",
	}
	if err := db.Model(&models.User{}).Where("username = ?", seed.username).Updates(updates).Error; err != nil {
		return fmt.Errorf("update %s: %w", seed.username, err)
	}
	fmt.Printf("updated %s (%s)\n", seed.username, seed.userType)
	return nil
}

func envString(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envInt(key string, fallback int) (int, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 || parsed > 100 {
		return 0, fmt.Errorf("%s must be an integer between 1 and 100", key)
	}
	return parsed, nil
}
