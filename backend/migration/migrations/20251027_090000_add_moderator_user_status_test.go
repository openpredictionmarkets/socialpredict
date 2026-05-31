package migrations_test

import (
	"testing"
	"time"

	"socialpredict/migration/migrations"
	"socialpredict/models"
	"socialpredict/models/modelstesting"

	"gorm.io/gorm"
)

type UserWithoutModeratorGovernance struct {
	ID                    int64 `gorm:"primaryKey"`
	CreatedAt             time.Time
	UpdatedAt             time.Time
	DeletedAt             gorm.DeletedAt `gorm:"index"`
	Username              string         `gorm:"unique;not null"`
	DisplayName           string         `gorm:"unique;not null"`
	UserType              string         `gorm:"not null"`
	InitialAccountBalance int64
	AccountBalance        int64
	Email                 string `gorm:"unique;not null"`
	APIKey                string `gorm:"unique"`
	Password              string `gorm:"not null"`
	MustChangePassword    bool   `gorm:"default:true"`
}

func (UserWithoutModeratorGovernance) TableName() string { return "users" }

func TestMigrateAddModeratorUserStatusAddsColumnsAuditTableAndBackfills(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	_ = db.Migrator().DropTable(&models.ModeratorRoleAudit{})
	_ = db.Migrator().DropColumn(&models.User{}, "ModeratorStatus")
	_ = db.Migrator().DropColumn(&models.User{}, "ModeratorSuspensionReason")
	_ = db.Migrator().DropColumn(&models.User{}, "ModeratorSuspendedBy")
	_ = db.Migrator().DropColumn(&models.User{}, "ModeratorSuspendedAt")

	seed := UserWithoutModeratorGovernance{
		ID:          77,
		Username:    "legacy_mod",
		DisplayName: "Legacy Mod",
		UserType:    "MODERATOR",
		Email:       "legacy-mod@example.com",
		APIKey:      "api-legacy-mod",
		Password:    "hash",
	}
	if err := db.Create(&seed).Error; err != nil {
		t.Fatalf("seed legacy user: %v", err)
	}

	if err := migrations.MigrateAddModeratorUserStatus(db); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	mig := db.Migrator()
	for _, column := range []string{"ModeratorStatus", "ModeratorSuspensionReason", "ModeratorSuspendedBy", "ModeratorSuspendedAt"} {
		if !mig.HasColumn(&models.User{}, column) {
			t.Fatalf("expected users.%s after migration", column)
		}
	}
	if !mig.HasTable(&models.ModeratorRoleAudit{}) {
		t.Fatalf("expected moderator_role_audits table after migration")
	}

	var out models.User
	if err := db.Where("username = ?", "legacy_mod").First(&out).Error; err != nil {
		t.Fatalf("load migrated user: %v", err)
	}
	if out.ModeratorStatus != "active" {
		t.Fatalf("moderator_status = %q, want active", out.ModeratorStatus)
	}
}
