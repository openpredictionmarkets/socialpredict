package migrations_test

import (
	"testing"

	"socialpredict/migration/migrations"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestMigrateAddSocialShareSettingsCreatesTable(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	_ = db.Migrator().DropTable(&models.SocialShareSettings{})

	if err := migrations.MigrateAddSocialShareSettings(db); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	if !db.Migrator().HasTable(&models.SocialShareSettings{}) {
		t.Fatalf("expected SocialShareSettings table after migration")
	}
	for _, column := range []string{"Slug", "SiteName", "DefaultDescription", "DefaultImageURL", "ImageAlt", "Version", "UpdatedBy"} {
		if !db.Migrator().HasColumn(&models.SocialShareSettings{}, column) {
			t.Fatalf("expected %s column after migration", column)
		}
	}
	if !db.Migrator().HasTable(&models.SocialShareImage{}) {
		t.Fatalf("expected SocialShareImage table after migration")
	}
	for _, column := range []string{"Slug", "FileName", "ContentType", "SizeBytes", "Data", "UpdatedBy"} {
		if !db.Migrator().HasColumn(&models.SocialShareImage{}, column) {
			t.Fatalf("expected SocialShareImage.%s column after migration", column)
		}
	}
}
