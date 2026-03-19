package publicuser

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
	"socialpredict/util"
)

func TestGetBetHistory_Empty(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	origDB := util.DB
	util.DB = db
	t.Cleanup(func() { util.DB = origDB })

	user := modelstesting.GenerateUser("alice", 0)
	user.MustChangePassword = false
	db.Create(&user)

	req := httptest.NewRequest("GET", "/v0/users/alice/bets", nil)
	req = mux.SetURLVars(req, map[string]string{"username": "alice"})
	rec := httptest.NewRecorder()

	GetBetHistory(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var result []BetHistoryItem
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty, got %d", len(result))
	}
}

func TestGetBetHistory_ReturnsBetsWithMarketTitle(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	origDB := util.DB
	util.DB = db
	t.Cleanup(func() { util.DB = origDB })

	user := modelstesting.GenerateUser("alice", 100)
	user.MustChangePassword = false
	db.Create(&user)

	market := modelstesting.GenerateMarket(1, "alice")
	db.Create(&market)

	bet1 := modelstesting.GenerateBet(100, "YES", "alice", uint(market.ID), 0)
	bet2 := modelstesting.GenerateBet(-50, "YES", "alice", uint(market.ID), time.Minute)
	db.Create(&bet1)
	db.Create(&bet2)

	req := httptest.NewRequest("GET", "/v0/users/alice/bets", nil)
	req = mux.SetURLVars(req, map[string]string{"username": "alice"})
	rec := httptest.NewRecorder()

	GetBetHistory(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result []BetHistoryItem
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 bets, got %d", len(result))
	}

	// newest first — bet2 (sell) should be first
	if result[0].Action != "SELL" {
		t.Fatalf("expected first result to be SELL, got %s", result[0].Action)
	}
	if result[1].Action != "BUY" {
		t.Fatalf("expected second result to be BUY, got %s", result[1].Action)
	}
	if result[0].QuestionTitle != market.QuestionTitle {
		t.Fatalf("expected market title %q, got %q", market.QuestionTitle, result[0].QuestionTitle)
	}
}

func TestGetBetHistory_OnlyShowsRequestedUser(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	origDB := util.DB
	util.DB = db
	t.Cleanup(func() { util.DB = origDB })

	alice := modelstesting.GenerateUser("alice", 100)
	alice.MustChangePassword = false
	bob := modelstesting.GenerateUser("bob", 100)
	bob.MustChangePassword = false
	db.Create(&alice)
	db.Create(&bob)

	market := modelstesting.GenerateMarket(1, "alice")
	db.Create(&market)

	db.Create(&modelstesting.GenerateBet(100, "YES", "alice", uint(market.ID), 0))
	db.Create(&modelstesting.GenerateBet(50, "NO", "bob", uint(market.ID), 0))

	req := httptest.NewRequest("GET", "/v0/users/alice/bets", nil)
	req = mux.SetURLVars(req, map[string]string{"username": "alice"})
	rec := httptest.NewRecorder()

	GetBetHistory(rec, req)

	var result []BetHistoryItem
	json.Unmarshal(rec.Body.Bytes(), &result)
	if len(result) != 1 {
		t.Fatalf("expected 1 bet for alice, got %d", len(result))
	}
	if result[0].Outcome != "YES" {
		t.Fatalf("expected alice's YES bet, got %s", result[0].Outcome)
	}
}
