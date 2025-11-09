package marketshandlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"
)

// GetMarketsHandler handles requests for listing all markets (alias for ListMarketsHandler)
func GetMarketsHandler(svc dmarkets.ServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Parse query parameters for filtering
		status, err := normalizeStatusParam(r.URL.Query().Get("status"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		limitStr := r.URL.Query().Get("limit")
		offsetStr := r.URL.Query().Get("offset")

		// Parse with defaults
		limit := 100
		if limitStr != "" {
			if parsedLimit, err := parseIntOrDefault(limitStr, 100); err == nil {
				limit = parsedLimit
			}
		}

		offset := 0
		if offsetStr != "" {
			if parsedOffset, err := parseIntOrDefault(offsetStr, 0); err == nil {
				offset = parsedOffset
			}
		}

		// 2. Build domain filter
		filters := dmarkets.ListFilters{
			Status: status,
			Limit:  limit,
			Offset: offset,
		}

		// 3. Call domain service
		markets, err := svc.ListMarkets(r.Context(), filters)
		if err != nil {
			// 4. Map domain errors to HTTP status codes
			switch err {
			case dmarkets.ErrInvalidInput:
				http.Error(w, "Invalid input parameters", http.StatusBadRequest)
			default:
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		overviews, err := buildMarketOverviewResponses(r.Context(), svc, markets)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// 7. Return response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(dto.ListMarketsResponse{
			Markets: overviews,
			Total:   len(overviews),
		})
	}
}

// Helper function for parsing integers with defaults
func parseIntOrDefault(s string, defaultVal int) (int, error) {
	if s == "" {
		return defaultVal, nil
	}
	return strconv.Atoi(s)
}
