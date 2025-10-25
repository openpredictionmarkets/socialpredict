package positions

import (
	"encoding/json"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	positionsmath "socialpredict/handlers/math/positions"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
	"socialpredict/util"

	"github.com/gorilla/mux"
)

func TestMarketDBPMPositionsHandler_IncludesZeroPositionUsers(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	origDB := util.DB
	util.DB = db
	t.Cleanup(func() {
		util.DB = origDB
	})
	_, _ = modelstesting.UseStandardTestEconomics(t)

	creator := modelstesting.GenerateUser("creator", 0)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("create creator: %v", err)
	}

	market := modelstesting.GenerateMarket(9001, creator.Username)
	market.IsResolved = true
	market.ResolutionResult = "YES"
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("create market: %v", err)
	}

	users := []string{"patrick", "jimmy", "jyron", "testuser03"}
	for _, username := range users {
		user := modelstesting.GenerateUser(username, 0)
		if err := db.Create(&user).Error; err != nil {
			t.Fatalf("create user %s: %v", username, err)
		}
	}

	bets := []struct {
		amount   int64
		outcome  string
		username string
		offset   time.Duration
	}{
		{50, "NO", "patrick", 0},
		{51, "NO", "jimmy", time.Second},
		{51, "NO", "jimmy", 2 * time.Second},
		{10, "YES", "jyron", 3 * time.Second},
		{30, "YES", "testuser03", 4 * time.Second},
	}

	for _, b := range bets {
		bet := modelstesting.GenerateBet(b.amount, b.outcome, b.username, uint(market.ID), b.offset)
		if err := db.Create(&bet).Error; err != nil {
			t.Fatalf("create bet: %v", err)
		}
	}

	req := httptest.NewRequest("GET", "/v0/markets/positions/"+strconv.FormatInt(market.ID, 10), nil)
	req = mux.SetURLVars(req, map[string]string{
		"marketId": strconv.FormatInt(market.ID, 10),
	})
	rec := httptest.NewRecorder()

	MarketDBPMPositionsHandler(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var positions []positionsmath.MarketPosition
	if err := json.Unmarshal(rec.Body.Bytes(), &positions); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	var locked *positionsmath.MarketPosition
	for i := range positions {
		if positions[i].Username == "testuser03" {
			locked = &positions[i]
			break
		}
	}

	if locked == nil {
		t.Fatalf("expected locked bettor to be present in handler response: %+v", positions)
	}

	if locked.YesSharesOwned != 0 || locked.NoSharesOwned != 0 || locked.Value != 0 {
		t.Fatalf("expected zero-valued position for locked bettor, got %+v", locked)
	}

	var totals models.Bet
	if err := db.Where("username = ? AND market_id = ?", "testuser03", market.ID).First(&totals).Error; err != nil {
		t.Fatalf("verify bets: %v", err)
	}
}
