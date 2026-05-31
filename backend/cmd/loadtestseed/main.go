package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	appruntime "socialpredict/internal/app/runtime"
	"socialpredict/models"
	"socialpredict/setup"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	defaultUserCount       = 10
	defaultModeratorCount  = 2
	defaultMarketCount     = 5
	defaultHotMarketCount  = 1
	defaultPassword        = "loadtest-password"
	defaultUserPrefix      = "loaduser"
	defaultModeratorPrefix = "loadmod"
	defaultFixtureDir      = "../loadtest/fixtures"
	defaultBcryptCost      = bcrypt.MinCost
	defaultUserBalance     = 1_000_000
)

type config struct {
	AppEnv          string
	Enabled         bool
	AllowProduction bool
	Reset           bool
	UserCount       int
	ModeratorCount  int
	MarketCount     int
	HotMarketCount  int
	Password        string
	UserPrefix      string
	ModeratorPrefix string
	FixtureDir      string
	BcryptCost      int
	UserBalance     int64
	ResolutionDays  int
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "load-test seed failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	if err := validateSafety(cfg); err != nil {
		return err
	}

	dbCfg, err := appruntime.LoadDBConfigFromEnv()
	if err != nil {
		return err
	}
	db, err := appruntime.InitDB(dbCfg, appruntime.PostgresFactory{})
	if err != nil {
		return err
	}
	defer func() { _ = appruntime.CloseDB(db) }()

	configSvc, err := appruntime.LoadConfigService(setup.EmbeddedSource{})
	if err != nil {
		return err
	}
	economics := configSvc.Economics()

	var marketIDs []int64
	if err := db.Transaction(func(tx *gorm.DB) error {
		if cfg.Reset {
			if err := resetFixtures(tx, cfg); err != nil {
				return err
			}
		}
		if err := upsertUsers(tx, cfg); err != nil {
			return err
		}
		ids, err := upsertMarkets(tx, cfg, economics.MarketCreation.InitialMarketProbability)
		if err != nil {
			return err
		}
		marketIDs = ids
		return nil
	}); err != nil {
		return err
	}

	if err := writeFixtures(cfg, marketIDs); err != nil {
		return err
	}

	fmt.Printf("Load-test seed complete.\n")
	fmt.Printf("Users: %d regular + %d moderators\n", cfg.UserCount, cfg.ModeratorCount)
	fmt.Printf("Markets: %d (%d hot)\n", cfg.MarketCount, cfg.HotMarketCount)
	fmt.Printf("Fixtures: %s\n", cfg.FixtureDir)
	fmt.Printf("Password: [REDACTED]\n")
	fmt.Printf("MustChangePassword: false\n")
	return nil
}

func loadConfig() (config, error) {
	cfg := config{
		AppEnv:          strings.ToLower(strings.Trim(strings.TrimSpace(os.Getenv("APP_ENV")), "'\"")),
		Enabled:         envBool("LOAD_TEST_ENABLED", false),
		AllowProduction: envBool("LOAD_TEST_ALLOW_PRODUCTION", false),
		Reset:           envBool("LOAD_TEST_RESET", false),
		Password:        envString("LOAD_TEST_PASSWORD", defaultPassword),
		UserPrefix:      envString("LOAD_TEST_USER_PREFIX", defaultUserPrefix),
		ModeratorPrefix: envString("LOAD_TEST_MODERATOR_PREFIX", defaultModeratorPrefix),
		FixtureDir:      envString("LOAD_TEST_FIXTURE_DIR", defaultFixtureDir),
	}
	var err error
	if cfg.UserCount, err = envIntErr("LOAD_TEST_USER_COUNT", defaultUserCount, 1, 50_000); err != nil {
		return config{}, err
	}
	if cfg.ModeratorCount, err = envIntErr("LOAD_TEST_MODERATOR_COUNT", defaultModeratorCount, 1, 10_000); err != nil {
		return config{}, err
	}
	if cfg.MarketCount, err = envIntErr("LOAD_TEST_MARKET_COUNT", defaultMarketCount, 1, 50_000); err != nil {
		return config{}, err
	}
	if cfg.HotMarketCount, err = envIntErr("LOAD_TEST_HOT_MARKET_COUNT", defaultHotMarketCount, 0, 50_000); err != nil {
		return config{}, err
	}
	if cfg.BcryptCost, err = envIntErr("LOAD_TEST_BCRYPT_COST", defaultBcryptCost, bcrypt.MinCost, bcrypt.MaxCost); err != nil {
		return config{}, err
	}
	if cfg.UserBalance, err = envInt64Err("LOAD_TEST_USER_BALANCE", defaultUserBalance, 1, 1_000_000_000_000); err != nil {
		return config{}, err
	}
	if cfg.ResolutionDays, err = envIntErr("LOAD_TEST_RESOLUTION_DAYS", 30, 1, 3650); err != nil {
		return config{}, err
	}
	if cfg.HotMarketCount > cfg.MarketCount {
		return config{}, fmt.Errorf("LOAD_TEST_HOT_MARKET_COUNT cannot exceed LOAD_TEST_MARKET_COUNT")
	}
	return cfg, nil
}

func validateSafety(cfg config) error {
	if !cfg.Enabled {
		return errors.New("refusing to seed load-test fixtures unless LOAD_TEST_ENABLED=true")
	}
	if cfg.AppEnv == "production" && !cfg.AllowProduction {
		return errors.New("refusing to seed load-test fixtures when APP_ENV=production unless LOAD_TEST_ALLOW_PRODUCTION=true")
	}
	if err := validatePrefix("LOAD_TEST_USER_PREFIX", cfg.UserPrefix); err != nil {
		return err
	}
	if err := validatePrefix("LOAD_TEST_MODERATOR_PREFIX", cfg.ModeratorPrefix); err != nil {
		return err
	}
	if cfg.UserPrefix == cfg.ModeratorPrefix {
		return errors.New("LOAD_TEST_USER_PREFIX and LOAD_TEST_MODERATOR_PREFIX must differ")
	}
	if len(cfg.Password) < 8 {
		return errors.New("LOAD_TEST_PASSWORD must be at least 8 characters")
	}
	return nil
}

func validatePrefix(label, prefix string) error {
	if prefix == "" {
		return fmt.Errorf("%s is required", label)
	}
	for i, r := range prefix {
		if i == 0 && (r < 'a' || r > 'z') {
			return fmt.Errorf("%s must start with a lowercase letter", label)
		}
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')) {
			return fmt.Errorf("%s must contain only lowercase letters and numbers", label)
		}
	}
	return nil
}

func resetFixtures(db *gorm.DB, cfg config) error {
	creatorLike := cfg.ModeratorPrefix + "%"
	userLike := cfg.UserPrefix + "%"
	modLike := cfg.ModeratorPrefix + "%"
	if err := db.Exec("DELETE FROM bets WHERE market_id IN (SELECT id FROM markets WHERE creator_username LIKE ?)", creatorLike).Error; err != nil {
		return fmt.Errorf("delete fixture market bets: %w", err)
	}
	if err := db.Exec("DELETE FROM bets WHERE username LIKE ? OR username LIKE ?", userLike, modLike).Error; err != nil {
		return fmt.Errorf("delete fixture user bets: %w", err)
	}
	if err := db.Unscoped().Where("creator_username LIKE ?", creatorLike).Delete(&models.Market{}).Error; err != nil {
		return fmt.Errorf("delete fixture markets: %w", err)
	}
	if err := db.Unscoped().Where("username LIKE ? OR username LIKE ?", userLike, modLike).Delete(&models.User{}).Error; err != nil {
		return fmt.Errorf("delete fixture users: %w", err)
	}
	return nil
}

func upsertUsers(db *gorm.DB, cfg config) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.Password), cfg.BcryptCost)
	if err != nil {
		return fmt.Errorf("hash load-test password: %w", err)
	}
	passwordHash := string(hash)
	for i := 1; i <= cfg.UserCount; i++ {
		username := fmt.Sprintf("%s%06d", cfg.UserPrefix, i)
		if err := upsertUser(db, seededUser{
			Username:        username,
			DisplayName:     fmt.Sprintf("Load Test User %06d", i),
			Email:           fmt.Sprintf("%s@example.loadtest.local", username),
			APIKey:          fmt.Sprintf("loadtest-api-key-%s", username),
			UserType:        "REGULAR",
			ModeratorStatus: "none",
		}, passwordHash, cfg.UserBalance); err != nil {
			return err
		}
	}
	for i := 1; i <= cfg.ModeratorCount; i++ {
		username := fmt.Sprintf("%s%06d", cfg.ModeratorPrefix, i)
		if err := upsertUser(db, seededUser{
			Username:        username,
			DisplayName:     fmt.Sprintf("Load Test Moderator %06d", i),
			Email:           fmt.Sprintf("%s@example.loadtest.local", username),
			APIKey:          fmt.Sprintf("loadtest-api-key-%s", username),
			UserType:        "MODERATOR",
			ModeratorStatus: "active",
		}, passwordHash, cfg.UserBalance); err != nil {
			return err
		}
	}
	return nil
}

type seededUser struct {
	Username        string
	DisplayName     string
	Email           string
	APIKey          string
	UserType        string
	ModeratorStatus string
}

func upsertUser(db *gorm.DB, seed seededUser, passwordHash string, balance int64) error {
	user := models.User{
		PublicUser: models.PublicUser{
			Username:              seed.Username,
			DisplayName:           seed.DisplayName,
			UserType:              seed.UserType,
			InitialAccountBalance: balance,
			AccountBalance:        balance,
			PersonalEmoji:         "LT",
			Description:           "Load-test fixture user",
		},
		PrivateUser:         models.PrivateUser{Email: seed.Email, APIKey: seed.APIKey, Password: passwordHash},
		ModeratorGovernance: models.ModeratorGovernance{ModeratorStatus: seed.ModeratorStatus},
		MustChangePassword:  false,
	}
	var existing models.User
	err := db.Where("username = ?", seed.Username).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := db.Create(&user).Error; err != nil {
			return fmt.Errorf("create user %s: %w", seed.Username, err)
		}
		if err := db.Model(&models.User{}).Where("username = ?", seed.Username).Update("must_change_password", false).Error; err != nil {
			return fmt.Errorf("clear password-change flag for user %s: %w", seed.Username, err)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("load user %s: %w", seed.Username, err)
	}
	updates := map[string]any{
		"display_name":            seed.DisplayName,
		"user_type":               seed.UserType,
		"initial_account_balance": balance,
		"account_balance":         balance,
		"personal_emoji":          "LT",
		"description":             "Load-test fixture user",
		"email":                   seed.Email,
		"api_key":                 seed.APIKey,
		"password":                passwordHash,
		"must_change_password":    false,
		"moderator_status":        seed.ModeratorStatus,
	}
	if err := db.Model(&models.User{}).Where("username = ?", seed.Username).Updates(updates).Error; err != nil {
		return fmt.Errorf("update user %s: %w", seed.Username, err)
	}
	return nil
}

func upsertMarkets(db *gorm.DB, cfg config, initialProbability float64) ([]int64, error) {
	if initialProbability == 0 {
		initialProbability = 0.5
	}
	now := time.Now().UTC()
	resolution := now.AddDate(0, 0, cfg.ResolutionDays)
	ids := make([]int64, 0, cfg.MarketCount)
	for i := 1; i <= cfg.MarketCount; i++ {
		creator := fmt.Sprintf("%s%06d", cfg.ModeratorPrefix, ((i-1)%cfg.ModeratorCount)+1)
		title := fmt.Sprintf("Load Test Market %06d", i)
		market := models.Market{
			QuestionTitle:      title,
			Description:        fmt.Sprintf("Load-test fixture market %06d generated for API load testing.", i),
			OutcomeType:        "BINARY",
			ResolutionDateTime: resolution,
			UTCOffset:          0,
			IsResolved:         false,
			ResolutionResult:   "",
			InitialProbability: initialProbability,
			YesLabel:           "YES",
			NoLabel:            "NO",
			LifecycleStatus:    "published",
			ApprovedBy:         "loadtestseed",
			ProposalCost:       0,
			CreatorUsername:    creator,
		}
		var existing models.Market
		err := db.Where("question_title = ? AND creator_username = ?", title, creator).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := db.Create(&market).Error; err != nil {
				return nil, fmt.Errorf("create market %s: %w", title, err)
			}
			ids = append(ids, market.ID)
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("load market %s: %w", title, err)
		}
		updates := map[string]any{
			"description":          market.Description,
			"outcome_type":         market.OutcomeType,
			"resolution_date_time": market.ResolutionDateTime,
			"utc_offset":           0,
			"is_resolved":          false,
			"resolution_result":    "",
			"initial_probability":  market.InitialProbability,
			"yes_label":            market.YesLabel,
			"no_label":             market.NoLabel,
			"lifecycle_status":     market.LifecycleStatus,
			"approved_by":          market.ApprovedBy,
			"proposal_cost":        int64(0),
		}
		if err := db.Model(&models.Market{}).Where("id = ?", existing.ID).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("update market %s: %w", title, err)
		}
		ids = append(ids, existing.ID)
	}
	return ids, nil
}

func writeFixtures(cfg config, marketIDs []int64) error {
	if err := os.MkdirAll(cfg.FixtureDir, 0o755); err != nil {
		return err
	}
	if err := writeUsersCSV(filepath.Join(cfg.FixtureDir, "users.csv"), cfg); err != nil {
		return err
	}
	if err := writeMarketsCSV(filepath.Join(cfg.FixtureDir, "markets.csv"), cfg, marketIDs); err != nil {
		return err
	}
	return nil
}

func writeUsersCSV(path string, cfg config) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()
	if err := writer.Write([]string{"username", "password"}); err != nil {
		return err
	}
	for i := 1; i <= cfg.UserCount; i++ {
		if err := writer.Write([]string{fmt.Sprintf("%s%06d", cfg.UserPrefix, i), cfg.Password}); err != nil {
			return err
		}
	}
	return writer.Error()
}

func writeMarketsCSV(path string, cfg config, marketIDs []int64) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()
	if err := writer.Write([]string{"market_id", "kind"}); err != nil {
		return err
	}
	for i, id := range marketIDs {
		kind := "normal"
		if i < cfg.HotMarketCount {
			kind = "hot"
		}
		if err := writer.Write([]string{strconv.FormatInt(id, 10), kind}); err != nil {
			return err
		}
	}
	return writer.Error()
}

func envString(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envBool(key string, fallback bool) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if value == "" {
		return fallback
	}
	return value == "1" || value == "true" || value == "yes" || value == "on"
}

func envIntErr(key string, fallback int, min int, max int) (int, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < min || parsed > max {
		return 0, fmt.Errorf("%s must be an integer between %d and %d", key, min, max)
	}
	return parsed, nil
}

func envInt64Err(key string, fallback int64, min int64, max int64) (int64, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed < min || parsed > max {
		return 0, fmt.Errorf("%s must be an integer between %d and %d", key, min, max)
	}
	return parsed, nil
}
