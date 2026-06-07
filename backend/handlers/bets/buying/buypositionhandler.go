package buybetshandlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"socialpredict/handlers"
	"socialpredict/handlers/authhttp"
	"socialpredict/handlers/bets/dto"
	dbets "socialpredict/internal/domain/bets"
	dmarkets "socialpredict/internal/domain/markets"
	dusers "socialpredict/internal/domain/users"
	authsvc "socialpredict/internal/service/auth"
	"socialpredict/logger"
)

type readModelInvalidator interface {
	InvalidateAfterMarketTransaction(ctx context.Context, username string, marketID int64, reason string) error
}

// PlaceBetHandler returns an HTTP handler that delegates bet placement to the bets domain service.
func PlaceBetHandler(betsSvc dbets.ServiceInterface, usersSvc dusers.ServiceInterface) http.HandlerFunc {
	return PlaceBetHandlerWithInvalidator(betsSvc, usersSvc, nil)
}

// PlaceBetHandlerWithInvalidator returns an HTTP handler that delegates bet
// placement and then marks display read models stale after a successful write.
func PlaceBetHandlerWithInvalidator(betsSvc dbets.ServiceInterface, usersSvc dusers.ServiceInterface, invalidator readModelInvalidator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			_ = handlers.WriteFailure(w, http.StatusMethodNotAllowed, handlers.ReasonMethodNotAllowed)
			return
		}

		user, authErr := authsvc.ValidateUserAndEnforcePasswordChangeGetUser(r, usersSvc)
		if authErr != nil {
			_ = authhttp.WriteFailure(w, authErr)
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
		if invalidator != nil {
			if err := invalidator.InvalidateAfterMarketTransaction(r.Context(), placedBet.Username, int64(placedBet.MarketID), "bet_accepted"); err != nil {
				logger.LogError("PlaceBet", "InvalidateReadModels", err)
			}
		}
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
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonValidationFailed)
	case dbets.ErrMarketClosed:
		_ = handlers.WriteFailure(w, http.StatusConflict, handlers.ReasonMarketClosed)
	case dbets.ErrInsufficientBalance:
		_ = handlers.WriteFailure(w, http.StatusUnprocessableEntity, handlers.ReasonInsufficientBalance)
	case dmarkets.ErrMarketNotFound:
		_ = handlers.WriteFailure(w, http.StatusNotFound, handlers.ReasonMarketNotFound)
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
