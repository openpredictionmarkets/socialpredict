package metricshandlers

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"socialpredict/models/modelstesting"
	"socialpredict/util"
)

func TestGetGlobalLeaderboardHandler_Success(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	orig := util.DB
	util.DB = db
	t.Cleanup(func() {
		util.DB = orig
	})

	_, _ = modelstesting.UseStandardTestEconomics(t)

	users := []string{"creator", "patrick", "jimmy", "jyron"}
	for _, username := range users {
		user := modelstesting.GenerateUser(username, 0)
		if err := db.Create(&user).Error; err != nil {
			t.Fatalf("create user %s: %v", username, err)
		}
	}

	market := modelstesting.GenerateMarket(12001, "creator")
	market.IsResolved = true
	market.ResolutionResult = "YES"
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("create market: %v", err)
	}

	bets := []struct {
		amount   int64
		outcome  string
		username string
		offset   time.Duration
	}{
		{50, "NO", "patrick", 0},
		{51, "NO", "jimmy", time.Second},
		{11, "YES", "jyron", 2 * time.Second},
	}

	for _, b := range bets {
		bet := modelstesting.GenerateBet(b.amount, b.outcome, b.username, uint(market.ID), b.offset)
		if err := db.Create(&bet).Error; err != nil {
			t.Fatalf("create bet: %v", err)
		}
	}

	req := httptest.NewRequest("GET", "/v0/global/leaderboard", nil)
	rec := httptest.NewRecorder()

	GetGlobalLeaderboardHandler(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var payload []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if len(payload) == 0 {
		t.Fatalf("expected at least one leaderboard entry")
	}

	if _, ok := payload[0]["username"]; !ok {
		t.Fatalf("expected username field in leaderboard entry: %+v", payload[0])
	}
}

func TestGetGlobalLeaderboardHandler_Error(t *testing.T) {
	orig := util.DB
	util.DB = nil
	defer func() { util.DB = orig }()

	req := httptest.NewRequest("GET", "/v0/global/leaderboard", nil)
	rec := httptest.NewRecorder()

	GetGlobalLeaderboardHandler(rec, req)

	if rec.Code != 500 {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
}
