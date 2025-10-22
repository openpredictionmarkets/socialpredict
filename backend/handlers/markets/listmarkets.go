package marketshandlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"
)

// ListMarketsHandler handles the HTTP request for listing markets with enriched data
func ListMarketsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("ListMarketsHandler: Request received")
	if r.Method != http.MethodGet {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	// Parse query parameters
	status := r.URL.Query().Get("status")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	// Parse limit with default
	limit := 50
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
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

	// TODO: Get service from dependency injection - for now this will fail
	// This needs to be wired through the container when we implement full DI
	var svc dmarkets.Service // This will be nil and cause panic - needs proper wiring

	// Call domain service for enriched market data
	overviews, err := svc.GetMarketOverviews(context.Background(), filters)
	if err != nil {
		// Map domain errors to HTTP status codes
		switch err {
		case dmarkets.ErrMarketNotFound:
			http.Error(w, "Markets not found", http.StatusNotFound)
		default:
			log.Printf("Error fetching market overviews: %v", err)
			http.Error(w, "Error fetching markets", http.StatusInternalServerError)
		}
		return
	}

	// Convert domain overviews to response DTOs
	var responseOverviews []*dto.MarketOverviewResponse
	for _, overview := range overviews {
		// For now, create a basic creator response from the available data
		creator := &dto.CreatorResponse{
			Username:      overview.Market.CreatorUsername,
			PersonalEmoji: "👤", // Default emoji - TODO: Get from user service
			DisplayName:   overview.Market.CreatorUsername,
		}

		responseOverview := &dto.MarketOverviewResponse{
			Market: &dto.MarketResponse{
				ID:                 overview.Market.ID,
				QuestionTitle:      overview.Market.QuestionTitle,
				Description:        overview.Market.Description,
				OutcomeType:        overview.Market.OutcomeType,
				ResolutionDateTime: overview.Market.ResolutionDateTime,
				CreatorUsername:    overview.Market.CreatorUsername,
				YesLabel:           overview.Market.YesLabel,
				NoLabel:            overview.Market.NoLabel,
				Status:             overview.Market.Status,
				CreatedAt:          overview.Market.CreatedAt,
				UpdatedAt:          overview.Market.UpdatedAt,
			},
			Creator:         creator,
			LastProbability: overview.LastProbability,
			NumUsers:        overview.NumUsers,
			TotalVolume:     overview.TotalVolume,
		}
		responseOverviews = append(responseOverviews, responseOverview)
	}

	// Ensure empty array instead of null
	if responseOverviews == nil {
		responseOverviews = make([]*dto.MarketOverviewResponse, 0)
	}

	// Build response
	response := dto.ListMarketsResponse{
		Markets: responseOverviews,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Legacy handler factory function for backward compatibility
func ListMarketsHandlerFactory(svc dmarkets.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("ListMarketsHandler: Request received")
		if r.Method != http.MethodGet {
			http.Error(w, "Method is not supported.", http.StatusNotFound)
			return
		}

		// Parse query parameters
		status := r.URL.Query().Get("status")
		limitStr := r.URL.Query().Get("limit")
		offsetStr := r.URL.Query().Get("offset")

		// Parse limit with default
		limit := 50
		if limitStr != "" {
			if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
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

		// Call domain service for enriched market data
		overviews, err := svc.GetMarketOverviews(context.Background(), filters)
		if err != nil {
			// Map domain errors to HTTP status codes
			switch err {
			case dmarkets.ErrMarketNotFound:
				http.Error(w, "Markets not found", http.StatusNotFound)
			default:
				log.Printf("Error fetching market overviews: %v", err)
				http.Error(w, "Error fetching markets", http.StatusInternalServerError)
			}
			return
		}

		// Convert domain overviews to response DTOs
		var responseOverviews []*dto.MarketOverviewResponse
		for _, overview := range overviews {
			// For now, create a basic creator response from the available data
			creator := &dto.CreatorResponse{
				Username:      overview.Market.CreatorUsername,
				PersonalEmoji: "👤", // Default emoji - TODO: Get from user service
				DisplayName:   overview.Market.CreatorUsername,
			}

			responseOverview := &dto.MarketOverviewResponse{
				Market: &dto.MarketResponse{
					ID:                 overview.Market.ID,
					QuestionTitle:      overview.Market.QuestionTitle,
					Description:        overview.Market.Description,
					OutcomeType:        overview.Market.OutcomeType,
					ResolutionDateTime: overview.Market.ResolutionDateTime,
					CreatorUsername:    overview.Market.CreatorUsername,
					YesLabel:           overview.Market.YesLabel,
					NoLabel:            overview.Market.NoLabel,
					Status:             overview.Market.Status,
					CreatedAt:          overview.Market.CreatedAt,
					UpdatedAt:          overview.Market.UpdatedAt,
				},
				Creator:         creator,
				LastProbability: overview.LastProbability,
				NumUsers:        overview.NumUsers,
				TotalVolume:     overview.TotalVolume,
			}
			responseOverviews = append(responseOverviews, responseOverview)
		}

		// Ensure empty array instead of null
		if responseOverviews == nil {
			responseOverviews = make([]*dto.MarketOverviewResponse, 0)
		}

		// Build response
		response := dto.ListMarketsResponse{
			Markets: responseOverviews,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
