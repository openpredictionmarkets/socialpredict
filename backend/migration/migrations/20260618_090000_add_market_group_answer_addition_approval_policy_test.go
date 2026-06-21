package migrations_test

import (
	"testing"

	"socialpredict/migration/migrations"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestMigrateAddMarketGroupAnswerAdditionApprovalPolicyCreatesColumn(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	if err := migrations.MigrateAddMarketGroupAnswerAdditionApprovalPolicy(db); err != nil {
		t.Fatalf("MigrateAddMarketGroupAnswerAdditionApprovalPolicy returned error: %v", err)
	}
	if !db.Migrator().HasColumn(&models.MarketGovernanceSettings{}, "MarketGroupAnswerAdditionApprovalPolicy") {
		t.Fatalf("expected MarketGroupAnswerAdditionApprovalPolicy column")
	}
}

func TestMigrateAddMarketGroupAnswerAdditionApprovalPolicyBackfillsLegacyBoolean(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	if err := db.AutoMigrate(&models.MarketGovernanceSettings{}); err != nil {
		t.Fatalf("auto migrate settings: %v", err)
	}
	if err := db.Create(&models.MarketGovernanceSettings{
		ID:                            1,
		AutoApproveMarketGroupAnswers: true,
		Version:                       1,
	}).Error; err != nil {
		t.Fatalf("seed settings: %v", err)
	}
	if err := db.Model(&models.MarketGovernanceSettings{}).Where("id = 1").Update("market_group_answer_addition_approval_policy", "").Error; err != nil {
		t.Fatalf("clear policy: %v", err)
	}

	if err := migrations.MigrateAddMarketGroupAnswerAdditionApprovalPolicy(db); err != nil {
		t.Fatalf("MigrateAddMarketGroupAnswerAdditionApprovalPolicy returned error: %v", err)
	}

	var row models.MarketGovernanceSettings
	if err := db.First(&row, 1).Error; err != nil {
		t.Fatalf("reload settings: %v", err)
	}
	if row.MarketGroupAnswerAdditionApprovalPolicy != "auto" {
		t.Fatalf("policy = %q, want auto", row.MarketGroupAnswerAdditionApprovalPolicy)
	}
}

func TestMigrateAddMarketGroupAnswerAdditionApprovalPolicyIsIdempotent(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	if err := migrations.MigrateAddMarketGroupAnswerAdditionApprovalPolicy(db); err != nil {
		t.Fatalf("first migration returned error: %v", err)
	}
	if err := migrations.MigrateAddMarketGroupAnswerAdditionApprovalPolicy(db); err != nil {
		t.Fatalf("second migration returned error: %v", err)
	}
}
