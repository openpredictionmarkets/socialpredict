package migration

import (
	"fmt"
	"log"
	"sort"
	"time"

	"gorm.io/gorm"

	// core models for fallback
	"socialpredict/models"
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
	// Ensure tracking table exists
	if err := db.AutoMigrate(&SchemaMigration{}); err != nil {
		return fmt.Errorf("auto-migrate SchemaMigration: %w", err)
	}

	// Load already-applied
	applied := map[string]bool{}
	var rows []SchemaMigration
	if err := db.Find(&rows).Error; err != nil {
		return fmt.Errorf("load SchemaMigration: %w", err)
	}
	for _, r := range rows {
		applied[r.ID] = true
	}

	// Sort and apply
	ids := make([]string, 0, len(registry))
	for id := range registry {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		if applied[id] {
			continue
		}
		up := registry[id]
		if up == nil {
			// <-- Fix: return a proper error instead of panic
			return fmt.Errorf("migration %s has nil Up()", id)
		}
		if err := up(db); err != nil {
			return fmt.Errorf("migration %s failed: %w", id, err)
		}
		if err := db.Create(&SchemaMigration{ID: id, AppliedAt: time.Now()}).Error; err != nil {
			return fmt.Errorf("record SchemaMigration %s: %w", id, err)
		}
		// optional: log.Printf("migration - applied %s", id)
	}
	return nil
}

// MigrateDB is the public entry; it never crashes the app.
// If there are zero registered migrations, we WARN and fallback to AutoMigrate core tables.
func MigrateDB(db *gorm.DB) error {
	log.Printf("migration - MigrateDB: starting database migrations")

	if len(registry) == 0 {
		log.Printf("migration - WARN: no registered migrations found; falling back to AutoMigrate for baseline schema")
		// Baseline schema so the app can run:
		// Keep this list tight (public, stable domain models only).
		if err := db.AutoMigrate(
			&models.User{},
			&models.Market{},
			&models.Bet{},
			&models.HomepageContent{},
		); err != nil {
			return fmt.Errorf("fallback AutoMigrate failed: %w", err)
		}
		log.Printf("migration - fallback AutoMigrate completed")
		return nil
	}

	if err := Run(db); err != nil {
		return err
	}

	log.Printf("migration - MigrateDB: database migrations completed")
	return nil
}
