package migrations_test

import (
	"testing"
	"time"

	"socialpredict/migration/migrations"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestMigrateAddMarketGroupsCreatesTablesAndIndexes(t *testing.T) {
	db := modelstesting.NewTestDB(t)

	if err := migrations.MigrateAddMarketGroups(db); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	for _, table := range []any{
		&models.MarketGroup{},
		&models.MarketGroupMember{},
	} {
		if !db.Migrator().HasTable(table) {
			t.Fatalf("expected table for %T", table)
		}
	}

	if !db.Migrator().HasIndex(&models.MarketGroup{}, "idx_market_groups_lifecycle_status") {
		t.Fatalf("expected lifecycle status index")
	}
	if !db.Migrator().HasIndex(&models.MarketGroupMember{}, "uniq_market_group_member_market") {
		t.Fatalf("expected unique group/member market index")
	}
	if !db.Migrator().HasIndex(&models.MarketGroupMember{}, "idx_market_group_members_group_order") {
		t.Fatalf("expected group display-order index")
	}
}

func TestMigrateAddMarketGroupsIsIdempotent(t *testing.T) {
	db := modelstesting.NewTestDB(t)
	if err := migrations.MigrateAddMarketGroups(db); err != nil {
		t.Fatalf("first migration failed: %v", err)
	}
	if err := migrations.MigrateAddMarketGroups(db); err != nil {
		t.Fatalf("second migration failed: %v", err)
	}
}

func TestMigrateAddMarketGroupsAppliesDefaults(t *testing.T) {
	db := modelstesting.NewTestDB(t)
	if err := migrations.MigrateAddMarketGroups(db); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	group := models.MarketGroup{
		QuestionTitle:      "Who will win?",
		Description:        "Multiple choice binary market group.",
		CreatorUsername:    "moderator",
		ResolutionDateTime: time.Now().UTC().Add(24 * time.Hour),
	}
	if err := db.Create(&group).Error; err != nil {
		t.Fatalf("create group: %v", err)
	}

	var got models.MarketGroup
	if err := db.First(&got, group.ID).Error; err != nil {
		t.Fatalf("reload group: %v", err)
	}
	if got.GroupType != "MULTIPLE_CHOICE_BINARY" {
		t.Fatalf("GroupType = %q", got.GroupType)
	}
	if got.ProbabilityPolicy != "INDEPENDENT_BINARY" {
		t.Fatalf("ProbabilityPolicy = %q", got.ProbabilityPolicy)
	}
	if got.ResolutionPolicy != "INDEPENDENT_CHILDREN" {
		t.Fatalf("ResolutionPolicy = %q", got.ResolutionPolicy)
	}
	if got.LifecycleStatus != "proposed" {
		t.Fatalf("LifecycleStatus = %q", got.LifecycleStatus)
	}

	member := models.MarketGroupMember{
		GroupID:     group.ID,
		MarketID:    123,
		AnswerLabel: "Team A",
	}
	if err := db.Create(&member).Error; err != nil {
		t.Fatalf("create member: %v", err)
	}
	var gotMember models.MarketGroupMember
	if err := db.First(&gotMember, member.ID).Error; err != nil {
		t.Fatalf("reload member: %v", err)
	}
	if gotMember.DisplayOrder != 0 {
		t.Fatalf("DisplayOrder = %d", gotMember.DisplayOrder)
	}
}
