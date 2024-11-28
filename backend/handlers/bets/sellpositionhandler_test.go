package betshandlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"socialpredict/handlers/positions"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Test parseRedeemRequest
func TestParseRedeemRequest(t *testing.T) {
	validBet := models.Bet{
		MarketID: 1,
		Amount:   10,
		Outcome:  "YES",
	}

	body, _ := json.Marshal(validBet)
	req := httptest.NewRequest(http.MethodPost, "/sell", bytes.NewReader(body))
	w := httptest.NewRecorder()

	redeemRequest, err := parseRedeemRequest(w, req)
	assert.NoError(t, err)
	assert.NotNil(t, redeemRequest)
	assert.Equal(t, validBet.MarketID, redeemRequest.MarketID)
	assert.Equal(t, validBet.Outcome, redeemRequest.Outcome)
	assert.Equal(t, validBet.Amount, redeemRequest.Amount)
}

func TestParseRedeemRequest_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/sell", bytes.NewReader([]byte("invalid-json")))
	w := httptest.NewRecorder()

	redeemRequest, err := parseRedeemRequest(w, req)
	assert.Error(t, err)
	assert.Nil(t, redeemRequest)
	assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
}

// Test validateRedeemAmount
func TestValidateRedeemAmount(t *testing.T) {
	userNetPosition := positions.UserMarketPosition{
		YesSharesOwned: 10,
		NoSharesOwned:  5,
	}

	validBet := &models.Bet{Amount: 5, Outcome: "YES"}
	invalidBet := &models.Bet{Amount: 15, Outcome: "YES"}

	err := validateRedeemAmount(validBet, userNetPosition)
	assert.NoError(t, err)

	err = validateRedeemAmount(invalidBet, userNetPosition)
	assert.Error(t, err)
	assert.Equal(t, "redeem amount exceeds available position", err.Error())
}

// Test createBet
func TestCreateBet(t *testing.T) {
	redeemRequest := &models.Bet{
		MarketID: 1,
		Amount:   -5,
		Outcome:  "NO",
	}
	username := "testuser"

	bet := models.CreateBet(username, redeemRequest.MarketID, redeemRequest.Amount, redeemRequest.Outcome)
	assert.Equal(t, redeemRequest.MarketID, bet.MarketID)
	assert.Equal(t, redeemRequest.Amount, bet.Amount)
	assert.Equal(t, redeemRequest.Outcome, bet.Outcome)
	assert.Equal(t, username, bet.Username)
	assert.WithinDuration(t, time.Now(), bet.PlacedAt, time.Second)
}

// Test reduceUserAccountBalance
//func TestReduceUserAccountBalance(t *testing.T) {
//	db := modelstesting.NewFakeDB(t)

//	user := &models.User{AccountBalance: 50}
//	bet := &models.Bet{Amount: 10}

//	err := reduceUseAccountBalance(db, user, bet)
//	assert.NoError(t, err)
//	assert.Equal(t, 40, user.AccountBalance)

//	bet.Amount = 60
//	err = reduceUseAccountBalance(db, user, bet)
//	assert.Error(t, err)
//	assert.Contains(t, err.Error(), "negative balance not allowed")
//}

// Test createBetInDatabase
func TestCreateBetInDatabase(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	bet := &models.Bet{MarketID: 1, Username: "testuser", Amount: -5}
	err := createBetInDatabase(db, bet)
	assert.NoError(t, err)

	// Verify bet was added to the fake database
	var fetchedBet models.Bet
	db.First(&fetchedBet)
	assert.Equal(t, bet.MarketID, fetchedBet.MarketID)
	assert.Equal(t, bet.Username, fetchedBet.Username)
	assert.Equal(t, bet.Amount, fetchedBet.Amount)
}

// Test respondSuccess
func TestRespondSuccess(t *testing.T) {
	redeemRequest := &models.Bet{
		MarketID: 1,
		Amount:   -5,
		Outcome:  "YES",
	}

	w := httptest.NewRecorder()
	respondSuccess(w, redeemRequest)

	assert.Equal(t, http.StatusCreated, w.Result().StatusCode)

	var response models.Bet
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, redeemRequest, &response)
}
