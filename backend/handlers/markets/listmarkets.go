package marketshandlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"
)

// ListMarketsHandlerFactory creates an HTTP handler for listing markets with service injection
func ListMarketsHandlerFactory(svc dmarkets.ServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("ListMarketsHandler: Request received")
		if r.Method != http.MethodGet {
			http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
			return
		}

		// Parse query parameters
		status, statusErr := normalizeStatusParam(r.URL.Query().Get("status"))
		if statusErr != nil {
			http.Error(w, statusErr.Error(), http.StatusBadRequest)
			return
		}
		limitStr := r.URL.Query().Get("limit")
		offsetStr := r.URL.Query().Get("offset")

		// Parse limit with default
		limit := 50
		if limitStr != "" {
			if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
				limit = parsedLimit
			}
		}

		// Parse offset with default
		offset := 0
		if offsetStr != "" {
			if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
				offset = parsedOffset
			}
		}

		// Build domain filter
		filters := dmarkets.ListFilters{
			Status: status,
			Limit:  limit,
			Offset: offset,
		}

		// If status is provided, delegate to ListByStatus; otherwise use List
		var (
			markets []*dmarkets.Market
			err     error
		)
		if status != "" {
			// Use ListByStatus for status-specific queries
			page := dmarkets.Page{Limit: limit, Offset: offset}
			markets, err = svc.ListByStatus(r.Context(), status, page)
		} else {
			// Use general List method
			markets, err = svc.ListMarkets(r.Context(), filters)
		}

		if err != nil {
			// Map domain errors to HTTP status codes
			switch err {
			case dmarkets.ErrInvalidInput:
				http.Error(w, "Invalid input parameters", http.StatusBadRequest)
			default:
				log.Printf("Error fetching markets: %v", err)
				http.Error(w, "Error fetching markets", http.StatusInternalServerError)
			}
			return
		}

		overviews, err := buildMarketOverviewResponses(r.Context(), svc, markets)
		if err != nil {
			log.Printf("Error building market overviews: %v", err)
			http.Error(w, "Error fetching markets", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(dto.ListMarketsResponse{
			Markets: overviews,
			Total:   len(overviews),
		}); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
		}
	}
}
