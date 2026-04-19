package usershandlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"socialpredict/handlers"

	dmarkets "socialpredict/internal/domain/markets"
	dusers "socialpredict/internal/domain/users"
	authsvc "socialpredict/internal/service/auth"
)

// UserMarketPositionHandlerWithService returns an HTTP handler that resolves the authenticated
// user's position in the specified market via the markets service.
func UserMarketPositionHandlerWithService(marketSvc dmarkets.ServiceInterface, usersSvc dusers.ServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			_ = handlers.WriteFailure(w, http.StatusMethodNotAllowed, handlers.ReasonMethodNotAllowed)
			return
		}

		user, httperr := authsvc.ValidateUserAndEnforcePasswordChangeGetUser(r, usersSvc)
		if httperr != nil {
			_ = handlers.WriteFailure(w, httperr.StatusCode, profileAuthFailureReason(httperr))
			return
		}

		marketID, err := parseMarketID(mux.Vars(r)["marketId"])
		if err != nil {
			_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
			return
		}

		position, err := marketSvc.GetUserPositionInMarket(r.Context(), marketID, user.Username)
		if err != nil {
			writeUserPositionError(w, marketID, user.Username, err)
			return
		}

		if err := handlers.WriteResult(w, http.StatusOK, position); err != nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
		}
	}
}

func parseMarketID(marketIDStr string) (int64, error) {
	if marketIDStr == "" {
		return 0, errors.New("Market ID is required")
	}

	marketID, err := strconv.ParseInt(marketIDStr, 10, 64)
	if err != nil {
		return 0, errors.New("Invalid market ID")
	}
	return marketID, nil
}

func writeUserPositionError(w http.ResponseWriter, marketID int64, username string, err error) {
	switch err {
	case dmarkets.ErrMarketNotFound:
		_ = handlers.WriteFailure(w, http.StatusNotFound, "MARKET_NOT_FOUND")
	case dmarkets.ErrUserNotFound:
		_ = handlers.WriteFailure(w, http.StatusNotFound, "USER_NOT_FOUND")
	case dmarkets.ErrInvalidInput:
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
	default:
		_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
	}
}
