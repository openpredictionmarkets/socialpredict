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
		params, err := parseGetMarketsParams(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// 3. Call domain service
		markets, err := svc.ListMarkets(r.Context(), params.filters)
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

type getMarketsParams struct {
	status  string
	limit   int
	offset  int
	filters dmarkets.ListFilters
}

func parseGetMarketsParams(r *http.Request) (getMarketsParams, error) {
	status, err := normalizeStatusParam(r.URL.Query().Get("status"))
	if err != nil {
		return getMarketsParams{}, err
	}

	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := parseIntOrDefault(limitStr, 100); err == nil {
			limit = parsedLimit
		}
	}

	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := parseIntOrDefault(offsetStr, 0); err == nil {
			offset = parsedOffset
		}
	}

	return getMarketsParams{
		status: status,
		limit:  limit,
		offset: offset,
		filters: dmarkets.ListFilters{
			Status: status,
			Limit:  limit,
			Offset: offset,
		},
	}, nil
}
