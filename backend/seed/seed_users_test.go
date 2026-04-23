package seed

import (
	"testing"

	configsvc "socialpredict/internal/service/config"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestSeedUsersRequiresConfigService(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	t.Setenv("ADMIN_PASSWORD", "secret-password")

	err := SeedUsers(db, nil)
	if err == nil {
		t.Fatalf("expected error when config service is nil")
	}
}

func TestSeedUsersCreatesAdminUsingInjectedConfig(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	t.Setenv("ADMIN_PASSWORD", "secret-password")

	cfg := modelstesting.GenerateEconomicConfig()
	cfg.Economics.User.InitialAccountBalance = 321

	if err := SeedUsers(db, configsvc.NewStaticService(cfg)); err != nil {
		t.Fatalf("SeedUsers returned error: %v", err)
	}

	var admin models.User
	if err := db.Where("username = ?", "admin").First(&admin).Error; err != nil {
		t.Fatalf("load admin user: %v", err)
	}

	if admin.InitialAccountBalance != 321 {
		t.Fatalf("expected injected initial balance 321, got %d", admin.InitialAccountBalance)
	}
	if admin.AccountBalance != 321 {
		t.Fatalf("expected injected account balance 321, got %d", admin.AccountBalance)
	}
}
