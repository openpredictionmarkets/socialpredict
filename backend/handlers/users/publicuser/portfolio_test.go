package publicuser

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"socialpredict/models"
	"socialpredict/models/modelstesting"
	"socialpredict/util"

	"github.com/gorilla/mux"
)

func TestFetchUserBetsReturnsBetsOrderedByPlacedAtDesc(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	user := modelstesting.GenerateUser("portfolio_bettor", 0)
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	market := modelstesting.GenerateMarket(8001, user.Username)
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("create market: %v", err)
	}

	first := modelstesting.GenerateBet(10, "YES", user.Username, uint(market.ID), -2*time.Hour)
	second := modelstesting.GenerateBet(15, "NO", user.Username, uint(market.ID), -1*time.Hour)
	if err := db.Create(&first).Error; err != nil {
		t.Fatalf("create bet1: %v", err)
	}
	if err := db.Create(&second).Error; err != nil {
		t.Fatalf("create bet2: %v", err)
	}

	bets, err := fetchUserBets(db, user.Username)
	if err != nil {
		t.Fatalf("fetchUserBets returned error: %v", err)
	}

	if len(bets) != 2 {
		t.Fatalf("expected 2 bets, got %d", len(bets))
	}
	if !bets[0].PlacedAt.After(bets[1].PlacedAt) {
		t.Fatalf("expected bets ordered by most recent first, got %v then %v", bets[0].PlacedAt, bets[1].PlacedAt)
	}
}

func TestMakeUserMarketMapTracksLastBet(t *testing.T) {
	now := time.Now()
	bets := []models.Bet{
		{MarketID: 1, PlacedAt: now.Add(-2 * time.Hour)},
		{MarketID: 1, PlacedAt: now.Add(-1 * time.Hour)},
		{MarketID: 2, PlacedAt: now.Add(-3 * time.Hour)},
	}

	result := makeUserMarketMap(bets)
	if len(result) != 2 {
		t.Fatalf("expected 2 markets, got %d", len(result))
	}

	if last := result[1].LastBetPlaced; !last.Equal(bets[1].PlacedAt) {
		t.Fatalf("expected last bet for market 1 to be %v, got %v", bets[1].PlacedAt, last)
	}
}

func TestProcessMarketMapReturnsPositionsWithTitles(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	_, _ = modelstesting.UseStandardTestEconomics(t)

	user := modelstesting.GenerateUser("portfolio_user", 0)
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	creator := modelstesting.GenerateUser("creator", 0)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("create creator: %v", err)
	}

	market := modelstesting.GenerateMarket(8101, creator.Username)
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("create market: %v", err)
	}

	other := modelstesting.GenerateUser("other", 0)
	if err := db.Create(&other).Error; err != nil {
		t.Fatalf("create other user: %v", err)
	}

	bets := []struct {
		amount   int64
		outcome  string
		username string
		offset   time.Duration
	}{
		{amount: 40, outcome: "YES", username: user.Username, offset: 0},
		{amount: 30, outcome: "NO", username: other.Username, offset: time.Second},
	}
	for _, b := range bets {
		bet := modelstesting.GenerateBet(b.amount, b.outcome, b.username, uint(market.ID), b.offset)
		if err := db.Create(&bet).Error; err != nil {
			t.Fatalf("create bet: %v", err)
		}
	}

	marketMap := map[uint]PortfolioItem{
		uint(market.ID): {
			MarketID:      uint(market.ID),
			LastBetPlaced: time.Now(),
		},
	}

	portfolio, err := processMarketMap(db, marketMap, user.Username)
	if err != nil {
		t.Fatalf("processMarketMap returned error: %v", err)
	}

	if len(portfolio) != 1 {
		t.Fatalf("expected single portfolio item, got %d", len(portfolio))
	}

	item := portfolio[0]
	if item.QuestionTitle != market.QuestionTitle {
		t.Fatalf("expected question title %q, got %q", market.QuestionTitle, item.QuestionTitle)
	}
	if item.YesSharesOwned == 0 && item.NoSharesOwned == 0 {
		t.Fatalf("expected shares to be populated, got %+v", item)
	}
}

func TestCalculateTotalSharesSumsPortfolio(t *testing.T) {
	portfolio := []PortfolioItem{
		{YesSharesOwned: 10, NoSharesOwned: 5},
		{YesSharesOwned: 3, NoSharesOwned: 2},
	}
	total := calculateTotalShares(portfolio)
	if total != 20 {
		t.Fatalf("expected total shares 20, got %d", total)
	}
}

func TestGetPortfolioReturnsAggregatedTotals(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	_, _ = modelstesting.UseStandardTestEconomics(t)

	orig := util.DB
	util.DB = db
	t.Cleanup(func() { util.DB = orig })

	user := modelstesting.GenerateUser("portfolio_handler", 0)
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	creator := modelstesting.GenerateUser("creator_portfolio", 0)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("create creator: %v", err)
	}

	market := modelstesting.GenerateMarket(8201, creator.Username)
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("create market: %v", err)
	}

	bet := modelstesting.GenerateBet(60, "YES", user.Username, uint(market.ID), 0)
	if err := db.Create(&bet).Error; err != nil {
		t.Fatalf("create bet: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v0/users/"+user.Username+"/portfolio", nil)
	req = mux.SetURLVars(req, map[string]string{"username": user.Username})
	rec := httptest.NewRecorder()

	GetPortfolio(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload PortfolioTotal
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if len(payload.PortfolioItems) != 1 {
		t.Fatalf("expected 1 portfolio item, got %d", len(payload.PortfolioItems))
	}

	if payload.TotalSharesOwned == 0 {
		t.Fatalf("expected total shares to be populated, got %d", payload.TotalSharesOwned)
	}
}
