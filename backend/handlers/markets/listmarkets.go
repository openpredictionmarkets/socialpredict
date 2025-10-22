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
func ListMarketsHandlerFactory(svc dmarkets.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("ListMarketsHandler: Request received")
		if r.Method != http.MethodGet {
			http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
			return
		}

		// Parse query parameters
		status := r.URL.Query().Get("status")
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
		var markets []*dmarkets.Market
		var err error

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

		// Convert domain markets to response DTOs
		var marketResponses []*dto.MarketResponse
		for _, market := range markets {
			marketResponse := &dto.MarketResponse{
				ID:                 market.ID,
				QuestionTitle:      market.QuestionTitle,
				Description:        market.Description,
				OutcomeType:        market.OutcomeType,
				ResolutionDateTime: market.ResolutionDateTime,
				CreatorUsername:    market.CreatorUsername,
				YesLabel:           market.YesLabel,
				NoLabel:            market.NoLabel,
				Status:             market.Status,
				CreatedAt:          market.CreatedAt,
				UpdatedAt:          market.UpdatedAt,
			}
			marketResponses = append(marketResponses, marketResponse)
		}

		// Normalize empty list → [] (ensure empty array instead of null)
		if marketResponses == nil {
			marketResponses = make([]*dto.MarketResponse, 0)
		}

		// Encode dto.ListResponse
		response := dto.SimpleListMarketsResponse{
			Markets: marketResponses,
			Total:   len(marketResponses),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
		}
	}
}
