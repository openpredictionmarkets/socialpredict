package positions

import (
	"errors"
	"net/http"
	"strconv"

	"socialpredict/handlers"
	dmarkets "socialpredict/internal/domain/markets"
	"socialpredict/logger"

	"github.com/gorilla/mux"
)

// MarketPositionsHandlerWithService creates a service-injected positions handler for all users
func MarketPositionsHandlerWithService(svc dmarkets.ServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			_ = handlers.WriteFailure(w, http.StatusMethodNotAllowed, handlers.ReasonMethodNotAllowed)
			return
		}

		marketID, err := parseMarketID(mux.Vars(r)["marketId"])
		if err != nil {
			_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
			return
		}

		positions, err := getMarketPositions(r, svc, marketID)
		if err != nil {
			writePositionsError(w, marketID, err)
			return
		}

		responses := mapPositionsToResponses(positions)

		if err := handlers.WriteResult(w, http.StatusOK, responses); err != nil {
			logger.LogError("MarketPositions", "WriteResponse", err)
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
		}
	}
}

func getMarketPositions(r *http.Request, svc dmarkets.ServiceInterface, marketID int64) (dmarkets.MarketPositions, error) {
	if !hasPaginationQuery(r) {
		return svc.GetMarketPositions(r.Context(), marketID)
	}
	return svc.GetMarketPositionsPage(r.Context(), marketID, parsePage(r, 20))
}

// MarketUserPositionHandlerWithService creates a service-injected handler for a specific user's position
func MarketUserPositionHandlerWithService(svc dmarkets.ServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			_ = handlers.WriteFailure(w, http.StatusMethodNotAllowed, handlers.ReasonMethodNotAllowed)
			return
		}

		marketID, username, err := parseMarketUserParams(mux.Vars(r))
		if err != nil {
			_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
			return
		}

		position, err := svc.GetUserPositionInMarket(r.Context(), marketID, username)
		if err != nil {
			writeUserPositionError(w, marketID, username, err)
			return
		}

		if err := handlers.WriteResult(w, http.StatusOK, newUserPositionResponse(position)); err != nil {
			logger.LogError("MarketUserPosition", "WriteResponse", err)
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

func hasPaginationQuery(r *http.Request) bool {
	query := r.URL.Query()
	return query.Get("limit") != "" || query.Get("offset") != ""
}

func parsePage(r *http.Request, defaultLimit int) dmarkets.Page {
	query := r.URL.Query()
	limit := defaultLimit
	if raw := query.Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}
	offset := 0
	if raw := query.Get("offset"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			offset = parsed
		}
	}
	return dmarkets.Page{Limit: limit, Offset: offset}
}

func writePositionsError(w http.ResponseWriter, marketID int64, err error) {
	switch err {
	case dmarkets.ErrMarketNotFound:
		_ = handlers.WriteFailure(w, http.StatusNotFound, handlers.ReasonMarketNotFound)
	case dmarkets.ErrInvalidInput:
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonValidationFailed)
	default:
		logger.LogError("MarketPositions", "GetMarketPositions", err)
		_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
	}
}

func mapPositionsToResponses(positions []*dmarkets.UserPosition) []userPositionResponse {
	responses := make([]userPositionResponse, 0, len(positions))
	for _, pos := range positions {
		if pos == nil {
			continue
		}
		responses = append(responses, newUserPositionResponse(pos))
	}
	return responses
}

func parseMarketUserParams(vars map[string]string) (int64, string, error) {
	marketID, err := parseMarketID(vars["marketId"])
	if err != nil {
		return 0, "", err
	}

	username := vars["username"]
	if username == "" {
		return 0, "", errors.New("Username is required")
	}

	return marketID, username, nil
}

func writeUserPositionError(w http.ResponseWriter, marketID int64, username string, err error) {
	switch err {
	case dmarkets.ErrMarketNotFound:
		_ = handlers.WriteFailure(w, http.StatusNotFound, handlers.ReasonMarketNotFound)
	case dmarkets.ErrUserNotFound:
		_ = handlers.WriteFailure(w, http.StatusNotFound, handlers.ReasonUserNotFound)
	case dmarkets.ErrInvalidInput:
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonValidationFailed)
	default:
		logger.LogError("MarketUserPosition", "GetUserPositionInMarket", err)
		_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
	}
}
