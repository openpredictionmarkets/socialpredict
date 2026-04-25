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

// SellPositionHandler returns an HTTP handler that delegates sales to the bets service.
func SellPositionHandler(betsSvc bets.ServiceInterface, usersSvc dusers.ServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			_ = handlers.WriteFailure(w, http.StatusMethodNotAllowed, handlers.ReasonMethodNotAllowed)
			return
		}

		user, httpErr := authsvc.ValidateUserAndEnforcePasswordChangeGetUser(r, usersSvc)
		if httpErr != nil {
			_ = handlers.WriteFailure(w, httpErr.StatusCode, handlers.AuthFailureReason(httpErr.StatusCode, httpErr.Message))
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
		_ = handlers.WriteFailure(w, http.StatusUnprocessableEntity, handlers.ReasonDustCapExceeded)
		return
	}

	switch {
	case errors.Is(err, bets.ErrInvalidOutcome), errors.Is(err, bets.ErrInvalidAmount):
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonValidationFailed)
	case errors.Is(err, bets.ErrMarketClosed):
		_ = handlers.WriteFailure(w, http.StatusConflict, handlers.ReasonMarketClosed)
	case errors.Is(err, bets.ErrNoPosition):
		_ = handlers.WriteFailure(w, http.StatusUnprocessableEntity, handlers.ReasonNoPosition)
	case errors.Is(err, bets.ErrInsufficientShares):
		_ = handlers.WriteFailure(w, http.StatusUnprocessableEntity, handlers.ReasonInsufficientShares)
	case errors.Is(err, dmarkets.ErrMarketNotFound):
		_ = handlers.WriteFailure(w, http.StatusNotFound, handlers.ReasonMarketNotFound)
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
