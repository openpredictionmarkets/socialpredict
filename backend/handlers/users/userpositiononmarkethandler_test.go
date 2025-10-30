package usershandlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"

	positionsmath "socialpredict/internal/domain/math/positions"
	"socialpredict/internal/app"
	"socialpredict/middleware"
	"socialpredict/models/modelstesting"
)

func TestUserMarketPositionHandlerReturnsUserPosition(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	_, _ = modelstesting.UseStandardTestEconomics(t)
	config := modelstesting.GenerateEconomicConfig()
	container := app.BuildApplication(db, config)

	t.Setenv("JWT_SIGNING_KEY", "test-secret-key")

	creator := modelstesting.GenerateUser("creator", 0)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("create creator: %v", err)
	}

	market := modelstesting.GenerateMarket(7001, creator.Username)
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("create market: %v", err)
	}

	user := modelstesting.GenerateUser("bettor", 0)
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	other := modelstesting.GenerateUser("otherbettor", 0)
	if err := db.Create(&other).Error; err != nil {
		t.Fatalf("create other user: %v", err)
	}

	bets := []struct {
		amount   int64
		outcome  string
		username string
		offset   time.Duration
	}{
		{amount: 50, outcome: "YES", username: user.Username, offset: 0},
		{amount: 25, outcome: "NO", username: other.Username, offset: time.Second},
	}

	for _, b := range bets {
		bet := modelstesting.GenerateBet(b.amount, b.outcome, b.username, uint(market.ID), b.offset)
		if err := db.Create(&bet).Error; err != nil {
			t.Fatalf("create bet: %v", err)
		}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &middleware.UserClaims{
		Username: user.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour).Unix(),
		},
	})
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SIGNING_KEY")))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v0/user/markets/"+strconv.FormatInt(market.ID, 10), nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	req = mux.SetURLVars(req, map[string]string{
		"marketId": strconv.FormatInt(market.ID, 10),
	})
	rec := httptest.NewRecorder()

	handler := UserMarketPositionHandlerWithService(container.GetMarketsService(), container.GetUsersService())
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var position positionsmath.UserMarketPosition
	if err := json.Unmarshal(rec.Body.Bytes(), &position); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if position.YesSharesOwned == 0 && position.NoSharesOwned == 0 {
		t.Fatalf("expected non-zero shares for bettor, got %+v", position)
	}
}

func TestUserMarketPositionHandlerUnauthorizedWithoutToken(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	_, _ = modelstesting.UseStandardTestEconomics(t)
	config := modelstesting.GenerateEconomicConfig()
	container := app.BuildApplication(db, config)

	req := httptest.NewRequest(http.MethodGet, "/v0/user/markets/1", nil)
	req = mux.SetURLVars(req, map[string]string{"marketId": "1"})
	rec := httptest.NewRecorder()

	handler := UserMarketPositionHandlerWithService(container.GetMarketsService(), container.GetUsersService())
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}
