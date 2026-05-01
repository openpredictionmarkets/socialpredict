package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	appruntime "socialpredict/internal/app/runtime"
	configsvc "socialpredict/internal/service/config"
	"socialpredict/migration"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestPostgresStartupReadinessAndMigrationPosture(t *testing.T) {
	dsn, ok := startupPostgresIntegrationDSN()
	if !ok {
		t.Skip("set SOCIALPREDICT_POSTGRES_TEST_DSN or POSTGRES_TEST_DSN to run real-Postgres startup readiness verification")
	}

	db := openIsolatedPostgresSchema(t, dsn)
	readiness := appruntime.NewReadiness()
	probe := appruntime.NewServingProbe(db, readiness)

	if err := probe.Ready(context.Background()); err == nil {
		t.Fatalf("expected readiness to stay closed before startup mutations complete")
	}

	migration.ClearRegistry()
	t.Cleanup(migration.ClearRegistry)

	migrationErr := errors.New("intentional postgres migration failure")
	registerMigration(t, "20260501000100", func(*gorm.DB) error {
		return migrationErr
	})

	var seedsRan bool
	err := runStartupMutations(db, configsvc.NewStaticService(nil), appruntime.StartupMutationMode{Writer: true}, startupMutationHooks{
		migrate: migration.MigrateDB,
		seedUsers: func(*gorm.DB, configsvc.Service) error {
			seedsRan = true
			return nil
		},
		seedHomepage: func(*gorm.DB, string) error {
			seedsRan = true
			return nil
		},
	})
	if !errors.Is(err, migrationErr) {
		t.Fatalf("expected startup writer to return migration failure, got %v", err)
	}
	if seedsRan {
		t.Fatalf("expected startup writer to stop before seeds after migration failure")
	}
	if readiness.Ready() {
		t.Fatalf("readiness gate must remain closed after migration failure")
	}
	if err := probe.Ready(context.Background()); err == nil {
		t.Fatalf("expected serving readiness to fail after migration failure")
	}
	assertSchemaMigrationMissing(t, db, "20260501000100")

	migration.ClearRegistry()
	var writerMigrationRan bool
	registerMigration(t, "20260501000200", func(db *gorm.DB) error {
		writerMigrationRan = true
		return db.Exec(`CREATE TABLE startup_postgres_probe (id integer PRIMARY KEY, note text NOT NULL)`).Error
	})

	err = runStartupMutations(db, configsvc.NewStaticService(nil), appruntime.StartupMutationMode{Writer: true}, startupMutationHooks{
		migrate: migration.MigrateDB,
		seedUsers: func(*gorm.DB, configsvc.Service) error {
			return nil
		},
		seedHomepage: func(*gorm.DB, string) error {
			return nil
		},
	})
	if err != nil {
		t.Fatalf("expected successful startup writer migration, got %v", err)
	}
	if !writerMigrationRan {
		t.Fatalf("expected startup writer to execute unapplied migration")
	}
	readiness.MarkReady()
	if err := probe.Ready(context.Background()); err != nil {
		t.Fatalf("expected readiness to pass after startup success against real Postgres: %v", err)
	}
	assertSchemaMigrationPresent(t, db, "20260501000200")

	migration.ClearRegistry()
	var nonWriterMigrationRan bool
	registerMigration(t, "20260501000200", func(db *gorm.DB) error {
		nonWriterMigrationRan = true
		return db.Exec(`CREATE TABLE startup_postgres_probe_nonwriter_unexpected (id integer PRIMARY KEY)`).Error
	})

	err = runStartupMutations(db, configsvc.NewStaticService(nil), appruntime.StartupMutationMode{}, startupMutationHooks{
		verify: migration.VerifyApplied,
	})
	if err != nil {
		t.Fatalf("expected non-writer to verify already-applied migrations, got %v", err)
	}
	if nonWriterMigrationRan {
		t.Fatalf("non-writer must verify applied migrations without executing migration bodies")
	}

	migration.ClearRegistry()
	registerMigration(t, "20260501000300", func(db *gorm.DB) error {
		return db.Exec(`CREATE TABLE startup_postgres_probe_nonwriter_missing (id integer PRIMARY KEY)`).Error
	})

	err = runStartupMutations(db, configsvc.NewStaticService(nil), appruntime.StartupMutationMode{}, startupMutationHooks{
		verify: migration.VerifyApplied,
	})
	if err == nil {
		t.Fatalf("expected non-writer startup to fail when a registered migration is missing")
	}
	if db.Migrator().HasTable("startup_postgres_probe_nonwriter_missing") {
		t.Fatalf("non-writer verification must not create missing migration tables")
	}
}

func openIsolatedPostgresSchema(t *testing.T, dsn string) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open postgres test database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("sql db handle: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	schema := fmt.Sprintf("sp_startup_it_%d", time.Now().UnixNano())
	if err := db.Exec(`CREATE SCHEMA ` + quotePostgresIdentifier(schema)).Error; err != nil {
		t.Fatalf("create isolated postgres schema: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Exec(`DROP SCHEMA IF EXISTS ` + quotePostgresIdentifier(schema) + ` CASCADE`).Error
		_ = sqlDB.Close()
	})
	if err := db.Exec(`SET search_path TO ` + quotePostgresIdentifier(schema)).Error; err != nil {
		t.Fatalf("set isolated postgres schema search_path: %v", err)
	}
	if db.Dialector.Name() != "postgres" {
		t.Fatalf("expected postgres dialector, got %q", db.Dialector.Name())
	}

	return db
}

func startupPostgresIntegrationDSN() (string, bool) {
	for _, key := range []string{"SOCIALPREDICT_POSTGRES_TEST_DSN", "POSTGRES_TEST_DSN"} {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value, true
		}
	}
	return "", false
}

func quotePostgresIdentifier(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

func registerMigration(t *testing.T, id string, up func(*gorm.DB) error) {
	t.Helper()

	if err := migration.Register(id, up); err != nil {
		t.Fatalf("register migration %s: %v", id, err)
	}
}

func assertSchemaMigrationMissing(t *testing.T, db *gorm.DB, id string) {
	t.Helper()

	var count int64
	if err := db.Model(&migration.SchemaMigration{}).Where("id = ?", id).Count(&count).Error; err != nil {
		t.Fatalf("count schema migration %s: %v", id, err)
	}
	if count != 0 {
		t.Fatalf("expected schema migration %s to be absent, got count %d", id, count)
	}
}

func assertSchemaMigrationPresent(t *testing.T, db *gorm.DB, id string) {
	t.Helper()

	var count int64
	if err := db.Model(&migration.SchemaMigration{}).Where("id = ?", id).Count(&count).Error; err != nil {
		t.Fatalf("count schema migration %s: %v", id, err)
	}
	if count != 1 {
		t.Fatalf("expected schema migration %s to be present once, got count %d", id, count)
	}
}
