package seed

import (
	"testing"

	configsvc "socialpredict/internal/service/config"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

type panicOnCurrentSeedConfigService struct {
	economics configsvc.Economics
}

func (s panicOnCurrentSeedConfigService) Current() *configsvc.AppConfig {
	panic("Current should not be called")
}

func (s panicOnCurrentSeedConfigService) Economics() configsvc.Economics {
	return s.economics
}

func (s panicOnCurrentSeedConfigService) Frontend() configsvc.Frontend {
	panic("Frontend should not be called")
}

func (s panicOnCurrentSeedConfigService) ChartSigFigs() int {
	panic("ChartSigFigs should not be called")
}

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

func TestSeedUsersAvoidsWholeTreeConfigAccess(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	t.Setenv("ADMIN_PASSWORD", "secret-password")

	cfg := modelstesting.GenerateEconomicConfig()
	cfg.Economics.User.InitialAccountBalance = 654

	if err := SeedUsers(db, panicOnCurrentSeedConfigService{economics: configsvc.FromSetup(cfg).Economics}); err != nil {
		t.Fatalf("SeedUsers returned error: %v", err)
	}

	var admin models.User
	if err := db.Where("username = ?", "admin").First(&admin).Error; err != nil {
		t.Fatalf("load admin user: %v", err)
	}

	if admin.InitialAccountBalance != 654 {
		t.Fatalf("expected injected initial balance 654, got %d", admin.InitialAccountBalance)
	}
	if admin.AccountBalance != 654 {
		t.Fatalf("expected injected account balance 654, got %d", admin.AccountBalance)
	}
}
