package marketshandlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"
	"socialpredict/security"
)

// SearchMarketsHandler handles HTTP requests for searching markets - HTTP-only with service injection
func SearchMarketsHandler(svc dmarkets.ServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("SearchMarketsHandler: Request received")
		if r.Method != http.MethodGet {
			http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
			return
		}

		// Read q, limit, offset from query
		query := r.URL.Query().Get("q")
		if query == "" {
			// Also check 'query' parameter for backward compatibility
			query = r.URL.Query().Get("query")
		}
		status := r.URL.Query().Get("status")
		limitStr := r.URL.Query().Get("limit")
		offsetStr := r.URL.Query().Get("offset")

		// Validate and sanitize input
		if query == "" {
			http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
			return
		}

		// Sanitize the search query
		sanitizer := security.NewSanitizer()
		sanitizedQuery, err := sanitizer.SanitizeMarketTitle(query)
		if err != nil {
			log.Printf("SearchMarketsHandler: Sanitization failed for query '%s': %v", query, err)
			http.Error(w, "Invalid search query: "+err.Error(), http.StatusBadRequest)
			return
		}
		if len(sanitizedQuery) > 100 {
			http.Error(w, "Query too long (max 100 characters)", http.StatusBadRequest)
			return
		}

		// Parse limit and offset
		limit := 20 // Default
		if limitStr != "" {
			if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 50 {
				limit = parsedLimit
			}
		}

		offset := 0 // Default
		if offsetStr != "" {
			if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
				offset = parsedOffset
			}
		}

		// Build f := dmarkets.SearchFilters{Limit: limit, Offset: offset}
		filters := dmarkets.SearchFilters{
			Status: status, // Can be empty, "active", "closed", "resolved", or "all"
			Limit:  limit,
			Offset: offset,
		}

		// ms, err := h.service.SearchMarkets(r.Context(), q, f)
		searchResults, err := svc.SearchMarkets(r.Context(), sanitizedQuery, filters)
		if err != nil {
			// Map errors
			switch err {
			case dmarkets.ErrInvalidInput:
				http.Error(w, "Invalid search parameters", http.StatusBadRequest)
			default:
				log.Printf("Error searching markets: %v", err)
				http.Error(w, "Error searching markets", http.StatusInternalServerError)
			}
			return
		}

		// Map domain → []dto.Market; ensure non-nil slice
		var primaryMarkets []dto.MarketResponse
		var fallbackMarkets []dto.MarketResponse

		// Convert primary results
		for _, market := range searchResults.PrimaryResults {
			marketDTO := dto.MarketResponse{
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
			primaryMarkets = append(primaryMarkets, marketDTO)
		}

		// Convert fallback results
		for _, market := range searchResults.FallbackResults {
			marketDTO := dto.MarketResponse{
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
			fallbackMarkets = append(fallbackMarkets, marketDTO)
		}

		// Ensure non-nil slice
		if primaryMarkets == nil {
			primaryMarkets = make([]dto.MarketResponse, 0)
		}
		if fallbackMarkets == nil {
			fallbackMarkets = make([]dto.MarketResponse, 0)
		}

		// Build search response using domain service results
		response := dto.SearchResponse{
			PrimaryResults:  primaryMarkets,
			FallbackResults: fallbackMarkets,
			Query:           searchResults.Query,
			PrimaryStatus:   searchResults.PrimaryStatus,
			PrimaryCount:    searchResults.PrimaryCount,
			FallbackCount:   searchResults.FallbackCount,
			TotalCount:      searchResults.TotalCount,
			FallbackUsed:    searchResults.FallbackUsed,
		}

		// Encode response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Error encoding search response: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
