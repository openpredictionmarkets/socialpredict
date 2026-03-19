package betshandlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
	"socialpredict/util"
	"testing"
	"time"

	"github.com/gorilla/mux"
)

func TestMarketBetsDisplayHandler(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	util.DB = db

	// Create a market with a known CreatedAt time
	market := &models.Market{
		QuestionTitle:      "Will it rain?",
		Description:        "Test market",
		OutcomeType:        "BINARY",
		ResolutionDateTime: time.Now().Add(24 * time.Hour),
		InitialProbability: 0.5,
		CreatorUsername:    "alice",
	}
	db.Create(market)

	marketIDStr := fmt.Sprintf("%d", market.ID)

	t.Run("returns empty array when no bets placed", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v0/markets/bets/"+marketIDStr, nil)
		req = mux.SetURLVars(req, map[string]string{"marketId": marketIDStr})
		rr := httptest.NewRecorder()

		MarketBetsDisplayHandler(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
		// Body should be a JSON array (empty or null)
		var result interface{}
		if err := json.Unmarshal(rr.Body.Bytes(), &result); err != nil {
			t.Errorf("expected valid JSON response, got: %s", rr.Body.String())
		}
	})

	t.Run("returns 404 for unknown market", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v0/markets/bets/99999", nil)
		req = mux.SetURLVars(req, map[string]string{"marketId": "99999"})
		rr := httptest.NewRecorder()

		MarketBetsDisplayHandler(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rr.Code)
		}
	})

	t.Run("returns 400 for invalid market ID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v0/markets/bets/notanumber", nil)
		req = mux.SetURLVars(req, map[string]string{"marketId": "notanumber"})
		rr := httptest.NewRecorder()

		MarketBetsDisplayHandler(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rr.Code)
		}
	})

	t.Run("returns bets sorted by placedAt ascending", func(t *testing.T) {
		now := time.Now()
		bets := []models.Bet{
			{
				MarketID: uint(market.ID),
				Username: "alice",
				Outcome:  "YES",
				Amount:   10,
				PlacedAt: now.Add(2 * time.Second),
			},
			{
				MarketID: uint(market.ID),
				Username: "bob",
				Outcome:  "NO",
				Amount:   5,
				PlacedAt: now.Add(1 * time.Second),
			},
		}
		for i := range bets {
			db.Create(&bets[i])
		}

		req, _ := http.NewRequest("GET", "/v0/markets/bets/"+marketIDStr, nil)
		req = mux.SetURLVars(req, map[string]string{"marketId": marketIDStr})
		rr := httptest.NewRecorder()

		MarketBetsDisplayHandler(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
		}

		var result []BetDisplayInfo
		if err := json.Unmarshal(rr.Body.Bytes(), &result); err != nil {
			t.Fatalf("error parsing response: %v", err)
		}
		if len(result) < 2 {
			t.Fatalf("expected at least 2 bets, got %d", len(result))
		}
		// Should be sorted ascending by PlacedAt — bob (earlier) before alice
		if result[0].Username != "bob" {
			t.Errorf("expected bob first (earlier bet), got %s", result[0].Username)
		}
	})
}
