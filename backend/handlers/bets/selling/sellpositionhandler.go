package sellbetshandlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"socialpredict/handlers"
	"socialpredict/handlers/bets/dto"
	bets "socialpredict/internal/domain/bets"
	dmarkets "socialpredict/internal/domain/markets"
	dusers "socialpredict/internal/domain/users"
	authsvc "socialpredict/internal/service/auth"
)

const (
	reasonSellValidationFailed handlers.FailureReason = "SELL_VALIDATION_FAILED"
	reasonSellMarketClosed     handlers.FailureReason = "MARKET_CLOSED"
	reasonNoPosition           handlers.FailureReason = "NO_POSITION"
	reasonInsufficientShares   handlers.FailureReason = "INSUFFICIENT_SHARES"
	reasonDustCapExceeded      handlers.FailureReason = "DUST_CAP_EXCEEDED"
	reasonSellMarketNotFound   handlers.FailureReason = "MARKET_NOT_FOUND"
)

// SellPositionHandler returns an HTTP handler that delegates sales to the bets service.
func SellPositionHandler(betsSvc bets.ServiceInterface, usersSvc dusers.ServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			_ = handlers.WriteFailure(w, http.StatusMethodNotAllowed, handlers.ReasonMethodNotAllowed)
			return
		}

		user, httpErr := authsvc.ValidateUserAndEnforcePasswordChangeGetUser(r, usersSvc)
		if httpErr != nil {
			_ = handlers.WriteFailure(w, httpErr.StatusCode, reasonFromAuthError(httpErr))
			return
		}

		req, decodeErr := decodeSellRequest(r)
		if decodeErr != nil {
			_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
			return
		}

		result, err := betsSvc.Sell(r.Context(), toSellRequest(req, user.Username))
		if err != nil {
			handleSellError(w, err)
			return
		}

		writeSellResponse(w, result)
	}
}

func decodeSellRequest(r *http.Request) (dto.SellBetRequest, error) {
	var req dto.SellBetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return dto.SellBetRequest{}, errors.New("Invalid request body")
	}
	return req, nil
}

func toSellRequest(req dto.SellBetRequest, username string) bets.SellRequest {
	return bets.SellRequest{
		Username: username,
		MarketID: req.MarketID,
		Amount:   req.Amount,
		Outcome:  req.Outcome,
	}
}

func handleSellError(w http.ResponseWriter, err error) {
	if _, ok := err.(bets.ErrDustCapExceeded); ok {
		_ = handlers.WriteFailure(w, http.StatusUnprocessableEntity, reasonDustCapExceeded)
		return
	}

	switch {
	case errors.Is(err, bets.ErrInvalidOutcome), errors.Is(err, bets.ErrInvalidAmount):
		_ = handlers.WriteFailure(w, http.StatusBadRequest, reasonSellValidationFailed)
	case errors.Is(err, bets.ErrMarketClosed):
		_ = handlers.WriteFailure(w, http.StatusConflict, reasonSellMarketClosed)
	case errors.Is(err, bets.ErrNoPosition), errors.Is(err, bets.ErrInsufficientShares):
		reason := reasonNoPosition
		if errors.Is(err, bets.ErrInsufficientShares) {
			reason = reasonInsufficientShares
		}
		_ = handlers.WriteFailure(w, http.StatusUnprocessableEntity, reason)
	case errors.Is(err, dmarkets.ErrMarketNotFound):
		_ = handlers.WriteFailure(w, http.StatusNotFound, reasonSellMarketNotFound)
	default:
		_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
	}
}

func writeSellResponse(w http.ResponseWriter, result *bets.SellResult) {
	response := dto.SellBetResponse{
		Username:      result.Username,
		MarketID:      result.MarketID,
		SharesSold:    result.SharesSold,
		SaleValue:     result.SaleValue,
		Dust:          result.Dust,
		Outcome:       result.Outcome,
		TransactionAt: result.TransactionAt,
	}

	_ = handlers.WriteResult(w, http.StatusCreated, response)
}

func reasonFromAuthError(err *authsvc.HTTPError) handlers.FailureReason {
	if err == nil {
		return handlers.ReasonInternalError
	}

	switch err.Message {
	case "Authorization header is required", "Invalid token":
		return handlers.ReasonInvalidToken
	case "Password change required":
		return handlers.ReasonPasswordChangeRequired
	case "User not found":
		return handlers.ReasonUserNotFound
	default:
		if err.StatusCode >= http.StatusInternalServerError {
			return handlers.ReasonInternalError
		}
		return handlers.ReasonInvalidToken
	}
}
