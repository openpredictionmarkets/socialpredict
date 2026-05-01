package runtime

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var (
	db   *gorm.DB
	dbMu sync.RWMutex
)

// DBConfig holds the normalized database configuration.
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
	TimeZone string

	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	RequireTLS      bool
}

// DBFactory provides a hook to open a database using a given configuration.
type DBFactory interface {
	Open(DBConfig) (*gorm.DB, error)
}

// PostgresFactory implements DBFactory using the postgres driver.
type PostgresFactory struct {
	GormConfig *gorm.Config
}

// LoadDBConfigFromEnv normalizes env vars into a DBConfig.
func LoadDBConfigFromEnv() (DBConfig, error) {
	cfg := DBConfig{
		Host:            firstNonEmpty(os.Getenv("DB_HOST"), os.Getenv("DBHOST")),
		User:            firstNonEmpty(os.Getenv("POSTGRES_USER"), os.Getenv("DB_USER")),
		Password:        firstNonEmpty(os.Getenv("POSTGRES_PASSWORD"), os.Getenv("DB_PASS"), os.Getenv("DB_PASSWORD")),
		Name:            firstNonEmpty(os.Getenv("POSTGRES_DATABASE"), os.Getenv("POSTGRES_DB"), os.Getenv("DB_NAME")),
		Port:            firstNonEmpty(os.Getenv("POSTGRES_PORT"), os.Getenv("DB_PORT"), "5432"),
		SSLMode:         firstNonEmpty(os.Getenv("DB_SSLMODE"), os.Getenv("POSTGRES_SSLMODE"), os.Getenv("PGSSLMODE"), "disable"),
		TimeZone:        firstNonEmpty(os.Getenv("DB_TIMEZONE"), os.Getenv("PGTZ"), "UTC"),
		MaxOpenConns:    intFromEnv(25, "DB_MAX_OPEN_CONNS", "POSTGRES_MAX_OPEN_CONNS"),
		MaxIdleConns:    intFromEnv(5, "DB_MAX_IDLE_CONNS", "POSTGRES_MAX_IDLE_CONNS"),
		ConnMaxLifetime: durationFromEnv(30*time.Minute, "DB_CONN_MAX_LIFETIME", "POSTGRES_CONN_MAX_LIFETIME"),
		ConnMaxIdleTime: durationFromEnv(5*time.Minute, "DB_CONN_MAX_IDLE_TIME", "POSTGRES_CONN_MAX_IDLE_TIME"),
		RequireTLS:      boolFromEnv(isProductionRuntime(), "DB_REQUIRE_TLS", "POSTGRES_REQUIRE_TLS"),
	}

	if cfg.Host == "" {
		return DBConfig{}, fmt.Errorf("missing DB host (DB_HOST or DBHOST)")
	}
	if cfg.User == "" {
		return DBConfig{}, fmt.Errorf("missing DB user (POSTGRES_USER or DB_USER)")
	}
	if cfg.Name == "" {
		return DBConfig{}, fmt.Errorf("missing DB name (POSTGRES_DATABASE/POSTGRES_DB/DB_NAME)")
	}

	return cfg, nil
}

// BuildPostgresDSN assembles the postgres DSN from config.
func BuildPostgresDSN(cfg DBConfig) (string, error) {
	cfg.Host = strings.TrimSpace(cfg.Host)
	cfg.User = strings.TrimSpace(cfg.User)
	cfg.Name = strings.TrimSpace(cfg.Name)
	cfg.Port = strings.TrimSpace(cfg.Port)
	cfg.SSLMode = strings.ToLower(strings.TrimSpace(cfg.SSLMode))
	cfg.TimeZone = strings.TrimSpace(cfg.TimeZone)

	if cfg.Host == "" || cfg.User == "" || cfg.Name == "" {
		return "", fmt.Errorf("invalid DB config: host/user/name required")
	}

	port := cfg.Port
	if port == "" {
		port = "5432"
	}

	sslMode := cfg.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}
	if err := validateSSLMode(sslMode, cfg.RequireTLS); err != nil {
		return "", err
	}

	timeZone := cfg.TimeZone
	if timeZone == "" {
		timeZone = "UTC"
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		cfg.Host,
		cfg.User,
		cfg.Password,
		cfg.Name,
		port,
		sslMode,
		timeZone,
	)

	return dsn, nil
}

// Open opens a postgres-backed gorm DB using the provided configuration.
func (f PostgresFactory) Open(cfg DBConfig) (*gorm.DB, error) {
	dsn, err := BuildPostgresDSN(cfg)
	if err != nil {
		return nil, err
	}

	gormCfg := f.GormConfig
	if gormCfg == nil {
		gormCfg = &gorm.Config{}
	}

	if gormCfg.Logger == nil {
		gormCfg.Logger = newFilteredGormLogger(gormlogger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			gormlogger.Config{
				LogLevel:                  gormlogger.Warn,
				IgnoreRecordNotFoundError: true,
			},
		))
	}

	return gorm.Open(postgres.Open(dsn), gormCfg)
}

// InitDB opens the explicit runtime DB handle with the provided factory and config.
func InitDB(cfg DBConfig, factory DBFactory) (*gorm.DB, error) {
	if factory == nil {
		factory = PostgresFactory{}
	}

	conn, err := factory.Open(cfg)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if conn == nil {
		return nil, fmt.Errorf("open db: factory returned nil db")
	}
	if err := ConfigureDBPool(conn, cfg); err != nil {
		return nil, fmt.Errorf("configure db pool: %w", err)
	}
	return conn, nil
}

// ConfigureDBPool applies runtime-owned connection pool and lifetime policy.
func ConfigureDBPool(db *gorm.DB, cfg DBConfig) error {
	if db == nil {
		return fmt.Errorf("database handle unavailable")
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("sql db handle unavailable: %w", err)
	}

	sqlDB.SetMaxOpenConns(normalizeNonNegative(cfg.MaxOpenConns))
	sqlDB.SetMaxIdleConns(normalizeNonNegative(cfg.MaxIdleConns))
	sqlDB.SetConnMaxLifetime(normalizeNonNegativeDuration(cfg.ConnMaxLifetime))
	sqlDB.SetConnMaxIdleTime(normalizeNonNegativeDuration(cfg.ConnMaxIdleTime))
	return nil
}

// CheckDBReadiness verifies that the backing SQL connection is reachable for request handling.
func CheckDBReadiness(ctx context.Context, db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database handle unavailable")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("sql db handle unavailable: %w", err)
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// CloseDB closes the underlying SQL connection pool during runtime shutdown.
func CloseDB(db *gorm.DB) error {
	if db == nil {
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("sql db handle unavailable: %w", err)
	}
	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("close db: %w", err)
	}
	return nil
}

// SetDB stores a process-global handle for legacy tests and narrow migration bridges.
//
// Production startup should pass the DB handle explicitly from InitDB instead of using
// this fallback.
func SetDB(conn *gorm.DB) {
	dbMu.Lock()
	db = conn
	dbMu.Unlock()
}

// GetDB returns the legacy process-global database connection.
//
// Production startup should pass the DB handle explicitly from InitDB instead of using
// this fallback.
func GetDB() *gorm.DB {
	dbMu.RLock()
	defer dbMu.RUnlock()
	return db
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if trimmed := strings.TrimSpace(v); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func intFromEnv(defaultValue int, keys ...string) int {
	value := firstExistingEnv(keys...)
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}

func durationFromEnv(defaultValue time.Duration, keys ...string) time.Duration {
	value := firstExistingEnv(keys...)
	if value == "" {
		return defaultValue
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}

func boolFromEnv(defaultValue bool, keys ...string) bool {
	value := strings.ToLower(firstExistingEnv(keys...))
	switch value {
	case "":
		return defaultValue
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return defaultValue
	}
}

func firstExistingEnv(keys ...string) string {
	for _, key := range keys {
		if value, ok := os.LookupEnv(key); ok {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func isProductionRuntime() bool {
	env := strings.ToLower(firstNonEmpty(
		os.Getenv("APP_ENV"),
		os.Getenv("APP_ENVIRONMENT"),
		os.Getenv("ENVIRONMENT"),
		os.Getenv("GO_ENV"),
	))
	return env == "prod" || env == "production"
}

func validateSSLMode(sslMode string, requireTLS bool) error {
	switch sslMode {
	case "disable", "allow", "prefer", "require", "verify-ca", "verify-full":
	default:
		return fmt.Errorf("invalid DB sslmode %q", sslMode)
	}
	if requireTLS {
		switch sslMode {
		case "require", "verify-ca", "verify-full":
			return nil
		default:
			return fmt.Errorf("DB sslmode %q does not satisfy required TLS posture", sslMode)
		}
	}
	return nil
}

func normalizeNonNegative(value int) int {
	if value < 0 {
		return 0
	}
	return value
}

func normalizeNonNegativeDuration(value time.Duration) time.Duration {
	if value < 0 {
		return 0
	}
	return value
}
