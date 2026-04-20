package buybetshandlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"socialpredict/handlers"
	"socialpredict/handlers/bets/dto"
	dbets "socialpredict/internal/domain/bets"
	dmarkets "socialpredict/internal/domain/markets"
	dusers "socialpredict/internal/domain/users"
	authsvc "socialpredict/internal/service/auth"
)

const (
	reasonBetValidationFailed handlers.FailureReason = "BET_VALIDATION_FAILED"
	reasonMarketClosed        handlers.FailureReason = "MARKET_CLOSED"
	reasonInsufficientBalance handlers.FailureReason = "INSUFFICIENT_BALANCE"
	reasonMarketNotFound      handlers.FailureReason = "MARKET_NOT_FOUND"
)

// PlaceBetHandler returns an HTTP handler that delegates bet placement to the bets domain service.
func PlaceBetHandler(betsSvc dbets.ServiceInterface, usersSvc dusers.ServiceInterface) http.HandlerFunc {
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

		req, decodeErr := decodePlaceBetRequest(r)
		if decodeErr != nil {
			_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
			return
		}

		placedBet, err := betsSvc.Place(r.Context(), toPlaceRequest(req, user.Username))
		if err != nil {
			writePlaceBetError(w, err)
			return
		}

		writePlaceBetResponse(w, placedBet)
	}
}

func decodePlaceBetRequest(r *http.Request) (dto.PlaceBetRequest, error) {
	var req dto.PlaceBetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return dto.PlaceBetRequest{}, errors.New("Invalid request body")
	}
	return req, nil
}

func toPlaceRequest(req dto.PlaceBetRequest, username string) dbets.PlaceRequest {
	return dbets.PlaceRequest{
		Username: username,
		MarketID: req.MarketID,
		Amount:   req.Amount,
		Outcome:  req.Outcome,
	}
}

func writePlaceBetError(w http.ResponseWriter, err error) {
	switch err {
	case dbets.ErrInvalidOutcome, dbets.ErrInvalidAmount:
		_ = handlers.WriteFailure(w, http.StatusBadRequest, reasonBetValidationFailed)
	case dbets.ErrMarketClosed:
		_ = handlers.WriteFailure(w, http.StatusConflict, reasonMarketClosed)
	case dbets.ErrInsufficientBalance:
		_ = handlers.WriteFailure(w, http.StatusUnprocessableEntity, reasonInsufficientBalance)
	case dmarkets.ErrMarketNotFound:
		_ = handlers.WriteFailure(w, http.StatusNotFound, reasonMarketNotFound)
	default:
		_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
	}
}

func writePlaceBetResponse(w http.ResponseWriter, placedBet *dbets.PlacedBet) {
	response := dto.PlaceBetResponse{
		Username: placedBet.Username,
		MarketID: placedBet.MarketID,
		Amount:   placedBet.Amount,
		Outcome:  placedBet.Outcome,
		PlacedAt: placedBet.PlacedAt,
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
