package adminhandlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"socialpredict/models"
	"socialpredict/models/modelstesting"
	"socialpredict/util"
)

func setupAdminExportTest(t *testing.T) (adminJWT string) {
	t.Helper()
	db := modelstesting.NewFakeDB(t)
	origDB := util.DB
	util.DB = db
	t.Cleanup(func() { util.DB = origDB })
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")

	admin := modelstesting.GenerateUser("admin_user", 0)
	admin.UserType = "ADMIN"
	admin.MustChangePassword = false
	if err := db.Create(&admin).Error; err != nil {
		t.Fatalf("create admin: %v", err)
	}
	return modelstesting.GenerateValidJWT("admin_user")
}

func TestExportBetsCSV_AdminAccess(t *testing.T) {
	jwt := setupAdminExportTest(t)

	// Seed a bet
	user := modelstesting.GenerateUser("alice", 100)
	user.MustChangePassword = false
	util.DB.Create(&user)
	market := modelstesting.GenerateMarket(1, "admin_user")
	util.DB.Create(&market)
	bet := modelstesting.GenerateBet(50, "YES", "alice", uint(market.ID), 0)
	util.DB.Create(&bet)

	req := httptest.NewRequest("GET", "/v0/admin/export/bets", nil)
	req.Header.Set("Authorization", "Bearer "+jwt)
	rec := httptest.NewRecorder()

	ExportBetsCSVHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Header().Get("Content-Type"), "text/csv") {
		t.Error("expected Content-Type text/csv")
	}
	body := rec.Body.String()
	if !strings.Contains(body, "id,username,market_id") {
		t.Errorf("expected CSV header row, got: %s", body)
	}
	if !strings.Contains(body, "alice") {
		t.Error("expected alice in bets CSV")
	}
}

func TestExportBetsCSV_NonAdminForbidden(t *testing.T) {
	setupAdminExportTest(t)

	regularUser := modelstesting.GenerateUser("bob", 0)
	regularUser.MustChangePassword = false
	util.DB.Create(&regularUser)

	req := httptest.NewRequest("GET", "/v0/admin/export/bets", nil)
	req.Header.Set("Authorization", "Bearer "+modelstesting.GenerateValidJWT("bob"))
	rec := httptest.NewRecorder()

	ExportBetsCSVHandler(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for non-admin, got %d", rec.Code)
	}
}

func TestExportMarketsCSV_AdminAccess(t *testing.T) {
	jwt := setupAdminExportTest(t)

	market := modelstesting.GenerateMarket(1, "admin_user")
	util.DB.Create(&market)

	req := httptest.NewRequest("GET", "/v0/admin/export/markets", nil)
	req.Header.Set("Authorization", "Bearer "+jwt)
	rec := httptest.NewRecorder()

	ExportMarketsCSVHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, "id,creator_username,question_title") {
		t.Errorf("expected CSV header, got: %s", body)
	}
	if !strings.Contains(body, "Test Market") {
		t.Error("expected market title in CSV")
	}
}

func TestExportUsersCSV_ExcludesPasswordHashes(t *testing.T) {
	jwt := setupAdminExportTest(t)

	req := httptest.NewRequest("GET", "/v0/admin/export/users", nil)
	req.Header.Set("Authorization", "Bearer "+jwt)
	rec := httptest.NewRecorder()

	ExportUsersCSVHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	// Should have user data
	if !strings.Contains(body, "admin_user") {
		t.Error("expected admin_user in CSV")
	}
	// Should NOT contain password field in header
	if strings.Contains(body, "password") {
		t.Error("CSV must not include password field")
	}
}

func TestExportUsersCSV_Unauthenticated(t *testing.T) {
	setupAdminExportTest(t)

	req := httptest.NewRequest("GET", "/v0/admin/export/users", nil)
	rec := httptest.NewRecorder()

	ExportUsersCSVHandler(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestExportBetsCSV_ActionColumn(t *testing.T) {
	jwt := setupAdminExportTest(t)

	user := modelstesting.GenerateUser("alice", 100)
	user.MustChangePassword = false
	util.DB.Create(&user)
	market := modelstesting.GenerateMarket(1, "admin_user")
	util.DB.Create(&market)

	// BUY bet (positive amount)
	buyBet := models.Bet{Username: "alice", MarketID: uint(market.ID), Amount: 100, Outcome: "YES"}
	// SELL bet (negative amount)
	sellBet := models.Bet{Username: "alice", MarketID: uint(market.ID), Amount: -40, Outcome: "YES"}
	util.DB.Create(&buyBet)
	util.DB.Create(&sellBet)

	req := httptest.NewRequest("GET", "/v0/admin/export/bets", nil)
	req.Header.Set("Authorization", "Bearer "+jwt)
	rec := httptest.NewRecorder()

	ExportBetsCSVHandler(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "BUY") {
		t.Error("expected BUY in action column")
	}
	if !strings.Contains(body, "SELL") {
		t.Error("expected SELL in action column")
	}
}
