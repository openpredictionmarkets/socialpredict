package runtime

import (
	"context"
	"testing"
	"time"

	"socialpredict/models/modelstesting"

	"gorm.io/gorm"
)

func TestInitDBReturnsExplicitHandleWithoutSharedFallback(t *testing.T) {
	original := GetDB()
	t.Cleanup(func() {
		SetDB(original)
	})

	want := modelstesting.NewFakeDB(t)
	legacy := modelstesting.NewFakeDB(t)
	SetDB(legacy)
	factory := stubFactory{db: want}

	got, err := InitDB(DBConfig{
		Host:            "localhost",
		User:            "postgres",
		Name:            "socialpredict",
		MaxOpenConns:    12,
		MaxIdleConns:    4,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
	}, factory)
	if err != nil {
		t.Fatalf("InitDB returned error: %v", err)
	}

	if got != want {
		t.Fatalf("expected InitDB to return factory db")
	}
	if shared := GetDB(); shared != legacy {
		t.Fatalf("expected InitDB not to replace shared db handle")
	}

	sqlDB, err := got.DB()
	if err != nil {
		t.Fatalf("db.DB: %v", err)
	}
	stats := sqlDB.Stats()
	if stats.MaxOpenConnections != 12 {
		t.Fatalf("expected max open connections to be configured, got %d", stats.MaxOpenConnections)
	}
}

func TestBuildPostgresDSNDefaultsAndValidation(t *testing.T) {
	dsn, err := BuildPostgresDSN(DBConfig{
		Host:     " localhost ",
		User:     " postgres ",
		Password: "secret",
		Name:     " socialpredict ",
	})
	if err != nil {
		t.Fatalf("BuildPostgresDSN returned error: %v", err)
	}

	want := "host=localhost user=postgres password=secret dbname=socialpredict port=5432 sslmode=disable TimeZone=UTC"
	if dsn != want {
		t.Fatalf("unexpected dsn: %q", dsn)
	}

	if _, err := BuildPostgresDSN(DBConfig{}); err == nil {
		t.Fatalf("expected validation error for empty config")
	}
}

func TestBuildPostgresDSNRequiresTLSWhenConfigured(t *testing.T) {
	cfg := DBConfig{
		Host:       "localhost",
		User:       "postgres",
		Name:       "socialpredict",
		SSLMode:    "disable",
		RequireTLS: true,
	}
	if _, err := BuildPostgresDSN(cfg); err == nil {
		t.Fatalf("expected TLS posture validation error")
	}

	cfg.SSLMode = "verify-full"
	if _, err := BuildPostgresDSN(cfg); err != nil {
		t.Fatalf("expected verify-full to satisfy required TLS posture: %v", err)
	}
}

func TestLoadDBConfigFromEnv(t *testing.T) {
	t.Setenv("DB_HOST", "dbhost")
	t.Setenv("POSTGRES_USER", "pguser")
	t.Setenv("POSTGRES_PASSWORD", "pgpass")
	t.Setenv("POSTGRES_DATABASE", "pgdb")
	t.Setenv("POSTGRES_PORT", "6543")
	t.Setenv("DB_SSLMODE", "verify-full")
	t.Setenv("DB_REQUIRE_TLS", "true")
	t.Setenv("DB_MAX_OPEN_CONNS", "11")
	t.Setenv("DB_MAX_IDLE_CONNS", "3")
	t.Setenv("DB_CONN_MAX_LIFETIME", "45m")
	t.Setenv("DB_CONN_MAX_IDLE_TIME", "90s")

	cfg, err := LoadDBConfigFromEnv()
	if err != nil {
		t.Fatalf("LoadDBConfigFromEnv returned error: %v", err)
	}

	if cfg.Host != "dbhost" || cfg.User != "pguser" || cfg.Password != "pgpass" || cfg.Name != "pgdb" || cfg.Port != "6543" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
	if cfg.SSLMode != "verify-full" || cfg.TimeZone != "UTC" || !cfg.RequireTLS {
		t.Fatalf("unexpected ssl/timezone config: %+v", cfg)
	}
	if cfg.MaxOpenConns != 11 || cfg.MaxIdleConns != 3 || cfg.ConnMaxLifetime != 45*time.Minute || cfg.ConnMaxIdleTime != 90*time.Second {
		t.Fatalf("unexpected pool config: %+v", cfg)
	}
}

func TestLoadDBConfigFromEnvProductionRequiresTLSByDefault(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("DB_HOST", "dbhost")
	t.Setenv("POSTGRES_USER", "pguser")
	t.Setenv("POSTGRES_DATABASE", "pgdb")

	cfg, err := LoadDBConfigFromEnv()
	if err != nil {
		t.Fatalf("LoadDBConfigFromEnv returned error: %v", err)
	}
	if !cfg.RequireTLS {
		t.Fatalf("expected production runtime to require TLS by default")
	}
	if _, err := BuildPostgresDSN(cfg); err == nil {
		t.Fatalf("expected default disabled sslmode to fail production TLS posture")
	}
}

func TestCheckDBReadiness(t *testing.T) {
	t.Run("ready", func(t *testing.T) {
		db := modelstesting.NewFakeDB(t)

		if err := CheckDBReadiness(context.Background(), db); err != nil {
			t.Fatalf("CheckDBReadiness returned error: %v", err)
		}
	})

	t.Run("nil db", func(t *testing.T) {
		if err := CheckDBReadiness(context.Background(), nil); err == nil {
			t.Fatal("expected error for nil db")
		}
	})

	t.Run("closed db", func(t *testing.T) {
		db := modelstesting.NewFakeDB(t)
		sqlDB, err := db.DB()
		if err != nil {
			t.Fatalf("db.DB: %v", err)
		}
		if err := sqlDB.Close(); err != nil {
			t.Fatalf("close sql db: %v", err)
		}

		if err := CheckDBReadiness(context.Background(), db); err == nil {
			t.Fatal("expected readiness check to fail for closed db")
		}
	})
}

func TestCloseDB(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	if err := CloseDB(db); err != nil {
		t.Fatalf("CloseDB returned error: %v", err)
	}
	if err := CheckDBReadiness(context.Background(), db); err == nil {
		t.Fatalf("expected closed db to fail readiness")
	}
}

type stubFactory struct {
	db  *gorm.DB
	err error
}

func (s stubFactory) Open(DBConfig) (*gorm.DB, error) {
	return s.db, s.err
}
