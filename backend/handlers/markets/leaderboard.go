package marketshandlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"

	"github.com/gorilla/mux"
)

// MarketLeaderboardHandler handles requests for market profitability leaderboards
func MarketLeaderboardHandler(svc dmarkets.ServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Parse HTTP parameters
		vars := mux.Vars(r)
		marketIdStr := vars["marketId"]

		marketId, err := strconv.ParseInt(marketIdStr, 10, 64)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(dto.ErrorResponse{Error: "Invalid market ID"})
			return
		}

		// 2. Parse pagination parameters
		limitStr := r.URL.Query().Get("limit")
		offsetStr := r.URL.Query().Get("offset")

		limit := 100
		if limitStr != "" {
			if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
				limit = parsedLimit
			}
		}

		offset := 0
		if offsetStr != "" {
			if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
				offset = parsedOffset
			}
		}

		page := dmarkets.Page{
			Limit:  limit,
			Offset: offset,
		}

		// 3. Call domain service
		leaderboard, err := svc.GetMarketLeaderboard(r.Context(), marketId, page)
		if err != nil {
			// 4. Map domain errors to HTTP status codes
			switch err {
			case dmarkets.ErrMarketNotFound:
				http.Error(w, "Market not found", http.StatusNotFound)
			case dmarkets.ErrInvalidInput:
				http.Error(w, "Invalid request parameters", http.StatusBadRequest)
			default:
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		// 5. Convert to response DTO
		var leaderRows []dto.LeaderboardRow
		for _, row := range leaderboard {
			leaderRows = append(leaderRows, dto.LeaderboardRow{
				Username: row.Username,
				Profit:   row.Profit,
				Volume:   row.Volume,
				Rank:     row.Rank,
			})
		}

		// 6. Ensure empty array instead of null
		if leaderRows == nil {
			leaderRows = make([]dto.LeaderboardRow, 0)
		}

		// 7. Return response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(dto.LeaderboardResponse{
			MarketID:    marketId,
			Leaderboard: leaderRows,
			Total:       len(leaderRows),
		})
	}
}
