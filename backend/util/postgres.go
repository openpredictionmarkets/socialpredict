package util

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// DBConfig holds the normalized database configuration.
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
	TimeZone string
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
		Host:     firstNonEmpty(os.Getenv("DB_HOST"), os.Getenv("DBHOST")),
		User:     firstNonEmpty(os.Getenv("POSTGRES_USER"), os.Getenv("DB_USER")),
		Password: firstNonEmpty(os.Getenv("POSTGRES_PASSWORD"), os.Getenv("DB_PASS"), os.Getenv("DB_PASSWORD")),
		Name:     firstNonEmpty(os.Getenv("POSTGRES_DATABASE"), os.Getenv("POSTGRES_DB"), os.Getenv("DB_NAME")),
		Port:     firstNonEmpty(os.Getenv("POSTGRES_PORT"), os.Getenv("DB_PORT"), "5432"),
		SSLMode:  "disable",
		TimeZone: "UTC",
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

	return gorm.Open(postgres.Open(dsn), gormCfg)
}

// InitDB initializes the global DB with the provided factory and config.
func InitDB(cfg DBConfig, factory DBFactory) (*gorm.DB, error) {
	if factory == nil {
		factory = PostgresFactory{}
	}

	db, err := factory.Open(cfg)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	DB = db
	return db, nil
}

// GetDB returns the database connection.
func GetDB() *gorm.DB {
	return DB
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
