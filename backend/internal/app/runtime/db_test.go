package runtime

import (
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

type stubFactory struct {
	db  *gorm.DB
	err error
}

func (s stubFactory) Open(DBConfig) (*gorm.DB, error) {
	return s.db, s.err
}
