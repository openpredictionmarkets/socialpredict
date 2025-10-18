package migration

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"socialpredict/logger"

	"gorm.io/gorm"
)

type SchemaMigration struct {
	ID        string `gorm:"primaryKey;size:32"` // e.g., "20251013080000"
	AppliedAt time.Time
}

type step struct {
	ID string
	Up func(db *gorm.DB) error
}

var registry = make(map[string]step)

// for testing
func ClearRegistry() {
	registry = make(map[string]step)
}

func Register(id string, up func(*gorm.DB) error) error {
	if _, exists := registry[id]; exists {
		err := errors.New("duplicate migration id: " + id)
		logger.LogError("migration", "Register", err)
		return err
	}
	registry[id] = step{ID: id, Up: up}
	logger.LogInfo("migration", "Register", fmt.Sprintf("registered migration %s", id))
	return nil
}

func ensureTable(db *gorm.DB) error {
	return db.AutoMigrate(&SchemaMigration{})
}

func appliedSet(db *gorm.DB) (map[string]struct{}, error) {
	if err := ensureTable(db); err != nil {
		return nil, err
	}
	var rows []SchemaMigration
	if err := db.Order("id asc").Find(&rows).Error; err != nil {
		return nil, err
	}
	m := make(map[string]struct{}, len(rows))
	for _, r := range rows {
		m[r.ID] = struct{}{}
	}
	return m, nil
}

func Run(db *gorm.DB) error {
	applied, err := appliedSet(db)
	if err != nil {
		return err
	}

	ids := make([]string, 0, len(registry))
	for id := range registry {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		if _, ok := applied[id]; ok {
			continue
		}
		step := registry[id]
		if step.Up == nil {
			return errors.New("migration has no Up(): " + id)
		}
		if err := step.Up(db); err != nil {
			return err
		}
		if err := db.Create(&SchemaMigration{ID: id, AppliedAt: time.Now()}).Error; err != nil {
			return err
		}
	}
	return nil
}

func MigrateDB(db *gorm.DB) error {
	logger.LogInfo("migration", "MigrateDB", "starting database migrations")

	err := Run(db)
	if err != nil {
		logger.LogError("migration", "MigrateDB", err)
		return fmt.Errorf("error running migrations: %w", err)
	}

	logger.LogInfo("migration", "MigrateDB", "database migrations completed successfully")
	return nil
}
