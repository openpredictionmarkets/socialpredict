package migrations

import (
	"socialpredict/migration"
	"socialpredict/models"

	"gorm.io/gorm"
)

func MigrateAddModeratorUserStatus(db *gorm.DB) error {
	m := db.Migrator()

	if !m.HasColumn(&models.User{}, "ModeratorStatus") {
		if err := m.AddColumn(&models.User{}, "ModeratorStatus"); err != nil {
			return err
		}
	}
	if !m.HasColumn(&models.User{}, "ModeratorSuspensionReason") {
		if err := m.AddColumn(&models.User{}, "ModeratorSuspensionReason"); err != nil {
			return err
		}
	}
	if !m.HasColumn(&models.User{}, "ModeratorSuspendedBy") {
		if err := m.AddColumn(&models.User{}, "ModeratorSuspendedBy"); err != nil {
			return err
		}
	}
	if !m.HasColumn(&models.User{}, "ModeratorSuspendedAt") {
		if err := m.AddColumn(&models.User{}, "ModeratorSuspendedAt"); err != nil {
			return err
		}
	}

	if err := db.AutoMigrate(&models.ModeratorRoleAudit{}); err != nil {
		return err
	}

	if err := db.Model(&models.User{}).
		Where("moderator_status IS NULL OR moderator_status = ''").
		Update("moderator_status", "none").Error; err != nil {
		return err
	}

	return db.Model(&models.User{}).
		Where("LOWER(user_type) = ? AND moderator_status = ?", "moderator", "none").
		Update("moderator_status", "active").Error
}

func init() {
	migration.Register("20251027090000", func(db *gorm.DB) error {
		return MigrateAddModeratorUserStatus(db)
	})
}
