package migration

import (
	"fmt"
	"sort"
	"time"

	"gorm.io/gorm"
	"socialpredict/logger"
)

// Registry of migrations; you already have tests that exercise this.
var registry = map[string]func(*gorm.DB) error{}

type SchemaMigration struct {
	ID        string    `gorm:"primaryKey;size:32"`
	AppliedAt time.Time `gorm:"autoCreateTime"`
}

func Register(id string, up func(*gorm.DB) error) error {
	if _, exists := registry[id]; exists {
		return fmt.Errorf("duplicate migration id: %s", id)
	}
	registry[id] = up
	return nil
}

func ClearRegistry() { // used by tests
	for k := range registry {
		delete(registry, k)
	}
}

// Run applies registered migrations in ID order and records them.
func Run(db *gorm.DB) error {
	if len(registry) == 0 {
		return fmt.Errorf("no registered migrations found")
	}
	if err := ensureSchemaTable(db); err != nil {
		return err
	}

	applied, err := loadAppliedMigrations(db)
	if err != nil {
		return err
	}

	for _, id := range sortedRegistryIDs() {
		if applied[id] {
			continue
		}
		if err := applyMigration(db, id); err != nil {
			return err
		}
	}
	return nil
}

// VerifyApplied confirms that all registered migrations have already been
// recorded without mutating the schema. Request-serving startup paths use this
// when they are not the explicit startup writer.
func VerifyApplied(db *gorm.DB) error {
	if len(registry) == 0 {
		return fmt.Errorf("no registered migrations found")
	}
	if !db.Migrator().HasTable(&SchemaMigration{}) {
		return fmt.Errorf("schema migrations table is missing")
	}

	applied, err := loadAppliedMigrations(db)
	if err != nil {
		return err
	}

	for _, id := range sortedRegistryIDs() {
		if !applied[id] {
			return fmt.Errorf("registered migration %s has not been applied", id)
		}
	}
	return nil
}

func ensureSchemaTable(db *gorm.DB) error {
	if err := db.AutoMigrate(&SchemaMigration{}); err != nil {
		return fmt.Errorf("auto-migrate SchemaMigration: %w", err)
	}
	return nil
}

func loadAppliedMigrations(db *gorm.DB) (map[string]bool, error) {
	applied := map[string]bool{}
	var rows []SchemaMigration
	if err := db.Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("load SchemaMigration: %w", err)
	}
	for _, r := range rows {
		applied[r.ID] = true
	}
	return applied, nil
}

func sortedRegistryIDs() []string {
	ids := make([]string, 0, len(registry))
	for id := range registry {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func applyMigration(db *gorm.DB, id string) error {
	up := registry[id]
	if up == nil {
		return fmt.Errorf("migration %s has nil Up()", id)
	}
	if err := up(db); err != nil {
		return fmt.Errorf("migration %s failed: %w", id, err)
	}
	if err := db.Create(&SchemaMigration{ID: id, AppliedAt: time.Now()}).Error; err != nil {
		return fmt.Errorf("record SchemaMigration %s: %w", id, err)
	}
	return nil
}

// MigrateDB is the public entry for explicit startup writers.
func MigrateDB(db *gorm.DB) error {
	logger.LogInfo("Migration", "MigrateDB", "starting database migrations")

	if len(registry) == 0 {
		return fmt.Errorf("no registered migrations found")
	}

	if err := Run(db); err != nil {
		return err
	}

	logger.LogInfo("Migration", "MigrateDB", "database migrations completed")
	return nil
}
