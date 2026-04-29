package runtime

import (
	"context"
	"testing"

	"socialpredict/models/modelstesting"

	"gorm.io/gorm"
)

func TestInitDBSetsSharedHandle(t *testing.T) {
	original := GetDB()
	t.Cleanup(func() {
		SetDB(original)
	})

	want := modelstesting.NewFakeDB(t)
	factory := stubFactory{db: want}

	got, err := InitDB(DBConfig{Host: "localhost", User: "postgres", Name: "socialpredict"}, factory)
	if err != nil {
		t.Fatalf("InitDB returned error: %v", err)
	}

	if got != want {
		t.Fatalf("expected InitDB to return factory db")
	}
	if shared := GetDB(); shared != want {
		t.Fatalf("expected shared db handle to be set")
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

func TestLoadDBConfigFromEnv(t *testing.T) {
	t.Setenv("DB_HOST", "dbhost")
	t.Setenv("POSTGRES_USER", "pguser")
	t.Setenv("POSTGRES_PASSWORD", "pgpass")
	t.Setenv("POSTGRES_DATABASE", "pgdb")
	t.Setenv("POSTGRES_PORT", "6543")

	cfg, err := LoadDBConfigFromEnv()
	if err != nil {
		t.Fatalf("LoadDBConfigFromEnv returned error: %v", err)
	}

	if cfg.Host != "dbhost" || cfg.User != "pguser" || cfg.Password != "pgpass" || cfg.Name != "pgdb" || cfg.Port != "6543" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
	if cfg.SSLMode != "disable" || cfg.TimeZone != "UTC" {
		t.Fatalf("unexpected defaults: %+v", cfg)
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

type stubFactory struct {
	db  *gorm.DB
	err error
}

func (s stubFactory) Open(DBConfig) (*gorm.DB, error) {
	return s.db, s.err
}
