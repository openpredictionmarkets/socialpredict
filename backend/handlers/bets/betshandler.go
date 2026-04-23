package betshandlers

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"socialpredict/handlers"
	dmarkets "socialpredict/internal/domain/markets"

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

		betsDisplayInfo, err := svc.GetMarketBets(r.Context(), marketID)
		if err != nil {
			writeMarketBetsError(w, marketID, err)
			return
		}

		if err := handlers.WriteResult(w, http.StatusOK, betsDisplayInfo); err != nil {
			log.Printf("Error encoding bets response: %v", err)
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

func writeMarketBetsError(w http.ResponseWriter, marketID int64, err error) {
	switch err {
	case dmarkets.ErrMarketNotFound:
		_ = handlers.WriteFailure(w, http.StatusNotFound, handlers.ReasonMarketNotFound)
	case dmarkets.ErrInvalidInput:
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
	default:
		log.Printf("Error getting market bets for market %d: %v", marketID, err)
		_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
	}
}
