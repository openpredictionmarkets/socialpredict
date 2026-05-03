package runtime

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"socialpredict/models/modelstesting"

	"gorm.io/gorm"
)

// These runtime DB tests use fake handles or SQLite-backed modelstesting helpers
// for fast package-local coverage. They prove config, handle ownership, and
// basic ping/close behavior, but real Postgres remains the source of truth for
// post-ready DB loss/recovery, pool/lifetime behavior, SSL posture, and HA
// startup semantics not covered by the WAVE07 startup contract test.

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

func TestLoadDBConfigFromEnvDefaultPoolConfig(t *testing.T) {
	t.Setenv("DB_HOST", "dbhost")
	t.Setenv("POSTGRES_USER", "pguser")
	t.Setenv("POSTGRES_DATABASE", "pgdb")

	cfg, err := LoadDBConfigFromEnv()
	if err != nil {
		t.Fatalf("LoadDBConfigFromEnv returned error: %v", err)
	}

	pool := EffectiveDBPoolConfig(cfg)
	if pool.MaxOpenConns != defaultDBMaxOpenConns {
		t.Fatalf("expected default max open conns 25, got %d", pool.MaxOpenConns)
	}
	if pool.MaxIdleConns != defaultDBMaxIdleConns {
		t.Fatalf("expected default max idle conns 5, got %d", pool.MaxIdleConns)
	}
	if pool.ConnMaxLifetime != defaultDBConnMaxLifetime {
		t.Fatalf("expected default conn max lifetime 30m, got %s", pool.ConnMaxLifetime)
	}
	if pool.ConnMaxIdleTime != defaultDBConnMaxIdleTime {
		t.Fatalf("expected default conn max idle time 5m, got %s", pool.ConnMaxIdleTime)
	}
}

func TestLoadDBConfigFromEnvUsesDefaultPoolConfigForInvalidValues(t *testing.T) {
	t.Setenv("DB_HOST", "dbhost")
	t.Setenv("POSTGRES_USER", "pguser")
	t.Setenv("POSTGRES_DATABASE", "pgdb")
	t.Setenv("DB_MAX_OPEN_CONNS", "not-an-int")
	t.Setenv("DB_MAX_IDLE_CONNS", "invalid")
	t.Setenv("DB_CONN_MAX_LIFETIME", "forever")
	t.Setenv("DB_CONN_MAX_IDLE_TIME", "eventually")

	cfg, err := LoadDBConfigFromEnv()
	if err != nil {
		t.Fatalf("LoadDBConfigFromEnv returned error: %v", err)
	}

	pool := EffectiveDBPoolConfig(cfg)
	if pool.MaxOpenConns != defaultDBMaxOpenConns ||
		pool.MaxIdleConns != defaultDBMaxIdleConns ||
		pool.ConnMaxLifetime != defaultDBConnMaxLifetime ||
		pool.ConnMaxIdleTime != defaultDBConnMaxIdleTime {
		t.Fatalf("expected invalid pool env values to fall back to defaults, got %+v", pool)
	}
}

func TestEffectiveDBPoolConfigNormalizesNegativeValues(t *testing.T) {
	pool := EffectiveDBPoolConfig(DBConfig{
		MaxOpenConns:    -1,
		MaxIdleConns:    -2,
		ConnMaxLifetime: -time.Second,
		ConnMaxIdleTime: -time.Minute,
	})

	if pool.MaxOpenConns != 0 || pool.MaxIdleConns != 0 || pool.ConnMaxLifetime != 0 || pool.ConnMaxIdleTime != 0 {
		t.Fatalf("expected negative pool values to normalize to zero, got %+v", pool)
	}
}

func TestEffectiveDBPoolConfigCapsIdleConnectionsAtMaxOpen(t *testing.T) {
	pool := EffectiveDBPoolConfig(DBConfig{
		MaxOpenConns: 2,
		MaxIdleConns: 5,
	})

	if pool.MaxIdleConns != 2 {
		t.Fatalf("expected idle conns to be capped at max open conns, got %+v", pool)
	}
}

func TestSnapshotDBPoolReportsConfiguredPoolPosture(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	if err := ConfigureDBPool(db, DBConfig{
		MaxOpenConns: 3,
		MaxIdleConns: 1,
	}); err != nil {
		t.Fatalf("ConfigureDBPool returned error: %v", err)
	}

	snapshot := SnapshotDBPool(db)
	if snapshot.MaxOpenConnections != 3 {
		t.Fatalf("expected max open connections 3, got %+v", snapshot)
	}
	if snapshot.OpenConnections < snapshot.InUseConnections {
		t.Fatalf("expected open connections to cover in-use connections, got %+v", snapshot)
	}
}

func TestDBPoolSnapshotFromSQLStatsIncludesSaturationAndWaitLatency(t *testing.T) {
	snapshot := DBPoolSnapshotFromSQLStats(sql.DBStats{
		MaxOpenConnections: 2,
		OpenConnections:    2,
		InUse:              2,
		Idle:               0,
		WaitCount:          4,
		WaitDuration:       7 * time.Millisecond,
		MaxIdleClosed:      3,
		MaxLifetimeClosed:  1,
	})

	if snapshot.MaxOpenConnections != 2 ||
		snapshot.OpenConnections != 2 ||
		snapshot.InUseConnections != 2 ||
		snapshot.IdleConnections != 0 {
		t.Fatalf("expected pool saturation fields to map from sql stats, got %+v", snapshot)
	}
	if snapshot.WaitCount != 4 || snapshot.WaitDurationNanoseconds != (7*time.Millisecond).Nanoseconds() {
		t.Fatalf("expected pool wait latency fields to map from sql stats, got %+v", snapshot)
	}
	if snapshot.MaxIdleClosedConnections != 3 || snapshot.MaxLifetimeClosedConnections != 1 {
		t.Fatalf("expected pool close counters to map from sql stats, got %+v", snapshot)
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
