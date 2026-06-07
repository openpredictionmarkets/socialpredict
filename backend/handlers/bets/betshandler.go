package betshandlers

import (
	"errors"
	"net/http"
	"strconv"

	"socialpredict/handlers"
	dmarkets "socialpredict/internal/domain/markets"
	"socialpredict/logger"

	"github.com/gorilla/mux"
)

// MarketBetsHandlerWithService creates a service-injected bets handler
func MarketBetsHandlerWithService(svc dmarkets.ServiceInterface) http.HandlerFunc {
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

		betsDisplayInfo, err := getMarketBets(r, svc, marketID)
		if err != nil {
			writeMarketBetsError(w, marketID, err)
			return
		}

		if err := handlers.WriteResult(w, http.StatusOK, betsDisplayInfo); err != nil {
			logger.LogError("MarketBets", "WriteResponse", err)
		}
	}
}

func getMarketBets(r *http.Request, svc dmarkets.ServiceInterface, marketID int64) ([]*dmarkets.BetDisplayInfo, error) {
	if !hasPaginationQuery(r) {
		return svc.GetMarketBets(r.Context(), marketID)
	}
	return svc.GetMarketBetsPage(r.Context(), marketID, parsePage(r, 20))
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

func writeMarketBetsError(w http.ResponseWriter, marketID int64, err error) {
	switch err {
	case dmarkets.ErrMarketNotFound:
		_ = handlers.WriteFailure(w, http.StatusNotFound, handlers.ReasonMarketNotFound)
	case dmarkets.ErrInvalidInput:
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
	default:
		logger.LogError("MarketBets", "GetMarketBets", err)
		_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
	}
}
