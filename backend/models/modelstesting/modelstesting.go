// package modelstesting provides support for the
package modelstesting

import (
	"testing"

	"socialpredict/migration"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// NewFakeDB returns a sqlite db running in memory as a gorm.DB
func NewFakeDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to the database: %v", err)
	}
	migration.MigrateDB(db)
	return db
}
